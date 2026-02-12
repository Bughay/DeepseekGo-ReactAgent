package deepseekgoreactagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatTemplate struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func DeepseekOneshot(userMessage string) (string, error) {
	const deepseekURL = "https://api.deepseek.com/v1/chat/completions"
	if err := godotenv.Load(); err != nil {
		return "", fmt.Errorf("load .env: %w", err)
	}

	apiKey := os.Getenv("DEEPSEEKAPIKEY")
	if apiKey == "" {
		return "", fmt.Errorf("DEEPSEEKAPIKEY not set")
	}

	chat := &ChatTemplate{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "system", Content: "you are a helpful assistant"},
			{Role: "user", Content: userMessage},
		},
		Stream:      false,
		Temperature: 0.5,
	}

	jsonData, err := json.Marshal(chat)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, deepseekURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}
