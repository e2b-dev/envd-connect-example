package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type SandboxResponse struct {
	SandboxID       string  `json:"sandboxID"`
	ClientID        string  `json:"clientID"`
	EnvdAccessToken *string `json:"envdAccessToken,omitempty"`
}

func CreateSandbox(templateID string, timeout int) (SandboxResponse, error) {
	apiKey := os.Getenv("E2B_API_KEY")
	if apiKey == "" {
		return SandboxResponse{}, fmt.Errorf("E2B_API_KEY environment variable is not set")
	}

	url := "https://api.e2b.dev/sandboxes"
	payload := map[string]interface{}{
		"templateID": templateID,
		"timeout":    timeout,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return SandboxResponse{}, fmt.Errorf("failed to marshal request payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return SandboxResponse{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return SandboxResponse{}, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return SandboxResponse{}, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var result SandboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return SandboxResponse{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if result.SandboxID == "" || result.ClientID == "" {
		return SandboxResponse{}, fmt.Errorf("invalid response: missing sandboxID or clientID")
	}

	return result, nil
}
