package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupMember struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	GroupID  primitive.ObjectID `bson:"group_id"`
	UserID   primitive.ObjectID `bson:"user_id"`
	JoinedAt time.Time          `bson:"joined_at"`
}
