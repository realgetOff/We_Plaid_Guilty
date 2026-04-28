package handler

import (
	"fmt"
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
