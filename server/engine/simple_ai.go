package engine

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/divinity/core/action"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

// professionDefaultActions maps profession to a preferred action when idle.
var professionDefaultActions = map[string]string{
	"farmer":          "gather_crops",
	"hunter":          "hunt",
	"blacksmith":      "smelt",
	"carpenter":       "chop_wood",
	"merchant":        "trade",
	"herbalist":       "gather_herbs",
	"healer":          "heal_patient",
	"scribe":          "write_journal",
	"miner":           "mine_ore",
	"fisher":          "fish",
	"tailor":          "craft_clothing",
	"potter":          "craft_pottery",
	"barmaid":         "serve_customer",
	"baker":           "bake_bread",
	"cook":            "cook_meal",
	"innkeeper":       "serve_customer",
	"guard":           "patrol",
	"stablehand":      "groom_horse",
	"carriage_driver": "explore",
	"king":            "hold_court",
	"queen":           "hold_court",
	"duke":            "levy_taxes",
	"count":           "hold_court",
	"baron":           "hold_court",
	"knight":          "patrol",
	"steward":         "trade",
}

// RunTier3Decision makes a simple rule-based decision for a background NPC.
// Returns the chosen action ID and a target name (empty if none).
func RunTier3Decision(n *npc.NPC, w *world.World) (string, string) {
	// Priority-based survival cascade
	if n.HP < 20 {
		if n.HasItem("healing potion") != nil {
			return "use_item", "healing potion"
		}
		if n.HasItem("bread") != nil || n.HasItem("cooked meal") != nil || n.HasItem("berries") != nil {
			return "eat", ""
		}
	}

	// High nobles skip survival cascade — servants handle their needs
	if n.IsHighNoble() {
		if n.Needs.Fatigue > 85 || w.IsNight() {
			return "sleep", ""
		}
		if profAction, ok := professionDefaultActions[n.Profession]; ok {
			return profAction, ""
		}
		return "hold_court", ""
	}

	if n.Needs.Thirst < 15 {
		return "drink", ""
	}

	if n.Needs.Hunger < 20 {
		return "eat", ""
	}

	if n.Needs.Fatigue > 85 {
		return "sleep", ""
	}

	// Nighttime: go home if possible
	if w.IsNight() && n.HomeID != "" && n.LocationID != n.HomeID {
		return "go_home", ""
	}
	if w.IsNight() {
		return "sleep", ""
	}

	// Low hygiene
	if n.Hygiene < 20 {
		return "wash", ""
	}

	// Profession-specific work
	if profAction, ok := professionDefaultActions[n.Profession]; ok {
		a := action.FindAction(profAction)
		if a != nil {
			// Quick check if conditions are met
			func() {
				defer func() { recover() }()
				if a.Conditions(n, w) {
					return
				}
			}()
		}
		// Try the action even if conditions aren't perfectly checked
		return profAction, ""
	}

	// Fallback: explore or socialize
	if n.Needs.SocialNeed > 70 {
		return "talk", ""
	}
	return "explore", ""
}

// ProcessTier3NPCs runs simple rule-based AI for all Tier 3 NPCs.
// This is called from the engine tick to give background NPCs automatic behavior.
func (e *Engine) ProcessTier3NPCs(tier3 []*npc.NPC) {
	w := e.World
	tickNum := e.tickNum

	for _, n := range tier3 {
		// Skip busy NPCs or those currently traveling
		if n.BusyUntilTick > tickNum || n.TargetLocationID != "" {
			continue
		}
		// Skip recently acted NPCs (don't spam actions)
		if n.LastActionTick > 0 && tickNum-n.LastActionTick < 2 {
			continue
		}

		actionID, targetName := RunTier3Decision(n, w)
		if actionID == "" {
			continue
		}

		a := action.FindAction(actionID)
		if a == nil {
			continue
		}

		// Check conditions (with panic recovery since conditions can be complex)
		conditionsMet := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[SimpleAI] Conditions panic for %s action %s: %v", n.Name, actionID, r)
				}
			}()
			conditionsMet = a.Conditions(n, w)
		}()
		if !conditionsMet {
			// Try a simpler fallback
			if actionID != "explore" && actionID != "sleep" {
				a = action.FindAction("explore")
				if a == nil {
					continue
				}
				actionID = "explore"
			} else {
				continue
			}
		}

		// Check if action requires travel
		destID := ""
		if a.Destination != nil {
			destID = a.Destination(n, w)
		}
		if destID != "" && destID != n.LocationID {
			// Initiate travel (mount-aware)
			travelTicks, mounted, inCarriage := w.TravelTicksMounted(n.LocationID, destID, e.Config.Game.GameMinutesPerTick, e.Config.Game.TravelMinutesPerUnit, n.ID)
			if travelTicks > 0 && actionID != "sleep" && actionID != "go_home" {
				// Travel fatigue: walking=3.0, riding=1.0, carriage=0.0 per 10 ticks
				fatiguePer10 := 3.0
				if mounted {
					fatiguePer10 = 1.0
				}
				if inCarriage {
					fatiguePer10 = 0.0
				}
				travelFatigue := fatiguePer10 * float64(travelTicks) / 10.0
				n.Needs.Fatigue = math.Min(100, math.Max(0, n.Needs.Fatigue+travelFatigue))
				n.PreviousLocationID = n.LocationID
				n.TargetLocationID = destID
				n.TravelStartTick = tickNum
				n.TravelArrivalTick = tickNum + travelTicks
			}
			continue
		}

		// Execute action
		ticks, _ := a.Duration(n, e.Config.Game.GameMinutesPerTick)
		n.Busy = true
		n.BusyUntilTick = tickNum + ticks
		n.PendingActionID = actionID
		n.PendingTargetName = targetName
		n.PendingReason = fmt.Sprintf("tier3_auto: %s", actionID)
		n.LastActionTick = tickNum

		// Log briefly (not every action to avoid spam)
		if rand.Intn(20) == 0 {
			log.Printf("[SimpleAI] %s (%s) → %s", n.Name, n.Profession, actionID)
		}
	}
}
