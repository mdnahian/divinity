package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CoreClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewCoreClient(baseURL string) *CoreClient {
	return &CoreClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *CoreClient) SetToken(token string) {
	c.token = token
}

type SpawnResponse struct {
	Token       string `json:"token"`
	NPCID       string `json:"npc_id"`
	Name        string `json:"name"`
	Profession  string `json:"profession"`
	Personality string `json:"personality"`
	Location    string `json:"location"`
	HP          int    `json:"hp"`
}

type ResumeResponse struct {
	NPCID      string `json:"npc_id"`
	Name       string `json:"name"`
	Profession string `json:"profession"`
	Location   string `json:"location"`
	HP         int    `json:"hp"`
	Alive      bool   `json:"alive"`
	Busy       bool   `json:"busy"`
}

type ToolsResponse struct {
	NPCID string                   `json:"npc_id"`
	Name  string                   `json:"name"`
	Busy  bool                     `json:"busy"`
	Tools []map[string]interface{} `json:"tools"`
}

type ToolCallResponse struct {
	Tool   string `json:"tool"`
	Result string `json:"result"`
}

type CommitResponse struct {
	Status   string `json:"status"`
	ActionID string `json:"action_id"`
	NPCID    string `json:"npc_id"`
}

type StateResponse struct {
	NPCID        string  `json:"npc_id"`
	Name         string  `json:"name"`
	Alive        bool    `json:"alive"`
	Busy         bool    `json:"busy"`
	BusyUntil    int     `json:"busy_until_tick"`
	CurrentTick  int     `json:"current_tick"`
	PendingAction string `json:"pending_action"`
	Location     string  `json:"location"`
	HP           int     `json:"hp"`
	Hunger       float64 `json:"hunger"`
	Thirst       float64 `json:"thirst"`
	Fatigue      float64 `json:"fatigue"`
}

func (c *CoreClient) Spawn() (*SpawnResponse, error) {
	body, err := c.post("/api/agent/spawn", nil, false)
	if err != nil {
		return nil, err
	}
	var resp SpawnResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode spawn response: %w", err)
	}
	return &resp, nil
}

func (c *CoreClient) Resume() (*ResumeResponse, error) {
	body, err := c.post("/api/agent/resume", nil, true)
	if err != nil {
		return nil, err
	}
	var resp ResumeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode resume response: %w", err)
	}
	return &resp, nil
}

func (c *CoreClient) GetTools() (*ToolsResponse, error) {
	body, err := c.get("/api/agent/tools")
	if err != nil {
		return nil, err
	}
	var resp ToolsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode tools response: %w", err)
	}
	return &resp, nil
}

func (c *CoreClient) CallTool(tool string, args json.RawMessage) (*ToolCallResponse, error) {
	payload := map[string]interface{}{
		"tool": tool,
		"args": args,
	}
	body, err := c.post("/api/agent/call", payload, true)
	if err != nil {
		return nil, err
	}
	var resp ToolCallResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode tool call response: %w", err)
	}
	return &resp, nil
}

func (c *CoreClient) CommitAction(actionID, target, dialogue, goal, location, reason string) (*CommitResponse, error) {
	payload := map[string]string{
		"action_id": actionID,
		"target":    target,
		"dialogue":  dialogue,
		"goal":      goal,
		"location":  location,
		"reason":    reason,
	}
	body, err := c.post("/api/agent/commit", payload, true)
	if err != nil {
		return nil, err
	}
	var resp CommitResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode commit response: %w", err)
	}
	return &resp, nil
}

func (c *CoreClient) GetState() (*StateResponse, error) {
	body, err := c.get("/api/agent/state")
	if err != nil {
		return nil, err
	}
	var resp StateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode state response: %w", err)
	}
	return &resp, nil
}

type PromptResponse struct {
	NPCID        string `json:"npc_id"`
	Name         string `json:"name"`
	SystemPrompt string `json:"system_prompt"`
	UserPrompt   string `json:"user_prompt"`
}

func (c *CoreClient) GetPrompt() (*PromptResponse, error) {
	body, err := c.get("/api/agent/prompt")
	if err != nil {
		return nil, err
	}
	var resp PromptResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode prompt response: %w", err)
	}
	return &resp, nil
}

func (c *CoreClient) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d: %s", path, resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *CoreClient) post(path string, payload interface{}, auth bool) ([]byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequest("POST", c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", path, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST %s: status %d: %s", path, resp.StatusCode, string(body))
	}
	return body, nil
}
