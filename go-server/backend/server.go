package main

import (
	// "strings"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-jwt/jwt/v5"
	"main.go/gamemanager"
	"net/http"
)


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

func socketLogic(conn *websocket.Conn, db *pgxpool.Pool, hub *gamemanager.Hub) {
	dispatcher := NewDispatcher()

	ctx := &WSContext{
		Db: db,
		Conn: conn,
		Hub: hub,
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

		if len(base.Players) == 0 {
			hub.DeleteRoom(base.ID)
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

func handleWebsocket(c *gin.Context, db *pgxpool.Pool, hub *gamemanager.Hub) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	socketLogic(conn, db, hub)
}
