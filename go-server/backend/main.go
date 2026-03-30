package main

import (
	//"encoding/json"
	"context"
	"log"
	"os"
	"fmt"
	"strings"
	"main.go/gamemanager"
	"net/http"
	// following two are for lobby generation
	//"math/rand/v2"
	// "sync"

	"github.com/gin-gonic/gin"
	//"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
The message structure contains the json information to be sent / received by the websocket for room generation
type: state before / after generation of the room code
code: room code
omitempty: omits empty strings, lowering network traffic

*/

var globalHub *gamemanager.Hub
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


func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")

	db, err := connectToDatabase()

	if err != nil {
		log.Fatalf("Couldn't connect to the PostgreSQL database: %v", err)
	}
	defer db.Close()

	globalHub = &gamemanager.Hub{
		Rooms: make(map[string]*gamemanager.Room),
	}

	globalAIHub = &gamemanager.AIHub{
    	Rooms: make(map[string]*gamemanager.AIRoom),
	}

	// if err := loadSecretsFromVault(); err != nil {
	// 	log.Fatalf("Failed to load secrets from Vault: %v", err)
	// }
	// Gin router with default "middleware"
	router := gin.Default();
	// gin.SetMode(gin.ReleaseMode)
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies 

	router.Static("/assets", "./static/assets")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})


	router.GET("/api/ai-rooms/:code", func(c *gin.Context) {
    code := strings.ToUpper(c.Param("code"))
    room, err := globalAIHub.GetRoom(code)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "ai room not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "code":    room.ID,
        "status":  room.Status,
        "players": len(room.Players),
    })
})
	router.GET("/api/rooms/:code", func(c *gin.Context) {
		code := strings.ToUpper(c.Param("code"))
		room, err := globalHub.GetRoom(code)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    room.ID,
			"status":  room.Status,
			"players": len(room.Players),
		})
	})
	router.GET("/ping", pong)
	router.GET("/health", health)
	router.GET("/api/config", vaultstatus)
	router.GET("/ws", func (c *gin.Context){
		handleWebsocket(c, db, globalHub, globalAIHub)
	})
	router.POST("/api/auth/player", func (c *gin.Context){
		handleGuestAuth(c, db)
	})

	// get the port defined in the environment variables, if theres fuckall, 8080

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
		
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)	
	}
}
