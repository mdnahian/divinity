package action

import (
	"fmt"
	"math/rand"

	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var movementActions = []Action{
	{
		ID: "explore", Label: "Wander and explore the area (travel time varies)", Category: "movement",
		Destination: func(n *npc.NPC, w *world.World) string {
			cur := w.LocationByID(n.LocationID)
			var locs []*world.Location
			for _, l := range w.Locations {
				if l.ID != n.LocationID && l.Type != "home" {
					if l.IsNobleRestricted() && !n.CanAccessNobleArea() {
						continue
					}
					// Cap explore to nearby locations (within 60 units ~= 60 min travel)
					if cur != nil {
						dx := (cur.X + cur.W/2) - (l.X + l.W/2)
						dy := (cur.Y + cur.H/2) - (l.Y + l.H/2)
						if dx < 0 {
							dx = -dx
						}
						if dy < 0 {
							dy = -dy
						}
						if dx+dy > 60 {
							continue
						}
					}
					locs = append(locs, l)
				}
			}
			if len(locs) == 0 {
				return ""
			}
			return locs[rand.Intn(len(locs))].ID
		},
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.Needs.Fatigue < 70
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			dest := w.LocationByID(n.LocationID)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+5, 0, 100)
			destName := "the area"
			if dest != nil {
				destName = dest.Name
			}
			if rand.Float64() < 0.2 {
				finds := []string{"odd trinket", "pretty stone", "old coin", "carved bone"}
				item := finds[rand.Intn(len(finds))]
				n.AddItem(item, 1)
				return fmt.Sprintf("Wandered to %s and found a %s.", destName, item)
			}
			return fmt.Sprintf("Wandered to %s, taking in the sights.", destName)
		},
	},
	{
		ID: "travel", Label: "Travel to a different location", Category: "movement",
		Candidates: func(n *npc.NPC, w *world.World) []*world.Location {
			var locs []*world.Location
			for _, l := range w.Locations {
				if l.ID != n.LocationID && l.Type != "home" {
					if l.IsNobleRestricted() && !n.CanAccessNobleArea() {
						continue
					}
					locs = append(locs, l)
				}
			}
			return locs
		},
		Destination: func(n *npc.NPC, w *world.World) string {
			var locs []*world.Location
			for _, l := range w.Locations {
				if l.ID != n.LocationID && l.Type != "home" {
					if l.IsNobleRestricted() && !n.CanAccessNobleArea() {
						continue
					}
					locs = append(locs, l)
				}
			}
			if len(locs) == 0 {
				return ""
			}
			return locs[rand.Intn(len(locs))].ID
		},
		Conditions: func(n *npc.NPC, _ *world.World) bool {
			return n.Needs.Fatigue < 80
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, _ memory.Store) string {
			dest := w.LocationByID(n.LocationID)
			n.Needs.Fatigue = clampF(n.Needs.Fatigue+3, 0, 100)
			destName := "the destination"
			if dest != nil {
				destName = dest.Name
			}
			return fmt.Sprintf("Traveled to %s.", destName)
		},
	},
	{
		ID: "go_home", Label: "Head home", Category: "movement",
		Destination: func(n *npc.NPC, w *world.World) string {
			return n.HomeID
		},
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return n.HomeID != "" && w.IsNight() && n.LocationID != n.HomeID
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, _ *world.World, _ memory.Store) string {
			return "Headed home for the night."
		},
	},
}
