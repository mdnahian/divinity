package action

import (
	"fmt"
	"math"

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
			// Pick the NEAREST safe location. Previously a random safe
			// location was chosen, which could send NPCs on 160+ minute
			// treks while wolves attacked them every tick.
			cur := w.LocationByID(n.LocationID)
			var best *world.Location
			bestDist := math.MaxFloat64
			var cx, cy float64
			if cur != nil {
				cx = float64(cur.X + cur.W/2)
				cy = float64(cur.Y + cur.H/2)
			}
			for _, l := range w.Locations {
				if l.ID == n.LocationID {
					continue
				}
				if len(w.EnemiesAtLocation(l.ID)) != 0 {
					continue
				}
				if cur == nil {
					return l.ID
				}
				dx := float64(l.X+l.W/2) - cx
				dy := float64(l.Y+l.H/2) - cy
				d := math.Abs(dx) + math.Abs(dy)
				if d < bestDist {
					bestDist = d
					best = l
				}
			}
			if best == nil {
				return ""
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
	// guard_patrol: a guard/knight-specific action that patrols a location,
	// gaining combat skill, reputation, and reducing stress for any NPCs
	// present. Fills the gap where guards had zero unique profession actions
	// despite having barracks locations.
	{
		ID: "guard_patrol", Label: "Patrol the area (guard/knight profession)", Category: "combat", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Profession != "guard" && n.Profession != "knight" && n.GetSkillLevel("combat") < 25 {
				return false
			}
			if n.Needs.Fatigue >= 75 || n.LastAction == "guard_patrol" {
				return false
			}
			return true
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			locName := "the area"
			if loc != nil {
				locName = loc.Name
			}
			n.GainSkill("combat", 0.3)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			n.Stats.Reputation = clamp(n.Stats.Reputation+1, 0, 100)
			// If enemies are present, the patrol spots them
			enemies := w.EnemiesAtLocation(n.LocationID)
			if len(enemies) > 0 {
				n.Stress = clamp(n.Stress+5, 0, 100)
				mem.Add(n.ID, memory.Entry{
					Text:       fmt.Sprintf("I patrolled %s and spotted %d hostile creature(s)!", locName, len(enemies)),
					Time:       w.TimeString(),
					Importance: 0.5,
					Category:   memory.CatCombat,
					Tags:       []string{"patrol", n.LocationID},
				})
				return fmt.Sprintf("Patrolled %s — spotted %d enemy(ies)! (+combat skill, +1 rep).", locName, len(enemies))
			}
			// Calm NPCs at the location feel safer
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				o.Stress = clamp(o.Stress-3, 0, 100)
			}
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("I patrolled %s. All clear.", locName),
				Time:       w.TimeString(),
				Importance: 0.2,
				Category:   memory.CatRoutine,
				Tags:       []string{"patrol", n.LocationID},
			})
			return fmt.Sprintf("Patrolled %s — all clear (+combat skill, +1 rep).", locName)
		},
	},
}
