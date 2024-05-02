package incidents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"issue-reporting/auth"
	"issue-reporting/database"
	"issue-reporting/notification"
	"issue-reporting/schedules"
	"issue-reporting/slack"
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
	var user auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Invalid Credentials"})

	}

	// initiate an incident template
	incident := Incident{
		Severity:     "Low",
		Status:       "Open",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
		Timeline:     []Timepoint{},
		AssignedTo:   []auth.User{},
	}

	if err := c.BodyParser(&incident); err != nil {
		// Handle parsing error
		log.Println(err)
		return err
	}

	// create a timeline item for when incident is created
	data := map[string]interface{}{
		"createdby": user.Name,
		"subtext":   fmt.Sprintf("Initiated by %s", user.Name),
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
	incident.TeamId = user.TeamId
	incident.Resolved = false
	incident.Timeline = append(incident.Timeline, Timepoint{
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
	schedule, err := schedules.Scheduled(time.Now(), user.TeamId)
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
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Invalid Credentials"})

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
		"createdby": user.Name,
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

func GetIncident(c *fiber.Ctx) error {
	id := c.Params(("id"))

	var incident Incident
	err := database.FindOne("incidents", bson.M{"id": id}).Decode(&incident)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return c.Status(fiber.StatusExpectationFailed).JSON(fiber.Map{
			"message": "Something went wrong",
			"status":  false,
		})
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
			"message": "No incidents found",
			"status":  false,
		})
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
	var team auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Invalid API key"})

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
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Unauthorized", "message": "error finding incidents"})

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

	var user auth.User
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
	var team auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&team)
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
	var team auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&team)
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
	var team auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&team)
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
		Title:     "Incident Assigned",
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

	if incident.AssignedTo[0].Name == "" {
		return nil, errors.New("no users found")
	}

	if len(incident.AssignedTo) > 0 {
		for _, user := range incident.AssignedTo {
			// assignedToNames[i] = fmt.Sprintf("%s <%s>", user.Name, user.GithubHandle)
			notification.SendNotification(fmt.Sprintf("You have been assigned to: \nIncident #%s\nTitle: %s\nDescription: %s\nSeverity: %s", incident.Id, incident.Title, incident.Description, incident.Severity), user)
		}
	}

	return &incident, nil
}

func GetLogs(c *fiber.Ctx) error {
	ctx := context.Background()
	email := c.Locals("email").(string)
	var team auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Pagination parameters
	page := 1      // default page number
	pageSize := 10 // default page size

	if pageStr := c.Query("page"); pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		pageSize, _ = strconv.Atoi(pageSizeStr)
	}

	// MongoDB filter
	filter := bson.M{"teamid": team.TeamId}

	// MongoDB options for sorting
	sortOptions := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}) // Sort by _id field in descending order

	// MongoDB options for pagination
	paginationOptions := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := database.GetDatabase().Database("IssueReporting").Collection("logs").Find(ctx, filter, sortOptions, paginationOptions)
	if err != nil {
		return fmt.Errorf("error finding logs: %v", err)
	}
	defer cursor.Close(ctx)

	var logs []Log
	for cursor.Next(ctx) {
		var log Log
		if err := cursor.Decode(&log); err != nil {
			return fmt.Errorf("error decoding log: %v", err)
		}
		logs = append(logs, log)
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "logs data",
		"logs":    logs,
	})
}
