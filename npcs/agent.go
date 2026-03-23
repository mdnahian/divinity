package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

const maxToolIterations = 8

type Agent struct {
	client *CoreClient
	llm    *LLMClient
	name   string
	prof   string
	pers   string
}

func NewAgent(client *CoreClient, llm *LLMClient, name, profession, personality string) *Agent {
	return &Agent{
		client: client,
		llm:    llm,
		name:   name,
		prof:   profession,
		pers:   personality,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		state, err := a.client.GetState()
		if err != nil {
			log.Printf("[Agent] Failed to get state: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if !state.Alive {
			log.Printf("[Agent] %s is dead. Shutting down.", a.name)
			return fmt.Errorf("NPC %s is dead", a.name)
		}

		if state.Busy {
			waitTicks := state.BusyUntil - state.CurrentTick
			if waitTicks < 1 {
				waitTicks = 1
			}
			sleepDur := time.Duration(waitTicks) * 30 * time.Second
			if sleepDur > 2*time.Minute {
				sleepDur = 2 * time.Minute
			}
			log.Printf("[Agent] %s is busy for ~%d ticks, sleeping %v", a.name, waitTicks, sleepDur)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sleepDur):
			}
			continue
		}

		if err := a.decide(ctx, state); err != nil {
			log.Printf("[Agent] %s decision error: %v", a.name, err)
			time.Sleep(10 * time.Second)
		}

		// Random interval 0–300s to stagger 500+ agents and avoid LLM stampedes
		time.Sleep(time.Duration(rand.Intn(301)) * time.Second)
	}
}

func (a *Agent) decide(ctx context.Context, state *StateResponse) error {
	toolsResp, err := a.client.GetTools()
	if err != nil {
		return fmt.Errorf("get tools: %w", err)
	}

	if toolsResp.Busy {
		log.Printf("[Agent] %s became busy between state check and tools fetch", a.name)
		return nil
	}

	toolSpecs := convertToolSpecs(toolsResp.Tools)
	toolNames := make(map[string]bool)
	for _, t := range toolSpecs {
		toolNames[t.Function.Name] = true
	}

	// Fetch canonical prompt from the core engine
	promptResp, err := a.client.GetPrompt()
	if err != nil {
		return fmt.Errorf("get prompt: %w", err)
	}

	messages := []ChatMessage{
		{Role: "system", Content: promptResp.SystemPrompt},
		{Role: "user", Content: promptResp.UserPrompt},
	}

	toolCache := make(map[string]string) // "toolName|args" -> cached result

	for i := 0; i < maxToolIterations; i++ {
		if i == 5 {
			messages = append(messages, ChatMessage{
				Role:    "user",
				Content: "You have used most of your tool calls. Call commit_action NOW with one of the actions from list_actions.",
			})
		}
		llmResp, err := a.llm.ChatComplete(ctx, messages, toolSpecs)
		if err != nil {
			return fmt.Errorf("LLM call (iteration %d): %w", i, err)
		}

		if len(llmResp.ToolCalls) == 0 {
			if i == 0 {
				messages = append(messages, ChatMessage{Role: "assistant", Content: llmResp.Content})
				messages = append(messages, ChatMessage{Role: "user", Content: "You must use your tools. Call list_actions to see what you can do, then commit_action to act."})
				continue
			}
			log.Printf("[Agent] %s: LLM gave no tool calls after %d iterations, giving up", a.name, i)
			return nil
		}

		assistantMsg := ChatMessage{
			Role:      "assistant",
			Content:   llmResp.Content,
			ToolCalls: llmResp.ToolCalls,
		}
		messages = append(messages, assistantMsg)

		committed := false
		for _, tc := range llmResp.ToolCalls {
			toolName := tc.Function.Name

			if toolName == "commit_action" {
				result, err := a.handleCommit(tc.Function.Arguments)
				if err != nil {
					messages = append(messages, ChatMessage{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    fmt.Sprintf("Error: %v", err),
					})
					continue
				}
				messages = append(messages, ChatMessage{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    result,
				})
				committed = true
				continue
			}

			if !toolNames[toolName] {
				messages = append(messages, ChatMessage{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("Unknown tool %q. Use list_actions to see available tools.", toolName),
				})
				continue
			}

			cacheKey := toolName + "|" + tc.Function.Arguments
			if cached, ok := toolCache[cacheKey]; ok {
				messages = append(messages, ChatMessage{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    cached + "\n\n[Same result as before. Please call commit_action now.]",
				})
				continue
			}

			result, err := a.client.CallTool(toolName, json.RawMessage(tc.Function.Arguments))
			if err != nil {
				messages = append(messages, ChatMessage{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("Error: %v", err),
				})
				continue
			}

			toolCache[cacheKey] = result.Result
			messages = append(messages, ChatMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result.Result,
			})
		}

		if committed {
			log.Printf("[Agent] %s committed an action", a.name)
			return nil
		}
	}

	log.Printf("[Agent] %s: exhausted %d iterations without committing", a.name, maxToolIterations)
	return nil
}

func (a *Agent) handleCommit(argsJSON string) (string, error) {
	var args struct {
		ActionID string `json:"action_id"`
		Target   string `json:"target"`
		Dialogue string `json:"dialogue"`
		Goal     string `json:"goal"`
		Location string `json:"location"`
		Reason   string `json:"reason"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse commit args: %w", err)
	}

	resp, err := a.client.CommitAction(args.ActionID, args.Target, args.Dialogue, args.Goal, args.Location, args.Reason)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Action %q committed successfully (NPC: %s)", resp.ActionID, resp.NPCID), nil
}

func convertToolSpecs(raw []map[string]interface{}) []ToolSpec {
	var specs []ToolSpec
	for _, t := range raw {
		fnRaw, ok := t["function"]
		if !ok {
			continue
		}
		fnMap, ok := fnRaw.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := fnMap["name"].(string)
		desc, _ := fnMap["description"].(string)
		params := fnMap["parameters"]

		// Filter out tool type prefix if present
		toolType, _ := t["type"].(string)
		if toolType == "" {
			toolType = "function"
		}

		specs = append(specs, ToolSpec{
			Type: toolType,
			Function: ToolFunction{
				Name:        name,
				Description: desc,
				Parameters:  params,
			},
		})
	}
	return specs
}
