package incidents

import (
	"context"
	"errors"
	"fmt"
	"issue-reporting/auth"
	"issue-reporting/database"
	"issue-reporting/schedules"
	"issue-reporting/slack"
	"issue-reporting/users"
	"issue-reporting/utils"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handlers for incident-related routes

func NewIncident() *Incident {
	return &Incident{
		Severity:     "Low",
		Status:       "Open",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
	}
}

func CreateIncident(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	incident := NewIncident()
	incident.TeamId = team.TeamId
	incident.Resolved = false
	incident.Timeline[0] = Timepoint{
		Title:     "Incident Created",
		CreatedAt: time.Now(),
	}
	if err := c.BodyParser(&incident); err != nil {
		// Handle parsing error
		return err
	}

	// check who is on-call
	schedule, err := schedules.ScheduledNow()
	if err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "can not get on-call engineer")
	}

	if schedule != nil {
		incident.AssignedTo = append(incident.AssignedTo, schedule.User)
	}

	code, err := utils.GenerateRandomCode(6)
	if err != nil {
		return err
	}

	incident.Id = code

	// Someone is on-call
	_, err = database.InsertOne("incidents", incident)
	if err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "incident not created")
	}

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
	if len(incident.AssignedTo) > 0 {
		assignedToNames := make([]string, len(incident.AssignedTo))
		for i, user := range incident.AssignedTo {
			assignedToNames[i] = fmt.Sprintf("%s %s <%s>", user.FirstName, user.LastName, user.SlackHandle)
		}
		assignedToList := strings.Join(assignedToNames, ", ")
		incident.Timeline = append(incident.Timeline, Timepoint{
			Title:     "Assign",
			CreatedAt: time.Now(),
			Metadata:  `{"assignedTo": "` + assignedToList + `"}`,
		})
		text = fmt.Sprintf("Incident #%s created and assigned to %s\n\n%s\n%s\nSeverity: %s", incident.Id, assignedToList, incident.Title, incident.Description, severity)
	} else {
		text = fmt.Sprintf("Incident #%s created and unassigned\n\n%s\n%s\nSeverity: %s", incident.Id, incident.Title, incident.Description, severity)
	}
	err = slack.Notify(&slack.NotifyParams{Text: text})
	if err != nil {
		log.Println(err)
	}

	incident.Timeline = append(incident.Timeline, Timepoint{
		Title:     "Alert",
		CreatedAt: time.Now(),
	})

	return c.Status(200).JSON(fiber.Map{
		"message":  "incident created",
		"incident": incident.Id,
	})
}

func GetIncident(c *fiber.Ctx) error {
	id := c.Params(("id"))

	var incident Incident
	err := database.FindOne("incidents", bson.M{"id": id}).Decode(&incident)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusNoContent, "No incidents found")
	}

	return c.Status(200).JSON(fiber.Map{
		"message":  "incident data",
		"incident": &incident,
	})
}

func GetIncidents(c *fiber.Ctx) error {
	ctx := context.Background()
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	filter := bson.M{"teamid": team.TeamId}
	cursor, err := database.Find("incidents", filter)
	if err != nil {
		return fmt.Errorf("error finding incidents: %v", err)
	}
	defer cursor.Close(ctx)

	var incidents []Incident
	for cursor.Next(ctx) {
		var incident Incident
		if err := cursor.Decode(&incident); err != nil {
			return fmt.Errorf("error decoding incident: %v", err)
		}
		incidents = append(incidents, incident)
	}

	return c.Status(200).JSON(fiber.Map{
		"message":   "incidents data",
		"incidents": &incidents,
	})
}

func UpdateIncident(c *fiber.Ctx) error {
	// Parse the incoming request body to extract the fields to update
	var incidentUpdate map[string]interface{}
	if err := c.BodyParser(&incidentUpdate); err != nil {
		log.Println(err)
		return err
	}

	// Get the incident code from the request parameters
	incidentCode := c.Params("id")

	// Build the filter to find the incident by their code
	filter := bson.M{"id": incidentCode}

	// Initialize the update operation
	var update bson.M
	if len(incidentUpdate) > 0 {
		delete(incidentUpdate, "id")
		delete(incidentUpdate, "created_at")
		delete(incidentUpdate, "teamId")
		delete(incidentUpdate, "acknowledged_at")
		delete(incidentUpdate, "assigned_to")
		update = bson.M{"$set": incidentUpdate}
	} else {
		// If no fields provided, return an error or handle it as needed
		return fiber.NewError(fiber.StatusBadRequest, "No fields provided for update")
	}

	// Perform the update operation
	var incident Incident
	err := database.FindOneAndUpdate("incidents", filter, update).Decode(&incident)
	if err != nil {
		return fiber.NewError(fiber.StatusNoContent, err.Error())
	}

	// Return the updated incident as response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "incident data updated",
		"incident": &incident,
	})
}

