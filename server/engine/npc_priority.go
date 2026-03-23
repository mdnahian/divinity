package engine

import (
	"sort"

	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

// NPCTier represents the AI processing tier for an NPC.
type NPCTier int

const (
	Tier1FullLLM  NPCTier = 1 // Full LLM agent — nobles, faction leaders, crisis NPCs
	Tier2BatchLLM NPCTier = 2 // Batched LLM — important NPCs not in focus
	Tier3RuleBased NPCTier = 3 // Deterministic rule-based AI
)

// ScoredNPC holds an NPC reference and its priority score.
type ScoredNPC struct {
	NPC   *npc.NPC
	Score float64
}

// ScoreNPCPriority calculates how important this NPC is for full AI processing.
func ScoreNPCPriority(n *npc.NPC, w *world.World) float64 {
	score := 0.0

	// Nobles always high priority
	switch n.NobleRank {
	case "king", "queen":
		score += 80
	case "duke":
		score += 60
	case "count", "prince", "princess":
		score += 50
	case "baron":
		score += 40
	case "knight":
		score += 30
	}

	// Faction leaders
	for _, f := range w.Factions {
		if f.LeaderID == n.ID {
			score += 30
			break
		}
	}

	// Business owners
	if n.IsBusinessOwner {
		score += 10
	}

	// Crisis state boosts priority
	if n.HP < 30 {
		score += 25
	}
	if n.Needs.Hunger < 15 {
		score += 15
	}
	if n.Needs.Thirst < 10 {
		score += 15
	}
	if n.Stress > 80 {
		score += 10
	}

	// Combat situation
	loc := w.LocationByID(n.LocationID)
	if loc != nil {
		enemies := w.EnemiesAtLocation(loc.ID)
		if len(enemies) > 0 {
			score += 20
		}
	}

	return score
}

// ClassifyNPCs assigns every alive, claimed NPC to a tier based on priority scoring.
// Returns three slices: tier1, tier2, tier3.
func ClassifyNPCs(w *world.World, tier1Count, tier2Count int) ([]*npc.NPC, []*npc.NPC, []*npc.NPC) {
	alive := w.AliveNPCs()
	claimed := make([]*npc.NPC, 0, len(alive))
	for _, n := range alive {
		if n.Claimed {
			claimed = append(claimed, n)
		}
	}

	// Score and sort
	scored := make([]ScoredNPC, len(claimed))
	for i, n := range claimed {
		scored[i] = ScoredNPC{NPC: n, Score: ScoreNPCPriority(n, w)}
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	var tier1, tier2, tier3 []*npc.NPC
	for i, s := range scored {
		if i < tier1Count {
			tier1 = append(tier1, s.NPC)
		} else if i < tier1Count+tier2Count {
			tier2 = append(tier2, s.NPC)
		} else {
			tier3 = append(tier3, s.NPC)
		}
	}

	return tier1, tier2, tier3
}
