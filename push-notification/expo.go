package pushnotification

import (
	"fmt"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

type PushParams struct {
	PushToken string
	Title     string
	Body      string
	Type      string
}

func SendPushNotification(payload *PushParams) {
	pushToken, err := expo.NewExponentPushToken(payload.PushToken)
	if err != nil {
		panic(err)
	}

	// Create a new Expo SDK client
	client := expo.NewPushClient(nil)

	// Publish message
	response, err := client.Publish(
		&expo.PushMessage{
			To:       []expo.ExponentPushToken{pushToken},
			Body:     payload.Body,
			Data:     map[string]string{"type": payload.Type},
			Sound:    "default",
			Title:    payload.Title,
			Priority: expo.DefaultPriority,
		},
	)

	if err != nil {
		panic(err)
	}

	if response.ValidateResponse() != nil {
		fmt.Println(response.PushMessage.To, "failed")
	}
}
