package gametools

import (
	"encoding/json"

	"github.com/divinity/core/config"
	"github.com/divinity/core/graph"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

// TrendFormatter provides world trend data for GOD strategic analysis.
type TrendFormatter interface {
	FormatForGod() string
}

// TerritoryBriefFormatter provides a pre-computed intelligence brief for a territory.
type TerritoryBriefFormatter interface {
	FormatForGod() string
}

type AgentContext struct {
	NPC              *npc.NPC
	World            *world.World
	Memory           memory.Store
	Relationships    *memory.RelationshipStore
	SharedMemory     *memory.SharedMemoryStore
	Trends           TrendFormatter
	TerritoryBriefs  map[string]TerritoryBriefFormatter
	SocialGraph      *graph.SocialGraph
	NPCTiers         map[string]int // NPC ID → tier (1=key figure, 2=notable, 3=commoner)
	Config           *config.Config
}

type ToolDef struct {
	Name        string                                               `json:"name"`
	Description string                                               `json:"description"`
	Parameters  json.RawMessage                                      `json:"parameters"`
	IsTerminal  bool                                                 `json:"-"`
	Handler     func(ctx *AgentContext, args json.RawMessage) (string, error) `json:"-"`
}

func (t *ToolDef) ToOpenAI() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name,
			"description": t.Description,
			"parameters":  json.RawMessage(t.Parameters),
		},
	}
}

func ToolDefsToOpenAI(tools []*ToolDef) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tools))
	for i, t := range tools {
		result[i] = t.ToOpenAI()
	}
	return result
}
