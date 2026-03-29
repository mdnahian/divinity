package gametools

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/divinity/core/action"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/world"
)

func NPCTools() []*ToolDef {
	return []*ToolDef{
		observeTool(),
		checkSelfTool(),
		inspectPersonTool(),
		recallMemoriesTool(),
		checkLocationTool(),
		listActionsTool(),
		commitActionTool(),
	}
}

func observeTool() *ToolDef {
	return &ToolDef{
		Name:        "observe",
		Description: "Look around your current location. See who is here, enemies, items on the ground, and resources.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			n := ctx.NPC
			w := ctx.World
			loc := w.LocationByID(n.LocationID)
			if loc == nil {
				return "You are in an unknown place.", nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Location: %s (%s)", loc.Name, loc.Type))
			if loc.Description != "" {
				sb.WriteString(fmt.Sprintf(" — %s", loc.Description))
			}
			sb.WriteString(fmt.Sprintf("\nTime: %s", w.TimeString()))
			if w.IsNight() {
				sb.WriteString(" (night)")
			}
			sb.WriteString(fmt.Sprintf("\nWeather: %s — %s", w.Weather, w.WeatherDescription()))

			if loc.Resources != nil {
				var rParts []string
				for k, v := range loc.Resources {
					if v > 0 {
						rParts = append(rParts, fmt.Sprintf("%s: %d", k, v))
					}
				}
				if len(rParts) > 0 {
					sb.WriteString(fmt.Sprintf("\nResources here: %s", strings.Join(rParts, ", ")))
				}
			}

			locOwner := w.GetLocationOwner(n.LocationID)
			if locOwner != nil && locOwner.ID != n.ID {
				sb.WriteString(fmt.Sprintf("\nOwned by: %s", locOwner.Name))
			}

			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			if len(nearby) > 0 {
				sb.WriteString("\nPeople here:")
				for _, other := range nearby {
					rel := n.GetRelationship(other.ID)
					relLabel := "neutral"
					switch {
					case rel > 30:
						relLabel = "friend"
					case rel > 10:
						relLabel = "acquaintance"
					case rel < -30:
						relLabel = "enemy"
					case rel < -10:
						relLabel = "disliked"
					}
					sb.WriteString(fmt.Sprintf("\n  - %s (%s, mood: %s, relation: %s [%d])",
						other.Name, other.Profession, other.Mood(), relLabel, rel))
				}
			} else {
				sb.WriteString("\nNobody else here.")
			}

			enemies := w.EnemiesAtLocation(n.LocationID)
			if len(enemies) > 0 {
				sb.WriteString("\nENEMIES HERE:")
				for _, e := range enemies {
					sb.WriteString(fmt.Sprintf("\n  - %s (%s, HP: %d/%d, str: %d)",
						e.Name, e.Category, e.HP, e.MaxHP, e.Strength))
				}
			}

			groundItems := w.GroundItemsAt(n.LocationID)
			if len(groundItems) > 0 {
				var gItems []string
				for _, g := range groundItems {
					gItems = append(gItems, fmt.Sprintf("%s x%d", g.Name, g.Qty))
				}
				sb.WriteString(fmt.Sprintf("\nItems on ground: %s", strings.Join(gItems, ", ")))
			}

			// Show location memories (significant past events at this location)
			if ctx.SharedMemory != nil {
				locMems := ctx.SharedMemory.LocationMemories(n.LocationID)
				if len(locMems) > 0 {
					sb.WriteString("\nThis place is known for:")
					for _, lm := range locMems {
						sb.WriteString(fmt.Sprintf("\n  - %s", lm.Text))
					}
				}
			}

			var activeEvents []string
			seen := make(map[string]bool)
			for _, e := range w.ActiveEvents {
				if e.TicksLeft > 0 && !seen[e.Name] {
					seen[e.Name] = true
					activeEvents = append(activeEvents, e.Name)
				}
			}
			if len(activeEvents) > 0 {
				sb.WriteString(fmt.Sprintf("\nActive events: %s", strings.Join(activeEvents, ", ")))
			}

			return sb.String(), nil
		},
	}
}

func needsBar(val float64) string {
	return fmt.Sprintf("%.0f/100 (%d%%)", val, int(math.Round(val)))
}

func needsBarInt(val int) string {
	return fmt.Sprintf("%d/100", val)
}

