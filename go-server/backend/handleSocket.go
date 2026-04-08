package main

import (
	"strings"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"context"
	"main.go/gamemanager"
	"sync"
)

// NEW STRUCT FOR CLIENT / WEBSOCKET MANAGEMENT

type Client struct {
	CurrUsrID		*string				// UserID, to identify the
	CurrUsrName		*string				// Ditto, for the username
	Conn			*websocket.Conn		// The actual websocket
	Hub				*gamemanager.Hub	// Reference to the game manager
	CurrentRoom		gamemanager.GameRoom// The current room of a GAMER
}

type ClientHub struct {
	Clients map[string]*Client

	Db		*pgxpool.Pool
	mu		sync.RWMutex
}

type WSContext struct {
	client *Client
	chub *ClientHub
}

type HandleFunc func(ctx *WSContext, msg Message)
type PipeFunc func(ctx	 *WSContext, msg Message) bool

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
	if len(strings.TrimSpace(msg.Text)) == -1 { return false }
	return true
}

func (d *Dispatcher) PipeIsAuth(ctx *WSContext, msg Message) bool {

	if ctx.client.CurrUsrID == nil || *ctx.client.CurrUsrID == "" {
		return false
	}

	if ctx.client.CurrUsrName == nil || *ctx.client.CurrUsrName == "" {
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
	tmpRoom, err := ctx.client.Hub.GetRoom(msg.Code)
	if err != nil || tmpRoom == nil {
		return false
	}
	ctx.client.CurrentRoom = tmpRoom
	return true
}

func (d *Dispatcher) Dispatch(ctx *WSContext, msg Message) {
	handler, ok := d.handlers[msg.Type]
	if !ok {
		fmt.Printf("DEBUG: %s not found\n", msg.Type)
		ctx.client.Conn.WriteJSON(map[string]string{"type": "error", "message": "Action not recognized by the server",})
		return
	}
	handler(ctx, msg)
}

type Friend struct {
	ID string `json:"id"`
	Username string `json:"username"`
	Online bool `json:"online"`
}

type FriendsListResponse struct {
		Type string `json:"type"`
		Friends []Friend `json:"friends,omitempty"`
		Friend Friend `json:"friend,omitempty"`
		Success bool `json:"success,omitempty"`
		FriendID string `json:"friend_id,omitempty"`
}


func (d *Dispatcher) HandleGetFriend(ctx *WSContext, msg Message) {
	fmt.Println("DEBUG: HandleGetFriend triggered!")
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) { 
		return
	}

	query := `	SELECT users.id, users.username
				FROM users
				JOIN friends ON (users.id = friends.requester_id OR users.id = friends.addressee_id)
				WHERE
				(friends.requester_id = $1 OR friends.addressee_id = $1)
				AND users.id != $1
			`

	rows, err := ctx.chub.Db.Query(context.Background(), query, msg.ID)
	if err != nil {
		fmt.Printf("Failed to open rows :: %v\n", err)
		return
	}
	defer rows.Close()

	friends := make([]Friend, 0)
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.ID, &f.Username); err != nil {
			fmt.Printf("Error scanning friends row :: %v\n", err)
			continue 
		}
		if (ctx.chub.Clients[f.ID] != nil) {
			f.Online = true
		}
		friends = append(friends, f)
	}

	response := FriendsListResponse {
		Type: "friends_list",
		Friends: friends, 
	}

	err = ctx.client.Conn.WriteJSON(response)
	if err != nil {
		fmt.Printf("failed to send friends list: %v", err)
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
	err := ctx.chub.Db.QueryRow(context.Background(), query, ctx.client.CurrUsrID, msg.Username).Scan(&friend_id)
	if (err != nil) {
		fmt.Printf("Friend remove failed :: %v\n", err)
		return
	}

	response := FriendsListResponse {
		Type: "friend_removed",
		FriendID: friend_id,
	}

	err = ctx.client.Conn.WriteJSON(response)
	if err != nil {
		fmt.Printf("Failed to send friend_id on removal: %v", err)
	}

	response = FriendsListResponse {
		Type: "friend_removed",
		FriendID: *ctx.client.CurrUsrID,
	}

	fmt.Println("TRYING TO REMOVE SENDER FROM ADDRESSEE FRIENDLIST")

	err = ctx.chub.Clients[friend_id].Conn.WriteJSON(response)
	if err != nil {
		fmt.Printf("Failed to send friend_id to addressee on removal: %v", err)
	}

}

