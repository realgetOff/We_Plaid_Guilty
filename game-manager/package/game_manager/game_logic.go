package gamemanager

import (
	"fmt"
	"math/rand/v2"
	"time"
)

/*
* This function waiting if every player finish before the timer.
*/
func (r *Room) waitForPhase(timeout time.Duration) {
	timer := time.After(timeout)

	select {
	case <-timer:
		fmt.Printf("Time out !\n")
	case <-r.FinishedChan:
		fmt.Printf("Everybody finished !\n")
	}
}

/*
* The RunGameLoop function manages the main game cycle of a room.
* It alternates between writing and drawing phases for several rounds.
* When all rounds are completed, it sets the game state to finished.
*/
func (r *Room) RunGameLoop() {

	TotalRound := len(r.Players)

	if TotalRound % 2 == 0 {
		TotalRound++
	}

	for round := 1; round <= TotalRound; round++ {
		r.resetPlayer()

		r.mu.Lock()
		r.FinishedChan = make(chan bool, len(r.Players))
		r.mu.Unlock()

		if round > 1 {
			r.rotateBook()
		}

		if (round % 2) != 0 {
			r.updateStatus(StateWriting)
			fmt.Printf("Round %d : Writing starting...\n", round)
			r.waitForPhase(45 * time.Second)
		} else {
			r.updateStatus(StateDrawing)
			fmt.Printf("Round %d : Drawing starting...\n", round)
			r.waitForPhase(90 * time.Second)
		}
	}
	r.updateStatus(StateFinished)
	fmt.Printf("GG everyone game end !")
}

/*
* Fill in the notebook with the player’s ID and the text that is either an IMAGE or a SENTENCE
*/
func (r *Room) SubmiteAction(playerID int, text string) {
	
	r.mu.Lock()
	defer r.mu.Unlock()

	ptrBook := r.Books[playerID]

	newEntry := Entry{
		AuthorID: playerID,
		Content: text,
	}
	if r.Status == StateWriting {
		newEntry.Type = "TEXT"
	} else {
		newEntry.Type = "IMAGE"
	}

	ptrBook.Entries = append(ptrBook.Entries, newEntry)

	r.Players[playerID].IsReady = true

	readyCount := 0
	for _, tmp := range r.Players {
		if tmp.IsReady == true {
			readyCount++
		}
	}
	if readyCount == len(r.Players) {
		r.FinishedChan <- true
	}
}

/*
* This function allows a rotation of the notebooks.
* This allows each player to write or draw in the sequence.
*/
func (r *Room) rotateBook() {
	r.mu.Lock()
	defer r.mu.Unlock()

	nextBook := make(map[int]*Book)

	for i, donorPlayerID := range r.PlayerOrder {
		nextIndex := (i + 1) % len(r.PlayerOrder)
		catcherPlayerID := r.PlayerOrder[nextIndex]
		nextBook[catcherPlayerID] = r.Books[donorPlayerID]
	}

	r.Books = nextBook
}

/*
* Attention, all players check if they can enter or not. 
* Update the status of the roon from waiting to writing and launch the loop game.
*/
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

	r.Books = make(map[int]*Book)
	r.PlayerOrder = []int{}

	for playerID := range r.Players {
			r.Books[playerID] = &Book{
				OwnerID: playerID,
		}
		r.PlayerOrder = append(r.PlayerOrder, playerID)
	}

	rand.Shuffle(len(r.PlayerOrder), func(j, i int) {
		r.PlayerOrder[i], r.PlayerOrder[j] = r.PlayerOrder[j], r.PlayerOrder[i]
	})

	r.FinishedChan = make(chan bool, len(r.Players))

	go r.RunGameLoop()

	return nil
}
