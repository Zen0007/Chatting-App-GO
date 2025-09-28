package main

import (
	"fmt"
	"log"
	"main/endpoint"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	envErr := godotenv.Load(".env")
	if envErr != nil {
		log.Println("Error loading .env file")
	}
	dbLink := os.Getenv("DATABASE_LINK")
	fmt.Println("db Link", dbLink)

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"POST", "GET"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	r.Use(cors.New(config))
	endpoint.AuthRouters(r)
	endpoint.HandlerRouter(r)
	endpoint.ProtectedRouters(r)

	r.Run(":8080")
}

/*
	connection "ws://localhost:8080/ws"

*/
