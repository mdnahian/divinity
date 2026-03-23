package action

import (
	"fmt"
	"math"
	"strings"

	bld "github.com/divinity/core/building"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

// buildingDisplayName converts an internal building type like "market_stall"
// into a human-readable name like "Market Stall".
func buildingDisplayName(typeName string) string {
	parts := strings.Split(typeName, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

var constructionActions = []Action{
	{
		ID: "begin_construction", Label: "Begin building a structure (needs materials + carpentry)", Category: "construction", BaseGameMinutes: 120, SkillKey: "carpentry",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			for typeName := range bld.Types {
				if canConstruct(n, typeName) {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			for typeName, bt := range bld.Types {
				if !canConstruct(n, typeName) {
					continue
				}
				for matName, qty := range bt.Materials {
					n.RemoveItem(matName, qty)
				}
				c := &bld.Construction{
					ID:           fmt.Sprintf("construction_%d", len(w.Constructions)+1),
					BuildingType: typeName,
					Name:         typeName,
					OwnerID:      n.ID,
					Progress:     10,
					MaxProgress:  100,
					Durability:   100,
					LocationID:   n.LocationID,
				}
				w.Constructions = append(w.Constructions, c)
				n.Needs.Fatigue = clampF(n.Needs.Fatigue+15, 0, 100)
				return fmt.Sprintf("Began constructing a %s (progress: %.0f%%).", typeName, c.Progress)
			}
			return "Could not start any construction."
		},
	},
	{
		ID: "repair_building", Label: "Repair a building you own", Category: "construction", BaseGameMinutes: 45, SkillKey: "carpentry",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			for _, loc := range w.Locations {
				if loc.BuildingOwnerID == n.ID && loc.BuildingDurability < 80 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			for _, loc := range w.Locations {
				if loc.BuildingOwnerID == n.ID && loc.BuildingDurability < 80 {
					skill := math.Max(n.GetSkillLevel("carpentry"), float64(n.Stats.Carpentry))
					loc.BuildingDurability = math.Min(100, loc.BuildingDurability+10+skill*0.2)
					n.Needs.Fatigue = clampF(n.Needs.Fatigue+10, 0, 100)
					n.GainSkill("carpentry", 0.3)
					return fmt.Sprintf("Repaired %s (durability now %.0f).", loc.Name, loc.BuildingDurability)
				}
			}
			return "No building to repair."
		},
	},
	{
		ID: "commission_construction", Label: "Pay a carpenter to build a structure for you (10 gold)", Category: "construction", BaseGameMinutes: 60,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.GoldCount() < 10 {
				return false
			}
			if n.GetSkillLevel("carpentry") >= 20 || n.Stats.Carpentry >= 20 {
				return false
			}
			for _, other := range w.AliveNPCs() {
				if other.ID != n.ID && (other.Profession == "carpenter" || other.GetSkillLevel("carpentry") >= 30) {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			var carpenter *npc.NPC
			for _, other := range w.AliveNPCs() {
				if other.ID != n.ID && (other.Profession == "carpenter" || other.GetSkillLevel("carpentry") >= 30) {
					carpenter = other
					break
				}
			}
			if carpenter == nil {
				return "No carpenter available."
			}
			if n.GoldCount() < 10 {
				return "Couldn't afford to commission a building (need 10 gold)."
			}
			for typeName, bt := range bld.Types {
				if !canConstruct(carpenter, typeName) {
					continue
				}
				n.RemoveItem("gold", 10)
				carpenter.AddItem("gold", 10)
				for matName, qty := range bt.Materials {
					carpenter.RemoveItem(matName, qty)
				}
				c := &bld.Construction{
					ID:             fmt.Sprintf("construction_%d", len(w.Constructions)+1),
					BuildingType:   typeName,
					Name:           buildingDisplayName(typeName),
					OwnerID:        carpenter.ID,
					CommissionerID: n.ID,
					Progress:       10,
					MaxProgress:    100,
					Durability:     100,
					LocationID:     carpenter.LocationID,
				}
				w.Constructions = append(w.Constructions, c)
				n.AdjustRelationship(carpenter.ID, 5)
				carpenter.AdjustRelationship(n.ID, 5)
				mem.Add(carpenter.ID, memory.Entry{Text: fmt.Sprintf("%s paid me 10 gold to build a %s.", n.Name, typeName), Time: w.TimeString()})
				return fmt.Sprintf("Commissioned %s to build a %s for 10 gold.", carpenter.Name, typeName)
			}
			return fmt.Sprintf("%s doesn't have the materials to build anything right now.", carpenter.Name)
		},
	},
}

func canConstruct(n *npc.NPC, typeName string) bool {
	bt, ok := bld.Types[typeName]
	if !ok {
		return false
	}
	carpEntry := max(int(n.GetSkillLevel("carpentry")), n.Stats.Carpentry)
	if carpEntry < bt.MinCarpentry {
		return false
	}
	for matName, qty := range bt.Materials {
		item := n.HasItem(matName)
		if item == nil || item.Qty < qty {
			return false
		}
	}
	return true
}
