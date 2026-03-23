package action

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/divinity/core/faction"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var socialActions = []Action{
	{
		ID: "talk", Label: "Socialise with a nearby NPC", Category: "social",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return n.Needs.SocialNeed > 10 && len(w.NPCsAtLocation(n.LocationID, n.ID)) > 0
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				if len(nearby) == 0 {
					return "Looked for someone to talk to, but was alone."
				}
				target = nearby[randInt(0, len(nearby)-1)]
			}
			n.Needs.SocialNeed = clampF(n.Needs.SocialNeed-20, 0, 100)
			target.Needs.SocialNeed = clampF(target.Needs.SocialNeed-10, 0, 100)
			n.Happiness = clamp(n.Happiness+3, 0, 100)
			relBonus := 3
			if n.Stats.Charisma > 60 {
				relBonus = 6
			}
			n.AdjustRelationship(target.ID, relBonus)
			target.AdjustRelationship(n.ID, int(math.Ceil(float64(relBonus)/2)))
			if sameFaction(n, target, w) {
				n.AdjustRelationship(target.ID, 5)
				target.AdjustRelationship(n.ID, 5)
			}
			chatText := fmt.Sprintf("Had a chat with %s.", n.Name)
		chatImportance := 0.2
		if n.LastDialogue != "" {
			chatText = fmt.Sprintf("%s said: \"%s\"", n.Name, n.LastDialogue)
			chatImportance = 0.4
		}
		mem.Add(target.ID, memory.Entry{Text: chatText, Time: w.TimeString(), Importance: chatImportance, Category: memory.CatSocial, Tags: []string{n.ID}})

			// Rumor propagation: speaker shares their most notable memory with listener
			topMems := mem.HighestImportance(n.ID, 1)
			if len(topMems) > 0 && topMems[0].Importance >= 0.4 && topMems[0].Category != memory.CatHeard {
				rumor := memory.PropagateRumor(topMems[0], n.Name)
				mem.Add(target.ID, rumor)
			}

			return fmt.Sprintf("Chatted with %s (social +20, relationship +%d).", target.Name, relBonus)
		},
	},
	{
		ID: "gift", Label: "Give food to a hungry nearby NPC", Category: "social",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Stats.Empathy < 40 && n.Stats.Virtue < 50 {
				return false
			}
			if n.HasItemOfCategory("food") == nil {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				if o.Needs.Hunger < 35 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					if o.Needs.Hunger < 35 {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "Wanted to share food but nobody nearby seemed hungry."
			}
			food := n.HasItemOfCategory("food")
			if food == nil {
				return "Had nothing to share."
			}
			foodName := food.Name
			n.RemoveItem(foodName, 1)
			target.AddItem(foodName, 1)
			n.AdjustRelationship(target.ID, 12)
			target.AdjustRelationship(n.ID, 18)
			n.Stats.Reputation = clamp(n.Stats.Reputation+3, 0, 100)
			n.Happiness = clamp(n.Happiness+5, 0, 100)
			if sameFaction(n, target, w) {
				n.AdjustRelationship(target.ID, 5)
				target.AdjustRelationship(n.ID, 5)
			}
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s generously gave me %s.", n.Name, foodName), Time: w.TimeString(), Importance: 0.4, Category: memory.CatSocial, Tags: []string{n.ID}})
			return fmt.Sprintf("Gave %s to %s who was hungry (rep +3, relationship +12).", foodName, target.Name)
		},
	},
	{
		ID: "eat_together", Label: "Share a meal with a nearby NPC", Category: "social",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Hunger > 70 {
				return false
			}
			if n.HasItemOfCategory("food") == nil {
				return false
			}
			return len(w.NPCsAtLocation(n.LocationID, n.ID)) > 0
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				if len(nearby) == 0 {
					return "Nobody to eat with."
				}
				target = nearby[randInt(0, len(nearby)-1)]
			}
			food := n.HasItemOfCategory("food")
			if food == nil {
				return "Had no food to share."
			}
			foodName := food.Name
			n.RemoveItem(foodName, 1)
			n.Needs.Hunger = clampF(n.Needs.Hunger+15, 0, 100)
			target.Needs.Hunger = clampF(target.Needs.Hunger+15, 0, 100)
			n.Needs.SocialNeed = clampF(n.Needs.SocialNeed-15, 0, 100)
			target.Needs.SocialNeed = clampF(target.Needs.SocialNeed-10, 0, 100)
			n.AdjustRelationship(target.ID, 8)
			target.AdjustRelationship(n.ID, 8)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("Shared a meal of %s with %s.", foodName, n.Name), Time: w.TimeString(), Importance: 0.2, Category: memory.CatSocial, Tags: []string{n.ID}})
			if sameFaction(n, target, w) {
				n.AdjustRelationship(target.ID, 5)
				target.AdjustRelationship(n.ID, 5)
			}
			return fmt.Sprintf("Shared %s with %s (+hunger, +social, relationship +8).", foodName, target.Name)
		},
	},
	{
		ID: "drink_together", Label: "Share a drink with a nearby NPC", Category: "social",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return n.HasItemOfCategory("drink") != nil && len(w.NPCsAtLocation(n.LocationID, n.ID)) > 0
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				if len(nearby) == 0 {
					return "Nobody to drink with."
				}
				target = nearby[randInt(0, len(nearby)-1)]
			}
			drinkItem := n.HasItemOfCategory("drink")
			if drinkItem == nil {
				return "Had nothing to drink."
			}
			drinkName := drinkItem.Name
			n.RemoveItem(drinkName, 1)
			n.Sobriety = clamp(n.Sobriety-15, 0, 100)
			target.Sobriety = clamp(target.Sobriety-15, 0, 100)
			n.Needs.SocialNeed = clampF(n.Needs.SocialNeed-20, 0, 100)
			target.Needs.SocialNeed = clampF(target.Needs.SocialNeed-15, 0, 100)
			n.AdjustRelationship(target.ID, 10)
			target.AdjustRelationship(n.ID, 8)
			n.Stress = clamp(n.Stress-8, 0, 100)
			target.Stress = clamp(target.Stress-5, 0, 100)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("Had drinks with %s. Good times.", n.Name), Time: w.TimeString(), Importance: 0.2, Category: memory.CatSocial, Tags: []string{n.ID}})
			if sameFaction(n, target, w) {
				n.AdjustRelationship(target.ID, 5)
				target.AdjustRelationship(n.ID, 5)
			}
			return fmt.Sprintf("Shared %s with %s (-sobriety, +social, relationship +10).", drinkName, target.Name)
		},
	},
	{
		ID: "work_together", Label: "Work alongside a nearby NPC of same profession", Category: "social", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue > 70 {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				if o.Profession == n.Profession {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					if o.Profession == n.Profession {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "No colleague to work with."
			}
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+12, 0, 100)
			target.Needs.Fatigue = clampF(target.Needs.Fatigue+8, 0, 100)
			n.GainSkill(n.Profession, 0.6)
			target.GainSkill(n.Profession, 0.4)
			n.AdjustRelationship(target.ID, 5)
			target.AdjustRelationship(n.ID, 5)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("Worked together with %s. Improved our skills.", n.Name), Time: w.TimeString(), Importance: 0.3, Category: memory.CatSocial, Tags: []string{n.ID}})
			return fmt.Sprintf("Worked together with %s — improved %s skills and bond.", target.Name, n.Profession)
		},
	},
	{
		ID: "comfort", Label: "Comfort a stressed nearby NPC", Category: "social",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Stats.Empathy < 45 {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				if o.Stress > 50 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					if o.Stress > 50 {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "Nobody seemed to need comfort."
			}
			target.Stress = clamp(target.Stress-15, 0, 100)
			target.Stats.Trauma = clamp(target.Stats.Trauma-3, 0, 100)
			n.AdjustRelationship(target.ID, 10)
			target.AdjustRelationship(n.ID, 12)
			n.Needs.SocialNeed = clampF(n.Needs.SocialNeed-10, 0, 100)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s comforted me when I was stressed.", n.Name), Time: w.TimeString(), Importance: 0.4, Category: memory.CatSocial, Tags: []string{n.ID}})
			if sameFaction(n, target, w) {
				n.AdjustRelationship(target.ID, 5)
				target.AdjustRelationship(n.ID, 5)
			}
			return fmt.Sprintf("Comforted %s (-15 stress, -3 trauma, relationship +10).", target.Name)
		},
	},
	{
		ID: "flirt", Label: "Flirt with a nearby NPC", Category: "social",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				if n.GetRelationship(o.ID) > 20 && n.Stats.Attractiveness > 40 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					if n.GetRelationship(o.ID) > 20 {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "Nobody to flirt with."
			}
			charm := float64(n.Stats.Attractiveness+n.Stats.Charisma) / 2
			if rand.Float64()*100 < charm {
				n.AdjustRelationship(target.ID, 12)
				target.AdjustRelationship(n.ID, 10)
				n.Happiness = clamp(n.Happiness+5, 0, 100)
				target.Happiness = clamp(target.Happiness+3, 0, 100)
				mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s flirted with me... I think I liked it.", n.Name), Time: w.TimeString(), Importance: 0.3, Category: memory.CatSocial, Tags: []string{n.ID}})
				return fmt.Sprintf("Flirted with %s successfully (relationship +12).", target.Name)
			}
			n.AdjustRelationship(target.ID, -3)
			n.Stress = clamp(n.Stress+5, 0, 100)
			return fmt.Sprintf("Tried to flirt with %s but it fell flat.", target.Name)
		},
	},
	{
		ID: "recruit_to_faction", Label: "Recruit a nearby NPC to your faction", Category: "social", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.FactionID == "" {
				return false
			}
			var fac *faction.Faction
			for _, f := range w.Factions {
				if f.ID == n.FactionID {
					fac = f
					break
				}
			}
			if fac == nil || fac.LeaderID != n.ID {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				if o.FactionID == "" && n.GetRelationship(o.ID) > 10 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			var fac *faction.Faction
			for _, f := range w.Factions {
				if f.ID == n.FactionID {
					fac = f
					break
				}
			}
			if fac == nil {
				return "Nobody suitable to recruit."
			}
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					if o.FactionID == "" && n.GetRelationship(o.ID) > 10 {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "Nobody suitable to recruit."
			}
			persuasion := n.Stats.Persuasion + n.Stats.Charisma
			resistance := 0
			if target.Stats.Conformity <= 50 {
				resistance = 30
			}
			if rand.Float64()*200 < float64(persuasion-resistance) {
				target.FactionID = fac.ID
				fac.MemberIDs = append(fac.MemberIDs, target.ID)
				n.AdjustRelationship(target.ID, 8)
				target.AdjustRelationship(n.ID, 5)
				mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("Joined %s after %s recruited me.", fac.Name, n.Name), Time: w.TimeString(), Importance: 0.6, Category: memory.CatFaction, Tags: []string{n.ID, fac.ID}})
				return fmt.Sprintf("Recruited %s into %s.", target.Name, fac.Name)
			}
			return fmt.Sprintf("Tried to recruit %s to %s but they declined.", target.Name, fac.Name)
		},
	},
	{
		ID: "share_journal", Label: "Share a written journal with a nearby NPC", Category: "social",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			hasJournal := false
			for _, it := range n.Inventory {
				if (it.Name == "written journal" || it.Name == "journal entry") && it.Qty > 0 {
					hasJournal = true
					break
				}
			}
			if !hasJournal {
				return false
			}
			return len(w.NPCsAtLocation(n.LocationID, n.ID)) > 0
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				if len(nearby) == 0 {
					return "Nobody nearby to share with."
				}
				target = nearby[randInt(0, len(nearby)-1)]
			}
			if !n.RemoveItem("written journal", 1) {
				n.RemoveItem("journal entry", 1)
			}
			target.AddItem("written journal", 1)
			n.AdjustRelationship(target.ID, 5)
			target.AdjustRelationship(n.ID, 5)
			n.Needs.SocialNeed = clampF(n.Needs.SocialNeed-10, 0, 100)
			target.Needs.SocialNeed = clampF(target.Needs.SocialNeed-10, 0, 100)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s shared a journal with me.", n.Name), Time: w.TimeString(), Importance: 0.3, Category: memory.CatSocial, Tags: []string{n.ID}})
			return fmt.Sprintf("Shared a journal with %s (relationship +5).", target.Name)
		},
	},
}

// sameFaction returns true when both NPCs belong to the same non-empty faction.
func sameFaction(a *npc.NPC, b *npc.NPC, _ *world.World) bool {
	if a.FactionID == "" || b.FactionID == "" {
		return false
	}
	return a.FactionID == b.FactionID
}
