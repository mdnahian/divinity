package god

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/divinity/core/config"
	"github.com/divinity/core/gametools"
	"github.com/divinity/core/graph"
	"github.com/divinity/core/llm"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/world"
)

const GodSystemPrompt = `You are a distant, inscrutable deity overseeing a medieval realm. You CANNOT directly affect mortals — no healing, no cursing, no gifting items, no changing stats. Your influence is STRICTLY INDIRECT: dreams, visions, omens, weather, and shaping the natural landscape. NPCs make their own choices; you can only inspire, warn, and guide them through the world of dreams and signs. Use your tools to understand the current state, then commit your divine actions. ALL text must be in English.`

func RunGodAgentTurn(ctx context.Context, w *world.World, router *llm.Router, mem memory.Store, cfg *config.Config, ticksSinceLastSpawn int, trends gametools.TrendFormatter, focusTerritoryID string, relationships *memory.RelationshipStore, sharedMemory *memory.SharedMemoryStore, territoryBriefs map[string]gametools.TerritoryBriefFormatter, socialGraph *graph.SocialGraph, npcTiers map[string]int) (results []*GodResult, didSpawnEnemy bool) {
	if cfg.API.OpenRouterKey == "" {
		return nil, false
	}
	if !cfg.GodAgent.Enabled {
		return nil, false
	}

	tools := gametools.GodTools()
	agentCtx := &gametools.AgentContext{
		World:           w,
		Memory:          mem,
		Relationships:   relationships,
		SharedMemory:    sharedMemory,
		Config:          cfg,
		Trends:          trends,
		TerritoryBriefs: territoryBriefs,
		SocialGraph:     socialGraph,
		NPCTiers:        npcTiers,
	}

	var toolSpecs []llm.ToolSpec
	toolMap := make(map[string]*gametools.ToolDef)
	for _, t := range tools {
		toolSpecs = append(toolSpecs, llm.ToolSpec{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
		toolMap[t.Name] = t
	}

	identityPrompt := fmt.Sprintf(
		"You are a distant, inscrutable deity. You CANNOT directly manipulate NPCs or their stats.\n"+
			"Your only tools of influence are: dreams (send_dream), visions (send_vision for praying NPCs), omens (send_omen for all sensitive NPCs), weather control, spawning creatures, and shaping the natural landscape.\n"+
			"Ticks since last enemy spawn: %d.\n\n"+
			"STRATEGY: Dreams are your primary tool — they become NPC memories that shape their decisions.\n"+
			"Examples: wealth hoarding → dream inspiring a rival to form a faction; food shortage → bring rain + dream a farmer about new techniques; danger → spawn enemies + dream a warning to a brave warrior.\n\n"+
			"WEATHER: Rain fills wells with water. Manage weather strategically — clear skies are pleasant but wells run dry.\n\n"+
			"Use your tools to check the village state, then use god_act to issue your divine commands.\n"+
			"Priorities: population survival > narrative depth > environmental balance > interesting challenges.\n\n"+
			"STRATEGIC TOOLS: Use 'god_memories' to recall past decisions (filter by category: spawn, dream, vision, omen, resource, growth, weather). "+
			"Use 'world_trends' to see how key metrics have changed over recent days — this helps you evaluate whether your interventions are working.",
		ticksSinceLastSpawn,
	)

	if focusTerritoryID != "" {
		t := w.TerritoryByID(focusTerritoryID)
		if t != nil {
			identityPrompt += fmt.Sprintf("\n\nCURRENT FOCUS: %s (%s)\nPrioritize this territory this turn. You may still act elsewhere if urgent.", t.Name, t.ID)
			if brief, ok := territoryBriefs[focusTerritoryID]; ok && brief != nil {
				identityPrompt += "\n\n" + brief.FormatForGod()
			}
		}
	}

	messages := []llm.Message{
		{Role: "system", Content: GodSystemPrompt},
		{Role: "user", Content: identityPrompt},
	}

	maxCalls := cfg.Game.GodMaxToolCalls
	if maxCalls <= 0 {
		maxCalls = 8
	}
	deadline := time.Now().Add(2 * time.Minute)

	log.Printf("[GOD] === GOD turn start === (model=%s, maxCalls=%d, ticksSinceSpawn=%d)", cfg.GodAgent.Model, maxCalls, ticksSinceLastSpawn)
	turnStart := time.Now()

	for i := 0; i <= maxCalls; i++ {
		if time.Now().After(deadline) {
			log.Printf("[GOD] Timeout after %d iterations (%.1fs elapsed), aborting", i, time.Since(turnStart).Seconds())
			return []*GodResult{{Text: "The GOD stirs but cannot act (timeout).", Type: "god"}}, false
		}

		log.Printf("[GOD] LLM call iteration %d/%d (messages=%d)", i, maxCalls, len(messages))
		resp, err := router.CallGodWithTools(ctx, messages, toolSpecs, cfg.GodAgent.Model, cfg.GodAgent.MaxTokens, cfg.GodAgent.Temperature)
		if err != nil {
			log.Printf("[GOD] LLM call FAILED (iteration %d): %v", i, err)
			return []*GodResult{{Text: fmt.Sprintf("The GOD stirs but cannot act (%v).", err), Type: "god"}}, false
		}

		if len(resp.ToolCalls) == 0 {
			log.Printf("[GOD] No tool calls returned (iteration %d), content: %s", i, truncate(resp.Content, 300))
			if i == maxCalls {
				log.Printf("[GOD] === GOD turn end === (no actions, %.1fs elapsed)", time.Since(turnStart).Seconds())
				return []*GodResult{{Text: "The GOD watches silently.", Type: "god"}}, false
			}
			messages = append(messages, llm.Message{Role: "assistant", Content: resp.Content})
			messages = append(messages, llm.Message{Role: "user", Content: "Use your tools to check the village state, then use god_act to issue commands."})
			continue
		}

		log.Printf("[GOD] Received %d tool call(s) in iteration %d", len(resp.ToolCalls), i)
		if resp.Content != "" {
			log.Printf("[GOD] Assistant content: %s", truncate(resp.Content, 300))
		}

		assistantMsg := llm.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		messages = append(messages, assistantMsg)

		for _, tc := range resp.ToolCalls {
			tool, ok := toolMap[tc.Function.Name]
			if !ok {
				log.Printf("[GOD] Unknown tool called: %q", tc.Function.Name)
				messages = append(messages, llm.Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("Unknown tool %q.", tc.Function.Name),
				})
				continue
			}

			log.Printf("[GOD] Tool call: %s | args: %s", tc.Function.Name, truncate(tc.Function.Arguments, 500))
			callStart := time.Now()
			result, _ := tool.Handler(agentCtx, json.RawMessage(tc.Function.Arguments))
			log.Printf("[GOD] Tool result (%s, %dms): %s", tc.Function.Name, time.Since(callStart).Milliseconds(), truncate(result, 500))

			if tool.IsTerminal && result == "GOD_ACTION_COMMITTED" {
				log.Printf("[GOD] Terminal tool god_act invoked — validating actions")
				w.Mu.RLock()
				validationErrs, ok := ValidateGodActions(tc.Function.Arguments, w)
				w.Mu.RUnlock()
				if !ok {
					log.Printf("[GOD] Validation FAILED: %s", strings.Join(validationErrs, "; "))
					msg := "Validation failed. Fix the following and call god_act again with corrected params:\n" + strings.Join(validationErrs, "\n")
					messages = append(messages, llm.Message{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    msg,
					})
					continue
				}
				log.Printf("[GOD] Validation passed — executing god_act actions (%.1fs elapsed)", time.Since(turnStart).Seconds())
				return parseAndExecuteGodActions(tc.Function.Arguments, w, mem, cfg)
			}

			messages = append(messages, llm.Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result,
			})
		}
	}

	log.Printf("[GOD] === GOD turn end === (max iterations reached, %.1fs elapsed)", time.Since(turnStart).Seconds())
	return []*GodResult{{Text: "The GOD watches silently.", Type: "god"}}, false
}

