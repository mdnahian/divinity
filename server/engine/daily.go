package engine

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/divinity/core/faction"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

func (e *Engine) runDailyTick() {
	w := e.World
	cfg := e.Config

	for _, n := range w.AliveNPCs() {
		n.DailyHygieneSobriety()
		n.DecayInventoryDurability()
		applyHomeTierBuffs(n, w)
		// Apply memory vividness decay
		e.Memory.ApplyDecay(n.ID)
	}
	// Also decay GOD memories
	e.Memory.ApplyDecay(memory.GodEntityID)
	// Decay relationship sentiment toward neutral over time
	e.Relationships.DecayAll(int64(w.GameDay), 0.01)
	// Decay location memories
	e.SharedMemory.DecayLocationMemories()
	// Snapshot world trends for GOD strategic analysis
	e.Trends.Snapshot(w)

	for _, n := range w.AliveNPCs() {
		cause := n.CheckOldAge(w.GameDay, cfg.Game.GameDaysPerYear)
		if cause != "" {
			e.onNPCDeath(n, cause)
		}
	}

	// Periodic reflection: every 3 days, NPCs with enough memories generate insights
	if w.GameDay%3 == 0 {
		for _, n := range w.AliveNPCs() {
			go GenerateReflection(context.Background(), n, w, e.Router, e.Memory, cfg)
		}
	}

	e.checkReproduction()
	e.checkComingOfAge()

	events := updateLeadershipAndFactions(w, cfg.Game.GameDaysPerYear)
	for _, evt := range events {
		w.LogEvent(evt.Text, evt.Type)
	}

	w.DecayGroundItems()
	w.DecayEconomy()
	w.RegenResources()
	w.TickTraps() // Process hunting trap catches
	e.payWages()
	e.collectFactionFees()
	e.expireFactionContracts()

	collapsed := w.DecayBuildings()
	if collapsed != nil {
		w.LogEvent(fmt.Sprintf("%s has collapsed from neglect!", collapsed.Name), "world")
	}

	// Decay mount hunger/grooming, handle starvation
	w.DecayMounts()

	// Auto-feed stabled horses (stablehand bonus)
	for _, m := range w.Mounts {
		if !m.Alive || m.Hunger > 50 {
			continue
		}
		loc := w.LocationByID(m.LocationID)
		if loc == nil || loc.Type != "stable" {
			continue
		}
		// Check if stable has hay
		if loc.Resources != nil && loc.Resources["hay"] > 0 {
			loc.Resources["hay"]--
			m.Hunger = m.Hunger + 20
			if m.Hunger > 100 {
				m.Hunger = 100
			}
		}
	}

	// Rebuild social graph and compute metrics for GOD AI
	e.SocialGraph.Rebuild(e.World, e.Relationships)
	e.SocialGraph.ComputeMetrics()

	// Compute territory intelligence briefs for GOD AI
	e.TerritoryBriefs = e.computeTerritoryBriefs()

	// Classify NPCs into tiers for GOD tool context
	pop := len(w.AliveNPCs())
	tier1Count := max(3, pop/5)
	tier2Count := max(5, pop/3)
	t1, t2, t3 := ClassifyNPCs(w, tier1Count, tier2Count)
	e.NPCTiers = make(map[string]NPCTier, pop)
	for _, n := range t1 {
		e.NPCTiers[n.ID] = Tier1FullLLM
	}
	for _, n := range t2 {
		e.NPCTiers[n.ID] = Tier2BatchLLM
	}
	for _, n := range t3 {
		e.NPCTiers[n.ID] = Tier3RuleBased
	}
	log.Printf("[Engine] NPC tiering: %d tier1, %d tier2, %d tier3", len(t1), len(t2), len(t3))

	go e.generateDailyChronicle(context.Background())
}

func applyHomeTierBuffs(n *npc.NPC, w *world.World) {
	if !n.Alive {
		return
	}
	tier := 0
	if n.HomeBuildingID != "" {
		for _, loc := range w.Locations {
			if loc.BuildingID == n.HomeBuildingID || loc.ID == n.HomeBuildingID {
				tier = loc.GetTier()
				break
			}
		}
	}
	switch {
	case tier == 0:
		n.Stress = clamp(n.Stress+5, 0, 100)
	case tier == 1:
		n.Stress = clamp(n.Stress+2, 0, 100)
	case tier == 2:
		n.Stress = clamp(n.Stress-5, 0, 100)
		n.Stats.DiseaseResistance = clamp(n.Stats.DiseaseResistance+1, 0, 100)
	case tier == 3:
		n.Stress = clamp(n.Stress-10, 0, 100)
		n.Happiness = clamp(n.Happiness+3, 0, 100)
	case tier >= 4:
		n.Stress = clamp(n.Stress-15, 0, 100)
		n.Happiness = clamp(n.Happiness+5, 0, 100)
		n.Stats.Reputation = clamp(n.Stats.Reputation+1, 0, 100)
	}
}

