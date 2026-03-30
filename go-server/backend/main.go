package main

import (
	//"encoding/json"
	"context"
	"log"
	"os"
	"fmt"
	"main.go/gamemanager"
	// following two are for lobby generation
	//"math/rand/v2"
	// "sync"

	"github.com/gin-gonic/gin"
	//"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var globalAIHub *gamemanager.AIHub

func connectToDatabase () (*pgxpool.Pool, error) {

	db_host := os.Getenv("DB_HOST")
	db_port := os.Getenv("DB_PORT")
	db_user := os.Getenv("DB_USER")
	db_password := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")

	connection_url := "postgres://" + db_user + ":" + db_password + "@" + db_host + ":" + db_port + "/" + db_name

	fmt.Println("Attempting to connect to :" + connection_url)

	db, err := pgxpool.New(context.Background(), connection_url)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connection to PostgreSQL database successful")

	return db, nil
}


type serverVarsStruct struct { // the name is temporary
	globalHub *gamemanager.Hub
	globalAIHub *gamemanager.AIHub
	router *gin.Engine
	db *pgxpool.Pool
}

func NewServerStructure () *serverVarsStruct {
	
	// Try to connect to the database, fatally exit if we can't reach it
	dbPool, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Couldn't connect to the PostgreSQL database: %v", err)
	}
	hub := &gamemanager.Hub{
		Rooms: make(map[string]*gamemanager.Room),
	}
	AIHub := &gamemanager.AIHub{
    	Rooms: make(map[string]*gamemanager.AIRoom),
	}
	r := gin.Default();

	return &serverVarsStruct{
		globalHub:		hub,
		globalAIHub:	AIHub,
		router:			r,
		db:				dbPool,
	}
}


func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")

	serverVars := NewServerStructure()

	defer serverVars.db.Close()

	// if err := loadSecretsFromVault(); err != nil {
	// 	log.Fatalf("Failed to load secrets from Vault: %v", err)
	// }
	// Gin router with default "middleware"
	
	// gin.SetMode(gin.ReleaseMode)
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies 

	serverVars.router.Static("/assets", "./static/assets")
	serverVars.router.StaticFile("/favicon.ico", "./static/favicon.ico")
	serverVars.router.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	serverVars.router.GET("/api/ai-rooms/:code", func(c *gin.Context) { // smells like AI generated code.
		findRoom(c, serverVars, 1)
	})
	serverVars.router.GET("/api/rooms/:code", func(c *gin.Context) { // smells like AI generated code.
		findRoom(c, serverVars, 0)
	})
	serverVars.router.GET("/ping", pong)
	serverVars.router.GET("/health", health)
	serverVars.router.GET("/api/config", vaultstatus)
	serverVars.router.GET("/ws", func (c *gin.Context){
		handleWebsocket(c, serverVars.db, serverVars.globalHub)
	})
	serverVars.router.POST("/api/auth/player", func (c *gin.Context){
		handleGuestAuth(c, serverVars.db)
	})

	// get the port defined in the environment variables, if theres fuckall, 8080

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
		
	if err := serverVars.router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)	
	}
}