func parseAndExecuteGodActions(argsJSON string, w *world.World, mem memory.Store, cfg *config.Config) ([]*GodResult, bool) {
	var params struct {
		Analysis string `json:"analysis"`
		Actions  []struct {
			Action string                 `json:"action"`
			Reason string                 `json:"reason"`
			Params map[string]interface{} `json:"params"`
		} `json:"actions"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		log.Printf("[GOD] Failed to parse god_act args: %v", err)
		return []*GodResult{{Text: "The GOD's will was unclear (parse error).", Type: "god"}}, false
	}

	log.Printf("[GOD] Parsing god_act: %d action(s) planned", len(params.Actions))
	if params.Analysis != "" {
		log.Printf("[GOD] Analysis: %s", params.Analysis)
	}

	var results []*GodResult
	didSpawnEnemy := false
	seen := make(map[string]bool)

	var actionSummaries []string
	var maxImportance float64
	var allTags []string

	for idx, a := range params.Actions {
		if a.Action == "" || seen[a.Action] {
			if a.Action != "" {
				log.Printf("[GOD] Action %d/%d: %s (skipped — duplicate)", idx+1, len(params.Actions), a.Action)
			}
			continue
		}
		if strings.TrimSpace(a.Reason) == "" {
			log.Printf("[GOD] Action %d/%d: %s (skipped — no reason)", idx+1, len(params.Actions), a.Action)
			continue
		}
		seen[a.Action] = true

		paramsJSON, _ := json.Marshal(a.Params)
		log.Printf("[GOD] Action %d/%d: %s | reason: %s | params: %s", idx+1, len(params.Actions), a.Action, truncate(a.Reason, 200), truncate(string(paramsJSON), 300))

		details := make(map[string]interface{})
		details["action"] = a.Action
		details["reason"] = a.Reason
		if a.Params != nil {
			for k, v := range a.Params {
				details[k] = v
			}
		}

		r := ExecuteGodAction(details, w, mem, cfg)
		log.Printf("[GOD] Action %d result: %s", idx+1, truncate(r.Text, 300))
		w.LogEvent(r.Text, "god")
		r.Logged = true
		results = append(results, &r)

		if a.Action == "spawn_enemy" {
			didSpawnEnemy = true
		}

		if a.Reason != "" {
			actionSummaries = append(actionSummaries, fmt.Sprintf("%s: %s", a.Action, a.Reason))
		} else {
			actionSummaries = append(actionSummaries, a.Action)
		}

		// Track highest importance among actions for the composite memory
		imp := godActionImportance(a.Action)
		if imp > maxImportance {
			maxImportance = imp
		}

		// Collect tags from action params (target names, locations, etc.)
		if a.Params != nil {
			for _, key := range []string{"target_name", "name", "location_id", "enemy_type", "event_type"} {
				if v, ok := a.Params[key]; ok {
					if s, ok := v.(string); ok && s != "" {
						allTags = append(allTags, s)
					}
				}
			}
		}

		// Store individual action memories with subcategories for filtered retrieval
		subcat := godActionSubcategory(a.Action)
		var actionTags []string
		if a.Params != nil {
			for _, key := range []string{"target_name", "name", "location_id", "enemy_type"} {
				if v, ok := a.Params[key]; ok {
					if s, ok := v.(string); ok && s != "" {
						actionTags = append(actionTags, s)
					}
				}
			}
		}
		actionText := a.Action
		if a.Reason != "" {
			actionText = fmt.Sprintf("%s: %s", a.Action, a.Reason)
		}
		mem.Add(memory.GodEntityID, memory.Entry{
			Text:       actionText,
			Time:       w.TimeString(),
			TS:         int64(w.GameDay),
			Importance: imp,
			Category:   subcat,
			Tags:       actionTags,
		})
	}

	if maxImportance == 0 {
		maxImportance = 0.5
	}

	// Store composite summary memory for the overall turn
	memText := params.Analysis
	if len(actionSummaries) > 0 {
		actStr := strings.Join(actionSummaries, "; ")
		if memText != "" {
			memText = memText + " Actions: " + actStr
		} else {
			memText = "Actions: " + actStr
		}
	}
	if memText != "" {
		mem.Add(memory.GodEntityID, memory.Entry{
			Text:       memText,
			Time:       w.TimeString(),
			TS:         int64(w.GameDay),
			Importance: maxImportance,
			Category:   memory.CatGod,
			Tags:       allTags,
		})
	}

	if len(results) == 0 {
		return []*GodResult{{Text: "The GOD chose to do nothing.", Type: "god"}}, false
	}

	return results, didSpawnEnemy
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// godActionSubcategory maps a GOD action to a memory subcategory for filtered retrieval.
func godActionSubcategory(action string) string {
	switch action {
	case "spawn_npc", "spawn_enemy":
		return memory.CatGodSpawn
	case "send_dream", "stir_dreams":
		return memory.CatGodDream
	case "send_vision":
		return memory.CatGodVision
	case "send_omen", "send_prophecy":
		return memory.CatGodOmen
	case "replenish_resources":
		return memory.CatGodResource
	case "create_location", "expand_grid", "introduce_item", "introduce_technique", "introduce_profession":
		return memory.CatGodGrowth
	case "advance_weather":
		return memory.CatGodWeather
	default:
		return memory.CatGod
	}
}

// godActionImportance returns an importance score for GOD memories based on action type.
// High-impact actions (miracles, spawns, curses) are remembered more vividly.
func godActionImportance(action string) float64 {
	switch action {
	case "send_vision":
		return 0.9
	case "send_dream", "stir_dreams":
		return 0.7
	case "spawn_npc", "spawn_enemy":
		return 0.7
	case "send_prophecy":
		return 0.8
	case "send_omen":
		return 0.6
	case "create_location", "expand_grid":
		return 0.6
	case "introduce_item", "introduce_technique", "introduce_profession":
		return 0.5
	case "replenish_resources", "advance_weather":
		return 0.3
	case "do_nothing":
		return 0.1
	default:
		return 0.5
	}
}