func (d *Dispatcher) HandleAddFriend(ctx *WSContext, msg Message) {
	fmt.Println("DEBUG: HandleAddFriend triggered!")
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) {
		return
	}

	query := `
		INSERT INTO friends (requester_id, addressee_id)
		SELECT $1, id
		FROM users
		WHERE username = $2
		RETURNING addressee_id;
	`

	var addressee_id string

	err := ctx.chub.Db.QueryRow(context.Background(), query, ctx.client.CurrUsrID, msg.Username).Scan(&addressee_id)
	if (err != nil) {
		fmt.Printf("Friend invite failed :: %v\n", err)
		return
	}

	var addressee_online bool
	addressee_online = (ctx.chub.Clients[addressee_id] != nil)

	response := FriendsListResponse {
		Friend : Friend {
			ID : addressee_id,
			Username : msg.Username,
			Online : addressee_online,
		},
		Type : "friend_added",
		Success : true,
	}

	fmt.Printf("DEBUG: SENDING TYPE %v SUCCESS %v ID %v USERNAME %v ONLINE %v\n", "friend_added", true, addressee_id, msg.Username, false)

	err = ctx.client.Conn.WriteJSON(response)
	if err != nil {
		fmt.Printf("failed to send friends list: %v", err)
	}

	response = FriendsListResponse {
		Friend : Friend {
			ID : *ctx.client.CurrUsrID,
			Username : *ctx.client.CurrUsrName,
			Online : true,
		},
		Type : "friend_added",
		Success : true,
	}

	fmt.Printf("DEBUG: SENDING TYPE %v SUCCESS %v ID %v USERNAME %v ONLINE %v\n", "friend_added", true, addressee_id, msg.Username, false)

	err = ctx.chub.Clients[addressee_id].Conn.WriteJSON(response)
	if err != nil {
		fmt.Printf("failed to send friends list to addressee: %v", err)
	}
}

type ProfileStyle struct {
	Color string `json:"color,omitempty"`
	Font string `json:"font,omitempty"`
}

type ProfileUser struct {
	Username string `json:"username,omitempty"`
	Email string `json:"email,omitempty"`
	Online bool `json:"online,omitempty"`
	Style ProfileStyle `json:"style,omitempty"`
}

type ProfileResponse struct {
	Type string `json:"type"`
	User ProfileUser `json:"user,omitempty"`
	Success bool `json:"success,omitempty"`
	IsCaller bool `json:"is_me,omitempty"`
}

func (d* Dispatcher) HandleGetProfile(ctx *WSContext, msg Message) {
	fmt.Println("DEBUG: HandleGetProfile triggered!")
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) {
		return
	}

	var user ProfileUser
	var profileID string

	query := `SELECT p.id, p.display_name, p.color, p.font 
				FROM profiles p
				INNER JOIN users u ON p.id = u.id
				WHERE u.username = $1;`
	err := ctx.chub.Db.QueryRow(context.Background(), query, msg.Username).Scan(&profileID, &user.Username, &user.Style.Color, &user.Style.Font)

	var response ProfileResponse
	
	response.Type = "profile_data"

	user.Online = (ctx.chub.Clients[profileID] != nil) 

	if (err != nil) {
		response.Success = false
	} else {
		response.Success = true
		response.User = user
		response.IsCaller = (profileID == *ctx.client.CurrUsrID)
	}

	// fmt.Printf("REQUESTED USERNAME: %v\n", msg.Username)
	// fmt.Printf("USER USERNAME %v ONLINE %v COLOR %v FONT %v\n", user.Username, user.Online, user.Style.Color, user.Style.Font)

	err = ctx.client.Conn.WriteJSON(response)
	if err != nil {
		fmt.Printf("failed to send profile data: %v", err)
	}
}

