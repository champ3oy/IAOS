package users

type Users struct {
	Items []User
}

type User struct {
	FirstName    string `bson:"firstName"`
	LastName     string `bson:"lastName"`
	SlackHandle  string `bson:"slackHandle"`
	GithubHandle string `bson:"githubHandle"`
	PhoneNumber  string `bson:"phoneNumber"`
	PushToken    string `bson:"pushToken"`
	Role         []Role `bson:"role"`
	Email        string `bson:"email"`
	Code         string `bson:"code"`
	Team         string `bson:"team"`
}

type Role string

const (
	Engineer Role = "Enngineer"
	Lead     Role = "Lead"
	Admin    Role = "Amin"
)
