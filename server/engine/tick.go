package engine

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/divinity/core/db"
	"github.com/divinity/core/enemy"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

func (e *Engine) tick(ctx context.Context) {
	e.mu.Lock()
	e.tickNum++
	tickNum := e.tickNum
	e.mu.Unlock()

	w := e.World
	cfg := e.Config

	// Phase 1: fast world-state mutations (hold write lock)
	w.Mu.Lock()

	w.AdvanceTime(cfg.Game.GameMinutesPerTick)

	for i := len(w.ActiveEvents) - 1; i >= 0; i-- {
		w.ActiveEvents[i].TicksLeft--
		if w.ActiveEvents[i].TicksLeft <= 0 {
			w.ActiveEvents = append(w.ActiveEvents[:i], w.ActiveEvents[i+1:]...)
		}
	}

	for _, n := range w.AliveNPCs() {
		n.DecayNeeds(cfg)
		e.dirtyNPCs[n.ID] = true
	}

	// Servants attend to high noble needs
	for _, n := range w.AliveNPCs() {
		if !n.IsHighNoble() {
			continue
		}
		// Water is free — servants fetch it from the well regardless of treasury
		if n.Needs.Thirst < 50 {
			n.Needs.Thirst = 100
		}
		t := w.TerritoryByID(n.TerritoryID)
		if t == nil || t.Treasury <= 0 {
			continue
		}
		if n.Needs.Hunger < 50 {
			n.Needs.Hunger = min(100, n.Needs.Hunger+25)
			t.Treasury--
		}
	}

	// Auto-collapse: NPCs pass out at max fatigue
	for _, n := range w.AliveNPCs() {
		if n.BusyUntilTick > tickNum {
			continue
		}
		if n.Needs.Fatigue >= 100 {
			log.Printf("[AutoRest] %s collapses from exhaustion", n.Name)
			n.Needs.Fatigue -= 15
			if n.Needs.Fatigue < 0 {
				n.Needs.Fatigue = 0
			}
			n.Happiness = max(0, n.Happiness-10)
			n.Stress = min(100, n.Stress+10)
			n.HP = max(0, n.HP-3)
			n.BusyUntilTick = tickNum + 3
			n.PendingActionID = "rest"
			n.PendingReason = "collapsed from exhaustion"
			e.Memory.Add(n.ID, memory.Entry{
				Text:       "I collapsed from exhaustion and passed out. I need to find somewhere safe to sleep.",
				Time:       w.TimeString(),
				Importance: 0.8,
				Category:   memory.CatRoutine,
				Tags:       []string{"exhaustion", "collapse"},
			})
		}
	}

	// Children follow parent (only claimed children under 18)
	for _, n := range w.NPCs {
		if !n.Alive || !n.Claimed || n.ParentID == "" {
			continue
		}
		if n.GetAge(w.GameDay, cfg.Game.GameDaysPerYear) >= 18 {
			continue
		}
		parent := w.FindNPCByID(n.ParentID)
		if parent == nil || !parent.Alive {
			// Try to find the other parent via ChildIDs cross-reference
			var otherParent *npc.NPC
			for _, other := range w.AliveNPCs() {
				for _, cid := range other.ChildIDs {
					if cid == n.ID && other.ID != n.ParentID {
						otherParent = other
						break
					}
				}
				if otherParent != nil {
					break
				}
			}
			if otherParent != nil {
				n.ParentID = otherParent.ID
				n.LocationID = otherParent.LocationID
			} else {
				// Both parents dead — child dies
				log.Printf("[Death] %s (child) dies — no surviving parent", n.Name)
				n.Alive = false
				n.CauseOfDeath = "orphaned — no surviving parent"
				e.onNPCDeath(n, "orphaned — no surviving parent")
			}
			continue
		}
		if n.LocationID != parent.LocationID {
			n.LocationID = parent.LocationID
		}
	}

	for _, n := range w.AliveNPCs() {
		cause := n.CheckDeath(cfg.Game.DeathGraceTicks)
		if cause != "" {
			log.Printf("[Death] %s died: %s | HP: %d | hunger: %.0f | thirst: %.0f | fatigue: %.0f | stress: %d",
				n.Name, cause, n.HP, n.Needs.Hunger, n.Needs.Thirst, n.Needs.Fatigue, n.Stress)
			e.onNPCDeath(n, cause)
		}
	}

	if w.IsNewDay() {
		log.Printf("[Engine] === New day: Day %d === | Pop: %d | Enemies: %d | Weather: %s | Treasury: %d",
			w.GameDay, len(w.AliveNPCs()), len(w.AliveEnemies()), w.Weather, w.Treasury)
		e.lastGodSlot = -1
		e.runDailyTick()
	}

	e.tickConstructions()

	e.runEnemyAttackPhase()

	if tickNum%5 == 0 {
		w.RegenResources()
		w.RegenWellWater()
	}

	alive := w.AliveNPCs()
	if len(alive) == 0 {
		if w.HasClaimedNPCs() {
			w.LogEvent("All villagers have perished. The village is silent.", "death")
			e.Pause()
		}
		w.Mu.Unlock()
		return
	}

	w.Mu.Unlock()

	// Phase 2: Complete finished actions, handle interrupts
	e.World.Mu.Lock()
	e.processCompletionsAndInterrupts()
	e.World.Mu.Unlock()

	// Compute focus territory for GOD
	focusTerritoryID := ""
	if len(w.Territories) > 0 {
		focusIdx := e.focusTerritoryIdx % len(w.Territories)
		// Bump any territory with 3+ crisis NPCs to front
		for i, t := range w.Territories {
			crisisCount := 0
			for _, n := range w.AliveNPCs() {
				if n.TerritoryID == t.ID && (n.HP < 20 || n.Needs.Hunger < 10) {
					crisisCount++
				}
			}
			if crisisCount >= 3 {
				focusIdx = i
				break
			}
		}
		focusTerritoryID = w.Territories[focusIdx].ID
		e.focusTerritoryIdx++
	}

	// Phase 4: GOD turn
	godSlot := (w.GameHour*60 + w.GameMinute) / 205
	if godSlot > e.lastGodSlot {
		e.runGodTurn(ctx, focusTerritoryID)
		e.lastGodSlot = godSlot
	}

	// Phase 5: persist (incremental by default, full-save every 100 ticks)
	if e.DB != nil {
		fullSave := tickNum%100 == 0

		if fullSave {
			w.Mu.RLock()
			err := e.DB.SaveWorld(ctx, w)
			w.Mu.RUnlock()
			if err != nil {
				log.Printf("[Engine] Full save failed: %v", err)
			}
		} else {
			// Incremental: upsert only dirty NPCs and locations
			if len(e.dirtyNPCs) > 0 {
				w.Mu.RLock()
				var dirtyNPCList []*npc.NPC
				for _, n := range w.NPCs {
					if e.dirtyNPCs[n.ID] {
						dirtyNPCList = append(dirtyNPCList, n)
					}
				}
				wid := w.ID
				w.Mu.RUnlock()
				if err := e.DB.UpsertNPCs(ctx, wid, dirtyNPCList); err != nil {
					log.Printf("[Engine] Incremental NPC save failed: %v", err)
				}
			}
			if len(e.dirtyLocations) > 0 {
				w.Mu.RLock()
				var dirtyLocList []*world.Location
				for _, l := range w.Locations {
					if e.dirtyLocations[l.ID] {
						dirtyLocList = append(dirtyLocList, l)
					}
				}
				wid := w.ID
				w.Mu.RUnlock()
				if err := e.DB.UpsertLocations(ctx, wid, dirtyLocList); err != nil {
					log.Printf("[Engine] Incremental location save failed: %v", err)
				}
			}
		}

		// Clear dirty sets
		e.dirtyNPCs = make(map[string]bool)
		e.dirtyLocations = make(map[string]bool)

		if err := e.DB.SaveRelationships(ctx, e.Relationships); err != nil {
			log.Printf("[Engine] Save relationships failed: %v", err)
		}
		if err := e.DB.SaveSharedMemories(ctx, e.SharedMemory); err != nil {
			log.Printf("[Engine] Save shared memories failed: %v", err)
		}
		if err := e.saveTrends(ctx); err != nil {
			log.Printf("[Engine] Save trends failed: %v", err)
		}
	}

	if tickNum%10 == 0 {
		w.Mu.RLock()
		log.Printf("[Engine] Tick %d | Day %d %02d:%02d | Pop: %d | Enemies: %d",
			tickNum, w.GameDay, w.GameHour, w.GameMinute, len(w.AliveNPCs()), len(w.AliveEnemies()))
		w.Mu.RUnlock()
	}

	if e.AfterTick != nil {
		e.AfterTick()
	}
}

