package handler

import (
	"fmt"
	"errors"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"

	"github.com/realgetOff/We_Plaid_Guilty/internal/db"
	"github.com/realgetOff/We_Plaid_Guilty/internal/webutil"
)

func (d *Dispatcher) HandleGetFriend(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) { 
		return
	}
	fmt.Println("DEBUG: HandleGetFriend triggered!")

	userID := msg.ID
	if userID == "" {
		userID = *ctx.Client.CurrUsrID
	}

	if ctx.Client.IsGuest {
		_ = ctx.Client.Conn.WriteJSON(FriendsListResponse{
			Type:			"friends_list",
			GuestNoFriends:	true,
		})
		return
	}


	qAccepted := `SELECT u.id, u.username FROM users u
		JOIN friends f ON (u.id = f.requester_id OR u.id = f.addressee_id)
		WHERE (f.requester_id = $1 OR f.addressee_id = $1) AND u.id != $1 AND f.status = 'accepted'`

	rows, err := db.DBQueryRows(ctx.Chub.Db, qAccepted, userID)

	if err != nil {
		fmt.Printf("Failed to open accepted friends :: %v\n", err)
		return
	}

	friends := scanFriendRows(ctx, rows)

	qIn := `SELECT u.id, u.username FROM users u
		JOIN friends f ON u.id = f.requester_id
		WHERE f.addressee_id = $1 AND f.status = 'pending'`
	rowsIn, err := db.DBQueryRows(ctx.Chub.Db, qIn, userID)

	if err != nil {
		fmt.Printf("Failed pending_in :: %v\n", err)
		return
	}

	pendingIn := scanFriendRows(ctx, rowsIn)


	qOut := `SELECT u.id, u.username FROM users u
		JOIN friends f ON u.id = f.addressee_id
		WHERE f.requester_id = $1 AND f.status = 'pending'`
	rowsOut, err := db.DBQueryRows(ctx.Chub.Db, qOut, userID)
	if err != nil {
		fmt.Printf("Failed pending_out :: %v\n", err)
		return
	}

	pendingOut := scanFriendRows(ctx, rowsOut)

	err = ctx.Client.Conn.WriteJSON(FriendsListResponse{
		Type:        "friends_list",
		Friends:     friends,
		PendingIn:   pendingIn,
		PendingOut:  pendingOut,
	})
	if err != nil {
		fmt.Printf("failed to send friends list: %v\n", err)
	}
}

func (d* Dispatcher) HandleRemoveFriend(ctx *WSContext, msg Message) {
	fmt.Println("DEBUG: HandleRemoveFriend triggered!")
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) {
		return
	}

	query := `
	DELETE FROM friends 
	WHERE (requester_id = $1 AND addressee_id = (SELECT id FROM users WHERE username = $2))
		OR (requester_id = (SELECT id FROM users WHERE username = $2) AND addressee_id = $1)
	RETURNING (SELECT id FROM users WHERE username = $2);
	`

	var friend_id string
	err := db.DBQuery(ctx.Chub.Db, query, []any{ctx.Client.CurrUsrID, msg.Username}, &friend_id)
	if (err != nil) {
		fmt.Printf("Friend remove failed :: %v\n", err)
		return
	}

	response := FriendsListResponse {
		Type: "friend_removed",
		FriendID: friend_id,
	}

	err = ctx.Client.Conn.WriteJSON(response)
	if err != nil {
		fmt.Printf("Failed to send friend_id on removal: %v", err)
	}

	response = FriendsListResponse {
		Type: "friend_removed",
		FriendID: *ctx.Client.CurrUsrID,
	}

	if peer := ctx.Chub.Clients[friend_id]; peer != nil {
		err = peer.Conn.WriteJSON(response)
		if err != nil {
			fmt.Printf("Failed to send friend_id to peer on removal: %v\n", err)
		}
	}

}

func (d *Dispatcher) HandleInviteFriend(ctx *WSContext, msg Message) {
	if !RunPipeLine(ctx, msg, d.PipeIsAuth) {
		return
	}
	if ctx.Client.IsGuest {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "invite_sent", "success": false, "error": "Guests cannot send invites.",
		})
		return
	}
	to := strings.TrimSpace(msg.To)
	code := strings.ToUpper(strings.TrimSpace(msg.Code))
	if to == "" || code == "" {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "invite_sent", "success": false, "error": "Missing recipient or room code.",
		})
		return
	}
	var targetID string

	err := db.DBQuery(ctx.Chub.Db,  `SELECT id FROM users WHERE username = $1`, []any{to}, &targetID)
	if err != nil {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "invite_sent", "success": false, "error": "User not found.",
		})
		return
	}

	target := ctx.Chub.Clients[targetID]
	if target == nil {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "invite_sent", "success": false, "error": "User is offline.",
		})
		return
	}
	payload := map[string]interface{}{
		"type": "game_invite",
		"from": *ctx.Client.CurrUsrName,
		"code": code,
	}
	if msg.IsAI {
		payload["is_ai"] = true
	}
	err = target.Conn.WriteJSON(payload)
	if err != nil {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "invite_sent", "success": false, "error": "Failed to deliver invite.",
		})
		return
	}
	_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{"type": "invite_sent", "success": true})
}

