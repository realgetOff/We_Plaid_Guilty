package gamemanager

import "sync"

const (
	StateAIWaiting  GameStates = "ai_waiting"
	StateAIDrawing  GameStates = "ai_draw"
	StateAIVoting   GameStates = "ai_vote"
	StateAIFinished GameStates = "ai_finished"
)

type AIDrawing struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Drawing    string `json:"drawing"`
}

type AIVote struct {
	VoterID  string `json:"voter_id"`
	TargetID string `json:"target_id"`
	Score    int    `json:"score"`
}

const StateAIGallery = "gallery"

type AIResult struct {
	PlayerID   string  `json:"player_id"`
	PlayerName string  `json:"player_name"`
	Drawing    string  `json:"drawing"`
	Score      float64 `json:"score"`
}

type AIRoom struct {
	ID           string
	Status       GameStates
	Prompt       string
	Players      map[string]*Player
	Drawings     map[string]*AIDrawing
	Votes        []AIVote
	DrawingsDone int
	VotesDone    int
	DrawChan     chan bool
	VoteChan     chan bool
	MessageChan  chan Notification
	mu           sync.Mutex
}
