package handler

import (
	"fmt"
	"context"

	"main.go/gamemanager"
	"main.go/metrics"

)

func (d *Dispatcher) HandleCreateRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) { return }

	fmt.Printf("DEBUG: create_room\n")
	var color string
	var font string

	if !ctx.Client.IsGuest {
		metrics.DbRequests.Inc()
		query := `SELECT color, font FROM profiles WHERE id = $1`

		err := ctx.Chub.Db.QueryRow(context.Background(), query, ctx.Client.CurrUsrID).Scan(&color, &font)

		if err != nil {
			fmt.Printf(NOT_FOUND_DB)
			return
		}
		metrics.DbRequestsSucessful.Inc()
	} else {
		color = DEFAULT_COLOR
		font = "normal"
	}

	ctx.Client.CurrentRoom = ctx.Client.Hub.CreateRoom(false)
	base := ctx.Client.CurrentRoom.GetBase()
	err := base.AddPlayer(*ctx.Client.CurrUsrID, *ctx.Client.CurrUsrName, ctx.Client.Conn, color, font)
	if err != nil {
		fmt.Println("DEBUG: ", err)
		return
	}

	fmt.Printf("DEBUG: %s success\n", msg.Type)

	base.MessageChan <- gamemanager.Notification{
		PlayerID: *ctx.Client.CurrUsrID,
		Data: map[string]interface{}{
			"type":		"room_created",
			"code":		base.ID,
			"players":	[]map[string]interface{}{
				{
					"id":	ctx.Client.CurrUsrID,
					"name":	ctx.Client.CurrUsrName,
					"host":	true,
				},
			},
		},
	}
	
	base.BroadcastLobbyState()
	
}

func (d *Dispatcher) HandleJoinRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	var color string
	var font string
	if !ctx.Client.IsGuest {
		metrics.DbRequests.Inc()
		query := `SELECT color, font FROM profiles WHERE id = $1`

		err := ctx.Chub.Db.QueryRow(context.Background(), query, ctx.Client.CurrUsrID).Scan(&color, &font)

		if err != nil {
			fmt.Printf(NOT_FOUND_DB)
			return
		}
		metrics.DbRequestsSucessful.Inc()
	} else {
		color = DEFAULT_COLOR
		font = "normal"
	}

	base := ctx.Client.CurrentRoom.GetBase()
	err := base.AddPlayer(*ctx.Client.CurrUsrID, *ctx.Client.CurrUsrName, ctx.Client.Conn, color, font)
	if err != nil {
		fmt.Println("AddPlayer error:", err)
		return
	}
	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {
		classicRoom.SendSystemMsg(fmt.Sprintf("%s join the lobby !", *ctx.Client.CurrUsrName))
	}

	base.BroadcastLobbyState()

}

func (d *Dispatcher) HandleJoinGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }

	base := ctx.Client.CurrentRoom.GetBase()
	base.JoinGame(*ctx.Client.CurrUsrID, ctx.Client.Conn)
	fmt.Printf("DEBUG join_game: code='%s' user='%s'\n", msg.Code, *ctx.Client.CurrUsrName)
	
	var task gamemanager.GameStateRecord
	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {
		task = classicRoom.GetPlayerTask(*ctx.Client.CurrUsrID)
	}
	writeErr := ctx.Client.Conn.WriteJSON(task)
	if writeErr != nil {
		fmt.Printf(ERROR_WRITE_WS, writeErr)
	}
}

func (d *Dispatcher) HandleLeaveLobby(ctx *WSContext, msg Message) { 
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }


	base := ctx.Client.CurrentRoom.GetBase()
	if base.Status != gamemanager.StateWaiting { return }

	fmt.Printf("DEBUG leave_lobby: code='%s' user='%s'\n", msg.Code, *ctx.Client.CurrUsrName)
	isHost := false

	if p, err := base.GetPlayer(*ctx.Client.CurrUsrID); err == nil {
		isHost = p.IsHost
	}

	base.RemovePlayer(*ctx.Client.CurrUsrID)
	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {

		classicRoom.SendSystemMsg(fmt.Sprintf("%s leave the room !", *ctx.Client.CurrUsrName))

		if len(base.Players) == 0 {
			classicRoom.MessageChan <- gamemanager.Notification{
				End: true,
			}
			ctx.Client.Hub.DeleteRoom(base.ID)
			return
		}
	}

	if isHost {
		base.TransferHost()
	}
	base.BroadcastLobbyState()
}

