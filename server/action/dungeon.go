package action

import (
	"fmt"
	"math/rand"

	"github.com/divinity/core/enemy"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var dungeonActions = []Action{
	{
		ID: "enter_dungeon", Label: "Enter Dungeon", Category: "dungeon",
		BaseGameMinutes: 15, SkillKey: "combat",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return false
			}
			if loc.Type != "cave" && loc.Type != "dungeon_entrance" {
				return false
			}
			dungeon := w.DungeonAtLocation(n.LocationID)
			return dungeon != nil && dungeon.Discovered
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			dungeon := w.DungeonAtLocation(n.LocationID)
			if dungeon == nil {
				return "There is no dungeon here."
			}
			n.GainSkill("combat", 0.3)
			return fmt.Sprintf("Entered the dungeon: %s. It has %d floors.", dungeon.Name, len(dungeon.Floors))
		},
	},
	{
		ID: "explore_floor", Label: "Explore Dungeon Floor", Category: "dungeon",
		BaseGameMinutes: 60, SkillKey: "combat",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || (loc.Type != "cave" && loc.Type != "dungeon_entrance") {
				return false
			}
			dungeon := w.DungeonAtLocation(n.LocationID)
			if dungeon == nil || !dungeon.Discovered {
				return false
			}
			// Must have an uncleared floor
			for _, f := range dungeon.Floors {
				if !f.Cleared {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			dungeon := w.DungeonAtLocation(n.LocationID)
			if dungeon == nil {
				return "No dungeon to explore."
			}

			// Find first uncleared floor
			var floor *world.DungeonFloor
			floorIdx := -1
			for i := range dungeon.Floors {
				if !dungeon.Floors[i].Cleared {
					floor = &dungeon.Floors[i]
					floorIdx = i
					break
				}
			}
			if floor == nil {
				return "All floors have been cleared."
			}

			// Fight enemies on this floor
			killCount := 0
			damageTaken := 0
			for _, enemyID := range floor.EnemyIDs {
				e := findEnemyByID(w, enemyID)
				if e == nil || !e.Alive {
					killCount++
					continue
				}

				// Simple combat: compare NPC strength+weapon vs enemy
				npcPower := n.Stats.Strength + n.EquippedWeaponBonus()
				enemyPower := e.Strength + e.WeaponBonus()

				if npcPower+rand.Intn(20) > enemyPower+rand.Intn(20) {
					e.Alive = false
					killCount++
					// Loot
					for _, item := range e.Inventory {
						n.AddItem(item.Name, item.Qty)
					}
				} else {
					dmg := max(1, enemyPower-n.EquippedArmorBonus()-n.Stats.Endurance/3)
					n.HP -= dmg
					damageTaken += dmg
					if n.HP <= 0 {
						return fmt.Sprintf("Was slain on floor %d of %s by %s.", floorIdx+1, dungeon.Name, e.Name)
					}
				}
			}

			// Check if floor is cleared
			allDead := true
			for _, enemyID := range floor.EnemyIDs {
				e := findEnemyByID(w, enemyID)
				if e != nil && e.Alive {
					allDead = false
					break
				}
			}
			if allDead {
				floor.Cleared = true
				// Award floor loot
				for _, loot := range floor.Loot {
					if rand.Float64() < loot.DropChance {
						qty := loot.MinQty + rand.Intn(max(1, loot.MaxQty-loot.MinQty+1))
						n.AddItem(loot.ItemName, qty)
					}
				}
			}

			n.GainSkill("combat", 1.5)
			if allDead {
				return fmt.Sprintf("Cleared floor %d of %s! Killed %d enemies, took %d damage.", floorIdx+1, dungeon.Name, killCount, damageTaken)
			}
			return fmt.Sprintf("Fought on floor %d of %s. Killed %d enemies, took %d damage. Floor not yet cleared.", floorIdx+1, dungeon.Name, killCount, damageTaken)
		},
	},
	{
		ID: "retreat_dungeon", Label: "Retreat from Dungeon", Category: "dungeon",
		BaseGameMinutes: 15, SkillKey: "",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return false
			}
			return loc.Type == "cave" || loc.Type == "dungeon_entrance"
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			return "Retreated from the dungeon to safety."
		},
	},
}

func findEnemyByID(w *world.World, id string) *enemy.Enemy {
	for _, e := range w.Enemies {
		if e.ID == id {
			return e
		}
	}
	return nil
}

func init() {
	Registry = append(Registry, dungeonActions...)
}
