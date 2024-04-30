package auth

import (
	"fmt"
	"issue-reporting/database"
	"issue-reporting/middleware"
	"issue-reporting/utils"
	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	var existingUser User
	err := database.FindOne("users", bson.M{"email": body.Email}).Decode(&existingUser)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}

	if existingUser.Email == body.Email {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Email already exists",
		})
	}

	if body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Email and password are required",
		})
	}

	// generate unique codes
	code, err := utils.GenerateRandomCode(8)
	if err != nil {
		return err
	}
	teamId, err := utils.GenerateRandomCode(18)
	if err != nil {
		return err
	}
	apiKey, err := utils.GenerateRandomCode(16)
	if err != nil {
		return err
	}
	apiKey2, err := utils.GenerateRandomCode(16)
	if err != nil {
		return err
	}

	// Create user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var user User
	user.Name = body.Name
	user.Email = body.Email
	user.Password = string(hashedPassword)
	user.TeamId = teamId
	user.Code = code

	// create team for user
	var team Team
	team.APIKey = apiKey + "." + apiKey2
	team.TeamName = body.TeamName
	team.TeamId = teamId

	//insert everthing to the database
	result1, err := database.InsertOne("users", user)
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "user not created")
	}
	result, err := database.InsertOne("teams", team)
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "team not created")
	}

	insertedID1 := result1.InsertedID.(primitive.ObjectID).Hex()
	insertedID := result.InsertedID.(primitive.ObjectID).Hex()

	return c.Status(200).JSON(fiber.Map{
		"message": "user created",
		"data": fiber.Map{
			"userId": insertedID1,
			"teamId": insertedID,
		},
	})
}

func UpdateTeam(c *fiber.Ctx) error {
	var teamUpdate []Notification
	if err := c.BodyParser(&teamUpdate); err != nil {
		log.Println(err)
		return err
	}

	email := c.Locals("email").(string)
	var user User
	err := database.FindOne("users", bson.M{"email": email}).Decode(&user)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	filter := bson.M{"teamId": user.TeamId}
	update := bson.M{"$set": bson.M{"notifications": teamUpdate}}

	var team Team
	err = database.FindOneAndUpdate("teams", filter, update).Decode(&team)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "team data updated",
		"team":    &team,
	})
}

func JoinTeam(c *fiber.Ctx) error {
	var body JoinTeamBody
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	var existingUser User
	err := database.FindOne("users", bson.M{"email": body.Email}).Decode(&existingUser)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}

	if existingUser.Email == body.Email {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Email already exists",
		})
	}

	if body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Email and password are required",
		})
	}

	// get team id
	teamId := body.TeamId

	// Create user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var user User
	user.Name = body.Name
	user.Email = body.Email
	user.Password = string(hashedPassword)
	user.TeamId = teamId

	//insert everthing to the database
	result, err := database.InsertOne("users", user)
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "user not created")
	}

	insertedID := result.InsertedID.(primitive.ObjectID).Hex()

	return c.Status(200).JSON(fiber.Map{
		"message": "user created",
		"data": fiber.Map{
			"id": insertedID,
		},
	})
}

func Login(c *fiber.Ctx) error {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&credentials); err != nil {
		return err
	}

	var foundDoc User
	err := database.FindOne("users", bson.M{"email": credentials.Email}).Decode(&foundDoc)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundDoc.Password), []byte(credentials.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	token, err := middleware.GenerateToken(credentials.Email)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "login successfull",
		"token":   token,
	})
}

var secretKey = []byte("your-secret-key")

var tokenstr struct {
	TokenString string `json:tokenString`
}

func Refresh(c *fiber.Ctx) error {
	tokenString := tokenstr
	if err := c.BodyParser(&tokenString); err != nil {
		// Handle parsing error
		log.Println(err)
		return err
	}
	token, _ := jwt.ParseWithClaims(tokenString.TokenString, &middleware.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	// if err != nil {
	// 	log.Println(err, token)
	// 	return err
	// }
	claims, _ := token.Claims.(*middleware.Claims)
	newToken, err := middleware.GenerateToken(claims.Email)
	if err != nil {
		log.Println(err)
		return err
	}
	return c.JSON(fiber.Map{"token": newToken})
}
