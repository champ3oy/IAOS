package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Teams struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `bson:"name"`
	Email         string             `bson:"email"`
	Password      string             `bson:"password"`
	TeamId        string             `bson:"teamId"`
	Notifications []Notification     `bson:"notifications"`
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
