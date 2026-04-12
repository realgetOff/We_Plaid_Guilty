package main

import (
	"fmt"
	"math/rand"
	"time"
	"context"
	"net/http"
	"strings"

	"encoding/json"

	"github.com/gin-gonic/gin"
	// "github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-jwt/jwt/v5"
)

type LobbySettings struct {
	Rounds int `json:"rounds"`
}

type playerNameTemp struct {
	PlayerName string `json:"playerName"`
}

type CreateLobbyRequest struct {
	HostID string `json:"hostId"`
	Settings LobbySettings `json:"settings"`
}

type CreateLobbyResponse struct {
	LobbyCode string `json:"lobbyCode"`
}

/*
The message structure contains the json information to be sent / received by the websocket for room generation
type: state before / after generation of the room code
code: room code
omitempty: omits empty strings, lowering network traffic
*/


type Message struct {
	Type	string			`json:"type"`
	Text	string			`json:"text,omitempty"`
	Token	string			`json:"token,omitempty"`
	Code	string			`json:"code,omitempty"`
	Reason	string			`json:"reason,omitempty"`
	Prompt	string			`json:"prompt,omitempty"`
	Drawing	string			`json:"drawing,omitempty"`
	Guess	string			`json:"guess,omitempty"`
	Votes	map[string]int	`json:"votes,omitempty"`
	Title		string		`json:"title,omitempty"`
	Description string		`json:"description,omitempty"`
	Username string			`json:"username,omitempty"`
	To		string			`json:"to,omitempty"`
	ID string				`json:"id,omitempty"`
	IsAI	bool			`json:"is_ai,omitempty"`
	Style	ProfileStyle	`json:"style,omitempty"`
}

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

func generateLobbyCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" // Removed the numeric arguements as 0123456789
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	ret := make([]byte, length)
	for i := range ret {
		ret[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(ret)
}

func createLobby(c *gin.Context) { // obsolete code
	lobbyCode := generateLobbyCode(6)
	
	//fmt.Println("The generated lobby code is: " + lobbyCode) //debug command

	c.JSON(http.StatusOK, CreateLobbyResponse{
		LobbyCode: lobbyCode,
	})
}

var jwtSecret = []byte("replace_with_env_or_equivalent_later")

type MyCustomClaims struct {
	Username string `json:"username"`
	UserID string `json:"id"`
	jwt.RegisteredClaims
}

func generateJWT(userID string, guestName string) (string, error) {
	claims := MyCustomClaims{
		Username:	guestName,
		UserID:		userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // change later, temporarily 24h
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	signedToken, err := token.SignedString(jwtSecret)
	if ( err != nil ) {
		fmt.Println("Couldn't sign / generate JWT for guest " + guestName + " where id = " + userID)
		return "", err
	}
	return signedToken, nil
}

func validateAndGetClaims(tokenString string) (*MyCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || token == nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token is invalid or claims are corrupted")
}

func handleGuestAuth(c *gin.Context, dbs *DBSafe){
	guestName := fmt.Sprintf("guest_%d%d", rand.Intn(99), time.Now().UnixNano()%1000)
	var userID string
	db := dbs.GetPool()	
	
	userQuery := "INSERT INTO users (username, is_guest) VALUES ($1, TRUE) RETURNING id;"
	err := db.QueryRow(context.Background(), userQuery, guestName).Scan(&userID);
	if (err != nil) {
		fmt.Println("Guest creation failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a guest user in the database."})
		return
	}
	fmt.Println("Guest name: " + guestName + " guest ID = " + userID)

	profileQuery := "INSERT INTO profiles (id, display_name) VALUES ($1, $2)"
	_, err = db.Exec(context.Background(), profileQuery, userID, guestName)
		if (err != nil) {
		fmt.Println("Guest profile creation failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a guest profile in the database."})
		return
	}


	var SignedString string
	SignedString, err = generateJWT(userID, guestName)
	if (err != nil) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't sign / generate JWT for guest."})
		return 
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: SignedString,
		})
}

func findRoom(c *gin.Context, serverVars *serverVarsStruct){
	code := strings.ToUpper(c.Param("code"))
	room, err := serverVars.globalHub.GetRoom(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Couldn't find the room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":		room.GetBase().ID,
		"status":	room.GetBase().Status,
		"players":	len(room.GetBase().Players),
	})
}

func FortyTwoCallback(c *gin.Context, dbs *DBSafe){
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
        // Email string `json:"email"`
        // ID    int    `json:"id"`
        // Image struct {
        //     Link string `json:"link"`
        // } `json:"image"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&userProfile); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
        return
    }

    // 5. PRINT TO TERMINAL
    fmt.Println("\n--- NEW LOGIN DETECTED ---")
    fmt.Printf("Username: %s\n", userProfile.Login)
    // fmt.Printf("Email:    %s\n", userProfile.Email)
    // fmt.Printf("42 ID:    %d\n", userProfile.ID)
    fmt.Println("--------------------------\n")


	var userID string
	db := dbs.GetPool()	
	
	userQuery := 	`INSERT INTO users (username, is_guest) 
					VALUES ($1, FALSE) 
					ON CONFLICT (username) 
					DO UPDATE SET username = EXCLUDED.username
					RETURNING id;`
	err = db.QueryRow(context.Background(), userQuery, userProfile.Login).Scan(&userID);
	if (err != nil) {
		fmt.Printf("User creation failed %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't insert a user in the database."})
		return
	}
	fmt.Println("Username: " + userProfile.Login + " user ID = " + userID)

	if (userID != ""){
	profileQuery := 	`INSERT INTO profiles (id, display_name)
						VALUES ($1, $2)
						ON CONFLICT (id) DO NOTHING;`
	_, err = db.Exec(context.Background(), profileQuery, userID, userProfile.Login)
		if (err != nil) {
		fmt.Printf("User profile creation failed %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a user profile in the database."})
		return
	}
	} // dunno if this works


	var SignedString string
	SignedString, err = generateJWT(userID, userProfile.Login)
	if (err != nil) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't sign / generate JWT for user."})
		return 
	}

	// 5. Redirect back to the React Frontend
	// We pass the token in the URL so the React app can grab it and save it.
	frontendRedirectURL := fmt.Sprintf("http://localhost:8080/callback?token=%s", SignedString)
	
	c.Redirect(http.StatusTemporaryRedirect, frontendRedirectURL)
}

// func handleLogin(c *gin.Context) {
// 	var name playerNameTemp

// 	if err := c.ShouldBindJSON(&name); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
// 		return
// 	}

// 	fmt.Println("Player name is : " + name.PlayerName)	

// 	c.JSON(http.StatusOK, gin.H{
// 		"login": "success",
// 	})
// }
