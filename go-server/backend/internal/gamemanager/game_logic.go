package gamemanager

import (
	"fmt"
	"math/rand"
	"time"
)

func (r *Room) waitForPhase(timeout time.Duration) {
	timer := time.After(timeout)

	select {
	case <-timer:
		fmt.Printf("Timeout!\n")
	case <-r.FinishedChan:
		fmt.Printf("Everybody's finished!\n")
	}
}

func (b *BaseRoom) listenForNotifaction() {

	for notification := range b.MessageChan {

		if (notification.End == true) {
			break
		}
		b.mu.Lock()
		player, ok := b.Players[notification.PlayerID]
		b.mu.Unlock()

		if !ok || player == nil || player.Conn == nil{
			continue
		}
		conn := player.Conn

		player.WriteMu.Lock()
		err := conn.WriteJSON(notification.Data)
		player.WriteMu.Unlock()

		if err != nil {
			fmt.Printf("DEBUG: Error WriteJSON: %v\n", err)
		}
	}
}

func (r *Room) RunGameLoop() {
	fmt.Printf("DEBUG: Start the go_routine RunGameLoop: Room %s\n", r.ID)
	TotalRound := len(r.Players)

	if TotalRound % 2 == 0 {
		TotalRound++
	}

	for round := 1; round <= TotalRound; round++ {
		r.resetPlayer()

		r.mu.Lock()
		r.FinishedChan = make(chan bool, 1)
		r.mu.Unlock()

		if round == 1 {
			r.Phase = string(StateWriting)
		} else if (round % 2) != 0 {
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
				Data:     task,
			}
		}

		fmt.Printf("Round %d : %s starting...\n", round, r.Phase)
		if r.Phase == string(StateDrawing) {
			r.waitForPhase(140 * time.Second)
		} else {
			r.waitForPhase(100 * time.Second)
		}

	}

	r.mu.Lock()
	r.Phase = "gallery"
	r.mu.Unlock()

	r.updateStatus(StateFinished)
	r.broadcastGallery()

	fmt.Printf("DEBUG: End of go_routine RunGameLoop: Room %s\n", r.ID)
}

func (r *Room) SubmitAction(playerID string, data map[string]interface{}, isFinal bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player, ok := r.Players[playerID]
	if !ok {
		return fmt.Errorf("Player not found!")
	}
	fmt.Printf("DEBUG: action_submited for player = %s\n", player.Name)

	content := ""
	if val, ok := data["prompt"].(string); ok {
		content = val
	} else if val, ok := data["drawing"].(string); ok {
		content = val
	} else if val, ok := data["guess"].(string); ok {
		content = val
	}


	if !isFinal {
		return nil
	}

	ptrBook, ok := r.Books[playerID]
	if !ok {
		return fmt.Errorf("Book not found!")
	}

	newEntry := Entry{
		AuthorID: playerID,
		Content:  content,
	}
	if r.Phase == string(StateWriting) || r.Phase == string(StateGuess) {
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
	fmt.Printf("%s valid round\n", player.Name)
	if readyCount == len(r.Players) {
		select {
		case r.FinishedChan <- true:
		default:
		}
	}
	return nil
}

func (r *Room) rotateBook() {
	r.mu.Lock()
	defer r.mu.Unlock()

	nextBook := make(map[string]*Book)

	for i, donorPlayerID := range r.PlayerOrder {

		nextIndex := (i + 1) % len(r.PlayerOrder)

		catcherPlayerID := r.PlayerOrder[nextIndex]

		nextBook[catcherPlayerID] = r.Books[donorPlayerID]
	}

	r.Books = nextBook
}

func (r *Room) GetPlayerTask(playerID string) GameStateRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := GameStateRecord{
		Type:  "game_state",
		Phase: r.Phase,
		Room:  r.ID,
	}

	val, ok := r.Books[playerID]
	if !ok {
		return res
	}

	lenEntries := len(val.Entries)

	if lenEntries == 0 {
		return res
	} else if lenEntries > 0 {
		last := val.Entries[lenEntries-1]
		if last.Type == "TEXT" {
			res.Prompt = last.Content
		} else {
			res.Drawing = last.Content
		}
	}

	return res
}

func (r *Room) StartGame() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) < 3 {
		return fmt.Errorf("3 or more players are required to play")
	}

	if r.Status != StateWaiting {
		return fmt.Errorf("The game has already started.")
	}

	r.Status = "started"
	r.Phase = string(StateWriting)

	r.Books = make(map[string]*Book)
	r.PlayerOrder = []string{}

	for playerID := range r.Players {
		r.Books[playerID] = &Book{
			OwnerID: playerID,
		}
		r.PlayerOrder = append(r.PlayerOrder, playerID)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(r.PlayerOrder), func(i, j int) {
		r.PlayerOrder[i], r.PlayerOrder[j] = r.PlayerOrder[j], r.PlayerOrder[i]
	})

	go r.RunGameLoop()

	return nil
}
