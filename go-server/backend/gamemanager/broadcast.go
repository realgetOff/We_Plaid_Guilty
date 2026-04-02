package gamemanager

import (
	"fmt"
	"time"
)

func (r *Room) SendSystemMsg(content string) {
	r.BroadcastChat("SYSTEM", content)
}

func (r *Room) BroadcastChat(playerID string, content string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var userName string
	var isSystem bool

	messageId := fmt.Sprintf("%d", time.Now().UnixNano())

	if playerID == "SYSTEM" {
		userName = "📢 System"
		isSystem = true
	} else {
		sender, ok := r.Players[playerID]
		if !ok {
			return
		}
		userName = sender.Name
		isSystem = false
	}

	for _, p := range r.Players {
		if !p.IsConnected {
			continue
		}

		r.MessageChan <- Notification{
			PlayerID: p.ID,
			Data: map[string]interface{}{
				"type":      "chat_message",
				"user":      userName,
				"text":      content,
				"id":        messageId,
				"is_system": isSystem,
				"room": r.ID,
			},
		}
	}
}

func (b *BaseRoom) BroadcastLobbyState() {
	b.mu.Lock()
	
	playerList := make([]map[string]interface{}, 0)
	type toNotify struct {
		id   string
		name string
		host bool
	}
	targets := make([]toNotify, 0)

	for _, p := range b.Players {
		playerList = append(playerList, map[string]interface{}{
			"id":     p.ID,
			"name":   p.Name,
			"host":   p.IsHost,
			"online": p.IsConnected,
		})
		targets = append(targets, toNotify{id: p.ID, name: p.Name, host: p.IsHost})
	}
	
	roomID := b.ID
	b.mu.Unlock()

	for _, target := range targets {
		
		b.MessageChan <- Notification{
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

// func (r *Room) BroadcastLobbyState() {
	// r.mu.Lock()
	// 
	// playerList := make([]map[string]interface{}, 0)
	// type toNotify struct {
		// id   string
		// name string
		// host bool
	// }
	// targets := make([]toNotify, 0)
// 
	// for _, p := range r.Players {
		// playerList = append(playerList, map[string]interface{}{
			// "id":     p.ID,
			// "name":   p.Name,
			// "host":   p.IsHost,
			// "online": p.IsConnected,
		// })
		// targets = append(targets, toNotify{id: p.ID, name: p.Name, host: p.IsHost})
	// }
	// 
	// roomID := r.ID
	// r.mu.Unlock()
// 
	// for _, target := range targets {
		// 
		// r.MessageChan <- Notification{
			// PlayerID: target.id,
			// Data: map[string]interface{}{
				// "type":    "lobby_state",
				// "room":    roomID,
				// "players": playerList,
				// "me": map[string]interface{}{
					// "id":   target.id,
					// "name": target.name,
					// "host": target.host,
				// },
			// },
		// }
	// }
// }

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
			ID: fmt.Sprintf("chain-%s", book.OwnerID),
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

func (b *BaseRoom) BroadcastToAll(data map[string]interface{}) {
	b.mu.Lock()
    ids := make([]string, 0, len(b.Players))
    for id := range b.Players {
        ids = append(ids, id)
    }
    roomID := b.ID
    b.mu.Unlock()

    for _, id := range ids {
        payload := make(map[string]interface{}, len(data)+2)
        for k, v := range data {
            payload[k] = v
        }
        payload["room"] = roomID
		payload["code"] = roomID
        b.MessageChan <- Notification{
            PlayerID: id,
            Data:     payload,
        }
    }
}

// func (r *Room) BroadcastToAll(data map[string]interface{}) {
	// r.mu.Lock()
	// ids := make([]string, 0, len(r.Players))
	// for id := range r.Players {
		// ids = append(ids, id)
	// }
	// roomID := r.ID
	// r.mu.Unlock()
// 
	// data["room"] = roomID
	// for _, id := range ids {
		// r.MessageChan <- Notification{
			// PlayerID: id,
			// Data:     data,
		// }
	// }
// }
