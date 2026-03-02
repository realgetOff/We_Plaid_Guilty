package main

import (
	"context"
	//"encoding/json"
	"log"
	"fmt"
	"net/http"
	"os"
	// following two are for lobby generation
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	//"github.com/jackc/pgx/v5"
	vault "github.com/hashicorp/vault/api"
	awsauth "github.com/hashicorp/vault/api/auth/aws"
)

type Config struct {
	APIKey     string `json:"api_key"`
	DBPassword string `json:"db_password"`
	JWTSecret  string `json:"jwt_secret"`
}

type LobbySettings struct {
	Rounds int `json:"rounds"`
	Timer int `json:"timer"`
}

type CreateLobbyRequest struct {
	HostID string `json:"hostId"`
	Settings LobbySettings `json:"settings"`
}

type CreateLobbyResponse struct {
	LobbyCode string `json:"lobbyCode"`
}

var config Config

func generateLobbyCode(length int) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	ret := make([]byte, length)
	for i := range ret {
		ret[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(ret)
}

func createLobby(c *gin.Context) {
	// var req CreateLobbyRequest
// 
	// if err := c.ShouldBindJSON(&req); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		// return
	// }

	lobbyCode := generateLobbyCode(6)
	
	//fmt.Println("The generated lobby code is: " + lobbyCode) //debug command

	c.JSON(http.StatusOK, CreateLobbyResponse{
		LobbyCode: lobbyCode,
	})
}

func pong(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
			"status": "ok",
	})
}

func vaultstatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"api_key_loaded" : config.APIKey != "",
		"db_password_loaded" : config.DBPassword != "",
		"jwt_secret_loaded" : config.JWTSecret != "",
	})
}

func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")

	/*
	if err := loadSecretsFromVault(); err != nil {
		log.Fatalf("Failed to load secrets from Vault: %v", err)
	}*/
	config.APIKey = "dummy_key"
    config.DBPassword = "dummy_password"
    config.JWTSecret = "dummy_secret"

	// Gin router with default "middleware"
	router := gin.Default();
	// gin.SetMode(gin.ReleaseMode)
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies 
	// define a port, ie 443 / 80 so we can connect over https / http

	router.StaticFile("/", "./static/index.html")
	router.Static("/assets", "./static/assets")

//CHANGEMENT POUR FAIRE TOURNER
/*

	router.StaticFile("/", "../ft_transcendance/dist/index.html") // for a single file
	router.Static("/assets", "../ft_transcendance/dist/assets")
*/
	// GET endpoint in the router
	router.GET("/ping", pong)
	router.GET("/health", health)
	router.GET("/api/config", vaultstatus)
	router.POST("/api/rooms", createLobby)

	// get the port defined in the environment variables, if theres fuckall, 8080

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
		
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)	
	}
}

func loadSecretsFromVault() error {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		vaultAddr = "http://vault:8200"
	}

	cfg := vault.DefaultConfig()
	cfg.Address = vaultAddr

	client, err := vault.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("unable to create vault client: %w", err)
	}

	// Auth via AWS IAM
	awsAuth, err := awsauth.NewAWSAuth(
		awsauth.WithRole("app-role"),
	)
	if err != nil {
		return fmt.Errorf("unable to create AWS auth: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), awsAuth)
	if err != nil {
		return fmt.Errorf("vault login failed: %w", err)
	}
	if authInfo == nil {
		return fmt.Errorf("no auth info returned")
	}

	// Lecture des secrets KV v2
	kv, err := client.KVv2("secret").Get(context.Background(), "app/config")
	if err != nil {
		return fmt.Errorf("failed to read secrets: %w", err)
	}

	config.APIKey = kv.Data["api_key"].(string)
	config.DBPassword = kv.Data["db_password"].(string)
	config.JWTSecret = kv.Data["jwt_secret"].(string)

	log.Println("Secrets loaded from Vault successfully")
	return nil
}
