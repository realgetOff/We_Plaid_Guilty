package main

import (
	"context"
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
	var currentUsername string
	var currentUserID string
	var currentRoom *gamemanager.Room

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		if msg.Type == "authenticate" {
			claims, err := validateAndGetClaims(msg.Token)
			if err != nil {
				fmt.Println("WS Auth Failed:", err)
				return
			}
			currentUsername = claims.Username
			currentUserID = claims.UserID
			fmt.Printf("WS Authenticated: %s (ID: %s)\n", currentUsername, currentUserID)
			continue
		}

		if msg.Type == "create_room" {
			if currentUsername == "" { continue }

			newRoom := hub.CreateRoom()
			err = newRoom.AddPlayer(currentUserID, currentUsername, conn)

			currentRoom = newRoom
			newRoom.MessageChan <- gamemanager.Notification{
				PlayerID: currentUserID,
				Data: map[string]interface{}{
					"type": "room_created",
					"code": newRoom.ID,
					"players": []map[string]interface{}{
						{
							"id":   currentUserID,
							"name": currentUsername,
							"host": true,
						},
					},
				},
			}

			newRoom.BroadcastLobbyState()

			go func(roomID string) {
				db.Exec(context.Background(), "INSERT INTO rooms (room_code, created_at) VALUES ($1, NOW())", roomID)
			}(newRoom.ID)
		}

		if msg.Type == "join_room" {
			if currentUsername == "" { continue }

			room, err := hub.GetRoom(msg.Code)
			if err != nil || room == nil {
				fmt.Println("Join failed: room is nil or not found")
				conn.WriteJSON(map[string]string{"type": "error", "message": "room not found"})
				continue
			}

			err = room.AddPlayer(currentUserID, currentUsername, conn)
			if err != nil {
				fmt.Println("AddPlayer error:", err)
				continue
			}

			currentRoom = room

			room.BroadcastLobbyState()
		}

		if msg.Type == "leave_lobby" {
			if currentUsername == "" { continue }
			room, err := hub.GetRoom(msg.Code)
			if err != nil || room == nil { continue }
			if room.Status != gamemanager.StateWaiting {
				continue
			}
			isHost := false
			if p, err := room.GetPlayer(currentUserID); err == nil {
				isHost = p.IsHost
			}
			room.RemovePlayer(currentUserID)
			currentRoom = nil
			if len(room.Players) == 0 {
				hub.DeleteRoom(room.ID)
				continue
			}
			if isHost {
				room.TransferHost()
			}
			room.BroadcastLobbyState()
		}

		if msg.Type == "start_game" {
			if currentUsername == "" { continue }

			room, err := hub.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			err = room.StartGame()
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				continue
			}
			room.BroadcastToAll(map[string]interface{}{
				"type": "start_game",
				"room": room.ID,
			})
		}
		if msg.Type == "join_game" {
			fmt.Printf("DEBUG join_game: code='%s' user='%s'\n", msg.Code, currentUsername)

			room, err := hub.GetRoom(msg.Code)
			fmt.Printf("DEBUG join_game GetRoom: room=%v err=%v\n", room != nil, err)
			if err != nil || room == nil { continue }

			room.UpdatePlayerConn(currentUserID, conn)
			currentRoom = room

			task := room.GetPlayerTask(currentUserID)
			fmt.Printf("DEBUG join_game task: %+v\n", task)
			conn.WriteJSON(task)
		}
		if msg.Type == "join_game" {
			fmt.Printf("DEBUG join_game: code='%s' user='%s'\n", msg.Code, currentUsername)
			if currentUsername == "" { continue }

			room, err := hub.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			room.UpdatePlayerConn(currentUserID, conn)
			currentRoom = room

			task := room.GetPlayerTask(currentUserID)
			conn.WriteJSON(task)
		}
		if (msg.Type == "prompt_submitted") {
			fmt.Printf("DEBUG: PROMPT")
			if (currentRoom == nil) { continue }
			data := map[string]interface{}{
				"type": "prompt",
				"prompt": msg.Prompt,
			}
			err := currentRoom.SubmiteAction(currentUserID, data, true);
			if (err != nil) {
				fmt.Printf("Error: Submited prompt: %v\n", err);
			}
		}
		if (msg.Type == "drawing_submitted") {
			fmt.Printf("DEBUG: DRAW")
			if (currentRoom == nil) { continue }
			data := map[string]interface{}{
				"type": "draw",
				"drawing": msg.Drawing,
			}
			err := currentRoom.SubmiteAction(currentUserID, data, true);
			if (err != nil) {
				fmt.Printf("Error: Submited draw: %v\n", err);
			}
		}
		if (msg.Type == "guess_submitted") {
			fmt.Printf("DEBUG: GUESS")
			if (currentRoom == nil) { continue }
			data := map[string]interface{}{
				"type": "guess",
				"guess": msg.Guess,
			}
			err := currentRoom.SubmiteAction(currentUserID, data, true);
			if (err != nil) {
				fmt.Printf("Error: Submited guess: %v\n", err);
			}
		}

		// if (msg.Type == )
	}
	
	if currentRoom != nil && currentUserID != "" {
		isHost := false
		if p, err := currentRoom.GetPlayer(currentUserID); err == nil {
			isHost = p.IsHost
		}

		currentRoom.RemovePlayer(currentUserID)

		if len(currentRoom.Players) == 0 {
			hub.DeleteRoom(currentRoom.ID)
		} else {
			if isHost {
				currentRoom.TransferHost()
			}
			currentRoom.BroadcastLobbyState()
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
