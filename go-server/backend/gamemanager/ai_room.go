package gamemanager

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func NewAIRoom(id string) *AIRoom {
	return &AIRoom{
		ID:          id,
		Status:      StateAIWaiting,
		Players:     make(map[string]*Player),
		Drawings:    make(map[string]*AIDrawing),
		Votes:       []AIVote{},
		DrawChan:    make(chan bool, 1),
		VoteChan:    make(chan bool, 1),
		MessageChan: make(chan Notification, 100),
	}
}

func (r *AIRoom) AddPlayer(playerID string, name string, conn *websocket.Conn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= 8 {
		return fmt.Errorf("Room is full")
	}

	newPlayer := &Player{
		ID:          playerID,
		Name:        name,
		Conn:        conn,
		IsHost:      len(r.Players) == 0,
		IsConnected: true,
		IsReady:     false,
		Score:       0,
	}

	r.Players[playerID] = newPlayer
	return nil
}

func (r *AIRoom) listenForNotification() {
    for notification := range r.MessageChan {
        r.mu.Lock()
        player, ok := r.Players[notification.PlayerID]
        if !ok {
            r.mu.Unlock()
            continue
        }
        conn := player.Conn
        writeMu := &player.WriteMu
        r.mu.Unlock()

        writeMu.Lock()
        err := conn.WriteJSON(notification.Data)
        writeMu.Unlock()
        if err != nil {
            fmt.Printf("AIRoom WriteJSON error: %v\n", err)
        }
    }
}

func (r *AIRoom) BroadcastToAll(data map[string]interface{}) {
    r.mu.Lock()
    ids := make([]string, 0, len(r.Players))
    for id := range r.Players {
        ids = append(ids, id)
    }
    roomID := r.ID
    r.mu.Unlock()

    for _, id := range ids {
        payload := make(map[string]interface{}, len(data)+1)
        for k, v := range data {
            payload[k] = v
        }
        payload["room"] = roomID
        r.MessageChan <- Notification{
            PlayerID: id,
            Data:     payload,
        }
    }
}
func (r *AIRoom) BroadcastLobbyState() {
	r.mu.Lock()
	type toNotify struct {
		id   string
		name string
		host bool
	}
	playerList := make([]map[string]interface{}, 0)
	targets    := make([]toNotify, 0)

	for _, p := range r.Players {
		playerList = append(playerList, map[string]interface{}{
			"id":     p.ID,
			"name":   p.Name,
			"host":   p.IsHost,
			"online": p.IsConnected,
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

func (r *AIRoom) UpdatePlayerConn(playerID string, conn interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.Players[playerID]; ok {
		type connSetter interface {
			WriteJSON(interface{}) error
		}
		_ = conn.(connSetter)
	}
	_ = playerID
}

func (r *AIRoom) GetPlayer(playerID string) (*Player, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.Players[playerID]
	return p, ok
}

func (r *AIRoom) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Players, playerID)
}

func (r *AIRoom) TransferHost() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.Players {
		p.IsHost = true
		break
	}
}

func (r *AIRoom) SubmitDrawing(playerID string, drawing string) {
	r.mu.Lock()
	name := ""
	if p, ok := r.Players[playerID]; ok {
		name = p.Name
	}
	r.Drawings[playerID] = &AIDrawing{
		PlayerID:   playerID,
		PlayerName: name,
		Drawing:    drawing,
	}
	r.DrawingsDone = len(r.Drawings)
	total := len(r.Players)
	r.mu.Unlock()

	if r.DrawingsDone >= total {
		select {
		case r.DrawChan <- true:
		default:
		}
	}
}

func (r *AIRoom) SubmitVotes(voterID string, votes map[string]int) {
	r.mu.Lock()
	for targetID, score := range votes {
		r.Votes = append(r.Votes, AIVote{
			VoterID:  voterID,
			TargetID: targetID,
			Score:    score,
		})
	}
	r.VotesDone++
	total := len(r.Players)
	r.mu.Unlock()

	// Déclenche dès que tous ont voté
	if r.VotesDone >= total {
		select {
		case r.VoteChan <- true:
		default:
		}
	}
}

func (r *AIRoom) ComputeResults() []AIResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	scores := make(map[string][]int)
	for _, vote := range r.Votes {
		scores[vote.TargetID] = append(scores[vote.TargetID], vote.Score)
	}

	results := []AIResult{}
	for playerID, drawing := range r.Drawings {
		avg := 0.0
		if s, ok := scores[playerID]; ok && len(s) > 0 {
			sum := 0
			for _, v := range s {
				sum += v
			}
			avg = float64(sum) / float64(len(s))
		}
		results = append(results, AIResult{
			PlayerID:   playerID,
			PlayerName: drawing.PlayerName,
			Drawing:    drawing.Drawing,
			Score:      avg,
		})
	}
	return results
}