func (d* Dispatcher) HandleProfileUpdate(ctx *WSContext, msg Message) {
	fmt.Println("DEBUG: HandleProfileUpdate triggered!")
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) {
		return
	}

	query := `UPDATE profiles
				SET color = $2, font = $3, display_name = $4
				WHERE id = $1`
	_, err := ctx.chub.Db.Exec(context.Background(), query, ctx.client.CurrUsrID, msg.Style.Color, msg.Style.Font, msg.Username)

	if (err != nil ) { // need to eventually use database transactions
		fmt.Printf("FAILED TO UPDATE THE PROFILE TABLE FOR USER %v : %v\n", *ctx.client.CurrUsrName, err)
		return
	}

	usrnmQuery := ` UPDATE users
					SET username = $2
					WHERE id = $1;
	`

	_, err = ctx.chub.Db.Exec(context.Background(), usrnmQuery, ctx.client.CurrUsrID, msg.Username)

	if (err != nil ) { 
		fmt.Printf("FAILED TO UPDATE THE USERNAME FOR USER %v : %v\n", *ctx.client.CurrUsrName, err)
		return
	}

	*ctx.client.CurrUsrName = msg.Username

	var response ProfileResponse

	response.Type = "profile_updated"

	if (err != nil) {
		response.Success = false
	} else {
		response.Success = true
		response.User.Username = *ctx.client.CurrUsrName
		response.User.Style.Color = msg.Style.Color
		response.User.Style.Font = msg.Style.Font
	}

	err = ctx.client.Conn.WriteJSON(response)
	if err != nil {
		fmt.Printf("failed to send profile data: %v", err)
	}
}

func NewDispatcher() *Dispatcher {
	d:= &Dispatcher{
		handlers: make(map[string]HandleFunc),
	}
	d.handlers["get_profile"] = d.HandleGetProfile
	d.handlers["update_profile"] = d.HandleProfileUpdate
	d.handlers["add_friend"] = d.HandleAddFriend
	d.handlers["get_friends"] = d.HandleGetFriend
	d.handlers["remove_friend"] = d.HandleRemoveFriend
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

func (d *Dispatcher) HandleLeaveAIGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }


	RoomIA := ctx.client.CurrentRoom.(*gamemanager.AIRoom)
	del := RoomIA.LeaveGame(*ctx.client.CurrUsrID)
	if del {
		ctx.client.Hub.DeleteRoom(msg.Code)
	}
}

func (d *Dispatcher) HandleJoinAIGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }

	base := ctx.client.CurrentRoom.GetBase()
	base.JoinGame(*ctx.client.CurrUsrID, ctx.client.Conn)
	fmt.Printf("DEBUG join_game: code='%s' user='%s'\n", msg.Code, *ctx.client.CurrUsrName)

	RoomIA := ctx.client.CurrentRoom.(*gamemanager.AIRoom)
	if (base.Status == gamemanager.StateAIDrawing) {
		base.MessageChan <- gamemanager.Notification{
			PlayerID: *ctx.client.CurrUsrID,
			Data: map[string]interface{}{
				"type":		"ai_game_state",
				"phase":	"draw",
				"prompt":	RoomIA.Prompt,
				"room":		base.ID,
			},
		}
	}
	fmt.Printf("DEGUB: %s\n", msg.Type)
}

func (d *Dispatcher) HandleJoinAIRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	base := ctx.client.CurrentRoom.GetBase()
	err := base.AddPlayer(*ctx.client.CurrUsrID, *ctx.client.CurrUsrName, ctx.client.Conn)
	if err != nil {
		fmt.Println("AddPlayer error:", err)
		return
	}
	if AIRoom, ok := ctx.client.CurrentRoom.(*gamemanager.AIRoom); ok {
		AIRoom.SendSystemMsg(fmt.Sprintf("%s join the lobby !", *ctx.client.CurrUsrName))
	}

	base.BroadcastLobbyState()

}

func (d *Dispatcher) HandleLeaveAILobby(ctx *WSContext, msg Message) { 
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }


	base := ctx.client.CurrentRoom.GetBase()
	if base.Status != gamemanager.StateAIWaiting { return }

	isHost := false

	if p, err := base.GetPlayer(*ctx.client.CurrUsrID); err == nil {
		isHost = p.IsHost
	}

	base.RemovePlayer(*ctx.client.CurrUsrID)
	if AIRoom, ok := ctx.client.CurrentRoom.(*gamemanager.AIRoom); ok {
		AIRoom.SendSystemMsg(fmt.Sprintf("%s leave the lobby !", *ctx.client.CurrUsrName))
	}

	if len(base.Players) == 0 {
		ctx.client.Hub.DeleteRoom(base.ID)
		return
	}
	if isHost {
		base.TransferHost()
	}
	base.BroadcastLobbyState()
}

