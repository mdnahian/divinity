package action

import (
	"fmt"
	"math"

	"github.com/divinity/core/knowledge"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var educationActions = []Action{
	{
		ID: "teach", Label: "Teach your skills to a nearby NPC (2 gold lesson fee)", Category: "education", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Stats.TeachingDrive < 40 {
				return false
			}
			topSkill, topVal := topSkillEntry(n)
			if topVal < 50 {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			generous := n.Stats.Generosity > 70
			for _, o := range nearby {
				if o.GetSkillLevel(topSkill) < topVal-10 && (o.GoldCount() >= 2 || generous) {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			topSkill, topVal := topSkillEntry(n)
			if topVal < 50 {
				return "Had nothing worth teaching."
			}
			generous := n.Stats.Generosity > 70
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					if o.GetSkillLevel(topSkill) < topVal-10 && (o.GoldCount() >= 2 || generous) {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "Found nobody to teach."
			}
			gain := topVal * 0.02 * (float64(target.Stats.Intelligence) / 100)
			actualGain := math.Max(0.2, math.Round(gain*10)/10)
			target.GainSkill(topSkill, actualGain)
			n.GainSkill("teaching", 0.2)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)

			payStr := ""
			if target.GoldCount() >= 2 {
				target.RemoveItem("gold", 2)
				n.AddItem("gold", 2)
				payStr = " for 2 gold"
			}
			n.AdjustRelationship(target.ID, 4)
			target.AdjustRelationship(n.ID, 6)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s taught me about %s%s (+%.1f).", n.Name, topSkill, payStr, actualGain), Time: w.TimeString(), Importance: 0.4, Category: memory.CatEducation, Tags: []string{n.ID, topSkill}})
			return fmt.Sprintf("Taught %s about %s%s (+%.1f skill).", target.Name, topSkill, payStr, actualGain)
		},
	},
	{
		ID: "teach_literacy", Label: "Teach someone to read and write (1 gold)", Category: "education", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Literacy < 60 || n.Stats.TeachingDrive < 35 {
				return false
			}
			generous := n.Stats.Generosity > 70
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				if o.Literacy < n.Literacy-20 && (o.GoldCount() >= 1 || generous) {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			generous := n.Stats.Generosity > 70
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					if o.Literacy < n.Literacy-20 && (o.GoldCount() >= 1 || generous) {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "Nobody nearby needed literacy lessons."
			}
			gain := max(1, int(math.Round(float64(n.Literacy)*0.03*float64(target.Stats.Intelligence)/100)))
			target.Literacy = min(100, target.Literacy+gain)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)

			payStr := ""
			if target.GoldCount() >= 1 {
				target.RemoveItem("gold", 1)
				n.AddItem("gold", 1)
				payStr = " for 1 gold"
			}
			n.AdjustRelationship(target.ID, 3)
			target.AdjustRelationship(n.ID, 5)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s gave me a reading lesson%s (+%d literacy).", n.Name, payStr, gain), Time: w.TimeString(), Importance: 0.4, Category: memory.CatEducation, Tags: []string{n.ID, "literacy"}})
			return fmt.Sprintf("Taught %s to read%s (+%d literacy, now %d/100 — %s).", target.Name, payStr, gain, target.Literacy, target.LiteracyLevel())
		},
	},
	{
		ID: "teach_technique", Label: "Teach a named technique to a nearby NPC", Category: "education", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			knownTechs := npcTechniques(n.ID, w)
			if len(knownTechs) == 0 {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				for _, t := range knownTechs {
					if !contains(t.KnownBy, o.ID) {
						return true
					}
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			knownTechs := npcTechniques(n.ID, w)
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					for _, t := range knownTechs {
						if !contains(t.KnownBy, o.ID) {
							target = o
							break
						}
					}
					if target != nil {
						break
					}
				}
			}
			if target == nil {
				return "Nobody nearby to teach techniques to."
			}
			for _, t := range knownTechs {
				if !contains(t.KnownBy, target.ID) {
					t.KnownBy = append(t.KnownBy, target.ID)
					n.AdjustRelationship(target.ID, 5)
					target.AdjustRelationship(n.ID, 8)
					mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s taught me the technique \"%s\".", n.Name, t.Name), Time: w.TimeString(), Importance: 0.5, Category: memory.CatEducation, Tags: []string{n.ID, t.Name}})
					return fmt.Sprintf("Taught %s the technique \"%s\" (%s).", target.Name, t.Name, t.BonusLabel)
				}
			}
			return fmt.Sprintf("%s already knows all my techniques.", target.Name)
		},
	},
}

func topSkillEntry(n *npc.NPC) (string, float64) {
	var bestName string
	var bestVal float64
	for k, v := range n.Skills {
		if v > bestVal {
			bestName = k
			bestVal = v
		}
	}
	return bestName, bestVal
}

func npcTechniques(npcID string, w *world.World) []*knowledge.Technique {
	var result []*knowledge.Technique
	for _, t := range w.Techniques {
		if contains(t.KnownBy, npcID) {
			result = append(result, t)
		}
	}
	return result
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

