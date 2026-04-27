package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"main.go/metrics"
	"main.go/handler"
)

type LobbySettings struct {
	Rounds int `json:"rounds"`
}

type playerNameTemp struct {
	PlayerName string `json:"playerName"`
}

type CreateLobbyRequest struct {
	HostID   string        `json:"hostId"`
	Settings LobbySettings `json:"settings"`
}

type CreateLobbyResponse struct {
	LobbyCode string `json:"lobbyCode"`
}

const USR_ID = " user ID = "
const JWT_ERROR = "Couldn't sign / generate JWT for user."

type AuthResponse struct {
	Token string `json:"token"`
}

// FUNCTIONS //

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

func handleGuestAuth(c *gin.Context, dbs *DBSafe) {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	guestName := fmt.Sprintf("guest_%04d", n.Int64())
	var userID string
	var err error

	// SQL QUERIES
	userQuery :=	`
					INSERT INTO users (username, type)
					VALUES ($1, 'guest')
					RETURNING id;
					`
	
	profileQuery := `
					INSERT INTO profiles (id, display_name)
					VALUES ($1, $2)
					`

	err = DBQuery(dbs, userQuery, []any{guestName}, &userID)
	if err != nil {
		fmt.Println("Guest creation failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a guest user in the database."})
		return
	}
	fmt.Println("Guest name: " + guestName + " guest ID = " + userID)
	
	err = DBQuery(dbs, profileQuery, []any{userID, guestName})
	if err != nil {
		fmt.Println("Guest profile creation failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a guest profile in the database."})
		return
	}

	var SignedString string
	SignedString, err = handler.GenerateJWT(userID, guestName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't sign / generate JWT for guest."})
		return
	}

	metrics.UserCountTotal.Inc()
	metrics.UserCountGuest.Inc()

	c.JSON(http.StatusOK, AuthResponse{
		Token: SignedString,
	})
}

func findRoom(c *gin.Context, serverVars *serverVarsStruct) {
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

type loginInfo struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email,omitempty"`
}

func handleRegister(c *gin.Context, dbs *DBSafe) {
	var login loginInfo
	var userID string

	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(login.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't hash password: " + err.Error()})
		return
	}

	// SQL QUERIES
	userQuery :=	`
					INSERT INTO users (username, email, password_hash)
					VALUES ($1, $2, $3)
					RETURNING id;
					`

	profileQuery :=	`
					INSERT INTO profiles (id, display_name)
					VALUES ($1, $2)
					`
	
	err = DBQuery(dbs, userQuery, []any{login.Username, login.Email, bytes}, &userID)
	if err != nil {
		fmt.Println("User registration failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't register a user in the database."})
		return
	}

	fmt.Println("Registered username: " + login.Username + USR_ID + userID)

	err = DBQuery(dbs, profileQuery, []any{userID, login.Username})
	if err != nil {
		fmt.Println("User profile creation failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a user profile in the database."})
		return
	}

	var SignedString string
	SignedString, err = handler.GenerateJWT(userID, login.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": JWT_ERROR})
		return
	}

	metrics.UserCountTotal.Inc()
	metrics.UserCountStandard.Inc()

	c.JSON(http.StatusOK, AuthResponse{
		Token: SignedString,
	})
}

func handleLogin(c *gin.Context, dbs *DBSafe) {
	var login loginInfo
	var userID string
	var passHash string
	var err error

	if err = c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	userQuery :=	`
					SELECT id, password_hash
					FROM users
					WHERE username = $1 AND type = 'standard';
					`

	err = DBQuery(dbs, userQuery, []any{login.Username}, &userID, &passHash)

	if err != nil {
		fmt.Println("Coulnd't get password hash for user: "+login.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server get the password hash for the user."})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passHash), []byte(login.Password))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "The passwords don't match / the hash comparison failed."})
		return
	}

	fmt.Println("Password valid for: " + login.Username + USR_ID + userID)

	var SignedString string
	SignedString, err = handler.GenerateJWT(userID, login.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": JWT_ERROR})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: SignedString,
	})
}

func FortyTwoCallback(c *gin.Context, dbs *DBSafe) {
	code := c.Query("code")

	token, err := fortyTwoOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Exchange failed"})
		return
	}

	client := fortyTwoOauthConfig.Client(context.Background(), token)

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

	err = DBQuery(dbs, userQuery, []any{userProfile.Login, userProfile.Email}, &userID, &isInsert)
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
		err = DBQuery(dbs, profileQuery, []any{userID, userProfile.Login})
		if err != nil {
			fmt.Printf("User profile creation failed %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Server couldn't create a user profile in the database."})
			return
		}
	}

	var SignedString string
	SignedString, err = handler.GenerateJWT(userID, userProfile.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": JWT_ERROR})
		return
	}

	frontendRedirectURL := fmt.Sprintf("http://localhost:8080/callback?token=%s", SignedString)
	c.Redirect(http.StatusTemporaryRedirect, frontendRedirectURL)
}
