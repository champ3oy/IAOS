package incidents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"issue-reporting/auth"
	"issue-reporting/database"
	"issue-reporting/schedules"
	"issue-reporting/slack"
	"issue-reporting/users"
	"issue-reporting/utils"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Handlers for incident-related routes

func CreateIncident(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	incident := Incident{
		Severity:     "Low",
		Status:       "Open",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
		Timeline:     []Timepoint{},
	}

	data := map[string]interface{}{
		"createdby": team.Name,
		"subtext":   fmt.Sprintf("Invitiated by %s", team.Name),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
	}

	jsonString := string(jsonData)

	incident.Metadata = jsonString

	incident.TeamId = team.TeamId
	incident.Resolved = false
	incident.Timeline = append(incident.Timeline, Timepoint{
		Title:     "Incident Created",
		CreatedAt: time.Now(),
		Metadata:  jsonString,
	})

	if err := c.BodyParser(&incident); err != nil {
		// Handle parsing error
		log.Println(err)
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
		log.Println(err)
		return err
	}

	incident.Id = code
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

		incident.Timeline = append(incident.Timeline, Timepoint{
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
		"createdby": team.Name,
		"subtext":   fmt.Sprint("Alert sent to everyone on-call and slack"),
	}

	jsonData, err = json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
	}

	jsonString = string(jsonData)

	incident.Timeline = append(incident.Timeline, Timepoint{
		Title:     "Alerted",
		CreatedAt: time.Now(),
		Metadata:  jsonString,
	})

	// Someone is on-call
	_, err = database.InsertOne("incidents", incident)
	if err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "incident not created")
	}

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

// func GetIncidents(c *fiber.Ctx) error {
// 	ctx := context.Background()
// 	email := c.Locals("email").(string)
// 	var team auth.Teams
// 	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
// 	if err != nil {
// 		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
// 	}

// 	filter := bson.M{"teamid": team.TeamId}
// 	cursor, err := database.Find("incidents", filter)
// 	if err != nil {
// 		return fmt.Errorf("error finding incidents: %v", err)
// 	}
// 	defer cursor.Close(ctx)

// 	var incidents []Incident
// 	for cursor.Next(ctx) {
// 		var incident Incident
// 		if err := cursor.Decode(&incident); err != nil {
// 			return fmt.Errorf("error decoding incident: %v", err)
// 		}
// 		incidents = append(incidents, incident)
// 	}

// 	return c.Status(200).JSON(fiber.Map{
// 		"message":   "incidents data",
// 		"incidents": &incidents,
// 	})
// }

func GetIncidents(c *fiber.Ctx) error {
	ctx := context.Background()
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Pagination parameters
	page := 1      // default page number
	pageSize := 10 // default page size
	incType := false

	if pageStr := c.Query("page"); pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		pageSize, _ = strconv.Atoi(pageSizeStr)
	}
	if incTypeStr := c.Query("incType"); incTypeStr != "" {
		incType, _ = strconv.ParseBool(incTypeStr)
	}

	// MongoDB filter
	filter := bson.M{"teamid": team.TeamId, "resolved": incType}

	// MongoDB options for sorting
	sortOptions := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}) // Sort by _id field in descending order

	// MongoDB options for pagination
	paginationOptions := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := database.GetDatabase().Database("IssueReporting").Collection("incidents").Find(ctx, filter, sortOptions, paginationOptions)
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
		"incidents": incidents,
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

func Resolve(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}
	incidentCode := c.Params("incidentId")

	data := map[string]interface{}{
		"resolvedBy": team,
		"subtext":    "Incident has been resolved",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
	}

	jsonString := string(jsonData)

	timepoint := Timepoint{
		Title:     "Resolved ✅",
		CreatedAt: time.Now(),
		Metadata:  jsonString,
	}

	// Build the filter to find the incident by their code
	filter := bson.M{"id": incidentCode}
	update := bson.M{"$set": bson.M{"resolved": true, "resolvedat": time.Now()}, "$push": bson.M{"timeline": timepoint}}

	// Perform the update operation
	var incident Incident
	err = database.FindOneAndUpdate("incidents", filter, update).Decode(&incident)
	if err != nil {
		return fiber.NewError(fiber.StatusNoContent, err.Error())
	}

	// Return the updated incident as response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "incident resolved",
		"resolved": true,
	})
}
func Acknowledge(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}
	incidentCode := c.Params("incidentId")

	data := map[string]interface{}{
		"acknowledBy": team,
		"subtext":     "Incident has been acknowledged",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
	}

	jsonString := string(jsonData)

	timepoint := Timepoint{
		Title:     "Acknowledged 👍🏼",
		CreatedAt: time.Now(),
		Metadata:  jsonString,
	}

	// Build the filter to find the incident by their code
	filter := bson.M{"id": incidentCode, "assignedto.FirstName": bson.M{"$ne": ""}}
	update := bson.M{"$set": bson.M{"acknowledged": true, "acknowledgedat": time.Now()}, "$push": bson.M{"timeline": timepoint}}

	// Perform the update operation
	var incident Incident
	err = database.FindOneAndUpdate("incidents", filter, update).Decode(&incident)
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

func ResolvedList() ([]Incident, error) {
	ctx := context.Background()
	filter := bson.M{"resolved": true, "reportCreated": false}

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
	data := map[string]interface{}{
		"assignedTo": params.User,
		"subtext":    fmt.Sprintf("Assigned to: %s", params.User),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
	}

	jsonString := string(jsonData)

	timepoint := Timepoint{
		Title:     "Incident",
		CreatedAt: time.Now(),
		Metadata:  jsonString,
	}
	update := bson.M{"$push": bson.M{"assignedto": params.User, "timeline": timepoint}}

	var incident Incident
	err = database.FindOneAndUpdate("incidents", filter, update).Decode(&incident)
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
