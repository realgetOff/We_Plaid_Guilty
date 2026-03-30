package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	// Legacy Vault direct auth — replaced by Vault Agent Injector
	// Keep for reference only, do not uncomment in production
	// "context"
	// vault "github.com/hashicorp/vault/api"
	// awsauth "github.com/hashicorp/vault/api/auth/aws"
)

type Config struct {
	APIKey     string `json:"api_key"`
	DBPassword string `json:"db_password"`
	JWTSecret  string `json:"jwt_secret"`
}

var config Config

func vaultstatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"api_key_loaded":     config.APIKey != "",
		"db_password_loaded": config.DBPassword != "",
		"jwt_secret_loaded":  config.JWTSecret != "",
	})
}

// loadSecrets reads secrets injected by Vault Agent Injector
// into /vault/secrets/app-secrets and sources them as env vars.
func loadSecrets() error {
	secretsFile := "/vault/secrets/app/config"

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

// Legacy Vault direct auth — replaced by Vault Agent Injector
// Keep for reference only, do not uncomment in production
//
// func loadSecretsFromVault() error {
// 	vaultAddr := os.Getenv("VAULT_ADDR")
// 	if vaultAddr == "" {
// 		vaultAddr = "http://vault:8200"
// 	}
// 	cfg := vault.DefaultConfig()
// 	cfg.Address = vaultAddr
// 	client, err := vault.NewClient(cfg)
// 	if err != nil {
// 		return fmt.Errorf("unable to create vault client: %w", err)
// 	}
// 	if token := os.Getenv("VAULT_TOKEN"); token != "" {
// 		client.SetToken(token)
// 		log.Println("Using direct VAULT_TOKEN (Local Dev/Test Mode)")
// 	} else {
// 		awsAuth, err := awsauth.NewAWSAuth(awsauth.WithRole("app-role"))
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
// 	kv, err := client.KVv2("secret").Get(context.Background(), "app/config")
// 	if err != nil {
// 		return fmt.Errorf("failed to read secrets: %w", err)
// 	}
// 	config.APIKey     = kv.Data["api_key"].(string)
// 	config.DBPassword = kv.Data["db_password"].(string)
// 	config.JWTSecret  = kv.Data["jwt_secret"].(string)
// 	log.Println("Secrets loaded from Vault successfully")
// 	return nil
// }
