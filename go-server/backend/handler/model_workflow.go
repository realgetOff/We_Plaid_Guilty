package handler

import (
	"main.go/gamemanager"
	"sync"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var JwtSecret = []byte("replace_with_env_or_equivalent_later")

type MyCustomClaims struct {
	Username string `json:"username"`
	UserID   string `json:"id"`
	jwt.RegisteredClaims
}

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

	Db		*pgxpool.Pool
	Mu		sync.RWMutex
}

type WSContext struct {
	Client *Client
	Chub *ClientHub
}


/*
The message structure contains the json information to be sent / received by the websocket for room generation
type: state before / after generation of the room code
code: room code
omitempty: omits empty strings, lowering network traffic
*/

type Message struct {
	Type        string         `json:"type"`
	Text        string         `json:"text,omitempty"`
	Token       string         `json:"token,omitempty"`
	Code        string         `json:"code,omitempty"`
	Reason      string         `json:"reason,omitempty"`
	Prompt      string         `json:"prompt,omitempty"`
	Drawing     string         `json:"drawing,omitempty"`
	Guess       string         `json:"guess,omitempty"`
	Votes       map[string]int `json:"votes,omitempty"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	Username    string         `json:"username,omitempty"`
	To          string         `json:"to,omitempty"`
	ID          string         `json:"id,omitempty"`
	IsAI        bool           `json:"is_ai,omitempty"`
	Style       ProfileStyle   `json:"style,omitempty"`
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
	Type			string		`json:"type"`
	Friends			[]Friend	`json:"friends,omitempty"`
	PendingIn		[]Friend	`json:"pending_in,omitempty"`
	PendingOut		[]Friend	`json:"pending_out,omitempty"`
	Friend			Friend		`json:"friend,omitempty"`
	Success			bool		`json:"success,omitempty"`
	FriendID		string		`json:"friend_id,omitempty"`
	GuestNoFriends	bool		`json:"guest_no_friends,omitempty"`
}

type ProfileStyle struct {
	Color	string	`json:"color,omitempty"`
	Font	string	`json:"font,omitempty"`
}

type ProfileUser struct {
	Username	string			`json:"username,omitempty"`
	Email		string			`json:"email,omitempty"`
	Online		bool			`json:"online,omitempty"`
	IsGuest		bool			`json:"is_guest,omitempty"`
	Style		ProfileStyle	`json:"style,omitempty"`
}

type ProfileResponse struct {
	Type		string		`json:"type"`
	User		ProfileUser	`json:"user,omitempty"`
	Success		bool		`json:"success,omitempty"`
	IsCaller	bool		`json:"is_me,omitempty"`
}

