package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type LLMClient struct {
	endpoint    string
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	httpClient  *http.Client
}

func NewLLMClient(endpoint, apiKey, model string, maxTokens int, temperature float64, timeoutMs int) *LLMClient {
	return &LLMClient{
		endpoint:    endpoint,
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		},
	}
}

type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolSpec struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Tools       []ToolSpec    `json:"tools,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

type LLMResponse struct {
	Content    string
	ToolCalls  []ToolCall
	StopReason string
}

func (c *LLMClient) ChatComplete(ctx context.Context, messages []ChatMessage, tools []ToolSpec) (*LLMResponse, error) {
	body := chatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: c.temperature,
	}
	if c.maxTokens > 0 {
		body.MaxTokens = c.maxTokens
	}
	if len(tools) > 0 {
		body.Tools = tools
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w (body: %s)", err, string(respBody))
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	choice := chatResp.Choices[0]

	for i, tc := range choice.Message.ToolCalls {
		if tc.Type == "" {
			choice.Message.ToolCalls[i].Type = "function"
		}
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &raw); err != nil {
			choice.Message.ToolCalls[i].Function.Arguments = tc.Function.Arguments
		}
	}

	return &LLMResponse{
		Content:    stripThinkTags(choice.Message.Content),
		ToolCalls:  choice.Message.ToolCalls,
		StopReason: choice.FinishReason,
	}, nil
}

var thinkRe = regexp.MustCompile(`(?s)<think>.*?</think>`)

func stripThinkTags(s string) string {
	s = thinkRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
