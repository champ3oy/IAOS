package cron

import (
	"context"
	"fmt"
	"issue-reporting/database"
	"issue-reporting/incidents"
	"issue-reporting/notification"
	"issue-reporting/reports"
	"issue-reporting/schedules"
	"log"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func StartNotifyAcknowlegedScheduler() {
	c := cron.New()

	_, err := c.AddFunc("@every 10m", func() {
		cursor, err := incidents.List()
		if err != nil {
			log.Printf("Error starting cronjob: %v", err)
		}
		if cursor == nil {
			log.Printf("Error starting cronjob: %v", err)
		}

		for _, incident := range cursor {
			for _, user := range incident.AssignedTo {
				var assignee string
				if user.Name != "" {
					assignee = fmt.Sprintf("%s @%s", user.Name, user.SlackHandle)
				} else {
					assignee = "Unassigned"
				}

				notification.SendNotification(fmt.Sprintf("This incident has not been acknowledged yet. \nPlease acknowledge them otherwise you will be reminded every 10 minutes: \n\n[%s] \n\n[#%s] %s", assignee, incident.Id, incident.Title), user)
			}
		}
	})
	if err != nil {
		log.Printf("Error adding cronjob: %v", err)
	}

	c.Start()
}

func StartNotifyAssignScheduler() {
	c := cron.New()
	_, err := c.AddFunc("@every 5m", func() {

		cursor, err := incidents.List()
		if err != nil {
			log.Printf("Error starting cronjob2: %v", err)
		}

		for _, incident := range cursor {
			schedule, err := schedules.ScheduledNow(incident.TeamId)
			if err != nil {
				log.Printf("Error starting cronjob1: %v", err)
			}
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
		log.Printf("Error adding cronjob: %v", err)
	}

	c.Start()
}

func ReportGeneratorScheduler() {
	c := cron.New()
	_, err := c.AddFunc("@every 30m", func() {
		cursor, err := incidents.ResolvedList()
		if err != nil {
			log.Printf("Error fetching resolved incidents: %v", err)
			return
		}
		if len(cursor) == 0 {
			log.Println("No resolved incidents found")
			return
		}

		var reportslist []reports.Report

		for _, incident := range cursor {
			report := reports.GeneratePDF(incident)
			reportslist = append(reportslist, report)
		}

		var interfaceSlice []interface{}
		for _, r := range reportslist {
			interfaceSlice = append(interfaceSlice, r)
		}

		// Add generated reports to the database
		if _, err := database.InsertMany("reports", interfaceSlice); err != nil {
			log.Printf("Error inserting reports to database: %v", err)
			return
		}

		var bulkWrites []mongo.WriteModel
		for _, incident2 := range cursor {
			filter := bson.M{"id": incident2.Id}
			update := bson.M{"$set": bson.M{"reportCreated": true}}
			model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
			bulkWrites = append(bulkWrites, model)
		}

		if len(bulkWrites) > 0 {
			_, err = database.GetDatabase().Database("IssueReporting").Collection("incidents").BulkWrite(context.Background(), bulkWrites)
			if err != nil {
				log.Println(err)
				return
			}
		}
	})
	if err != nil {
		log.Printf("Error adding cronjob: %v", err)
	}

	c.Start()
}
