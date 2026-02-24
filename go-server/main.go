package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	vault "github.com/hashicorp/vault/api"
	awsauth "github.com/hashicorp/vault/api/auth/aws"
)

type Config struct {
	APIKey     string `json:"api_key"`
	DBPassword string `json:"db_password"`
	JWTSecret  string `json:"jwt_secret"`
}

var config Config

func main() {
	if err := loadSecretsFromVault(); err != nil {
		log.Fatalf("Failed to load secrets from Vault: %v", err)
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/config", configHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"api_key_loaded":     config.APIKey != "",
		"db_password_loaded": config.DBPassword != "",
		"jwt_secret_loaded":  config.JWTSecret != "",
	})
}