func (d *Dispatcher) HandleAcceptFriend(ctx *WSContext, msg Message) {
	if !RunPipeLine(ctx, msg, d.PipeIsAuth) {
		return
	}
	if ctx.Client.IsGuest {
		return
	}
	me := *ctx.Client.CurrUsrID
	myName := *ctx.Client.CurrUsrName
	otherUsername := strings.TrimSpace(msg.Username)
	if otherUsername == "" {
		return
	}
	var requesterID string

	err := db.DBQuery(ctx.Chub.Db,  `UPDATE friends SET status = 'accepted', updated_at = NOW()
		WHERE addressee_id = $1 AND requester_id = (SELECT id FROM users WHERE username = $2) AND status = 'pending'
		RETURNING requester_id`, []any{me, otherUsername}, &requesterID)

	if err != nil {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "friend_accept_failed", "success": false, "error": "No pending request from that user.",
		})
		return
	}

	var requesterName string
	_ = db.DBQuery(ctx.Chub.Db, `SELECT username FROM users WHERE id = $1`, []any{requesterID}, &requesterName)

	d.broadcastFriendAdded(ctx, requesterID, requesterName, me, myName)
}

func (d *Dispatcher) HandleAddFriend(ctx *WSContext, msg Message) {
	if !RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeIsGuest){
		return
	}
	fmt.Println("DEBUG: HandleAddFriend triggered!")

	me := *ctx.Client.CurrUsrID
	myName := *ctx.Client.CurrUsrName
	targetName := strings.TrimSpace(msg.Username)
	if targetName == "" || targetName == myName {
		return
	}
	var targetID string
	var targetType string

	err := db.DBQuery(ctx.Chub.Db,`SELECT id, type FROM users WHERE username = $1`, []any{targetName}, &targetID, &targetType)
	if err != nil {
		fmt.Printf("Friend add: user not found :: %v\n", err)
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "friend_add_failed", "success": false, "error": "User not found.",
		})
		return
	}

	if targetType == "guest" {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "friend_add_failed", "success": false, "error": "Cannot add guest accounts as friends.",
		})
		return
	}
	var st string
	var reqID string

	err = db.DBQuery(ctx.Chub.Db,
		`SELECT f.status::text, f.requester_id::text FROM friends f WHERE
	 	(f.requester_id = $1::uuid AND f.addressee_id = $2::uuid)
	 	OR (f.requester_id = $2::uuid AND f.addressee_id = $1::uuid)`,
		[]any{me, targetID}, &st, &reqID)

		if errors.Is(err, pgx.ErrNoRows){
			err = db.DBQuery(ctx.Chub.Db,
			`INSERT INTO friends (requester_id, addressee_id, status) VALUES ($1::uuid, $2::uuid, 'pending')`,
			[]any{me, targetID})

			if err != nil {
				fmt.Printf("Friend invite insert failed :: %v\n", err)
				_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
					"type": "friend_add_failed", "success": false, "error": "Could not send friend request.",
				})
				return
			}

			addresseeOnline := ctx.Chub.Clients[targetID] != nil
			if c := ctx.Chub.Clients[targetID]; c != nil {
				_ = c.Conn.WriteJSON(map[string]interface{}{
					"type": "friend_request",
					"user": Friend{ID: me, Username: myName, Online: true},
				})
			}
			_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
				"type": "friend_request_sent",
				"user": Friend{ID: targetID, Username: targetName, Online: addresseeOnline},
			})

			return
		}
		if err != nil {
			fmt.Printf("Friend lookup failed :: %v\n", err)
			_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
				"type": "friend_add_failed", "success": false, "error": "Database error while looking up friend.",
			})
			return
		}

		if st == "accepted" {
			_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
				"type": "friend_add_failed", "success": false, "error": "Already friends.",
			})
			return
		}

		if st == "pending" && reqID == me {
				_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
					"type": "friend_add_failed", "success": false, "error": "Friend request already sent.",
				})
				return
			}
			err = db.DBQuery(ctx.Chub.Db,
			`UPDATE friends SET status = 'accepted', updated_at = NOW()
			WHERE requester_id = $1::uuid AND addressee_id = $2::uuid AND status = 'pending'`,
			[]any{targetID, me})

			if err != nil {
				fmt.Printf("mutual accept failed: %v\n", err)
				_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
					"type": "friend_add_failed", "success": false, "error": "Could not accept friend request.",
				})
				return
			}

			var requesterName string

			_ = db.DBQuery(ctx.Chub.Db, `SELECT username FROM users WHERE id = $1`, []any{targetID}, &requesterName)

			d.broadcastFriendAdded(ctx, targetID, requesterName, me, myName)
}

