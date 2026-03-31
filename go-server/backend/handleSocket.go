package main

import (
	"strings"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"context"
	"main.go/gamemanager"
)

type WSContext struct {
	CurrUsrID *string
	CurrUsrName *string
	Db *pgxpool.Pool
	Conn *websocket.Conn
	Hub *gamemanager.Hub
	CurrentRoom *gamemanager.BaseRoom
}

type HandleFunc func(context *WSContext, msg Message)
type PipeFunc func(ctx *WSContext, msg Message) bool

type Dispatcher struct {
	handlers map[string]HandleFunc
}

func RunPipeLine(ctx *WSContext, msg Message, pipes ...PipeFunc) bool {
	for _, pipe := range pipes {
		if !pipe(ctx, msg) {
			return false
		}
	}
	return true
}

func (d *Dispatcher) PipeIsValidChat(ctx *WSContext, msg Message) bool {
	if len(strings.TrimSpace(msg.Text)) == 0 { return false }
	return true
}

func (d *Dispatcher) PipeIsAuth(ctx *WSContext, msg Message) bool {
	if ctx.CurrUsrID == nil || ctx.CurrUsrName == nil || *ctx.CurrUsrName == "" {
		return false
	}
	return true
}

func (d *Dispatcher) PipeHasRoomCode(ctx *WSContext, msg Message) bool {
	if msg.Code == "" {
		return false
	}
	return true
}

func (d *Dispatcher) PipeRoomExist(ctx *WSContext, msg Message) bool {
	tmpRoom, err := ctx.Hub.GetRoom(msg.Code)
	if err != nil || tmpRoom == nil {
		return false
	}
	ctx.CurrentRoom = tmpRoom
	return true
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
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) { return }

	ctx.CurrentRoom = ctx.Hub.CreateRoom()
	err := ctx.CurrentRoom.AddPlayer(*ctx.CurrUsrID, *ctx.CurrUsrName, ctx.Conn)
	if err != nil {
		fmt.Println("DEBUG: ", err)
		return
	}

	fmt.Printf("DEBUG: %s success\n", msg.Type)

	ctx.CurrentRoom.MessageChan <- gamemanager.Notification{
		PlayerID: *ctx.CurrUsrID,
		Data: map[string]interface{}{
			"type": "room_created",
			"code": ctx.CurrentRoom.ID,
			"players": []map[string]interface{}{
				{
					"id":   ctx.CurrUsrID,
					"name": ctx.CurrUsrName,
					"host": true,
				},
			},
		},
	}
	
	ctx.CurrentRoom.BroadcastLobbyState()

	go func(roomID string) {
		ctx.Db.Exec(context.Background(), "INSERT INTO rooms (room_code, created_at) VALUES ($1, NOW())", roomID)
	}(ctx.CurrentRoom.ID)
}

func (d *Dispatcher) HandleJoinRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	err := ctx.CurrentRoom.AddPlayer(*ctx.CurrUsrID, *ctx.CurrUsrName, ctx.Conn)
	if err != nil {
		fmt.Println("AddPlayer error:", err)
		return
	}
	ctx.CurrentRoom.SendSystemMsg(fmt.Sprintf("%s join the lobby !", *ctx.CurrUsrName))

	ctx.CurrentRoom.BroadcastLobbyState()

}

func (d *Dispatcher) HandleJoinGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }

	ctx.CurrentRoom.JoinGame(*ctx.CurrUsrID, ctx.Conn)
	fmt.Printf("DEBUG join_game: code='%s' user='%s'\n", msg.Code, *ctx.CurrUsrName)

	task := ctx.CurrentRoom.GetPlayerTask(*ctx.CurrUsrID)
	ctx.Conn.WriteJSON(task)
}

func (d *Dispatcher) HandleLeaveLobby(ctx *WSContext, msg Message) { 
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }


	if ctx.CurrentRoom.Status != gamemanager.StateWaiting { return }

	fmt.Printf("DEBUG leave_lobby: code='%s' user='%s'\n", msg.Code, *ctx.CurrUsrName)
	isHost := false

	if p, err := ctx.CurrentRoom.GetPlayer(*ctx.CurrUsrID); err == nil {
		isHost = p.IsHost
	}

	ctx.CurrentRoom.RemovePlayer(*ctx.CurrUsrID)
	ctx.CurrentRoom.SendSystemMsg(fmt.Sprintf("%s leave the lobby !", *ctx.CurrUsrName))

	if len(ctx.CurrentRoom.Players) == 0 {
		ctx.Hub.DeleteRoom(ctx.CurrentRoom.ID)
		return
	}
	if isHost {
		ctx.CurrentRoom.TransferHost()
	}
	ctx.CurrentRoom.BroadcastLobbyState()
}

func (d *Dispatcher) HandleLeaveGame(ctx *WSContext, msg Message) { 
	fmt.Printf("DEBUG: leave_game msg %s\n", msg.Code)
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	del := ctx.CurrentRoom.LeaveGame(*ctx.CurrUsrID)
	if del {
		ctx.Hub.DeleteRoom(msg.Code)
		fmt.Printf("DEBUG: DELETE ROOM nobody is in") // TODO Fix le isReady et set Prompt a "..."
	}
}



func (d *Dispatcher) HandlePrompt(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	data := map[string]interface{}{
		"type": "prompt",
		"prompt": msg.Prompt,
	}
	fmt.Printf("DEGUB: %s\n", msg.Type)
	err := ctx.CurrentRoom.SubmiteAction(*ctx.CurrUsrID, data, true);
	if (err != nil) {
		fmt.Printf("Error: Submited prompt: %v\n", err);
	}
}

func (d *Dispatcher) HandleDraw(ctx *WSContext, msg Message) {
	fmt.Printf("DEBUG: draw_submitted code = %s\n", msg.Code)
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	data := map[string]interface{}{
		"type": "draw",
		"drawing": msg.Drawing,
	}
	fmt.Printf("DEGUB: %s\n", msg.Type)
	err := ctx.CurrentRoom.SubmiteAction(*ctx.CurrUsrID, data, true);
	if (err != nil) {
		fmt.Printf("Error: Submited draw: %v\n", err);
	}
}

func (d *Dispatcher) HandleGuess(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	data := map[string]interface{}{
		"type": "guess",
		"guess": msg.Guess,
	}
	fmt.Printf("DEGUB: %s\n", msg.Type)
	err := ctx.CurrentRoom.SubmiteAction(*ctx.CurrUsrID, data, true);
	if (err != nil) {
		fmt.Printf("Error: Submited guess: %v\n", err);
	}
}


func (d *Dispatcher) HandleStartGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeHasRoomCode)) { return }

	err := ctx.CurrentRoom.StartGame()
	if err != nil {
		ctx.Conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
		return
	}
	fmt.Printf("DEBUG: start_game\n")
	ctx.CurrentRoom.BroadcastToAll(map[string]interface{}{
		"type": "start_game",
		"code": ctx.CurrentRoom.ID,
	})
}

func (d *Dispatcher) HandleChat(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeIsValidChat)) { return }

	fmt.Printf("DEBUG: chat_message\n")
	ctx.CurrentRoom.BroadcastChat(*ctx.CurrUsrID, msg.Text)
}
