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

func destFirstOfType(t string) func(*npc.NPC, *world.World) string {
	return func(_ *npc.NPC, w *world.World) string {
		locs := w.LocationsByType(t)
		if len(locs) == 0 {
			return ""
		}
		return locs[0].ID
	}
}

// anyLocationHasResource returns true if any location of the given type has at
// least 1 of the named resource (or has nil Resources, which means unlimited).
func anyLocationHasResource(w *world.World, locType, resource string) bool {
	for _, loc := range w.LocationsByType(locType) {
		if loc.Resources == nil || loc.Resources[resource] > 0 {
			return true
		}
	}
	return false
}

func destNearestOfType(t string) func(*npc.NPC, *world.World) string {
	return func(n *npc.NPC, w *world.World) string {
		locs := w.LocationsByType(t)
		if len(locs) == 0 {
			return ""
		}
		cur := w.LocationByID(n.LocationID)
		if cur == nil {
			return locs[0].ID
		}
		var best *world.Location
		bestDist := math.MaxFloat64
		cx, cy := float64(cur.X+cur.W/2), float64(cur.Y+cur.H/2)
		for _, l := range locs {
			if w.IsLocationFull(l.ID) {
				continue // skip locations at capacity
			}
			// Skip wells with no water
			if t == "well" && l.Resources != nil && l.Resources["water"] <= 0 {
				continue
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
			return locs[0].ID // fallback if all full
		}
		return best.ID
	}
}

func candidatesOfType(t string) func(*npc.NPC, *world.World) []*world.Location {
	return func(_ *npc.NPC, w *world.World) []*world.Location {
		return w.LocationsByType(t)
	}
}

var gatherActions = []Action{
	{
		ID: "forage", Label: "Forage for food in the forest", Category: "gather", BaseGameMinutes: 30, SkillKey: "herbalism",
		Destination: destNearestOfType("forest"),
		Candidates:  candidatesOfType("forest"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 75 {
				return false
			}
			return anyLocationHasResource(w, "forest", "berries")
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			avail := 99
			if loc != nil && loc.Resources != nil {
				avail = loc.Resources["berries"]
			}
			found := int(math.Min(float64(randInt(1, 3)), float64(avail)))
			if found <= 0 {
				return "The forest has no berries left to forage."
			}
			if loc != nil && loc.Resources != nil {
				loc.Resources["berries"] = max(0, loc.Resources["berries"]-found)
			}
			n.AddItem("berries", found)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			if rand.Float64() < 0.15 {
				herbs := 1
				if loc != nil && loc.Resources != nil {
					herbs = loc.Resources["herbs"]
					if herbs > 0 {
						loc.Resources["herbs"]--
					}
				}
				if herbs > 0 {
					n.AddItem("herbs", 1)
					return fmt.Sprintf("Foraged in the woods and found %d berries and some herbs.", found)
				}
			}
			return fmt.Sprintf("Foraged in the woods and found %d berries.", found)
		},
	},
	{
		ID: "hunt", Label: "Hunt game in the forest", Category: "gather", BaseGameMinutes: 45, SkillKey: "hunter",
		Destination: destNearestOfType("forest"),
		Candidates:  candidatesOfType("forest"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 65 || n.Stats.Strength <= 30 {
				return false
			}
			return anyLocationHasResource(w, "forest", "game")
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+15, 0, 100)
			avail := 99
			if loc != nil && loc.Resources != nil {
				avail = loc.Resources["game"]
			}
			if avail <= 0 {
				return "No game left in the forest to hunt."
			}
			skill := float64(n.Stats.Agility+n.Stats.Strength) / 2
			if rand.Float64()*100 < skill {
				meat := int(math.Min(float64(randInt(1, 2)), float64(avail)))
				if loc != nil && loc.Resources != nil {
					loc.Resources["game"] = max(0, loc.Resources["game"]-meat)
				}
				n.AddItem("raw meat", meat)
				n.GainSkill("hunter", 0.5)
				if rand.Float64() < 0.4 {
					n.AddItem("hide", 1)
					return fmt.Sprintf("Hunted successfully — %d raw meat and a hide.", meat)
				}
				return fmt.Sprintf("Hunted successfully — %d raw meat.", meat)
			}
			n.GainSkill("hunter", 0.1)
			return "Spent hours tracking game but came back empty-handed (+hunting experience)."
		},
	},
	{
		ID: "farm", Label: "Work the wheat fields", Category: "gather", BaseGameMinutes: 30, SkillKey: "farmer",
		Destination: destNearestOfType("farm"),
		Candidates:  candidatesOfType("farm"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 70 {
				return false
			}
			return anyLocationHasResource(w, "farm", "wheat")
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			avail := 99
			if loc != nil && loc.Resources != nil {
				avail = loc.Resources["wheat"]
			}
			harvested := int(math.Min(float64(randInt(1, 4)), float64(avail)))
			if harvested <= 0 {
				return "The fields are barren — no wheat to harvest."
			}
			bonus := knowledge.GetTechniqueBonus(n.ID, "crop_yield", w.Techniques)
			if bonus > 0 {
				harvested = int(math.Ceil(float64(harvested) * (1 + bonus)))
			}
			if loc != nil && loc.Resources != nil {
				loc.Resources["wheat"] = max(0, loc.Resources["wheat"]-harvested)
			}
			n.AddItem("wheat", harvested)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+12, 0, 100)
			n.GainSkill("farmer", 0.3)
			return fmt.Sprintf("Worked the fields and harvested %d wheat.", harvested)
		},
	},
	{
		ID: "fish", Label: "Fish at the dock or river", Category: "gather", BaseGameMinutes: 30, SkillKey: "fisher",
		Destination: destNearestOfType("dock"),
		Candidates:  candidatesOfType("dock"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 70 {
				return false
			}
			return anyLocationHasResource(w, "dock", "fish")
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			avail := 99
			if loc != nil && loc.Resources != nil {
				avail = loc.Resources["fish"]
			}
			if avail <= 0 {
				return "The waters are fished out — nothing to catch."
			}
			if rand.Float64() < 0.6 {
				caught := int(math.Min(float64(randInt(1, 3)), float64(avail)))
				bonus := knowledge.GetTechniqueBonus(n.ID, "fish_catch", w.Techniques)
				if bonus > 0 {
					caught = int(math.Ceil(float64(caught) * (1 + bonus)))
				}
				if loc != nil && loc.Resources != nil {
					loc.Resources["fish"] = max(0, loc.Resources["fish"]-caught)
				}
				n.AddItem("fish", caught)
				n.GainSkill("fisher", 0.3)
				return fmt.Sprintf("Caught %d fish.", caught)
			}
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			n.GainSkill("fisher", 0.1)
			return "Sat by the water but caught nothing (+fishing experience)."
		},
	},
	{
		ID: "chop_wood", Label: "Chop wood in the forest (needs axe)", Category: "gather", BaseGameMinutes: 45, SkillKey: "carpentry",
		Destination: destNearestOfType("forest"),
		Candidates:  candidatesOfType("forest"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 70 || n.Equipment.Weapon == nil || n.Equipment.Weapon.Name != "iron axe" {
				return false
			}
			return anyLocationHasResource(w, "forest", "wood")
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			avail := 99
			if loc != nil && loc.Resources != nil {
				avail = loc.Resources["wood"]
			}
			qty := int(math.Min(float64(randInt(1, 3)), float64(avail)))
			if qty <= 0 {
				return "No trees left to chop in the forest."
			}
			if loc != nil && loc.Resources != nil {
				loc.Resources["wood"] = max(0, loc.Resources["wood"]-qty)
			}
			n.AddItem("logs", qty)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+12, 0, 100)
			n.GainSkill("woodcutter", 0.3)
			if n.Equipment.Weapon != nil {
				n.Equipment.Weapon.Durability = math.Max(0, n.Equipment.Weapon.Durability-1)
			}
			return fmt.Sprintf("Chopped %d logs in the forest.", qty)
		},
	},
	{
		ID: "mine_stone", Label: "Mine stone (needs pickaxe)", Category: "gather", BaseGameMinutes: 45, SkillKey: "miner",
		Destination: destNearestOfType("mine"),
		Candidates:  candidatesOfType("mine"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 70 || n.Equipment.Weapon == nil || n.Equipment.Weapon.Name != "pickaxe" {
				return false
			}
			return anyLocationHasResource(w, "mine", "stone")
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			avail := 99
			if loc != nil && loc.Resources != nil {
				avail = loc.Resources["stone"]
			}
			qty := int(math.Min(float64(randInt(1, 2)), float64(avail)))
			if qty <= 0 {
				return "The mine has no stone left to extract."
			}
			if loc != nil && loc.Resources != nil {
				loc.Resources["stone"] = max(0, loc.Resources["stone"]-qty)
			}
			n.AddItem("stone", qty)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+14, 0, 100)
			n.GainSkill("miner", 0.3)
			if n.Equipment.Weapon != nil {
				n.Equipment.Weapon.Durability = math.Max(0, n.Equipment.Weapon.Durability-1)
			}
			return fmt.Sprintf("Mined %d stone.", qty)
		},
	},
	{
		ID: "mine_ore", Label: "Mine iron ore (needs pickaxe)", Category: "gather", BaseGameMinutes: 45, SkillKey: "miner",
		Destination: destNearestOfType("mine"),
		Candidates:  candidatesOfType("mine"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 65 || n.Equipment.Weapon == nil || n.Equipment.Weapon.Name != "pickaxe" {
				return false
			}
			return anyLocationHasResource(w, "mine", "iron_ore")
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			avail := 99
			if loc != nil && loc.Resources != nil {
				avail = loc.Resources["iron_ore"]
			}
			qty := int(math.Min(float64(randInt(1, 2)), float64(avail)))
			if qty <= 0 {
				return "The mine has no iron ore left."
			}
			if loc != nil && loc.Resources != nil {
				loc.Resources["iron_ore"] = max(0, loc.Resources["iron_ore"]-qty)
			}
			n.AddItem("iron ore", qty)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+16, 0, 100)
			n.GainSkill("miner", 0.4)
			if n.Equipment.Weapon != nil {
				n.Equipment.Weapon.Durability = math.Max(0, n.Equipment.Weapon.Durability-1)
			}
			return fmt.Sprintf("Mined %d iron ore from the depths.", qty)
		},
	},
	{
		ID: "gather_thatch", Label: "Gather thatch from the fields", Category: "gather", BaseGameMinutes: 30,
		Destination: destNearestOfType("farm"),
		Candidates:  candidatesOfType("farm"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 70 {
				return false
			}
			return anyLocationHasResource(w, "farm", "thatch")
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			avail := 99
			if loc != nil && loc.Resources != nil {
				avail = loc.Resources["thatch"]
			}
			qty := int(math.Min(float64(randInt(1, 3)), float64(avail)))
			if qty <= 0 {
				return "No thatch left in the fields to gather."
			}
			if loc != nil && loc.Resources != nil {
				loc.Resources["thatch"] = max(0, loc.Resources["thatch"]-qty)
			}
			n.AddItem("thatch", qty)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+8, 0, 100)
			return fmt.Sprintf("Gathered %d thatch from the fields.", qty)
		},
	},
	{
		ID: "gather_clay", Label: "Gather clay near water", Category: "gather", BaseGameMinutes: 30,
		Candidates: func(_ *npc.NPC, w *world.World) []*world.Location {
			var locs []*world.Location
			for _, l := range w.LocationsByType("dock") {
				if l.Resources != nil && l.Resources["clay"] > 0 {
					locs = append(locs, l)
				}
			}
			for _, l := range w.LocationsByType("well") {
				if l.Resources != nil && l.Resources["clay"] > 0 {
					locs = append(locs, l)
				}
			}
			return locs
		},
		Destination: func(n *npc.NPC, w *world.World) string {
			// Check if current location has clay
			cur := w.LocationByID(n.LocationID)
			if cur != nil && cur.Resources != nil && cur.Resources["clay"] > 0 {
				return n.LocationID
			}
			// Find nearest dock or well with clay
			var candidates []*world.Location
			for _, l := range w.LocationsByType("dock") {
				if l.Resources != nil && l.Resources["clay"] > 0 {
					candidates = append(candidates, l)
				}
			}
			for _, l := range w.LocationsByType("well") {
				if l.Resources != nil && l.Resources["clay"] > 0 {
					candidates = append(candidates, l)
				}
			}
			if len(candidates) == 0 {
				return ""
			}
			if cur == nil {
				return candidates[0].ID
			}
			cx, cy := float64(cur.X+cur.W/2), float64(cur.Y+cur.H/2)
			var best *world.Location
			bestDist := math.MaxFloat64
			for _, l := range candidates {
				d := math.Abs(float64(l.X+l.W/2)-cx) + math.Abs(float64(l.Y+l.H/2)-cy)
				if d < bestDist {
					bestDist = d
					best = l
				}
			}
			return best.ID
		},
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Needs.Fatigue >= 70 {
				return false
			}
			// Check current location
			cur := w.LocationByID(n.LocationID)
			if cur != nil && cur.Resources != nil && cur.Resources["clay"] > 0 {
				return true
			}
			// Check any dock or well with clay
			for _, l := range w.LocationsByType("dock") {
				if l.Resources != nil && l.Resources["clay"] > 0 {
					return true
				}
			}
			for _, l := range w.LocationsByType("well") {
				if l.Resources != nil && l.Resources["clay"] > 0 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			avail := 99
			if loc != nil && loc.Resources != nil {
				avail = loc.Resources["clay"]
			}
			qty := int(math.Min(float64(randInt(1, 2)), float64(avail)))
			if qty <= 0 {
				return "No clay left to gather here."
			}
			if loc != nil && loc.Resources != nil {
				loc.Resources["clay"] = max(0, loc.Resources["clay"]-qty)
			}
			n.AddItem("clay", qty)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+10, 0, 100)
			return fmt.Sprintf("Gathered %d clay near the water.", qty)
		},
	},
	{
		ID: "scavenge", Label: "Pick up items/loot from the ground at your location", Category: "gather",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return len(w.GroundItemsAt(n.LocationID)) > 0
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			items := w.GroundItemsAt(n.LocationID)
			if len(items) == 0 {
				return "Searched around but found nothing."
			}
			var results []string
			for _, item := range items {
				picked := w.PickUpGroundItem(n, item.Name)
				if picked != nil {
					results = append(results, fmt.Sprintf("%dx %s", picked.Qty, picked.Name))
				}
			}
			if len(results) == 0 {
				return "Searched around but found nothing useful."
			}
			result := "Picked up loot from the ground: "
			for i, r := range results {
				if i > 0 {
					result += ", "
				}
				result += r
			}
			return result + "."
		},
	},
}
