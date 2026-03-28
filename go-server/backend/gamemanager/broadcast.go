package gamemanager

import "fmt"

func (r *Room) BroadcastLobbyState() {
    r.mu.Lock()
    
    playerList := make([]map[string]interface{}, 0)
    type toNotify struct {
        id   int
        name string
        host bool
    }
    targets := make([]toNotify, 0)

    for _, p := range r.Players {
        playerList = append(playerList, map[string]interface{}{
            "id":     p.ID,
            "name":   p.Name,
            "host":   p.IsHost,
            "online": p.isConnected,
        })
        targets = append(targets, toNotify{id: p.ID, name: p.Name, host: p.IsHost})
    }
    
    roomID := r.ID
    r.mu.Unlock()

    for _, target := range targets {
        
        r.MessageChan <- Notification{
            PlayerID: target.id,
            Data: map[string]interface{}{
                "type":    "lobby_state",
                "room":    roomID,
                "players": playerList,
                "me": map[string]interface{}{
                    "id":   target.id,
                    "name": target.name,
                    "host": target.host,
                },
            },
        }
    }
}

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
		Steps []Step `json:"steps"`
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
			chain.Steps = append(chain.Steps, Step{
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