func (e *Engine) payWages() {
	w := e.World
	alive := w.AliveNPCs()

	for _, n := range alive {
		if n.EmployerID == "" || n.Wage == 0 {
			continue
		}
		employer := w.FindNPCByID(n.EmployerID)
		if employer == nil || !employer.Alive {
			n.EmployerID = ""
			n.WorkplaceID = ""
			n.Wage = 0
			e.Memory.Add(n.ID, memory.Entry{
				Text: "My employer is gone. I am now unemployed.",
				Time: w.TimeString(),
				Importance: 0.6,
				Category:   memory.CatEmployment,
			})
			continue
		}
		gold := employer.GoldCount()
		if gold >= n.Wage {
			employer.RemoveItem("gold", n.Wage)
			n.AddItem("gold", n.Wage)
		} else if gold > 0 {
			// Part B: Partial wage payment
			employer.RemoveItem("gold", gold)
			n.AddItem("gold", gold)
			n.UnpaidDays++
			n.Stress = clamp(n.Stress+3, 0, 100)
			e.Memory.Add(n.ID, memory.Entry{
				Text: fmt.Sprintf("%s could only pay %d of %d gold wages (%d days partially unpaid).", employer.Name, gold, n.Wage, n.UnpaidDays),
				Time: w.TimeString(),
				Importance: 0.4,
				Category:   memory.CatEmployment,
				Tags:       []string{employer.ID},
			})
			if n.UnpaidDays >= 3 {
				n.AdjustRelationship(employer.ID, -10)
			}
		} else {
			n.UnpaidDays++
			n.Stress = clamp(n.Stress+5, 0, 100)
			e.Memory.Add(n.ID, memory.Entry{
				Text: fmt.Sprintf("%s failed to pay my wages (%d days unpaid).", employer.Name, n.UnpaidDays),
				Time: w.TimeString(),
				Importance: 0.5,
				Category:   memory.CatEmployment,
				Tags:       []string{employer.ID},
			})
			if n.UnpaidDays >= 3 {
				n.AdjustRelationship(employer.ID, -10)
			}
		}
	}
}

func (e *Engine) checkReproduction() {
	w := e.World
	cfg := e.Config
	alive := w.AliveNPCs()
	if len(alive) >= 30 {
		return
	}

	for i := 0; i < len(alive); i++ {
		a := alive[i]
		aAge := a.GetAge(w.GameDay, cfg.Game.GameDaysPerYear)
		if aAge < 18 || aAge > 55 || a.Fertility < 30 {
			continue
		}
		for j := i + 1; j < len(alive); j++ {
			b := alive[j]
			bAge := b.GetAge(w.GameDay, cfg.Game.GameDaysPerYear)
			if bAge < 18 || bAge > 55 || b.Fertility < 30 {
				continue
			}
			if a.LocationID != b.LocationID {
				continue
			}
			if a.GetRelationship(b.ID) < 30 || b.GetRelationship(a.ID) < 20 {
				continue
			}
			chance := float64(a.Fertility+b.Fertility) / 2 * 0.003
			if rand.Float64() > chance {
				continue
			}

			child := npc.CreateChild(a, b, w.GameDay, cfg)
			child.LocationID = a.LocationID
			w.NPCs = append(w.NPCs, child)

			a.ChildIDs = append(a.ChildIDs, child.ID)
			b.ChildIDs = append(b.ChildIDs, child.ID)
			a.AdjustRelationship(b.ID, 10)
			b.AdjustRelationship(a.ID, 10)

			e.Memory.Add(a.ID, memory.Entry{
				Text:       fmt.Sprintf("I have a new child: %s! I must protect and provide for them.", child.Name),
				Time:       w.TimeString(),
				Importance: 0.95,
				Category:   memory.CatSocial,
				Tags:       []string{"child", child.ID},
			})
			e.Memory.Add(b.ID, memory.Entry{
				Text:       fmt.Sprintf("I have a new child: %s! I must protect and provide for them.", child.Name),
				Time:       w.TimeString(),
				Importance: 0.95,
				Category:   memory.CatSocial,
				Tags:       []string{"child", child.ID},
			})

			w.LogEventNPC(fmt.Sprintf("%s and %s welcome a new child: %s!", a.Name, b.Name, child.Name), "birth", child.ID)
			log.Printf("[Birth] %s and %s have a child: %s", a.Name, b.Name, child.Name)
			return
		}
	}
}

