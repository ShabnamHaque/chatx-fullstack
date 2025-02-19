package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ShabnamHaque/chatx/backend/database"
	"github.com/ShabnamHaque/chatx/backend/models"
	"github.com/ShabnamHaque/chatx/backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserDetails(c *gin.Context) {
	receiverID := c.Query("id") // Extract ID from URL
	usersCollection := database.Users
	//log.Print("Receiver ID: ", receiverID)
	receiverObjID, err := primitive.ObjectIDFromHex(receiverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format", "details": err.Error()})
		return
	}
	var user models.User
	err = usersCollection.FindOne(context.TODO(), bson.M{"_id": receiverObjID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
		return
	}
	//log.Print("Receiver name successfully sent: ", user.Username)
	c.JSON(http.StatusOK, gin.H{"receiver": user.Username})
}

// DeleteContactHandler removes a contact from the user's contact list
func DeleteContactHandler(c *gin.Context) {
	contactID := c.Query("contactId")
	if contactID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing contactId"})
		return
	}
	token := c.GetHeader("Authorization")
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}
	contactObjID, err := primitive.ObjectIDFromHex(contactID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID format"})
		return
	}
	if userID.IsZero() || contactObjID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing userId or contactId"})
		return
	}
	filter := bson.M{"_id": userID}
	update := bson.M{"$pull": bson.M{"contacts": contactObjID}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := database.Users
	result, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete contact"})
		return
	}
	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found or already removed"})
		return
	}
	log.Print("deleted successfully ", contactID)
	c.JSON(http.StatusOK, gin.H{"error": nil, "message": "Contact deleted successfully."})
}

func AddContact(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	token := c.GetHeader("Authorization")
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var contactUser models.User
	usersCollection := database.Users
	err = usersCollection.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&contactUser)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if contactUser.ID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot add yourself as a contact"})
		return
	}
	var currentUser models.User //check if added email is registered or not
	err = usersCollection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
		return
	}
	update := bson.M{"$addToSet": bson.M{"contacts": contactUser.ID}} // Prevents duplicates
	_, err = database.Users.UpdateOne(context.TODO(), bson.M{"_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add contact", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contact added successfully"})
}
func GetContacts(c *gin.Context) {
	token := c.GetHeader("Authorization")
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	userID, _ := primitive.ObjectIDFromHex(claims.UserID)
	var user models.User
	err = database.Users.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusOK, gin.H{"contacts": []bson.M{}}) // Return empty list if no contacts found
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	var contacts []models.User
	cursor, err := database.Users.Find(context.TODO(), bson.M{"_id": bson.M{"$in": user.Contacts}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contacts"})
		return
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var contact models.User
		if err := cursor.Decode(&contact); err == nil {
			contacts = append(contacts, contact)
		}
	}
	var contactList []bson.M
	for _, contact := range contacts {
		contactList = append(contactList, bson.M{
			"id":          contact.ID,
			"username":    contact.Username,
			"email":       contact.Email,
			"profile_pic": contact.ProfilePic,
		})
	}
	c.JSON(http.StatusOK, gin.H{"contacts": contactList})
}

/*
	func GetChatHistory(c *gin.Context) {
	    senderID, err := primitive.ObjectIDFromHex(c.Query("sender_id"))
	    if err != nil {
	        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sender ID"})
	        return
	    }

	    receiverID, err := primitive.ObjectIDFromHex(c.Query("receiver_id"))
	    if err != nil {
	        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
	        return
	    }

	    // Pagination parameters
	    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))  // Default page 1
	    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20")) // Default 20 messages per request
	    skip := (page - 1) * limit

	    filter := bson.M{
	        "$or": []bson.M{
	            {"sender_id": senderID, "receiver_id": receiverID},
	            {"sender_id": receiverID, "receiver_id": senderID},
	        },
	    }

	    opts := options.Find().
	        SetSort(bson.M{"timestamp": -1}). // Fetch latest messages first
	        SetLimit(int64(limit)).
	        SetSkip(int64(skip))

	    cursor, err := database.Messages.Find(context.TODO(), filter, opts)
	    if err != nil {
	        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chat history"})
	        return
	    }
	    defer cursor.Close(context.TODO())

	    var messages []bson.M
	    if err = cursor.All(context.TODO(), &messages); err != nil {
	        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing chat history"})
	        return
	    }

	    c.JSON(http.StatusOK, messages)
	}

database.Messages.Indexes().CreateOne(

	context.TODO(),
	mongo.IndexModel{
	    Keys: bson.D{
	        {Key: "sender_id", Value: 1},
	        {Key: "receiver_id", Value: 1},
	        {Key: "timestamp", Value: -1}, // Index for fast sorting
	    },
	},

)
run this in mongodb shell or go driver to create index
*/
