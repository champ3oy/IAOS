package auth

import (
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

func RegisterAuthRoutes(app *fiber.App) {
	authRoutes := app.Group("/auth")

	authRoutes.Post("/register", Register)
	authRoutes.Post("/login", Login)
	authRoutes.Post("/refresh", Refresh)
}

func Register(c *fiber.Ctx) error {
	var team Teams
	if err := c.BodyParser(&team); err != nil {
		return err
	}

	var existingTeam Teams
	err := database.FindOne("teams", bson.M{"email": team.Email}).Decode(&existingTeam)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err)
		return fiber.NewError(fiber.StatusExpectationFailed, "Something went wrong")
	}

	if existingTeam.Email == team.Email {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Email already exists",
		})
	}

	if team.Email == "" || team.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Email and password are required",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(team.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	team.Password = string(hashedPassword)
	code, err := utils.GenerateRandomCode(6)
	if err != nil {
		return err
	}
	team.TeamId = code

	result, err := database.InsertOne("teams", team)
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "team not created")
	}

	insertedID := result.InsertedID.(primitive.ObjectID).Hex()

	return c.Status(200).JSON(fiber.Map{
		"message": "team created",
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

	var foundDoc Teams
	err := database.FindOne("teams", bson.M{"email": credentials.Email}).Decode(&foundDoc)
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
