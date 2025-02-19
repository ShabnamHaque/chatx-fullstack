package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username   string               `bson:"username" json:"username" validate:"required,min=3,max=30"`
	Email      string               `bson:"email" json:"email" validate:"required,email"`
	ProfilePic string               `bson:"profile_pic" json:"profile_pic" validate:"required"`
	Password   string               `bson:"password" json:"password" validate:"required,min=6"`
	CreatedAt  time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time            `bson:"updated_at" json:"updated_at"`
	Contacts   []primitive.ObjectID `bson:"contacts" json:"contacts"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}
