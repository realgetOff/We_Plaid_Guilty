package webutil

import (
	"fmt"
	"net/http"
	"context"
	"encoding/json"


	"github.com/gin-gonic/gin"
	"github.com/realgetOff/We_Plaid_Guilty/internal/db"
	"github.com/realgetOff/We_Plaid_Guilty/internal/metrics"
	"github.com/realgetOff/We_Plaid_Guilty/internal/config"
)

func FortyTwoCallback(c *gin.Context, dbs *db.DBSafe) {
	code := c.Query("code")

	config.DEBUGgetalloauthvars()

	token, err := config.FortyTwoOauth.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Exchange failed"})
		return
	}

	client := config.FortyTwoOauth.Client(context.Background(), token)

	resp, err := client.Get("https://api.intra.42.fr/v2/me")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reach 42 API"})
		return
	}
	defer resp.Body.Close()

	var userProfile struct {
		Login string `json:"login"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userProfile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	fmt.Println("\n--- NEW 42API LOGIN DETECTED ---")
	fmt.Printf("Username: %s\n", userProfile.Login)
	fmt.Printf("Email:    %s\n", userProfile.Email)
	fmt.Println("--------------------------\n")

	var userID string
	var isInsert bool

	userQuery :=	`
					INSERT INTO users (username, email, type) 
					VALUES ($1, $2, 'api42') 
					ON CONFLICT (username) 
					DO UPDATE SET 
						email = EXCLUDED.email,
						type = 'api42'
					RETURNING id, (xmax = 0);
					`

	profileQuery :=	`
					INSERT INTO profiles (id, display_name)
					VALUES ($1, $2)
					ON CONFLICT (id) DO NOTHING;
					`

	err = db.DBQuery(dbs, userQuery, []any{userProfile.Login, userProfile.Email}, &userID, &isInsert)
	if err != nil {
		fmt.Printf("User creation failed %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't insert a user in the database."})
		return
	}

	if (isInsert){
		metrics.UserCountTotal.Inc()
		metrics.UserCountAPI.Inc() 
	}

	fmt.Println("Username: " + userProfile.Login + USR_ID + userID)

	if userID != "" {
		err = db.DBQuery(dbs, profileQuery, []any{userID, userProfile.Login})
		if err != nil {
			fmt.Printf("User profile creation failed %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Server couldn't create a user profile in the database."})
			return
		}
	}

	var SignedString string
	SignedString, err = GenerateJWT(userID, userProfile.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": JWT_ERROR})
		return
	}

	frontendRedirectURL := fmt.Sprintf("http://localhost:8080/callback?token=%s", SignedString)
	c.Redirect(http.StatusTemporaryRedirect, frontendRedirectURL)
}
