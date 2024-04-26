package auth

import (
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
	ID                     primitive.ObjectID `bson:"_id,omitempty"`
	Name                   string             `bson:"name"`
	Email                  string             `bson:"email"`
	Password               string             `bson:"password"`
	TeamId                 string             `bson:"teamId"`
	SlackHandle            string             `bson:"slackHandle"`
	GithubHandle           string             `bson:"githubHandle"`
	PhoneNumber            string             `bson:"phoneNumber"`
	WhatsappNumber         string             `bson:"whatsappNumber"`
	PushToken              string             `bson:"pushToken"`
	Role                   []Role             `bson:"role"`
	Code                   string             `bson:"code"`
	NotificationType       string             `bson:"notificationType"`
	AcceptPushNotification bool               `bson:"acceptPushNotification"`
}

type Role string

const (
	Engineer Role = "Engineer"
	Lead     Role = "Lead"
	Admin    Role = "Admin"
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
