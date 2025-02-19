package routes

import (
	"github.com/ShabnamHaque/chatx/backend/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	handlers.InitChatHandler()
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", handlers.RegisterHandler)
		authGroup.POST("/login", handlers.LoginHandler)
	}
	chatGroup := router.Group("/api/chat")
	{
		chatGroup.GET("/unread-users", handlers.GetListOfUsersWithUnreadMessages) // the new unread messages tab
		chatGroup.GET("/ws", handlers.HandleWebSocket)                            // WebSocket session is established. - query param
		chatGroup.POST("/send", handlers.InitMessageHandler)                      // alt to send texts - also for msg history retrieval
		chatGroup.GET("/history", handlers.GetChatHistory)                        // Fetch chat history - query param
		chatGroup.POST("/contacts", handlers.AddContact)                          // add new contact
		chatGroup.GET("/contacts", handlers.GetContacts)                          // get all contacts
		chatGroup.DELETE("/contacts", handlers.DeleteContactHandler)              // delete a contact - query param
		chatGroup.GET("/user", handlers.GetUserDetails)                           // get user details - query param
	}
}
