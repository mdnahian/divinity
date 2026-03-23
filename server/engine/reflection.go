package engine

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/divinity/core/config"
	"github.com/divinity/core/llm"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

const reflectionSystemPrompt = `You are simulating the inner thoughts of a medieval villager reflecting on their recent experiences. Based on their memories, generate 1-2 short insight statements (lessons learned, beliefs formed, or conclusions drawn). Each insight should be a single sentence. Respond with ONLY the insights, one per line. No JSON, no explanation, no quotes. ALL text must be in English.`

// GenerateReflection produces insight memories for an NPC based on their recent experiences.
// Called periodically (e.g., every few game days) for NPCs with sufficient memories.
func GenerateReflection(ctx context.Context, n *npc.NPC, w *world.World, router *llm.Router, mem memory.Store, cfg *config.Config) {
	if cfg.API.OpenRouterKey == "" {
		return
	}

	// Need at least 10 memories to reflect
	all := mem.All(n.ID)
	if len(all) < 10 {
		return
	}

	// Don't reflect if already reflected recently (check for reflection in last 3 entries)
	recent := mem.Recent(n.ID, 3)
	for _, e := range recent {
		if e.Category == memory.CatReflection {
			return
		}
	}

	prompt := buildReflectionPrompt(n, w, mem)
	resp, err := router.Call(ctx, reflectionSystemPrompt, prompt, "", 200, cfg.API.Temperature)
	if err != nil {
		log.Printf("[Reflection] %s: LLM call failed: %v", n.Name, err)
		return
	}

	text := strings.TrimSpace(resp.Content)
	if text == "" {
		return
	}

	// Split into individual insights
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "- •")
		line = strings.TrimSpace(line)
		line = strings.Trim(line, "\"'")
		if line == "" {
			continue
		}
		mem.Add(n.ID, memory.Entry{
			Text:       "INSIGHT: " + line,
			Time:       w.TimeString(),
			TS:         int64(w.GameDay),
			Importance: 0.8,
			Category:   memory.CatReflection,
			Tags:       []string{n.ID},
		})
		log.Printf("[Reflection] %s reflected: %s", n.Name, line)
	}
}

func buildReflectionPrompt(n *npc.NPC, w *world.World, mem memory.Store) string {
	recent := mem.Recent(n.ID, 8)
	important := mem.HighestImportance(n.ID, 3)

	var memLines []string
	seen := make(map[string]bool)
	for _, e := range recent {
		if e.Category != memory.CatReflection && e.Category != memory.CatDream {
			memLines = append(memLines, fmt.Sprintf("- [%s] %s", e.Time, e.Text))
			seen[e.Text] = true
		}
	}
	for _, e := range important {
		if !seen[e.Text] && e.Category != memory.CatReflection && e.Category != memory.CatDream {
			memLines = append(memLines, fmt.Sprintf("- [%s] (important) %s", e.Time, e.Text))
		}
	}

	memStr := "No significant experiences."
	if len(memLines) > 0 {
		memStr = strings.Join(memLines, "\n")
	}

	return fmt.Sprintf(`%s is a %s, age %d, living in a medieval village. Their mood is %s.

Recent experiences:
%s

Based on these experiences, what has %s learned? Generate 1-2 brief insight statements — beliefs, lessons, or conclusions this person would naturally draw from their experiences.`,
		n.Name, n.Profession, n.GetAge(w.GameDay, 15), n.Mood(),
		memStr,
		n.Name)
}