func (d *Dispatcher) HandleCreateAIRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) { return }

	newRoom := ctx.client.Hub.CreateRoom(true)
	base := newRoom.GetBase()
	err := base.AddPlayer(*ctx.client.CurrUsrID, *ctx.client.CurrUsrName, ctx.client.Conn)
	if err != nil { return }
	fmt.Printf("DEGUB: AI_ROOM created\n")	

	base.MessageChan <- gamemanager.Notification{
		PlayerID: *ctx.client.CurrUsrID,
		Data: map[string]interface{}{
			"type": "ai_room_created",
			"code": base.ID,
			"palyers": []map[string]interface{}{
				{
					"id": *ctx.client.CurrUsrID,
					"name": *ctx.client.CurrUsrName,
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

	room, _ := ctx.client.Hub.GetRoom(msg.Code)
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
	fmt.Printf("DEBUG: %s\n", msg.Type)
	if RoomIA, ok := ctx.client.CurrentRoom.(*gamemanager.AIRoom); ok {
		fmt.Printf("DEBUG: RunLoopAI\n")
		go RoomIA.RunAIGameLoop(prompt)
	}
}

func (d *Dispatcher) HandleAIDraw(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeHasRoomCode)) { return }

	if RoomIA, ok := ctx.client.CurrentRoom.(*gamemanager.AIRoom); ok {
		RoomIA.SubmitDrawing(*ctx.client.CurrUsrID, msg.Drawing, msg.Title, msg.Description)
	}
}

func (d *Dispatcher) HandleAIVote(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeHasRoomCode)) { return }

	if RoomIA, ok := ctx.client.CurrentRoom.(*gamemanager.AIRoom); ok {
		RoomIA.SubmitVotes(*ctx.client.CurrUsrID, msg.Votes)
	}
}

func (d *Dispatcher) HandleChatAI(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeIsValidChat)) { return }

	fmt.Printf("DEBUG: chat_message\n")
	if AIRoom, ok := ctx.client.CurrentRoom.(*gamemanager.AIRoom); ok {
		AIRoom.BroadcastChat(*ctx.client.CurrUsrID, msg.Text)
	}
}

func (d *Dispatcher) HandleAuth(ctx *WSContext, msg Message) {
	claims, err := validateAndGetClaims(msg.Token)
	if err != nil {
		fmt.Println("WS Auth Failed:", err)
		ctx.client.Conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(4000, "token expired"))
		return
	}
	
	ctx.client.CurrUsrName = &claims.Username
	ctx.client.CurrUsrID = &claims.UserID

	ctx.chub.mu.Lock()
	ctx.chub.Clients[claims.UserID] = ctx.client
	ctx.chub.mu.Unlock()

	fmt.Printf("WS Authenticated: %s (ID: %s)\n", *ctx.client.CurrUsrName, *ctx.client.CurrUsrID)
}

func (d *Dispatcher) HandleCreateRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth)) { return }

	ctx.client.CurrentRoom = ctx.client.Hub.CreateRoom(false)
	base := ctx.client.CurrentRoom.GetBase()
	err := base.AddPlayer(*ctx.client.CurrUsrID, *ctx.client.CurrUsrName, ctx.client.Conn)
	if err != nil {
		fmt.Println("DEBUG: ", err)
		return
	}

	fmt.Printf("DEBUG: %s success\n", msg.Type)

	base.MessageChan <- gamemanager.Notification{
		PlayerID: *ctx.client.CurrUsrID,
		Data: map[string]interface{}{
			"type":		"room_created",
			"code":		base.ID,
			"players":	[]map[string]interface{}{
				{
					"id":	ctx.client.CurrUsrID,
					"name":	ctx.client.CurrUsrName,
					"host":	true,
				},
			},
		},
	}
	
	base.BroadcastLobbyState()

	// go func(roomID string) {
	// 	ctx.Db.Exec(context.Background(), "INSERT INTO rooms (room_code, created_at) VALUES ($1, NOW())", roomID)
	// }(base.ID)
}

func (d *Dispatcher) HandleJoinRoom(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	base := ctx.client.CurrentRoom.GetBase()
	err := base.AddPlayer(*ctx.client.CurrUsrID, *ctx.client.CurrUsrName, ctx.client.Conn)
	if err != nil {
		fmt.Println("AddPlayer error:", err)
		return
	}
	if classicRoom, ok := ctx.client.CurrentRoom.(*gamemanager.Room); ok {
		classicRoom.SendSystemMsg(fmt.Sprintf("%s join the lobby !", *ctx.client.CurrUsrName))
	}

	base.BroadcastLobbyState()

}