func (e *Engine) tickConstructions() {
	w := e.World
	for i := len(w.Constructions) - 1; i >= 0; i-- {
		c := w.Constructions[i]
		builder := w.FindNPCByID(c.OwnerID)
		if builder == nil || !builder.Alive {
			continue
		}
		carpSkill := builder.GetSkillLevel("carpentry")
		if float64(builder.Stats.Carpentry) > carpSkill {
			carpSkill = float64(builder.Stats.Carpentry)
		}
		endMod := float64(builder.Stats.Endurance) / 50.0
		c.Progress += carpSkill * 0.02 * endMod
		builder.GainSkill("carpentry", 0.3)

		if c.Progress >= c.MaxProgress {
			log.Printf("[Construction] %s completed building %s (%s) — progress: %.1f/%.1f",
				builder.Name, c.Name, c.BuildingType, c.Progress, c.MaxProgress)
			loc := w.CompleteConstruction(c)
			w.Constructions = append(w.Constructions[:i], w.Constructions[i+1:]...)
			if c.CommissionerID != "" {
				commissioner := w.FindNPCByID(c.CommissionerID)
				if commissioner != nil && commissioner.Alive && loc != nil {
					loc.OwnerID = commissioner.ID
					commissioner.IsBusinessOwner = true
					commissioner.WorkplaceID = loc.ID
				}
			}
			if loc != nil {
				w.LogEventNPC(fmt.Sprintf("%s finished building a %s: %s!", builder.Name, c.BuildingType, loc.Name), "world", builder.ID)
			}
		}
	}
}

