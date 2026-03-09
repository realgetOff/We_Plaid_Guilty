package main

import (
	//"encoding/json"
	"log"
	"fmt"
	"os"
	// following two are for lobby generation
	//"math/rand/v2"
	// "sync"

	"github.com/gin-gonic/gin"
	//"github.com/jackc/pgx/v5"
)

/*
The message structure contains the json information to be sent / received by the websocket for room generation
type: state before / after generation of the room code
code: room code
omitempty: omits empty strings, lowering network traffic

*/

func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")

	
	// if err := loadSecretsFromVault(); err != nil {
	// 	log.Fatalf("Failed to load secrets from Vault: %v", err)
	// }
	// Gin router with default "middleware"
	router := gin.Default();
	// gin.SetMode(gin.ReleaseMode)
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies 
	// define a port, ie 443 / 80 so we can connect over https / http

	router.Static("/assets", "./static/assets")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})


	// GET endpoint in the router
	router.GET("/ping", pong)
	router.GET("/health", health)
	router.GET("/api/config", vaultstatus)
	router.GET("/ws", handleWebsocket)
	//router.POST("/api/rooms", createLobby)
	//router.POST("/api/player", handleLogin)
	router.POST("/api/player", handleGuestAuth)

	// get the port defined in the environment variables, if theres fuckall, 8080

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
		
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)	
	}
}

