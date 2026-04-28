package server

import (
	"fmt"
	"time"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/realgetOff/We_Plaid_Guilty/internal/gamemanager"
	"github.com/realgetOff/We_Plaid_Guilty/internal/handler"
	"github.com/realgetOff/We_Plaid_Guilty/internal/metrics"
)

func socketLogic(client *handler.Client, serverVars *ServerVarsStruct) {
	dispatcher := handler.NewDispatcher()

	ctx := handler.WSContext{
		Client: client,
		Chub:   serverVars.ClientHub,
	}

	defer func() {
		if client.CurrUsrID == nil || *client.CurrUsrID == "" {
			return
		}
		serverVars.ClientHub.Mu.Lock()
		delete(serverVars.ClientHub.Clients, *client.CurrUsrID)
		serverVars.ClientHub.Mu.Unlock()
	}()
	for {
		var msg handler.Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		fmt.Printf("DEBUG: MSG = %s\n", msg.Type)
		dispatcher.Dispatch(&ctx, msg)
	}

	if client.CurrentRoom == nil || client.CurrUsrID == nil || *client.CurrUsrID == "" {
		return
	}

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
		time.Sleep(15 * time.Second)
		if len(base.Players) == 0 {
			serverVars.globalHub.DeleteRoom(base.ID)
			return
		}
	}

	if isHost && len(base.Players) != 0 {
		base.TransferHost()
	}

	base.BroadcastLobbyState()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebsocket(c *gin.Context, serverVars *ServerVarsStruct) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &handler.Client{
		Conn: conn,
		Hub:  serverVars.globalHub,
	}

	defer conn.Close()

	// increase / decrease the activeWebsockets gauge for metrics
	metrics.ActiveWebsockets.Inc()
	defer metrics.ActiveWebsockets.Dec()

	socketLogic(client, serverVars)
}
