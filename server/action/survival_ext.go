package action

import (
	"fmt"
	"math/rand"

	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

// survivalExtActions contains new survival actions added in the 2026-04-13 playtest.
// These address solo viability gaps: trap hunting, water canteen, basic tool crafting,
// and temporary shelter building.
var survivalExtActions = []Action{
	// --- Hunting Trap System ---
	{
		ID: "set_trap", Label: "Set a hunting trap in the forest (uses rope or logs)", Category: "gather", BaseGameMinutes: 30,
		Destination: destNearestOfType("forest"),
		Candidates:  candidatesOfType("forest"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 70 {
				return false
			}
			// Need rope or logs to build a snare
			hasRope := n.HasItem("rope") != nil
			hasLogs := n.HasItem("logs") != nil
			if !hasRope && !hasLogs {
				return false
			}
			// Limit to 3 traps per NPC
			traps := w.TrapsByOwner(n.ID)
			if len(traps) >= 3 {
				return false
			}
			// Need a forest with game
			for _, f := range w.LocationsByType("forest") {
				if f.Resources == nil || f.Resources["game"] > 0 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "forest" {
				return "Not in a forest. Cannot set a trap here."
			}
			// Consume material
			if n.HasItem("rope") != nil {
				n.RemoveItem("rope", 1)
			} else {
				n.RemoveItem("logs", 1)
			}
			w.Traps = append(w.Traps, world.Trap{
				OwnerID:    n.ID,
				LocationID: loc.ID,
				SetDay:     w.GameDay,
			})
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			n.GainSkill("hunter", 0.2)
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("I set a hunting trap at %s. I should return later to check it.", loc.Name),
				Time:       w.TimeString(),
				Importance: 0.3,
				Category:   memory.CatEconomic,
				Tags:       []string{"trap", loc.ID},
			})
			return fmt.Sprintf("Set a hunting trap at %s. Return later to collect any catch.", loc.Name)
		},
	},
	{
		ID: "check_trap", Label: "Check your hunting traps for caught game", Category: "gather", BaseGameMinutes: 15,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			traps := w.TrapsByOwner(n.ID)
			return len(traps) > 0
		},
		Destination: func(n *npc.NPC, w *world.World) string {
			traps := w.TrapsByOwner(n.ID)
			if len(traps) == 0 {
				return ""
			}
			// Go to the nearest trap location
			cur := w.LocationByID(n.LocationID)
			if cur == nil {
				return traps[0].LocationID
			}
			bestID := traps[0].LocationID
			bestDist := 999999.0
			for _, t := range traps {
				tl := w.LocationByID(t.LocationID)
				if tl == nil {
					continue
				}
				dx := float64(tl.X+tl.W/2) - float64(cur.X+cur.W/2)
				dy := float64(tl.Y+tl.H/2) - float64(cur.Y+cur.H/2)
				if dx < 0 {
					dx = -dx
				}
				if dy < 0 {
					dy = -dy
				}
				d := dx + dy
				if d < bestDist {
					bestDist = d
					bestID = t.LocationID
				}
			}
			return bestID
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			locName := "the area"
			if loc != nil {
				locName = loc.Name
			}
			// Check traps at current location
			var totalCaught int
			var caughtItem string
			for i := len(w.Traps) - 1; i >= 0; i-- {
				t := &w.Traps[i]
				if t.OwnerID != n.ID || t.LocationID != n.LocationID {
					continue
				}
				if t.Caught != "" && t.CaughtQty > 0 {
					n.AddItem(t.Caught, t.CaughtQty)
					caughtItem = t.Caught
					totalCaught += t.CaughtQty
				}
				// Remove trap after checking (single-use)
				w.Traps = append(w.Traps[:i], w.Traps[i+1:]...)
			}
			if totalCaught > 0 {
				n.GainSkill("hunter", 0.3)
				n.Happiness = clamp(n.Happiness+3, 0, 100)
				return fmt.Sprintf("Checked traps at %s and found %d %s. The trap is spent.", locName, totalCaught, caughtItem)
			}
			return fmt.Sprintf("Checked traps at %s but nothing was caught. The trap is spent.", locName)
		},
	},

	// --- Water Canteen System ---
	{
		ID: "craft_canteen", Label: "Craft a water canteen from leather and clay", Category: "craft", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.HasItem("leather") != nil && n.HasItem("clay") != nil && n.HasItem("canteen") == nil && n.HasItem("filled canteen") == nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("leather", 1)
			n.RemoveItem("clay", 1)
			n.AddItem("canteen", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			return "Crafted a water canteen from leather and clay. Fill it at a well."
		},
	},
	{
		ID: "fill_canteen", Label: "Fill your canteen at the well", Category: "survival",
		Destination: destNearestOfType("well"),
		Candidates:  candidatesOfType("well"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.HasItem("canteen") == nil {
				return false
			}
			for _, loc := range w.LocationsByType("well") {
				if loc.Resources != nil && loc.Resources["water"] >= 3 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "well" || loc.Resources == nil || loc.Resources["water"] < 3 {
				return "The well doesn't have enough water to fill the canteen."
			}
			loc.Resources["water"] -= 3
			n.RemoveItem("canteen", 1)
			n.AddItem("filled canteen", 1)
			return "Filled canteen with fresh water from the well (3 drinks)."
		},
	},
	{
		ID: "drink_canteen", Label: "Drink from your canteen (no well needed)", Category: "survival",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.HasItem("filled canteen") != nil && n.Needs.Thirst < 85
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			item := n.HasItem("filled canteen")
			if item == nil {
				return "Canteen is empty."
			}
			n.Needs.Thirst = clampF(n.Needs.Thirst+30, 0, 100)
			// Each canteen has 3 charges tracked via durability
			item.Durability -= 25
			if item.Durability <= 25 {
				n.RemoveItem("filled canteen", 1)
				n.AddItem("canteen", 1)
				return "Drank the last water from canteen (+30 thirst). Canteen is now empty."
			}
			return "Drank water from canteen (+30 thirst)."
		},
	},

	// --- Basic Tool Crafting (no forge needed) ---
	{
		ID: "craft_basic_tool", Label: "Craft a basic tool from sticks and stone (walking stick or club)", Category: "craft", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			hasLogs := n.HasItem("logs") != nil
			hasStone := n.HasItem("stone") != nil
			if !hasLogs {
				return false
			}
			// Can craft walking stick from just logs, or wooden club from logs+stone
			return true || hasStone
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			hasStone := n.HasItem("stone") != nil
			if hasStone {
				n.RemoveItem("logs", 1)
				n.RemoveItem("stone", 1)
				n.AddItem("wooden club", 1)
				n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
				n.GainSkill("carpentry", 0.3)
				return "Crafted a wooden club from logs and stone (weapon bonus +5)."
			}
			n.RemoveItem("logs", 1)
			n.AddItem("walking stick", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			n.GainSkill("carpentry", 0.2)
			return "Crafted a walking stick from a log (weapon bonus +2)."
		},
	},

	// --- Temporary Shelter (Lean-to) ---
	{
		ID: "build_lean_to", Label: "Build a temporary lean-to shelter (logs + thatch)", Category: "construction", BaseGameMinutes: 45,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 75 {
				return false
			}
			logs := n.HasItem("logs")
			thatch := n.HasItem("thatch")
			if logs == nil || logs.Qty < 2 || thatch == nil || thatch.Qty < 2 {
				return false
			}
			// Can only build outdoors (not inside buildings)
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return false
			}
			outdoorTypes := map[string]bool{
				"forest": true, "farm": true, "dock": true,
				"desert": true, "swamp": true, "tundra": true,
				"cave": true,
			}
			return outdoorTypes[loc.Type]
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return "Cannot build a shelter here."
			}
			n.RemoveItem("logs", 2)
			n.RemoveItem("thatch", 2)
			n.HomeID = loc.ID // Temporary home at this location
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+12, 0, 100)
			n.GainSkill("carpentry", 0.4)
			n.Happiness = clamp(n.Happiness+5, 0, 100)
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("I built a lean-to shelter at %s. I can sleep here without penalty.", loc.Name),
				Time:       w.TimeString(),
				Importance: 0.5,
				Category:   memory.CatRoutine,
				Tags:       []string{"shelter", loc.ID},
			})
			return fmt.Sprintf("Built a lean-to shelter at %s using 2 logs and 2 thatch. This is now your temporary home.", loc.Name)
		},
	},

	// --- Craft Snare (for traps) ---
	{
		ID: "craft_snare", Label: "Craft a snare from rope (for setting traps)", Category: "craft", BaseGameMinutes: 15,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.HasItem("rope") != nil && n.HasItem("snare") == nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("rope", 1)
			n.AddItem("snare", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+3, 0, 100)
			return "Crafted a snare from rope. Use it to set a hunting trap in the forest."
		},
	},

	// --- Foraging Knowledge (repeated foraging reveals rare items) ---
	// This is implemented via the herbalism skill bonus in forage action.
	// At higher herbalism skill, rare finds become more common.

	// --- Landmark Discovery Bonus ---
	// This enhances the existing explore action. When exploring, if the NPC
	// has never visited the destination, they get a bonus curiosity item.
	// We track this via a simple check: if the NPC has no memories tagged
	// with the destination location ID, it's a first visit.

	// --- Scavenge Value Sort Fix ---
	// The scavenge action picks items in insertion order. This fix sorts
	// ground items by estimated value before picking.
}

