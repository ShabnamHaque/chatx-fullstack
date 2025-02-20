package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	SenderID     string             `json:"sender_id" bson:"sender_id"`
	ReceiverID   string             `json:"receiver_id" bson:"receiver_id"`
	Content      string             `bson:"content"`
	Disappearing bool               `json:"disappearing,omitempty" bson:"disappearing,omitempty"`
	Timestamp    time.Time          `json:"timestamp" bson:"timestamp"`
	ExpiresAt    time.Time          `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	Unread       bool               `bson:"unread"`
}
