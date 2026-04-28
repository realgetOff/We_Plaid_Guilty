package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/realgetOff/We_Plaid_Guilty/internal/webutil"
	"github.com/realgetOff/We_Plaid_Guilty/internal/config"

)

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func findRoom(c *gin.Context, serverVars *ServerVarsStruct) {
	code := strings.ToUpper(c.Param("code"))
	room, err := serverVars.globalHub.GetRoom(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Couldn't find the room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    room.GetBase().ID,
		"status":  room.GetBase().Status,
		"players": len(room.GetBase().Players),
	})
}

func Routing(serverVars *ServerVarsStruct) {
	serverVars.router.Static("/assets", "./static/assets")
	serverVars.router.StaticFile("/favicon.ico", "./static/favicon.ico")
	serverVars.router.NoRoute(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.File("./static/index.html")
	})

	serverVars.router.GET("/api/rooms/:code", func(c *gin.Context) {
		findRoom(c, serverVars)
	})
	serverVars.router.GET("/api/ai-rooms/:code", func(c *gin.Context) {
		findRoom(c, serverVars)
	})
	serverVars.router.GET("/ping", ping)
	serverVars.router.GET("/ws", func (c *gin.Context){
		handleWebsocket(c, serverVars)
	})
	serverVars.router.POST("/api/auth/player", func (c *gin.Context){
		webutil.HandleGuestAuth(c, serverVars.db)
	})

	serverVars.router.GET("/api/auth/42/url", func (c *gin.Context){
		fmt.Println("ATTEMPTING TO GET LOGIN/42/URL FROM ROUTER")

		config.DEBUGgetalloauthvars()
		
		url := config.FortyTwoOauth.AuthCodeURL(config.OauthStateString)
		c.JSON(http.StatusOK, gin.H{"url": url})
	})
	serverVars.router.GET("/api/auth/42/callback", func(c *gin.Context){
		fmt.Println("42 CALLBACK URL")
		webutil.FortyTwoCallback(c, serverVars.db)
	})

	serverVars.router.POST("/api/auth/register", func(c *gin.Context){
		webutil.HandleRegister(c, serverVars.db)
	})
	serverVars.router.POST("/api/auth/login", func(c *gin.Context){
		webutil.HandleLogin(c, serverVars.db)
	})
}
