package handler

import (
	"fmt"
	"context"
	"errors"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"

	"main.go/gamemanager"
	"main.go/metrics"

)

func NewDispatcher() *Dispatcher {
	d:= &Dispatcher{
		handlers: make(map[string]HandleFunc),
	}
	d.handlers["get_profile"] = d.HandleGetProfile
	d.handlers["update_profile"] = d.HandleProfileUpdate
	d.handlers["add_friend"] = d.HandleAddFriend
	d.handlers["get_friends"] = d.HandleGetFriend
	d.handlers["remove_friend"] = d.HandleRemoveFriend
	d.handlers["accept_friend"] = d.HandleAcceptFriend
	d.handlers["invite_friend"] = d.HandleInviteFriend
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
	d.handlers["create_ai_room"] = d.HandleCreateAIRoom
	d.handlers["join_ai_room"] = d.HandleJoinAIRoom
	d.handlers["join_ai_game"] = d.HandleJoinAIGame
	d.handlers["leave_ai_room"] = d.HandleLeaveAILobby
	d.handlers["leave_ai_game"] = d.HandleLeaveAIGame
	d.handlers["ai_chat_message"] = d.HandleChatAI
	d.handlers["start_ai_game"] = d.HandleStartAIGame
	d.handlers["ai_drawing_submitted"] = d.HandleAIDraw
	d.handlers["ai_votes_submitted"] = d.HandleAIVote

	return d
}

func (d *Dispatcher) Dispatch(ctx *WSContext, msg Message) {
	handler, ok := d.handlers[msg.Type]
	if !ok {
		fmt.Printf("DEBUG: %s not found\n", msg.Type)
		writeErr := ctx.Client.Conn.WriteJSON(map[string]string{"type": "error", "message": "Action not recognized by the server",})
		if writeErr != nil {
			fmt.Printf(ERROR_WRITE_WS, writeErr)
		}
		return
	}
	fmt.Printf("MESSAGE: %s\n", msg.Type)
	handler(ctx, msg)
}


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
			Type:           "friends_list",
			GuestNoFriends: true,
		})
		return
	}

	// PROMETHEUS
	metrics.DbRequests.Inc()

	qAccepted := `SELECT u.id, u.username FROM users u
		JOIN friends f ON (u.id = f.requester_id OR u.id = f.addressee_id)
		WHERE (f.requester_id = $1 OR f.addressee_id = $1) AND u.id != $1 AND f.status = 'accepted'`

	rows, err := ctx.Chub.Db.Query(context.Background(), qAccepted, userID)
	if err != nil {
		fmt.Printf("Failed to open accepted friends :: %v\n", err)
		return
	}

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()

	friends := scanFriendRows(ctx, rows)

	// PROMETHEUS
	metrics.DbRequests.Inc()

	qIn := `SELECT u.id, u.username FROM users u
		JOIN friends f ON u.id = f.requester_id
		WHERE f.addressee_id = $1 AND f.status = 'pending'`
	rowsIn, err := ctx.Chub.Db.Query(context.Background(), qIn, userID)
	if err != nil {
		fmt.Printf("Failed pending_in :: %v\n", err)
		return
	}

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()

	pendingIn := scanFriendRows(ctx, rowsIn)

	// PROMETHEUS
	metrics.DbRequests.Inc()

	qOut := `SELECT u.id, u.username FROM users u
		JOIN friends f ON u.id = f.addressee_id
		WHERE f.requester_id = $1 AND f.status = 'pending'`
	rowsOut, err := ctx.Chub.Db.Query(context.Background(), qOut, userID)
	if err != nil {
		fmt.Printf("Failed pending_out :: %v\n", err)
		return
	}

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()

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

	// PROMETHEUS
	metrics.DbRequests.Inc()

	query := `
	DELETE FROM friends 
	WHERE (requester_id = $1 AND addressee_id = (SELECT id FROM users WHERE username = $2))
		OR (requester_id = (SELECT id FROM users WHERE username = $2) AND addressee_id = $1)
	RETURNING (SELECT id FROM users WHERE username = $2);
	`

	var friend_id string
	err := ctx.Chub.Db.QueryRow(context.Background(), query, ctx.Client.CurrUsrID, msg.Username).Scan(&friend_id)
	if (err != nil) {
		fmt.Printf("Friend remove failed :: %v\n", err)
		return
	}

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()

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

	// PROMETHEUS
	metrics.DbRequests.Inc()

	err := ctx.Chub.Db.QueryRow(context.Background(),
		`SELECT id FROM users WHERE username = $1`, to).Scan(&targetID)
	if err != nil {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "invite_sent", "success": false, "error": "User not found.",
		})
		return
	}

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()

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
	
	// PROMETHEUS
	metrics.DbRequests.Inc()
	
	err := ctx.Chub.Db.QueryRow(context.Background(),
		`UPDATE friends SET status = 'accepted', updated_at = NOW()
		 WHERE addressee_id = $1 AND requester_id = (SELECT id FROM users WHERE username = $2) AND status = 'pending'
		 RETURNING requester_id`,
		me, otherUsername).Scan(&requesterID)
	if err != nil {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "friend_accept_failed", "success": false, "error": "No pending request from that user.",
		})
		return
	}

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()

	metrics.DbRequests.Inc()

	var requesterName string
	_ = ctx.Chub.Db.QueryRow(context.Background(),
		`SELECT username FROM users WHERE id = $1`, requesterID).Scan(&requesterName)

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()
	
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

	// PROMETHEUS
	metrics.DbRequests.Inc()

	err := ctx.Chub.Db.QueryRow(context.Background(),
		`SELECT id, type FROM users WHERE username = $1`, targetName).Scan(&targetID, &targetType)
	if err != nil {
		fmt.Printf("Friend add: user not found :: %v\n", err)
		return
	}

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()

	if targetType == "guest" {
		_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
			"type": "friend_add_failed", "success": false, "error": "Cannot add guest accounts as friends.",
		})
		return
	}
	var st string
	var reqID string

	// PROMETHEUS
	metrics.DbRequests.Inc()

	err = ctx.Chub.Db.QueryRow(context.Background(),
		`SELECT f.status::text, f.requester_id::text FROM friends f WHERE
		 (f.requester_id = $1::uuid AND f.addressee_id = $2::uuid)
		 OR (f.requester_id = $2::uuid AND f.addressee_id = $1::uuid)`,
		me, targetID).Scan(&st, &reqID)

		if errors.Is(err, pgx.ErrNoRows){

			metrics.DbRequestsSucessful.Inc()

			// PROMETHEUS
			metrics.DbRequests.Inc()
			_, err = ctx.Chub.Db.Exec(context.Background(),
			`INSERT INTO friends (requester_id, addressee_id, status) VALUES ($1::uuid, $2::uuid, 'pending')`,
			me, targetID)

			if err != nil {
				fmt.Printf("Friend invite insert failed :: %v\n", err)
				_ = ctx.Client.Conn.WriteJSON(map[string]interface{}{
					"type": "friend_add_failed", "success": false, "error": "Could not send friend request.",
				})
				return
			}

			metrics.DbRequestsSucessful.Inc()

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

			// PROMETHEUS
			metrics.DbRequests.Inc()

			_, err = ctx.Chub.Db.Exec(context.Background(),
				`UPDATE friends SET status = 'accepted', updated_at = NOW()
				 WHERE requester_id = $1::uuid AND addressee_id = $2::uuid AND status = 'pending'`,
				targetID, me)

			if err != nil {
				fmt.Printf("mutual accept failed: %v\n", err)
				return
			}

			// PROMETHEUS
			metrics.DbRequestsSucessful.Inc()
			metrics.DbRequests.Inc()

			var requesterName string
			_ = ctx.Chub.Db.QueryRow(context.Background(),
				`SELECT username FROM users WHERE id = $1`, targetID).Scan(&requesterName)

			// PROMETHEUS
			metrics.DbRequestsSucessful.Inc()
			
			d.broadcastFriendAdded(ctx, targetID, requesterName, me, myName)
}