func (d *Dispatcher) HandleJoinGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }

	base := ctx.client.CurrentRoom.GetBase()
	base.JoinGame(*ctx.client.CurrUsrID, ctx.client.Conn)
	fmt.Printf("DEBUG join_game: code='%s' user='%s'\n", msg.Code, *ctx.client.CurrUsrName)
	
	var task gamemanager.GameStateRecord
	// task := base.GetPlayerTask(*WSContext.CurrUsrID)
	if classicRoom, ok := ctx.client.CurrentRoom.(*gamemanager.Room); ok {
		task = classicRoom.GetPlayerTask(*ctx.client.CurrUsrID)
	}
	ctx.client.Conn.WriteJSON(task)
}

func (d *Dispatcher) HandleLeaveLobby(ctx *WSContext, msg Message) { 
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeHasRoomCode, d.PipeRoomExist)) { return }


	base := ctx.client.CurrentRoom.GetBase()
	if base.Status != gamemanager.StateWaiting { return }

	fmt.Printf("DEBUG leave_lobby: code='%s' user='%s'\n", msg.Code, *ctx.client.CurrUsrName)
	isHost := false

	if p, err := base.GetPlayer(*ctx.client.CurrUsrID); err == nil {
		isHost = p.IsHost
	}

	base.RemovePlayer(*ctx.client.CurrUsrID)
	if classicRoom, ok := ctx.client.CurrentRoom.(*gamemanager.Room); ok {
		classicRoom.SendSystemMsg(fmt.Sprintf("%s leave the lobby !", *ctx.client.CurrUsrName))
	}

	if len(base.Players) == 0 {
		ctx.client.Hub.DeleteRoom(base.ID)
		return
	}
	if isHost {
		base.TransferHost()
	}
	base.BroadcastLobbyState()
}

func (d *Dispatcher) HandleLeaveGame(ctx *WSContext, msg Message) { 
	fmt.Printf("DEBUG: leave_game msg %s\n", msg.Code)
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	if classicRoom, ok := ctx.client.CurrentRoom.(*gamemanager.Room); ok {
		classicRoom.SendSystemMsg(fmt.Sprintf("%s leave the lobby !", *ctx.client.CurrUsrName))
		del := classicRoom.LeaveGame(*ctx.client.CurrUsrID)
		if del {
			ctx.client.Hub.DeleteRoom(msg.Code)
			fmt.Printf("DEBUG: DELETE ROOM nobody is in") // TODO Fix le isReady et set Prompt a "..."
		}
	}
}



func (d *Dispatcher) HandlePrompt(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist)) { return }

	data := map[string]interface{}{
		"type": "prompt",
		"prompt": msg.Prompt,
	}
	fmt.Printf("DEGUB: %s\n", msg.Type)
	if classicRoom, ok := ctx.client.CurrentRoom.(*gamemanager.Room); ok {
		err := classicRoom.SubmiteAction(*ctx.client.CurrUsrID, data, true);
		if (err != nil) {
			fmt.Printf("Error: Submited prompt: %v\n", err);
		}
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
	if classicRoom, ok := ctx.client.CurrentRoom.(*gamemanager.Room); ok {
		err := classicRoom.SubmiteAction(*ctx.client.CurrUsrID, data, true);
		if (err != nil) {
			fmt.Printf("Error: Submited draw: %v\n", err);
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
	if classicRoom, ok := ctx.client.CurrentRoom.(*gamemanager.Room); ok {
		err := classicRoom.SubmiteAction(*ctx.client.CurrUsrID, data, true);
		if (err != nil) {
			fmt.Printf("Error: Submited guess: %v\n", err);
		}
	}
}


func (d *Dispatcher) HandleStartGame(ctx *WSContext, msg Message) {
	if (!RunPipeLine(ctx, msg, d.PipeIsAuth, d.PipeRoomExist, d.PipeHasRoomCode)) { return }

	base := ctx.client.CurrentRoom.GetBase()
	if classicRoom, ok := ctx.client.CurrentRoom.(*gamemanager.Room); ok {
		err := classicRoom.StartGame()
		if err != nil {
			ctx.client.Conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
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
	if classicRoom, ok := ctx.client.CurrentRoom.(*gamemanager.Room); ok {
		classicRoom.BroadcastChat(*ctx.client.CurrUsrID, msg.Text)
	}
}
