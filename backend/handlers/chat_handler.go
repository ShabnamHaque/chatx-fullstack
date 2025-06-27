package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ShabnamHaque/chatx/backend/database"
	"github.com/ShabnamHaque/chatx/backend/models"
	"github.com/ShabnamHaque/chatx/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var m *melody.Melody

// InitChatHandler initializes the Melody WebSocket instance
func InitChatHandler() {
	m = melody.New() // Initialize WebSocket server

	m.HandleConnect(func(s *melody.Session) {
		//	receiver_id := s.Request.URL.Query().Get("receiver_id")
		sender_id := s.Request.URL.Query().Get("sender_id") //just store who is connected

		s.Set("sender_id", sender_id) //how the ws conn be uniquely identified.
		log.Printf("🟢 WebSocket connected | Sender: %s", sender_id)
	})

	m.HandleDisconnect(func(s *melody.Session) {
		//receiver_id, _ := s.Get("receiver_id")
		sender_id, _ := s.Get("sender_id")
		log.Printf("🔴WebSocket disconnected|senderId :%s", sender_id)
	})
}

//	func HandleWebSocket(c *gin.Context) {
//		err := m.HandleRequest(c.Writer, c.Request)
//		if err != nil {
//			log.Println("❌ WebSocket connection error:", err)
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket connection failed"})
//			return
//		}
//		log.Println("✅ WebSocket request handled successfully")
//	}

