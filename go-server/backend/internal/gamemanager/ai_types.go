package gamemanager

const (
	StateAIWaiting  GameStates = "ai_waiting"
	StateAIDrawing  GameStates = "ai_draw"
	StateAIVoting   GameStates = "ai_vote"
	StateAIFinished GameStates = "ai_finished"
)

type AIDrawing struct {
	PlayerID    string `json:"player_id"`
	PlayerName  string `json:"player_name"`
	Drawing     string `json:"drawing"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type AIVote struct {
	VoterID  string `json:"voter_id"`
	TargetID string `json:"target_id"`
	Score    int    `json:"score"`
}

const StateAIGallery = "gallery"

type AIResult struct {
	PlayerID    string  `json:"player_id"`
	PlayerName  string  `json:"player_name"`
	Drawing     string  `json:"drawing"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
}
