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
		if (p.IsConnected){
			p.IsReady = false
		}
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
func (b *BaseRoom) AddPlayer(playerID string, name string, conn *websocket.Conn) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.Players) >= 8 {
		return fmt.Errorf("Room is full")
	}

	newPlayer := &Player{
		ID:          playerID,
		Name:        name,
		Conn:        conn,
		IsHost:      len(b.Players) == 0,
		IsConnected: true,
		IsReady:     false,
		Score:       0,
	}

	b.Players[playerID] = newPlayer
	return nil
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

func (r *Room) LeaveGame(playerID string) (bool){
	r.mu.Lock()
	defer r.mu.Unlock()
	if oldPlayer, ok := r.Players[playerID]; ok {
		oldPlayer.IsReady = true
		oldPlayer.IsConnected = false
	}
	var isAllDisconnect bool
	for _, p := range r.Players {
		if !p.IsConnected {
			isAllDisconnect = true
		} else {
			isAllDisconnect = false
			break
		}
	}
	return isAllDisconnect
}

func (r *Room) JoinGame(playerID string, newConn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if oldPlayer, ok := r.Players[playerID]; ok {
		oldPlayer.Conn = newConn
		oldPlayer.IsConnected = true
	}
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
