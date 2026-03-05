package gamemanager

import "fmt"

func (r *Room) broadcastGallery() {
	r.mu.Lock()
	defer r.mu.Unlock()

	type Step struct {
		Type string `json:"type"`
		Content string `json:"content"`
	}
	type Chain struct {
		ID string `json:"id"`
		Prompt string `json:"prompt"`
		Step []Step `json:"step"`
	}

	var allChain []Chain

	for _, book := range r.Books {

		if len(book.Entries) == 0 {
			continue
		}

		chain := Chain{
			ID: fmt.Sprintf("chain-%d", book.OwnerID),
			Prompt: book.Entries[0].Content,
		}

		for i := 1; i < len(book.Entries); i++{
			entryType := "drawing"
			if book.Entries[i].Type == "TEXT" {
				entryType = "guess"
			}
			chain.Step = append(chain.Step, Step{
				Type: entryType,
				Content: book.Entries[i].Content,
			})
		}
		allChain = append(allChain, chain)
	}

	finalData := map[string]interface{}{
		"type": "game_state",
		"phase": "gallery",
		"room": r.ID,
		"chains": allChain,
	}

	for _, p := range r.Players {
		r.MessageChan <- Notification{
			PlayerID: p.ID,
			Data: finalData,
		}
	}
}
