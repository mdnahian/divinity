package action

import (
	"fmt"
	"math/rand"

	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var mountActions = []Action{
	{
		ID: "buy_horse", Label: "Buy Horse", Category: "mount",
		BaseGameMinutes: 30, SkillKey: "",
		Destination: func(n *npc.NPC, w *world.World) string {
			for _, l := range w.Locations {
				if l.Type == "stable" {
					return l.ID
				}
			}
			return n.LocationID
		},
		Candidates: func(n *npc.NPC, w *world.World) []*world.Location {
			return w.LocationsByType("stable")
		},
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.MountID != "" {
				return false // already has a mount
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "stable" {
				return false
			}
			// Must have enough gold for cheapest horse (pony=250)
			return n.GoldCount() >= 250
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			// Pick mount type based on budget
			gold := n.GoldCount()
			var mountType string
			var price int
			if gold >= 800 {
				mountType = "war_horse"
				price = 800
			} else if gold >= 500 {
				mountType = "horse"
				price = 500
			} else {
				mountType = "pony"
				price = 250
			}

			tmpl, ok := world.MountTemplates[mountType]
			if !ok {
				return "No horses available."
			}

			mountID := fmt.Sprintf("mount_%d_%d", rand.Intn(10000), len(w.Mounts))
			mount := &world.Mount{
				ID:          mountID,
				Name:        fmt.Sprintf("%s's %s", n.Name, mountType),
				Type:        tmpl.Type,
				OwnerID:     n.ID,
				LocationID:  n.LocationID,
				HP:          tmpl.HP,
				MaxHP:       tmpl.MaxHP,
				Speed:       tmpl.Speed,
				CarryWeight: tmpl.CarryWeight,
				Hunger:      100,
				Grooming:    100,
				Price:       price,
				Alive:       true,
			}
			w.Mounts = append(w.Mounts, mount)
			n.MountID = mountID
			n.RemoveItem("gold", price)

			// Pay stablehand if present
			for _, s := range w.NPCsAtLocation(n.LocationID, n.ID) {
				if s.Profession == "stablehand" {
					s.AddItem("gold", price/5)
					break
				}
			}

			return fmt.Sprintf("Bought a %s for %d gold.", mountType, price)
		},
	},
	{
		ID: "sell_horse", Label: "Sell Horse", Category: "mount",
		BaseGameMinutes: 15, SkillKey: "",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.MountID == "" {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			return loc != nil && loc.Type == "stable"
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			mount := w.MountByID(n.MountID)
			if mount == nil {
				n.MountID = ""
				return "Had no horse to sell."
			}
			sellPrice := mount.Price * 3 / 4 // 75% of purchase price
			n.AddItem("gold", sellPrice)
			mount.Alive = false
			mount.OwnerID = ""
			n.MountID = ""
			// Unhitch carriage if any
			if n.CarriageID != "" {
				carriage := w.CarriageByID(n.CarriageID)
				if carriage != nil {
					carriage.HorseID = ""
				}
			}
			return fmt.Sprintf("Sold horse for %d gold.", sellPrice)
		},
	},
	{
		ID: "feed_horse", Label: "Feed Horse", Category: "mount",
		BaseGameMinutes: 15, SkillKey: "animal_care",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.MountID == "" {
				return false
			}
			mount := w.MountByID(n.MountID)
			if mount == nil || !mount.Alive {
				return false
			}
			// Needs food: wheat, berries, or hay
			return n.HasItem("wheat") != nil || n.HasItem("berries") != nil || n.HasItem("hay") != nil
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			mount := w.MountByID(n.MountID)
			if mount == nil {
				return "Horse not found."
			}
			// Consume feed
			fed := false
			for _, feed := range []string{"hay", "wheat", "berries"} {
				if n.HasItem(feed) != nil {
					n.RemoveItem(feed, 1)
					mount.Hunger = clampF(mount.Hunger+30, 0, 100)
					fed = true
					break
				}
			}
			if !fed {
				return "Had nothing to feed the horse."
			}
			n.GainSkill("animal_care", 0.5)
			return fmt.Sprintf("Fed %s. Hunger now at %.0f.", mount.Name, mount.Hunger)
		},
	},
	{
		ID: "groom_horse", Label: "Groom Horse", Category: "mount",
		BaseGameMinutes: 30, SkillKey: "animal_care",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.MountID == "" {
				return false
			}
			mount := w.MountByID(n.MountID)
			return mount != nil && mount.Alive && mount.Grooming < 80
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			mount := w.MountByID(n.MountID)
			if mount == nil {
				return "Horse not found."
			}
			mount.Grooming = clampF(mount.Grooming+25, 0, 100)
			if mount.HP < mount.MaxHP {
				mount.HP = min(mount.MaxHP, mount.HP+5)
			}
			n.GainSkill("animal_care", 0.8)
			return fmt.Sprintf("Groomed %s. The horse looks healthier.", mount.Name)
		},
	},
	{
		ID: "buy_carriage", Label: "Buy Carriage", Category: "mount",
		BaseGameMinutes: 30, SkillKey: "",
		Destination: func(n *npc.NPC, w *world.World) string { return n.LocationID },
		Candidates:  func(n *npc.NPC, w *world.World) []*world.Location { return nil },
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.CarriageID != "" || n.MountID == "" {
				return false // already has carriage or no horse
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "stable" {
				return false
			}
			return n.GoldCount() >= 600 // cheapest is cart at 600
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			gold := n.GoldCount()
			var carriageType string
			var tmpl world.Carriage
			if gold >= 1200 {
				carriageType = "wagon"
				tmpl = world.CarriageTemplates["wagon"]
			} else {
				carriageType = "cart"
				tmpl = world.CarriageTemplates["cart"]
			}

			carriageID := fmt.Sprintf("carriage_%d_%d", rand.Intn(10000), len(w.Carriages))
			carriage := &world.Carriage{
				ID:          carriageID,
				Name:        fmt.Sprintf("%s's %s", n.Name, tmpl.Name),
				OwnerID:     n.ID,
				HorseID:     n.MountID,
				LocationID:  n.LocationID,
				CargoSlots:  tmpl.CargoSlots,
				CargoWeight: tmpl.CargoWeight,
				Durability:  100,
				Price:       tmpl.Price,
			}
			w.Carriages = append(w.Carriages, carriage)
			n.CarriageID = carriageID
			n.RemoveItem("gold", tmpl.Price)
			return fmt.Sprintf("Bought a %s for %d gold. Can carry %d item slots.", carriageType, tmpl.Price, tmpl.CargoSlots)
		},
	},
}

func init() {
	Registry = append(Registry, mountActions...)
}
