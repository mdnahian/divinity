package engine

import (
	"context"
	"log"

	"github.com/divinity/core/gametools"
	"github.com/divinity/core/god"
)

func (e *Engine) runGodTurn(ctx context.Context, focusTerritoryID string) {
	ticksSinceSpawn := e.tickNum - e.lastEnemySpawnTick
	log.Printf("[Engine] Triggering GOD turn at tick %d (Day %d %02d:%02d)",
		e.tickNum, e.World.GameDay, e.World.GameHour, e.World.GameMinute)
	// Convert TerritoryBriefs to the interface map expected by god package
	briefMap := make(map[string]gametools.TerritoryBriefFormatter)
	for k, v := range e.TerritoryBriefs {
		briefMap[k] = v
	}
	// Convert NPCTiers from engine.NPCTier to int for the gametools interface
	tierMap := make(map[string]int, len(e.NPCTiers))
	for id, t := range e.NPCTiers {
		tierMap[id] = int(t)
	}
	results, didSpawn := god.RunGodAgentTurn(ctx, e.World, e.Router, e.Memory, e.Config, ticksSinceSpawn, e.Trends, focusTerritoryID, e.Relationships, e.SharedMemory, briefMap, e.SocialGraph, tierMap)
	if didSpawn {
		e.lastEnemySpawnTick = e.tickNum
		log.Printf("[Engine] GOD spawned enemy — resetting spawn timer")
	}
	log.Printf("[Engine] GOD turn complete — %d result(s)", len(results))
	for _, r := range results {
		log.Printf("[GOD] Result: %s", r.Text)
		if r.Type == "god" && !r.Logged {
			e.World.LogEvent(r.Text, "god")
		}
	}
}
