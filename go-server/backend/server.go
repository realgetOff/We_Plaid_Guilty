package main

import (
	"strings"
	// "time"
	// "context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-jwt/jwt/v5"
	"main.go/gamemanager"
	"net/http"
)

type WSContext struct {
	CurrUsrID *string
	CurrUsrName *string
	Db *pgxpool.Pool
	Conn *websocket.Conn
	HubAI *gamemanager.AIHub
	Hub *gamemanager.Hub
	CurrentRoom *gamemanager.Room
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

func socketLogic(conn *websocket.Conn, db *pgxpool.Pool, hub *gamemanager.Hub, hubAI *gamemanager.AIHub) {
	var currentUsername string
	var currentUserID string
	var currentRoom *gamemanager.Room
	var currentAIRoom *gamemanager.AIRoom

	dispatcher := NewDispatcher()

	ctx := &WSContext{
		Db: db,
		Conn: conn,
		Hub: hub,
		HubAI: hubAI,
	}
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		dispatcher.Dispatch(ctx, msg)

		// AI_GAME_GESTION
		if msg.Type == "create_ai_room" {
			if currentUsername == "" { continue }

			newRoom := hubAI.CreateRoom()
			currentAIRoom = newRoom
			err := newRoom.AddPlayer(currentUserID, currentUsername, conn)
			if err != nil { continue }
			fmt.Printf("DEGUB: AI_ROOM created\n")	

			newRoom.MessageChan <- gamemanager.Notification{
				PlayerID: currentUserID,
				Data: map[string]interface{}{
					"type": "ai_room_created",
					"code": newRoom.ID,
					"palyers": []map[string]interface{}{
						{
							"id": currentUserID,
							"name": currentUsername,
							"host": true,
						},
					},
				},
			}
			fmt.Printf("DEBUG: %s\n", msg.Type)
			newRoom.BroadcastLobbyState()
		}
		if msg.Type == "join_ai_room" {
			if currentUsername == "" { continue }

			room, err := hubAI.GetRoom(msg.Code)
			if err != nil || room == nil {
				conn.WriteJSON(map[string]interface{}{
					"type":    "error",
					"message": "AI room not found",
				})
				continue
			}

			_, err = hubAI.AddPlayerToRoom(msg.Code, currentUserID, currentUsername, conn)
			if err != nil {
				conn.WriteJSON(map[string]interface{}{
					"type":    "error",
					"message": err.Error(),
				})
				continue
			}

			currentAIRoom = room

			room.MessageChan <- gamemanager.Notification{
				PlayerID: currentUserID,
				Data: map[string]interface{}{
					"type":  "ai_room_joined",
					"code":  room.ID,
					"my_id": currentUserID,
				},
			}
			fmt.Printf("DEBUG: %s rejoint par %s\n", msg.Code, currentUsername)

			room.SendSystemMsg(fmt.Sprintf("%s join the lobby !", currentUsername))
			room.BroadcastLobbyState()
		}

		if msg.Type == "ai_chat_message" {
			if (currentUsername == "") { continue }
			if (len(strings.TrimSpace(msg.Text)) == 0) { continue }
			room, err := hubAI.GetRoom(msg.Code)
			if (err != nil) { continue }
			fmt.Printf("DEGUB: %s\n", msg.Type)

			room.BroadcastChat(currentUserID, msg.Text)
		}

		if msg.Type == "start_ai_game" {
			if currentUsername == "" { continue }

			room, err := hubAI.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			if room.Status != gamemanager.StateAIWaiting { continue }

			prompt, err := gamemanager.CallAI("")
			if err != nil {
				fmt.Println("callAI error:", err)
				prompt = "Error API : add Credits"
			}

			currentAIRoom = room

			currentAIRoom.BroadcastToAll(map[string]interface{}{
				"type": "start_ai_game",
				"code": room.ID,
			})
			fmt.Printf("DEBUG: %s\n", msg.Type)
			go currentAIRoom.RunAIGameLoop(prompt)
		}

		if msg.Type == "join_ai_game" {
			if currentUsername == "" { continue }

			room, err := hubAI.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			hubAI.UpdatePlayerConn(msg.Code, currentUserID, conn)
			currentAIRoom = room
			if (room.Status == gamemanager.StateAIDrawing) {
				room.MessageChan <- gamemanager.Notification{
					PlayerID: currentUserID,
					Data: map[string]interface{}{
						"type":   "ai_game_state",
						"phase":  "draw",
						"prompt": room.Prompt,
						"room":   room.ID,
					},
				}
			}
			fmt.Printf("DEGUB: %s\n", msg.Type)
		}

		if msg.Type == "ai_drawing_submitted" {
    		if currentUsername == "" { continue }

    		room, err := globalAIHub.GetRoom(msg.Code)
   			if err != nil || room == nil { continue }

    		room.SubmitDrawing(currentUserID, msg.Drawing, msg.Title, msg.Description)
		}

		if msg.Type == "ai_votes_submitted" {
			if currentUsername == "" { continue }

			room, err := hubAI.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			room.SubmitVotes(currentUserID, msg.Votes)
		}
		if msg.Type == "leave_ai_game" {
			if currentUsername == "" { continue }

			room, err := hubAI.GetRoom(msg.Code)
			if err != nil { continue }

			del := room.LeaveGame(currentUserID)
			if del {
				hubAI.DeleteRoom(room.ID)
				fmt.Printf("DEBUG: Delete ROOM everybody quit\n")
			}
		}
		if msg.Type == "leave_ai_room" {
			if currentUsername == "" { continue }

			room, err := hubAI.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			if room.Status != gamemanager.StateAIWaiting { continue }

			isHost := false
			if p, ok := room.GetPlayer(currentUserID); ok {
				isHost = p.IsHost
			}

			room.RemovePlayer(currentUserID)

			if len(room.Players) == 0 {
				hubAI.DeleteRoom(room.ID)
				continue
			}

			if isHost {
				room.TransferHost()
			}

			room.SendSystemMsg(fmt.Sprintf("%s leave the lobby !", currentUsername))
			room.BroadcastLobbyState()

		}
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

func handleWebsocket(c *gin.Context, db *pgxpool.Pool, hub *gamemanager.Hub, hubAI *gamemanager.AIHub) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	socketLogic(conn, db, hub, hubAI)
}
