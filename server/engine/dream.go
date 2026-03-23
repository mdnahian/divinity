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

const dreamSystemPrompt = `You are simulating the subconscious dream of a medieval villager. Respond with ONLY one vivid sentence describing the dream. No JSON, no explanation, no quotes — just the dream itself in plain text. ALL text must be in English.`

func GenerateDream(ctx context.Context, n *npc.NPC, w *world.World, router *llm.Router, mem memory.Store, cfg *config.Config) {
	if cfg.API.OpenRouterKey == "" {
		return
	}
	prompt := BuildDreamPrompt(n, w, mem)
	resp, err := router.Call(ctx, dreamSystemPrompt, prompt, "", 150, cfg.API.Temperature)
	if err != nil {
		log.Printf("[Dream] %s: LLM call failed: %v", n.Name, err)
		return
	}
	dream := strings.TrimSpace(resp.Content)
	dream = strings.Trim(dream, "\"'")
	if dream == "" {
		return
	}

	// Weight dream importance toward the NPC's highest-importance recent memories
	importance := 0.3
	recentMems := mem.HighestImportance(n.ID, 1)
	if len(recentMems) > 0 && recentMems[0].Importance > 0.8 {
		importance = 0.5 // recurring dream potential from traumatic memory
	}

	mem.Add(n.ID, memory.Entry{
		Text:       memory.DreamPrefix + dream,
		Time:       w.TimeString(),
		TS:         int64(w.GameDay),
		Importance: importance,
		Category:   memory.CatDream,
		Tags:       []string{n.ID},
	})
	log.Printf("[Dream] %s dreamed: %s", n.Name, dream)
}

func BuildDreamPrompt(n *npc.NPC, w *world.World, mem memory.Store) string {
	// Use highest-importance recent memories to seed dreams, not just last 3
	allRecent := mem.Recent(n.ID, 10)
	var candidates []memory.Entry
	for _, e := range allRecent {
		if !strings.HasPrefix(e.Text, memory.DreamPrefix) {
			candidates = append(candidates, e)
		}
	}
	// Sort by importance descending, take top 3
	if len(candidates) > 1 {
		for i := 0; i < len(candidates)-1; i++ {
			for j := i + 1; j < len(candidates); j++ {
				if candidates[j].Importance > candidates[i].Importance {
					candidates[i], candidates[j] = candidates[j], candidates[i]
				}
			}
		}
	}
	if len(candidates) > 3 {
		candidates = candidates[:3]
	}

	var recentMems []string
	for _, e := range candidates {
		recentMems = append(recentMems, "- "+e.Text)
	}
	recentMemStr := "Nothing memorable lately."
	if len(recentMems) > 0 {
		recentMemStr = strings.Join(recentMems, "\n")
	}

	return fmt.Sprintf(`%s is a %s living in a small medieval village. They are %s years old and their mood is %s.
Recent experiences:
%s

As %s drifts into sleep, their mind wanders. Describe their dream in one vivid sentence. It can be anything — fears, desires, memories of someone, strange visions, a hopeful fantasy, or pure surreal imagery. Be creative and personal.`,
		n.Name, n.Profession, fmt.Sprint(n.GetAge(w.GameDay, 15)), n.Mood(),
		recentMemStr,
		n.Name)
}

func ComputeDreamChance(n *npc.NPC) float64 {
	bonus := float64(max(0, n.Stats.MysticalAptitude-50)+max(0, n.Stats.Creativity-50)) / 300.0
	chance := 0.20 + bonus
	if chance < 0.15 {
		return 0.15
	}
	if chance > 0.35 {
		return 0.35
	}
	return chance
}

func HasDreamToday(mem memory.Store, npcID string, gameDay int64) bool {
	for _, e := range mem.Recent(npcID, memory.MaxMemories) {
		if e.TS == gameDay && (strings.HasPrefix(e.Text, memory.DreamPrefix) || strings.HasPrefix(e.Text, memory.DivineDreamPrefix)) {
			return true
		}
	}
	return false
}
