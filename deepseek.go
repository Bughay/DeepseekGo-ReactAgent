package deepseek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatTemplate struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	Stream         bool            `json:"stream"`
	Temperature    float64         `json:"temperature"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

const deepseekURL = "https://api.deepseek.com/v1/chat/completions"

func DeepseekOneshot(systemMessage string, userMessage string, temperature float64) (string, error) {
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
			{Role: "system", Content: systemMessage},
			{Role: "user", Content: userMessage},
		},
		Stream:      false,
		Temperature: temperature,
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

func DeepseekOneshotJSON(messages []Message, temperature float64) (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", fmt.Errorf("load .env: %w", err)
	}

	apiKey := os.Getenv("DEEPSEEKAPIKEY")
	if apiKey == "" {
		return "", fmt.Errorf("DEEPSEEKAPIKEY not set")
	}

	chat := &ChatTemplate{
		Model:          "deepseek-chat",
		Messages:       messages,
		Stream:         false,
		Temperature:    temperature,
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	}

	jsonData, _ := json.Marshal(chat)
	req, _ := http.NewRequest(http.MethodPost, deepseekURL, bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second} // 🔥 increased
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	trimmed := strings.TrimSpace(string(bodyBytes))
	if trimmed == "" {
		return "", fmt.Errorf("empty response body")
	}

	var response ChatResponse
	if err := json.Unmarshal([]byte(trimmed), &response); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := response.Choices[0].Message.Content
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("empty content in response")
	}

	return content, nil
}
