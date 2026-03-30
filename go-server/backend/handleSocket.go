package main

import (
	"strings"
	"fmt"
	"github.com/gorilla/websocket"
	"context"
	"main.go/gamemanager"
)

type HandleFunc func(context *WSContext, msg Message)

type Dispatcher struct {
	handlers map[string]HandleFunc
}

func (d *Dispatcher) Dispatch(ctx *WSContext, msg Message) {
	handler, ok := d.handlers[msg.Type]
	if !ok {
		fmt.Printf("DEBUG: %s not found\n", msg.Type)
		ctx.Conn.WriteJSON(map[string]string{"type": "error", "message": "Action not recognized by the server",})
		return
	}
	handler(ctx, msg)
}

func NewDispatcher() *Dispatcher {
	d:= &Dispatcher{
		handlers: make(map[string]HandleFunc),
	}
	d.handlers["authenticate"] = d.HandleAuth
	d.handlers["create_room"] = d.HandleCreateRoom
	d.handlers["join_room"] = d.HandleJoinRoom
	d.handlers["join_game"] = d.HandleJoinGame
	d.handlers["leave_lobby"] = d.HandleLeaveLobby
	d.handlers["leave_game"] = d.HandleLeaveGame
	d.handlers["prompt_submitted"] = d.HandlePrompt
	d.handlers["drawing_submitted"] = d.HandleDraw
	d.handlers["guess_submitted"] = d.HandleGuess
	d.handlers["chat_message"] = d.HandleChat
	d.handlers["start_game"] = d.HandleStartGame

	return d
}

func (d *Dispatcher) HandleAuth(ctx *WSContext, msg Message) {
	claims, err := validateAndGetClaims(msg.Token)
	if err != nil {
		fmt.Println("WS Auth Failed:", err)
		ctx.Conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(4001, "token expired"))
		return
	}
	ctx.CurrUsrName = &claims.Username
	ctx.CurrUsrID = &claims.UserID
	fmt.Printf("WS Authenticated: %s (ID: %s)\n", *ctx.CurrUsrName, *ctx.CurrUsrID)
}

func (d *Dispatcher) HandleCreateRoom(ctx *WSContext, msg Message) {
	if *ctx.CurrUsrName == "" { return }

	newRoom := ctx.Hub.CreateRoom()
	err := newRoom.AddPlayer(*ctx.CurrUsrID, *ctx.CurrUsrName, ctx.Conn)
	if err != nil {
		fmt.Println("DEBUG: ", err)
		return
	}

	fmt.Printf("DEBUG: %s success\n", msg.Type)

	newRoom.MessageChan <- gamemanager.Notification{
		PlayerID: *ctx.CurrUsrID,
		Data: map[string]interface{}{
			"type": "room_created",
			"code": newRoom.ID,
			"players": []map[string]interface{}{
				{
					"id":   ctx.CurrUsrID,
					"name": ctx.CurrUsrName,
					"host": true,
				},
			},
		},
	}
	
	newRoom.BroadcastLobbyState()

	go func(roomID string) {
		ctx.Db.Exec(context.Background(), "INSERT INTO rooms (room_code, created_at) VALUES ($1, NOW())", roomID)
	}(newRoom.ID)
}

func (d *Dispatcher) HandleJoinRoom(ctx *WSContext, msg Message) {
	if *ctx.CurrUsrName == "" { return }

	room, err := ctx.Hub.GetRoom(msg.Code)
	if err != nil || room == nil {
		room.MessageChan <- gamemanager.Notification{
			PlayerID: *ctx.CurrUsrID,
			Data: map[string]string{"type": "error", "message": "room not found"},
		}
		return
	}

	err = room.AddPlayer(*ctx.CurrUsrID, *ctx.CurrUsrName, ctx.Conn)
	if err != nil {
		fmt.Println("AddPlayer error:", err)
		return
	}
	room.SendSystemMsg(fmt.Sprintf("%s join the lobby !", *ctx.CurrUsrName))

	room.BroadcastLobbyState()

}

func (d *Dispatcher) HandleJoinGame(ctx *WSContext, msg Message) {
	room, err := ctx.Hub.GetRoom(msg.Code)
	if err != nil || room == nil { return }

	room.JoinGame(*ctx.CurrUsrID, ctx.Conn)
	fmt.Printf("DEBUG join_game: code='%s' user='%s'\n", msg.Code, *ctx.CurrUsrName)

	task := room.GetPlayerTask(*ctx.CurrUsrID)
	ctx.Conn.WriteJSON(task)
}

