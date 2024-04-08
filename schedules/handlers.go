package schedules

import (
	"context"
	"errors"
	"fmt"
	"issue-reporting/database"
	"issue-reporting/users"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateSchedules(c *fiber.Ctx) error {
	var body struct {
		TimeRange TimeRange
	}
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	userCode := c.Params("userCode")

	if body.TimeRange.Start.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "start timestamp in the past",
		})
	}

	err := VerifyTimeRange(body.TimeRange)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "invalid time range",
		})
	}

	existingSchedule, err := ScheduledAt(body.TimeRange.Start.Format(time.RFC3339), body.TimeRange.End.Format(time.RFC3339))
	if err != nil {
		return err
	}
	if existingSchedule != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "schedule already exists within this time range",
		})
	}

	var user users.User
	err = database.FindOne("users", bson.M{"code": userCode}).Decode(&user)
	if err != nil {
		log.Println(err, userCode)
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "no user found",
			})
		}
	}

	schedule := Schedule{User: user, Time: body.TimeRange}

	result, err := database.InsertOne("schedules", schedule)
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "team not created")
	}

	insertedID := result.InsertedID.(primitive.ObjectID).Hex()

	return c.Status(200).JSON(fiber.Map{
		"message":    "schedule created",
		"scheduleId": insertedID,
	})
}

func ScheduledAt(startTimestamp, endTimestamp string) (*Schedule, error) {
	startTime, err := time.Parse(time.RFC3339, startTimestamp)
	if err != nil {
		return nil, errors.New("start timestamp is not in a valid format")
	}

	endTime, err := time.Parse(time.RFC3339, endTimestamp)
	if err != nil {
		return nil, errors.New("end timestamp is not in a valid format")
	}

	overlappingFilter := bson.M{
		"$or": []bson.M{
			bson.M{"time.start": bson.M{"$lt": endTime}, "time.end": bson.M{"$gt": startTime}},
			bson.M{"time.start": bson.M{"$lt": endTime}, "time.end": endTime},
			bson.M{"time.start": startTime, "time.end": bson.M{"$gt": startTime}},
		},
	}

	var schedule Schedule
	err = database.FindOne("schedules", overlappingFilter).Decode(&schedule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No overlapping schedule found
		}
		return nil, err
	}

	return &schedule, nil
}

func SchedulesWithinRange(startTimestamp, endTimestamp string) ([]Schedule, error) {
	ctx := context.Background()
	startTime, err := time.Parse(time.RFC3339, startTimestamp)
	if err != nil {
		return nil, errors.New("start timestamp is not in a valid format")
	}

	endTime, err := time.Parse(time.RFC3339, endTimestamp)
	if err != nil {
		return nil, errors.New("end timestamp is not in a valid format")
	}

	withinRangeFilter := bson.M{
		"$and": []bson.M{
			bson.M{"time.start": bson.M{"$gte": startTime}},
			bson.M{"time.end": bson.M{"$lte": endTime}},
		},
	}

	cursor, err := database.Find("schedules", withinRangeFilter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var schedules []Schedule
	for cursor.Next(ctx) {
		var schedule Schedule
		if err := cursor.Decode(&schedule); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

func Scheduled(timestamp time.Time) (*Schedule, error) {
	filter := bson.M{
		"$or": []bson.M{
			bson.M{"time.start": bson.M{"$lt": timestamp.UTC()}, "time.end": bson.M{"$gt": timestamp.UTC()}},
			bson.M{"time.start": bson.M{"$lt": timestamp.UTC()}, "time.end": timestamp.UTC()},
			bson.M{"time.start": timestamp.UTC(), "time.end": bson.M{"$gt": timestamp.UTC()}},
		},
	}

	var schedule Schedule
	err := database.FindOne("schedules", filter).Decode(&schedule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &schedule, nil
}

func VerifyTimeRange(timeRange TimeRange) error {

	if timeRange.Start.Equal(timeRange.End) {
		return errors.New("start timestamp cannot be equal to end timestamp")
	}

	if timeRange.Start.After(timeRange.End) {
		return errors.New("start timestamp cannot be greater than end timestamp")
	}

	return nil
}

func ScheduledNow() (*Schedule, error) {
	return Scheduled(time.Now())
}

func GetScheduledNow(c *fiber.Ctx) error {
	schedule, err := ScheduledNow()
	if err != nil {
		log.Printf("Error starting cronjob1: %v", err)
	}

	return c.Status(200).JSON(fiber.Map{
		"message":  "schedule data",
		"schedule": schedule,
	})
}

func GetScheduledAt(c *fiber.Ctx) error {
	timestamp := c.Params(("timestamp"))
	parsedtime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "Invalid timestamp")
	}

	schedule, err := Scheduled(parsedtime)
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong getting schedules")
	}

	return c.Status(200).JSON(fiber.Map{
		"message":  "schedule data at " + timestamp,
		"schedule": schedule,
	})
}

func ListByTimeRange(c *fiber.Ctx) error {
	var body struct {
		TimeRange TimeRange
	}
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	schedules, err := SchedulesWithinRange(body.TimeRange.Start.Format(time.RFC3339), body.TimeRange.End.Format(time.RFC3339))
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{
		"message":   "schedule data",
		"schedules": schedules,
	})
}

func DeleteSchedule(c *fiber.Ctx) error {
	id := c.Params("id")
	var schedule Schedule

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	err = database.FindOne("schedules", bson.M{"_id": objID}).Decode(&schedule)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "No schedule found")
	}

	_, err = database.InsertOne("deletedschedules", schedule)
	if err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "schedule not created")
	}

	_, err = database.DeleteOne("schedules", bson.M{"_id": objID})
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "No schedule found")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "schedule deleted",
		"schedule": &schedule,
	})
}

func UpdateSchedules(c *fiber.Ctx) error {
	var scheduleUpdate TimeRange
	if err := c.BodyParser(&scheduleUpdate); err != nil {
		log.Println(err)
		return err
	}

	scheduleCode := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(scheduleCode)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}

	update := bson.M{"$set": scheduleUpdate}

	var schedule Schedule
	err = database.FindOneAndUpdate("schedules", filter, update).Decode(&schedule)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "schedule data updated",
		"schedule": &schedule,
	})
}
