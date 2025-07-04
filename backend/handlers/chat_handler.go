package handlers

import (
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
		log.Printf("🔴WebSocket disconnected|senderId :%s", sender_id) //executing
	})
}

func HandleWebSocket(c *gin.Context) {
	// Extract user ID from query parameters
	tokenStr := c.Query("token")
	_, err := utils.ValidateJWT(tokenStr)
	if err != nil {
		log.Println("❌ Invalid WebSocket token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	senderID := c.Query("sender_id")
	if senderID == "" {
		log.Println("⚠️ Missing senderID in WebSocket connection")
		c.JSON(http.StatusBadRequest, gin.H{"error": "senderID is required"})
		return
	}

	err = m.HandleRequest(c.Writer, c.Request)
	if err != nil {
		log.Println("❌ WebSocket connection error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket connection failed"})
		return
	}
	log.Printf("✅ WebSocket request handled successfully for sender:%s", senderID) //executing
}
func InitMessageHandler(c *gin.Context) {
	token := c.GetHeader("Authorization")
	claims, err := utils.ValidateJWT(token)
	if err != nil {
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

	users, err := database.GetUnreadUsers(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch unread message users"})
		return
	}

	var userList []bson.M
	for _, u := range users {
		userList = append(userList, bson.M{
			"id":          u.ID.Hex(),
			"username":    u.Username,
			"email":       u.Email,
			"profile_pic": u.ProfilePic,
		})
	}

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
	/*	token, err := utils.ExtractToken(c)
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
	*/
	// senderID, exists := c.Get("UserID")
	// if !exists {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "UserID not found in context"})
	// 	return
	// }
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
	senderID_objID, _ := primitive.ObjectIDFromHex(senderID)
	receiverID_objID, _ := primitive.ObjectIDFromHex(receiverID)

	err := database.MarkMessagesAsRead(senderID_objID, receiverID_objID)
	if err != nil {
		log.Println("⚠️ Failed to update read status:", err)
	}
	log.Printf("Updated read status btn %s and %s", senderID, receiverID)

	messagesSlice, err := database.GetMessages(senderID_objID, receiverID_objID)
	if err != nil {
		log.Println("❌ Failed to fetch chat history: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve messages"})
		return
	}
	log.Printf("📜 Chat history retrieved | Between: %s & %s", senderID, receiverID)
	c.JSON(http.StatusOK, gin.H{"messages": messagesSlice})
}
