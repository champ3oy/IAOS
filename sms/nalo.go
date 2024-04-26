package sms

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type SMSParams struct {
	Recipients string
	Message    string
}

// func encodeAPIKey(apiKey string) string {
// 	// Encode the API key
// 	encodedKey := url.QueryEscape(apiKey)
// 	// Replace any special characters that are not encoded by url.QueryEscape
// 	encodedKey = strings.ReplaceAll(encodedKey, "@", "%40")
// 	// You might need to replace other characters based on your API key's format

// 	return encodedKey
// }

func SendWithNalo(payload SMSParams) error {
	phoneNumber := payload.Recipients
	phoneNumberWithoutPlus := strings.ReplaceAll(phoneNumber, "+", "")

	log.Println(phoneNumberWithoutPlus, payload.Message)

	key := os.Getenv("NALO_API")
	apiURL := "https://sms.nalosolutions.com/smsbackend/Resl_Nalo/send-message/"

	params := url.Values{}
	params.Set("key", key)
	params.Set("type", "0")
	params.Set("destination", phoneNumberWithoutPlus)
	params.Set("dlr", "1")
	params.Set("source", "Samarithan")
	params.Set("message", payload.Message)
	apiURL += "?" + params.Encode()

	fmt.Println("HTTP JSON GET URL:", apiURL)

	var jsonData = []byte(`{}`)

	req, err := http.NewRequest("GET", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating HTTP request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error sending SMS:", err)
		return err
	}

	defer res.Body.Close()
	log.Println("nalo response", req.Body)
	return nil
}
