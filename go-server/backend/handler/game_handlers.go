package handler

import (
	"fmt"

	"main.go/gamemanager"
)

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

func (d *Dispatcher) HandleLeaveAIGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }

	RoomIA := ctx.Client.CurrentRoom.(*gamemanager.AIRoom)
	del := RoomIA.LeaveGame(*ctx.Client.CurrUsrID)
	if del {
		ctx.Client.Hub.DeleteRoom(msg.Code)
	}
}

func (d *Dispatcher) HandlePrompt(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	data := map[string]interface{}{
		"type": "prompt",
		"prompt": msg.Prompt,
	}
	if classicRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.Room); ok {
		err := classicRoom.SubmitAction(*ctx.Client.CurrUsrID, data, true)
		if (err != nil) {
			fmt.Printf("Error: Submitted prompt: %v\n", err)
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
		err := classicRoom.SubmitAction(*ctx.Client.CurrUsrID, data, true)
		if (err != nil) {
			fmt.Printf("ERROR: Submitted draw: %v\n", err)
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
		err := classicRoom.SubmitAction(*ctx.Client.CurrUsrID, data, true)
		if (err != nil) {
			fmt.Printf("ERROR: submitted guess: %v\n", err)
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


func (d *Dispatcher) HandleChatAI(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeIsValidChat)) { return }

	fmt.Printf("DEBUG: chat_message\n")
	if AIRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.AIRoom); ok {
		AIRoom.BroadcastChat(*ctx.Client.CurrUsrID, msg.Text)
	}
}
