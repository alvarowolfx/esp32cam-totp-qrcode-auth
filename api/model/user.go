package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID         string             `json:"id"`
	InternalID primitive.ObjectID `json:"_id" bson:"_id"`
	Email      string             `json:"email"`
	Password   string             `json:"-"`
	Secret     string             `json:"-"`
}
