package main

import (
	// "strings"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	// "github.com/jackc/pgx/v5/pgxpool"
	// "github.com/golang-jwt/jwt/v5"
	"net/http"

	"main.go/gamemanager"
)


func socketLogic(client *Client, serverVars *serverVarsStruct) {
	dispatcher := NewDispatcher()

	ctx := WSContext {
		client: client,
		chub: serverVars.ClientHub,
	}

	// go ctx.client.Hub.LogRoom()
	defer func() {
		if client.CurrUsrID == nil || *client.CurrUsrID == "" {
			return
		}
		serverVars.ClientHub.mu.Lock()
		delete(serverVars.ClientHub.Clients, *client.CurrUsrID)
		serverVars.ClientHub.mu.Unlock()
	}()
	for {
		var msg Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		fmt.Printf("DEBUG: MSG = %s\n", msg.Type)
		dispatcher.Dispatch(&ctx, msg)
	}

	if client.CurrentRoom != nil && client.CurrUsrID != nil && *client.CurrUsrID != "" {
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
			time.Sleep(1 * time.Second)
			if len(base.Players) == 0 {
				serverVars.globalHub.DeleteRoom(base.ID)
			}
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
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		Conn:	conn,
		Hub:	serverVars.globalHub,
	}
	
	defer conn.Close()

	// increase / decrease the activeWebsockets gauge for metrics
	activeWebsockets.Inc()
	defer activeWebsockets.Dec()

	socketLogic(client, serverVars)
}
