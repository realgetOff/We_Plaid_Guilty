package gamemanager

import (
	"sync"
	"fmt"
	"time"
)

type GameStates string

const (
	StateWaiting	GameStates = "WAITING"
	StateWriting	GameStates = "WRITING"
	StateDrawing	GameStates = "DRAWING"
	StateFinished	GameStates = "FINISHED"
)

type Player struct {
	ID int
	Name string
	Score int
	IsReady bool
}

type Room struct {
	ID string
	Players map[int]*Player
	FinishedChan chan bool
	Status GameStates
	mu sync.Mutex
}

func NewRoom(id string) (*Room) {
	return &Room {
		ID: id,
		Status: StateWaiting,
		Players: make(map[int]*Player),
	}
}

func (r *Room) AddPlayer(p *Player) (error){
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Status != StateWaiting {
		return fmt.Errorf("Game started, too late!\n")
	}
	_, ok := r.Players[p.ID]
	if ok == true {
		return fmt.Errorf("Player %s is already in the game!\n", p.Name)
	}
	r.Players[p.ID] = p
	return nil
}

func (r *Room) waitForPhase(timout time.Duration) {
	
}

func (r *Room) updateStatus(newStatus GameStates) {
	r.mu.Lock()
	r.Status = newStatus
	r.mu.Unlock()
}

func (r *Room) RunGameLoop() {

	TotalRound := len(r.Players)

	for i := 1; i < TotalRound; i++ {
		r.updateStatus(StateWriting)
	
		r.updateStatus(StateDrawing)
	}
}

func (r *Room) StartGame() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) < 3 {
		return fmt.Errorf("pas assez de joueur pour commencer")
	}

	if r.Status != StateWaiting {
		return fmt.Errorf("la partie a deja commencer")
	}

	r.Status = StateWriting

	r.FinishedChan = make(chan bool, len(r.Players))

	// go r.RunGameLoop()

	return nil
}
