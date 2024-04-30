package middleware

import (
	"issue-reporting/database"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Team struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	TeamName      string             `bson:"teamName"`
	TeamId        string             `bson:"teamId"`
	Notifications []Notification     `bson:"notifications"`
	APIKey        string             `json:"apiKey"`
}

type Notification struct {
	Channel Channel `bson:"channel"`
	Use     bool    `bson:"use"`
}

type Channel string

const (
	SMS              Channel = "SMS"
	Slack            Channel = "Slack"
	Email            Channel = "Email"
	Call             Channel = "Call"
	Whatsapp         Channel = "Whatsapp"
	PushNotification Channel = "PushNotification"
)

func VerifyAPI() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Missing authentication token"})
		}

		var team Team
		err := database.FindOne("teams", bson.M{"apikey": token}).Decode(&team)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Invalid API key"})
		}

		if team.APIKey != token {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Invalid API key"})
		}

		c.Locals("teamId", team.TeamId)
		return c.Next()
	}
}
