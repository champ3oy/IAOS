package users

import (
	"context"
	"fmt"
	"issue-reporting/auth"
	"issue-reporting/database"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUser(c *fiber.Ctx) error {
	userCode := c.Params("userCode")
	var user auth.User
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
	var team auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	filter := bson.M{"teamId": team.TeamId}
	cursor, err := database.Find("users", filter)
	if err != nil {
		return fmt.Errorf("error finding users: %v", err)
	}
	defer cursor.Close(ctx)

	var users []auth.User
	for cursor.Next(ctx) {
		var user auth.User
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

func GetCurrentUser(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	var user auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&user)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "users data",
		"user":    &user,
	})
}

func GetTeam(c *fiber.Ctx) error {
	email := c.Locals("email").(string)
	log.Println(email)
	var user auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&user)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	var team auth.Team
	err = database.FindOne("teams", bson.M{"teamId": user.TeamId}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "teams data",
		"team":    &team,
	})
}

func UpdateUser(c *fiber.Ctx) error {
	var userUpdate map[string]interface{}
	if err := c.BodyParser(&userUpdate); err != nil {
		log.Println(err)
		return err
	}

	email := c.Locals("email").(string)
	log.Println(email)
	var user auth.User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&user)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	filter := bson.M{"code": user.Code}

	var update bson.M
	if len(userUpdate) > 0 {
		delete(userUpdate, "teamId")
		delete(userUpdate, "code")
		delete(userUpdate, "password")
		delete(userUpdate, "id")
		delete(userUpdate, "_id")
		delete(userUpdate, "role")
		update = bson.M{"$set": userUpdate}
	} else {
		return fiber.NewError(fiber.StatusBadRequest, "No fields provided for update")
	}

	err = database.FindOneAndUpdate("users", filter, update).Decode(&user)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user data updated",
		"user":    &user,
	})
}

func AssignRole(c *fiber.Ctx) error {
	var userUpdate []auth.Role
	if err := c.BodyParser(&userUpdate); err != nil {
		log.Println(err)
		return err
	}

	userCode := c.Params("userCode")

	filter := bson.M{"code": userCode}

	var update bson.M
	if len(userUpdate) > 0 {

		update = bson.M{"$set": bson.M{"role": userUpdate}}
	} else {

		return fiber.NewError(fiber.StatusBadRequest, "No fields provided for update")
	}

	var user auth.User
	err := database.FindOneAndUpdate("users", filter, update).Decode(&user)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user data updated",
		"user":    &user,
	})
}

func DeleteUser(c *fiber.Ctx) error {
	userCode := c.Params("userCode")
	var user auth.User
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
