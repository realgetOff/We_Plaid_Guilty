package gamemanager

import (
	"sync"
	"github.com/gorilla/websocket"
)

type GameStates string

const (
	StateWaiting	GameStates = "waiting"
	StateWriting	GameStates = "write"
	StateDrawing	GameStates = "draw"
	StateGuess		GameStates = "guess"
	StateFinished	GameStates = "gallery"
)

type GameStateRecord struct {
	Type    string `json:"type"`
	Phase   string `json:"phase"`
	Room    string `json:"room"`
	Prompt  string `json:"prompt,omitempty"`
	Drawing string `json:"drawing,omitempty"`
}

/* * The entry structure is like a new page from a book.
*/
type Entry struct {
	AuthorID string `json:"authorId"`
	Content  string `json:"content"`
	Type     string `json:"type"`
}

/* * The Book structure is a mere structure for the Players.
*/
type Book struct {
	OwnerID string  `json:"ownerId"`
	Entries []Entry `json:"entries"`
}

/*
* A simple Player struct.
*/
type Player struct {
	ID          string          `json:"id"`
	Score       int             `json:"score"`
	Name        string          `json:"name"`
	Conn        *websocket.Conn `json:"-"`
	LastDraft   string          `json:"lastDraft"`
	IsHost      bool            `json:"isHost"`
	IsReady     bool            `json:"isReady"`
	WriteMu     sync.Mutex      `json:"-"`
	IsConnected bool            `json:"isConnected"`
}

type Notification struct {
	PlayerID string      `json:"playerId"`
	Data     interface{} `json:"data"`
}

type BaseRoom struct {
	ID           string
	Phase        string
	TotalRound   int
	Players      map[string]*Player
	PlayerOrder  []string  
	FinishedChan chan bool
	Status       GameStates
	MessageChan  chan Notification
	
	mu           sync.Mutex
}

type AIRoom struct {
	BaseRoom
	Prompt       string
	Drawings     map[string]*AIDrawing
	Votes        []AIVote
	DrawingsDone int
	VotesDone    int
	DrawChan     chan bool
	VoteChan     chan bool
}

type Room struct {
	BaseRoom
	Books map[string]*Book
}

/*
* All the others structs are contained in the Room structure.
*/
// type Room struct {
	// ID           string
	// Phase        string
	// TotalRound   int
	// Players      map[string]*Player
	// Books        map[string]*Book
	// PlayerOrder  []string  
	// FinishedChan chan bool
	// Status       GameStates
	// MessageChan  chan Notification
	// 
	// mu           sync.Mutex
// }
