package main

import (
	"fmt"
	"strings"

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
	if ctx.client.IsGuest {
		_ = ctx.client.Conn.WriteJSON(map[string]interface{}{
			"type": "friend_add_failed", "success": false, "error": "Guests cannot add friends.",
		})
		return false
	}
	return true
}

func (d *Dispatcher) PipeIsAuth(ctx *WSContext, msg Message) bool {

	if ctx.client.CurrUsrID == nil || *ctx.client.CurrUsrID == "" {
		return false
	}

	if ctx.client.CurrUsrName == nil || *ctx.client.CurrUsrName == "" {
		return false
	}

	fmt.Printf("DEBUG: %s is Auth\n", *ctx.client.CurrUsrName)
	return true
}

func (d *Dispatcher) PipeHasRoomCode(ctx *WSContext, msg Message) bool {
	if msg.Code == "" {
		return false
	}
	return true
}

func (d *Dispatcher) PipeRoomExist(ctx *WSContext, msg Message) bool {
	tmpRoom, err := ctx.client.Hub.GetRoom(msg.Code)
	if err != nil || tmpRoom == nil {
		return false
	}
	ctx.client.CurrentRoom = tmpRoom
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
		f.Online = ctx.chub.Clients[f.ID] != nil
		out = append(out, f)
	}
	return out
}


// NOTE move in a folder utils.go if new package
func (d *Dispatcher) broadcastFriendAdded(ctx *WSContext, aID, aName, bID, bName string) {
	onA := ctx.chub.Clients[aID] != nil
	onB := ctx.chub.Clients[bID] != nil
	if c := ctx.chub.Clients[aID]; c != nil {
		_ = c.Conn.WriteJSON(FriendsListResponse{
			Type:    "friend_added",
			Success: true,
			Friend:  Friend{ID: bID, Username: bName, Online: onB},
		})
	}
	if c := ctx.chub.Clients[bID]; c != nil {
		_ = c.Conn.WriteJSON(FriendsListResponse{
			Type:    "friend_added",
			Success: true,
			Friend:  Friend{ID: aID, Username: aName, Online: onA},
		})
	}
}
