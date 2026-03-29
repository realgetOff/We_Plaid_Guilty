package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"main.go/gamemanager"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var globalHub *gamemanager.Hub

func connectToDatabase() (*pgxpool.Pool, error) {
	db_host     := os.Getenv("DB_HOST")
	db_port     := os.Getenv("DB_PORT")
	db_user     := os.Getenv("DB_USER")
	db_password := os.Getenv("DB_PASSWORD")
	db_name     := os.Getenv("DB_NAME")

	connection_url := "postgres://" + db_user + ":" + db_password + "@" + db_host + ":" + db_port + "/" + db_name

	fmt.Println("Attempting to connect to :" + connection_url)

	db, err := pgxpool.New(context.Background(), connection_url)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connection to PostgreSQL database successful")
	return db, nil
}

func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")

	// Load secrets from Vault Agent Injector before connecting to DB
	if err := loadSecrets(); err != nil {
		log.Fatalf("Failed to load secrets: %v", err)
	}

	db, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Couldn't connect to the PostgreSQL database: %v", err)
	}
	defer db.Close()

	router := gin.Default()

	router.Static("/assets", "./static/assets")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	router.GET("/ping", pong)
	router.GET("/health", health)
	router.GET("/api/config", vaultstatus)
	router.GET("/ws", func(c *gin.Context) {
		handleWebsocket(c, db, globalHub)
	})
	router.POST("/api/player", handleGuestAuth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
