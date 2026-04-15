package engine

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"

	"github.com/divinity/core/action"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

func (e *Engine) processCompletionsAndInterrupts() {
	w := e.World
	currentTick := e.TickCount()
	alive := w.AliveNPCs()

	for _, n := range alive {
		if n.TravelArrivalTick > 0 && n.TravelArrivalTick <= currentTick && n.TargetLocationID != "" {
			loc := w.LocationByID(n.TargetLocationID)
			locName := n.TargetLocationID
			if loc != nil {
				locName = loc.Name
			}
			// Bounce commoners away from noble-restricted areas
			if loc != nil && loc.IsNobleRestricted() && !n.CanAccessNobleArea() {
				log.Printf("[Travel] %s denied entry to %s (noble-restricted)", n.Name, locName)
				n.TargetLocationID = ""
				n.TravelStartTick = 0
				n.TravelArrivalTick = 0
				continue
			}
			log.Printf("[Travel] %s arrived at %s", n.Name, locName)
			n.LocationID = n.TargetLocationID
			n.TravelStartTick = 0
			n.TravelArrivalTick = 0
		}
	}

	for _, n := range alive {
		if n.BusyUntilTick > 0 && n.BusyUntilTick <= currentTick {
			completedID := n.PendingActionID
			targetName := n.PendingTargetName

			loc := w.LocationByID(n.LocationID)
			locName := n.LocationID
			if loc != nil {
				locName = loc.Name
			}
			log.Printf("[Turn] %s completing action: %s | target: %q | location: %s | HP: %d | hunger: %.0f | fatigue: %.0f",
				n.Name, completedID, targetName, locName, n.HP, n.Needs.Hunger, n.Needs.Fatigue)

			result := action.ExecuteAction(completedID, n, w, e.Memory, targetName)
			n.LastAction = completedID
			n.LastActionTick = currentTick

			// Push notification to agent WebSocket
			if e.OnActionComplete != nil {
				go e.OnActionComplete(n.ID, completedID, result)
			}

			memText := result
			if n.LastDialogue != "" {
				memText = fmt.Sprintf("%s Said: \"%s\"", result, n.LastDialogue)
			}
			e.Memory.AddWithCap(n.ID, memory.Entry{
				Text:       memText,
				Time:       w.TimeString(),
				Importance: 0.1,
				Category:   actionMemoryCategory(completedID),
			}, memory.EffectiveCap(n.Stats.MemoryCapacity))

			displayDialogue := ""
			if n.LastDialogue != "" {
				displayDialogue = fmt.Sprintf("  \"%s\"", n.LastDialogue)
			}

			reasonText := ""
			if n.PendingReason != "" {
				reasonText = fmt.Sprintf(" [%s]", n.PendingReason)
			}

			logMsg := fmt.Sprintf("%s: %s%s%s", n.Name, result, displayDialogue, reasonText)
			w.LogEventNPC(logMsg, "npc", n.ID)
			log.Printf("[Turn] %s", logMsg)

			// Update relationship sentiment store for social interactions
			if targetName != "" && e.Relationships != nil {
				target := w.FindNPCByNameAtLocation(targetName, n.LocationID)
				if target != nil && target.ID == n.ID {
					target = w.FindNPCByNameAtLocationExcluding(targetName, n.LocationID, n.ID)
				}
				if target != nil {
					sentimentDelta := actionSentimentDelta(completedID)
					if sentimentDelta != 0 {
						log.Printf("[Relationship] %s → %s: %+.2f (from %s)", n.Name, target.Name, sentimentDelta, completedID)
						e.Relationships.Update(n.ID, target.ID, target.Name, sentimentDelta, completedID, int64(w.GameDay))
						e.Relationships.Update(target.ID, n.ID, n.Name, sentimentDelta*0.5, completedID, int64(w.GameDay))
					}
				}
			}

			if completedID == "sleep" {
				hasDream := HasDreamToday(e.Memory, n.ID, int64(w.GameDay))
				dreamChance := ComputeDreamChance(n)
				roll := rand.Float64()
				canDream := !hasDream && roll < dreamChance

				// Try to deliver a pending divine dream (priority over regular dreams)
				pending := w.PopDivineDreams(n.ID)
				log.Printf("[Dream] %s wakes from sleep | hasDreamToday: %v | dreamChance: %.2f | roll: %.2f | canDream: %v | pendingDivine: %d",
					n.Name, hasDream, dreamChance, roll, canDream, len(pending))
				if canDream && len(pending) > 0 {
					pd := pending[0]
					e.Memory.Add(n.ID, memory.Entry{
						Text:       pd.Text,
						Time:       w.TimeString(),
						TS:         int64(w.GameDay),
						Importance: pd.Importance,
						Vividness:  pd.Vividness,
						Category:   pd.Category,
						Tags:       pd.Tags,
					})
					log.Printf("[DivineDream] %s received divine dream: %s", n.Name, pd.Text)
					// Re-queue remaining divine dreams
					for _, d := range pending[1:] {
						w.QueueDivineDream(d)
					}
					if len(pending) > 1 {
						log.Printf("[DivineDream] %s: %d divine dream(s) re-queued for next sleep", n.Name, len(pending)-1)
					}
				} else {
					// Re-queue all pending divine dreams
					for _, d := range pending {
						w.QueueDivineDream(d)
					}
					if len(pending) > 0 {
						log.Printf("[DivineDream] %s: dream conditions not met, %d divine dream(s) re-queued", n.Name, len(pending))
					}
					// Try regular dream if conditions allow
					if canDream {
						log.Printf("[Dream] %s: generating LLM dream", n.Name)
						go GenerateDream(context.Background(), n, w, e.Router, e.Memory, e.Config)
					}
				}
			}

			if completedID == "pray" {
				pending := w.PopDivineDreams(n.ID)
				if len(pending) > 0 {
					pd := pending[0]
					e.Memory.Add(n.ID, memory.Entry{
						Text:       strings.Replace(pd.Text, "dream", "vision", 1),
						Time:       w.TimeString(),
						TS:         int64(w.GameDay),
						Importance: pd.Importance * 0.8,
						Vividness:  pd.Vividness * 0.8,
						Category:   pd.Category,
						Tags:       pd.Tags,
					})
					log.Printf("[DivineVision] %s received divine vision during prayer: %s", n.Name, pd.Text)
					for _, d := range pending[1:] {
						w.QueueDivineDream(d)
					}
				}
			}

			n.BusyUntilTick = 0
			n.PendingActionID = ""
			n.PendingTargetName = ""
			n.PendingReason = ""
			n.TargetLocationID = ""
			n.PreviousLocationID = ""
		}
	}

	for _, n := range alive {
		if n.BusyUntilTick <= currentTick {
			continue
		}
		enemiesHere := w.EnemiesAtLocation(n.LocationID)
		shouldInterrupt := false
		if len(enemiesHere) > 0 && n.HP < 40 {
			shouldInterrupt = true
		}
		// Only interrupt at very low HP if enemies are present.
		// Previously this fired unconditionally, which trapped
		// safe NPCs in an interrupt loop at low HP — they couldn't
		// eat, heal, or drink because every action was immediately
		// cancelled. Now they can act normally when there's no danger.
		if len(enemiesHere) > 0 && n.HP < 15 {
			shouldInterrupt = true
		}
		if shouldInterrupt {
			log.Printf("[Interrupt] %s interrupted from %s (HP: %d, enemies: %d) — will resume later",
				n.Name, n.PendingActionID, n.HP, len(enemiesHere))
			// Extra penalties for interrupted sleep
			if n.PendingActionID == "sleep" {
				n.Stress = clamp(n.Stress+15, 0, 100)
				n.Happiness = clamp(n.Happiness-10, 0, 100)
				log.Printf("[Interrupt] %s sleep was interrupted! (+15 stress, -10 happiness)", n.Name)
				e.Memory.Add(n.ID, memory.Entry{
					Text:       "My sleep was violently interrupted by danger! I'm exhausted and terrified.",
					Time:       w.TimeString(),
					Importance: 0.8,
					Category:   memory.CatRoutine,
					Tags:       []string{"sleep", "interrupted", "danger"},
				})
			}
			n.ResumeActionID = n.PendingActionID
			n.ResumeTicksLeft = n.BusyUntilTick - currentTick
			n.BusyUntilTick = 0
			n.PendingActionID = ""
			n.PendingTargetName = ""
		}
	}
}