func (d* Dispatcher) HandleGetProfile(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) {
		return
	}
	fmt.Println("DEBUG: HandleGetProfile triggered!")

	var user ProfileUser
	var profileID string

	// PROMETHEUS
	metrics.DbRequests.Inc()

	query := `SELECT p.id, p.display_name, p.color, p.font
				FROM profiles p
				INNER JOIN users u ON p.id = u.id
				WHERE u.username = $1;`
	err := ctx.Chub.Db.QueryRow(context.Background(), query, msg.Username).Scan(
		&profileID, &user.Username, &user.Style.Color, &user.Style.Font)

	var response ProfileResponse
	
	response.Type = "profile_data"

	user.Online = (ctx.Chub.Clients[profileID] != nil) 
	user.IsGuest = ctx.Client.IsGuest

	if (err != nil) {
		response.Success = false
	} else {
		// PROMETHEUS
		metrics.DbRequestsSucessful.Inc()
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

	// PROMETHEUS
	metrics.DbRequests.Inc()

    query := `UPDATE profiles
                SET color = $2, font = $3, display_name = $4
                WHERE id = $1`
    _, err := ctx.Chub.Db.Exec(context.Background(), query, ctx.Client.CurrUsrID, msg.Style.Color, msg.Style.Font, msg.Username)

    if (err != nil ) { 
        fmt.Printf("FAILED TO UPDATE THE PROFILE TABLE FOR USER %v : %v\n", *ctx.Client.CurrUsrName, err)
        return
    }

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()

	metrics.DbRequests.Inc()

    usrnmQuery := ` UPDATE users
                    SET username = $2
                    WHERE id = $1;
    `

    _, err = ctx.Chub.Db.Exec(context.Background(), usrnmQuery, ctx.Client.CurrUsrID, msg.Username)

    if (err != nil ) { 
        fmt.Printf("FAILED TO UPDATE THE USERNAME FOR USER %v : %v\n", *ctx.Client.CurrUsrName, err)
        return
    }

	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()

    oldUsername := *ctx.Client.CurrUsrName
    *ctx.Client.CurrUsrName = msg.Username
    
    fmt.Printf("Session username updated from %s to %s (ID: %s)\n", oldUsername, *ctx.Client.CurrUsrName, *ctx.Client.CurrUsrID)

    newToken, err := GenerateJWT(*ctx.Client.CurrUsrID, *ctx.Client.CurrUsrName)
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

func (d *Dispatcher) HandleChatAI(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeIsValidChat)) { return }

	fmt.Printf("DEBUG: chat_message\n")
	if AIRoom, ok := ctx.Client.CurrentRoom.(*gamemanager.AIRoom); ok {
		AIRoom.BroadcastChat(*ctx.Client.CurrUsrID, msg.Text)
	}
}

func (d *Dispatcher) HandleAuth(ctx *WSContext, msg Message) {
	claims, err := validateAndGetClaims(msg.Token)
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

	// PROMETHEUS
	metrics.DbRequests.Inc()

	_ = ctx.Chub.Db.QueryRow(context.Background(),
		`SELECT type FROM users WHERE id = $1`, claims.UserID).Scan(&clientType)
	
	
	// PROMETHEUS
	metrics.DbRequestsSucessful.Inc()
	
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
