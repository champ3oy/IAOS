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

// StartScheduler starts the cron scheduler
func StartNotifyAcknowlegedScheduler() {
	// Create a new cron scheduler
	c := cron.New()

	// Add a cronjob to the scheduler
	_, err := c.AddFunc("@every 10m", func() {
		cursor, err := incidents.List() // we never query for acknowledged incidents
		if err != nil {
			log.Fatalf("Error starting cronjob: %v", err)
		}
		if cursor == nil {
			log.Fatalf("Error starting cronjob: %v", err)
		}

		var items = []string{"These incidents have not been acknowledged yet. \nPlease acknowledge them otherwise you will be reminded every 10 minutes: \n"}
		for _, incident := range cursor {
			var assignee string
			if incident.AssignedTo.FirstName != "" {
				assignee = fmt.Sprintf("%s %s @%s", incident.AssignedTo.FirstName, incident.AssignedTo.LastName, incident.AssignedTo.SlackHandle)
			} else {
				assignee = "Unassigned"
			}

			items = append(items, fmt.Sprintf("[%s] \n\n[#%s] %s", assignee, incident.Id, incident.Title))
		}

		if len(cursor) > 0 {
			_ = slack.Notify(&slack.NotifyParams{Text: strings.Join(items, "\n")})
		}

	})
	if err != nil {
		log.Fatalf("Error adding cronjob: %v", err)
	}

	// Start the cron scheduler
	c.Start()
}

// StartScheduler starts the cron scheduler
func StartNotifyAssignScheduler() {
	c := cron.New()

	// Add a cronjob to the scheduler
	_, err := c.AddFunc("@every 5m", func() {
		// If this code fail, it is either a server error or someone isn't on-call
		schedule, err := schedules.ScheduledNow()
		if err != nil {
			log.Printf("Error starting cronjob1: %v", err)
		}

		cursor, err := incidents.List() // we never query for acknowledged incidents
		if err != nil {
			log.Printf("Error starting cronjob2: %v", err)
		}

		for _, incident := range cursor {
			if incident.AssignedTo.FirstName != "" {
				continue // this incident has already been assigned
			}

			_, err := incidents.Assign(incident.Id, &incidents.AssignParams{User: schedule.User})

			if err == nil {
				log.Println("OK assigned unassigned incident", "incident", incident, "user", schedule.User)
			} else {
				log.Println(err)
				log.Println("FAIL to assign unassigned incident", "incident", incident, "user", schedule.User, "err", err)

			}
		}
	})
	if err != nil {
		log.Fatalf("Error adding cronjob: %v", err)
	}

	// Start the cron scheduler
	c.Start()
}