// npcLock returns (or creates) a per-NPC mutex so that agent submissions
// for different NPCs never block each other.
func (e *Engine) npcLock(npcID string) *sync.Mutex {
	e.npcMu.Lock()
	mu, ok := e.npcLocks[npcID]
	if !ok {
		mu = &sync.Mutex{}
		e.npcLocks[npcID] = mu
	}
	e.npcMu.Unlock()
	return mu
}

func (e *Engine) SubmitExternalAction(n *npc.NPC, actionID, target, dialogue, goal, location, reason string, force bool) {
	// Per-NPC lock: two requests for the same NPC serialize, but different NPCs proceed in parallel.
	mu := e.npcLock(n.ID)
	mu.Lock()
	defer mu.Unlock()

	// RLock for world reads — safe because the per-NPC lock prevents concurrent
	// writes to the same NPC, and the tick holds w.Mu.Lock() which waits for us.
	e.World.Mu.RLock()
	defer e.World.Mu.RUnlock()

	currentTick := e.TickCount()

	// Voluntary interruption: agent forced a new action while busy
	if force && n.BusyUntilTick > currentTick {
		log.Printf("[Interrupt] %s voluntarily interrupted %s (%d ticks remaining)",
			n.Name, n.PendingActionID, n.BusyUntilTick-currentTick)

		n.ResumeActionID = n.PendingActionID
		n.ResumeTicksLeft = n.BusyUntilTick - currentTick

		// Cancel travel if mid-travel
		if n.TravelArrivalTick > currentTick && n.TargetLocationID != "" {
			log.Printf("[Interrupt] %s: cancelling travel to %s", n.Name, n.TargetLocationID)
			n.TargetLocationID = ""
			n.TravelStartTick = 0
			n.TravelArrivalTick = 0
			n.PreviousLocationID = ""
		}

		// Clear busy state
		n.BusyUntilTick = 0
		n.PendingActionID = ""
		n.PendingTargetName = ""
		n.PendingReason = ""

		// Mild stress penalty for changing plans
		n.Stress = clamp(n.Stress+5, 0, 100)

		e.Memory.Add(n.ID, memory.Entry{
			Text:       fmt.Sprintf("I stopped %s before finishing to do something more urgent.", n.ResumeActionID),
			Time:       e.World.TimeString(),
			Importance: 0.2,
			Category:   memory.CatRoutine,
			Tags:       []string{"interrupted", "voluntary"},
		})
	}

	if actionID == "resume" && n.ResumeActionID != "" {
		// Validate that the resumed action is still possible
		resumeAct := action.FindAction(n.ResumeActionID)
		if resumeAct != nil && resumeAct.Conditions != nil && !resumeAct.Conditions(n, e.World) {
			log.Printf("[Submit] %s tried to resume %s but conditions no longer met; idling", n.Name, n.ResumeActionID)
			n.ResumeActionID = ""
			n.ResumeTicksLeft = 0
			n.BusyUntilTick = currentTick + 2
			n.PendingActionID = "rest"
			return
		}
		log.Printf("[Submit] %s resuming interrupted action: %s (%d ticks left)", n.Name, n.ResumeActionID, n.ResumeTicksLeft)
		n.BusyUntilTick = currentTick + n.ResumeTicksLeft
		n.PendingActionID = n.ResumeActionID
		n.PendingTargetName = ""
		n.ResumeActionID = ""
		n.ResumeTicksLeft = 0
		return
	}

	n.ResumeActionID = ""
	n.ResumeTicksLeft = 0

	act := action.FindAction(actionID)
	if act != nil && act.Conditions != nil && !act.Conditions(n, e.World) {
		log.Printf("[Submit] %s tried %s but conditions not met; idling", n.Name, actionID)
		n.BusyUntilTick = currentTick + 2
		n.PendingActionID = "rest"
		return
	}

	mpt := e.Config.Game.GameMinutesPerTick
	ticks := 2
	if act != nil {
		ticks, _ = act.Duration(n, mpt)
	}

	travelTicks := 0
	if act != nil {
		destID := ""

		if location != "" && act.Candidates != nil {
			candidates := act.Candidates(n, e.World)
			destID = matchCandidateLocation(location, candidates, n.LocationID)
			if destID == "" {
				log.Printf("[Submit] %s: location %q did not match any candidate for %s", n.Name, location, actionID)
			}
		}

		if destID == "" && act.Destination != nil {
			destID = act.Destination(n, e.World)
		}

		if destID != "" && destID != n.LocationID {
			var mounted, inCarriage bool
			travelTicks, mounted, inCarriage = e.World.TravelTicksMounted(n.LocationID, destID, mpt, e.Config.Game.TravelMinutesPerUnit, n.ID)
			// Travel fatigue: walking=3.0, riding=1.0, carriage=0.0 per 10 ticks
			// Skip travel fatigue for sleep/go_home — the NPC is heading to
			// bed and shouldn't risk collapse from the journey.
			// Weather modifies travel fatigue: rain +30%, storm +60%.
			if travelTicks > 0 && actionID != "sleep" && actionID != "go_home" {
				fatiguePer10 := 3.0
				if mounted {
					fatiguePer10 = 1.0
				}
				if inCarriage {
					fatiguePer10 = 0.0
				}
				travelFatigue := fatiguePer10 * float64(travelTicks) / 10.0 * e.World.WeatherTravelFatigueMod()
				n.Needs.Fatigue = math.Min(100, math.Max(0, n.Needs.Fatigue+travelFatigue))
			}
			n.PreviousLocationID = n.LocationID
			n.TargetLocationID = destID
			n.TravelStartTick = currentTick
			n.TravelArrivalTick = currentTick + travelTicks
			destLoc := e.World.LocationByID(destID)
			destName := destID
			if destLoc != nil {
				destName = destLoc.Name
			}
			modeStr := "walking"
			if inCarriage {
				modeStr = "by carriage"
			} else if mounted {
				modeStr = "riding"
			}
			log.Printf("[Submit] %s: traveling to %s for %s (%d travel ticks, %s)", n.Name, destName, actionID, travelTicks, modeStr)
		}
	}

	log.Printf("[Submit] %s: scheduled %s | target: %q | ticks: %d (travel: %d + action: %d) | busyUntil: %d",
		n.Name, actionID, target, travelTicks+ticks, travelTicks, ticks, currentTick+travelTicks+ticks)

	n.BusyUntilTick = currentTick + travelTicks + ticks
	n.PendingActionID = actionID
	n.PendingTargetName = target
	n.LastDialogue = dialogue
	n.CurrentGoal = goal
	n.PendingReason = reason
}

