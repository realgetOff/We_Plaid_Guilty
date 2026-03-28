package main

import (
	"os"
	"fmt"
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"context"
	
	vault "github.com/hashicorp/vault/api"
	awsauth "github.com/hashicorp/vault/api/auth/aws"
)

type Config struct {
	APIKey     string `json:"api_key"`
	DBPassword string `json:"db_password"`
	JWTSecret  string `json:"jwt_secret"`
}

var config Config

func vaultstatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"api_key_loaded" : config.APIKey != "",
		"db_password_loaded" : config.DBPassword != "",
		"jwt_secret_loaded" : config.JWTSecret != "",
	})
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

	// Check for local dev/test token first
    if token := os.Getenv("VAULT_TOKEN"); token != "" {
        client.SetToken(token)
        log.Println("Using direct VAULT_TOKEN (Local Dev/Test Mode)")
    } else {
        // Fallback to Production AWS IAM Auth
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
        log.Println("Authenticated via AWS IAM")
    }

    // ... continue to your KV v2 secret reading ...

	// Lecture des secrets KV v2
	kv, err := client.KVv2("secret").Get(context.Background(), "app/config")
	if err != nil {
		return fmt.Errorf("failed to read secrets: %w", err)
	}

	config.APIKey = kv.Data["api_key"].(string)
	config.DBPassword = kv.Data["db_password"].(string)
	config.JWTSecret = kv.Data["jwt_secret"].(string)

	fmt.Println("this is wicked insecure, remove this line before moving further on in prod " + config.APIKey + " " + config.DBPassword + " " + config.JWTSecret )

	log.Println("Secrets loaded from Vault successfully")
	return nil
}
