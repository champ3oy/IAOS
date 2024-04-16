package sms

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type SMSParams struct {
	Recipients string
	Message    string
}

func SendWithNalo(payload SMSParams) error {
	phoneNumber := payload.Recipients
	phoneNumberWithoutPlus := strings.ReplaceAll(phoneNumber, "+", "")

	log.Println(phoneNumberWithoutPlus, payload.Message)

	key := "3p7pah73@4fkwei7bod@4xjkbanz_6bj14u)@r17zr_u(0ge@0jx(ntg8uuhukox"
	apiURL := "https://sms.nalosolutions.com/smsbackend/clientapi/Resl_Nalo/send-message/"

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

	return nil
}
