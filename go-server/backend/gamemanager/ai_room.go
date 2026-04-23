package gamemanager

import (
	"fmt"
	"time"

)

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

func (r *AIRoom) SubmitDrawing(playerID string, drawing string, title string, description string) {
	r.mu.Lock()
	name := ""
	if p, ok := r.Players[playerID]; ok {
		name = p.Name
	}
	r.Drawings[playerID] = &AIDrawing{
		PlayerID:    playerID,
		PlayerName:  name,
		Drawing:     drawing,
		Title:       title,
		Description: description,
	}
	r.DrawingsDone = len(r.Drawings)
	total := len(r.Players)

	p := r.Players[playerID]
	p.IsReady = true
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

	p := r.Players[voterID]
	p.IsReady = true
	r.mu.Unlock()
	if r.VotesDone >= total {
		select {
		case r.VoteChan <- true:
		default:
		}
	}
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
		userName = "📢 System"
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
				"color": p.Color,
				"font": p.Font,
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

	var isReadyCount int

	for _, p := range r.Players {
		if p.IsReady {
			isReadyCount++
		}
	}
	if isReadyCount == len(r.Players) {
		if r.Status == StateAIVoting {
			select {
			case r.VoteChan <- true:
			default:
			}
		}
		
		if r.Status == StateAIDrawing {
			select {
			case r.DrawChan <- true:
			default:
			}
		}
	}

	return isAllDisconnect
}


func (r *AIRoom) RunAIGameLoop(prompt string) {
	r.mu.Lock()
	r.Status = StateAIDrawing
	r.Prompt = prompt
	r.Drawings = make(map[string]*AIDrawing)
	r.Votes = []AIVote{}
	r.DrawingsDone = 0
	r.VotesDone = 0
	r.mu.Unlock()

	r.BroadcastToAll(map[string]interface{}{
		"type":   "ai_game_state",
		"phase":  "draw",
		"prompt": prompt,
	})

	<-r.DrawChan

	r.mu.Lock()
	r.Status = StateAIVoting
	roomID := r.ID
	allDrawings := make([]*AIDrawing, 0, len(r.Drawings))
	for _, d := range r.Drawings {
		allDrawings = append(allDrawings, d)
	}

	playerIDs := make([]string, 0, len(r.Players))
	for id := range r.Players {
		playerIDs = append(playerIDs, id)
	}
	r.mu.Unlock()

	for _, pID := range playerIDs {
		filteredList := make([]map[string]interface{}, 0)
		for _, d := range allDrawings {
			if d.PlayerID != pID { 
				filteredList = append(filteredList, map[string]interface{}{
					"player_id": d.PlayerID,
					"name":      d.PlayerName,
					"drawing":   d.Drawing,
				})
			}
		}

		r.MessageChan <- Notification{
			PlayerID: pID,
			Data: map[string]interface{}{
				"type":     "ai_vote_state",
				"phase":    "vote",
				"drawings": filteredList,
				"room":     roomID,
			},
		}
	}

	<-r.VoteChan

	results := r.ComputeResults()

	r.mu.Lock()
	r.Status = StateAIFinished
	r.mu.Unlock()

	fmt.Printf("DEBUG: gallery\n")
	r.BroadcastToAll(map[string]interface{}{
		"type":    "ai_results",
		"phase":   "gallery",
		"results": results,
	})
	fmt.Printf("DEBUG: Etape5\n")
}


func (r *AIRoom) ComputeResults() []AIResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	scoresMap := make(map[string][]int)
	for _, v := range r.Votes {
		scoresMap[v.TargetID] = append(scoresMap[v.TargetID], v.Score)
	}

	results := make([]AIResult, 0)
	for pID, d := range r.Drawings {
		avg := 0.0
		if s, ok := scoresMap[pID]; ok && len(s) > 0 {
			sum := 0
			for _, val := range s {
				sum += val
			}
			avg = float64(sum) / float64(len(s))
		}
		results = append(results, AIResult{
			PlayerID:   pID,
			PlayerName: d.PlayerName,
			Drawing:    d.Drawing,
			Score:      avg,
		})
		fmt.Printf("DEBUG ComputeResults: %d drawings, %d votes\n", len(r.Drawings), len(r.Votes))
	}
	return results
}
