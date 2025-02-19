package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ShabnamHaque/chatx/backend/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

var client *mongo.Client
var DB *mongo.Database
var Users *mongo.Collection
var Messages *mongo.Collection

func ConnectDB() {
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")
	if mongoURI == "" || dbName == "" {
		log.Fatal("MONGO_URI or DB_NAME is not set in environment variables")
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	var err error
	client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Could not ping MongoDB:", err)
	}
	DB = client.Database(dbName)
	Users = DB.Collection("users")
	Messages = DB.Collection("messages")
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"expires_at": 1},
		Options: options.Index().SetExpireAfterSeconds(0), // TTL Index
	}
	_, err = Messages.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal("Failed to create TTL index:", err)
	}
	fmt.Println("✅ MongoDB connected & TTL index set")
}
func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}

func GetMessages(senderID, receiverID string) ([]models.Message, error) {
	var messages []models.Message
	filter := bson.M{ // Query messages where sender/receiver match
		"$or": []bson.M{
			{"sender_id": senderID, "receiver_id": receiverID},
			{"sender_id": receiverID, "receiver_id": senderID},
		},
	}
	cursor, err := Messages.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &messages); err != nil {
		return nil, err //matching results stored in messages
	}
	return messages, nil
}
func MarkMessagesAsRead(receiverID, senderID string) error {
	filter := bson.M{
		"sender_id":   senderID,
		"receiver_id": receiverID,
		"unread":      true,
	}
	update := bson.M{"$set": bson.M{"unread": false}}
	_, err := Messages.UpdateMany(context.TODO(), filter, update)
	return err
}

func SaveMessageToDB(message models.Message) error {
	doc := bson.M{
		"sender_id":    message.SenderID,
		"receiver_id":  message.ReceiverID,
		"content":      message.Content,
		"timestamp":    message.Timestamp,
		"disappearing": message.Disappearing,
		"unread":       true,
	}
	if message.Disappearing {
		doc["expires_at"] = time.Now().Add(24 * time.Hour) // Message expires in 24 hours
	}
	_, err := Messages.InsertOne(context.TODO(), doc)
	return err
}
func DisconnectDB() {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := client.Disconnect(ctx)
		if err != nil {
			log.Println("Error disconnecting MongoDB:", err)
		} else {
			log.Println("Disconnected from MongoDB")
		}
	}
}
