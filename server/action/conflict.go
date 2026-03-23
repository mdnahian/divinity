package action

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var conflictActions = []Action{
	{
		ID: "steal", Label: "Steal gold from a nearby NPC (RISKY)", Category: "conflict",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Stats.Virtue > 45 || n.LastAction == "steal" {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				if o.GoldCount() > 0 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			if target == nil {
				var rich []*npc.NPC
				for _, o := range nearby {
					if o.GoldCount() > 0 {
						rich = append(rich, o)
					}
				}
				if len(rich) == 0 {
					return "Eyed someone's coin purse but thought better of it."
				}
				target = rich[randInt(0, len(rich)-1)]
			}
			if target == nil || target.GoldCount() <= 0 {
				return "Eyed someone's coin purse but thought better of it."
			}
			thiefSkill := float64(n.Stats.Agility) + float64(n.Stats.Intelligence)*0.5
			victimAwareness := float64(target.Stats.Wisdom) + float64(target.Stats.Agility)*0.5
			catchChance := math.Min(0.9, math.Max(0.2, (victimAwareness-thiefSkill+100)/200))
			caught := rand.Float64() < catchChance

			if caught {
				fine := n.GoldCount() / 2
				if fine > 0 {
					n.RemoveItem("gold", fine)
					target.AddItem("gold", fine)
				}
				n.AdjustRelationship(target.ID, -25)
				target.AdjustRelationship(n.ID, -40)
				n.Stats.Reputation = clamp(n.Stats.Reputation-15, 0, 100)
				n.Stats.Infamy = clamp(n.Stats.Infamy+10, 0, 100)
				n.Stress = clamp(n.Stress+15, 0, 100)

				fightBack := ""
				if rand.Float64() < 0.4 {
					dmg := randInt(5, 15)
					n.HP = max(0, n.HP-dmg)
					fightBack = fmt.Sprintf(" %s fought back and dealt %d damage!", target.Name, dmg)
					mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("I caught %s stealing and gave them a beating.", n.Name), Time: w.TimeString(), Importance: 0.7, Category: memory.CatSocial, Tags: []string{n.ID}})
				}
				mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("Caught %s trying to steal from me! They paid %d gold as a fine.", n.Name, fine), Time: w.TimeString(), Importance: 0.7, Category: memory.CatSocial, Tags: []string{n.ID}})
				for _, wit := range nearby {
					if wit.ID != n.ID && wit.ID != target.ID {
						wit.AdjustRelationship(n.ID, -10)
						mem.Add(wit.ID, memory.Entry{Text: fmt.Sprintf("Witnessed %s trying to steal from %s.", n.Name, target.Name), Time: w.TimeString(), Importance: 0.5, Category: memory.CatSocial, Tags: []string{n.ID, target.ID}})
					}
				}
				return fmt.Sprintf("Tried to steal from %s but got caught! Paid %d gold fine. (rep -15, infamy +10)%s", target.Name, fine, fightBack)
			}

			stolen := min(target.GoldCount(), randInt(1, 3))
			target.RemoveItem("gold", stolen)
			n.AddItem("gold", stolen)
			n.AdjustRelationship(target.ID, -5)
			n.Stats.Reputation = clamp(n.Stats.Reputation-3, 0, 100)
			n.Stats.Infamy = clamp(n.Stats.Infamy+2, 0, 100)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("Noticed %d gold missing from my pouch...", stolen), Time: w.TimeString(), Importance: 0.4, Category: memory.CatSocial})
			return fmt.Sprintf("Quietly stole %d gold from %s.", stolen, target.Name)
		},
	},
	{
		ID: "fight", Label: "Pick a fight with a nearby NPC (SEVERE reputation penalty)", Category: "conflict",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Stats.Aggression < 60 || n.Stress < 60 {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				if n.GetRelationship(o.ID) < -30 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					if n.GetRelationship(o.ID) < -30 {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "Clenched fists but held back."
			}
			atkPower := float64(n.Stats.Strength) + float64(n.Stats.Agility)*0.5 + float64(n.Stats.Courage)*0.3
			defPower := float64(target.Stats.Strength) + float64(target.Stats.Agility)*0.5 + float64(target.Stats.Courage)*0.3
			npcRoll := atkPower * (0.5 + rand.Float64())
			targetRoll := defPower * (0.5 + rand.Float64())

			var winner, loser *npc.NPC
			if npcRoll >= targetRoll {
				winner, loser = n, target
			} else {
				winner, loser = target, n
			}
			dmgToLoser := randInt(10, 25)
			dmgToWinner := randInt(3, 10)
			loser.HP = max(0, loser.HP-dmgToLoser)
			winner.HP = max(0, winner.HP-dmgToWinner)
			loser.Stress = clamp(loser.Stress+15, 0, 100)
			winner.Stress = clamp(winner.Stress+5, 0, 100)
			loser.Stats.Trauma = clamp(loser.Stats.Trauma+5, 0, 100)

			n.AdjustRelationship(target.ID, -15)
			target.AdjustRelationship(n.ID, -20)
			n.Stats.Reputation = clamp(n.Stats.Reputation-15, 0, 100)
			n.Stats.Infamy = clamp(n.Stats.Infamy+10, 0, 100)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+10, 0, 100)
			target.Needs.Fatigue = clampF(target.Needs.Fatigue+10, 0, 100)

			for _, wit := range w.NPCsAtLocation(n.LocationID, n.ID) {
				if wit.ID != n.ID && wit.ID != target.ID {
					wit.AdjustRelationship(n.ID, -10)
					mem.Add(wit.ID, memory.Entry{Text: fmt.Sprintf("Witnessed %s attack %s.", n.Name, target.Name), Time: w.TimeString(), Importance: 0.6, Category: memory.CatSocial, Tags: []string{n.ID, target.ID}})
				}
			}

			if n.Equipment.Weapon != nil {
				n.Equipment.Weapon.Durability = math.Max(0, n.Equipment.Weapon.Durability-2)
			}
			if target.Equipment.Weapon != nil {
				target.Equipment.Weapon.Durability = math.Max(0, target.Equipment.Weapon.Durability-2)
			}
			if n.Equipment.Armor != nil {
				n.Equipment.Armor.Durability = math.Max(0, n.Equipment.Armor.Durability-1)
			}
			if target.Equipment.Armor != nil {
				target.Equipment.Armor.Durability = math.Max(0, target.Equipment.Armor.Durability-1)
			}

			if loser.HP <= 0 && rand.Float64() < 0.3 && loser.GoldCount() > 0 {
				loot := min(loser.GoldCount(), randInt(1, 2))
				loser.RemoveItem("gold", loot)
				winner.AddItem("gold", loot)
			}

			wonByNpc := winner == n
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("Got into a brawl with %s. %s", n.Name, map[bool]string{true: "I lost.", false: "I won."}[wonByNpc]), Time: w.TimeString(), Importance: 0.7, Category: memory.CatSocial, Tags: []string{n.ID}})
			if wonByNpc {
				return fmt.Sprintf("Fought %s and won! (-%d HP self, -%d HP %s) (rep -15, infamy +10)", target.Name, dmgToWinner, dmgToLoser, target.Name)
			}
			return fmt.Sprintf("Fought %s and lost! (-%d HP self, -%d HP %s) (rep -15, infamy +10)", target.Name, dmgToLoser, dmgToWinner, target.Name)
		},
	},
}
