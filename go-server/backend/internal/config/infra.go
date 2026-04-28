package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// https://api.intra.42.fr/apidoc/guides/web_application_flow#exchange-your-code-for-an-access-token
// https://pkg.go.dev/golang.org/x/oauth2#Endpoint

var (
	FortyTwoOauth *oauth2.Config
	OauthStateString = "pseudo-random-state"
)

func LoadSecretsIntoOauth() {
	var redirectUrl, clientId, clientSecret, authUrl, tokenUrl string
	myMap, err := godotenv.Read("/vault/secrets/app/config")

	if err == nil {
		clientId = myMap["CLIENT_ID"]
		clientSecret = myMap["CLIENT_SECRET"]
	} else { // LOCAL
		clientId = os.Getenv("CLIENT_ID")
		clientSecret = os.Getenv("CLIENT_SECRET")
	}
	redirectUrl = os.Getenv("REDIRECT_URL_42")
	authUrl = os.Getenv("AUTH_URL")
	tokenUrl = os.Getenv("TOKEN_URL")

	FortyTwoOauth = &oauth2.Config {
		RedirectURL: redirectUrl,
		ClientID: clientId,
		ClientSecret: clientSecret,
		Scopes: []string{"public"},
		Endpoint:	oauth2.Endpoint {
			AuthURL: authUrl,
			TokenURL: tokenUrl,
		},
	}
}

func DEBUGgetalloauthvars () {
	fmt.Println(" -- DEBUG -- ")
	fmt.Printf(`REDIRURL : %v | ID : %v | SECRET : %v | AUTH : %v | TOKEN : %v\n`,
	FortyTwoOauth.RedirectURL, FortyTwoOauth.ClientID, FortyTwoOauth.ClientSecret)
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
