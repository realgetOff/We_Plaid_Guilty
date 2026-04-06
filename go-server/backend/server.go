package main

import (
	// "strings"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	// "github.com/jackc/pgx/v5/pgxpool"
	// "github.com/golang-jwt/jwt/v5"
	"main.go/gamemanager"
	"net/http"
)


func socketLogic(client *Client, serverVars *serverVarsStruct) {
	dispatcher := NewDispatcher()

	ctx := WSContext {
		client: client,
		chub: serverVars.ClientHub,
	}

	for {
		var msg Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		dispatcher.Dispatch(&ctx, msg)
	}

	if client.CurrentRoom != nil && *client.CurrUsrID != "" {
		isHost := false
		base := client.CurrentRoom.GetBase()
		if p, err := base.GetPlayer(*client.CurrUsrID); err == nil {
			isHost = p.IsHost
		}

		base.RemovePlayer(*client.CurrUsrID)
		if classicRoom, ok := client.CurrentRoom.(*gamemanager.Room); ok {
			classicRoom.SendSystemMsg(fmt.Sprintf("%s leave the lobby !", *client.CurrUsrName))
		}

		if len(base.Players) == 0 {
			serverVars.globalHub.DeleteRoom(base.ID)
		} else {
			if isHost {
				base.TransferHost()
			}
			base.BroadcastLobbyState()
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebsocket(c *gin.Context, serverVars *serverVarsStruct) {

	userID := c.GetString("userID")
    userName := c.GetString("userName")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
        CurrUsrID:   &userID,
        CurrUsrName: &userName,
        Conn:        conn,
        Hub:         serverVars.globalHub, // Your gamemanager.Hub
    }

    // 4. Register the client in the Global Registry
    serverVars.ClientHub.mu.Lock()
    serverVars.ClientHub.Clients[userID] = client
    serverVars.ClientHub.mu.Unlock()
	
	defer func() {
		serverVars.ClientHub.mu.Lock()
        delete(serverVars.ClientHub.Clients, userID)
        serverVars.ClientHub.mu.Unlock()
        conn.Close()
	}()

	socketLogic(client, serverVars)
}
