package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenRouterClient struct {
	url    string
	apiKey string
	model  string
	client *http.Client
}

func NewOpenRouterClient(url, apiKey, model string, timeoutMs int) *OpenRouterClient {
	return &OpenRouterClient{
		url:    url,
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		},
	}
}

type orReasoning struct {
	Effort string `json:"effort"`
}

type orRequest struct {
	Model       string       `json:"model"`
	Messages    []orMsg      `json:"messages"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Reasoning   *orReasoning `json:"reasoning,omitempty"`
	Tools       []ToolSpec   `json:"tools,omitempty"`
}

type orMsg struct {
	Role       string      `json:"role"`
	Content    *string     `json:"content"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	Name       string      `json:"name,omitempty"`
}

type orResponseMessage struct {
	Content          string     `json:"content"`
	ReasoningContent string     `json:"reasoning_content"`
	ToolCalls        []ToolCall `json:"tool_calls"`
}

type orResponse struct {
	Choices []struct {
		Message      orResponseMessage `json:"message"`
		FinishReason string            `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int     `json:"prompt_tokens"`
		CompletionTokens int     `json:"completion_tokens"`
		Cost             float64 `json:"cost"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func strPtr(s string) *string { return &s }

func (c *OpenRouterClient) Complete(ctx context.Context, systemPrompt, prompt string, model string, maxTokens int, temperature float64) (*Response, error) {
	m := c.model
	if model != "" {
		m = model
	}
	msgs := []orMsg{}
	if systemPrompt != "" {
		msgs = append(msgs, orMsg{Role: "system", Content: strPtr(systemPrompt)})
	}
	msgs = append(msgs, orMsg{Role: "user", Content: strPtr(prompt)})
	body := orRequest{
		Model:       m,
		Messages:    msgs,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Reasoning:   &orReasoning{Effort: "none"},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openrouter marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("openrouter request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter status %d: %s", resp.StatusCode, string(b))
	}

	var result orResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("openrouter decode: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("openrouter API error (code %d): %s", result.Error.Code, result.Error.Message)
	}

	content := ""
	if len(result.Choices) > 0 {
		content = result.Choices[0].Message.Content
		if content == "" && result.Choices[0].Message.ReasoningContent != "" {
			content = result.Choices[0].Message.ReasoningContent
		}
	}
	return &Response{
		Content:          content,
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
		Cost:             result.Usage.Cost,
		Provider:         "openrouter",
	}, nil
}

func (c *OpenRouterClient) CompleteWithTools(ctx context.Context, messages []Message, tools []ToolSpec, model string, maxTokens int, temperature float64) (*ToolResponse, error) {
	m := c.model
	if model != "" {
		m = model
	}

	var msgs []orMsg
	for _, msg := range messages {
		om := orMsg{
			Role:       msg.Role,
			ToolCalls:  msg.ToolCalls,
			ToolCallID: msg.ToolCallID,
			Name:       msg.Name,
		}
		// For assistant messages with tool_calls and no content, send content as null
		// (nil pointer) instead of empty string. Some providers (e.g., Qwen on Alibaba)
		// reject empty string content alongside tool_calls.
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 && msg.Content == "" {
			om.Content = nil
		} else if msg.Role == "tool" && msg.Content == "" {
			// Ensure tool result content is never empty — some providers reject empty tool results.
			om.Content = strPtr("(no output)")
		} else {
			om.Content = strPtr(msg.Content)
		}
		msgs = append(msgs, om)
	}

	body := orRequest{
		Model:       m,
		Messages:    msgs,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Tools:       tools,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openrouter marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("openrouter request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter status %d: %s", resp.StatusCode, string(b))
	}

	var result orResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("openrouter decode: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("openrouter API error (code %d): %s", result.Error.Code, result.Error.Message)
	}

	tr := &ToolResponse{
		Response: Response{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			Cost:             result.Usage.Cost,
			Provider:         "openrouter",
		},
	}

	if len(result.Choices) > 0 {
		msg := result.Choices[0].Message
		tr.Content = msg.Content
		if tr.Content == "" && msg.ReasoningContent != "" {
			tr.Content = msg.ReasoningContent
		}
		tr.ToolCalls = msg.ToolCalls
	}
	return tr, nil
}
