package action

import (
	"fmt"
	"math/rand"

	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var employmentActions = []Action{
	{
		ID: "hire_employee", Label: "Hire a nearby unemployed NPC to work at your establishment", Category: "employment",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if !n.IsBusinessOwner {
				return false
			}
			if n.LastAction == "hire_employee" || n.LastAction == "fire_employee" {
				return false
			}
			var loc *world.Location
			for _, l := range w.Locations {
				if l.OwnerID == n.ID {
					loc = l
					break
				}
			}
			if loc == nil {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, o := range nearby {
				if o.EmployerID == "" && !o.IsBusinessOwner {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			var loc *world.Location
			for _, l := range w.Locations {
				if l.OwnerID == n.ID {
					loc = l
					break
				}
			}
			if loc == nil {
				return "You don't own an establishment."
			}
			if target != nil && target.EmployerID == n.ID {
				return fmt.Sprintf("%s already works for you.", target.Name)
			}
			if target == nil || target.EmployerID != "" {
				target = nil
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, o := range nearby {
					if o.EmployerID == "" && !o.IsBusinessOwner {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "No one nearby is looking for work."
			}
			wage := 2 + rand.Intn(3)
			target.EmployerID = n.ID
			target.WorkplaceID = loc.ID
			target.Wage = wage
			target.UnpaidDays = 0
			n.AdjustRelationship(target.ID, 5)
			target.AdjustRelationship(n.ID, 5)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s hired me to work at %s for %d gold/day.", n.Name, loc.Name, wage), Time: w.TimeString()})
			return fmt.Sprintf("Hired %s to work at %s for %d gold/day.", target.Name, loc.Name, wage)
		},
	},
	{
		ID: "fire_employee", Label: "Fire an employee from your establishment", Category: "employment",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if !n.IsBusinessOwner {
				return false
			}
			if n.LastAction == "fire_employee" || n.LastAction == "hire_employee" {
				return false
			}
			for _, o := range w.AliveNPCs() {
				if o.EmployerID == n.ID {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				for _, o := range w.AliveNPCs() {
					if o.EmployerID == n.ID {
						target = o
						break
					}
				}
			}
			if target == nil {
				return "No employees to fire."
			}
			locName := "the establishment"
			if loc := w.LocationByID(target.WorkplaceID); loc != nil {
				locName = loc.Name
			}
			target.EmployerID = ""
			target.WorkplaceID = ""
			target.Wage = 0
			target.UnpaidDays = 0
			n.AdjustRelationship(target.ID, -5)
			target.AdjustRelationship(n.ID, -15)
			target.Stress = clamp(target.Stress+10, 0, 100)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s fired me from %s.", n.Name, locName), Time: w.TimeString()})
			return fmt.Sprintf("Fired %s from %s.", target.Name, locName)
		},
	},
	{
		ID: "quit_job", Label: "Quit your current job", Category: "employment",
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			if n.EmployerID == "" {
				return false
			}
			if n.LastAction == "seek_employment" {
				return false
			}
			relWithBoss := n.GetRelationship(n.EmployerID)
			return n.Stress > 60 || relWithBoss < -10 || n.UnpaidDays >= 2
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			employer := w.FindNPCByID(n.EmployerID)
			locName := "the establishment"
			if loc := w.LocationByID(n.WorkplaceID); loc != nil {
				locName = loc.Name
			}
			empName := "my employer"
			if employer != nil {
				empName = employer.Name
				n.AdjustRelationship(employer.ID, -5)
				employer.AdjustRelationship(n.ID, -8)
				mem.Add(employer.ID, memory.Entry{Text: fmt.Sprintf("%s quit working at %s.", n.Name, locName), Time: w.TimeString()})
			}
			n.EmployerID = ""
			n.WorkplaceID = ""
			n.Wage = 0
			n.UnpaidDays = 0
			return fmt.Sprintf("Quit working at %s for %s.", locName, empName)
		},
	},
	{
		ID: "seek_employment", Label: "Ask the owner of this establishment for work", Category: "employment",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.EmployerID != "" || n.IsBusinessOwner {
				return false
			}
			if n.LastAction == "quit_job" || n.LastAction == "seek_employment" {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.OwnerID == "" || loc.OwnerID == n.ID {
				return false
			}
			owner := w.FindNPCByID(loc.OwnerID)
			return owner != nil && owner.Alive
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			owner := w.FindNPCByID(loc.OwnerID)
			if owner == nil {
				return "The owner is not available."
			}
			wage := 2 + rand.Intn(2)
			n.EmployerID = owner.ID
			n.WorkplaceID = loc.ID
			n.Wage = wage
			n.UnpaidDays = 0
			n.AdjustRelationship(owner.ID, 5)
			owner.AdjustRelationship(n.ID, 3)
			mem.Add(owner.ID, memory.Entry{Text: fmt.Sprintf("%s asked for work and I hired them at %s for %d gold/day.", n.Name, loc.Name, wage), Time: w.TimeString()})
			return fmt.Sprintf("Found work at %s under %s for %d gold/day.", loc.Name, owner.Name, wage)
		},
	},
	{
		ID: "start_business", Label: "Claim an unowned establishment as your own", Category: "employment", BaseGameMinutes: 45,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.IsBusinessOwner || n.GoldCount() < 5 {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.OwnerID != "" {
				return false
			}
			excluded := map[string]bool{"home": true, "forest": true, "farm": true, "well": true}
			return !excluded[loc.Type]
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.OwnerID != "" {
				return "This place already has an owner."
			}
			n.RemoveItem("gold", 5)
			w.Treasury += 5
			loc.OwnerID = n.ID
			n.IsBusinessOwner = true
			n.WorkplaceID = loc.ID
			if n.EmployerID != "" {
				oldEmployer := w.FindNPCByID(n.EmployerID)
				if oldEmployer != nil {
					mem.Add(oldEmployer.ID, memory.Entry{Text: fmt.Sprintf("%s left to start their own business.", n.Name), Time: w.TimeString()})
				}
				n.EmployerID = ""
				n.Wage = 0
			}
			return fmt.Sprintf("Claimed %s as my own business for 5 gold!", loc.Name)
		},
	},
}
