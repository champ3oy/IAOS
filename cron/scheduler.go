package cron

import (
	"context"
	"fmt"
	"issue-reporting/database"
	"issue-reporting/incidents"
	"issue-reporting/reports"
	"issue-reporting/schedules"
	"issue-reporting/slack"
	"log"
	"strings"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

	_, err := c.AddFunc("@every 5m", func() {

		schedule, err := schedules.ScheduledNow()
		if err != nil {
			log.Printf("Error starting cronjob1: %v", err)
		}

		cursor, err := incidents.List()
		if err != nil {
			log.Printf("Error starting cronjob2: %v", err)
		}

		for _, incident := range cursor {
			if incident.AssignedTo == nil || len(incident.AssignedTo) == 0 {
				_, err := incidents.Assign(incident.Id, &incidents.AssignParams{User: schedule.User})

				if err == nil {
					log.Println("OK assigned unassigned incident", "incident", incident, "user", schedule.User)
				} else {
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

func ReportGeneratorScheduler() {
	c := cron.New()
	_, err := c.AddFunc("@every 30m", func() {
		cursor, err := incidents.ResolvedList()
		if err != nil {
			log.Fatalf("Error starting cronjob gen: %v", err)
		}
		if cursor == nil {
			log.Fatalf("Error starting cronjob gen: %v", err)
		}

		reportslist := make([]reports.Report, len(cursor))

		for i, incident := range cursor {
			report := reports.GeneratePDF(incident)
			reportslist[i] = report
		}

		var interfaceSlice []interface{}
		for _, r := range reportslist {
			interfaceSlice = append(interfaceSlice, r)
		}

		// Add generated reports to the database
		database.InsertMany("reports", interfaceSlice)

		var bulkWrites []mongo.WriteModel
		for _, incident2 := range cursor {
			filter := bson.M{"id": incident2.Id}
			update := bson.M{"$set": bson.M{"reportCreated": true}}
			model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
			bulkWrites = append(bulkWrites, model)
		}

		// Execute the bulk update
		_, err = database.GetDatabase().Database("IssueReporting").Collection("incidents").BulkWrite(context.Background(), bulkWrites)
		if err != nil {
			log.Println(err)
		}
	})
	if err != nil {
		log.Fatalf("Error adding cronjob: %v", err)
	}

	c.Start()
}
