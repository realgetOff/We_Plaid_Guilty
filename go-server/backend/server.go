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


func socketLogic(conn *websocket.Conn, serverVars *serverVarsStruct) {
	dispatcher := NewDispatcher()

	ctx := &WSContext{
		Db: serverVars.db,
		Conn: conn,
		Hub: serverVars.globalHub,
	}
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		dispatcher.Dispatch(ctx, msg)
	}

	if ctx.CurrentRoom != nil && *ctx.CurrUsrID != "" {
		isHost := false
		base := ctx.CurrentRoom.GetBase()
		if p, err := base.GetPlayer(*ctx.CurrUsrID); err == nil {
			isHost = p.IsHost
		}

		base.RemovePlayer(*ctx.CurrUsrID)
		if classicRoom, ok := ctx.CurrentRoom.(*gamemanager.Room); ok {
			classicRoom.SendSystemMsg(fmt.Sprintf("%s leave the lobby !", *ctx.CurrUsrName))
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
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	socketLogic(conn, serverVars)
}