func (e *Engine) runEnemyAttackPhase() {
	w := e.World
	enemies := w.AliveEnemies()
	if len(enemies) == 0 {
		return
	}

	attacked := make(map[string]bool)

	for _, en := range enemies {
		if rand.Float64() > 0.5 {
			continue
		}
		npcsHere := w.NPCsAtLocation(en.LocationID, "")
		var candidates []*npc.NPC
		for _, n := range npcsHere {
			if !attacked[n.ID] {
				candidates = append(candidates, n)
			}
		}
		if len(candidates) == 0 {
			continue
		}

		target := candidates[rand.Intn(len(candidates))]
		attacked[target.ID] = true

		weaponBonus := en.WeaponBonus()
		armorBonus := target.EquippedArmorBonus()
		dmg := enemy.CalculateCombatDamage(en.Strength, en.Agility, weaponBonus, target.Stats.Strength, armorBonus)
		target.HP = max(0, target.HP-dmg)

		loc := w.LocationByID(en.LocationID)
		locName := en.LocationID
		if loc != nil {
			locName = loc.Name
		}

		log.Printf("[Combat] %s (STR:%d AGI:%d wpn:%d HP:%d/%d) attacks %s (STR:%d armor:%d HP:%d→%d) at %s | dmg: %d",
			en.Name, en.Strength, en.Agility, weaponBonus, en.HP, en.MaxHP,
			target.Name, target.Stats.Strength, armorBonus, target.HP+dmg, target.HP,
			locName, dmg)
		w.LogEventNPC(fmt.Sprintf("A %s attacks %s at %s for %d damage! (%s HP: %d)",
			en.Name, target.Name, locName, dmg, target.Name, target.HP), "combat", target.ID)

		e.Memory.Add(target.ID, memory.Entry{
			Text:       fmt.Sprintf("A %s attacked me for %d damage! My HP is now %d. I must fight back, flee, or heal!", en.Name, dmg, target.HP),
			Time:       w.TimeString(),
			Importance: 0.9,
			Category:   memory.CatCombat,
			Tags:       []string{en.Name, en.LocationID},
		})
		target.Stress = clamp(target.Stress+15, 0, 100)

		for _, witness := range npcsHere {
			if witness.ID != target.ID {
				e.Memory.Add(witness.ID, memory.Entry{
					Text:       fmt.Sprintf("I saw a %s attack %s! There is danger here!", en.Name, target.Name),
					Time:       w.TimeString(),
					Importance: 0.6,
					Category:   memory.CatCombat,
					Tags:       []string{en.Name, target.ID, en.LocationID},
				})
			}
		}

		cause := target.CheckDeath(0)
		if cause != "" {
			e.onNPCDeath(target, fmt.Sprintf("%s (killed by %s)", cause, en.Name))
			// Record significant event at this location
			e.SharedMemory.AddLocationMemory(en.LocationID,
				fmt.Sprintf("%s was killed by a %s here.", target.Name, en.Name),
				w.TimeString(), int64(w.GameDay))
			// If the victim was a child, devastating memory for parent
			if target.ParentID != "" {
				if parent := w.FindNPCByID(target.ParentID); parent != nil && parent.Alive {
					e.Memory.Add(parent.ID, memory.Entry{
						Text:       fmt.Sprintf("My child %s was killed by a %s! I am devastated.", target.Name, en.Name),
						Time:       w.TimeString(),
						Importance: 1.0,
						Vividness:  1.0,
						Category:   memory.CatCombat,
						Tags:       []string{"child_death", target.ID, en.Name},
					})
					parent.Happiness = max(0, parent.Happiness-30)
					parent.Stress = min(100, parent.Stress+30)
				}
			}
		}
	}
}

// saveTrends converts WorldSnapshots to DB docs and persists them.
func (e *Engine) saveTrends(ctx context.Context) error {
	snaps := e.Trends.Snapshots()
	docs := make([]db.TrendSnapshotDoc, len(snaps))
	for i, s := range snaps {
		docs[i] = db.TrendSnapshotDoc{
			GameDay:         s.GameDay,
			Population:      s.Population,
			DeadCount:       s.DeadCount,
			TotalGold:       s.TotalGold,
			Treasury:        s.Treasury,
			HungryCount:     s.HungryCount,
			StarvingCount:   s.StarvingCount,
			BrokeCount:      s.BrokeCount,
			EnemyCount:      s.EnemyCount,
			FactionCount:    s.FactionCount,
			DepletedResLocs: s.DepletedResLocs,
			TotalResLocs:    s.TotalResLocs,
			AvgHappiness:    s.AvgHappiness,
			AvgStress:       s.AvgStress,
			CrisisLevel:     s.CrisisLevel,
		}
	}
	return e.DB.SaveTrends(ctx, docs)
}