func (r *AIRoom) SendSystemMsg(content string) {
	r.BroadcastChat("SYSTEM", content)
}

func (r *AIRoom) BroadcastChat(playerID string, content string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var userName string
	var messageId string
	var isSystem bool

	if (playerID == "SYSTEM") {
		userName = "📢 Système"
		messageId = fmt.Sprintf("%d", time.Now().UnixNano())
		isSystem = true
	} else {
		sender, ok := r.Players[playerID]
		if !ok { return }
		userName = sender.Name
		isSystem = false
	}

	messageId = fmt.Sprintf("%d", time.Now().UnixNano())

	for _,p := range r.Players {
		if !p.IsConnected {
			continue
		}

		r.MessageChan <- Notification{
			PlayerID: p.ID,
			Data: map[string]interface{}{
				"type": "ai_chat_message",
				"user": userName,
				"text": content,
				"id": messageId,
				"is_system": isSystem,
				"room": r.ID,
			},
		}
	}
}

func (r *AIRoom) LeaveGame(playerID string) (bool){
	r.mu.Lock()
	defer r.mu.Unlock()
	if oldPlayer, ok := r.Players[playerID]; ok {
		oldPlayer.IsReady = true
		oldPlayer.IsConnected = false
	}
	var isAllDisconnect bool
	for _, p := range r.Players {
		if !p.IsConnected {
			isAllDisconnect = true
		} else {
			isAllDisconnect = false
			break
		}
	}
	return isAllDisconnect
}


func (r *AIRoom) RunAIGameLoop(prompt string) {
	// 1. PHASE DE DESSIN
	r.mu.Lock()
	r.Status = StateAIDrawing
	r.Prompt = prompt
	r.Drawings = make(map[string]*AIDrawing) // Reset des dessins précédents
	r.mu.Unlock()

	r.BroadcastToAll(map[string]interface{}{
		"type":   "ai_game_state",
		"phase":  "draw",
		"prompt": prompt,
	})

	// ATTENTE : On bloque jusqu'à ce que tout le monde ait dessiné
	<-r.DrawChan

	// 2. PHASE DE VOTE
	r.mu.Lock()
	r.Status = StateAIVoting
	drawingsList := make([]map[string]interface{}, 0)
	for id, d := range r.Drawings {
		drawingsList = append(drawingsList, map[string]interface{}{
			"playerId": id,
			"name":     d.PlayerName,
			"drawing":  d.Drawing,
		})
	}
	r.mu.Unlock()

	r.BroadcastToAll(map[string]interface{}{
		"type":     "ai_vote_state",
		"phase":    "vote",
		"drawings": drawingsList,
	})

	// ATTENTE : On bloque jusqu'à ce que tout le monde ait voté
	<-r.VoteChan

	// 3. PHASE DE RÉSULTATS
	results := r.ComputeResults()
	r.mu.Lock()
	r.Status = StateAIGallery
	r.mu.Unlock()

	r.BroadcastToAll(map[string]interface{}{
		"type":    "ai_results",
		"phase":   "gallery",
		"results": results,
	})
}
