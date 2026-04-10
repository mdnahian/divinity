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
	// meditate: a solo-friendly shrine action. Works for ANY NPC (not gated
	// on desperation like `pray`). Restores stress, improves spiritual
	// sensitivity over time, and gives mild happiness. Encourages shrine use
	// during normal play and fills the gap at shrines with no other actions.
	{
		ID: "meditate", Label: "Meditate quietly at the shrine (free, solo-friendly)", Category: "wellbeing", BaseGameMinutes: 30,
		Destination: destNearestOfType("shrine"),
		Candidates:  candidatesOfType("shrine"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.LastAction == "meditate" {
				return false
			}
			return len(w.LocationsByType("shrine")) > 0
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "shrine" {
				return "No shrine to meditate at."
			}
			n.Stress = clamp(n.Stress-12, 0, 100)
			n.Happiness = clamp(n.Happiness+4, 0, 100)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue-3, 0, 100)
			// Slowly awaken spiritual aptitude through practice
			if n.Stats.SpiritualSensitivity < 100 {
				n.Stats.SpiritualSensitivity = clamp(n.Stats.SpiritualSensitivity+1, 0, 100)
			}
			mem.Add(n.ID, memory.Entry{
				Text:       "I meditated in silence at the shrine. My mind felt clearer.",
				Time:       w.TimeString(),
				Importance: 0.4,
				Category:   memory.CatRoutine,
				Tags:       []string{"shrine", "meditate"},
			})
			return "Meditated quietly at the shrine (-12 stress, +4 happiness, -3 fatigue, +spiritual insight)."
		},
	},
	// tend_shrine: a solo-friendly shrine action that rewards pro-social
	// behavior (reputation) for ANY NPC. Uses water from a well if carried
	// otherwise it's a symbolic tending. Encourages worship/routine loops.
	{
		ID: "tend_shrine", Label: "Tend and clean the shrine (free, gains reputation)", Category: "wellbeing", BaseGameMinutes: 30,
		Destination: destNearestOfType("shrine"),
		Candidates:  candidatesOfType("shrine"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.LastAction == "tend_shrine" || n.Needs.Fatigue >= 80 {
				return false
			}
			return len(w.LocationsByType("shrine")) > 0
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "shrine" {
				return "No shrine to tend."
			}
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+6, 0, 100)
			n.Happiness = clamp(n.Happiness+3, 0, 100)
			n.Stats.Reputation = clamp(n.Stats.Reputation+2, 0, 100)
			// If the shrine has a dilapidated building, restore a little
			// durability — thematically the NPC is maintaining the site.
			restored := ""
			if loc.BuildingDurability > 0 && loc.BuildingDurability < 100 {
				loc.BuildingDurability = math.Min(100, loc.BuildingDurability+5)
				restored = fmt.Sprintf(", shrine durability now %.0f", loc.BuildingDurability)
			}
			// Offer an item if the NPC has one to give (curiosity, flowers, clay, etc.)
			offeringText := ""
			for _, pref := range []string{"pretty stone", "carved figurine", "carved bone", "old coin", "berries", "herbs"} {
				if n.HasItem(pref) != nil {
					n.RemoveItem(pref, 1)
					n.Stats.Reputation = clamp(n.Stats.Reputation+1, 0, 100)
					n.Stats.SpiritualSensitivity = clamp(n.Stats.SpiritualSensitivity+1, 0, 100)
					offeringText = fmt.Sprintf(" Left an offering of %s at the altar.", pref)
					break
				}
			}
			mem.Add(n.ID, memory.Entry{
				Text:       "I tended to the shrine — a small, quiet act of devotion.",
				Time:       w.TimeString(),
				Importance: 0.3,
				Category:   memory.CatRoutine,
				Tags:       []string{"shrine", "tend"},
			})
			return fmt.Sprintf("Tended the shrine%s (+3 happiness, +2 reputation).%s", restored, offeringText)
		},
	},
	// stargaze: night-only solo action. Available anywhere outdoors (not
	// inside buildings like inns, palaces, mines). Free, restores stress
	// and boosts happiness — perfect for empty worlds with no socials.
	{
		ID: "stargaze", Label: "Gaze at the night sky (night-only, free)", Category: "wellbeing", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.LastAction == "stargaze" || n.Needs.Fatigue >= 85 {
				return false
			}
			if !w.IsNight() {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return false
			}
			// Must be at an outdoor-type location
			outdoor := map[string]bool{
				"market": true, "farm": true, "forest": true, "dock": true,
				"well": true, "shrine": true, "stable": true, "garden": true,
				"cave": true, "mine": true, "desert": true, "swamp": true,
				"tundra": true, "barracks": true,
			}
			return outdoor[loc.Type]
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			n.Stress = clamp(n.Stress-8, 0, 100)
			n.Happiness = clamp(n.Happiness+6, 0, 100)
			// Rare discovery: a shooting star grants a small bit of wisdom
			wisdomText := ""
			if rand.Float64() < 0.12 {
				if n.Stats.Wisdom < 100 {
					n.Stats.Wisdom = clamp(n.Stats.Wisdom+1, 0, 100)
				}
				wisdomText = " A shooting star cut across the sky — a moment of wonder."
				mem.Add(n.ID, memory.Entry{
					Text:       "I saw a shooting star while stargazing. It felt like a sign.",
					Time:       w.TimeString(),
					Importance: 0.6,
					Category:   memory.CatRoutine,
					Tags:       []string{"stargaze", "omen"},
				})
			}
			return fmt.Sprintf("Gazed at the stars in the quiet night (-8 stress, +6 happiness).%s", wisdomText)
		},
	},
}
