package notification

import (
	"fmt"
	"issue-reporting/auth"
	"issue-reporting/call"
	"issue-reporting/email"
	pushnotification "issue-reporting/push-notification"
	"issue-reporting/schedules"
	"issue-reporting/slack"
	"issue-reporting/sms"
	"log"
	"time"
)

func SendNotification(message string, team auth.Teams) {
	done := make(chan bool)
	start := time.Now()

	schedule, err := schedules.ScheduledNow()
	if err != nil {
		log.Printf("Error getting schedule: %v", err)
	}

	for _, notification := range team.Notifications {
		go func(notification auth.Notification) {
			<-time.After(notification.TimeToTrigger)

			switch notification.Channel {
			case "SMS":
				p := sms.SMSParams{
					Recipients: schedule.User.PhoneNumber,
					Message:    message,
				}
				sms.SendWithNalo(p)
			case "Slack":
				_ = slack.Notify(&slack.NotifyParams{Text: message})
			case "Email":
				email.SendWithResend(email.EmailParams{
					Recipients: team.Email,
					Subject:    "Incident Report Alert 🆘🚨",
					Message:    message,
				})
			case "Push notification":
				pushnotification.SendPushNotification(&pushnotification.PushParams{
					Title: "Incident Report Alert 🆘🚨",
					Body:  message,
					Type:  "incident",
				})
			case "Call":
				call.MakeCall(schedule.User.PhoneNumber, "Incident Report Alert"+message)
			default:
				fmt.Println("Unknown notification method:", notification.Channel)
			}

			done <- true
		}(notification)
	}

	for range team.Notifications {
		<-done
	}

	fmt.Println("Total time taken:", time.Since(start))

}

func sendPushNotification(message string) {
	fmt.Println("Sending Push notification:", message)
}
