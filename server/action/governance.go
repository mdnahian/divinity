package action

import (
	"fmt"
	"math/rand"

	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var governanceActions = []Action{
	{
		ID: "levy_taxes", Label: "Levy Taxes", Category: "governance",
		BaseGameMinutes: 60, SkillKey: "governance",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return n.NobleRank == "duke" || n.NobleRank == "count" || n.NobleRank == "king"
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			territory := w.TerritoryByID(n.TerritoryID)
			if territory == nil {
				return "Had no territory to tax."
			}
			totalTax := 0
			taxed := 0
			for _, other := range w.AliveNPCs() {
				if other.TerritoryID != n.TerritoryID || other.ID == n.ID {
					continue
				}
				if other.NobleRank != "" {
					continue // Don't tax other nobles
				}
				tax := int(float64(other.GoldCount()) * territory.TaxRate)
				if tax < 1 {
					tax = 1
				}
				if other.GoldCount() >= tax {
					other.RemoveItem("gold", tax)
					totalTax += tax
					taxed++
				}
			}
			n.AddItem("gold", totalTax/2)
			territory.Treasury += totalTax - totalTax/2
			n.GainSkill("governance", 0.5)
			return fmt.Sprintf("Levied taxes from %d subjects, collecting %d gold total.", taxed, totalTax)
		},
	},
	{
		ID: "decree_law", Label: "Decree Law", Category: "governance",
		BaseGameMinutes: 45, SkillKey: "governance",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return n.NobleRank == "duke" || n.NobleRank == "king"
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			territory := w.TerritoryByID(n.TerritoryID)
			if territory == nil {
				return "Had no territory to govern."
			}
			laws := []string{
				"increased market tariffs",
				"reduced taxes for farmers",
				"mandatory guard training",
				"curfew after dark",
				"festival day declared",
				"bounty on bandits",
				"trade route protection",
			}
			law := laws[rand.Intn(len(laws))]
			territory.Laws = append(territory.Laws, law)
			if len(territory.Laws) > 5 {
				territory.Laws = territory.Laws[len(territory.Laws)-5:]
			}
			n.GainSkill("governance", 0.8)
			return fmt.Sprintf("Decreed: %s.", law)
		},
	},
	{
		ID: "grant_title", Label: "Grant Title", Category: "governance",
		BaseGameMinutes: 30, SkillKey: "leadership",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.NobleRank != "king" && n.NobleRank != "duke" {
				return false
			}
			// Must have someone nearby without a noble rank
			for _, other := range w.NPCsAtLocation(n.LocationID, n.ID) {
				if other.NobleRank == "" && other.Stats.Reputation >= 40 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				// Find a worthy subject
				for _, other := range w.NPCsAtLocation(n.LocationID, n.ID) {
					if other.NobleRank == "" && other.Stats.Reputation >= 40 {
						target = other
						break
					}
				}
			}
			if target == nil {
				return "Found no one worthy of a title."
			}
			var rank string
			switch n.NobleRank {
			case "king":
				rank = "baron"
			case "duke":
				rank = "knight"
			default:
				rank = "knight"
			}
			target.NobleRank = rank
			target.LiegeID = n.ID
			n.VassalIDs = append(n.VassalIDs, target.ID)
			target.TerritoryID = n.TerritoryID
			n.GainSkill("leadership", 1.0)
			return fmt.Sprintf("Granted the title of %s to %s.", rank, target.Name)
		},
	},
	{
		ID: "hold_court", Label: "Hold Court", Category: "governance",
		BaseGameMinutes: 120, SkillKey: "diplomacy",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return n.NobleRank == "duke" || n.NobleRank == "king" || n.NobleRank == "queen"
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			subjects := w.NPCsAtLocation(n.LocationID, n.ID)
			resolved := 0
			for _, s := range subjects {
				if s.Stress > 50 {
					s.Stress = clamp(s.Stress-15, 0, 100)
					resolved++
				}
				// Improve relationship
				rel := n.Relationships[s.ID]
				n.Relationships[s.ID] = clamp(rel+5, -100, 100)
			}
			n.GainSkill("diplomacy", 1.0)
			n.GainSkill("governance", 0.5)
			return fmt.Sprintf("Held court with %d subjects, resolved %d grievances.", len(subjects), resolved)
		},
	},
	{
		ID: "form_alliance", Label: "Form Alliance", Category: "governance",
		BaseGameMinutes: 60, SkillKey: "diplomacy",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.NobleRank != "duke" && n.NobleRank != "king" {
				return false
			}
			// Must be near another noble
			for _, other := range w.NPCsAtLocation(n.LocationID, n.ID) {
				if other.NobleRank == "duke" || other.NobleRank == "king" {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				for _, other := range w.NPCsAtLocation(n.LocationID, n.ID) {
					if (other.NobleRank == "duke" || other.NobleRank == "king") && other.TerritoryID != n.TerritoryID {
						target = other
						break
					}
				}
			}
			if target == nil {
				return "Found no other rulers to ally with."
			}
			myTerritory := w.TerritoryByID(n.TerritoryID)
			theirTerritory := w.TerritoryByID(target.TerritoryID)
			if myTerritory != nil && theirTerritory != nil {
				myTerritory.Allies = append(myTerritory.Allies, theirTerritory.ID)
				theirTerritory.Allies = append(theirTerritory.Allies, myTerritory.ID)
			}
			rel := n.Relationships[target.ID]
			n.Relationships[target.ID] = clamp(rel+20, -100, 100)
			target.Relationships[n.ID] = clamp(target.Relationships[n.ID]+20, -100, 100)
			n.GainSkill("diplomacy", 1.5)
			return fmt.Sprintf("Formed an alliance with %s.", target.Name)
		},
	},
	{
		ID: "knight_ceremony", Label: "Knight Ceremony", Category: "governance",
		BaseGameMinutes: 30, SkillKey: "leadership",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.NobleRank != "king" && n.NobleRank != "duke" {
				return false
			}
			for _, other := range w.NPCsAtLocation(n.LocationID, n.ID) {
				if other.Profession == "guard" && other.NobleRank == "" && other.GetSkillLevel("combat") >= 40 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				for _, other := range w.NPCsAtLocation(n.LocationID, n.ID) {
					if other.Profession == "guard" && other.NobleRank == "" && other.GetSkillLevel("combat") >= 40 {
						target = other
						break
					}
				}
			}
			if target == nil {
				return "Found no guard worthy of knighting."
			}
			target.NobleRank = "knight"
			target.Profession = "knight"
			target.LiegeID = n.ID
			n.VassalIDs = append(n.VassalIDs, target.ID)
			n.GainSkill("leadership", 1.0)
			return fmt.Sprintf("Knighted %s in a solemn ceremony.", target.Name)
		},
	},
}

func init() {
	Registry = append(Registry, governanceActions...)
}
