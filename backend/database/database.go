package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ShabnamHaque/chatx/backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

var client *mongo.Client
var DB *mongo.Database
var Users *mongo.Collection
var Messages *mongo.Collection
var Groups *mongo.Collection

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
	Groups = DB.Collection("groups")
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
func GetMessages(senderID, receiverID primitive.ObjectID) ([]models.Message, error) {
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
func MarkMessagesAsRead(receiverID, senderID primitive.ObjectID) error {
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
		"is_group":     false,
	}
	if message.Disappearing {
		doc["expires_at"] = time.Now().Add(24 * time.Hour) // Message expires in 24 hours
	}
	_, err := Messages.InsertOne(context.TODO(), doc)
	return err
}
func GetGroupMessages(groupID primitive.ObjectID) ([]models.Message, error) {
	filter := bson.M{"group_id": groupID}

	cursor, err := Messages.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	var messages []models.Message
	if err := cursor.All(context.Background(), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
func SaveGroupMessagesToDB(message models.Message) error {
	doc := bson.M{
		"sender_id":    message.SenderID,
		"is_group":     true,
		"group_id":     message.GroupID,
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
func AddMemberToGroup(userID, groupID primitive.ObjectID) error {

	filter := bson.M{"_id": groupID}
	update := bson.M{
		"$addToSet": bson.M{"members": userID},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	res, err := Groups.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.New("group not found")
	}

	return nil
}

// deletecontactbyID
func DeleteContactByID(userID, contactID primitive.ObjectID) error {
	if userID.IsZero() || contactID.IsZero() {
		return errors.New("missing user ID or contact ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": userID}
	update := bson.M{"$pull": bson.M{"contacts": contactID}}

	result, err := Users.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return errors.New("contact not found or already removed")
	}

	return nil
}

func GetUserByID(objID primitive.ObjectID) (models.User, error) {
	var user models.User

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := Users.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, errors.New("user not found")
		}
		return user, err
	}
	return user, nil
}

func GetUnreadUsers(receiverID string) ([]models.User, error) {
	receiverObjID, err := primitive.ObjectIDFromHex(receiverID)
	if err != nil {
		return nil, errors.New("invalid receiver ID format")
	}

	cursor, err := Messages.Find(context.TODO(), bson.M{
		"receiver_id": receiverObjID,
		"unread":      true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	senderMap := make(map[primitive.ObjectID]struct{})
	for cursor.Next(context.TODO()) {
		var msg models.Message
		if err := cursor.Decode(&msg); err == nil {
			senderMap[msg.SenderID] = struct{}{}
		}
	}
	var senderIDs []primitive.ObjectID
	for id := range senderMap {
		senderIDs = append(senderIDs, id)
	}
	if len(senderIDs) == 0 {
		return []models.User{}, nil
	}
	userCursor, err := Users.Find(context.TODO(), bson.M{
		"_id": bson.M{"$in": senderIDs},
	})
	if err != nil {
		return nil, err
	}
	defer userCursor.Close(context.TODO())

	var users []models.User
	for userCursor.Next(context.TODO()) {
		var user models.User
		if err := userCursor.Decode(&user); err == nil {
			users = append(users, user)
		}
	}
	return users, nil
}

func GetUsersByIDs(userIDs []primitive.ObjectID) ([]models.User, error) {
	if len(userIDs) == 0 {
		return []models.User{}, nil
	}

	cursor, err := Users.Find(context.TODO(), bson.M{
		"_id": bson.M{"$in": userIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var users []models.User
	for cursor.Next(context.TODO()) {
		var user models.User
		if err := cursor.Decode(&user); err == nil {
			users = append(users, user)
		}
	}

	return users, nil
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