func (d* Dispatcher) HandleGetProfile(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) {
		return
	}
	fmt.Println("DEBUG: HandleGetProfile triggered!")

	var user ProfileUser
	var profileID string

	query := `SELECT p.id, p.display_name, p.color, p.font
				FROM profiles p
				INNER JOIN users u ON p.id = u.id
				WHERE u.username = $1;`

	err := db.DBQuery(ctx.Chub.Db, query, []any{msg.Username}, &profileID,
		&user.Username, &user.Style.Color, &user.Style.Font)

	var response ProfileResponse
	
	response.Type = "profile_data"

	user.Online = (ctx.Chub.Clients[profileID] != nil) 
	user.IsGuest = ctx.Client.IsGuest

	if (err != nil) {
		response.Success = false
	} else {
		response.Success = true
		response.User = user
		response.IsCaller = (profileID == *ctx.Client.CurrUsrID)
	}

	err = ctx.Client.Conn.WriteJSON(response)
	if err != nil {
		fmt.Printf("failed to send profile data: %v\n", err)
	}
}

func (d* Dispatcher) HandleProfileUpdate(ctx *WSContext, msg Message) {
	fmt.Println("DEBUG: HandleProfileUpdate triggered!")
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) {
		return
	}
	if ctx.Client.IsGuest {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "profile_updated", "success": false,
			"error":  "Guest accounts cannot edit their profile.",
		})
		return
	}

	query :=	`
				UPDATE profiles
				SET color = $2, font = $3, display_name = $4
				WHERE id = $1
				`
	err := db.DBQuery(ctx.Chub.Db, query, []any{ ctx.Client.CurrUsrID, msg.Style.Color, msg.Style.Font, msg.Username })

	if (err != nil ) { 
		fmt.Printf("FAILED TO UPDATE THE PROFILE TABLE FOR USER %v : %v\n", *ctx.Client.CurrUsrName, err)
		return
	}

	usrnmQuery :=	`
					UPDATE users
					SET username = $2
					WHERE id = $1;
					`

	err = db.DBQuery(ctx.Chub.Db, query, []any{ usrnmQuery, ctx.Client.CurrUsrID, msg.Username })

	if (err != nil ) { 
		fmt.Printf("FAILED TO UPDATE THE USERNAME FOR USER %v : %v\n", *ctx.Client.CurrUsrName, err)
		return
	}

	oldUsername := *ctx.Client.CurrUsrName
	*ctx.Client.CurrUsrName = msg.Username

	fmt.Printf("Session username updated from %s to %s (ID: %s)\n", oldUsername, *ctx.Client.CurrUsrName, *ctx.Client.CurrUsrID)

	newToken, err := webutil.GenerateJWT(*ctx.Client.CurrUsrID, *ctx.Client.CurrUsrName)
	if err != nil {
		fmt.Printf("Failed to generate new JWT: %v\n", err)
		return
	}

var response ProfileResponse

response.Type = "profile_updated"
response.Success = true
response.User.Username = *ctx.Client.CurrUsrName
response.User.Style.Color = msg.Style.Color
response.User.Style.Font = msg.Style.Font

writeErr := ctx.Client.Conn.WriteJSON(map[string]interface{}{
	"type": "profile_updated",
	"success": true,
	"user": map[string]interface{}{
		"username": *ctx.Client.CurrUsrName,
		"style": map[string]interface{}{
			"color": msg.Style.Color,
			"font": msg.Style.Font,
		},
	},
	"token": newToken,
})
if writeErr != nil {
	fmt.Printf(ERROR_WRITE_WS, writeErr)
}
}



func (d *Dispatcher) HandleAuth(ctx *WSContext, msg Message) {
	claims, err := webutil.ValidateAndGetClaims(msg.Token)
	if err != nil {
		fmt.Println("WS Auth Failed:", err)
		writeErr := ctx.Client.Conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(4000, "token expired"))
		if writeErr != nil {
			fmt.Printf(ERROR_WRITE_WS, writeErr)
		}
		return
	}
	
	ctx.Client.CurrUsrName = &claims.Username
	ctx.Client.CurrUsrID = &claims.UserID

	var clientType string

	_ = db.DBQuery(ctx.Chub.Db, `SELECT type FROM users WHERE id = $1`,
		[]any{ claims.UserID }, &clientType)

	
	ctx.Client.IsGuest = (clientType == "guest")

	ctx.Chub.Mu.Lock()
	ctx.Chub.Clients[claims.UserID] = ctx.Client
	ctx.Chub.Mu.Unlock()

	_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
		"type":     "auth_ok",
		"is_guest": ctx.Client.IsGuest,
	})

	fmt.Printf("WS Authenticated: %s (ID: %s) guest=%v\n", *ctx.Client.CurrUsrName, *ctx.Client.CurrUsrID, ctx.Client.IsGuest)
}
