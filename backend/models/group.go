package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	AdminID     primitive.ObjectID   `bson:"admin_id" json:"admin_id"`
	MemberIDs   []primitive.ObjectID `bson:"member_ids,omitempty"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	Description string               `bson:"description,omitempty" json:"description,omitempty"`
	IsPrivate   bool                 `bson:"is_private" json:"is_private"`
}

//private group - only the admin can add people