func (d *Dispatcher) HandleLeaveGame(ctx *WSContext, msg Message) { 
	fmt.Printf("DEBUG: leave_game msg %s\n", msg.Code)
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {
		del := classicRoom.LeaveGame(*ctx.Client.CurrUsrID)
		if del {
			ctx.Client.Hub.DeleteRoom(msg.Code)
			fmt.Printf("DEBUG: DELETE ROOM nobody is in\n")
		}
	}
}

func (d *Dispatcher) HandlePrompt(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	data := map[string]interface{}{
		"type": "prompt",
		"prompt": msg.Prompt,
	}
	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {
		err := classicRoom.SubmiteAction(*ctx.Client.CurrUsrID, data, true)
		if (err != nil) {
			fmt.Printf("Error: Submited prompt: %v\n", err)
		}
	}
}

func (d *Dispatcher) HandleDraw(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	data := map[string]interface{}{
		"type": "draw",
		"drawing": msg.Drawing,
	}
	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {
		fmt.Printf("DEBUG: draw_submitted code = %s\n", msg.Code)
		err := classicRoom.SubmiteAction(*ctx.Client.CurrUsrID, data, true)
		if (err != nil) {
			fmt.Printf("ERROR: Submited draw: %v\n", err)
		}
	}
}

func (d *Dispatcher) HandleGuess(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	data := map[string]interface{}{
		"type": "guess",
		"guess": msg.Guess,
	}
	fmt.Printf("DEGUB: %s\n", msg.Type)
	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {
		err := classicRoom.SubmiteAction(*ctx.Client.CurrUsrID, data, true)
		if (err != nil) {
			fmt.Printf("ERROR: Submited guess: %v\n", err)
		}
	}
}


func (d *Dispatcher) HandleStartGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeHasRoomCode)) { return }

	base := ctx.Client.CurrentRoom.GetBase()
	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {
		err := classicRoom.StartGame()
		if err != nil {
			writeErr := ctx.Client.Conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
			if writeErr != nil {
				fmt.Printf(ERROR_WRITE_WS, writeErr)
			}
			return
		}
	}
	fmt.Printf("DEBUG: start_game\n")
	base.BroadcastToAll(map[string]interface{}{
		"type": "start_game",
		"code": base.ID,
	})
}

func (d *Dispatcher) HandleChat(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeIsValidChat)) { return }

	fmt.Printf("DEBUG: chat_message\n")
	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {
		classicRoom.BroadcastChat(*ctx.Client.CurrUsrID, msg.Text)
	}
}

func (d *Dispatcher) HandleAIDraw(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeHasRoomCode)) { return }

	if RoomIA, ok := ctx.Client.CurrentRoom.(*gamemanager.AIRoom); ok {
		RoomIA.SubmitDrawing(*ctx.Client.CurrUsrID, msg.Drawing, msg.Title, msg.Description)
	}
}

func (d *Dispatcher) HandleAIVote(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeHasRoomCode)) { return }

	if RoomIA, ok := ctx.Client.CurrentRoom.(*gamemanager.AIRoom); ok {
		RoomIA.SubmitVotes(*ctx.Client.CurrUsrID, msg.Votes)
	}
}


func (d *Dispatcher) HandleCreateAIRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) { return }

	var color string
	var font string
	if !ctx.Client.IsGuest {
		metrics.DbRequests.Inc()
		query := `SELECT color, font FROM profiles WHERE id = $1`

		err := ctx.Chub.Db.QueryRow(context.Background(), query, ctx.Client.CurrUsrID).Scan(&color, &font)

		if err != nil {
			fmt.Printf(NOT_FOUND_DB)
			return
		}
		metrics.DbRequestsSucessful.Inc()
	} else {
		color = DEFAULT_COLOR
		font = "normal"
	}
	newRoom := ctx.Client.Hub.CreateRoom(true)
	base := newRoom.GetBase()
	err := base.AddPlayer(*ctx.Client.CurrUsrID, *ctx.Client.CurrUsrName, ctx.Client.Conn, color, font)
	if err != nil { return }
	fmt.Printf("DEGUB: AI_ROOM created\n")	

	base.MessageChan <- gamemanager.Notification{
		PlayerID: *ctx.Client.CurrUsrID,
		Data: map[string]interface{}{
			"type": "ai_room_created",
			"code": base.ID,
			"palyers": []map[string]interface{}{
				{
					"id": *ctx.Client.CurrUsrID,
					"name": *ctx.Client.CurrUsrName,
					"host": true,
				},
			},
		},
	}
	fmt.Printf("DEBUG: %s\n", msg.Type)
	base.BroadcastLobbyState()
}

