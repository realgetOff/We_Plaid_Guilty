package gamemanager

import (
	"fmt"
)

/*
* Verify if player is currently connected
*/
func (r *Room) SetPlayerStatus(playerID int, connected bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.Players[playerID] 
	if !ok { return }

	p.isConnected = connected

	if !p.isConnected && r.Status != StateWaiting {
		p.IsReady = true

		readyCount := 0
		for _, tmp := range r.Players {
			if tmp.IsReady {
				readyCount++
			}
		}

		if readyCount == len(r.Players) {
			select {
				case r.FinishedChan <- true:
				default:
			}
		}
	}
}

/*
* Create a new Room.
 */
func NewRoom(id string, timer int, totalRound int) (*Room) {
	return &Room {
		ID: id,
		Status: StateWaiting,
		Phase: "waiting",
		Timer: timer,
		TotalRound: totalRound,
		CurrentRound: 0,
		MessageChan: make(chan Notification, 100),
		Players: make(map[int]*Player),
		Books: make(map[int]*Book),
		PlayerOrder: []int{},
		FinishedChan: make(chan bool),
	}
}

/*
* Add player in the room.
*/
func (r *Room) AddPlayer(id int, name string) (error){
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != StateWaiting {
		return fmt.Errorf("Game started, too late!\n")
	}
	_, ok := r.Players[id]
	if ok == true {
		return fmt.Errorf("Player %s is already in the game!\n", name)
	}

	isFirst := len(r.Players) == 0

	newPlayer := &Player{
		ID: id,
		Name: name,
		IsHost: isFirst,
		isConnected: true,
		IsReady: false,
	}

	r.Players[id] = newPlayer
	r.PlayerOrder = append(r.PlayerOrder, id)

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
		if tmp.isConnected == false {
			newEntry := Entry{
				AuthorID: tmp.ID,
			}

			if r.Phase == string(StateWriting) || r.Phase == string(StateGuess) {
				newEntry.Type = "TEXT"
				newEntry.Content = "..."
			} else {
				newEntry.Type = "IMAGE"
				newEntry.Content = "EMPTY_IMAGE"
			}

			if ptrBook, ok := r.Books[tmp.ID]; ok {
				ptrBook.Entries = append(ptrBook.Entries, newEntry)
			}

			if !tmp.IsReady {
				tmp.IsReady = true
			}
		}
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
