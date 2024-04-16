package cron

import (
	"fmt"
	"issue-reporting/incidents"
	"issue-reporting/schedules"
	"issue-reporting/slack"
	"log"
	"strings"

	"github.com/robfig/cron/v3"
)

func StartNotifyAcknowlegedScheduler() {
	c := cron.New()

	_, err := c.AddFunc("@every 10m", func() {
		cursor, err := incidents.List()
		if err != nil {
			log.Fatalf("Error starting cronjob: %v", err)
		}
		if cursor == nil {
			log.Fatalf("Error starting cronjob: %v", err)
		}

		var items = []string{"These incidents have not been acknowledged yet. \nPlease acknowledge them otherwise you will be reminded every 10 minutes: \n"}
		for _, incident := range cursor {
			for _, user := range incident.AssignedTo {
				var assignee string
				if user.FirstName != "" {
					assignee = fmt.Sprintf("%s %s @%s", user.FirstName, user.LastName, user.SlackHandle)
				} else {
					assignee = "Unassigned"
				}

				items = append(items, fmt.Sprintf("[%s] \n\n[#%s] %s", assignee, incident.Id, incident.Title))

			}
		}

		if len(cursor) > 0 {
			_ = slack.Notify(&slack.NotifyParams{Text: strings.Join(items, "\n")})
		}

	})
	if err != nil {
		log.Fatalf("Error adding cronjob: %v", err)
	}

	c.Start()
}

func StartNotifyAssignScheduler() {
	c := cron.New()

	_, err := c.AddFunc("@every 2m", func() {

		schedule, err := schedules.ScheduledNow()
		if err != nil {
			log.Printf("Error starting cronjob1: %v", err)
		}

		cursor, err := incidents.List()
		if err != nil {
			log.Printf("Error starting cronjob2: %v", err)
		}

		for _, incident := range cursor {
			for _, user := range incident.AssignedTo {
				if user.FirstName != "" {
					continue
				}

				_, err := incidents.Assign(incident.Id, &incidents.AssignParams{User: schedule.User})

				if err == nil {
					log.Println("OK assigned unassigned incident", "incident", incident, "user", schedule.User)
				} else {
					log.Println(err)
					log.Println("FAIL to assign unassigned incident", "incident", incident, "user", schedule.User, "err", err)

				}
			}
		}
	})
	if err != nil {
		log.Fatalf("Error adding cronjob: %v", err)
	}

	c.Start()
}
