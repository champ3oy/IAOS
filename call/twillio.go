package call

import (
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

func MakeCall(to string, message string) error {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	from := os.Getenv("TWILIO_FROM_PHONE_NUMBER")

	// Construct TwiML for the message
	twiml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
                        <Response>
                            <Say voice="woman" language="en-gb" input="test">%s</Say>
							<Hangup/>
                        </Response>`, message)

	// Make call
	params := &twilioApi.CreateCallParams{}
	params.SetTo(to)
	params.SetFrom(from)
	// params.SetUrl("http://twimlets.com/holdmusic?Bucket=com.twilio.music.ambient")
	params.SetTwiml(twiml)

	resp, err := client.Api.CreateCall(params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Call Status: " + *resp.Status)
		fmt.Println("Call Sid: " + *resp.Sid)
		fmt.Println("Call Direction: " + *resp.Direction)
	}

	return nil
}
