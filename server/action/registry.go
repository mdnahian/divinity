package action

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

type Action struct {
	ID              string
	Label           string
	Category        string
	BaseGameMinutes int
	SkillKey        string
	Destination     func(n *npc.NPC, w *world.World) string
	Candidates      func(n *npc.NPC, w *world.World) []*world.Location
	Conditions      func(n *npc.NPC, w *world.World) bool
	Execute         func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string
}

func (a *Action) Duration(n *npc.NPC, minutesPerTick int) (ticks int, gameMinutes int) {
	gm := a.BaseGameMinutes
	if gm <= 0 {
		gm = 15
	}
	if a.SkillKey != "" {
		skill := n.GetSkillLevel(a.SkillKey)
		reduction := int(skill/33) * 5
		gm = max(10, gm-reduction)
	}
	ticks = max(1, (gm+minutesPerTick-1)/minutesPerTick)
	return ticks, gm
}

func FindAction(actionID string) *Action {
	for i := range Registry {
		if Registry[i].ID == actionID {
			return &Registry[i]
		}
	}
	return nil
}

var Registry []Action

func init() {
	Registry = append(Registry, survivalActions...)
	Registry = append(Registry, gatherActions...)
	Registry = append(Registry, craftActions...)
	Registry = append(Registry, economyActions...)
	Registry = append(Registry, socialActions...)
	Registry = append(Registry, conflictActions...)
	Registry = append(Registry, combatActions...)
	Registry = append(Registry, wellbeingActions...)
	Registry = append(Registry, educationActions...)
	Registry = append(Registry, constructionActions...)
	Registry = append(Registry, equipmentActions...)
	Registry = append(Registry, movementActions...)
	Registry = append(Registry, employmentActions...)
	Registry = append(Registry, factionActions...)
}

// commonerCategories are action categories that high nobles should never perform.
var commonerCategories = map[string]bool{"gather": true, "craft": true}

// commonerActions are specific action IDs that high nobles should never perform.
var commonerActions = map[string]bool{
	"mine_ore": true, "chop_wood": true, "gather_crops": true, "gather_herbs": true,
	"hunt": true, "fish": true, "smelt": true, "tan_hide": true,
	"bake_bread": true, "cook_meal": true, "serve_customer": true, "build_shelter": true,
}

func isCommonerAction(category, id string) bool {
	return commonerCategories[category] || commonerActions[id]
}

func GetAvailableActions(n *npc.NPC, w *world.World) []Action {
	isHighNoble := n.IsHighNoble()
	var result []Action
	for _, a := range Registry {
		func() {
			defer func() { recover() }()
			if isHighNoble && isCommonerAction(a.Category, a.ID) {
				return
			}
			if !a.Conditions(n, w) {
				return
			}
			if a.Destination != nil {
				if a.Candidates != nil {
					// Multi-destination action: only filter if ALL candidates are full
					candidates := a.Candidates(n, w)
					allFull := len(candidates) > 0
					for _, c := range candidates {
						if !w.IsLocationFull(c.ID) {
							allFull = false
							break
						}
					}
					if allFull {
						return
					}
				}
				// Actions with Destination but no Candidates pick a random
				// location each call. Filtering by capacity makes them vanish
				// ~50% of the time. Let them always appear; handle "full" at
				// execution time instead.
			}
			result = append(result, a)
		}()
	}
	return result
}

var InnFeeExempt = map[string]bool{
	"sleep": true, "rest": true, "go_home": true, "flee_area": true,
	"serve_customer": true, "brew_ale": true, "hire_employee": true, "quit_job": true,
	"set_faction_goal": true, "set_faction_fee": true, "set_faction_cut": true,
	"toggle_external_jobs": true, "post_commission": true, "accept_contract": true,
	"report_contract": true, "approve_contract": true, "reject_contract": true,
	"abandon_contract": true, "leave_faction": true, "kick_member": true,
}

func ExecuteAction(actionID string, n *npc.NPC, w *world.World, mem memory.Store, targetName string) string {
	var action *Action
	for i := range Registry {
		if Registry[i].ID == actionID {
			action = &Registry[i]
			break
		}
	}
	if action == nil {
		return "Stood around, confused (unknown action: " + actionID + ")."
	}

	var target *npc.NPC
	if targetName != "" {
		target = w.FindNPCByName(targetName)
	}

	result := action.Execute(n, target, w, mem)

	inns := w.LocationsByType("inn")
	if len(inns) > 0 {
		inn := inns[0]
		owner := w.GetLocationOwner(inn.ID)
		if owner != nil && n.LocationID == inn.ID && !w.IsWorkerAt(n, inn.ID) && !InnFeeExempt[actionID] {
			guests := len(w.NPCsAtLocation(inn.ID, ""))
			fee := 3 + guests
			if n.GoldCount() >= fee {
				n.RemoveItem("gold", fee)
				if owner.ID != n.ID {
					owner.AddItem("gold", fee)
				} else {
					w.Treasury += fee
				}
				result += fmt.Sprintf(" (paid %d gold inn table fee)", fee)
			}
		}
	}
	return result
}

// innSleepCost returns the gold cost for sleeping overnight at the inn.
func innSleepCost(guests int) int { return 3 + min(guests, 7) }

// nearestAffordableInn returns the ID of the closest non-full inn the NPC can
// afford, or "" if none qualifies.
func nearestAffordableInn(n *npc.NPC, w *world.World) string {
	inns := w.LocationsByType("inn")
	if len(inns) == 0 {
		return ""
	}
	cur := w.LocationByID(n.LocationID)
	type candidate struct {
		id   string
		dist float64
	}
	var candidates []candidate
	for _, inn := range inns {
		if innFull(inn, w) {
			continue
		}
		if w.GetLocationOwner(inn.ID) == nil {
			continue
		}
		guests := len(w.NPCsAtLocation(inn.ID, ""))
		cost := innSleepCost(guests)
		if n.GoldCount() < cost {
			continue
		}
		d := 0.0
		if cur != nil {
			dx := float64(inn.X+inn.W/2) - float64(cur.X+cur.W/2)
			dy := float64(inn.Y+inn.H/2) - float64(cur.Y+cur.H/2)
			d = math.Abs(dx) + math.Abs(dy)
		}
		candidates = append(candidates, candidate{inn.ID, d})
	}
	if len(candidates) == 0 {
		return ""
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.dist < best.dist {
			best = c
		}
	}
	return best.id
}

// innRestCost returns the gold cost for a short rest at the inn.
func innRestCost() int { return 1 }

// innFull returns true if the inn has reached its guest capacity.
func innFull(inn *world.Location, w *world.World) bool {
	if inn.Capacity <= 0 {
		return false
	}
	return len(w.NPCsAtLocation(inn.ID, "")) >= inn.Capacity
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

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func randInt(lo, hi int) int {
	if lo >= hi {
		return lo
	}
	return lo + rand.Intn(hi-lo+1)
}
