package auth

import "go.mongodb.org/mongo-driver/bson/primitive"

type Teams struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`
	Email    string             `bson:"email"`
	Password string             `bson:"password"`
	TeamId   string             `bson:"teamId"`
}
