package main

import (
	"fmt"
	"math/rand"
	"time"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-jwt/jwt/v5"
)

type LobbySettings struct {
	Rounds int `json:"rounds"`
	Timer int `json:"timer"`
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
    Type    string         `json:"type"`
	Text    string         `json:"text,omitempty"`
    Token   string         `json:"token,omitempty"`
    Code    string         `json:"code,omitempty"`
    Reason  string         `json:"reason,omitempty"`
    Prompt  string         `json:"prompt,omitempty"`
    Drawing string         `json:"drawing,omitempty"`
    Guess   string         `json:"guess,omitempty"`
    Votes   map[string]int `json:"votes,omitempty"`
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


func handleGuestAuth(c *gin.Context, db *pgxpool.Pool){
	guestName := fmt.Sprintf("guest_%d%d", rand.Intn(99), time.Now().UnixNano()%1000)
	var userID string

	query := "INSERT INTO users (username, is_guest) VALUES ($1, TRUE) RETURNING id;"
	err := db.QueryRow(context.Background(), query, guestName).Scan(&userID);

	if (err != nil) {
		fmt.Println("Guest creation failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Server couldn't create a guest user in the database."})
		return
	}
	fmt.Println("Guest name: " + guestName + " guest ID = " + userID)

	var SignedString string
	SignedString, err = generateJWT(userID, guestName)
	if (err != nil) {
		// fmt.Println("Guest creation failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't sign / generate JWT for guest."})
		return 
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: SignedString,
		})
}

func findRoom(c *gin.Context, serverVars *serverVarsStruct, roomtype int){
    code := strings.ToUpper(c.Param("code"))

	if roomtype == 1 {
    	room, err := globalAIHub.GetRoom(code)
    	if err != nil {
    	    c.JSON(http.StatusNotFound, gin.H{"error": "Couldn't find the AI room"})
    	    return
		}
		c.JSON(http.StatusOK, gin.H{
        "code":    room.ID,
        "status":  room.Status,
        "players": len(room.Players),
		})
	} else {
		room, err := serverVars.globalHub.GetRoom(code)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Couldn't find the room"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
        "code":    room.ID,
        "status":  room.Status,
        "players": len(room.Players),
		})
	}
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