func (e *Engine) checkComingOfAge() {
	w := e.World
	cfg := e.Config
	for _, n := range w.NPCs {
		if !n.Alive || !n.Claimed || n.ParentID == "" {
			continue
		}
		age := n.GetAge(w.GameDay, cfg.Game.GameDaysPerYear)
		if age >= 18 {
			parentID := n.ParentID
			n.ParentID = ""
			w.SpawnQueue = append(w.SpawnQueue, world.SpawnEntry{NPCID: n.ID})
			w.LogEventNPC(fmt.Sprintf("%s has come of age and is ready to make their own way!", n.Name), "coming_of_age", n.ID)
			log.Printf("[ComingOfAge] %s (age %d) added to spawn queue", n.Name, age)
			// Notify parent
			if parent := w.FindNPCByID(parentID); parent != nil && parent.Alive {
				e.Memory.Add(parent.ID, memory.Entry{
					Text:       fmt.Sprintf("My child %s has grown up and is now independent!", n.Name),
					Time:       w.TimeString(),
					Importance: 0.8,
					Category:   memory.CatSocial,
					Tags:       []string{"child", n.ID, "coming_of_age"},
				})
			}
		}
	}
}

type factionEvent struct {
	Text string
	Type string
}

func computeLeadershipScore(n *npc.NPC, worldDay, daysPerYear int) float64 {
	s := &n.Stats
	return float64(s.Dominance)*0.25 +
		float64(s.Reputation)*0.25 +
		float64(s.Charisma)*0.2 +
		float64(n.GetEffectiveWisdom(worldDay, daysPerYear))*0.15 +
		float64(s.Persuasion)*0.15
}

func determineFactionType(leader *npc.NPC) string {
	best := "political"
	bestVal := 0
	for ftype, def := range faction.FactionTypes {
		val := leader.Stats.ByKey(def.StatKey)
		if val > bestVal && val >= def.Threshold {
			best = ftype
			bestVal = val
		}
	}
	return best
}

func updateLeadershipAndFactions(w *world.World, daysPerYear int) []factionEvent {
	alive := w.AliveNPCs()
	events := make([]factionEvent, 0)

	for _, n := range alive {
		stage := n.GetLifeStage(w.GameDay, daysPerYear)
		if stage == "infant" || stage == "child" {
			n.LeadershipScore = 0
			continue
		}
		n.LeadershipScore = computeLeadershipScore(n, w.GameDay, daysPerYear)
	}

	for i := len(w.Factions) - 1; i >= 0; i-- {
		f := w.Factions[i]
		var leader *npc.NPC
		for _, n := range alive {
			if n.ID == f.LeaderID {
				leader = n
				break
			}
		}

		if leader == nil {
			members := make([]*npc.NPC, 0)
			for _, n := range alive {
				if n.FactionID == f.ID {
					members = append(members, n)
				}
			}
			var newLeader *npc.NPC
			bestScore := 0.0
			for _, m := range members {
				if m.LeadershipScore > bestScore {
					bestScore = m.LeadershipScore
					newLeader = m
				}
			}
			if newLeader != nil && bestScore > 40 {
				f.LeaderID = newLeader.ID
				f.LeaderName = newLeader.Name
				events = append(events, factionEvent{
					Text: fmt.Sprintf("%s takes over leadership of %s.", newLeader.Name, f.Name),
					Type: "faction",
				})
			} else {
				for _, m := range members {
					m.FactionID = ""
				}
				w.Factions = append(w.Factions[:i], w.Factions[i+1:]...)
				events = append(events, factionEvent{
					Text: fmt.Sprintf("%s has dissolved.", f.Name),
					Type: "faction",
				})
			}
			continue
		}

		memberIDs := make([]string, 0)
		for _, n := range alive {
			if n.FactionID == f.ID {
				memberIDs = append(memberIDs, n.ID)
			}
		}
		f.MemberIDs = memberIDs

		if len(f.MemberIDs) < 2 {
			leader.FactionID = ""
			w.Factions = append(w.Factions[:i], w.Factions[i+1:]...)
			events = append(events, factionEvent{
				Text: fmt.Sprintf("%s has disbanded — too few members remain.", f.Name),
				Type: "faction",
			})
		}
	}

	for _, n := range alive {
		if n.FactionID != "" {
			continue
		}
		if n.LeadershipScore < 55 {
			continue
		}
		stage := n.GetLifeStage(w.GameDay, daysPerYear)
		if stage == "adolescent" {
			continue
		}

		positiveRels := make([]*npc.NPC, 0)
		for _, other := range alive {
			if other.ID == n.ID || other.FactionID != "" {
				continue
			}
			if n.GetRelationship(other.ID) > 15 && other.GetRelationship(n.ID) > 5 {
				positiveRels = append(positiveRels, other)
			}
		}
		if len(positiveRels) < 2 {
			continue
		}

		ftype := determineFactionType(n)
		members := positiveRels
		if len(members) > 4 {
			members = members[:4]
		}

		memberIDs := make([]string, 0, len(members)+1)
		memberIDs = append(memberIDs, n.ID)
		for _, m := range members {
			memberIDs = append(memberIDs, m.ID)
		}

		f := faction.NewFaction(ftype, n.ID, n.Name, memberIDs, w.GameDay)
		n.FactionID = f.ID
		for _, m := range members {
			m.FactionID = f.ID
		}
		w.Factions = append(w.Factions, f)

		fdef := faction.FactionTypes[ftype]
		events = append(events, factionEvent{
			Text: fmt.Sprintf("%s founds \"%s\" (%s) with %d followers.",
				n.Name, f.Name, fdef.Label, len(members)),
			Type: "faction",
		})
	}

	return events
}