func (d *Dispatcher) HandleLeaveLobby(ctx *WSContext, msg Message) { 
	if *ctx.CurrUsrName == "" { return }
	room, err := ctx.Hub.GetRoom(msg.Code)

	if err != nil || room == nil { return }

	if room.Status != gamemanager.StateWaiting { return }

	fmt.Printf("DEBUG leave_lobby: code='%s' user='%s'\n", msg.Code, *ctx.CurrUsrName)
	isHost := false

	if p, err := room.GetPlayer(*ctx.CurrUsrID); err == nil {
		isHost = p.IsHost
	}

	room.RemovePlayer(*ctx.CurrUsrID)
	room.SendSystemMsg(fmt.Sprintf("%s leave the lobby !", *ctx.CurrUsrName))

	if len(room.Players) == 0 {
		ctx.Hub.DeleteRoom(room.ID)
		return
	}
	if isHost {
		room.TransferHost()
	}
	room.BroadcastLobbyState()
}

func (d *Dispatcher) HandleLeaveGame(ctx *WSContext, msg Message) { 
	fmt.Printf("DEBUG: leave_game msg %s\n", msg.Code)
	currentRoom, err := ctx.Hub.GetRoom(msg.Code)
	if (err != nil) { return }
	del := currentRoom.LeaveGame(*ctx.CurrUsrID)
	if del {
		ctx.Hub.DeleteRoom(msg.Code)
		fmt.Printf("DEBUG: DELETE ROOM nobody is in") // TODO Fix le isReady et set Prompt a "..."
	}
}



func (d *Dispatcher) HandlePrompt(ctx *WSContext, msg Message) {
	currentRoom, err := ctx.Hub.GetRoom(msg.Code)
	if (err != nil) { 
		fmt.Printf("Error: %v\n", err)
		return
	}
	data := map[string]interface{}{
		"type": "prompt",
		"prompt": msg.Prompt,
	}
	fmt.Printf("DEGUB: %s\n", msg.Type)
	err = currentRoom.SubmiteAction(*ctx.CurrUsrID, data, true);
	if (err != nil) {
		fmt.Printf("Error: Submited prompt: %v\n", err);
	}
}

func (d *Dispatcher) HandleDraw(ctx *WSContext, msg Message) {
	fmt.Printf("DEBUG: draw_submitted code = %s\n", msg.Code)
	currentRoom, err := ctx.Hub.GetRoom(msg.Code)
	if (err != nil) { 
		fmt.Printf("Error: %v\n", err)
		return
	}
	data := map[string]interface{}{
		"type": "draw",
		"drawing": msg.Drawing,
	}
	fmt.Printf("DEGUB: %s\n", msg.Type)
	err = currentRoom.SubmiteAction(*ctx.CurrUsrID, data, true);
	if (err != nil) {
		fmt.Printf("Error: Submited draw: %v\n", err);
	}
}

func (d *Dispatcher) HandleGuess(ctx *WSContext, msg Message) {
	currentRoom, err := ctx.Hub.GetRoom(msg.Code)
	if (err != nil) { 
		fmt.Printf("Error: %v\n", err)
		return
	}
	data := map[string]interface{}{
		"type": "guess",
		"guess": msg.Guess,
	}
	fmt.Printf("DEGUB: %s\n", msg.Type)
	err = currentRoom.SubmiteAction(*ctx.CurrUsrID, data, true);
	if (err != nil) {
		fmt.Printf("Error: Submited guess: %v\n", err);
	}
}


func (d *Dispatcher) HandleStartGame(ctx *WSContext, msg Message) {
	if *ctx.CurrUsrName == "" { return }

	room, err := ctx.Hub.GetRoom(msg.Code)
	if err != nil || room == nil { return }

	err = room.StartGame()
	if err != nil {
		ctx.Conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
		return
	}
	fmt.Printf("DEBUG: start_game\n")
	room.BroadcastToAll(map[string]interface{}{
		"type": "start_game",
		"code": room.ID,
	})
}

func (d *Dispatcher) HandleChat(ctx *WSContext, msg Message) {
	if (*ctx.CurrUsrName == "") { return }
	if (len(strings.TrimSpace(msg.Text)) == 0) { return }
	room, err := ctx.Hub.GetRoom(msg.Code)
	if (err != nil) { return }
	fmt.Printf("DEBUG: chat_message\n")
	room.BroadcastChat(*ctx.CurrUsrID, msg.Text)
}
