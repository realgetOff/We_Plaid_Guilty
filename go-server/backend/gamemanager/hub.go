package gamemanager

import (
	"fmt"
	"strings"
	"math/rand"
	"sync"
	"time"
)

type GameRoom interface {
	GetID() string
	GetBase() *BaseRoom
}

func (r *Room) GetID() string   { return r.ID }
func (r *Room) GetBase() *BaseRoom { return &r.BaseRoom }

func (r *AIRoom) GetID() string { return r.ID }
func (r *AIRoom) GetBase() *BaseRoom { return &r.BaseRoom }

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Hub struct {
	Rooms map[string]GameRoom
	mu sync.RWMutex
}

func (h *Hub) generateRandID(lenght int) (roomId string) {

	ran_str := make([]byte, lenght)

	for i:= 0; i < lenght; i++ {
		ran_str[i] = charset[rand.Intn(len(charset))]
	}

	return string(ran_str)
}

func (h *Hub) DeleteRoom(id string) {
    h.mu.Lock()
    defer h.mu.Unlock()
	// NOTE to acces IsAi bool h.Rooms[id].GetBase().Isi
	fmt.Printf("DEBUG: DELETE ROOM\n")
    delete(h.Rooms, id)
}

func NewAIRoom(id string) *AIRoom {
	return &AIRoom{
		BaseRoom: BaseRoom{
			ID:          id,
			Status:      StateAIWaiting,
			Players:     make(map[string]*Player),
			MessageChan: make(chan Notification, 200),
		},
		Drawings:    make(map[string]*AIDrawing),
		Votes:       []AIVote{},
		DrawChan:    make(chan bool, 1),
		VoteChan:    make(chan bool, 1),
	}
}


func NewRoom(id string) *Room {
	return &Room{
		BaseRoom: BaseRoom{
			ID:           id,
			Status:       StateWaiting,
			Players:      make(map[string]*Player),
			PlayerOrder:  []string{},
			MessageChan:  make(chan Notification, 100),
			FinishedChan: make(chan bool, 1),
		},
		Books:        make(map[string]*Book),
	}
}


func (h *Hub) GetRoom(id string) (GameRoom, error) {
	id = strings.ToUpper(strings.TrimSpace(id))
	h.mu.Lock()
	defer h.mu.Unlock()

	ptr, ok := h.Rooms[id]
	if !ok {
		return nil, fmt.Errorf("room with id %s not found", id)
	}
	
	return ptr, nil
}

func (h *Hub) LogRoom() {
	for {
		fmt.Printf("__________{NUMBER_OF_ROOM}_________\n")
		fmt.Printf("nb_rooms = %d\n", len(h.Rooms))
		for _, r := range h.Rooms {
			base := r.GetID()
			fmt.Printf("ROOM_%s\n", base)
			fmt.Printf("______________")
		}
		time.Sleep(15 * time.Second)
	}
}

func (h *Hub) CreateRoom(isAI bool) (GameRoom){
	var IdRoom string
	rand.Seed(time.Now().UnixNano())

	for {
		IdRoom = h.generateRandID(6)

		h.mu.Lock()
		_, ok := h.Rooms[IdRoom]
		h.mu.Unlock()

		if !ok {
			break
		}	
	}

	var newRoom GameRoom

	if (isAI) {
		newRoom = NewAIRoom(IdRoom)
		newRoom.GetBase().IsAi = true
	} else {
		newRoom = NewRoom(IdRoom)
		newRoom.GetBase().IsAi = false
	}



	fmt.Println("HUB: Tentative de lancement de la Goroutine...")

	go newRoom.GetBase().listenForNotifaction()

	fmt.Println("HUB: Goroutine lancée, sortie de boucle.")

	h.mu.Lock()
	h.Rooms[IdRoom] = newRoom
	h.mu.Unlock()

	return newRoom
}
