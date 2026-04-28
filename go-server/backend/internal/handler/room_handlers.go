package handler

import (
	"fmt"

	"github.com/realgetOff/We_Plaid_Guilty/internal/gamemanager"
	"github.com/realgetOff/We_Plaid_Guilty/internal/db"

)

func (d *Dispatcher) HandleCreateRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) { return }

	fmt.Printf("DEBUG: create_room\n")
	var color string
	var font string

	if !ctx.Client.IsGuest {
		query := `SELECT color, font FROM profiles WHERE id = $1`

		err := db.DBQuery(ctx.Chub.Db, query, []any{ ctx.Client.CurrUsrID }, &color, &font)
		

		if err != nil {
			fmt.Printf(NOT_FOUND_DB)
			return
		}
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
		query := `SELECT color, font FROM profiles WHERE id = $1`

		err := db.DBQuery(ctx.Chub.Db, query, []any{ ctx.Client.CurrUsrID }, &color, &font)

		if err != nil {
			fmt.Printf(NOT_FOUND_DB)
			return
		}
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
		classicRoom.SendSystemMsg(fmt.Sprintf("%s joined the lobby!", *ctx.Client.CurrUsrName))
	}

	base.BroadcastLobbyState()

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

		classicRoom.SendSystemMsg(fmt.Sprintf("%s left the room!", *ctx.Client.CurrUsrName))

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

func (d *Dispatcher) HandleCreateAIRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) { return }

	var color string
	var font string
	if !ctx.Client.IsGuest {
		query := `SELECT color, font FROM profiles WHERE id = $1`

		err := db.DBQuery(ctx.Chub.Db, query, []any{ ctx.Client.CurrUsrID }, &color, &font)

		if err != nil {
			fmt.Printf(NOT_FOUND_DB)
			return
		}
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
		AIRoom.SendSystemMsg(fmt.Sprintf("%s left the AI room!", *ctx.Client.CurrUsrName))
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


func (d *Dispatcher) HandleJoinAIRoom(ctx *WSContext, msg Message) {
	fmt.Println("TIGGER: HandleJoinAIRoom")
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	var color string
	var font string
	if !ctx.Client.IsGuest {
		query := `SELECT color, font FROM profiles WHERE id = $1`

		err := db.DBQuery(ctx.Chub.Db, query, []any{ ctx.Client.CurrUsrID }, &color, &font)

		if err != nil {
			fmt.Printf(NOT_FOUND_DB)
			return
		}
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
		AIRoom.SendSystemMsg(fmt.Sprintf("%s joined the lobby!", *ctx.Client.CurrUsrName))
	}

	base.BroadcastLobbyState()

}
