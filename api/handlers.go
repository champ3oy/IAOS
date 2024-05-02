package api

import (
	"encoding/json"
	"fmt"
	"issue-reporting/auth"
	"issue-reporting/database"
	"issue-reporting/incidents"
	"issue-reporting/notification"
	"issue-reporting/schedules"
	"issue-reporting/slack"
	"issue-reporting/utils"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateIncident(c *fiber.Ctx) error {
	teamId := c.Locals("teamId").(string)

	var team auth.Team
	err := database.FindOne("teams", bson.M{"teamId": teamId}).Decode(&team)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Invalid API key"})
	}

	// initiate an incident template
	incident := incidents.Incident{
		Severity:     "Low",
		Status:       "Open",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
		Timeline:     []incidents.Timepoint{},
		AssignedTo:   []auth.User{},
	}

	if err := c.BodyParser(&incident); err != nil {
		// Handle parsing error
		log.Println(err)
		return err
	}

	// create a timeline item for when incident is created
	data := map[string]interface{}{
		"createdby": team.TeamName,
		"subtext":   fmt.Sprintf("Initiated by %s", team.TeamName),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
	}
	jsonString := string(jsonData)

	// start populating the incident with main details
	code, err := utils.GenerateRandomCode(6)
	if err != nil {
		log.Println(err)
		return err
	}
	incident.Id = code
	incident.Metadata = jsonString
	incident.TeamId = team.TeamId
	incident.Resolved = false
	incident.Timeline = append(incident.Timeline, incidents.Timepoint{
		Title:     "Incident Created",
		CreatedAt: time.Now(),
		Metadata:  jsonString,
	})
	var text string
	var severity string
	switch incident.Severity {
	case "Low":
		severity = "Low 🟨"
	case "Medium":
		severity = "Medium 🟥"
	case "High":
		severity = "High 🆘"
	default:
		severity = "Unknown"
	}

	// check who is on-call
	schedule, err := schedules.Scheduled(time.Now(), team.TeamId)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{
			"message": "can not get on-call engineer",
			"status":  false,
		})
	}

	var scheduledUser auth.User
	if schedule != nil {
		err := database.FindOne("users", bson.M{"email": schedule.User.Email}).Decode(&scheduledUser)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid credentials",
				"status":  false,
			})
		}
		incident.AssignedTo = append(incident.AssignedTo, scheduledUser)
	}

	if len(incident.AssignedTo) > 0 {
		assignedToNames := make([]string, len(incident.AssignedTo))
		for i, user := range incident.AssignedTo {
			assignedToNames[i] = fmt.Sprintf("%s <%s>", user.Name, user.GithubHandle)
		}
		assignedToList := strings.Join(assignedToNames, ", ")
		data := map[string]interface{}{
			"assignedTo": assignedToList,
			"subtext":    fmt.Sprintf("Assigned to: %s", assignedToList),
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
		}

		jsonString := string(jsonData)
		incident.Metadata = jsonString
		incident.Timeline = append(incident.Timeline, incidents.Timepoint{
			Title:     "Incident Assigned",
			CreatedAt: time.Now(),
			Metadata:  jsonString,
		})
		text = fmt.Sprintf("Incident #%s created and assigned to %s\n\n%s\n%s\nSeverity: %s", incident.Id, assignedToList, incident.Title, incident.Description, severity)
	} else {
		text = fmt.Sprintf("Incident #%s created and unassigned\n\n%s\n%s\nSeverity: %s", incident.Id, incident.Title, incident.Description, severity)
	}
	err = slack.Notify(&slack.NotifyParams{Text: text})
	if err != nil {
		log.Println(err)
	}

	data = map[string]interface{}{
		"createdby": team.TeamName,
		"subtext":   "Alert sent to everyone on-call and slack",
	}

	jsonData, err = json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
	}

	jsonString = string(jsonData)

	incident.Timeline = append(incident.Timeline, incidents.Timepoint{
		Title:     "Alerted",
		CreatedAt: time.Now(),
		Metadata:  jsonString,
	})

	// Someone is on-call
	_, err = database.InsertOne("incidents", incident)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{
			"message": "incident not created",
			"status":  false,
		})
	}

	if len(incident.AssignedTo) > 0 {
		for _, user := range incident.AssignedTo {
			// assignedToNames[i] = fmt.Sprintf("%s <%s>", user.Name, user.GithubHandle)
			notification.SendNotification(fmt.Sprintf("You have been assigned to: \nIncident #%s\nTitle: %s\nDescription: %s\nSeverity: %s", incident.Id, incident.Title, incident.Description, incident.Severity), user)
		}
	}

	return c.Status(200).JSON(fiber.Map{
		"message":  "incident created",
		"incident": incident.Id,
	})
}

func CreateLog(c *fiber.Ctx) error {
	teamId := c.Locals("teamId").(string)

	var team auth.Team
	err := database.FindOne("teams", bson.M{"teamId": teamId}).Decode(&team)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Invalid API key"})
	}

	// initiate an log template
	logx := incidents.Log{
		CreatedAt: time.Now(),
	}

	if err := c.BodyParser(&logx); err != nil {
		// Handle parsing error
		return err
	}

	// start populating the incident with main details
	code, err := utils.GenerateRandomCode(6)
	if err != nil {
		log.Println(err)
		return err
	}
	logx.Id = code
	logx.TeamId = team.TeamId

	// Someone is on-call
	_, err = database.InsertOne("logs", logx)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{
			"message": "log not created",
			"status":  false,
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "log created",
		"log":     logx.Id,
	})
}
