package gamemanager

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Hub struct {
	Rooms map[string]*Room
	mu sync.RWMutex
}

func (h *Hub) generateRandID(lenght int) (roomId string) {

	ran_str := make([]byte, lenght)

	for i:= 0; i < lenght; i++ {
		ran_str[i] = charset[rand.Intn(len(charset))]
	}

	return string(ran_str)
}

func (h *Hub) GetRoom(id string) (* Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ptr, err := h.Rooms[id]
	if err {
		return nil, fmt.Errorf("invalid id %d", id)
	}
	return ptr, nil
}

func (h *Hub) CreateRoom() (* Room){
	var R *Room
	var IdRoom string
	rand.Seed(time.Now().UnixNano())

	for {
		IdRoom = h.generateRandID(6)

		h.mu.Lock()
		_, ok := h.Rooms[IdRoom]
		h.mu.Unlock()

		if ok {
			continue
		} else {
			R = NewRoom(IdRoom, 60, 0)
			break
		}
	}

	h.mu.Lock()
	h.Rooms[IdRoom] = R
	h.mu.Unlock()

	return R
}