func checkSelfTool() *ToolDef {
	return &ToolDef{
		Name:        "check_self",
		Description: "Check your own state: HP, satiation, hydration, fatigue, stress, happiness, inventory, equipment, gold, skills, and relationships.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			n := ctx.NPC
			w := ctx.World
			cfg := ctx.Config

			age := n.GetAge(w.GameDay, cfg.Game.GameDaysPerYear)
			lifeStage := n.GetLifeStage(w.GameDay, cfg.Game.GameDaysPerYear)

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Name: %s | Age: %d (%s) | Profession: %s\n",
				n.Name, age, lifeStage, n.Profession))
			sb.WriteString(fmt.Sprintf("Personality: %s | Mood: %s\n",
				n.PersonalitySummary(), n.Mood()))
			sb.WriteString(fmt.Sprintf("HP: %d/100 | Satiation: %s | Hydration: %s\n",
				n.HP, needsBar(n.Needs.Hunger), needsBar(n.Needs.Thirst)))
			sb.WriteString(fmt.Sprintf("Fatigue: %s | Social: %s\n",
				needsBar(n.Needs.Fatigue), needsBar(100-n.Needs.SocialNeed)))
			sb.WriteString(fmt.Sprintf("Happiness: %s | Stress: %s\n",
				needsBarInt(n.Happiness), needsBarInt(n.Stress)))
			sb.WriteString(fmt.Sprintf("Hygiene: %s | Sobriety: %s | Reputation: %d\n",
				needsBarInt(n.Hygiene), needsBarInt(n.Sobriety), n.Stats.Reputation))

			var skills []struct {
				Name  string
				Value float64
			}
			for k, v := range n.Skills {
				skills = append(skills, struct {
					Name  string
					Value float64
				}{k, v})
			}
			sort.Slice(skills, func(i, j int) bool { return skills[i].Value > skills[j].Value })
			if len(skills) > 5 {
				skills = skills[:5]
			}
			var skillStrs []string
			for _, s := range skills {
				skillStrs = append(skillStrs, fmt.Sprintf("%s: %d", s.Name, int(math.Round(s.Value))))
			}
			if len(skillStrs) > 0 {
				sb.WriteString(fmt.Sprintf("Top skills: %s\n", strings.Join(skillStrs, ", ")))
			}

			var equippedSlots []string
			if n.Equipment.Weapon != nil {
				equippedSlots = append(equippedSlots, fmt.Sprintf("%s (weapon)", n.Equipment.Weapon.Name))
			}
			if n.Equipment.Armor != nil {
				equippedSlots = append(equippedSlots, fmt.Sprintf("%s (armor)", n.Equipment.Armor.Name))
			}
			if n.Equipment.Bag1 != nil {
				equippedSlots = append(equippedSlots, fmt.Sprintf("%s (bag)", n.Equipment.Bag1.Name))
			}
			if n.Equipment.Bag2 != nil {
				equippedSlots = append(equippedSlots, fmt.Sprintf("%s (bag)", n.Equipment.Bag2.Name))
			}
			if len(equippedSlots) > 0 {
				sb.WriteString(fmt.Sprintf("Equipped: %s\n", strings.Join(equippedSlots, ", ")))
			}

			if len(n.Inventory) > 0 {
				var items []string
				for _, it := range n.Inventory {
					items = append(items, fmt.Sprintf("%s x%d", it.Name, it.Qty))
				}
				sb.WriteString(fmt.Sprintf("Inventory: %s\n", strings.Join(items, ", ")))
			} else {
				sb.WriteString("Inventory: empty\n")
			}

			type relEntry struct {
				Name string
				Val  int
			}
			var rels []relEntry
			for id, val := range n.Relationships {
				if abs(val) > 5 {
					other := w.FindNPCByID(id)
					name := "someone"
					if other != nil {
						name = other.Name
					}
					rels = append(rels, relEntry{name, val})
				}
			}
			sort.Slice(rels, func(i, j int) bool { return abs(rels[i].Val) > abs(rels[j].Val) })
			if len(rels) > 6 {
				rels = rels[:6]
			}
			if len(rels) > 0 {
				var relStrs []string
				for _, r := range rels {
					sign := "+"
					if r.Val < 0 {
						sign = ""
					}
					relStrs = append(relStrs, fmt.Sprintf("%s: %s%d", r.Name, sign, r.Val))
				}
				sb.WriteString(fmt.Sprintf("Relationships: %s\n", strings.Join(relStrs, ", ")))
			}

			if n.IsBusinessOwner {
				sb.WriteString("You own a business.\n")
			} else if n.EmployerID != "" {
				employer := w.FindNPCByID(n.EmployerID)
				empName := "?"
				if employer != nil {
					empName = employer.Name
				}
				sb.WriteString(fmt.Sprintf("Employed by %s (wage: %d gold/day)\n", empName, n.Wage))
			}

			if n.FactionID != "" {
				for _, f := range w.Factions {
					if f.ID == n.FactionID {
						role := "Member"
						if f.LeaderID == n.ID {
							role = "Leader"
						}
						sb.WriteString(fmt.Sprintf("Faction: %s of \"%s\" (%s)\n", role, f.Name, f.Type))
						break
					}
				}
			}

			return sb.String(), nil
		},
	}
}

