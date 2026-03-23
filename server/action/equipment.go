package action

import (
	"fmt"
	"math"

	"github.com/divinity/core/item"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var equipmentActions = []Action{
	{
		ID: "equip_item", Label: "Equip a weapon, armor, or bag from inventory", Category: "equipment",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.LastAction == "equip_item" || n.LastAction == "unequip_item" {
				return false
			}
			for _, it := range n.Inventory {
				info := item.GetInfo(it.Name)
				if info.Slot == "" {
					continue
				}
				switch info.Slot {
				case "weapon":
					if n.Equipment.Weapon != nil && n.Equipment.Weapon.Name == it.Name {
						continue
					}
					return true
				case "armor":
					if n.Equipment.Armor != nil && n.Equipment.Armor.Name == it.Name {
						continue
					}
					return true
				case "bag":
					if n.Equipment.Bag1 == nil || n.Equipment.Bag2 == nil {
						return true
					}
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			for _, it := range n.Inventory {
				info := item.GetInfo(it.Name)
				if info.Slot == "" {
					continue
				}
				if info.Slot == "bag" && n.Equipment.Bag1 != nil && n.Equipment.Bag2 != nil {
					continue
				}
				if n.EquipItem(it.Name) {
					return fmt.Sprintf("Equipped %s.", it.Name)
				}
			}
			return "Had nothing to equip."
		},
	},
	{
		ID: "unequip_item", Label: "Unequip a worn item back to inventory", Category: "equipment",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.LastAction == "unequip_item" || n.LastAction == "equip_item" {
				return false
			}
			for _, eq := range []*npc.EquipmentSlot{n.Equipment.Weapon, n.Equipment.Armor, n.Equipment.Bag1, n.Equipment.Bag2} {
				if eq != nil && n.CanCarry(eq.Name, 1) {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			for _, slot := range []string{"weapon", "armor", "bag_1", "bag_2"} {
				var eq *npc.EquipmentSlot
				switch slot {
				case "weapon":
					eq = n.Equipment.Weapon
				case "armor":
					eq = n.Equipment.Armor
				case "bag_1":
					eq = n.Equipment.Bag1
				case "bag_2":
					eq = n.Equipment.Bag2
				}
				if eq != nil {
					name := eq.Name
					if n.UnequipItem(slot) {
						return fmt.Sprintf("Unequipped %s.", name)
					}
					return "Inventory too full to unequip."
				}
			}
			return "Had nothing equipped."
		},
	},
	{
		ID: "repair_metal", Label: "Repair a metal weapon or tool (blacksmith)", Category: "equipment", BaseGameMinutes: 45, SkillKey: "blacksmith",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.Profession != "blacksmith" && n.GetSkillLevel("blacksmith") < 30 {
				return false
			}
			if n.HasItem("iron ingot") == nil {
				return false
			}
			return n.Equipment.Weapon != nil && n.Equipment.Weapon.Durability < 76
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("iron ingot", 1)
			eq := n.Equipment.Weapon
			restore := math.Min(95, eq.Durability+40)
			eq.Durability = restore
			n.GainSkill("blacksmith", 0.4)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			return fmt.Sprintf("Repaired %s with an iron ingot (durability now %.0f).", eq.Name, restore)
		},
	},
	{
		ID: "repair_clothing", Label: "Repair armor or bags (tailor)", Category: "equipment", BaseGameMinutes: 45, SkillKey: "tailor",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.Profession != "tailor" && n.GetSkillLevel("tailor") < 30 {
				return false
			}
			if n.HasItem("leather") == nil && n.HasItem("cloth") == nil {
				return false
			}
			for _, eq := range []*npc.EquipmentSlot{n.Equipment.Armor, n.Equipment.Bag1, n.Equipment.Bag2} {
				if eq != nil && eq.Durability < 76 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			mat := "leather"
			if n.HasItem("leather") == nil {
				mat = "cloth"
			}
			n.RemoveItem(mat, 1)
			for _, eq := range []*npc.EquipmentSlot{n.Equipment.Armor, n.Equipment.Bag1, n.Equipment.Bag2} {
				if eq != nil && eq.Durability < 76 {
					eq.Durability = math.Min(95, eq.Durability+35)
					n.GainSkill("tailor", 0.4)
					n.Needs.Fatigue = clampF(n.Needs.Fatigue+6, 0, 100)
					return fmt.Sprintf("Repaired %s with %s (durability now %.0f).", eq.Name, mat, eq.Durability)
				}
			}
			return "Had nothing to repair."
		},
	},
}
