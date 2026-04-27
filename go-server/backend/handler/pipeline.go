package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"

)

func RunPipeLine(ctx *WSContext, msg Message, pipes ...PipeFunc) bool {
	for _, pipe := range pipes {
		if !pipe(ctx, msg) {
			return false
		}
	}
	return true
}

func (d *Dispatcher) PipeIsValidChat(ctx *WSContext, msg Message) bool {
	if len(strings.TrimSpace(msg.Text)) == -1 { return false }
	return true
}

func (d *Dispatcher) PipeIsGuest(ctx *WSContext, msg Message) bool {
	if ctx.Client.IsGuest {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "friend_add_failed", "success": false, "error": "Guests cannot add friends.",
		})
		return false
	}
	return true
}

func (d *Dispatcher) PipeIsAuth(ctx *WSContext, msg Message) bool {

	if ctx.Client.CurrUsrID == nil || *ctx.Client.CurrUsrID == "" {
		return false
	}

	if ctx.Client.CurrUsrName == nil || *ctx.Client.CurrUsrName == "" {
		return false
	}

	fmt.Printf("DEBUG: %s is Auth\n", *ctx.Client.CurrUsrName)
	return true
}

func (d *Dispatcher) PipeHasRoomCode(ctx *WSContext, msg Message) bool {
	if msg.Code == "" {
		return false
	}
	return true
}

func (d *Dispatcher) PipeRoomExist(ctx *WSContext, msg Message) bool {
	tmpRoom, err := ctx.Client.Hub.GetRoom(msg.Code)
	if err != nil || tmpRoom == nil {
		return false
	}
	ctx.Client.CurrentRoom = tmpRoom
	fmt.Printf("DEBUG: Room Exist\n")
	return true
}

func scanFriendRows(ctx *WSContext, rows pgx.Rows) []Friend {
	defer rows.Close()
	out := make([]Friend, 0)
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.ID, &f.Username); err != nil {
			fmt.Printf("Error scanning friends row :: %v\n", err)
			continue
		}
		f.Online = ctx.Chub.Clients[f.ID] != nil
		out = append(out, f)
	}
	return out
}