func inspectPersonTool() *ToolDef {
	return &ToolDef{
		Name:        "inspect_person",
		Description: "Learn about a nearby NPC — their mood, profession, visible needs, and your relationship with them.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Name of the person to inspect"}},"required":["name"]}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			var params struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "Invalid parameters.", nil
			}

			n := ctx.NPC
			w := ctx.World

			target := w.FindNPCByNameAtLocation(params.Name, n.LocationID)
			if target == nil {
				return fmt.Sprintf("Nobody named %q exists.", params.Name), nil
			}
			if !target.Alive {
				return fmt.Sprintf("%s is dead.", params.Name), nil
			}
			if target.LocationID != n.LocationID {
				return fmt.Sprintf("%s is not at your location.", params.Name), nil
			}

			rel := n.GetRelationship(target.ID)
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%s — %s, mood: %s\n", target.Name, target.Profession, target.Mood()))
			sb.WriteString(fmt.Sprintf("Your relationship: %d\n", rel))
			sb.WriteString(fmt.Sprintf("HP: %d/100", target.HP))

			var needs []string
			if target.Needs.Hunger < 30 {
				needs = append(needs, "hungry")
			}
			if target.Needs.Thirst < 20 {
				needs = append(needs, "thirsty")
			}
			if target.Stress > 50 {
				needs = append(needs, "stressed")
			}
			if target.Needs.SocialNeed > 60 {
				needs = append(needs, "lonely")
			}
			if target.HP < 50 {
				needs = append(needs, "injured")
			}
			if len(needs) > 0 {
				sb.WriteString(fmt.Sprintf("\nVisible state: %s", strings.Join(needs, ", ")))
			}

			if target.IsBusinessOwner {
				sb.WriteString("\nOwns a business.")
			}
			if target.FactionID != "" {
				for _, f := range w.Factions {
					if f.ID == target.FactionID {
						sb.WriteString(fmt.Sprintf("\nFaction: %s", f.Name))
						break
					}
				}
			}

			return sb.String(), nil
		},
	}
}

func recallMemoriesTool() *ToolDef {
	return &ToolDef{
		Name:        "recall_memories",
		Description: "Search your recent memories. Optionally filter by a topic/keyword or category (combat, economic, social, employment, miracle, dream, education, faction).",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"topic":{"type":"string","description":"Optional keyword or topic to filter memories by"},"category":{"type":"string","description":"Optional category filter: combat, economic, social, employment, miracle, dream, education, faction"}}}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			var params struct {
				Topic    string `json:"topic"`
				Category string `json:"category"`
			}
			json.Unmarshal(args, &params)

			var entries []memory.Entry
			topic := strings.ToLower(strings.TrimSpace(params.Topic))
			category := strings.ToLower(strings.TrimSpace(params.Category))

			if category != "" {
				entries = ctx.Memory.ByCategory(ctx.NPC.ID, category)
			} else if topic != "" {
				entries = ctx.Memory.Search(ctx.NPC.ID, topic)
			} else {
				entries = ctx.Memory.Recent(ctx.NPC.ID, memory.MaxMemories)
			}

			if len(entries) == 0 {
				if topic != "" {
					return fmt.Sprintf("No memories about %q.", params.Topic), nil
				}
				if category != "" {
					return fmt.Sprintf("No %s memories.", category), nil
				}
				return "No memories yet.", nil
			}

			var sb strings.Builder
			for _, e := range entries {
				time := e.Time
				if time == "" {
					time = "?"
				}
				prefix := ""
				if e.Vividness < 0.3 {
					prefix = "(vague) "
				} else if e.Vividness < 0.5 {
					prefix = "(distant) "
				}
				sb.WriteString(fmt.Sprintf("[%s] %s%s\n", time, prefix, e.Text))
			}
			return sb.String(), nil
		},
	}
}

