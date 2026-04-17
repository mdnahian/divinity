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
			// If the NPC is at their own home, always give free home sleep
			// (even if home happens to be an inn)
			atHome := n.HomeID != "" && n.LocationID == n.HomeID
			atInn := loc != nil && loc.Type == "inn"
			needsInn := !atHome && (n.HomeID == "" || atInn)
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
			return fmt.Sprintf("Rested and slept (-%d fatigue, -5 stress).", restore)
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
	// camp_rest: rest outdoors using a lean-to shelter. Much better than
	// rough sleeping (no happiness/HP loss), though not as good as a real
	// bed. Available to any NPC carrying a shelter, no gold or inn needed.
	// Fills the critical gap where homeless NPCs were stuck in a death
	// spiral of rough sleep -> low HP -> can't work -> can't afford inn.
	{
		ID: "camp_rest", Label: "Rest in your lean-to shelter (no inn needed)", Category: "survival", BaseGameMinutes: 240,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.Needs.Fatigue <= 50 {
				return false
			}
			return n.HasItem("lean-to shelter") != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			// Shelter provides decent rest without the penalties of rough sleeping
			// and without the cost of an inn. Weather affects quality.
			restore := 30.0
			stressChange := -3
			hpRestore := 3
			weatherNote := ""
			switch w.Weather {
			case "rain":
				restore = 25.0
				weatherNote = " The rain drummed on your shelter roof."
			case "storm":
				restore = 20.0
				stressChange = 2
				hpRestore = 1
				weatherNote = " The storm rattled your shelter all night."
			case "clear":
				restore = 35.0
				stressChange = -5
				hpRestore = 4
				weatherNote = " The clear sky made for a peaceful night."
			}
			n.Needs.Fatigue = clampF(n.Needs.Fatigue-restore, 0, 100)
			n.Stress = clamp(n.Stress+stressChange, 0, 100)
			n.HP = min(100, n.HP+hpRestore)
			mem.Add(n.ID, memory.Entry{
				Text:       "Slept in my lean-to shelter. It's no inn, but it kept the wind out.",
				Time:       w.TimeString(),
				Importance: 0.3,
				Category:   memory.CatRoutine,
				Tags:       []string{"shelter", "sleep", "camp"},
			})
			return fmt.Sprintf("Rested in your lean-to shelter (-%.0f fatigue, %+d stress, +%d HP).%s", restore, stressChange, hpRestore, weatherNote)
		},
	},
	// apply_herbs: use herbs from inventory to heal wounds without a healer.
	// Less effective than a healing potion but uses a common forage item.
	// Fills the gap where injured solo NPCs had no way to heal without
	// finding a healer NPC or buying expensive potions.
	{
		ID: "apply_herbs", Label: "Apply herbs to your wounds (self-heal)", Category: "survival",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.HP < 70 && n.HasItem("herbs") != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("herbs", 1)
			restore := randInt(8, 15)
			n.HP = min(100, n.HP+restore)
			n.GainSkill("herbalism", 0.2)
			return fmt.Sprintf("Applied herbs to your wounds (+%d HP). The forest provides.", restore)
		},
	},
	// bandage_wound: use cloth to bandage wounds without a healer.
	// Less effective than herbs (5-10 HP vs 8-15) but cloth is more
	// widely available — sold at markets, dropped by tailors, spawned
	// with many professions. Fills the gap where an injured NPC has
	// cloth but no herbs (forests being too dangerous to reach).
	{
		ID: "bandage_wound", Label: "Bandage a wound with cloth (self-heal)", Category: "survival",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.HP < 75 && n.HasItem("cloth") != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("cloth", 1)
			restore := randInt(5, 10)
			n.HP = min(100, n.HP+restore)
			return fmt.Sprintf("Bandaged your wound with cloth (+%d HP). Not pretty, but it holds.", restore)
		},
	},
	// name_landmark: leave a personal name/mark for a location, creating
	// a vivid memory tied to that place. Happiness and social boost, no
	// cost. Anchors solo NPCs to the world by giving them small moments
	// of ownership over geography — a gentle antidote to wanderlust and
	// loneliness. Works anywhere outside established buildings.
	{
		ID: "name_landmark", Label: "Leave a personal mark, naming a landmark (free)", Category: "wellbeing", BaseGameMinutes: 20,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.LastAction == "name_landmark" {
				return false
			}
			if n.Needs.Fatigue >= 80 {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return false
			}
			// Natural / outdoor landmarks only (no inns, palaces, castles, shops)
			switch loc.Type {
			case "forest", "farm", "dock", "well", "shrine", "cave",
				"desert", "swamp", "tundra", "garden", "stable":
				return true
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return "Couldn't find the landmark."
			}
			// Pick a flavor phrase based on location type.
			phrase := "I carved my initial into the bark, marking this as mine."
			switch loc.Type {
			case "well":
				phrase = "I etched my mark on a stone at the well's edge."
			case "dock":
				phrase = "I scratched my name into the weathered dock post."
			case "cave":
				phrase = "I scratched my mark into the cave wall with a sharp stone."
			case "shrine":
				phrase = "I left a small pile of stones as my quiet offering at the shrine."
			case "farm":
				phrase = "I pressed my mark into the mud of the field edge."
			case "desert", "tundra":
				phrase = "I built a small cairn of stones to mark my passage."
			case "swamp":
				phrase = "I tied a knot of reeds to the branch overhead."
			case "garden":
				phrase = "I pressed a handprint into the garden soil."
			case "stable":
				phrase = "I scratched my mark into the stable's worn beam."
			}
			n.Happiness = clamp(n.Happiness+5, 0, 100)
			n.Stress = clamp(n.Stress-3, 0, 100)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+2, 0, 100)
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("At %s: %s", loc.Name, phrase),
				Time:       w.TimeString(),
				Importance: 0.55,
				Vividness:  0.8,
				Category:   memory.CatRoutine,
				Tags:       []string{"landmark", "memory", loc.Type},
			})
			return fmt.Sprintf("Named %s with a personal mark (+5 happiness, -3 stress). %s", loc.Name, phrase)
		},
	},
}
