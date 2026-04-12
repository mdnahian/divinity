package action

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/divinity/core/item"
	"github.com/divinity/core/knowledge"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

// outdoorTypes are location types considered "outdoors" where campfires can be built.
var outdoorTypes = map[string]bool{
	"forest": true, "farm": true, "dock": true, "well": true, "garden": true,
	"stable": true, "desert": true, "swamp": true, "tundra": true, "cave": true,
	"market": true, "shrine": true,
}

var wellbeingActions = []Action{
	{
		ID: "heal", Label: "Use medicine to restore HP", Category: "wellbeing",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.HP < 100 && n.HasItemOfCategory("medicine") != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			var bestName string
			var bestHeal float64
			for _, it := range n.Inventory {
				if it.Qty > 0 {
					def := item.GetInfo(it.Name)
					if def.Category == "medicine" {
						h := def.Effects["heal_hp_max"]
						if h > bestHeal {
							bestName = it.Name
							bestHeal = h
						}
					}
				}
			}
			if bestName == "" {
				return "Had nothing to heal with."
			}
			def := item.GetInfo(bestName)
			n.RemoveItem(bestName, 1)
			hpMin := int(def.Effects["heal_hp_min"])
			hpMax := int(def.Effects["heal_hp_max"])
			if hpMin <= 0 {
				hpMin = 10
			}
			if hpMax <= hpMin {
				hpMax = hpMin + 10
			}
			restore := randInt(hpMin, hpMax)
			bonus := knowledge.GetTechniqueBonus(n.ID, "healing_potency", w.Techniques)
			if bonus > 0 {
				restore = int(math.Ceil(float64(restore) * (1 + bonus)))
			}
			n.HP = min(100, n.HP+restore)
			return fmt.Sprintf("Used %s (+%d HP).", bestName, restore)
		},
	},
	{
		ID: "pray", Label: "Pray at the shrine (free)", Category: "wellbeing",
		Destination: destNearestOfType("shrine"),
		Candidates:  candidatesOfType("shrine"),
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			isBeliever := n.Stats.SpiritualSensitivity > 65
			isDesperate := n.Stress > 85 || n.HP < 30 || n.Happiness < 15
			if isBeliever {
				return n.Stress > 40 || n.HP < 70 || n.Happiness < 40
			}
			return isDesperate
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			n.Stress = clamp(n.Stress-10, 0, 100)
			n.Happiness = clamp(n.Happiness+5, 0, 100)

			prayerText := n.LastDialogue
			if prayerText == "" {
				prayerText = "a silent plea"
			}
			w.RecentPrayers = append(w.RecentPrayers, world.Prayer{
				NpcName: n.Name, NpcID: n.ID, Prayer: prayerText,
				Time: w.TimeString(), HP: n.HP, Stress: n.Stress, Hunger: n.Needs.Hunger,
			})
			if len(w.RecentPrayers) > 10 {
				w.RecentPrayers = w.RecentPrayers[1:]
			}

			if n.HP < 100 && n.Stats.SpiritualSensitivity > 60 {
				heal := randInt(3, 8)
				n.HP = min(100, n.HP+heal)
				return fmt.Sprintf("Prayed at the shrine (-10 stress, +5 happiness, +%d HP from deep faith).", heal)
			}
			return "Prayed at the shrine (-10 stress, +5 happiness)."
		},
	},
	{
		ID: "rest", Label: "Rest and relax at the inn", Category: "wellbeing", BaseGameMinutes: 15,
		Destination: destNearestOfType("inn"),
		Candidates:  candidatesOfType("inn"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			inns := w.LocationsByType("inn")
			hasAffordableInn := false
			for _, inn := range inns {
				if w.GetLocationOwner(inn.ID) == nil || innFull(inn, w) {
					continue
				}
				if n.GoldCount() >= innRestCost() {
					hasAffordableInn = true
					break
				}
			}
			if !hasAffordableInn {
				return false
			}
			hasOwnHome := n.HomeID != ""
			stressThreshold := 30
			fatigueThreshold := 40.0
			if hasOwnHome {
				stressThreshold = 70
				fatigueThreshold = 75
			}
			if n.Stress <= stressThreshold && n.Needs.Fatigue <= fatigueThreshold {
				return false
			}
			return true
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "inn" {
				return "No inn available."
			}
			owner := w.GetLocationOwner(loc.ID)
			if owner == nil {
				return "The inn is shut down."
			}
			if innFull(loc, w) {
				return "The inn is full — no rooms available."
			}
			cost := innRestCost()
			if n.GoldCount() < cost {
				return "Couldn't afford to rest at the inn."
			}
			n.RemoveItem("gold", cost)
			if owner.ID != n.ID {
				owner.AddItem("gold", cost)
				mem.Add(owner.ID, memory.Entry{Text: fmt.Sprintf("%s paid %d gold to rest at the inn.", n.Name, cost), Time: w.TimeString()})
			} else {
				w.Treasury += cost
			}
			n.Stress = clamp(n.Stress-8, 0, 100)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue-20, 0, 100)
			if n.HasItem("ale") != nil {
				n.RemoveItem("ale", 1)
				n.Happiness = clamp(n.Happiness+5, 0, 100)
				n.Needs.Thirst = clampF(n.Needs.Thirst+10, 0, 100)
				return fmt.Sprintf("Paid %d gold and relaxed at the inn with an ale (-8 stress, -20 fatigue, +5 happiness).", cost)
			}
			return fmt.Sprintf("Paid %d gold and sat by the hearth at the inn (-8 stress, -20 fatigue).", cost)
		},
	},
	{
		ID: "read_book", Label: "Read a written work", Category: "wellbeing", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.LastAction == "read_book" || n.LastAction == "write_journal" {
				return false
			}
			if n.Literacy <= 30 {
				return false
			}
			return n.HasItem("written journal") != nil || n.HasItem("journal entry") != nil || n.HasItem("book") != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			var itemName string
			for _, name := range []string{"written journal", "journal entry", "book"} {
				if n.HasItem(name) != nil {
					itemName = name
					break
				}
			}
			if itemName == "" {
				return "Had nothing to read."
			}
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+3, 0, 100)
			n.Happiness = clamp(n.Happiness+3, 0, 100)
			litGain := max(1, (100-n.Literacy)*2/100)
			n.Literacy = min(100, n.Literacy+litGain)

			// Journals and books are not consumed on reading
			_ = w
			_ = rand.Float64
			return fmt.Sprintf("Read a %s (+%d literacy, +happiness).", itemName, litGain)
		},
	},
	// --- NEW: Campfire actions ---
	{
		ID: "build_campfire", Label: "Build a campfire from logs (outdoor only)", Category: "wellbeing", BaseGameMinutes: 15,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			logs := n.HasItem("logs")
			if logs == nil || logs.Qty < 1 {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil || !outdoorTypes[loc.Type] {
				return false
			}
			// Don't build if already has campfire
			if loc.HasCampfire() {
				return false
			}
			// Can't build in storm
			if w.Weather == "storm" {
				return false
			}
			return true
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return "Couldn't find a place to build a fire."
			}
			// Rain makes it harder to build
			if w.Weather == "rain" && rand.Float64() < 0.4 {
				n.Needs.Fatigue = clampF(n.Needs.Fatigue+3, 0, 100)
				return "Tried to start a fire in the rain, but the wood was too wet."
			}
			n.RemoveItem("logs", 1)
			if loc.Resources == nil {
				loc.Resources = make(map[string]int)
			}
			loc.Resources["campfire"] = 8 // lasts ~8 resource ticks (40 game ticks)
			if loc.MaxResources == nil {
				loc.MaxResources = make(map[string]int)
			}
			loc.MaxResources["campfire"] = 0 // does not regen
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			n.GainSkill("carpentry", 0.1)
			locName := loc.Name
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("I built a campfire at %s. The warmth feels good.", locName),
				Time:       w.TimeString(),
				Importance: 0.2,
				Category:   "routine",
				Tags:       []string{"campfire", n.LocationID},
			})
			return fmt.Sprintf("Built a campfire at %s from logs. The fire crackles warmly.", locName)
		},
	},
	{
		ID: "warm_by_fire", Label: "Sit by the campfire to rest and recover", Category: "wellbeing", BaseGameMinutes: 15,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.LastAction == "warm_by_fire" {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			return loc != nil && loc.HasCampfire()
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || !loc.HasCampfire() {
				return "The fire has gone out."
			}
			n.Stress = clamp(n.Stress-8, 0, 100)
			n.Happiness = clamp(n.Happiness+4, 0, 100)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue-5, 0, 100)
			// Solo social recovery: sitting by fire with thoughts
			n.Needs.SocialNeed = clampF(n.Needs.SocialNeed-5, 0, 100)
			if w.IsNight() {
				n.Happiness = clamp(n.Happiness+2, 0, 100)
				return "Sat by the campfire under the stars (-8 stress, +6 happiness, -5 fatigue). The night feels peaceful."
			}
			return "Sat by the warm campfire (-8 stress, +4 happiness, -5 fatigue)."
		},
	},
	{
		ID: "campfire_cook", Label: "Cook food over the campfire (fish or meat)", Category: "wellbeing", BaseGameMinutes: 20,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || !loc.HasCampfire() {
				return false
			}
			return n.HasItem("fish") != nil || n.HasItem("raw meat") != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || !loc.HasCampfire() {
				return "The fire has gone out — can't cook."
			}
			var ingredient string
			if n.HasItem("fish") != nil {
				ingredient = "fish"
			} else if n.HasItem("raw meat") != nil {
				ingredient = "raw meat"
			}
			if ingredient == "" {
				return "Had nothing to cook."
			}
			qty := 1
			if it := n.HasItem(ingredient); it != nil && it.Qty >= 2 {
				qty = 2
			}
			n.RemoveItem(ingredient, qty)
			if ingredient == "fish" {
				n.AddItem("grilled fish", qty)
			} else {
				n.AddItem("cooked meal", qty)
			}
			n.GainSkill("cook", 0.2)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+3, 0, 100)
			return fmt.Sprintf("Cooked %d %s over the campfire.", qty, ingredient)
		},
	},
}
