package god

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/divinity/core/config"
	"github.com/divinity/core/llm"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

func SpawnOnDemand(ctx context.Context, w *world.World, router *llm.Router, cfg *config.Config) (*world.SpawnEntry, error) {
	w.Mu.RLock()
	profCounts := make(map[string]int)
	for _, n := range w.AliveNPCs() {
		profCounts[n.Profession]++
	}
	population := len(w.AliveNPCs())
	var profParts []string
	for p, c := range profCounts {
		profParts = append(profParts, fmt.Sprintf("%s: %d", p, c))
	}
	profSummary := strings.Join(profParts, ", ")
	if profSummary == "" {
		profSummary = "none"
	}
	timeStr := w.TimeString()
	w.Mu.RUnlock()

	prompt := fmt.Sprintf(`A new agent wants to join the village. The village currently has %d living residents.
Profession breakdown: %s
Current time: %s

Based on the village's needs, decide what new villager should arrive. Choose a name, profession, age (18-50), and a one-sentence personality.
Available professions: farmer, hunter, merchant, herbalist, barmaid, blacksmith, carpenter, scribe, tailor, miner, fisher, potter, baker, guard, priest.
Pick a profession the village needs most (fewest of, or completely missing).

Respond with ONLY a JSON object:
{"name": "<name>", "profession": "<profession>", "age": <number>, "personality": "<one sentence>"}`, population, profSummary, timeStr)

	resp, err := router.CallGod(ctx, GodSystemPrompt, prompt, cfg.GodAgent.Model, cfg.GodAgent.MaxTokens, cfg.GodAgent.Temperature)
	if err != nil {
		return nil, fmt.Errorf("GOD LLM call failed: %w", err)
	}

	var parsed struct {
		Name        string `json:"name"`
		Profession  string `json:"profession"`
		Age         int    `json:"age"`
		Personality string `json:"personality"`
	}

	content := strings.TrimSpace(resp.Content)
	if idx := strings.Index(content, "{"); idx >= 0 {
		content = content[idx:]
	}
	if idx := strings.LastIndex(content, "}"); idx >= 0 {
		content = content[:idx+1]
	}

	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("GOD response parse error: %w (content: %s)", err, resp.Content)
	}

	if parsed.Name == "" {
		parsed.Name = "Stranger"
	}
	if parsed.Profession == "" {
		parsed.Profession = "farmer"
	}
	if parsed.Age < 18 || parsed.Age > 60 {
		parsed.Age = 25
	}

	w.Mu.Lock()
	defer w.Mu.Unlock()

	// Pick a random inn as home, fall back to any location
	var homeID string
	inns := w.LocationsByType("inn")
	if len(inns) > 0 {
		homeID = inns[rand.Intn(len(inns))].ID
	} else if len(w.Locations) > 0 {
		homeID = w.Locations[rand.Intn(len(w.Locations))].ID
	}

	// Spawn at a random public location (market, inn, tavern, square) for variety
	startLocID := homeID
	publicTypes := []string{"market", "inn", "tavern", "square", "plaza"}
	var publicLocs []*world.Location
	for _, t := range publicTypes {
		publicLocs = append(publicLocs, w.LocationsByType(t)...)
	}
	if len(publicLocs) > 0 {
		startLocID = publicLocs[rand.Intn(len(publicLocs))].ID
	} else if len(w.Locations) > 0 {
		startLocID = w.Locations[rand.Intn(len(w.Locations))].ID
	}

	tmpl := npc.Template{
		Name:        parsed.Name,
		Profession:  parsed.Profession,
		Age:         parsed.Age,
		Personality: parsed.Personality,
		HomeID:      homeID,
		StartItems:  []npc.InventoryItem{{Name: "bread", Qty: 2}, {Name: "gold", Qty: 15}},
		Skills:      map[string]float64{parsed.Profession: float64(30 + rand.Intn(20))},
	}
	newNPC := npc.NewNPC(tmpl, len(w.NPCs), w.GameDay, cfg)
	newNPC.LocationID = startLocID
	w.NPCs = append(w.NPCs, newNPC)

	log.Printf("[GOD] On-demand spawn: %s the %s, age %d", parsed.Name, parsed.Profession, parsed.Age)
	w.LogEvent(fmt.Sprintf("A new villager arrives: %s the %s. \"%s\"", parsed.Name, parsed.Profession, parsed.Personality), "god")

	return &world.SpawnEntry{
		NPCID:      newNPC.ID,
		Name:       newNPC.Name,
		Profession: newNPC.Profession,
	}, nil
}
