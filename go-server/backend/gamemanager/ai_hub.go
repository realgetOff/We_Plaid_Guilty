	package gamemanager

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type AIHub struct {
	Rooms map[string]*AIRoom
	mu    sync.RWMutex
}

func (h *AIHub) generateRandID(length int) string {
	ran_str := make([]byte, length)
	for i := 0; i < length; i++ {
		ran_str[i] = charset[rand.Intn(len(charset))]
	}
	return string(ran_str)
}

func (h *AIHub) GetRoom(id string) (*AIRoom, error) {
	id = strings.ToUpper(strings.TrimSpace(id))
	h.mu.Lock()
	defer h.mu.Unlock()
	ptr, ok := h.Rooms[id]
	if !ok {
		return nil, fmt.Errorf("ai room %s not found", id)
	}
	return ptr, nil
}

func (h *AIHub) CreateRoom() *AIRoom {
	rand.Seed(time.Now().UnixNano())
	var id string
	for {
		id = h.generateRandID(6)
		h.mu.Lock()
		_, exists := h.Rooms[id]
		h.mu.Unlock()
		if !exists {
			break
		}
	}

	room := NewAIRoom(id)
	go room.listenForNotification()

	h.mu.Lock()
	h.Rooms[id] = room
	h.mu.Unlock()

	fmt.Printf("AIHub: room %s created\n", id)
	return room
}

func (h *AIHub) DeleteRoom(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.Rooms, id)
}

func (h *AIHub) AddPlayerToRoom(roomID string, playerID string, name string, conn *websocket.Conn) (*AIRoom, error) {
	room, err := h.GetRoom(roomID)
	if err != nil {
		return nil, err
	}

	room.mu.Lock()
	if len(room.Players) >= 8 {
		room.mu.Unlock()
		return nil, fmt.Errorf("room is full")
	}
	room.Players[playerID] = &Player{
		ID:          playerID,
		Name:        name,
		Conn:        conn,
		IsHost:      len(room.Players) == 0,
		IsConnected: true,
	}
	room.mu.Unlock()

	return room, nil
}

func (h *AIHub) UpdatePlayerConn(roomID string, playerID string, conn *websocket.Conn) {
    room, err := h.GetRoom(roomID)
    if err != nil { return }
    room.mu.Lock()
    if p, ok := room.Players[playerID]; ok {
        p.IsConnected = true
        p.Conn = conn
    }
    room.mu.Unlock()
}
