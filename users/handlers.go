package users

import (
	"context"
	"fmt"
	"issue-reporting/auth"
	"issue-reporting/database"
	"issue-reporting/utils"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateUser(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	var user User
	if err := c.BodyParser(&user); err != nil {
		// Handle parsing error
		log.Println(err)
		return err
	}

	if len(user.FirstName) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "first name is empty")
	}

	if len(user.LastName) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "last name is empty")
	}

	if len(user.Email) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "email is empty")
	}

	// Check if email already exists
	var existingUser User
	err = database.FindOne("users", bson.M{"email": user.Email}).Decode(&existingUser)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}

	if existingUser.Email == user.Email {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Email already exists",
		})
	}

	code, err := utils.GenerateRandomCode(5)
	if err != nil {
		log.Println(err)
		return err
	}

	user.Code = code
	user.Team = team.TeamId
	user.Role[0] = "Engineer"

	result, err := database.InsertOne("users", user)
	if err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "user not created")
	}

	insertedID := result.InsertedID.(primitive.ObjectID).Hex()

	return c.Status(200).JSON(fiber.Map{
		"message":  "user created",
		"userId":   insertedID,
		"userCode": code,
	})
}

func GetUser(c *fiber.Ctx) error {
	userCode := c.Params("userCode")
	var user User
	err := database.FindOne("users", bson.M{"code": userCode}).Decode(&user)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "No user found")
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "user data",
		"user":    &user,
	})
}

func GetUsers(c *fiber.Ctx) error {
	ctx := context.Background()
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	filter := bson.M{"team": team.TeamId}
	cursor, err := database.Find("users", filter)
	if err != nil {
		return fmt.Errorf("error finding users: %v", err)
	}
	defer cursor.Close(ctx)

	var users []User
	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			return fmt.Errorf("error decoding user: %v", err)
		}
		users = append(users, user)
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "users data",
		"users":   &users,
	})
}

func UpdateUser(c *fiber.Ctx) error {
	// Parse the incoming request body to extract the fields to update
	var userUpdate map[string]interface{}
	if err := c.BodyParser(&userUpdate); err != nil {
		log.Println(err)
		return err
	}

	// Get the user code from the request parameters
	userCode := c.Params("userCode")

	// Build the filter to find the user by their code
	filter := bson.M{"code": userCode}

	// Initialize the update operation
	var update bson.M
	if len(userUpdate) > 0 {
		delete(userUpdate, "team")
		delete(userUpdate, "code")
		// If there are fields in the update request, set them using $set
		update = bson.M{"$set": userUpdate}
	} else {
		// If no fields provided, return an error or handle it as needed
		return fiber.NewError(fiber.StatusBadRequest, "No fields provided for update")
	}

	// Perform the update operation
	var user User
	err := database.FindOneAndUpdate("users", filter, update).Decode(&user)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Return the updated user as response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user data updated",
		"user":    &user,
	})
}

func AssignRole(c *fiber.Ctx) error {
	// Parse the incoming request body to extract the fields to update
	var userUpdate []Role
	if err := c.BodyParser(&userUpdate); err != nil {
		log.Println(err)
		return err
	}

	// Get the user code from the request parameters
	userCode := c.Params("userCode")

	// Build the filter to find the user by their code
	filter := bson.M{"code": userCode}

	// Initialize the update operation
	var update bson.M
	if len(userUpdate) > 0 {
		// If there are fields in the update request, set them using $set
		update = bson.M{"$set": bson.M{"role": userUpdate}}
	} else {
		// If no fields provided, return an error or handle it as needed
		return fiber.NewError(fiber.StatusBadRequest, "No fields provided for update")
	}

	// Perform the update operation
	var user User
	err := database.FindOneAndUpdate("users", filter, update).Decode(&user)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Return the updated user as response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user data updated",
		"user":    &user,
	})
}

func DeleteUser(c *fiber.Ctx) error {
	userCode := c.Params("userCode")
	var user User
	err := database.FindOne("users", bson.M{"code": userCode}).Decode(&user)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "No user found")
	}

	_, err = database.InsertOne("deletedusers", user)
	if err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "user not created")
	}

	_, err = database.DeleteOne("users", bson.M{"code": userCode})
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}
	if err == mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "No user found")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user deleted",
		"user":    &user,
	})
}
