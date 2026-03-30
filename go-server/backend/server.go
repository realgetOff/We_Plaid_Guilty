package main

import (
	"strings"
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
				conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(4001, "token expired"))
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
			room.SendSystemMsg(fmt.Sprintf("%s join the lobby !", currentUsername))

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
			fmt.Printf("DEBUG leave_lobby: code='%s' user='%s'\n", msg.Code, currentUsername)
			isHost := false
			if p, err := room.GetPlayer(currentUserID); err == nil {
				isHost = p.IsHost
			}
			room.RemovePlayer(currentUserID)
			room.SendSystemMsg(fmt.Sprintf("%s leave the lobby !", currentUsername))
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

		if (msg.Type == "prompt_submitted") {
			if (currentRoom == nil) { continue }
			data := map[string]interface{}{
				"type": "prompt",
				"prompt": msg.Prompt,
			}
			fmt.Printf("DEBUG: PROMPT\n")
			err := currentRoom.SubmiteAction(currentUserID, data, true);
			if (err != nil) {
				fmt.Printf("Error: Submited prompt: %v\n", err);
			}
		}
		if (msg.Type == "drawing_submitted") {
			if (currentRoom == nil) { continue }
			data := map[string]interface{}{
				"type": "draw",
				"drawing": msg.Drawing,
			}
			fmt.Printf("DEBUG: DRAW")
			err := currentRoom.SubmiteAction(currentUserID, data, true);
			if (err != nil) {
				fmt.Printf("Error: Submited draw: %v\n", err);
			}
		}
		if (msg.Type == "guess_submitted") {
			if (currentRoom == nil) { continue }
			data := map[string]interface{}{
				"type": "guess",
				"guess": msg.Guess,
			}
			fmt.Printf("DEBUG: GUESS\n")
			err := currentRoom.SubmiteAction(currentUserID, data, true);
			if (err != nil) {
				fmt.Printf("Error: Submited guess: %v\n", err);
			}
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
			fmt.Printf("DEBUG: start_game\n")
			room.BroadcastToAll(map[string]interface{}{
				"type": "start_game",
				"room": room.ID,
			})
		}
		if (msg.Type == "leave_game") {
			fmt.Printf("DEBUG: leave_game msg %s\n", msg.Code)
			room, err := hub.GetRoom(msg.Code)
			if (err != nil) { continue }
			room.LeaveGame(currentUserID)
		}
		if msg.Type == "join_game" {

			room, err := hub.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			room.JoinGame(currentUserID, conn)
			// room.UpdatePlayerConn(currentUserID, conn)
			fmt.Printf("DEBUG join_game: code='%s' user='%s'\n", msg.Code, currentUsername)
			// fmt.Printf("DEBUG join_game GetRoom: room=%v err=%v\n", room != nil, err)
			currentRoom = room

			task := room.GetPlayerTask(currentUserID)
			// fmt.Printf("DEBUG join_game task: %+v\n", task) Système qui annonce les nou
			conn.WriteJSON(task)
		}

		if msg.Type == "create_ai_room" {
			if currentUsername == "" { continue }

			newRoom := globalAIHub.CreateRoom()
			newRoom.Players[currentUserID] = &gamemanager.Player{
				ID:          currentUserID,
				Name:        currentUsername,
				Conn:        conn,
				IsHost:      true,
				IsConnected: true,
			}

			newRoom.BroadcastLobbyState()

			newRoom.MessageChan <- gamemanager.Notification{
				PlayerID: currentUserID,
				Data: map[string]interface{}{
					"type": "ai_room_created",
					"code": newRoom.ID,
				},
			}
		}

		if msg.Type == "join_ai_room" {
			if currentUsername == "" { continue }

			room, err := globalAIHub.GetRoom(msg.Code)
			if err != nil || room == nil {
				// Fix : passe par WriteMu pour éviter le concurrent write
				room2, _ := globalAIHub.GetRoom(msg.Code)
				if room2 != nil {
					room2.MessageChan <- gamemanager.Notification{
						PlayerID: currentUserID,
						Data:     map[string]string{"type": "error", "message": "ai room not found"},
					}
				}
				continue
			}

			_, err = globalAIHub.AddPlayerToRoom(msg.Code, currentUserID, currentUsername, conn)
			if err != nil {
				room.MessageChan <- gamemanager.Notification{
					PlayerID: currentUserID,
					Data:     map[string]string{"type": "error", "message": err.Error()},
				}
				continue
			}

			room.SendSystemMsg(fmt.Sprintf("%s join the lobby !", currentUsername))
			room.BroadcastLobbyState()
		}

		if msg.Type == "chat_message" {
			if (currentUsername == "") { continue }
			if (len(strings.TrimSpace(msg.Text)) == 0) { continue }
			room, err := globalHub.GetRoom(msg.Code)
			if (err != nil) { continue }
			fmt.Printf("DEBUG: char_message\n")
			room.BroadcastChat(currentUserID, msg.Text)
		}

		if msg.Type == "ai_chat_message" {
			if (currentUsername == "") { continue }
			if (len(strings.TrimSpace(msg.Text)) == 0) { continue }
			room, err := globalAIHub.GetRoom(msg.Code)
			if (err != nil) { continue }

			room.BroadcastChat(currentUserID, msg.Text)
		}

		if msg.Type == "start_ai_game" {
			if currentUsername == "" { continue }

			room, err := globalAIHub.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			if room.Status != gamemanager.StateAIWaiting { continue }

			prompt, err := gamemanager.CallAI("")
			if err != nil {
				fmt.Println("callAI error:", err)
				prompt = "Dessine la meilleure façon de survivre à une réunion de travail"
			}

			room.BroadcastToAll(map[string]interface{}{
				"type": "start_ai_game",
				"room": room.ID,
			})

			go room.RunAIGameLoop(prompt)
		}

		if msg.Type == "join_ai_game" {
			if currentUsername == "" { continue }

			room, err := globalAIHub.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			globalAIHub.UpdatePlayerConn(msg.Code, currentUserID, conn)

			prompt := room.Prompt
			status := room.Status

			if status == gamemanager.StateAIDrawing {
				// Fix : passe par MessageChan au lieu de conn.WriteJSON direct
				room.MessageChan <- gamemanager.Notification{
					PlayerID: currentUserID,
					Data: map[string]interface{}{
						"type":   "ai_game_state",
						"phase":  "draw",
						"room":   msg.Code,
						"prompt": prompt,
						"my_id":  currentUserID,
					},
				}
			}
		}

		if msg.Type == "ai_drawing_submitted" {
			if currentUsername == "" { continue }

			room, err := globalAIHub.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			room.SubmitDrawing(currentUserID, msg.Drawing)
		}

		if msg.Type == "ai_votes_submitted" {
			if currentUsername == "" { continue }

			room, err := globalAIHub.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			room.SubmitVotes(currentUserID, msg.Votes)
		}

		if msg.Type == "leave_ai_room" {
			if currentUsername == "" { continue }

			room, err := globalAIHub.GetRoom(msg.Code)
			if err != nil || room == nil { continue }

			if room.Status != gamemanager.StateAIWaiting { continue }

			isHost := false
			if p, ok := room.GetPlayer(currentUserID); ok {
				isHost = p.IsHost
			}

			room.RemovePlayer(currentUserID)

			if len(room.Players) == 0 {
				globalAIHub.DeleteRoom(room.ID)
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

func handleWebsocket(c *gin.Context, db *pgxpool.Pool, hub *gamemanager.Hub) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	socketLogic(conn, db, hub)
}
