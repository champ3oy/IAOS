package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Body struct {
	Name     string `bson:"name"`
	Email    string `bson:"email"`
	Password string `bson:"password"`
	TeamName string `bson:"teamName"`
}

type JoinTeamBody struct {
	Name     string `bson:"name"`
	Email    string `bson:"email"`
	Password string `bson:"password"`
	TeamId   string `bson:"teamId"`
}

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`
	Email    string             `bson:"email"`
	Password string             `bson:"password"`
	TeamId   string             `bson:"teamId"`
}

type Team struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	TeamName      string             `bson:"teamName"`
	TeamId        string             `bson:"teamId"`
	Notifications []Notification     `bson:"notifications"`
	APIKey        string             `json:"apiKey"`
}

type Notification struct {
	Channel       Channel       `bson:"channel"`
	Priority      int           `bson:"priority"`
	TimeToTrigger time.Duration `bson:"timeToTrigger"`
}

type Channel string

const (
	SMS              Channel = "SMS"
	Slack            Channel = "Slack"
	Email            Channel = "Email"
	Call             Channel = "Call"
	PushNotification Channel = "PushNotification"
)
