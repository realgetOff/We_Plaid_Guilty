package gamemanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	// "os"
	"net/http"
)


// const grokAPIURL = "https://api.mistral.ai/v1/chat/completions"
// const grokModel  = "open-mistral-7b"
// const API_KEY = "0wZeVeQFag9DGevdgPZh1szVbMidYSuw"
const grokAPIURL = "https://api.x.ai/v1/chat/completions"
const grokModel  = "grok-4-1-fast"

// const grokAPIURL = "https://api.groq.com/openai/v1/chat/completions"
// const grokModel = "llama-3.1-8b-instant"
const API_KEY = "xai-qQnbfo3xYj0Fw28NaAlGTipAHBLdKiVp7AkGEjyyT5qJ78RJxjU2k7rCuDKBnY68NowCMo3do4khFcYP"

type grokMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type grokRequest struct {
	Model    string        `json:"model"`
	Messages []grokMessage `json:"messages"`
}

type grokChoice struct {
	Message grokMessage `json:"message"`
}

type grokResponse struct {
	Choices []grokChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func CallAI(prompt string) (string, error) {
	apiKey := API_KEY
	if apiKey == "" {
		return "", fmt.Errorf("API_KEY not set")
	}

	// reqBody := grokRequest{
		// Model: grokModel,
		// Messages: []grokMessage{
			// {
				// Role:    "user",
				// Content: prompt,
			// },
		// },
	// }
		reqBody := grokRequest{
		Model: grokModel,
		Messages: []grokMessage{
			{
				Role:    "system",
				Content: "Tu es un générateur de prompts pour un jeu de dessin, satirique et noir, style South Park ou Les Guignols de l'info. Tu génères UN SEUL énoncé choquant, provocateur, tabou, humour noir, gore et trash. UNE SEULE PHRASE, maximum 10 mots, sans explication, sans guillemets, sans mise en forme. Garde a l'esprit que les personnes doivent exposer leurs inventions devant les autres personnes",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", grokAPIURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type",  "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call API: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var grokResp grokResponse
	if err := json.Unmarshal(respBytes, &grokResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if grokResp.Error != nil {
		return "", fmt.Errorf("API error: %s", grokResp.Error.Message)
	}

	if len(grokResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return grokResp.Choices[0].Message.Content, nil
}
