package action

import (
	"fmt"
	"math"
	"math/rand"

	ec "github.com/divinity/core/enemy"
	"github.com/divinity/core/knowledge"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var combatActions = []Action{
	{
		ID: "attack_enemy", Label: "Attack an enemy at your location", Category: "combat",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return len(w.EnemiesAtLocation(n.LocationID)) > 0
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			enemies := w.EnemiesAtLocation(n.LocationID)
			if len(enemies) == 0 {
				return "No enemies to fight."
			}
			e := enemies[0]
			dmgToEnemy := ec.CalculateCombatDamage(n.Stats.Strength, n.Stats.Agility, n.EquippedWeaponBonus(), e.Strength, e.Defense)
			dmgToNpc := ec.CalculateCombatDamage(e.Strength, e.Agility, e.WeaponBonus(), n.Stats.Strength, n.EquippedArmorBonus())
			e.HP = max(0, e.HP-dmgToEnemy)
			n.HP = max(0, n.HP-dmgToNpc)
			n.Stress = clamp(n.Stress+8, 0, 100)

			if n.Equipment.Weapon != nil {
				weaponLoss := 5.0
				wdBonus := knowledge.GetTechniqueBonus(n.ID, "weapon_durability", w.Techniques)
				if wdBonus > 0 {
					weaponLoss = weaponLoss * (1 - wdBonus)
				}
				n.Equipment.Weapon.Durability = math.Max(0, n.Equipment.Weapon.Durability-weaponLoss)
			}
			if n.Equipment.Armor != nil {
				n.Equipment.Armor.Durability = math.Max(0, n.Equipment.Armor.Durability-3)
			}
			if e.HP <= 0 {
				e.Alive = false
				dropped := ec.DropLoot(e)
				for _, d := range dropped {
					w.GroundItems = append(w.GroundItems, world.GroundItem{Name: d.Name, Qty: d.Qty, LocationID: d.LocationID, DroppedDay: w.GameDay, Durability: 80})
				}
				lootText := ""
				if len(dropped) > 0 {
					lootText = " It dropped:"
					for _, d := range dropped {
						lootText += fmt.Sprintf(" %dx %s,", d.Qty, d.Name)
					}
					lootText = lootText[:len(lootText)-1] + "."
				}
				n.Happiness = clamp(n.Happiness+5, 0, 100)
				mem.Add(n.ID, memory.Entry{Text: fmt.Sprintf("I slew a %s!", e.Name), Time: w.TimeString(), Importance: 0.7, Category: memory.CatCombat, Tags: []string{e.Name, n.LocationID}})
				return fmt.Sprintf("Fought the %s and slew it! Dealt %d damage, took %d damage (HP: %d).%s", e.Name, dmgToEnemy, dmgToNpc, n.HP, lootText)
			}
			mem.Add(n.ID, memory.Entry{Text: fmt.Sprintf("I fought a %s but it still lives (HP: %d).", e.Name, e.HP), Time: w.TimeString(), Importance: 0.5, Category: memory.CatCombat, Tags: []string{e.Name, n.LocationID}})
			return fmt.Sprintf("Fought the %s: dealt %d, took %d damage (%s HP: %d, %s HP: %d).", e.Name, dmgToEnemy, dmgToNpc, n.Name, n.HP, e.Name, e.HP)
		},
	},
	{
		ID: "flee_area", Label: "Flee from enemies to a safe location", Category: "combat",
		Destination: func(n *npc.NPC, w *world.World) string {
			cur := w.LocationByID(n.LocationID)
			var safe []*world.Location
			for _, l := range w.Locations {
				if l.ID != n.LocationID && len(w.EnemiesAtLocation(l.ID)) == 0 {
					safe = append(safe, l)
				}
			}
			if len(safe) == 0 {
				return ""
			}
			// Pick nearest safe location instead of random
			if cur == nil {
				return safe[0].ID
			}
			best := safe[0]
			cx := float64(cur.X + cur.W/2)
			cy := float64(cur.Y + cur.H/2)
			bestDist := math.Abs(float64(best.X+best.W/2)-cx) + math.Abs(float64(best.Y+best.H/2)-cy)
			for _, l := range safe[1:] {
				d := math.Abs(float64(l.X+l.W/2)-cx) + math.Abs(float64(l.Y+l.H/2)-cy)
				if d < bestDist {
					best = l
					bestDist = d
				}
			}
			return best.ID
		},
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return len(w.EnemiesAtLocation(n.LocationID)) > 0
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			dest := w.LocationByID(n.LocationID)
			enemies := w.EnemiesAtLocation(n.LocationID)
			eName := "danger"
			if len(enemies) > 0 {
				eName = enemies[0].Name
			}
			destName := "safety"
			if dest != nil {
				destName = dest.Name
			}
			n.Stress = clamp(n.Stress+10, 0, 100)
			mem.Add(n.ID, memory.Entry{Text: fmt.Sprintf("I fled from a %s to %s.", eName, destName), Time: w.TimeString(), Importance: 0.9, Category: memory.CatCombat, Tags: []string{eName, n.LocationID}})
			return fmt.Sprintf("Fled from %s to %s!", eName, destName)
		},
	},
	{
		ID: "party_attack", Label: "Rally nearby NPCs and attack an enemy together", Category: "combat",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if len(w.EnemiesAtLocation(n.LocationID)) == 0 {
				return false
			}
			return len(w.NPCsAtLocation(n.LocationID, n.ID)) > 0
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			enemies := w.EnemiesAtLocation(n.LocationID)
			if len(enemies) == 0 {
				return "No enemies to fight."
			}
			e := enemies[0]
			allies := w.NPCsAtLocation(n.LocationID, n.ID)
			if len(allies) == 0 {
				return "No allies nearby to party with."
			}
			ally := allies[0]
			if target != nil {
				for _, a := range allies {
					if a.ID == target.ID {
						ally = a
						break
					}
				}
			}
			dmg1 := ec.CalculateCombatDamage(n.Stats.Strength, n.Stats.Agility, n.EquippedWeaponBonus(), e.Strength, e.Defense)
			dmg2 := ec.CalculateCombatDamage(ally.Stats.Strength, ally.Stats.Agility, ally.EquippedWeaponBonus(), e.Strength, e.Defense)
			totalDmg := dmg1 + dmg2
			dmgBack := max(1, ec.CalculateCombatDamage(e.Strength, e.Agility, e.WeaponBonus(), n.Stats.Strength, n.EquippedArmorBonus())*6/10)
			e.HP = max(0, e.HP-totalDmg)
			n.HP = max(0, n.HP-dmgBack)
			ally.HP = max(0, ally.HP-max(1, dmgBack-2))
			n.AdjustRelationship(ally.ID, 5)
			ally.AdjustRelationship(n.ID, 5)

			if e.HP <= 0 {
				e.Alive = false
				dropped := ec.DropLoot(e)
				for _, d := range dropped {
					w.GroundItems = append(w.GroundItems, world.GroundItem{Name: d.Name, Qty: d.Qty, LocationID: d.LocationID, DroppedDay: w.GameDay, Durability: 80})
				}
				lootText := ""
				if len(dropped) > 0 {
					lootText = " It dropped:"
					for _, d := range dropped {
						lootText += fmt.Sprintf(" %dx %s,", d.Qty, d.Name)
					}
					lootText = lootText[:len(lootText)-1] + "."
				}
				n.Happiness = clamp(n.Happiness+5, 0, 100)
				ally.Happiness = clamp(ally.Happiness+5, 0, 100)
				mem.Add(n.ID, memory.Entry{Text: fmt.Sprintf("%s and I slew a %s together!", ally.Name, e.Name), Time: w.TimeString(), Importance: 0.7, Category: memory.CatCombat, Tags: []string{ally.ID, e.Name, n.LocationID}})
				mem.Add(ally.ID, memory.Entry{Text: fmt.Sprintf("%s and I slew a %s together!", n.Name, e.Name), Time: w.TimeString(), Importance: 0.7, Category: memory.CatCombat, Tags: []string{n.ID, e.Name, n.LocationID}})
				return fmt.Sprintf("%s and %s fought the %s together and slew it! (%d total damage).%s", n.Name, ally.Name, e.Name, totalDmg, lootText)
			}
			mem.Add(n.ID, memory.Entry{Text: fmt.Sprintf("%s and I attacked a %s together.", ally.Name, e.Name), Time: w.TimeString(), Importance: 0.5, Category: memory.CatCombat, Tags: []string{ally.ID, e.Name, n.LocationID}})
			return fmt.Sprintf("%s and %s fought the %s together: dealt %d damage (%s HP: %d).", n.Name, ally.Name, e.Name, totalDmg, e.Name, e.HP)
		},
	},
}
