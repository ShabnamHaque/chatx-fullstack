package main

import (
	"fmt"
	"log"

	"github.com/ShabnamHaque/chatx/backend/database"
	"github.com/ShabnamHaque/chatx/backend/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Check if MONGO_URI is loaded
	// mongoURI := os.Getenv("MONGO_URI")
	// if mongoURI == "" {
	// 	log.Fatal("❌ MONGO_URI is not set")
	// }
	//log.Print(mongoURI)
	database.ConnectDB()
	defer database.DisconnectDB()
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5500", "http://127.0.0.1:5500",
			"https://chatx-01.netlify.app/", "https://shabnamhaque.github.io/chatx-fe/"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	routes.SetupRoutes(router)
	fmt.Println("🚀🚀 Server running on port 8080.. 🚀🚀")
	log.Fatal(router.Run(":8080"))
}
