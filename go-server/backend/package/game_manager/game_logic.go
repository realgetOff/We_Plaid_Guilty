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

		if round != 1 {
			r.resetPlayer()
		}

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
		r.forceValidation()
	}
	r.updateStatus(StateFinished)
	fmt.Printf("GG everyone game end !")
}

/*
* Force the validation if the player timeout
*/
func (r * Room) forceValidation() {
	r.mu.Lock()
	defer r.mu.Unlock()

	content := ""
	tType := ""
	for _, tmp := range r.Players {
		if tmp.IsReady == false {
		
			if r.Status == StateWriting {
				tType = "TEXT"
				content = "..."
			} else {
				tType = "IMAGE"
				content = "EMPTY_IMAGE"
			}

			if tmp.LastDraft != "" {
				content = tmp.LastDraft
			}
			
			newEntry := Entry{
				AuthorID: tmp.ID,
				Content: content,
				Type: tType,
			}

			
			ptrBook, ok := r.Books[tmp.ID]
			if !ok {
				continue
			}

			ptrBook.Entries = append(ptrBook.Entries, newEntry)

			tmp.IsReady = true
		}
	}
}

/*
* Fill in the notebook with the player’s ID and the text that is either an IMAGE or a SENTENCE
*/
func (r *Room) SubmiteAction(playerID int, text string, isFinal bool) {

	r.mu.Lock()
	defer r.mu.Unlock()

	player, ok := r.Players[playerID]
	if !ok {
		return
	}

	player.LastDraft = text

	if !isFinal {
		return
	}

	ptrBook, ok := r.Books[playerID]
	if !ok {
		return
	}

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
		select {
		case r.FinishedChan <- true:
		default:
		}
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

func (r *Room) GetPlayerTask(playerID int) (taskString string, content string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	val, ok := r.Books[playerID]
	if !ok { return "", "" }

	lenEntries := len(val.Entries)
	if lenEntries == 0 {
		return "TEXT", ""
	} else if lenEntries > 0 {
		last := val.Entries[lenEntries - 1]
		if last.Type == "TEXT" {
			return "IMAGE", last.Content
		} else {
			return "TEXT", last.Content
		}
	}
	return "", ""
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
