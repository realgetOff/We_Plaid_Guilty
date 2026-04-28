package server

import (
	"fmt"
	// "log"
	"net/http"
	// "os"
	"strings"
	// "time"

	"github.com/gin-gonic/gin"
	// "golang.org/x/oauth2"

	// "github.com/joho/godotenv"
	"github.com/realgetOff/We_Plaid_Guilty/internal/config"
	"github.com/realgetOff/We_Plaid_Guilty/internal/webutil"
	// "github.com/realgetOff/We_Plaid_Guilty/internal/config"
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

	// secretsFile := "/vault/secrets/app/config"
// 
	// maxRetries := 10
	// for i := 0; i < maxRetries; i++ {
		// if _, err := os.Stat(secretsFile); err == nil {
			// break
		// }
		// log.Printf("Waiting for secrets file %s... (%d/%d)", secretsFile, i+1, maxRetries)
		// time.Sleep(2 * time.Second)
	// }
// 
	// if err := godotenv.Load(secretsFile); err != nil {
		// return
	// }
// 
	// redirectUrl := os.Getenv("REDIRECT_URL_42")
	// clientId := os.Getenv("CLIENT_ID")
	// clientSecret := os.Getenv("CLIENT_SECRET")
	// authUrl := os.Getenv("AUTH_URL")
	// tokenUrl := os.Getenv("TOKEN_URL")

	// FortyTwoOauthConfig := &oauth2.Config {
		// RedirectURL: redirectUrl,
		// ClientID: clientId,
		// ClientSecret: clientSecret,
		// Scopes: []string{"public"},
		// Endpoint:	oauth2.Endpoint {
			// AuthURL: authUrl,
			// TokenURL: tokenUrl,
		// },
	// }
	// this should be turned into a randomly generated string
	// OauthStateString := "pseudo-random-state"

	// NEW LOGIN CODE
	serverVars.router.GET("/api/auth/42/url", func (c *gin.Context){
		fmt.Println("ATTEMPTING TO GET LOGIN/42/URL FROM ROUTER")

		config.DEBUGgetalloauthvars()
		
		url := config.FortyTwoOauth.AuthCodeURL(config.OauthStateString)
		c.JSON(http.StatusOK, gin.H{"url": url})
	})

	// CALLBACK FOR OAUTH2 WITH 42API

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
