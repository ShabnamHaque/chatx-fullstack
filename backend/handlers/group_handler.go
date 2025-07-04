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

var GroupCollection *mongo.Collection = database.Groups

func AddMember(c *gin.Context) {
	groupID := c.Query("group_id")
	userID := c.Query("user_id")

	if groupID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id and user_id are required"})
		return
	}

	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	userObjID, err2 := primitive.ObjectIDFromHex(userID)

	if err != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group_id or user_id"})
		return
	}

	filter := bson.M{"_id": groupObjID}
	update := bson.M{"$addToSet": bson.M{"members": userObjID}}

	_, err = GroupCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Member added"})
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

	var group models.Group
	err = GroupCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&group)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": group.MemberIDs})
}
func CreateGroup(c *gin.Context) {
	var group models.Group
	if err := c.BindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group data"})
		return
	}

	group.ID = primitive.NewObjectID()
	group.CreatedAt = time.Now()

	_, err := GroupCollection.InsertOne(context.Background(), group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group created", "group_id": group.ID.Hex()})
}
func GetGroupNameFromID(c *gin.Context) {
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing group ID"})
		return
	}

	// Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}

	// Define a struct to hold result
	var result struct {
		Name string `bson:"name" json:"name"`
	}

	collection := database.GetCollection("groups") // adjust to your setup
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
func GetAllGroupsForUser(c *gin.Context) {
	//
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
