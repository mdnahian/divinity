package action

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/divinity/core/faction"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

// findNPCFaction returns the faction the given NPC belongs to (as leader or member).
func findNPCFaction(n *npc.NPC, w *world.World) *faction.Faction {
	for _, f := range w.Factions {
		if f.LeaderID == n.ID {
			return f
		}
		for _, mid := range f.MemberIDs {
			if mid == n.ID {
				return f
			}
		}
	}
	return nil
}

// findNPCContractAsWorker returns an accepted/pending_review contract where NPC is the worker.
func findNPCContractAsWorker(n *npc.NPC, w *world.World) *faction.FactionContract {
	for _, c := range w.FactionContracts {
		if c.WorkerID == n.ID && (c.Status == "accepted" || c.Status == "pending_review") {
			return c
		}
	}
	return nil
}

// findNPCContractAsRequester returns a pending_review contract where NPC is the requester.
func findNPCContractAsRequester(n *npc.NPC, w *world.World) *faction.FactionContract {
	for _, c := range w.FactionContracts {
		if c.RequesterID == n.ID && c.Status == "pending_review" {
			return c
		}
	}
	return nil
}

// parseIntFromDialogue parses the first integer from dialogue text, returning defaultVal if none found.
func parseIntFromDialogue(dialogue string, defaultVal int) int {
	for _, word := range strings.Fields(dialogue) {
		word = strings.Trim(word, ".,!?;:")
		if v, err := strconv.Atoi(word); err == nil && v > 0 {
			return v
		}
	}
	return defaultVal
}

