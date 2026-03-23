package action

import (
	"fmt"

	"github.com/divinity/core/item"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var survivalActions = []Action{
	{
		ID: "eat", Label: "Eat food from inventory", Category: "survival",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.Needs.Hunger >= 90 {
				return false
			}
			for _, it := range n.Inventory {
				if it.Qty > 0 {
					def := item.GetInfo(it.Name)
					if def.Category == "food" && def.Effects["hunger_restore"] > 0 {
						return true
					}
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			var bestName string
			var bestRestore float64
			for _, it := range n.Inventory {
				if it.Qty > 0 {
					def := item.GetInfo(it.Name)
					if def.Category == "food" && def.Effects["hunger_restore"] > bestRestore {
						bestName = it.Name
						bestRestore = def.Effects["hunger_restore"]
					}
				}
			}
			if bestName == "" {
				return "Had nothing to eat."
			}
			restore := int(bestRestore)
			if restore == 0 {
				restore = 15
			}
			n.RemoveItem(bestName, 1)
			n.Needs.Hunger = clampF(n.Needs.Hunger+float64(restore), 0, 100)
			return fmt.Sprintf("Ate %s (+%d hunger).", bestName, restore)
		},
	},
	{
		ID: "drink", Label: "Go to the well and drink fresh water (free, uses well water)", Category: "survival",
		Destination: destNearestOfType("well"),
		Candidates:  candidatesOfType("well"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			for _, loc := range w.LocationsByType("well") {
				if loc.Resources != nil && loc.Resources["water"] > 0 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "well" || loc.Resources == nil || loc.Resources["water"] <= 0 {
				return "The well is dry. No water to drink."
			}
			loc.Resources["water"]--
			n.Needs.Thirst = 100
			return "Drank fresh water at the well (thirst fully restored)."
		},
	},
	{
		ID: "sleep", Label: "Sleep (at home: free; at inn: 3g+1/guest; rough: free but miserable)", Category: "survival", BaseGameMinutes: 240,
		Destination: func(n *npc.NPC, w *world.World) string {
			// Critical fatigue: rough-sleep wherever you are
			if n.Needs.Fatigue >= 80 {
				return n.LocationID
			}
			if n.HomeID != "" {
				return n.HomeID
			}
			// Find nearest non-full, affordable inn
			if inn := nearestAffordableInn(n, w); inn != "" {
				return inn
			}
			return n.LocationID // fallback: rough sleep
		},
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue <= 50 {
				return false
			}
			// Critical fatigue: can always rough-sleep wherever you are
			if n.Needs.Fatigue >= 80 {
				return true
			}
			// Normal fatigue: need home or affordable inn
			if n.HomeID != "" {
				return true
			}
			return nearestAffordableInn(n, w) != ""
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			atInn := loc != nil && loc.Type == "inn"
			needsInn := n.HomeID == "" || (atInn && n.HomeID == n.LocationID)
			// Rough sleep: no home and can't afford inn
			if needsInn {
				canAffordInn := false
				if atInn && !innFull(loc, w) {
					if owner := w.GetLocationOwner(loc.ID); owner != nil {
						guests := len(w.NPCsAtLocation(loc.ID, ""))
						cost := innSleepCost(guests)
						if n.GoldCount() >= cost {
							canAffordInn = true
						}
					}
				}
				if !canAffordInn {
					n.Needs.Fatigue = clampF(n.Needs.Fatigue-20, 0, 100)
					n.Happiness = clamp(n.Happiness-10, 0, 100)
					n.Stress = clamp(n.Stress+10, 0, 100)
					n.HP = max(0, n.HP-3)
					n.Hygiene = clamp(n.Hygiene-10, 0, 100)
					mem.Add(n.ID, memory.Entry{
						Text:       "I slept on the cold ground outside. I need to find a proper place to live.",
						Time:       w.TimeString(),
						Importance: 0.7,
						Category:   memory.CatRoutine,
						Tags:       []string{"homeless", "sleep"},
					})
					return "Slept rough on the ground (-20 fatigue, -10 happiness, +10 stress, -3 HP, -10 hygiene). You desperately need a home or inn."
				}
			}
			// Inn payment
			if needsInn && atInn {
				owner := w.GetLocationOwner(loc.ID)
				guests := len(w.NPCsAtLocation(loc.ID, ""))
				cost := innSleepCost(guests)
				n.RemoveItem("gold", cost)
				if owner != nil && owner.ID != n.ID {
					owner.AddItem("gold", cost)
					mem.Add(owner.ID, memory.Entry{Text: fmt.Sprintf("%s paid %d gold for a night at the inn.", n.Name, cost), Time: w.TimeString()})
				} else {
					w.Treasury += cost
				}
			}
			restore := 40
			n.Needs.Fatigue = clampF(n.Needs.Fatigue-float64(restore), 0, 100)
			n.Stress = clamp(n.Stress-5, 0, 100)
			n.HP = min(100, n.HP+5)
			if needsInn && atInn {
				return fmt.Sprintf("Paid gold for a room at the inn (-%d fatigue, -5 stress).", restore)
			}
			if n.HomeID != "" {
				return fmt.Sprintf("Slept at home (-%d fatigue, -5 stress).", restore)
			}
			return fmt.Sprintf("Slept at home (-%d fatigue, -5 stress).", restore)
		},
	},
	{
		ID: "bathe", Label: "Bathe at the well to clean up (uses well water)", Category: "survival",
		Destination: destNearestOfType("well"),
		Candidates:  candidatesOfType("well"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Hygiene >= 30 || n.LastAction == "bathe" {
				return false
			}
			for _, loc := range w.LocationsByType("well") {
				if loc.Resources != nil && loc.Resources["water"] >= 2 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "well" || loc.Resources == nil || loc.Resources["water"] < 2 {
				return "The well doesn't have enough water to bathe."
			}
			loc.Resources["water"] -= 2
			n.Hygiene = clamp(n.Hygiene+50, 0, 100)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+3, 0, 100)
			n.Happiness = clamp(n.Happiness+3, 0, 100)
			return "Bathed at the well (+50 hygiene)."
		},
	},
	{
		ID: "drink_ale", Label: "Drink a beverage from inventory", Category: "survival",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.HasItemOfCategory("drink") != nil && (n.Stress > 30 || n.Sobriety > 40)
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			drinkItem := n.HasItemOfCategory("drink")
			if drinkItem == nil {
				return "Had nothing to drink."
			}
			drinkName := drinkItem.Name
			def := item.GetInfo(drinkName)
			n.RemoveItem(drinkName, 1)
			if v := def.Effects["sobriety"]; v != 0 {
				n.Sobriety = clamp(n.Sobriety+int(v), 0, 100)
			}
			if v := def.Effects["stress"]; v != 0 {
				n.Stress = clamp(n.Stress+int(v), 0, 100)
			}
			if v := def.Effects["happiness"]; v != 0 {
				n.Happiness = clamp(n.Happiness+int(v), 0, 100)
			}
			if v := def.Effects["thirst"]; v != 0 {
				n.Needs.Thirst = clampF(n.Needs.Thirst+v, 0, 100)
			}
			return fmt.Sprintf("Drank %s.", drinkName)
		},
	},
}
