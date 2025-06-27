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