var factionActions = []Action{
	{
		ID: "set_faction_goal", Label: "Set your faction's goal (state it in your dialogue)", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			for _, f := range w.Factions {
				if f.LeaderID == n.ID {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			var fac *faction.Faction
			for _, f := range w.Factions {
				if f.LeaderID == n.ID {
					fac = f
					break
				}
			}
			if fac == nil {
				return "No faction to lead."
			}
			goal := strings.TrimSpace(n.LastDialogue)
			if goal == "" {
				return "Spoke vaguely about the faction's purpose but said nothing concrete."
			}
			fac.Goal = goal
			msg := fmt.Sprintf("%s has declared the faction goal: \"%s\"", n.Name, goal)
			w.LogEvent(msg, "faction")
			for _, mid := range fac.MemberIDs {
				if mid != n.ID {
					mem.Add(mid, memory.Entry{Text: fmt.Sprintf("Our leader %s has set a new faction goal: \"%s\"", n.Name, goal), Time: w.TimeString(), Importance: 0.4, Category: memory.CatFaction, Tags: []string{n.ID, fac.Name}})
				}
			}
			return fmt.Sprintf("Declared faction goal: \"%s\"", goal)
		},
	},
	{
		ID: "set_faction_fee", Label: "Set the faction membership fee (state amount in gold in your dialogue)", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			for _, f := range w.Factions {
				if f.LeaderID == n.ID {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			var fac *faction.Faction
			for _, f := range w.Factions {
				if f.LeaderID == n.ID {
					fac = f
					break
				}
			}
			if fac == nil {
				return "No faction to lead."
			}
			fee := parseIntFromDialogue(n.LastDialogue, -1)
			if fee < 0 || fee > 5 {
				fee = clamp(fee, 0, 5)
			}
			fac.MembershipFee = fee
			msg := fmt.Sprintf("%s has set the faction membership fee to %d gold/day.", n.Name, fee)
			w.LogEvent(msg, "faction")
			for _, mid := range fac.MemberIDs {
				if mid != n.ID {
					mem.Add(mid, memory.Entry{Text: fmt.Sprintf("Our leader %s changed the faction membership fee to %d gold/day.", n.Name, fee), Time: w.TimeString(), Importance: 0.4, Category: memory.CatFaction, Tags: []string{n.ID, fac.Name}})
				}
			}
			return fmt.Sprintf("Set faction membership fee to %d gold/day.", fee)
		},
	},
	{
		ID: "set_faction_cut", Label: "Set the faction's contract cut percentage (0-30, state in dialogue)", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			for _, f := range w.Factions {
				if f.LeaderID == n.ID {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			var fac *faction.Faction
			for _, f := range w.Factions {
				if f.LeaderID == n.ID {
					fac = f
					break
				}
			}
			if fac == nil {
				return "No faction to lead."
			}
			cut := parseIntFromDialogue(n.LastDialogue, fac.FactionCutPct)
			cut = clamp(cut, 0, 30)
			fac.FactionCutPct = cut
			msg := fmt.Sprintf("%s has set the faction contract cut to %d%%.", n.Name, cut)
			w.LogEvent(msg, "faction")
			for _, mid := range fac.MemberIDs {
				if mid != n.ID {
					mem.Add(mid, memory.Entry{Text: fmt.Sprintf("Our leader %s changed the faction contract cut to %d%%.", n.Name, cut), Time: w.TimeString(), Importance: 0.4, Category: memory.CatFaction, Tags: []string{n.ID, fac.Name}})
				}
			}
			return fmt.Sprintf("Set faction contract cut to %d%%.", cut)
		},
	},
	{
		ID: "toggle_external_jobs", Label: "Toggle whether outsiders can post contracts to your faction", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			for _, f := range w.Factions {
				if f.LeaderID == n.ID {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			var fac *faction.Faction
			for _, f := range w.Factions {
				if f.LeaderID == n.ID {
					fac = f
					break
				}
			}
			if fac == nil {
				return "No faction to lead."
			}
			fac.AllowExternalContracts = !fac.AllowExternalContracts
			state := "CLOSED"
			if fac.AllowExternalContracts {
				state = "OPEN"
			}
			msg := fmt.Sprintf("%s has set \"%s\" external jobs to %s.", n.Name, fac.Name, state)
			w.LogEvent(msg, "faction")
			for _, mid := range fac.MemberIDs {
				if mid != n.ID {
					mem.Add(mid, memory.Entry{Text: fmt.Sprintf("Our leader %s has set external jobs to %s — outsiders %s hire us.", n.Name, state, map[bool]string{true: "CAN", false: "CANNOT"}[fac.AllowExternalContracts]), Time: w.TimeString(), Importance: 0.4, Category: memory.CatFaction, Tags: []string{n.ID, fac.Name}})
				}
			}
			return fmt.Sprintf("External jobs are now %s for \"%s\".", state, fac.Name)
		},
	},
	{
		// post_commission: NPC encodes "Description | FulfillmentCondition | payment" in dialogue.
		// e.g. "Bring me iron ore | 5 iron ore | 15"
		// FactionCut is computed from faction's FactionCutPct.
		ID: "post_commission", Label: "Post a faction contract (dialogue: \"job description | fulfillment condition | payment gold\")", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if findNPCContractAsWorker(n, w) != nil {
				return false
			}
			for _, f := range w.Factions {
				isMember := f.LeaderID == n.ID
				if !isMember {
					for _, mid := range f.MemberIDs {
						if mid == n.ID {
							isMember = true
							break
						}
					}
				}
				if isMember || f.AllowExternalContracts {
					// Need at least some gold
					if n.GoldCount() >= 1 {
						return true
					}
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			// Find target faction: member's own faction first, or any external-open faction
			var fac *faction.Faction
			for _, f := range w.Factions {
				isMember := f.LeaderID == n.ID
				if !isMember {
					for _, mid := range f.MemberIDs {
						if mid == n.ID {
							isMember = true
							break
						}
					}
				}
				if isMember {
					fac = f
					break
				}
			}
			if fac == nil {
				// outsider — find first open faction
				for _, f := range w.Factions {
					if f.AllowExternalContracts {
						fac = f
						break
					}
				}
			}
			if fac == nil {
				return "No faction available to post a contract to."
			}

			// Parse dialogue: "description | condition | payment"
			parts := strings.SplitN(n.LastDialogue, "|", 3)
			description := "Unspecified job"
			condition := "Unspecified condition"
			paymentRaw := 5
			if len(parts) >= 1 {
				description = strings.TrimSpace(parts[0])
			}
			if len(parts) >= 2 {
				condition = strings.TrimSpace(parts[1])
			}
			if len(parts) >= 3 {
				paymentRaw = parseIntFromDialogue(parts[2], 5)
			}

			cutAmt := paymentRaw * fac.FactionCutPct / 100
			escrow := paymentRaw + cutAmt

			if n.GoldCount() < escrow {
				return fmt.Sprintf("Not enough gold — need %d (payment %d + faction cut %d) but only have %d.",
					escrow, paymentRaw, cutAmt, n.GoldCount())
			}

			n.RemoveItem("gold", escrow)
			contract := faction.NewContract(fac.ID, n.ID, n.Name, description, condition, paymentRaw, cutAmt, w.GameDay)
			w.FactionContracts = append(w.FactionContracts, contract)

			// Notify all faction members
			for _, mid := range fac.MemberIDs {
				if mid != n.ID {
					mem.Add(mid, memory.Entry{
						Text:       fmt.Sprintf("New faction contract posted by %s: \"%s\" — satisfied when: \"%s\" — pays %d gold.", n.Name, description, condition, paymentRaw),
						Time:       w.TimeString(),
						Importance: 0.5,
						Category:   memory.CatFaction,
						Tags:       []string{n.ID, fac.Name},
					})
				}
			}
			if fac.LeaderID != n.ID {
				mem.Add(fac.LeaderID, memory.Entry{
					Text:       fmt.Sprintf("New faction contract posted by %s: \"%s\" — satisfied when: \"%s\" — pays %d gold.", n.Name, description, condition, paymentRaw),
					Time:       w.TimeString(),
					Importance: 0.5,
					Category:   memory.CatFaction,
					Tags:       []string{n.ID, fac.Name},
				})
			}
			return fmt.Sprintf("Posted faction contract to \"%s\": \"%s\" (condition: \"%s\", pays %d gold, faction cut %d gold, total escrowed: %d).", fac.Name, description, condition, paymentRaw, cutAmt, escrow)
		},
	},
	{
		ID: "accept_contract", Label: "Accept an open faction contract", Category: "faction",
		Candidates: func(n *npc.NPC, w *world.World) []*world.Location {
			return nil // uses dialogue/target instead of location candidates
		},
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if findNPCContractAsWorker(n, w) != nil {
				return false // already has a job
			}
			fac := findNPCFaction(n, w)
			if fac == nil {
				return false
			}
			for _, c := range w.FactionContracts {
				if c.FactionID == fac.ID && c.Status == "open" && c.RequesterID != n.ID {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			fac := findNPCFaction(n, w)
			if fac == nil {
				return "You are not in a faction."
			}
			// Find the first open contract (or match by dialogue hint)
			var contract *faction.FactionContract
			for _, c := range w.FactionContracts {
				if c.FactionID == fac.ID && c.Status == "open" && c.RequesterID != n.ID {
					contract = c
					break
				}
			}
			if contract == nil {
				return "No open contracts to accept."
			}
			contract.WorkerID = n.ID
			contract.WorkerName = n.Name
			contract.Status = "accepted"

			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("I accepted a faction contract from %s: \"%s\" — I must fulfil: \"%s\" — I will earn %d gold on approval (due Day %d).", contract.RequesterName, contract.Description, contract.FulfillmentCondition, contract.Payment, contract.DueDay),
				Time:       w.TimeString(),
				Importance: 0.5,
				Category:   memory.CatFaction,
				Tags:       []string{n.ID, contract.RequesterID, fac.Name},
			})
			mem.Add(contract.RequesterID, memory.Entry{
				Text:       fmt.Sprintf("%s has accepted your faction contract: \"%s\".", n.Name, contract.Description),
				Time:       w.TimeString(),
				Importance: 0.5,
				Category:   memory.CatFaction,
				Tags:       []string{n.ID, contract.RequesterID, fac.Name},
			})
			return fmt.Sprintf("Accepted contract from %s: \"%s\" (condition: \"%s\", pays %d gold, due Day %d).", contract.RequesterName, contract.Description, contract.FulfillmentCondition, contract.Payment, contract.DueDay)
		},
	},
	{
		ID: "report_contract", Label: "Report back on your accepted faction contract (describe what you did in dialogue)", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return findNPCContractAsWorker(n, w) != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			contract := findNPCContractAsWorker(n, w)
			if contract == nil {
				return "No active contract to report on."
			}
			report := strings.TrimSpace(n.LastDialogue)
			if report == "" {
				report = "Job done."
			}
			contract.WorkerReport = report
			contract.Status = "pending_review"

			// Build inventory snapshot for requester
			requester := w.FindNPCByID(contract.RequesterID)
			invSnap := "nothing notable"
			if len(n.Inventory) > 0 {
				var items []string
				for _, it := range n.Inventory {
					if it.Name != "gold" {
						items = append(items, fmt.Sprintf("%dx %s", it.Qty, it.Name))
					}
				}
				if len(items) > 0 {
					invSnap = strings.Join(items, ", ")
				}
			}

			reviewMsg := fmt.Sprintf(
				"CONTRACT REVIEW NEEDED — Job: \"%s\" | You said satisfied when: \"%s\" | %s reports: \"%s\" | %s currently holds: %s | Use approve_contract or reject_contract (%d rejections used of 3).",
				contract.Description, contract.FulfillmentCondition, n.Name, report, n.Name, invSnap, contract.RejectionCount,
			)
			if requester != nil {
				mem.Add(requester.ID, memory.Entry{Text: reviewMsg, Time: w.TimeString(), Importance: 0.5, Category: memory.CatFaction, Tags: []string{n.ID, requester.ID}})
			}
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("I reported back on the contract for %s: \"%s\". Awaiting their approval.", contract.RequesterName, contract.Description),
				Time:       w.TimeString(),
				Importance: 0.5,
				Category:   memory.CatFaction,
				Tags:       []string{n.ID, contract.RequesterID},
			})
			return fmt.Sprintf("Reported back on contract: \"%s\". Awaiting %s's approval.", contract.Description, contract.RequesterName)
		},
	},
	{
		ID: "approve_contract", Label: "Approve a completed faction contract and pay the worker", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return findNPCContractAsRequester(n, w) != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			contract := findNPCContractAsRequester(n, w)
			if contract == nil {
				return "No contract awaiting your review."
			}
			contract.Status = "completed"

			// Pay worker from escrow
			worker := w.FindNPCByID(contract.WorkerID)
			if worker != nil {
				worker.AddItem("gold", contract.Payment)
				mem.Add(worker.ID, memory.Entry{
					Text:       fmt.Sprintf("%s approved your contract work: \"%s\". You received %d gold.", n.Name, contract.Description, contract.Payment),
					Time:       w.TimeString(),
					Importance: 0.6,
					Category:   memory.CatFaction,
					Tags:       []string{worker.ID, n.ID},
				})
			}

			// Faction cut goes to faction treasury
			var fac *faction.Faction
			for _, f := range w.Factions {
				if f.ID == contract.FactionID {
					fac = f
					break
				}
			}
			if fac != nil && contract.FactionCut > 0 {
				fac.Treasury += contract.FactionCut
			}

			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("Approved the contract for \"%s\" completed by %s. Paid %d gold (faction received %d gold cut).", contract.Description, contract.WorkerName, contract.Payment, contract.FactionCut),
				Time:       w.TimeString(),
				Importance: 0.6,
				Category:   memory.CatFaction,
				Tags:       []string{n.ID, contract.WorkerID},
			})
			return fmt.Sprintf("Approved contract: \"%s\". Paid %s %d gold. Faction treasury received %d gold.", contract.Description, contract.WorkerName, contract.Payment, contract.FactionCut)
		},
	},
	{
		ID: "reject_contract", Label: "Reject the reported faction contract (state reason in dialogue)", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return findNPCContractAsRequester(n, w) != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			contract := findNPCContractAsRequester(n, w)
			if contract == nil {
				return "No contract awaiting your review."
			}
			contract.RejectionCount++
			reason := strings.TrimSpace(n.LastDialogue)
			if reason == "" {
				reason = "Not satisfied with the work."
			}

			worker := w.FindNPCByID(contract.WorkerID)

			if contract.RejectionCount >= 3 {
				// Force expire — return escrow to requester
				contract.Status = "expired"
				n.AddItem("gold", contract.EscrowGold)
				if worker != nil {
					worker.Stats.Reputation = clamp(worker.Stats.Reputation-15, 0, 100)
					worker.Stress = clamp(worker.Stress+20, 0, 100)
					mem.Add(worker.ID, memory.Entry{
						Text:       fmt.Sprintf("My contract for %s was rejected 3 times and expired. I lost the job and my reputation suffered. Reason: \"%s\"", contract.RequesterName, reason),
						Time:       w.TimeString(),
						Importance: 0.6,
						Category:   memory.CatFaction,
						Tags:       []string{worker.ID, n.ID},
					})
					// Notify other faction members
					var fac *faction.Faction
					for _, f := range w.Factions {
						if f.ID == contract.FactionID {
							fac = f
							break
						}
					}
					if fac != nil {
						for _, mid := range fac.MemberIDs {
							if mid != worker.ID && mid != n.ID {
								mem.Add(mid, memory.Entry{
									Text:       fmt.Sprintf("%s failed to complete a faction contract for %s after 3 rejections.", worker.Name, n.Name),
									Time:       w.TimeString(),
									Importance: 0.6,
									Category:   memory.CatFaction,
									Tags:       []string{worker.ID, n.ID, fac.Name},
								})
							}
						}
					}
				}
				mem.Add(n.ID, memory.Entry{
					Text:       fmt.Sprintf("Rejected %s's contract work for the 3rd time. Contract expired, %d gold returned.", contract.WorkerName, contract.EscrowGold),
					Time:       w.TimeString(),
					Importance: 0.6,
					Category:   memory.CatFaction,
					Tags:       []string{n.ID, contract.WorkerID},
				})
				return fmt.Sprintf("Rejected contract for the 3rd time — contract expired. Escrow of %d gold returned.", contract.EscrowGold)
			}

			// Not yet expired — send worker back
			contract.Status = "accepted"
			if worker != nil {
				mem.Add(worker.ID, memory.Entry{
					Text:       fmt.Sprintf("%s rejected your contract report (%d/3 rejections). Reason: \"%s\" — try again or abandon_contract.", n.Name, contract.RejectionCount, reason),
					Time:       w.TimeString(),
					Importance: 0.6,
					Category:   memory.CatFaction,
					Tags:       []string{worker.ID, n.ID},
				})
			}
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("Rejected %s's contract work (%d/3). Reason: \"%s\"", contract.WorkerName, contract.RejectionCount, reason),
				Time:       w.TimeString(),
				Importance: 0.6,
				Category:   memory.CatFaction,
				Tags:       []string{n.ID, contract.WorkerID},
			})
			return fmt.Sprintf("Rejected contract work by %s (%d/3 rejections). Reason: \"%s\". They can try again.", contract.WorkerName, contract.RejectionCount, reason)
		},
	},
	{
		ID: "abandon_contract", Label: "Abandon your accepted faction contract (no penalty, escrow returned to requester)", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			return findNPCContractAsWorker(n, w) != nil
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			contract := findNPCContractAsWorker(n, w)
			if contract == nil {
				return "No active contract to abandon."
			}
			contract.Status = "abandoned"

			// Return escrow to requester
			requester := w.FindNPCByID(contract.RequesterID)
			if requester != nil {
				requester.AddItem("gold", contract.EscrowGold)
				mem.Add(requester.ID, memory.Entry{
					Text:       fmt.Sprintf("%s has abandoned your contract: \"%s\". Your %d gold escrow has been returned.", n.Name, contract.Description, contract.EscrowGold),
					Time:       w.TimeString(),
					Importance: 0.4,
					Category:   memory.CatFaction,
					Tags:       []string{n.ID, requester.ID},
				})
			}
			mem.Add(n.ID, memory.Entry{
				Text:       fmt.Sprintf("Abandoned the faction contract from %s: \"%s\". No penalty — escrow returned to them.", contract.RequesterName, contract.Description),
				Time:       w.TimeString(),
				Importance: 0.4,
				Category:   memory.CatFaction,
				Tags:       []string{n.ID, contract.RequesterID},
			})
			return fmt.Sprintf("Abandoned contract: \"%s\". Escrow of %d gold returned to %s.", contract.Description, contract.EscrowGold, contract.RequesterName)
		},
	},
	{
		ID: "leave_faction", Label: "Leave your faction (severe relationship penalty with all members)", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.FactionID == "" {
				return false
			}
			for _, f := range w.Factions {
				if f.ID == n.FactionID && f.LeaderID == n.ID {
					return false // leader cannot leave
				}
			}
			return true
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			var fac *faction.Faction
			for _, f := range w.Factions {
				if f.ID == n.FactionID {
					fac = f
					break
				}
			}
			if fac == nil {
				n.FactionID = ""
				return "Left an unknown faction."
			}

			// Remove from member list
			newMembers := make([]string, 0, len(fac.MemberIDs))
			for _, mid := range fac.MemberIDs {
				if mid != n.ID {
					newMembers = append(newMembers, mid)
				}
			}
			fac.MemberIDs = newMembers
			n.FactionID = ""

			n.Stress = clamp(n.Stress+10, 0, 100)
			n.AdjustRelationship(fac.LeaderID, -40)

			leaderMsg := fmt.Sprintf("%s has left \"%s\".", n.Name, fac.Name)
			mem.Add(fac.LeaderID, memory.Entry{Text: leaderMsg, Time: w.TimeString(), Importance: 0.6, Category: memory.CatFaction, Tags: []string{n.ID, fac.LeaderID, fac.Name}})

			for _, mid := range fac.MemberIDs {
				n.AdjustRelationship(mid, -30)
				other := w.FindNPCByID(mid)
				if other != nil {
					other.AdjustRelationship(n.ID, -20)
				}
				mem.Add(mid, memory.Entry{Text: leaderMsg, Time: w.TimeString(), Importance: 0.6, Category: memory.CatFaction, Tags: []string{n.ID, fac.Name}})
			}

			w.LogEvent(leaderMsg, "faction")
			return fmt.Sprintf("Left \"%s\". Relationship penalties applied with all %d members.", fac.Name, len(fac.MemberIDs)+1)
		},
	},
	{
		ID: "kick_member", Label: "Expel a member from your faction (name them as target)", Category: "faction",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			for _, f := range w.Factions {
				if f.LeaderID == n.ID && len(f.MemberIDs) > 1 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			var fac *faction.Faction
			for _, f := range w.Factions {
				if f.LeaderID == n.ID {
					fac = f
					break
				}
			}
			if fac == nil {
				return "You lead no faction."
			}
			if target == nil {
				return "Specify the member to expel as target."
			}

			// Verify target is a member
			found := false
			newMembers := make([]string, 0, len(fac.MemberIDs))
			for _, mid := range fac.MemberIDs {
				if mid == target.ID {
					found = true
				} else {
					newMembers = append(newMembers, mid)
				}
			}
			if !found {
				return fmt.Sprintf("%s is not a member of your faction.", target.Name)
			}
			fac.MemberIDs = newMembers
			target.FactionID = ""

			target.Stats.Reputation = clamp(target.Stats.Reputation-10, 0, 100)
			target.Stress = clamp(target.Stress+15, 0, 100)
			target.AdjustRelationship(n.ID, -40)

			expelMsg := fmt.Sprintf("%s has expelled %s from \"%s\".", n.Name, target.Name, fac.Name)
			mem.Add(target.ID, memory.Entry{
				Text:       fmt.Sprintf("I was expelled from \"%s\" by %s. I bear a -40 relationship penalty with them.", fac.Name, n.Name),
				Time:       w.TimeString(),
				Importance: 0.8,
				Category:   memory.CatFaction,
				Tags:       []string{target.ID, n.ID, fac.Name},
			})
			mem.Add(fac.LeaderID, memory.Entry{Text: expelMsg, Time: w.TimeString(), Importance: 0.8, Category: memory.CatFaction, Tags: []string{target.ID, n.ID, fac.Name}})

			for _, mid := range fac.MemberIDs {
				target.AdjustRelationship(mid, -20)
				mem.Add(mid, memory.Entry{Text: expelMsg, Time: w.TimeString(), Importance: 0.8, Category: memory.CatFaction, Tags: []string{target.ID, n.ID, fac.Name}})
			}

			w.LogEvent(expelMsg, "faction")
			return fmt.Sprintf("Expelled %s from \"%s\".", target.Name, fac.Name)
		},
	},
}
