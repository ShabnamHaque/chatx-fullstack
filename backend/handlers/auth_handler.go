package handlers

import (
	"context"
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

func RegisterHandler(c *gin.Context) {
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	usersCollection := database.Users // Get Users collection
	var existingUser models.User
	err := usersCollection.FindOne(context.TODO(), bson.M{"email": newUser.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email is already registered. Go To Login"})
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	hashedPassword, err := utils.HashPassword(newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	newUser.Password = hashedPassword
	newUser.CreatedAt = time.Now()
	newUser.UpdatedAt = time.Now()
	newUser.Contacts = []primitive.ObjectID{}
	_, err = usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving user"})
		return
	}
	token, err := utils.GenerateToken(newUser.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		//"user_id": newUser.ID.Hex(), // Return UserID
		"message": "User registered successfully",
		"token":   token,
	})
}
func LoginHandler(c *gin.Context) {
	var loginData models.LoginRequest
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	usersCollection := database.Users
	var user models.User
	err := usersCollection.FindOne(context.TODO(), bson.M{"email": loginData.Email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if !utils.CheckPassword(loginData.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}
	token, err := utils.GenerateToken(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}
	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}
	_, err = usersCollection.UpdateOne(context.TODO(), bson.M{"email": loginData.Email}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update last login time"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"user_id":       user.ID.Hex(), // Output the UserID
		"token":         token,
		"last_login_at": time.Now(),
		"username":      user.Username,
	})
}