func (e *Engine) collectFactionFees() {
	w := e.World
	for _, f := range w.Factions {
		if f.MembershipFee <= 0 {
			continue
		}
		for _, mid := range f.MemberIDs {
			if mid == f.LeaderID {
				continue
			}
			member := w.FindNPCByID(mid)
			if member == nil || !member.Alive {
				continue
			}
			if member.GoldCount() >= f.MembershipFee {
				member.RemoveItem("gold", f.MembershipFee)
				f.Treasury += f.MembershipFee
			} else {
				member.Stress = clamp(member.Stress+10, 0, 100)
				e.Memory.Add(member.ID, memory.Entry{
					Text: fmt.Sprintf("Could not pay the %d gold faction membership fee for \"%s\" today.", f.MembershipFee, f.Name),
					Time: w.TimeString(),
					Importance: 0.4,
					Category:   memory.CatFaction,
					Tags:       []string{f.Name},
				})
			}
		}
	}
}

func (e *Engine) expireFactionContracts() {
	w := e.World
	for _, c := range w.FactionContracts {
		if c.Status == "completed" || c.Status == "expired" || c.Status == "abandoned" {
			continue
		}
		shouldExpire := w.GameDay >= c.DueDay || (c.RejectionCount >= 3 && c.Status != "open")
		if !shouldExpire {
			continue
		}
		c.Status = "expired"

		// Return escrow to requester
		requester := w.FindNPCByID(c.RequesterID)
		if requester != nil {
			requester.AddItem("gold", c.EscrowGold)
		}

		if c.WorkerID == "" {
			// Nobody took the job — just log
			w.LogEvent(fmt.Sprintf("Faction contract \"%s\" posted by %s expired unfulfilled.", c.Description, c.RequesterName), "faction")
			continue
		}

		// Worker failed — apply penalties
		worker := w.FindNPCByID(c.WorkerID)
		if worker != nil {
			worker.Stats.Reputation = clamp(worker.Stats.Reputation-15, 0, 100)
			worker.Stress = clamp(worker.Stress+20, 0, 100)
			e.Memory.Add(worker.ID, memory.Entry{
				Text: fmt.Sprintf("My faction contract for %s (\"%s\") has expired. I failed to complete it in time and my reputation has suffered.", c.RequesterName, c.Description),
				Time: w.TimeString(),
				Importance: 0.6,
				Category:   memory.CatFaction,
				Tags:       []string{c.RequesterID},
			})
		}

		// Notify all faction members
		var fac *faction.Faction
		for _, f := range w.Factions {
			if f.ID == c.FactionID {
				fac = f
				break
			}
		}
		if fac != nil {
			failMsg := fmt.Sprintf("%s failed to complete the faction contract for %s: \"%s\".", c.WorkerName, c.RequesterName, c.Description)
			for _, mid := range fac.MemberIDs {
				if mid != c.WorkerID {
					e.Memory.Add(mid, memory.Entry{Text: failMsg, Time: w.TimeString(), Importance: 0.4, Category: memory.CatFaction, Tags: []string{c.WorkerID, c.RequesterID}})
				}
			}
			if fac.LeaderID != c.WorkerID {
				e.Memory.Add(fac.LeaderID, memory.Entry{Text: failMsg, Time: w.TimeString(), Importance: 0.4, Category: memory.CatFaction, Tags: []string{c.WorkerID, c.RequesterID}})
			}
		}
		w.LogEvent(fmt.Sprintf("Faction contract \"%s\" expired — %s failed to deliver for %s.", c.Description, c.WorkerName, c.RequesterName), "faction")
	}
}
