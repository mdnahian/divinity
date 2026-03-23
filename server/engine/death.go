package engine

import (
	"fmt"
	"math"

	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

func (e *Engine) onNPCDeath(n *npc.NPC, cause string) {
	w := e.World

	w.DropItemsOnDeath(n)
	w.LogEventNPC(fmt.Sprintf("%s has died of %s.", n.Name, cause), "death", n.ID)

	for _, other := range w.AliveNPCs() {
		rel := other.GetRelationship(n.ID)
		if rel > 10 {
			scale := math.Min(1, float64(rel)/50)
			other.Stress = clamp(other.Stress+int(math.Round(15*scale)), 0, 100)
			other.Happiness = clamp(other.Happiness-int(math.Round(10*scale)), 0, 100)
			other.Stats.Trauma = clamp(other.Stats.Trauma+int(math.Round(5*scale)), 0, 100)
			e.Memory.Add(other.ID, memory.Entry{
				Text: fmt.Sprintf("%s has died. I feel grief.", n.Name),
				Time: w.TimeString(),
			})
		}
	}

	if len(n.Apprentices) > 0 {
		for _, appID := range n.Apprentices {
			app := w.FindNPCByID(appID)
			if app == nil || !app.Alive {
				continue
			}
			for skill, val := range n.Skills {
				current := app.GetSkillLevel(skill)
				if val > current {
					app.Skills[skill] = math.Min(100, current+(val-current)*0.5)
				}
			}
			app.ApprenticeTo = ""
		}
	}

	for _, loc := range w.Locations {
		if loc.BuildingOwnerID == n.ID {
			spouse := findAliveWith(w, func(o *npc.NPC) bool {
				return o.GetRelationship(n.ID) > 50
			})
			factionMate := findAliveWith(w, func(o *npc.NPC) bool {
				return o.FactionID != "" && o.FactionID == n.FactionID
			})
			heir := spouse
			if heir == nil {
				heir = factionMate
			}
			if heir != nil {
				loc.BuildingOwnerID = heir.ID
				w.LogEventNPC(fmt.Sprintf("%s inherits %s from %s.", heir.Name, loc.Name, n.Name), "world", heir.ID)
			} else {
				loc.BuildingOwnerID = ""
			}
		}
	}

	checkTechniqueLoss(n.ID, w, e)

	for _, emp := range w.AliveNPCs() {
		if emp.EmployerID == n.ID {
			emp.EmployerID = ""
			emp.WorkplaceID = ""
			emp.Wage = 0
			e.Memory.Add(emp.ID, memory.Entry{
				Text: fmt.Sprintf("My employer %s has died. I lost my job.", n.Name),
				Time: w.TimeString(),
			})
		}
	}

	for _, loc := range w.Locations {
		if loc.OwnerID == n.ID {
			heir := findAliveWith(w, func(o *npc.NPC) bool {
				return o.EmployerID == n.ID
			})
			if heir == nil {
				heir = findAliveWith(w, func(o *npc.NPC) bool {
					return o.GetRelationship(n.ID) > 50
				})
			}
			if heir != nil {
				loc.OwnerID = heir.ID
				heir.IsBusinessOwner = true
				heir.WorkplaceID = loc.ID
				w.LogEventNPC(fmt.Sprintf("%s takes over %s after %s's passing.", heir.Name, loc.Name, n.Name), "world", heir.ID)
			} else {
				loc.OwnerID = ""
			}
		}
	}
}

func checkTechniqueLoss(deadID string, w *world.World, e *Engine) {
	for i := len(w.Techniques) - 1; i >= 0; i-- {
		tech := w.Techniques[i]
		filtered := make([]string, 0, len(tech.KnownBy))
		for _, id := range tech.KnownBy {
			if id != deadID {
				filtered = append(filtered, id)
			}
		}
		tech.KnownBy = filtered

		for j := range tech.WrittenIn {
			if tech.WrittenIn[j].AuthorID == deadID {
				tech.WrittenIn[j].AuthorAlive = false
			}
		}

		if len(tech.KnownBy) == 0 && len(tech.WrittenIn) == 0 {
			w.LogEvent(fmt.Sprintf("The technique \"%s\" has been lost forever!", tech.Name), "world")
			w.Techniques = append(w.Techniques[:i], w.Techniques[i+1:]...)
		}
	}
}

func findAliveWith(w *world.World, pred func(*npc.NPC) bool) *npc.NPC {
	for _, n := range w.AliveNPCs() {
		if pred(n) {
			return n
		}
	}
	return nil
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
