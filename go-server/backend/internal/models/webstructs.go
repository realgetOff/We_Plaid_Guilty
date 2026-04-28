package models

type LobbySettings struct {
	Rounds int `json:"rounds"`
}

type playerNameTemp struct {
	PlayerName string `json:"playerName"`
}

type CreateLobbyRequest struct {
	HostID   string        `json:"hostId"`
	Settings LobbySettings `json:"settings"`
}

type CreateLobbyResponse struct {
	LobbyCode string `json:"lobbyCode"`
}