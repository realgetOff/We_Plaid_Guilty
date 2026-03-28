package gamemanager

import (
	"fmt"
	"math/rand"
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

func (r *Room) listenForNotifaction() {
	for notification := range r.MessageChan {
        
        r.mu.Lock()
        player, ok := r.Players[notification.PlayerID]
        r.mu.Unlock()

        if !ok {
            continue
        }
		player.WriteMu.Lock()
        err := player.Conn.WriteJSON(notification.Data)
		player.WriteMu.Unlock()
        if err != nil {
            fmt.Printf("DEBUG: Erreur WriteJSON: %v\n", err)
        }
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
		r.CurrentRound = round
		r.FinishedChan = make(chan bool, 1)
		r.mu.Unlock()


		if round == 1 {
			r.Phase = string(StateWriting)
		} else if (round % 2) != 0{
			r.Phase = string(StateGuess)
		} else {
			r.Phase = string(StateDrawing)
		}

		if round > 1 {
			r.rotateBook()
		}

		for _, p := range r.Players {
			task := r.GetPlayerTask(p.ID)
			r.MessageChan <- Notification{
				PlayerID: p.ID,
				Data: task,
			}
		}

		if r.Phase == string(StateDrawing) {
			fmt.Printf("Round %d : Drawing starting...\n", round)
			r.waitForPhase(90 * time.Second)
		} else {
			r.waitForPhase(45 * time.Second)
			fmt.Printf("Round %d : Writting starting...\n", round)
		}

		r.forceValidation()
	}

	r.mu.Lock()
	r.Phase = "gallery"
	r.mu.Unlock()

	r.updateStatus(StateFinished)

	r.broadcastGallery()

	fmt.Printf("GG everyone game end !")
}

/*
* Force the validation if the player timeout
*/
func (r * Room) forceValidation() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, tmp := range r.Players {

		if tmp.IsReady == false {
		
			content := "DEFAULT_IMAGE"
			tType := "IMAGE"
			if r.Phase == string(StateWriting) || r.Phase == string(StateGuess) {
				tType = "TEXT"
				content = "..."
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
			r.Players[tmp.ID].LastDraft = ""
		}
	}
}

/*
* Fill in the notebook with the player’s ID and the text that is either an IMAGE or a SENTENCE
*/
func (r *Room) SubmiteAction(playerID int, data map[string]interface{}, isFinal bool) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	player, ok := r.Players[playerID]
	if !ok {
		return fmt.Errorf("Player not found!")
	}

	var content string
	if val, ok := data["prompt"].(string); ok {
		content = val
	} else if val, ok := data["drawing"].(string); ok {
		content = val
	} else if val, ok := data["guess"].(string); ok {
		content = val
	}

	player.LastDraft = content

	if !isFinal {
		return nil
	}

	ptrBook, ok := r.Books[playerID]
	if !ok {
		return fmt.Errorf("Book not found!")
	}

	newEntry := Entry{
		AuthorID: playerID,
		Content: content,
	}
	if r.Phase == string(StateWriting) || r.Phase == string(StateGuess) {
		newEntry.Type = "TEXT"
	} else {
		newEntry.Type = "IMAGE"
	}

	ptrBook.Entries = append(ptrBook.Entries, newEntry)

	r.Players[playerID].IsReady = true

	r.Players[playerID].LastDraft = ""

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
	return nil
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

func (r *Room) GetPlayerTask(playerID int) GameStateRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := GameStateRecord {
		Type: "game_state",
		Phase: r.Phase,
		Room: r.ID,
	}

	val, ok := r.Books[playerID]
	if !ok { return res }
	lenEntries := len(val.Entries)

	if lenEntries == 0 {

		return res
	} else if lenEntries > 0 {

		last := val.Entries[lenEntries - 1]

			if last.Type == "TEXT" {
			res.Prompt = last.Content
		} else {
			res.Drawing = last.Content
		}
	}
	return res
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

	r.Status = "started"
	r.Phase = string(StateWriting)
	r.CurrentRound = 1

	r.Books = make(map[int]*Book)
	r.PlayerOrder = []int{}

	for playerID := range r.Players {
			r.Books[playerID] = &Book{
				OwnerID: playerID,
		}
		r.PlayerOrder = append(r.PlayerOrder, playerID)
	}

	rand.Shuffle(len(r.PlayerOrder), func(j, i int) {
		rand.Seed(time.Now().UnixNano())
		r.PlayerOrder[i], r.PlayerOrder[j] = r.PlayerOrder[j], r.PlayerOrder[i]
	})

	go r.RunGameLoop()

	return nil
}