func HandleWebSocket(c *gin.Context) {
	// Extract user ID from query parameters
	senderID := c.Query("sender_id")
	if senderID == "" {
		log.Println("⚠️ Missing senderID in WebSocket connection")
		c.JSON(http.StatusBadRequest, gin.H{"error": "senderID is required"})
		return
	}

	err := m.HandleRequest(c.Writer, c.Request)
	if err != nil {
		log.Println("❌ WebSocket connection error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket connection failed"})
		return
	}
	log.Printf("✅ WebSocket request handled successfully for sender:%s", senderID)
}
func InitMessageHandler(c *gin.Context) {
	token, err := utils.ExtractToken(c)
	if err != nil {
		log.Println("❌ Token extraction failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "redirect": "/login"})
		return
	}
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		log.Println("❌ JWT validation failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "redirect": "/login"})
		return
	}
	var message models.Message
	if err := c.ShouldBindJSON(&message); err != nil {
		log.Println("⚠️ Invalid request body:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	userID_primitive, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error parsing str to objectID"})
		return
	}
	if userID_primitive != message.SenderID {
		log.Printf("⚠️ Unauthorized sender attempt. JWT UserID: %s | SenderID in request: %s", claims.UserID, message.SenderID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized sender"})
		return
	}
	message.Timestamp = time.Now()
	message.Unread = true
	if err := database.SaveMessageToDB(message); err != nil {
		log.Println("❌ Database insertion failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save message"})
		return
	}
	log.Printf("✅ Message saved | Sender: %s → Receiver: %s | Content: %s", message.SenderID, message.ReceiverID, message.Content)

	messageJSON, _ := json.Marshal(message)
	// messageJSON, _ := json.Marshal(map[string]interface{}{
	// 	"_id":          message.ID.Hex(),
	// 	"sender_id":   message.SenderID,
	// 	"receiver_id": message.ReceiverID,
	// 	"content":     message.Content,
	// 	"timestamp":   message.Timestamp,
	// 	"unread":      true, // Ensure it's true when sent
	//	})
	m.BroadcastFilter(messageJSON, func(q *melody.Session) bool {
		userID, exists := q.Get("sender_id") // Retrieve user ID from session
		if !exists {
			return false
		}
		userIDObj, ok := userID.(primitive.ObjectID)
		if !ok {
			return false
		}
		return userIDObj == message.SenderID || userIDObj == message.ReceiverID
	})
	log.Printf("📡 Message broadcasted to Receiver: %s and Sender: %s", message.ReceiverID, message.SenderID)
	c.JSON(http.StatusOK, gin.H{"message": "Message sent successfully", "isSender": true})
}
func GetListOfUsersWithUnreadMessages(c *gin.Context) {
	token := c.GetHeader("Authorization")
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims.UserID // Keep as string

	// Find unread messages where the user is the receiver
	cursor, err := database.Messages.Find(context.TODO(), bson.M{
		"receiver_id": userID, // No conversion needed
		"unread":      true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch unread messages"})
		return
	}
	defer cursor.Close(context.TODO())

	senderMap := make(map[primitive.ObjectID]bool)
	for cursor.Next(context.TODO()) {
		var msg models.Message
		if err := cursor.Decode(&msg); err == nil {
			senderMap[msg.SenderID] = true // Keep sender ID as string
		}
	}

	var unreadSenderIDs []primitive.ObjectID
	for senderID, err := range senderMap {
		if !err {
			unreadSenderIDs = append(unreadSenderIDs, senderID)
		}
	}

	var unreadUsers []models.User
	if len(unreadSenderIDs) > 0 {
		cursor, err := database.Users.Find(context.TODO(), bson.M{"_id": bson.M{"$in": unreadSenderIDs}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user details"})
			return
		}
		defer cursor.Close(context.TODO())

		for cursor.Next(context.TODO()) {
			var user models.User
			if err := cursor.Decode(&user); err == nil {
				unreadUsers = append(unreadUsers, user)
			}
		}
	}

	// Prepare JSON response
	var userList []bson.M
	for _, u := range unreadUsers {
		userList = append(userList, bson.M{
			"id":          u.ID.Hex(), // Convert ObjectID to string
			"username":    u.Username,
			"email":       u.Email,
			"profile_pic": u.ProfilePic,
		})
	}
	log.Println("unread users ", len(userList))
	c.JSON(http.StatusOK, gin.H{"users": userList})
}
func GetGroupChatHistory(c *gin.Context) {
	token, err := utils.ExtractToken(c)
	if err != nil {
		log.Println("❌ Token extraction failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "redirect": "/login"})
		return
	}
	_, err = utils.ValidateJWT(token)
	if err != nil {
		log.Println("❌ JWT validation failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "redirect": "/login"})
		return
	}

	groupIDStr := c.Query("group_id")
	if groupIDStr == "" {
		log.Println("⚠️ Missing group_id in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	groupID, err := primitive.ObjectIDFromHex(groupIDStr)

	if err != nil {
		log.Println("❌ Invalid group_id format:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group_id"})
		return
	}
	messages, err := database.GetGroupMessages(groupID)
	if err != nil {
		log.Println("❌ Failed to fetch group chat history:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve group messages"})
		return
	}

	log.Printf("📜 Group chat history retrieved | GroupID: %s", groupIDStr)
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func GetChatHistory(c *gin.Context) { //log.Println("Inside get chat history...")
	token, err := utils.ExtractToken(c)
	if err != nil {
		log.Println("❌ Token extraction failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "redirect": "/login"})
		return
	}
	_, err = utils.ValidateJWT(token)
	if err != nil {
		log.Println("❌ JWT validation failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "redirect": "/login"})
		return
	}
	senderID := c.Query("sender_id")
	receiverID := c.Query("receiver_id")
	if receiverID == "" {
		log.Println("⚠️ Missing receiverID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "receiver_id is required"})
		return
	}
	if senderID == "" {
		log.Println("⚠️ Missing senderID in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "senderID is required"})
		return
	}
	//Mark messages as read if they were unread
	err = database.MarkMessagesAsRead(senderID, receiverID)
	if err != nil {
		log.Println("⚠️ Failed to update read status:", err)
	}
	log.Printf("Updated read status btn %s and %s", senderID, receiverID)

	senderID_objID, _ := primitive.ObjectIDFromHex(senderID)
	receiverID_objID, _ := primitive.ObjectIDFromHex(receiverID)

	messagesSlice, err := database.GetMessages(senderID_objID, receiverID_objID)
	if err != nil {
		log.Println("❌ Failed to fetch chat history: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve messages"})
		return
	}
	log.Printf("📜 Chat history retrieved | Between: %s & %s", senderID, receiverID)
	c.JSON(http.StatusOK, gin.H{"messages": messagesSlice})
}
