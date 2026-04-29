package handler

import (
	"github.com/realgetOff/We_Plaid_Guilty/internal/db"
)

func (d *Dispatcher) broadcastFriendAdded(ctx *WSContext, aID, aName, bID, bName string) {
	onA := ctx.Chub.Clients[aID] != nil
	onB := ctx.Chub.Clients[bID] != nil
	if c := ctx.Chub.Clients[aID]; c != nil {
		_ = c.Conn.WriteJSON(FriendsListResponse{
			Type:    "friend_added",
			Success: true,
			Friend:  Friend{ID: bID, Username: bName, Online: onB},
		})
	}
	if c := ctx.Chub.Clients[bID]; c != nil {
		_ = c.Conn.WriteJSON(FriendsListResponse{
			Type:    "friend_added",
			Success: true,
			Friend:  Friend{ID: aID, Username: aName, Online: onA},
		})
	}
}

type FriendOnlineStatus struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Online   bool   `json:"online"`
}

func NotifyFriendsStatus(ctx *WSContext, userID string, username string, isOnline bool) {
	qAccepted := `SELECT u.id FROM users u
		JOIN friends f ON (u.id = f.requester_id OR u.id = f.addressee_id)
		WHERE (f.requester_id = $1 OR f.addressee_id = $1)
		AND u.id != $1 AND f.status = 'accepted'`

	rows, err := db.DBQueryRows(ctx.Chub.Db, qAccepted, userID)
	if err != nil {
		return
	}
	defer rows.Close()

	statusMsg := FriendOnlineStatus{
		Type:     "friend_online_status",
		Username: username,
		Online:   isOnline,
	}

	for rows.Next() {
		var friendID string
		if err := rows.Scan(&friendID); err != nil {
			continue
		}

		ctx.Chub.Mu.RLock()
		if targetClient, isConnected := ctx.Chub.Clients[friendID]; isConnected {
			_ = targetClient.Conn.WriteJSON(statusMsg)
		}
		ctx.Chub.Mu.RUnlock()
	}
}

func HandleUserConnect(ctx *WSContext, userID string, username string) {
	NotifyFriendsStatus(ctx, userID, username, true)
}

func HandleUserDisconnect(ctx *WSContext, userID string, username string) {
	NotifyFriendsStatus(ctx, userID, username, false)
}