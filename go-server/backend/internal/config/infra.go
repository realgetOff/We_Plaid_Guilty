package config

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"golang.org/x/oauth2"
)

type Config struct {
	APIKey		string `json:"api_key"`
	DBPassword	string `json:"db_password"`
	JWTSecret	string `json:"jwt_secret"`
}

var config Config

func vaultstatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"api_key_loaded":		config.APIKey != "",
		"db_password_loaded":	config.DBPassword != "",
		"jwt_secret_loaded":	config.JWTSecret != "",
	})
}

// https://api.intra.42.fr/apidoc/guides/web_application_flow#exchange-your-code-for-an-access-token
// https://pkg.go.dev/golang.org/x/oauth2#Endpoint

var (
	redirectUrl = os.Getenv("REDIRECT_URL_42")
	clientId = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	authUrl = os.Getenv("AUTH_URL")
	tokenUrl = os.Getenv("TOKEN_URL")

	FortyTwoOauthConfig = &oauth2.Config {
		RedirectURL: redirectUrl,
		ClientID: clientId,
		ClientSecret: clientSecret,
		Scopes: []string{"public"},
		Endpoint:	oauth2.Endpoint {
			AuthURL: authUrl,
			TokenURL: tokenUrl,
		},
	}
	// this should be turned into a randomly generated string
	OauthStateString = "pseudo-random-state"
)


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

	config.APIKey	 = os.Getenv("API_KEY")
	config.DBPassword = os.Getenv("DB_PASSWORD")
	config.JWTSecret  = os.Getenv("JWT_SECRET")

	log.Println("Secrets loaded from Vault Agent Injector successfully")
	return nil
}

func Addnewlinestotls() []byte {
	content, err := os.ReadFile("/vault/secrets/tls")
	if err != nil {
		return nil
	}
	delimiter := "-----END CERTIFICATE-----"
	replacement := delimiter + "\n"
	return []byte(strings.ReplaceAll(string(content), delimiter, replacement))
}
