package handlers

import (
	"context"
	"net/http"
	"time"

	database "github.com/ShabnamHaque/chatx/backend/database"
	models "github.com/ShabnamHaque/chatx/backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var RoomsCollection *mongo.Collection = database.Rooms

func AddMemberToRoom(c *gin.Context) {
	groupID := c.Query("group_id")

	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id are required"})
		return
	}

	groupObjID, err := primitive.ObjectIDFromHex(groupID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group_id or user_id"})
		return
	}
	var body struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	if body.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	user, err := database.GetUserByEmail(body.Email) // user_email contained in body
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User to Add not found"})
		return
	}
	if err := database.AddMemberToRoom(user.ID, groupObjID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Member added", "user_email": body.Email, "group_id": groupID})
}
func GetMembers(c *gin.Context) {
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group_id"})
		return
	}

	var group models.Room
	err = RoomsCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&group)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": group.MemberIDs})
}
func CreateRoom(c *gin.Context) {
	var req models.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	userIDRaw, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "UserID missing from context"})
		return
	}

	userID, ok := userIDRaw.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid UserID type"})
		return
	}

	roomID := primitive.NewObjectID()

	group := models.Room{
		ID:          roomID,
		Name:        req.Name,
		Description: req.Description,
		AdminID:     userID,
		MemberIDs:   []primitive.ObjectID{}, // optionally empty, let AddMember handle it
		CreatedAt:   time.Now(),
	}

	_, err := database.Rooms.InsertOne(context.Background(), group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group", "details": err.Error()})
		return
	}

	if err := database.AddMemberToRoom(userID, roomID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add creator to group", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group_id": roomID.Hex()})
}
func GetRoomNameFromID(c *gin.Context) {
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing group ID"})
		return
	}
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}
	var result struct {
		Name string `bson:"name" json:"name"`
	}
	collection := database.GetCollection("rooms") // adjust to your setup
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"group_name": result.Name})
}
func GetAllRoomsForUser(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id"})
		return
	}

	// Parse to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	// MongoDB collection
	collection := database.GetCollection("groups")

	// Filter: groups where user is a member
	filter := bson.M{"members": objID}

	// Optional: define what fields to return
	projection := bson.M{"name": 1, "description": 1}

	cursor, err := collection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}
	defer cursor.Close(context.TODO())

	// Struct for response
	var groups []bson.M
	if err := cursor.All(context.TODO(), &groups); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing groups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}