func (d *Dispatcher) HandleStartAIGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeHasRoomCode)) { return }

	room, _ := ctx.Client.Hub.GetRoom(msg.Code)
	base := room.GetBase()
	if base.Status != gamemanager.StateAIWaiting { return }

	prompt, err := gamemanager.CallAI("")
	if err != nil {
		fmt.Println("callAI error:", err)
		prompt = "Error API : add Credits"
	}


	base.BroadcastToAll(map[string]interface{}{
		"type": "start_ai_game",
		"code": base.ID,
	})
	if RoomIA, ok := ctx.Client.CurrentRoom.(*gamemanager.AIRoom); ok {
		fmt.Printf("DEBUG: RunLoopAI\n")
		go RoomIA.RunAIGameLoop(prompt)
	}
}


func (d *Dispatcher) HandleLeaveAILobby(ctx *WSContext, msg Message) { 
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }


	base := ctx.Client.CurrentRoom.GetBase()
	if base.Status != gamemanager.StateAIWaiting { return }

	isHost := false

	if p, err := base.GetPlayer(*ctx.Client.CurrUsrID); err == nil {
		isHost = p.IsHost
	}

	base.RemovePlayer(*ctx.Client.CurrUsrID)
	if AIRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.AIRoom); ok {
		AIRoom.SendSystemMsg(fmt.Sprintf("%s leave the AI room !", *ctx.Client.CurrUsrName))
	}

	if len(base.Players) == 0 {
		ctx.Client.Hub.DeleteRoom(base.ID)
		return
	}
	if isHost {
		base.TransferHost()
	}
	base.BroadcastLobbyState()
}

func (d *Dispatcher) HandleChatAI(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeIsValidChat)) { return }

	fmt.Printf("DEBUG: chat_message\n")
	if AIRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.AIRoom); ok {
		AIRoom.BroadcastChat(*ctx.Client.CurrUsrID, msg.Text)
	}
}

func (d *Dispatcher) HandleLeaveAIGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }

	RoomIA := ctx.Client.CurrentRoom.(*gamemanager.AIRoom)
	del := RoomIA.LeaveGame(*ctx.Client.CurrUsrID)
	if del {
		ctx.Client.Hub.DeleteRoom(msg.Code)
	}
}

func (d *Dispatcher) HandleJoinAIGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }

	base := ctx.Client.CurrentRoom.GetBase()
	base.JoinGame(*ctx.Client.CurrUsrID, ctx.Client.Conn)
	fmt.Printf("DEBUG join_game: code='%s' user='%s'\n", msg.Code, *ctx.Client.CurrUsrName)

	RoomIA := ctx.Client.CurrentRoom.(*gamemanager.AIRoom)
	if (base.Status == gamemanager.StateAIDrawing) {
		base.MessageChan <- gamemanager.Notification{
			PlayerID: *ctx.Client.CurrUsrID,
			Data: map[string]interface{}{
				"type":		"ai_game_state",
				"phase":	"draw",
				"prompt":	RoomIA.Prompt,
				"room":		base.ID,
			},
		}
	}
}

func (d *Dispatcher) HandleJoinAIRoom(ctx *WSContext, msg Message) {
	fmt.Println("TIGGER: HandleJoinAIRoom")
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	var color string
	var font string
	if !ctx.Client.IsGuest {
		metrics.DbRequests.Inc()
		query := `SELECT color, font FROM profiles WHERE id = $1`

		err := ctx.Chub.Db.QueryRow(context.Background(), query, ctx.Client.CurrUsrID).Scan(&color, &font)

		if err != nil {
			fmt.Printf(NOT_FOUND_DB)
			return
		}
		metrics.DbRequestsSucessful.Inc()
	} else {
		color = DEFAULT_COLOR
		font = "normal"
	}
	base := ctx.Client.CurrentRoom.GetBase()
	err := base.AddPlayer(*ctx.Client.CurrUsrID, *ctx.Client.CurrUsrName, ctx.Client.Conn, color, font)
	if err != nil {
		fmt.Println("AddPlayer error:", err)
		return
	}
	if AIRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.AIRoom); ok {
		AIRoom.SendSystemMsg(fmt.Sprintf("%s join the lobby !", *ctx.Client.CurrUsrName))
	}

	base.BroadcastLobbyState()

}
