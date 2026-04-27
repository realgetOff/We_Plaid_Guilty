package main

import (
	"main.go/gamemanager"
	"sync"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gorilla/websocket"
)

type Client struct {
	CurrUsrID		*string				// UserID, to identify the
	CurrUsrName		*string				// Ditto, for the username
	IsGuest			bool
	Conn			*websocket.Conn		// The actual websocket
	Hub				*gamemanager.Hub	// Reference to the game manager
	CurrentRoom		gamemanager.GameRoom// The current room of a GAMER
}

type ClientHub struct {
	Clients map[string]*Client

	db		*pgxpool.Pool
	mu		sync.RWMutex
}

type WSContext struct {
	client *Client
	chub *ClientHub
}

const DEFAULT_COLOR = "#000000"
const NOT_FOUND_DB = "ERROR: Player not found in the DB\n"
const ERROR_WRITE_WS = "ERROR: WS %v\n"

type HandleFunc func(ctx *WSContext, msg Message)
type PipeFunc func(ctx	 *WSContext, msg Message) bool

type Dispatcher struct {
	handlers map[string]HandleFunc
}

type Friend struct {
	ID string `json:"id"`
	Username string `json:"username"`
	Online bool `json:"online"`
}

type FriendsListResponse struct {
	Type         string   `json:"type"`
	Friends      []Friend `json:"friends,omitempty"`
	PendingIn    []Friend `json:"pending_in,omitempty"`
	PendingOut   []Friend `json:"pending_out,omitempty"`
	Friend       Friend   `json:"friend,omitempty"`
	Success      bool     `json:"success,omitempty"`
	FriendID     string   `json:"friend_id,omitempty"`
	GuestNoFriends bool   `json:"guest_no_friends,omitempty"`
}

type ProfileStyle struct {
	Color string `json:"color,omitempty"`
	Font string `json:"font,omitempty"`
}

type ProfileUser struct {
	Username string `json:"username,omitempty"`
	Email string `json:"email,omitempty"`
	Online bool `json:"online,omitempty"`
	IsGuest bool `json:"is_guest,omitempty"`
	Style ProfileStyle `json:"style,omitempty"`
}

type ProfileResponse struct {
	Type string `json:"type"`
	User ProfileUser `json:"user,omitempty"`
	Success bool `json:"success,omitempty"`
	IsCaller bool `json:"is_me,omitempty"`
}

