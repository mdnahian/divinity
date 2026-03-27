package action

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/divinity/core/knowledge"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var craftActions = []Action{
	{
		ID: "cook", Label: "Cook a meal at the inn (owner/employee)", Category: "craft", BaseGameMinutes: 30, SkillKey: "cook",
		Destination: destNearestOfType("inn"),
		Candidates:  candidatesOfType("inn"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if (n.HasItem("raw meat") == nil && n.HasItem("fish") == nil) || n.Needs.Fatigue >= 80 {
				return false
			}
			return world.IsWorkerAtType(n, "inn", w)
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			var ingredient *npc.InventoryItem
			if ingredient = n.HasItem("raw meat"); ingredient == nil {
				ingredient = n.HasItem("fish")
			}
			if ingredient == nil {
				return "Had nothing to cook."
			}
			qty := int(math.Min(float64(ingredient.Qty), 2))
			n.RemoveItem(ingredient.Name, qty)
			n.AddItem("cooked meal", qty)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			n.GainSkill("cook", 0.3)
			return fmt.Sprintf("Cooked %d %s into cooked meal(s) at the inn.", qty, ingredient.Name)
		},
	},
	{
		ID: "brew_ale", Label: "Brew ale from wheat or berries (at inn)", Category: "craft", BaseGameMinutes: 30, SkillKey: "barmaid",
		Destination: destNearestOfType("inn"),
		Candidates:  candidatesOfType("inn"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if !world.IsWorkerAtType(n, "inn", w) {
				return false
			}
			wheat := n.HasItem("wheat")
			berries := n.HasItem("berries")
			return (wheat != nil && wheat.Qty >= 2) || (berries != nil && berries.Qty >= 4)
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			wheat := n.HasItem("wheat")
			if wheat != nil && wheat.Qty >= 2 {
				n.RemoveItem("wheat", 2)
			} else {
				n.RemoveItem("berries", 4)
			}
			n.AddItem("ale", 2)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			n.GainSkill("barmaid", 0.4)
			return "Brewed 2 ale at the inn."
		},
	},
	{
		ID: "brew_potion", Label: "Brew a healing potion from herbs", Category: "craft", BaseGameMinutes: 60, SkillKey: "herbalist",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			herbs := n.HasItem("herbs")
			return herbs != nil && herbs.Qty >= 2
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("herbs", 2)
			n.AddItem("healing potion", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			n.GainSkill("herbalist", 0.5)
			return "Carefully brewed herbs into a healing potion."
		},
	},
	{
		ID: "smelt_ore", Label: "Smelt iron ore into ingots (at forge)", Category: "craft", BaseGameMinutes: 60, SkillKey: "blacksmith",
		Destination: destNearestOfType("forge"),
		Candidates:  candidatesOfType("forge"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if !n.HasProfessionOrSkill("blacksmith", "blacksmith", 20) {
				return false
			}
			ore := n.HasItem("iron ore")
			if ore == nil || ore.Qty < 2 {
				return false
			}
			return world.IsWorkerAtType(n, "forge", w)
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			n.RemoveItem("iron ore", 2)
			ingots := 1
			bonus := knowledge.GetTechniqueBonus(n.ID, "smelt_efficiency", w.Techniques)
			if bonus > 0 && rand.Float64() < bonus {
				ingots = 2
			}
			n.AddItem("iron ingot", ingots)
			n.GainSkill("blacksmith", 0.5)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+10, 0, 100)
			return fmt.Sprintf("Smelted 2 iron ore into %d iron ingot(s).", ingots)
		},
	},
	{
		ID: "forge_weapon", Label: "Forge an iron sword (at forge)", Category: "craft", BaseGameMinutes: 90, SkillKey: "blacksmith",
		Destination: destNearestOfType("forge"),
		Candidates:  candidatesOfType("forge"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.GetSkillLevel("blacksmith") < 30 {
				return false
			}
			ingots := n.HasItem("iron ingot")
			if ingots == nil || ingots.Qty < 2 {
				return false
			}
			return world.IsWorkerAtType(n, "forge", w)
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			n.RemoveItem("iron ingot", 2)
			n.AddItem("iron sword", 1)
			bonus := knowledge.GetTechniqueBonus(n.ID, "craft_quality", w.Techniques)
			if bonus > 0 {
				if it := n.HasItem("iron sword"); it != nil {
					it.Durability = it.Durability * (1 + bonus)
				}
			}
			n.GainSkill("blacksmith", 0.8)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+15, 0, 100)
			return "Forged an iron sword from 2 iron ingots."
		},
	},
	{
		ID: "forge_tool", Label: "Forge a tool (at forge)", Category: "craft", BaseGameMinutes: 60, SkillKey: "blacksmith",
		Destination: destNearestOfType("forge"),
		Candidates:  candidatesOfType("forge"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.GetSkillLevel("blacksmith") < 25 {
				return false
			}
			return n.HasItem("iron ingot") != nil && world.IsWorkerAtType(n, "forge", w)
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			n.RemoveItem("iron ingot", 1)
			tools := []string{"pickaxe", "iron axe", "hammer"}
			tool := tools[rand.Intn(len(tools))]
			n.AddItem(tool, 1)
			bonus := knowledge.GetTechniqueBonus(n.ID, "craft_quality", w.Techniques)
			if bonus > 0 {
				if it := n.HasItem(tool); it != nil {
					it.Durability = it.Durability * (1 + bonus)
				}
			}
			n.GainSkill("blacksmith", 0.6)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+12, 0, 100)
			return fmt.Sprintf("Forged a %s from an iron ingot.", tool)
		},
	},
	{
		ID: "tan_hide", Label: "Tan a hide into leather", Category: "craft", BaseGameMinutes: 45,
		Destination: func(n *npc.NPC, w *world.World) string {
			// Check if current location is already a dock or well
			cur := w.LocationByID(n.LocationID)
			if cur != nil && (cur.Type == "dock" || cur.Type == "well") {
				return n.LocationID
			}
			// Find nearest dock or well
			return destNearestOfType("dock")(n, w)
		},
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.HasItem("hide") == nil {
				return false
			}
			return len(w.LocationsByType("dock")) > 0 || len(w.LocationsByType("well")) > 0
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			n.RemoveItem("hide", 1)
			n.AddItem("leather", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			return "Tanned a hide into leather."
		},
	},
	{
		ID: "tailor_craft", Label: "Craft clothing or bag from leather (tailor)", Category: "craft", BaseGameMinutes: 45, SkillKey: "tailor",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if !n.HasProfessionOrSkill("tailor", "tailor", 20) {
				return false
			}
			return n.HasItem("leather") != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("leather", 1)
			items := []string{"belt pouch", "satchel", "cloth tunic"}
			item := items[rand.Intn(len(items))]
			n.AddItem(item, 1)
			n.GainSkill("tailor", 0.5)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			return fmt.Sprintf("Crafted a %s from leather.", item)
		},
	},
	{
		ID: "mill_grain", Label: "Mill wheat into flour (at mill)", Category: "craft", BaseGameMinutes: 30,
		Destination: destNearestOfType("mill"),
		Candidates:  candidatesOfType("mill"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			wheat := n.HasItem("wheat")
			if wheat == nil || wheat.Qty < 2 {
				return false
			}
			for _, mill := range w.LocationsByType("mill") {
				if w.IsWorkerAt(n, mill.ID) || mill.OwnerID == "" {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			n.RemoveItem("wheat", 2)
			n.AddItem("flour", 2)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+6, 0, 100)
			return "Milled 2 wheat into 2 flour."
		},
	},
	{
		ID: "bake_bread_adv", Label: "Bake bread from flour", Category: "craft", BaseGameMinutes: 30, SkillKey: "cook",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			flour := n.HasItem("flour")
			return flour != nil && flour.Qty >= 2
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("flour", 2)
			n.AddItem("bread", 3)
			n.GainSkill("cook", 0.4)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+6, 0, 100)
			return "Baked 3 loaves of bread from flour."
		},
	},
	{
		ID: "craft_pottery", Label: "Craft a ceramic from clay", Category: "craft", BaseGameMinutes: 45,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			clay := n.HasItem("clay")
			return clay != nil && clay.Qty >= 2
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("clay", 2)
			n.AddItem("ceramic", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			return "Shaped clay into a ceramic piece."
		},
	},
	{
		ID: "write_technique", Label: "Write down a technique as a scroll", Category: "craft", BaseGameMinutes: 45,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Literacy < 60 {
				return false
			}
			for _, t := range w.Techniques {
				for _, knower := range t.KnownBy {
					if knower == n.ID && len(t.WrittenIn) == 0 {
						return true
					}
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			for _, t := range w.Techniques {
				for _, knower := range t.KnownBy {
					if knower == n.ID && len(t.WrittenIn) == 0 {
						t.WrittenIn = append(t.WrittenIn, knowledge.WrittenRecord{AuthorID: n.ID, Day: w.GameDay, AuthorAlive: true})
						n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
						n.GainSkill("writing", 0.5)
						return fmt.Sprintf("Wrote down the technique \"%s\" for posterity.", t.Name)
					}
				}
			}
			return "All known techniques are already written."
		},
	},
	{
		ID: "write_journal", Label: "Write a journal about what you observe", Category: "craft", BaseGameMinutes: 45,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Literacy <= 40 || n.Needs.Fatigue >= 75 {
				return false
			}
			if n.LastAction == "write_journal" || n.LastAction == "read_book" {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return false
			}
			if loc.Type != "home" && loc.Type != "inn" && loc.Type != "library" {
				return false
			}
			journalCount := 0
			for _, it := range n.Inventory {
				if it.Name == "written journal" {
					journalCount += it.Qty
				}
			}
			if journalCount >= 3 {
				return false
			}
			return n.Stress > 40 || n.Needs.SocialNeed > 30
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			n.AddItem("written journal", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			n.Happiness = clamp(n.Happiness+3, 0, 100)
			n.Stress = clamp(n.Stress-3, 0, 100)
			n.GainSkill("writing", 0.3)
			return "Wrote a journal about personal reflections."
		},
	},
	{
		ID: "copy_text", Label: "Copy a text or write a manuscript (scribe)", Category: "craft", BaseGameMinutes: 45, SkillKey: "scribe",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if !n.HasProfessionOrSkill("scribe", "writing", 40) {
				return false
			}
			if n.Literacy <= 50 || n.Needs.Fatigue >= 75 {
				return false
			}
			for _, lib := range w.LocationsByType("library") {
				if lib.OwnerID == "" || w.IsWorkerAt(n, lib.ID) {
					return true
				}
			}
			// No libraries exist — allow copying anywhere
			return len(w.LocationsByType("library")) == 0
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.AddItem("book", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			n.GainSkill("writing", 0.4)
			n.GainSkill("scribe", 0.3)
			return "Copied a manuscript into a book (can sell at market)."
		},
	},
}
