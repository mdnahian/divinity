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
			docks := w.LocationsByType("dock")
			if len(docks) > 0 {
				return docks[0].ID
			}
			wells := w.LocationsByType("well")
			if len(wells) > 0 {
				return wells[0].ID
			}
			return ""
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
			libs := w.LocationsByType("library")
			// If no libraries exist or any is unowned, allow scribing
			if len(libs) == 0 {
				return true
			}
			for _, lib := range libs {
				if lib.OwnerID == "" || w.IsWorkerAt(n, lib.ID) {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.AddItem("book", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			n.GainSkill("writing", 0.4)
			n.GainSkill("scribe", 0.3)
			return "Copied a manuscript into a book (can sell at market)."
		},
	},
	// whittle: a solo-friendly craft that consumes 1 log and produces a
	// carved figurine curiosity. Available anywhere — no workshop needed.
	// Gains carpentry skill slowly. Good for solo players in empty worlds
	// and for builders who want to level carpentry from anywhere.
	{
		ID: "whittle", Label: "Whittle a log into a carved figurine (needs 1 log)", Category: "craft", BaseGameMinutes: 30, SkillKey: "carpentry",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.Needs.Fatigue >= 75 || n.LastAction == "whittle" {
				return false
			}
			return n.HasItem("logs") != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("logs", 1)
			// Quality depends on carpentry skill: higher skill → better chance of bonus
			carpSkill := math.Max(n.GetSkillLevel("carpentry"), float64(n.Stats.Carpentry))
			qty := 1
			bonusText := ""
			if carpSkill >= 50 && rand.Float64() < 0.4 {
				qty = 2
				bonusText = " The carving was so fine you made two."
			}
			n.AddItem("carved figurine", qty)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+4, 0, 100)
			n.Happiness = clamp(n.Happiness+2, 0, 100)
			n.GainSkill("carpentry", 0.3)
			return fmt.Sprintf("Whittled a log into %d carved figurine(s).%s", qty, bonusText)
		},
	},
	// cook_over_fire: cook raw meat or fish anywhere outdoors using 1 log
	// as fuel. No inn, no campfire system, no profession requirements.
	// Produces cooked meal from raw meat or grilled fish from fish. Solo-
	// friendly alternative to the inn-locked cook action.
	{
		ID: "cook_over_fire", Label: "Cook food over an open fire (needs 1 log + raw food)", Category: "craft", BaseGameMinutes: 30, SkillKey: "cook",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 75 {
				return false
			}
			// Can't light a fire in a storm (rain is borderline — allowed but harder)
			if w.Weather == "storm" {
				return false
			}
			if n.HasItem("logs") == nil {
				return false
			}
			hasRawFood := n.HasItem("raw meat") != nil || n.HasItem("fish") != nil
			if !hasRawFood {
				return false
			}
			// Must be at an outdoor location
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return false
			}
			outdoor := map[string]bool{
				"market": true, "farm": true, "forest": true, "dock": true,
				"well": true, "shrine": true, "stable": true, "garden": true,
				"cave": true, "mine": true, "desert": true, "swamp": true,
				"tundra": true, "barracks": true,
			}
			return outdoor[loc.Type]
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			n.RemoveItem("logs", 1)
			var ingredient *npc.InventoryItem
			if ingredient = n.HasItem("raw meat"); ingredient == nil {
				ingredient = n.HasItem("fish")
			}
			if ingredient == nil {
				return "Had nothing to cook."
			}
			qty := int(math.Min(float64(ingredient.Qty), 2))
			ingName := ingredient.Name
			n.RemoveItem(ingName, qty)
			product := "cooked meal"
			if ingName == "fish" {
				product = "grilled fish"
			}
			n.AddItem(product, qty)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+6, 0, 100)
			n.GainSkill("cook", 0.4)
			return fmt.Sprintf("Cooked %d %s into %s over an open fire.", qty, ingName, product)
		},
	},
	// craft_fishing_rod: craft a basic fishing rod from 1 log + 1 rope.
	// Equipping the rod as a weapon gives a fishing bonus (checked during
	// fish action). The rod is a weapon-slot item with minimal combat use
	// but great utility for solo fishing-focused gameplay.
	{
		ID: "craft_fishing_rod", Label: "Craft a fishing rod (1 log + 1 rope)", Category: "craft", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.HasItem("logs") != nil && n.HasItem("rope") != nil && n.Needs.Fatigue < 75
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("logs", 1)
			n.RemoveItem("rope", 1)
			n.AddItem("fishing rod", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			n.GainSkill("fisher", 0.2)
			return "Crafted a fishing rod from a log and rope. Equip it for better catches!"
		},
	},
	// mend_equipment: repair your equipped weapon or armor by using raw
	// materials. Stone repairs weapons, leather repairs armor. Available
	// to all NPCs — no forge or profession needed. Fills the gap where
	// equipment degraded with no self-service repair option.
	{
		ID: "mend_equipment", Label: "Mend your equipped weapon or armor (needs materials)", Category: "craft", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.Needs.Fatigue >= 75 {
				return false
			}
			// Check if weapon needs mending and we have materials
			if n.Equipment.Weapon != nil && n.Equipment.Weapon.Durability < 70 {
				if n.HasItem("stone") != nil || n.HasItem("iron ingot") != nil {
					return true
				}
			}
			// Check if armor needs mending and we have materials
			if n.Equipment.Armor != nil && n.Equipment.Armor.Durability < 70 {
				if n.HasItem("leather") != nil || n.HasItem("cloth") != nil {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			// Prefer repairing weapon first, then armor
			if n.Equipment.Weapon != nil && n.Equipment.Weapon.Durability < 70 {
				mat := ""
				restore := 20.0
				if n.HasItem("iron ingot") != nil {
					mat = "iron ingot"
					restore = 35.0
				} else if n.HasItem("stone") != nil {
					mat = "stone"
					restore = 20.0
				}
				if mat != "" {
					n.RemoveItem(mat, 1)
					before := n.Equipment.Weapon.Durability
					n.Equipment.Weapon.Durability = math.Min(100, n.Equipment.Weapon.Durability+restore)
					n.Needs.Fatigue = clampF(n.Needs.Fatigue+6, 0, 100)
					return fmt.Sprintf("Mended %s with %s (durability %.0f -> %.0f).", n.Equipment.Weapon.Name, mat, before, n.Equipment.Weapon.Durability)
				}
			}
			if n.Equipment.Armor != nil && n.Equipment.Armor.Durability < 70 {
				mat := ""
				restore := 20.0
				if n.HasItem("leather") != nil {
					mat = "leather"
					restore = 30.0
				} else if n.HasItem("cloth") != nil {
					mat = "cloth"
					restore = 15.0
				}
				if mat != "" {
					n.RemoveItem(mat, 1)
					before := n.Equipment.Armor.Durability
					n.Equipment.Armor.Durability = math.Min(100, n.Equipment.Armor.Durability+restore)
					n.Needs.Fatigue = clampF(n.Needs.Fatigue+6, 0, 100)
					return fmt.Sprintf("Mended %s with %s (durability %.0f -> %.0f).", n.Equipment.Armor.Name, mat, before, n.Equipment.Armor.Durability)
				}
			}
			return "Had nothing to mend or lacked the right materials."
		},
	},
	// craft_shelter: build a lean-to shelter from 2 logs + 2 thatch.
	// The shelter is an inventory item that enables "camp_rest" — a solo-
	// friendly way to sleep outdoors with reduced penalties. No home, inn,
	// or profession needed. Fills the gap where homeless NPCs had no way
	// to get decent rest without gold.
	{
		ID: "craft_shelter", Label: "Build a lean-to shelter (2 logs + 2 thatch)", Category: "craft", BaseGameMinutes: 45,
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.Needs.Fatigue >= 75 {
				return false
			}
			logs := n.HasItem("logs")
			thatch := n.HasItem("thatch")
			return logs != nil && logs.Qty >= 2 && thatch != nil && thatch.Qty >= 2
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			n.RemoveItem("logs", 2)
			n.RemoveItem("thatch", 2)
			n.AddItem("lean-to shelter", 1)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			n.GainSkill("carpentry", 0.3)
			return "Built a lean-to shelter from logs and thatch. You can now camp rest outdoors!"
		},
	},
}
