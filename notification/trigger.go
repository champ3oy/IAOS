package notification

import (
	"fmt"
	"issue-reporting/auth"
	"issue-reporting/call"
	"issue-reporting/database"
	"issue-reporting/email"
	pushnotification "issue-reporting/push-notification"
	"issue-reporting/slack"
	"issue-reporting/sms"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func SendNotification(message string, user auth.User) {
	done := make(chan bool)
	start := time.Now()

	var team auth.Team
	err := database.FindOne("teams", bson.M{"teamId": user.TeamId}).Decode(&team)
	if err != nil {
		fmt.Println("error find team")
	}

	for _, notification := range team.Notifications {
		go func(notification auth.Notification) {
			switch notification.Channel {
			case "SMS":
				if notification.Use {
					p := sms.SMSParams{
						Recipients: user.PhoneNumber,
						Message:    message,
					}
					sms.SendWithNalo(p)
				}
			case "Slack":
				if notification.Use {
					_ = slack.Notify(&slack.NotifyParams{Text: message})
				}
			case "Email":
				if notification.Use {
					email.SendWithResend(email.EmailParams{
						Recipients: user.Email,
						Subject:    "Incident Report Alert 🆘🚨",
						Message:    message,
					})
				} else {
					log.Println("Email was off")
				}
			case "Push notification":
				if notification.Use {
					pushnotification.SendPushNotification(&pushnotification.PushParams{
						Title: "Incident Report Alert 🆘🚨",
						Body:  message,
						Type:  "incident",
					})
				}
			case "Call":
				if notification.Use {
					call.MakeCall(user.PhoneNumber, "Incident Report Alert "+message)
				}
			default:
				fmt.Println("Unknown notification method: ", notification.Channel)
			}

			done <- true
		}(notification)
	}

	for range team.Notifications {
		<-done
	}

	fmt.Println("Total time taken:", time.Since(start))
}
