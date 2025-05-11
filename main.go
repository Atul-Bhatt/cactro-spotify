package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, World!, %s, %s", clientId, clientSecret)
	})
	r.Run() // defaults to ":8080"
}
