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
	// meditate: a solo-friendly shrine action. Available to ANY NPC without
	// stat gates (unlike pray which requires SpiritualSensitivity > 65 or
	// desperation). Reduces stress, boosts happiness, slowly builds
	// spiritual sensitivity. Fills the gap at shrines which otherwise have
	// zero actions for ~90% of NPCs.
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
	// reflect: a solo-friendly action for lonely NPCs. When social need is
	// high (> 40) and no other NPCs are around, the NPC can reflect on
	// their experiences to partially recover social need. This prevents the
	// social death spiral in empty worlds where social crashes to zero with
	// no recovery mechanism.
	{
		ID: "reflect", Label: "Reflect on your experiences (solo social recovery)", Category: "wellbeing", BaseGameMinutes: 20,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.LastAction == "reflect" {
				return false
			}
			if n.Needs.SocialNeed < 40 {
				return false
			}
			return len(w.NPCsAtLocation(n.LocationID, n.ID)) == 0
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			n.Needs.SocialNeed = clampF(n.Needs.SocialNeed-15, 0, 100)
			n.Stress = clamp(n.Stress-3, 0, 100)
			n.Happiness = clamp(n.Happiness+2, 0, 100)
			mem.Add(n.ID, memory.Entry{
				Text:       "I took time to reflect on my journey. The solitude felt purposeful.",
				Time:       w.TimeString(),
				Importance: 0.3,
				Category:   memory.CatRoutine,
				Tags:       []string{"reflect", "solo"},
			})
			return "Reflected on your experiences (-15 social need, -3 stress, +2 happiness)."
		},
	},
	// stargaze: a night-only outdoor action. Free, solo-friendly. Reduces
	// stress and boosts happiness. Small chance of gaining wisdom from
	// seeing a shooting star. Available at outdoor locations (markets,
	// farms, shrines, etc.) but not indoors (inns, libraries).
	{
		ID: "stargaze", Label: "Gaze at the stars (night, outdoor, free)", Category: "wellbeing", BaseGameMinutes: 20,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.LastAction == "stargaze" || !w.IsNight() {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return false
			}
			indoor := map[string]bool{"inn": true, "library": true, "home": true}
			return !indoor[loc.Type]
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			n.Stress = clamp(n.Stress-8, 0, 100)
			n.Happiness = clamp(n.Happiness+6, 0, 100)
			bonus := ""
			if rand.Float64() < 0.12 {
				n.Stats.Wisdom = clamp(n.Stats.Wisdom+1, 0, 100)
				bonus = " A shooting star streaked across the sky — you feel a moment of clarity (+1 wisdom)."
			}
			mem.Add(n.ID, memory.Entry{
				Text:       "I gazed at the stars. The night sky was beautiful.",
				Time:       w.TimeString(),
				Importance: 0.3,
				Category:   memory.CatRoutine,
				Tags:       []string{"stargaze", "night"},
			})
			return fmt.Sprintf("Gazed at the stars (-8 stress, +6 happiness).%s", bonus)
		},
	},
	// tend_shrine: a solo-friendly shrine maintenance action. Gains
	// reputation for tending public structures. Optionally consumes a
	// curiosity item as an offering for bonus spiritual/rep gains.
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
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			n.Stats.Reputation = clamp(n.Stats.Reputation+2, 0, 100)
			n.Happiness = clamp(n.Happiness+2, 0, 100)
			offeringText := ""
			// Check for curiosity items to offer
			curiosities := []string{"pretty stone", "carved bone", "odd trinket", "old coin"}
			for _, c := range curiosities {
				if n.HasItem(c) != nil {
					n.RemoveItem(c, 1)
					n.Stats.Reputation = clamp(n.Stats.Reputation+1, 0, 100)
					if n.Stats.SpiritualSensitivity < 100 {
						n.Stats.SpiritualSensitivity = clamp(n.Stats.SpiritualSensitivity+2, 0, 100)
					}
					offeringText = fmt.Sprintf(" Left a %s as an offering (+1 rep, +2 spiritual).", c)
					break
				}
			}
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("I tended the shrine at %s.%s", loc.Name, offeringText),
				Time:       w.TimeString(),
				Importance: 0.3,
				Category:   memory.CatRoutine,
				Tags:       []string{"shrine", "tend"},
			})
			return fmt.Sprintf("Tended the shrine (+2 reputation, +2 happiness).%s", offeringText)
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
	// healer_triage: a healer-specific self-heal action. Healers can use
	// herbs or medicine on themselves when injured, without needing another
	// NPC. This gives healers a unique solo advantage and fills the gap
	// where heal_patient required a paying target NPC.
	{
		ID: "healer_triage", Label: "Treat your own wounds (healer self-heal)", Category: "wellbeing", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.Profession != "healer" && n.GetSkillLevel("healer") < 20 {
				return false
			}
			if n.HP >= 90 {
				return false
			}
			// Need herbs or medicine
			return n.HasItemOfCategory("medicine") != nil || n.HasItem("herbs") != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			var matName string
			var hpRestore int
			if med := n.HasItemOfCategory("medicine"); med != nil {
				matName = med.Name
				def := item.GetInfo(matName)
				hpMin := int(def.Effects["heal_hp_min"])
				hpMax := int(def.Effects["heal_hp_max"])
				if hpMin <= 0 {
					hpMin = 10
				}
				if hpMax <= hpMin {
					hpMax = hpMin + 10
				}
				hpRestore = randInt(hpMin, hpMax)
				n.RemoveItem(matName, 1)
			} else if n.HasItem("herbs") != nil {
				matName = "herbs"
				n.RemoveItem("herbs", 1)
				hpRestore = randInt(5, 12)
			} else {
				return "Had no healing materials."
			}
			n.HP = min(100, n.HP+hpRestore)
			n.GainSkill("healer", 0.3)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("I treated my own wounds with %s (+%d HP).", matName, hpRestore),
				Time:       w.TimeString(),
				Importance: 0.5,
				Category:   memory.CatRoutine,
				Tags:       []string{"healer", "self-heal"},
			})
			return fmt.Sprintf("Treated your own wounds with %s (+%d HP, now %d/100).", matName, hpRestore, n.HP)
		},
	},
}
