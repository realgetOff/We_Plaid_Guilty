package main

import (
	"fmt"
	"math/rand"
	"time"

	"net/http"
	"github.com/gin-gonic/gin"
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


type Message struct {
	Type	string `json:"type"`
	Token	string `json:"token,omitempty"`
	Code	string `json:"code,omitempty"`
	Reason	string `json:"reason,omitempty"`
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

func createLobby(c *gin.Context) {
	// var req CreateLobbyRequest
// 
	// if err := c.ShouldBindJSON(&req); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		// return
	// }

	lobbyCode := generateLobbyCode(6)
	
	//fmt.Println("The generated lobby code is: " + lobbyCode) //debug command

	c.JSON(http.StatusOK, CreateLobbyResponse{
		LobbyCode: lobbyCode,
	})
}

func handleGuestAuth(c *gin.Context){
	guestName := fmt.Sprintf("guest_%d%d", rand.Intn(99), time.Now().UnixNano()%1000)

	fmt.Println("Guest name: " + guestName)
	c.JSON(http.StatusOK, AuthResponse{
		Token: guestName,
	})
}

func handleLogin(c *gin.Context) {
	var name playerNameTemp

	if err := c.ShouldBindJSON(&name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	fmt.Println("Player name is : " + name.PlayerName)

	c.JSON(http.StatusOK, gin.H{
		"login": "success",
	})
}
