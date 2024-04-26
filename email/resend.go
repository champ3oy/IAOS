package email

import (
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailParams struct {
	Recipients string
	Subject    string
	Message    string
}

func SendWithResend(payload EmailParams) {
	apiKey := os.Getenv("RESEND_API")

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "Samarithan <incident@stashterminal.com>",
		To:      []string{payload.Recipients},
		Html:    "<strong>" + payload.Message + "</strong>",
		Text:    payload.Message,
		Subject: payload.Subject,
		// Cc:      []string{"cc@example.com"},
		// Bcc:     []string{"bcc@example.com"},
		ReplyTo: "incident@stashterminal.com",
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(sent.Id)
}
