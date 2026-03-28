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

/* 
* The entry structure is like a new page from a book.
*/
type Entry struct {
	AuthorID int
	Content string
	Type string
}

/* 
* The Book structure is a mere structure for the Players.
* It is exchanged with all the players of the game/room. 
*/
type Book struct {
	OwnerID int
	Entries []Entry
}

/*
* A simple Player struct.
*/
type Player struct {
	ID int
	Score int
	Name string
	Conn *websocket.Conn
	LastDraft string
	IsHost bool
	IsReady bool
	isConnected bool
}

type Notification struct {
	PlayerID int
	Data interface{}
}

/*
* All the others structs are contained in the Room structure.
* The main structure for the game.
*/
type Room struct {
	ID string
	Phase string
	Timer int
	TotalRound int
	CurrentRound int
	Players map[int]*Player
	Books map[int]*Book
	PlayerOrder []int
	FinishedChan chan bool
	Status GameStates
	MessageChan chan Notification
	mu sync.Mutex
}