// actionMemoryCategory maps an action ID to the appropriate memory category
// so that recall_memories category filters return relevant results.
func actionMemoryCategory(actionID string) string {
	switch actionID {
	case "attack_enemy", "flee_area", "party_attack", "fight":
		return memory.CatCombat
	case "trade", "buy_food", "buy_ale", "buy_supplies", "serve_customer",
		"heal_patient", "offer_counsel", "farm", "fish", "hunt",
		"mine_ore", "mine_stone", "chop_wood", "forage",
		"gather_thatch", "gather_clay", "gather_firewood", "scavenge",
		"start_business", "craft_shelter", "craft_fishing_rod",
		"cook_over_fire", "mend_equipment":
		return memory.CatEconomic
	case "talk", "gift", "eat_together", "drink_together", "comfort",
		"flirt", "work_together", "share_journal", "steal":
		return memory.CatSocial
	case "hire_employee", "fire_employee", "quit_job", "seek_employment":
		return memory.CatEmployment
	case "recruit_to_faction", "leave_faction", "kick_member",
		"set_faction_goal", "set_faction_fee", "set_faction_cut":
		return memory.CatFaction
	case "teach", "teach_literacy", "teach_technique", "read_book", "copy_text":
		return memory.CatEducation
	default:
		return memory.CatRoutine
	}
}