func checkLocationTool() *ToolDef {
	return &ToolDef{
		Name:        "check_location",
		Description: "Scout a location you know about — see its resources, who is there, and travel time from your current position.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"location_id":{"type":"string","description":"ID of the location to check"}},"required":["location_id"]}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			var params struct {
				LocationID string `json:"location_id"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "Invalid parameters.", nil
			}

			n := ctx.NPC
			w := ctx.World
			mpt := ctx.Config.Game.GameMinutesPerTick

			loc := w.LocationByID(params.LocationID)
			if loc == nil {
				loc = w.LocationByName(params.LocationID)
			}
			if loc == nil {
				loc = w.LocationByNameFuzzy(params.LocationID)
			}
			if loc == nil {
				return fmt.Sprintf("Unknown location %q.", params.LocationID), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%s (%s)", loc.Name, loc.Type))
			if loc.Description != "" {
				sb.WriteString(fmt.Sprintf(" — %s", loc.Description))
			}

			if loc.Resources != nil {
				var rParts []string
				for k, v := range loc.Resources {
					mx := 0
					if loc.MaxResources != nil {
						mx = loc.MaxResources[k]
					}
					rParts = append(rParts, fmt.Sprintf("%s: %d/%d", k, v, mx))
				}
				if len(rParts) > 0 {
					sb.WriteString(fmt.Sprintf("\nResources: %s", strings.Join(rParts, ", ")))
				}
			}

			npcsHere := w.NPCsAtLocation(params.LocationID, "")
			if len(npcsHere) > 0 {
				var names []string
				for _, p := range npcsHere {
					names = append(names, p.Name)
				}
				sb.WriteString(fmt.Sprintf("\nPeople there: %s", strings.Join(names, ", ")))
			}

			enemies := w.EnemiesAtLocation(params.LocationID)
			if len(enemies) > 0 {
				sb.WriteString(fmt.Sprintf("\nEnemies: %d present!", len(enemies)))
			}

			if params.LocationID != n.LocationID {
				travelTicks := w.TravelTicks(n.LocationID, params.LocationID, mpt, ctx.Config.Game.TravelMinutesPerUnit)
				travelMins := travelTicks * mpt
				sb.WriteString(fmt.Sprintf("\nTravel time from here: ~%d min", travelMins))
			} else {
				sb.WriteString("\nYou are here.")
			}

			return sb.String(), nil
		},
	}
}

func listActionsTool() *ToolDef {
	return &ToolDef{
		Name:        "list_actions",
		Description: "List all actions available to you right now, with durations and location options.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			n := ctx.NPC
			w := ctx.World
			mpt := ctx.Config.Game.GameMinutesPerTick

			actions := action.GetAvailableActions(n, w)
			if len(actions) == 0 {
				return "No actions available.", nil
			}

			var sb strings.Builder
			if n.ResumeActionID != "" {
				resumeMins := n.ResumeTicksLeft * mpt
				resumeAct := action.FindAction(n.ResumeActionID)
				resumeLabel := n.ResumeActionID
				if resumeAct != nil {
					resumeLabel = resumeAct.Label
				}
				sb.WriteString(fmt.Sprintf("- \"resume\": Resume %s (~%d min remaining)\n", resumeLabel, resumeMins))
			}

			const maxCandidates = 10

			for _, a := range actions {
				_, mins := a.Duration(n, mpt)

				if a.Candidates != nil {
					candidates := a.Candidates(n, w)
					if len(candidates) > 0 {
						// Sort candidates by travel distance (current location first, then nearest)
						type candidateWithDist struct {
							loc      *world.Location
							footMins int
							here     bool
						}
						var cds []candidateWithDist
						for _, c := range candidates {
							if c.ID == n.LocationID {
								cds = append(cds, candidateWithDist{loc: c, footMins: 0, here: true})
							} else {
								footTicks := w.TravelTicks(n.LocationID, c.ID, mpt, ctx.Config.Game.TravelMinutesPerUnit)
								footMins := footTicks * mpt
								cds = append(cds, candidateWithDist{loc: c, footMins: footMins})
							}
						}
						sort.Slice(cds, func(i, j int) bool {
							return cds[i].footMins < cds[j].footMins
						})

						// Limit to nearest maxCandidates locations
						shown := cds
						if len(shown) > maxCandidates {
							shown = shown[:maxCandidates]
						}

						sb.WriteString(fmt.Sprintf("- \"%s\": %s [~%d min]\n", a.ID, a.Label, mins))
						for _, cd := range shown {
							if cd.here {
								sb.WriteString(fmt.Sprintf("    Location: \"%s\" (here)\n", cd.loc.Name))
							} else {
								mountedTicks, hasMnt, hasCarr := w.TravelTicksMounted(n.LocationID, cd.loc.ID, mpt, ctx.Config.Game.TravelMinutesPerUnit, n.ID)
								mountedMins := mountedTicks * mpt
								if hasMnt {
									label := "riding"
									if hasCarr {
										label = "by carriage"
									}
									sb.WriteString(fmt.Sprintf("    Location: \"%s\" (+%d min %s, %d min on foot)\n", cd.loc.Name, mountedMins, label, cd.footMins))
								} else {
									sb.WriteString(fmt.Sprintf("    Location: \"%s\" (+%d min travel)\n", cd.loc.Name, cd.footMins))
								}
							}
						}
						if len(cds) > maxCandidates {
							sb.WriteString(fmt.Sprintf("    ... and %d more distant locations\n", len(cds)-maxCandidates))
						}
						continue
					}
				}

				travelMins := 0
				travelLabel := ""
				if a.Destination != nil {
					destID := a.Destination(n, w)
					if destID != "" && destID != n.LocationID {
						footTicks := w.TravelTicks(n.LocationID, destID, mpt, ctx.Config.Game.TravelMinutesPerUnit)
						footMins := footTicks * mpt
						mountedTicks, hasMnt, hasCarr := w.TravelTicksMounted(n.LocationID, destID, mpt, ctx.Config.Game.TravelMinutesPerUnit, n.ID)
						mountedMins := mountedTicks * mpt
						if hasMnt {
							label := "riding"
							if hasCarr {
								label = "by carriage"
							}
							travelMins = mountedMins
							travelLabel = fmt.Sprintf("%d min %s, %d min on foot", mountedMins, label, footMins)
						} else {
							travelMins = footMins
							travelLabel = fmt.Sprintf("%d min travel", footMins)
						}
					}
				}
				if travelMins > 0 {
					sb.WriteString(fmt.Sprintf("- \"%s\": %s [~%d min + %s]\n", a.ID, a.Label, mins, travelLabel))
				} else {
					sb.WriteString(fmt.Sprintf("- \"%s\": %s [~%d min]\n", a.ID, a.Label, mins))
				}
			}

			return sb.String(), nil
		},
	}
}

func commitActionTool() *ToolDef {
	return &ToolDef{
		Name:        "commit_action",
		Description: "Commit to performing an action. This ends your turn — you will be busy until the action completes.",
		Parameters: json.RawMessage(`{"type":"object","properties":{
			"action_id":{"type":"string","description":"The action ID from list_actions"},
			"target":{"type":"string","description":"Name of a nearby NPC if the action involves one"},
			"dialogue":{"type":"string","description":"What you say, if the action is social or involves speaking"},
			"goal":{"type":"string","description":"Your current goal or intention"},
			"location":{"type":"string","description":"Location name from the @ list, if the action has location options"},
			"reason":{"type":"string","description":"Brief reason for choosing this action (1 short sentence)"}
		},"required":["action_id","reason"]}`),
		IsTerminal: true,
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			var params struct {
				ActionID string `json:"action_id"`
				Target   string `json:"target"`
				Dialogue string `json:"dialogue"`
				Goal     string `json:"goal"`
				Location string `json:"location"`
				Reason   string `json:"reason"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "Invalid parameters.", nil
			}

			if strings.TrimSpace(params.Reason) == "" {
				return "You must provide a reason for your action.", nil
			}

			n := ctx.NPC
			w := ctx.World

			act := action.FindAction(params.ActionID)
			if act == nil && params.ActionID != "resume" {
				return fmt.Sprintf("Unknown action %q. Use list_actions to see available actions.", params.ActionID), nil
			}

			if act != nil && act.Conditions != nil && !act.Conditions(n, w) {
				return fmt.Sprintf("Cannot do %q right now — conditions not met. Use list_actions to see what's available.", params.ActionID), nil
			}

			if params.Target != "" {
				// Prefer NPCs at the same location to avoid resolving the
				// wrong NPC when multiple share a name.
				target := w.FindNPCByNameAtLocation(params.Target, n.LocationID)
				if target == nil {
					return fmt.Sprintf("%s doesn't exist. Choose someone else.", params.Target), nil
				}
				if !target.Alive {
					return fmt.Sprintf("%s is dead.", params.Target), nil
				}
				if target.LocationID != n.LocationID {
					return fmt.Sprintf("%s is no longer here. Choose someone else or follow them.", params.Target), nil
				}
			}

			return "ACTION_COMMITTED", nil
		},
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