func DeleteIncident(c *fiber.Ctx) error {
	incidentCode := c.Params("id")
	var incident Incident
	err := database.FindOne("incidents", bson.M{"id": incidentCode}).Decode(&incident)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "No incident found")
	}

	_, err = database.InsertOne("deletedincidents", incident)
	if err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "incident not created")
	}

	_, err = database.DeleteOne("incidents", bson.M{"id": incidentCode})
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "No incident found")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "incident deleted",
		"incident": &incident,
	})
}

func AssignUser(c *fiber.Ctx) error {
	incidentCode := c.Params("incidentId")
	userId := c.Params("userId")

	var user users.User
	err := database.FindOne("users", bson.M{"code": userId}).Decode(&user)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "No user found")
	}

	// Build the filter to find the incident by their code
	filter := bson.M{"id": incidentCode}
	update := bson.M{"$set": bson.M{"assignedto": user}}

	// Perform the update operation
	var incident Incident
	err = database.FindOneAndUpdate("incidents", filter, update).Decode(&incident)
	if err != nil {
		return fiber.NewError(fiber.StatusNoContent, err.Error())
	}

	// Return the updated incident as response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "user assigned incident",
		"incident": &incident,
	})
}

func Acknowledge(c *fiber.Ctx) error {
	incidentCode := c.Params("incidentId")

	// Build the filter to find the incident by their code
	filter := bson.M{"id": incidentCode, "assignedto.FirstName": bson.M{"$ne": ""}}
	update := bson.M{"$set": bson.M{"acknowledged": true, "acknowledgedat": time.Now()}}

	// Perform the update operation
	var incident Incident
	err := database.FindOneAndUpdate("incidents", filter, update).Decode(&incident)
	if err != nil {
		return fiber.NewError(fiber.StatusNoContent, err.Error())
	}

	// Return the updated incident as response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "incident acknowledged",
		"incident": incident,
	})
}

func AcknowledgeAll(c *fiber.Ctx) error {
	ctx := context.Background()
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	filter := bson.M{"teamid": team.TeamId, "assignedto.firstName": bson.M{"$ne": ""}, "acknowledged": false}
	cursor, err := database.Find("incidents", filter)
	if err != nil {
		return fiber.NewError(fiber.StatusNoContent, err.Error())
	}
	defer cursor.Close(ctx)

	var incidents []Incident
	for cursor.Next(ctx) {
		var incident Incident
		update := bson.M{"$set": bson.M{"acknowledged": true, "acknowledgedat": time.Now()}}
		err := database.FindOneAndUpdate("incidents", bson.M{"id": incident.Id}, update).Decode(&incident)
		if err != nil {
			return fiber.NewError(fiber.StatusNoContent, err.Error())
		}
		incidents = append(incidents, incident)
	}

	return c.Status(200).JSON(fiber.Map{
		"message":   "incidents acknowledgef",
		"incidents": &incidents,
	})
}

func List() ([]Incident, error) {
	ctx := context.Background()
	filter := bson.M{"acknowledged": false}

	cursor, err := database.Find("incidents", filter)
	if err != nil {
		return nil, fmt.Errorf("error finding incidents: %v", err)
	}
	defer cursor.Close(ctx)

	var incidents []Incident
	for cursor.Next(ctx) {
		var incident Incident
		if err := cursor.Decode(&incident); err != nil {
			return nil, fmt.Errorf("error decoding incident: %v", err)
		}
		incidents = append(incidents, incident)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return incidents, nil
}

func Assign(incidentId string, params *AssignParams) (*Incident, error) {
	filter := bson.M{"acknowledged": false, "id": incidentId}
	update := bson.M{"$set": bson.M{"assignedto": params.User}}

	var incident Incident
	err := database.FindOneAndUpdate("incidents", filter, update).Decode(&incident)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	if incident.AssignedTo[0].FirstName == "" {
		return nil, errors.New("no users found")
	}

	if err := slack.Notify(&slack.NotifyParams{
		Text: fmt.Sprintf("Incident #%s is now assigned to %s %s [@%s]\n%s", incident.Id, incident.AssignedTo[0].FirstName, incident.AssignedTo[0].LastName, incident.AssignedTo[0].SlackHandle, incident.Title),
	}); err != nil {
		// Log error or handle it appropriately
		fmt.Println("Error sending Slack notification:", err)
	}

	return &incident, nil
}