// actionSentimentDelta returns the relationship sentiment change for an action.
// Positive for friendly interactions, negative for hostile ones.
func actionSentimentDelta(actionID string) float64 {
	switch actionID {
	case "talk", "eat_together", "drink_together", "work_together", "share_journal":
		return 0.05
	case "gift", "comfort":
		return 0.1
	case "flirt":
		return 0.08
	case "recruit_to_faction":
		return 0.06
	case "trade", "buy_food", "buy_ale", "buy_supplies", "serve_customer":
		return 0.03
	case "heal_patient", "offer_counsel":
		return 0.07
	case "steal":
		return -0.2
	case "fight":
		return -0.3
	default:
		return 0
	}
}

// matchCandidateLocation tries to match a free-text location string to one of the candidate locations
// using exact, normalized, and substring matching. When multiple candidates share the same name,
// it prefers the one at currentLocID (the NPC's current location) to avoid routing to a distant duplicate.
func matchCandidateLocation(input string, candidates []*world.Location, currentLocID string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// If input contains " or ", try each part
	if strings.Contains(strings.ToLower(input), " or ") {
		parts := strings.SplitN(input, " or ", 2)
		for _, part := range parts {
			if id := matchCandidateLocation(strings.TrimSpace(part), candidates, currentLocID); id != "" {
				return id
			}
		}
	}

	// 1. Exact case-insensitive match (after stripping parenthetical suffix)
	cleanLoc := input
	if idx := strings.Index(cleanLoc, " ("); idx >= 0 {
		cleanLoc = strings.TrimSpace(cleanLoc[:idx])
	}
	var exactMatches []*world.Location
	for _, c := range candidates {
		if strings.EqualFold(c.Name, cleanLoc) {
			exactMatches = append(exactMatches, c)
		}
	}
	if id := preferCurrentLocation(exactMatches, currentLocID); id != "" {
		return id
	}

	// 2. Normalized match
	norm := world.NormalizeLocationName(input)
	var normMatches []*world.Location
	for _, c := range candidates {
		if world.NormalizeLocationName(c.Name) == norm {
			normMatches = append(normMatches, c)
		}
	}
	if id := preferCurrentLocation(normMatches, currentLocID); id != "" {
		return id
	}

	// 3. Substring containment
	var subMatches []*world.Location
	for _, c := range candidates {
		cNorm := world.NormalizeLocationName(c.Name)
		if strings.Contains(norm, cNorm) || strings.Contains(cNorm, norm) {
			subMatches = append(subMatches, c)
		}
	}
	if id := preferCurrentLocation(subMatches, currentLocID); id != "" {
		return id
	}

	return ""
}

// preferCurrentLocation picks the best match from a list of candidates.
// If the NPC is already at one of them, return that one (no travel needed).
// Otherwise return the first match.
func preferCurrentLocation(matches []*world.Location, currentLocID string) string {
	if len(matches) == 0 {
		return ""
	}
	for _, m := range matches {
		if m.ID == currentLocID {
			return m.ID
		}
	}
	return matches[0].ID
}
