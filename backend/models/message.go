package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	SenderID     primitive.ObjectID `json:"sender_id" bson:"sender_id"`
	ReceiverID   primitive.ObjectID `json:"receiver_id,omitempty" bson:"receiver_id,omitempty"` // optional for direct messages
	GroupID      string             `json:"group_id,omitempty" bson:"group_id,omitempty"`       // optional for group messages
	Content      string             `json:"content" bson:"content"`
	Disappearing bool               `json:"disappearing,omitempty" bson:"disappearing,omitempty"`
	Timestamp    time.Time          `json:"timestamp" bson:"timestamp"`
	ExpiresAt    time.Time          `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	Unread       bool               `bson:"unread"`
	IsGroup      bool               `json:"is_group" bson:"is_group"`
}
