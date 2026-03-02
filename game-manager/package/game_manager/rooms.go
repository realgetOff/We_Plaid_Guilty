package gamemanager

import (
	"fmt"
)

/*
* Create a new Room.
*/
func NewRoom(id string) (*Room) {
	return &Room {
		ID: id,
		Status: StateWaiting,
		Players: make(map[int]*Player),
	}
}

/*
* Add player in the room.
*/
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

/*
* Reset the status player (Ready (to)-> NotReady)
*/
func (r *Room) resetPlayer() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, tmp := range r.Players {
		tmp.IsReady = false
	}
}

/*
* Update the Room status: { Waiting, Writing, Drawing, Finished }.
*/
func (r *Room) updateStatus(newStatus GameStates) {
	r.mu.Lock()
	r.Status = newStatus
	r.mu.Unlock()
}