func init() {
	// Register the extended survival actions with the global registry
	Registry = append(Registry, survivalExtActions...)
}

// forageKnowledgeBonus returns bonus rare item chance based on herbalism skill.
// Used by foraging to reveal rarer items with repeated visits.
func forageKnowledgeBonus(herbalismSkill float64) float64 {
	if herbalismSkill < 20 {
		return 0
	}
	return herbalismSkill / 200.0 // 10% at skill 20, 50% at skill 100
}

// scavengeItemValue estimates the value of a ground item for scavenge sorting.
func scavengeItemValue(name string) int {
	values := map[string]int{
		"gold": 100, "iron ore": 15, "iron ingot": 20, "logs": 8, "stone": 7,
		"leather": 12, "hide": 10, "rope": 9, "clay": 6, "thatch": 4,
		"iron sword": 30, "iron axe": 25, "pickaxe": 25, "hammer": 20,
		"leather armor": 25, "cloth tunic": 10,
		"healing potion": 18, "herbs": 8,
		"cooked meal": 6, "bread": 3, "fish": 4, "raw meat": 5, "berries": 2, "wheat": 3,
		"ale": 5, "flour": 4, "cloth": 5,
		"backpack": 20, "satchel": 12, "belt pouch": 8,
		"book": 10, "written journal": 5, "journal entry": 3,
		"spice bundle": 7, "odd trinket": 4, "pretty stone": 3, "old coin": 5, "carved bone": 4,
		"ceramic": 8, "snare": 5, "canteen": 10,
	}
	if v, ok := values[name]; ok {
		return v
	}
	return 1
}
