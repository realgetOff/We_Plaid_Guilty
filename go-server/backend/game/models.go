package gamemanager

import (
	"sync"
)

type GameStates string

const (
	StateWaiting	GameStates = "WAITING"
	StateWriting	GameStates = "WRITING"
	StateDrawing	GameStates = "DRAWING"
	StateFinished	GameStates = "FINISHED"
)

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
	LastDraft string
	IsReady bool
	isConnected bool
}

/*
* All the others structs are contained in the Room structure.
* The main structure for the game.
*/
type Room struct {
	ID string
	Players map[int]*Player
	Books map[int]*Book
	PlayerOrder []int
	FinishedChan chan bool
	Status GameStates
	mu sync.Mutex
}
