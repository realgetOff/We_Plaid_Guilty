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
	//"math/rand/v2"
	"time"
	// "sync"

	"github.com/gin-gonic/gin"
	//"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"

	// TODO: remove when Vault Agent Injector is fully configured
	// vault "github.com/hashicorp/vault/api"
	// awsauth "github.com/hashicorp/vault/api/auth/aws"
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
	lobbyCode := generateLobbyCode(6)
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
		"api_key_loaded":     config.APIKey != "",
		"db_password_loaded": config.DBPassword != "",
		"jwt_secret_loaded":  config.JWTSecret != "",
	})
}

// loadSecrets reads secrets from /vault/secrets/app-secrets
// injected by Vault Agent Injector sidecar.
// Retries until the file is available (agent-init-first handles ordering
// but we keep retries as safety net).
func loadSecrets() error {
	secretsFile := "/vault/secrets/app-secrets"

	// Wait for Vault Agent to write the secrets file
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		if _, err := os.Stat(secretsFile); err == nil {
			break
		}
		log.Printf("Waiting for secrets file %s... (%d/%d)", secretsFile, i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	if err := godotenv.Load(secretsFile); err != nil {
		return fmt.Errorf("failed to load secrets from %s: %w", secretsFile, err)
	}

	config.APIKey     = os.Getenv("API_KEY")
	config.DBPassword = os.Getenv("DB_PASSWORD")
	config.JWTSecret  = os.Getenv("JWT_SECRET")

	log.Println("Secrets loaded from Vault Agent Injector successfully")
	return nil
}

// TODO: remove when Vault Agent Injector is fully configured
// func loadSecretsFromVault() error {
// 	vaultAddr := os.Getenv("VAULT_ADDR")
// 	if vaultAddr == "" {
// 		vaultAddr = "http://vault:8200"
// 	}
// 
// 	cfg := vault.DefaultConfig()
// 	cfg.Address = vaultAddr
// 
// 	client, err := vault.NewClient(cfg)
// 	if err != nil {
// 		return fmt.Errorf("unable to create vault client: %w", err)
// 	}
// 
// 	if token := os.Getenv("VAULT_TOKEN"); token != "" {
// 		client.SetToken(token)
// 		log.Println("Using direct VAULT_TOKEN (Local Dev/Test Mode)")
// 	} else {
// 		awsAuth, err := awsauth.NewAWSAuth(
// 			awsauth.WithRole("app-role"),
// 		)
// 		if err != nil {
// 			return fmt.Errorf("unable to create AWS auth: %w", err)
// 		}
// 		authInfo, err := client.Auth().Login(context.Background(), awsAuth)
// 		if err != nil {
// 			return fmt.Errorf("vault login failed: %w", err)
// 		}
// 		if authInfo == nil {
// 			return fmt.Errorf("no auth info returned")
// 		}
// 		log.Println("Authenticated via AWS IAM")
// 	}
// 
// 	kv, err := client.KVv2("secret").Get(context.Background(), "app/config")
// 	if err != nil {
// 		return fmt.Errorf("failed to read secrets: %w", err)
// 	}
// 
// 	config.APIKey     = kv.Data["api_key"].(string)
// 	config.DBPassword = kv.Data["db_password"].(string)
// 	config.JWTSecret  = kv.Data["jwt_secret"].(string)
// 
// 	fmt.Println("this is wicked insecure, remove this line before moving further on in prod " + config.APIKey + " " + config.DBPassword + " " + config.JWTSecret)
// 
// 	log.Println("Secrets loaded from Vault successfully")
// 	return nil
// }

func main() {
	fmt.Println("~o~ This project was brought to you with hate by pmilner- mforest- namichel & lviravon! ~o~")
	fmt.Println(" ~~ Starting transcendence backend... ~~")

	if err := loadSecrets(); err != nil {
		log.Fatalf("Failed to load secrets: %v", err)
	}

	router := gin.Default()

	router.Static("/assets", "./static/assets")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	router.GET("/ping", pong)
	router.GET("/health", health)
	router.GET("/api/config", vaultstatus)
	router.POST("/api/rooms", createLobby)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

// unused imports kept for reference — remove with commented code above
var _ = context.Background
