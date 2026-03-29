package gamemanager

import (
	"fmt"
	"github.com/gorilla/websocket"
)

func (r *Room) GetPlayer(playerID string) (*Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	player, ok := r.Players[playerID]
	if !ok {
		return nil, fmt.Errorf("Player not found")
	}
	return player, nil
}

func (r *Room) SetPlayerOnline(playerID string, isOnline bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.Players[playerID]; ok {
		p.IsConnected = isOnline
	}
}

func (r *Room) updateStatus(status GameStates) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Status = status
}

func (r *Room) resetPlayer() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.Players {
		p.IsReady = false
	}
}

func (r *Room) TransferHost() {
    r.mu.Lock()
    defer r.mu.Unlock()

    for _, p := range r.Players {
        p.IsHost = true
        break
    }
}

func NewRoom(id string, rounds int, timer int) *Room {
	return &Room{
		ID:           id,
		Status:       StateWaiting,
		TotalRound:   rounds,
		Timer:        timer,
		Players:      make(map[string]*Player),
		Books:        make(map[string]*Book),
		PlayerOrder:  []string{},
		MessageChan:  make(chan Notification, 100),
		FinishedChan: make(chan bool, 1),
	}
}

/*
* Add a player to the room
*/
func (r *Room) AddPlayer(playerID string, name string, conn *websocket.Conn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= 8 {
		return fmt.Errorf("Room is full")
	}

	newPlayer := &Player{
		ID:          playerID,
		Name:        name,
		Conn:        conn,
		IsHost:      len(r.Players) == 0,
		IsConnected: true,
		IsReady:     false,
		Score:       0,
	}

	r.Players[playerID] = newPlayer
	return nil
}

/*
* Remove a player from the room
*/
func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Players, playerID)

	newOrder := []string{}
	for _, id := range r.PlayerOrder {
		if id != playerID {
			newOrder = append(newOrder, id)
		}
	}
	r.PlayerOrder = newOrder
}

func (r *Room) UpdatePlayerConn(playerID string, conn *websocket.Conn) {
    r.mu.Lock()
    defer r.mu.Unlock()
    if p, ok := r.Players[playerID]; ok {
        p.Conn = conn
    }
}
