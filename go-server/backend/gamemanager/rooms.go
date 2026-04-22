package gamemanager

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

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

func (b *BaseRoom) RemovePlayer(playerID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	newOrder := []string{}
	for _, id := range b.PlayerOrder {
		if id != playerID {
			newOrder = append(newOrder, id)
		}
	}
	b.PlayerOrder = newOrder
	delete(b.Players, playerID)
}

func (b *BaseRoom) TransferHost() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if (len(b.Players) == 0) {
		return
	}

	hasHost := false
	for _, p := range b.Players {
		if p.IsHost {
			hasHost = true
			return
		}
	}
	if !hasHost {
		for _, p := range b.Players {
				p.IsHost = true
				break
			}
		}
}

func (b *BaseRoom) GetPlayer(playerID string) (*Player, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p, ok := b.Players[playerID]
	if !ok {
		return nil, fmt.Errorf("player with id %s not found", playerID)
	}
	return p, nil
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

	var newPlayer* Player
	_, ok := b.Players[playerID]
	if (ok && len(b.Players) == 1) {
		newPlayer = &Player{
			ID:          playerID,
			Name:        name,
			Conn:        conn,
			IsHost:      true,
			IsConnected: true,
			IsReady:     false,
			Score:       0,
		}
	} else {
		newPlayer = &Player{
			ID:          playerID,
			Name:        name,
			Conn:        conn,
			IsHost:      len(b.Players) == 0,
			IsConnected: true,
			IsReady:     false,
			Score:       0,
		}
	}

	b.Players[playerID] = newPlayer
	fmt.Printf("DEBUG: Player %s joined room %s (Host: %v)\n", name, b.ID, newPlayer.IsHost)
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
	time.Sleep(15 * time.Second)
	for _, p := range r.Players {
		if !p.IsConnected {
			isAllDisconnect = true
		} else {
			isAllDisconnect = false
			break
		}
	}
	var isReadyCount int
	for _, p := range r.Players {
		if p.IsReady {
			isReadyCount++
		}
	}

	if isReadyCount == len(r.Players) {
		select {
		case r.FinishedChan <- true:
		default:
		}
	}
		
	return isAllDisconnect
}

func (b *BaseRoom) JoinGame(playerID string, newConn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if oldPlayer, ok := b.Players[playerID]; ok {
		oldPlayer.Conn = newConn
		oldPlayer.IsConnected = true
	}
}
