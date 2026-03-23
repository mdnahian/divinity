package gametools

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/divinity/core/enemy"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

func GodTools() []*ToolDef {
	return []*ToolDef{
		villageOverviewTool(),
		listNPCsTool(),
		queryNPCTool(),
		checkResourcesTool(),
		recentEventsTool(),
		checkEnemiesTool(),
		checkFactionsTool(),
		readPrayersTool(),
		godMemoriesTool(),
		godTrendsTool(),
		godActTool(),
		territoryDetailTool(),
		nobleHierarchyTool(),
		influenceMapTool(),
		socialClustersTool(),
		tensionAnalysisTool(),
	}
}

func villageOverviewTool() *ToolDef {
	return &ToolDef{
		Name:        "village_overview",
		Description: "Get a high-level summary: population, economy, crisis level, weather, time, buildings, and health analytics.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			w := ctx.World
			alive := w.AliveNPCs()
			deadCount := len(w.NPCs) - len(alive)

			totalGold := 0
			hungryCount := 0
			starvingCount := 0
			brokeCount := 0
			totalStr := 0
			for _, n := range alive {
				gold := n.GoldCount()
				totalGold += gold
				totalStr += n.Stats.Strength
				if n.Needs.Hunger < 30 {
					hungryCount++
				}
				if n.Needs.Hunger < 10 {
					starvingCount++
				}
				if gold == 0 {
					brokeCount++
				}
			}
			avgStr := 0
			if len(alive) > 0 {
				avgStr = totalStr / len(alive)
			}

			crisisLevel := "STABLE"
			if len(alive) > 0 {
				hungryPct := hungryCount * 100 / len(alive)
				brokePct := brokeCount * 100 / len(alive)
				if hungryPct >= 50 || brokePct >= 50 {
					crisisLevel = "CRITICAL"
				} else if hungryPct >= 25 || brokePct >= 25 {
					crisisLevel = "CONCERNING"
				}
			}

			buildingCount := 0
			for _, l := range w.Locations {
				if l.BuildingType != "" {
					buildingCount++
				}
			}

			profCounts := make(map[string]int)
			for _, n := range alive {
				profCounts[n.Profession]++
			}
			var profParts []string
			for p, c := range profCounts {
				profParts = append(profParts, fmt.Sprintf("%s: %d", p, c))
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Time: %s | Weather: %s\n", w.TimeString(), w.Weather))
			sb.WriteString(fmt.Sprintf("Population: %d alive, %d dead\n", len(alive), deadCount))
			sb.WriteString(fmt.Sprintf("Professions: %s\n", strings.Join(profParts, ", ")))
			sb.WriteString(fmt.Sprintf("Buildings: %d | Locations: %d | Grid: %dx%d\n", buildingCount, len(w.Locations), w.GridW, w.GridH))
			sb.WriteString(fmt.Sprintf("Avg NPC strength: %d | Enemies alive: %d\n", avgStr, len(w.AliveEnemies())))
			sb.WriteString(fmt.Sprintf("Total gold in circulation: %d\n", totalGold))
			sb.WriteString(fmt.Sprintf("Factions: %d | Techniques: %d\n", len(w.Factions), len(w.Techniques)))
			sb.WriteString(fmt.Sprintf("\nCRISIS LEVEL: %s\n", crisisLevel))
			sb.WriteString(fmt.Sprintf("Hungry (<30): %d/%d | Starving (<10): %d | Broke: %d/%d\n",
				hungryCount, len(alive), starvingCount, brokeCount, len(alive)))

			if len(w.ActiveEvents) > 0 {
				sb.WriteString("\nActive events:")
				for _, e := range w.ActiveEvents {
					sb.WriteString(fmt.Sprintf("\n  - %s (%d ticks left)", e.Name, e.TicksLeft))
				}
			}

			return sb.String(), nil
		},
	}
}

func listNPCsTool() *ToolDef {
	return &ToolDef{
		Name:        "list_npcs",
		Description: "List all living NPCs. If territory_id is provided, show detailed stats for NPCs in that territory. If omitted, show a summary by territory with counts and crisis NPCs only.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"territory_id":{"type":"string","description":"Optional territory ID to filter NPCs. Omit for a summary by territory."}}}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			var params struct {
				TerritoryID string `json:"territory_id"`
			}
			json.Unmarshal(args, &params)

			w := ctx.World
			daysPerYear := ctx.Config.Game.GameDaysPerYear
			alive := w.AliveNPCs()
			if len(alive) == 0 {
				return "No living NPCs.", nil
			}

			if params.TerritoryID != "" {
				// Filtered mode: show detailed stats for NPCs in the given territory
				var sb strings.Builder
				count := 0
				for _, n := range alive {
					if n.TerritoryID != params.TerritoryID {
						continue
					}
					count++
					age := n.GetAge(w.GameDay, daysPerYear)
					flags := ""
					if n.Needs.Hunger < 10 {
						flags += " !! STARVING"
					}
					if n.Needs.Thirst < 10 {
						flags += " !! DEHYDRATED"
					}
					if n.HP < 20 {
						flags += " !! NEAR DEATH"
					}
					if n.GoldCount() == 0 {
						flags += " BROKE"
					}
					sb.WriteString(fmt.Sprintf("- %s (%s, %dyo, mood: %s, HP: %d, loc: %s, gold: %d)%s\n",
						n.Name, n.Profession, age, n.Mood(), n.HP, n.LocationID, n.GoldCount(), flags))
				}
				if count == 0 {
					return fmt.Sprintf("No NPCs in territory %s.", params.TerritoryID), nil
				}
				return fmt.Sprintf("NPCs in %s (%d):\n%s", params.TerritoryID, count, sb.String()), nil
			}

			// Summary mode: count per territory + crisis NPCs only
			type terrSummary struct {
				name  string
				count int
			}
			terrCounts := make(map[string]*terrSummary)
			var crisisNPCs []string

			for _, n := range alive {
				tid := n.TerritoryID
				if tid == "" {
					tid = "(unassigned)"
				}
				s, ok := terrCounts[tid]
				if !ok {
					tName := tid
					if t := w.TerritoryByID(tid); t != nil {
						tName = t.Name
					}
					s = &terrSummary{name: tName}
					terrCounts[tid] = s
				}
				s.count++

				if n.HP < 20 || n.Needs.Hunger < 10 {
					flags := ""
					if n.HP < 20 {
						flags += " HP:" + fmt.Sprintf("%d", n.HP)
					}
					if n.Needs.Hunger < 10 {
						flags += fmt.Sprintf(" hunger:%.0f", n.Needs.Hunger)
					}
					crisisNPCs = append(crisisNPCs, fmt.Sprintf("  !! %s (%s, %s)%s", n.Name, n.Profession, tid, flags))
				}
			}

			var sb strings.Builder
			sb.WriteString("NPC Summary by Territory:\n")
			for tid, s := range terrCounts {
				sb.WriteString(fmt.Sprintf("  %s (%s): %d NPCs\n", s.name, tid, s.count))
			}

			if len(crisisNPCs) > 0 {
				sb.WriteString(fmt.Sprintf("\nCrisis NPCs (%d):\n", len(crisisNPCs)))
				for _, c := range crisisNPCs {
					sb.WriteString(c + "\n")
				}
			}
			sb.WriteString(fmt.Sprintf("\nTotal: %d alive\n", len(alive)))
			return sb.String(), nil
		},
	}
}

func queryNPCTool() *ToolDef {
	return &ToolDef{
		Name:        "query_npc",
		Description: "Deep-inspect a specific NPC: identity, vitals, mood, memories, relationships, goals, equipment, and top skills. Use this to understand individual motivations before sending dreams or visions.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"npc_name":{"type":"string","description":"The NPC's name (case-insensitive)"}},"required":["npc_name"]}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			var params struct {
				NPCName string `json:"npc_name"`
			}
			json.Unmarshal(args, &params)
			if params.NPCName == "" {
				return "npc_name is required.", nil
			}

			w := ctx.World

			// Case-insensitive search
			var n *npc.NPC
			nameLower := strings.ToLower(params.NPCName)
			for _, candidate := range w.NPCs {
				if candidate.Alive && strings.ToLower(candidate.Name) == nameLower {
					n = candidate
					break
				}
			}
			if n == nil {
				return fmt.Sprintf("No living NPC named %q found.", params.NPCName), nil
			}

			daysPerYear := ctx.Config.Game.GameDaysPerYear
			age := n.GetAge(w.GameDay, daysPerYear)

			var sb strings.Builder

			// Identity
			sb.WriteString(fmt.Sprintf("=== %s ===\n", n.Name))
			sb.WriteString(fmt.Sprintf("ID: %s | Profession: %s | Age: %d\n", n.ID, n.Profession, age))
			if n.Personality != "" {
				sb.WriteString(fmt.Sprintf("Personality: %s\n", n.Personality))
			}
			if n.NobleRank != "" {
				sb.WriteString(fmt.Sprintf("Noble rank: %s", n.NobleRank))
				if n.LiegeID != "" {
					if liege := w.FindNPCByID(n.LiegeID); liege != nil {
						sb.WriteString(fmt.Sprintf(" | Liege: %s", liege.Name))
					}
				}
				if len(n.VassalIDs) > 0 {
					sb.WriteString(fmt.Sprintf(" | Vassals: %d", len(n.VassalIDs)))
				}
				sb.WriteString("\n")
			}
			if n.TerritoryID != "" {
				tName := n.TerritoryID
				if t := w.TerritoryByID(n.TerritoryID); t != nil {
					tName = t.Name
				}
				sb.WriteString(fmt.Sprintf("Territory: %s\n", tName))
			}
			if n.FactionID != "" {
				sb.WriteString(fmt.Sprintf("Faction: %s\n", n.FactionID))
			}
			if ctx.NPCTiers != nil {
				if tier, ok := ctx.NPCTiers[n.ID]; ok {
					labels := map[int]string{1: "Tier 1 (key figure)", 2: "Tier 2 (notable)", 3: "Tier 3 (commoner)"}
					if label, ok := labels[tier]; ok {
						sb.WriteString(fmt.Sprintf("Importance: %s\n", label))
					}
				}
			}

			// Vitals & Mood
			sb.WriteString(fmt.Sprintf("\nVitals: HP %d | Happiness %d | Stress %d\n", n.HP, n.Happiness, n.Stress))
			sb.WriteString(fmt.Sprintf("Needs: Hunger %.0f | Thirst %.0f | Fatigue %.0f | Social %.0f\n",
				n.Needs.Hunger, n.Needs.Thirst, n.Needs.Fatigue, n.Needs.SocialNeed))
			sb.WriteString(fmt.Sprintf("Mood: %s\n", n.Mood()))

			// Location & Action
			locName := n.LocationID
			if loc := w.LocationByID(n.LocationID); loc != nil {
				locName = loc.Name
			}
			sb.WriteString(fmt.Sprintf("\nLocation: %s | Gold: %d\n", locName, n.GoldCount()))
			if n.CurrentGoal != "" {
				sb.WriteString(fmt.Sprintf("Current goal: %s\n", n.CurrentGoal))
			}
			if n.LastAction != "" {
				sb.WriteString(fmt.Sprintf("Last action: %s\n", n.LastAction))
			}
			if n.Busy {
				sb.WriteString("Status: BUSY\n")
			}

			// Top 3 skills
			if len(n.Skills) > 0 {
				type skillEntry struct {
					name  string
					level float64
				}
				var skills []skillEntry
				for k, v := range n.Skills {
					skills = append(skills, skillEntry{k, v})
				}
				sort.Slice(skills, func(i, j int) bool { return skills[i].level > skills[j].level })
				sb.WriteString("\nTop skills: ")
				limit := 3
				if len(skills) < limit {
					limit = len(skills)
				}
				var parts []string
				for _, s := range skills[:limit] {
					parts = append(parts, fmt.Sprintf("%s (%.1f)", s.name, s.level))
				}
				sb.WriteString(strings.Join(parts, ", ") + "\n")
			}

			// Equipment summary
			var equipped []string
			if n.Equipment.Weapon != nil {
				equipped = append(equipped, fmt.Sprintf("weapon: %s", n.Equipment.Weapon.Name))
			}
			if n.Equipment.Armor != nil {
				equipped = append(equipped, fmt.Sprintf("armor: %s", n.Equipment.Armor.Name))
			}
			if len(equipped) > 0 {
				sb.WriteString(fmt.Sprintf("Equipment: %s\n", strings.Join(equipped, ", ")))
			}

			// Recent 5 memories
			if ctx.Memory != nil {
				recent := ctx.Memory.Recent(n.ID, 5)
				if len(recent) > 0 {
					sb.WriteString("\nRecent memories:\n")
					for _, e := range recent {
						time := e.Time
						if time == "" {
							time = "?"
						}
						sb.WriteString(fmt.Sprintf("  [%s] %s\n", time, e.Text))
					}
				}

				// Top 3 defining (highest importance) memories
				important := ctx.Memory.HighestImportance(n.ID, 3)
				if len(important) > 0 {
					sb.WriteString("\nDefining memories:\n")
					for _, e := range important {
						time := e.Time
						if time == "" {
							time = "?"
						}
						sb.WriteString(fmt.Sprintf("  [%s] (importance: %.1f) %s\n", time, e.Importance, e.Text))
					}
				}
			}

			// Relationships
			if ctx.Relationships != nil {
				relStr := ctx.Relationships.FormatForLLM(n.ID)
				if relStr != "" {
					sb.WriteString("\n" + relStr + "\n")
				}
			}

			return sb.String(), nil
		},
	}
}

func checkResourcesTool() *ToolDef {
	return &ToolDef{
		Name:        "check_resources",
		Description: "Check resource levels at all locations that have resources.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			w := ctx.World
			var sb strings.Builder
			for _, l := range w.Locations {
				if l.Resources == nil || len(l.Resources) == 0 {
					continue
				}
				var rParts []string
				for k, v := range l.Resources {
					mx := 0
					if l.MaxResources != nil {
						mx = l.MaxResources[k]
					}
					rParts = append(rParts, fmt.Sprintf("%s: %d/%d", k, v, mx))
				}
				sb.WriteString(fmt.Sprintf("- [id=%s] %s (%s): %s\n", l.ID, l.Name, l.Type, strings.Join(rParts, ", ")))
			}
			if sb.Len() == 0 {
				return "No resource locations.", nil
			}
			return sb.String(), nil
		},
	}
}

func recentEventsTool() *ToolDef {
	return &ToolDef{
		Name:        "recent_events",
		Description: "Get the most recent event log entries from the world. Optionally filter by territory_id.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"count":{"type":"integer","description":"Number of recent events to retrieve (default 50)"},"territory_id":{"type":"string","description":"Optional territory ID to filter events."}}}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			var params struct {
				Count       int    `json:"count"`
				TerritoryID string `json:"territory_id"`
			}
			json.Unmarshal(args, &params)
			if params.Count <= 0 {
				params.Count = 50
			}

			w := ctx.World
			var filtered []world.EventEntry
			for i := len(w.EventLog) - 1; i >= 0 && len(filtered) < params.Count; i-- {
				e := w.EventLog[i]
				if params.TerritoryID != "" && e.TerritoryID != params.TerritoryID && e.TerritoryID != "" {
					continue
				}
				filtered = append(filtered, e)
			}

			if len(filtered) == 0 {
				return "No events yet.", nil
			}

			// Reverse to chronological order
			for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}

			var sb strings.Builder
			for _, e := range filtered {
				sb.WriteString(fmt.Sprintf("[%s] %s\n", e.Time, e.Text))
			}
			return sb.String(), nil
		},
	}
}

func checkEnemiesTool() *ToolDef {
	return &ToolDef{
		Name:        "check_enemies",
		Description: "List all living enemies, their locations, and density per location.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			w := ctx.World
			enemies := w.AliveEnemies()
			if len(enemies) == 0 {
				return "No enemies present.", nil
			}

			var sb strings.Builder
			locCounts := make(map[string]int)
			for _, e := range enemies {
				loc := w.LocationByID(e.LocationID)
				locName := e.LocationID
				if loc != nil {
					locName = loc.Name
				}
				sb.WriteString(fmt.Sprintf("- %s (%s, HP: %d/%d, loc: %s)\n",
					e.Name, e.Category, e.HP, e.MaxHP, locName))
				locCounts[locName]++
			}
			sb.WriteString("\nDensity: ")
			var parts []string
			for loc, cnt := range locCounts {
				parts = append(parts, fmt.Sprintf("%s: %d", loc, cnt))
			}
			sb.WriteString(strings.Join(parts, ", "))

			sb.WriteString(fmt.Sprintf("\nAvailable enemy types: %s", strings.Join(enemyTypeNames(), ", ")))
			return sb.String(), nil
		},
	}
}

func enemyTypeNames() []string {
	var names []string
	for k := range enemy.Templates {
		names = append(names, k)
	}
	return names
}

func checkFactionsTool() *ToolDef {
	return &ToolDef{
		Name:        "check_factions",
		Description: "Get details on all factions: members, treasury, goals, contracts.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			w := ctx.World
			if len(w.Factions) == 0 {
				return "No factions yet.", nil
			}

			var sb strings.Builder
			for _, f := range w.Factions {
				sb.WriteString(fmt.Sprintf("- %s (%s, leader: %s, %d members, treasury: %d gold)\n",
					f.Name, f.Type, f.LeaderName, len(f.MemberIDs), f.Treasury))
				if f.Goal != "" {
					sb.WriteString(fmt.Sprintf("  Goal: %s\n", f.Goal))
				}
			}
			return sb.String(), nil
		},
	}
}

func readPrayersTool() *ToolDef {
	return &ToolDef{
		Name:        "read_prayers",
		Description: "Read recent prayers made at the shrine by villagers seeking divine help.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			w := ctx.World
			if len(w.RecentPrayers) == 0 {
				return "No prayers recently.", nil
			}

			var sb strings.Builder
			for _, p := range w.RecentPrayers {
				sb.WriteString(fmt.Sprintf("- %s prayed: \"%s\" (HP: %d, stress: %d, hunger: %.0f) at %s\n",
					p.NpcName, p.Prayer, p.HP, p.Stress, p.Hunger, p.Time))
			}
			return sb.String(), nil
		},
	}
}

func godMemoriesTool() *ToolDef {
	return &ToolDef{
		Name:        "god_memories",
		Description: "Recall your past decisions. Optionally filter by category: spawn, dream, vision, omen, resource, growth, weather. Leave empty for all.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"category":{"type":"string","description":"Filter by category: spawn, dream, vision, omen, resource, growth, weather. Leave empty for all."}}}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			if ctx.Memory == nil {
				return "No memory store available.", nil
			}
			var params struct {
				Category string `json:"category"`
			}
			json.Unmarshal(args, &params)

			godID := memory.GodEntityID

			if params.Category != "" {
				// Map short names to full subcategory constants
				catMap := map[string]string{
					"spawn":    memory.CatGodSpawn,
					"dream":    memory.CatGodDream,
					"vision":   memory.CatGodVision,
					"omen":     memory.CatGodOmen,
					"resource": memory.CatGodResource,
					"crisis":   memory.CatGodCrisis,
					"growth":   memory.CatGodGrowth,
					"weather":  memory.CatGodWeather,
				}
				fullCat, ok := catMap[params.Category]
				if !ok {
					return fmt.Sprintf("Unknown category %q. Use: spawn, dream, vision, omen, resource, growth, weather.", params.Category), nil
				}
				entries := ctx.Memory.ByCategory(godID, fullCat)
				if len(entries) == 0 {
					return fmt.Sprintf("No %s decisions recorded.", params.Category), nil
				}
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("%s decisions (%d total):\n", params.Category, len(entries)))
				for i, e := range entries {
					time := e.Time
					if time == "" {
						time = "?"
					}
					sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, time, e.Text))
				}
				return sb.String(), nil
			}

			// Default: show strategic window (10 recent + 5 high-impact, deduped)
			recent := ctx.Memory.Recent(godID, 10)
			important := ctx.Memory.HighestImportance(godID, 5)

			var sb strings.Builder
			seen := make(map[string]bool)

			if len(recent) > 0 {
				sb.WriteString("Recent decisions:\n")
				for i, e := range recent {
					seen[e.Text] = true
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
					sb.WriteString(fmt.Sprintf("  %d. [%s] %s%s\n", i+1, time, prefix, e.Text))
				}
			}

			var highImpact []memory.Entry
			for _, e := range important {
				if !seen[e.Text] {
					highImpact = append(highImpact, e)
				}
			}
			if len(highImpact) > 0 {
				sb.WriteString("\nHigh-impact past decisions:\n")
				for _, e := range highImpact {
					time := e.Time
					if time == "" {
						time = "?"
					}
					sb.WriteString(fmt.Sprintf("  - [%s] (importance: %.1f) %s\n", time, e.Importance, e.Text))
				}
			}

			total := len(ctx.Memory.All(godID))
			sb.WriteString(fmt.Sprintf("\n(Total memories: %d/%d)\n", total, memory.GodMaxMemories))

			if sb.Len() == 0 {
				return "No previous decisions recorded.", nil
			}
			return sb.String(), nil
		},
	}
}

func godTrendsTool() *ToolDef {
	return &ToolDef{
		Name:        "world_trends",
		Description: "View world metric trends over recent days: population, economy, hunger, stress, happiness. Use this to understand whether your interventions are working.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			if ctx.Trends == nil {
				return "Trend tracking not available.", nil
			}
			return ctx.Trends.FormatForGod(), nil
		},
	}
}

func territoryDetailTool() *ToolDef {
	return &ToolDef{
		Name:        "territory_detail",
		Description: "Get detailed information about a specific territory: ruler, nobles, professions, resources, enemies, crisis NPCs, cities, treasury, laws, allies, enemies.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"territory_id":{"type":"string","description":"The territory ID to inspect"}},"required":["territory_id"]}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			var params struct {
				TerritoryID string `json:"territory_id"`
			}
			json.Unmarshal(args, &params)
			if params.TerritoryID == "" {
				return "territory_id is required.", nil
			}

			w := ctx.World
			t := w.TerritoryByID(params.TerritoryID)
			if t == nil {
				return fmt.Sprintf("Territory %q not found.", params.TerritoryID), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("=== %s (%s) ===\n", t.Name, t.ID))
			sb.WriteString(fmt.Sprintf("Type: %s | Biome: %s | Center: (%d,%d)\n", t.Type, t.BiomeHint, t.CenterX, t.CenterY))
			sb.WriteString(fmt.Sprintf("Treasury: %d gold | Tax rate: %.0f%%\n", t.Treasury, t.TaxRate*100))

			// Ruler
			if t.RulerID != "" {
				ruler := w.FindNPCByID(t.RulerID)
				if ruler != nil {
					sb.WriteString(fmt.Sprintf("Ruler: %s (%s, HP: %d, gold: %d)\n", ruler.Name, ruler.NobleRank, ruler.HP, ruler.GoldCount()))
				} else {
					sb.WriteString(fmt.Sprintf("Ruler: %s (missing)\n", t.RulerID))
				}
			} else {
				sb.WriteString("Ruler: (none)\n")
			}

			// Laws
			if len(t.Laws) > 0 {
				sb.WriteString(fmt.Sprintf("Laws: %s\n", strings.Join(t.Laws, ", ")))
			}

			// Allies / Enemies
			if len(t.Allies) > 0 {
				sb.WriteString(fmt.Sprintf("Allies: %s\n", strings.Join(t.Allies, ", ")))
			}
			if len(t.Enemies) > 0 {
				sb.WriteString(fmt.Sprintf("Enemies: %s\n", strings.Join(t.Enemies, ", ")))
			}

			// Cities
			if len(t.CityIDs) > 0 {
				var cityNames []string
				for _, cid := range t.CityIDs {
					loc := w.LocationByID(cid)
					if loc != nil {
						cityNames = append(cityNames, fmt.Sprintf("%s (%s)", loc.Name, cid))
					} else {
						cityNames = append(cityNames, cid)
					}
				}
				sb.WriteString(fmt.Sprintf("Cities: %s\n", strings.Join(cityNames, ", ")))
			}

			// NPCs in this territory
			alive := w.AliveNPCs()
			var nobles []*npc.NPC
			profCounts := make(map[string]int)
			var crisisNPCs []*npc.NPC
			npcCount := 0

			for _, n := range alive {
				if n.TerritoryID != t.ID {
					continue
				}
				npcCount++
				profCounts[n.Profession]++
				if n.NobleRank != "" {
					nobles = append(nobles, n)
				}
				if n.HP < 20 || n.Needs.Hunger < 10 {
					crisisNPCs = append(crisisNPCs, n)
				}
			}

			sb.WriteString(fmt.Sprintf("\nPopulation: %d\n", npcCount))

			// Nobles
			if len(nobles) > 0 {
				sb.WriteString("Nobles:\n")
				for _, n := range nobles {
					sb.WriteString(fmt.Sprintf("  - %s (%s, gold: %d, mood: %s, vassals: %d)\n",
						n.Name, n.NobleRank, n.GoldCount(), n.Mood(), len(n.VassalIDs)))
				}
			}

			// Professions
			if len(profCounts) > 0 {
				sb.WriteString("Professions: ")
				var parts []string
				for p, c := range profCounts {
					parts = append(parts, fmt.Sprintf("%s: %d", p, c))
				}
				sb.WriteString(strings.Join(parts, ", ") + "\n")
			}

			// Resource locations in this territory
			sb.WriteString("\nResource Locations:\n")
			resCount := 0
			// Build a set of location IDs in this territory
			terrLocIDs := make(map[string]bool)
			for _, l := range w.Locations {
				if l.TerritoryID != t.ID {
					continue
				}
				terrLocIDs[l.ID] = true
				if l.Resources == nil || len(l.Resources) == 0 {
					continue
				}
				resCount++
				var rParts []string
				for k, v := range l.Resources {
					mx := 0
					if l.MaxResources != nil {
						mx = l.MaxResources[k]
					}
					rParts = append(rParts, fmt.Sprintf("%s: %d/%d", k, v, mx))
				}
				sb.WriteString(fmt.Sprintf("  - %s (%s): %s\n", l.Name, l.ID, strings.Join(rParts, ", ")))
			}
			if resCount == 0 {
				sb.WriteString("  (none)\n")
			}

			// Enemies at territory locations
			enemies := w.AliveEnemies()
			var terrEnemies []string
			for _, e := range enemies {
				if terrLocIDs[e.LocationID] {
					loc := w.LocationByID(e.LocationID)
					locName := e.LocationID
					if loc != nil {
						locName = loc.Name
					}
					terrEnemies = append(terrEnemies, fmt.Sprintf("  - %s (%s, HP: %d/%d, at %s)", e.Name, e.Category, e.HP, e.MaxHP, locName))
				}
			}
			if len(terrEnemies) > 0 {
				sb.WriteString(fmt.Sprintf("\nEnemies (%d):\n", len(terrEnemies)))
				for _, e := range terrEnemies {
					sb.WriteString(e + "\n")
				}
			}

			// Crisis NPCs
			if len(crisisNPCs) > 0 {
				sb.WriteString(fmt.Sprintf("\nCrisis NPCs (%d):\n", len(crisisNPCs)))
				for _, n := range crisisNPCs {
					flags := ""
					if n.HP < 20 {
						flags += fmt.Sprintf(" HP:%d", n.HP)
					}
					if n.Needs.Hunger < 10 {
						flags += fmt.Sprintf(" hunger:%.0f", n.Needs.Hunger)
					}
					sb.WriteString(fmt.Sprintf("  !! %s (%s)%s\n", n.Name, n.Profession, flags))
				}
			}

			// Append intelligence brief if available
			if ctx.TerritoryBriefs != nil {
				if brief, ok := ctx.TerritoryBriefs[params.TerritoryID]; ok {
					sb.WriteString("\n--- Intelligence Brief ---\n")
					sb.WriteString(brief.FormatForGod())
				}
			}

			return sb.String(), nil
		},
	}
}

func nobleHierarchyTool() *ToolDef {
	return &ToolDef{
		Name:        "noble_hierarchy",
		Description: "Display the feudal hierarchy tree: kings/queens → dukes → counts → barons → knights, showing gold, territory, vassals, and combat stats.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(ctx *AgentContext, _ json.RawMessage) (string, error) {
			w := ctx.World
			alive := w.AliveNPCs()

			// Collect all nobles
			nobles := make(map[string]*npc.NPC)
			var topLevel []*npc.NPC // Nobles with no liege

			for _, n := range alive {
				if n.NobleRank == "" {
					continue
				}
				nobles[n.ID] = n
			}

			if len(nobles) == 0 {
				return "No nobles in the realm.", nil
			}

			// Identify top-level nobles (no liege or liege not alive/noble)
			for _, n := range nobles {
				if n.LiegeID == "" || nobles[n.LiegeID] == nil {
					topLevel = append(topLevel, n)
				}
			}

			// Sort top-level: kings/queens first, then by name
			rankOrder := map[string]int{
				"king": 0, "queen": 0,
				"prince": 1, "princess": 1,
				"duke": 2,
				"count": 3,
				"baron": 4,
				"knight": 5,
			}
			sort.Slice(topLevel, func(i, j int) bool {
				ri, rj := rankOrder[topLevel[i].NobleRank], rankOrder[topLevel[j].NobleRank]
				if ri != rj {
					return ri < rj
				}
				return topLevel[i].Name < topLevel[j].Name
			})

			var sb strings.Builder
			sb.WriteString("=== Feudal Hierarchy ===\n")

			// Recursive tree renderer
			var renderNoble func(n *npc.NPC, prefix string, isLast bool)
			renderNoble = func(n *npc.NPC, prefix string, isLast bool) {
				// Connector
				connector := "├─ "
				if isLast {
					connector = "└─ "
				}
				if prefix == "" {
					connector = ""
				}

				// Build the noble's info line
				terrName := ""
				if n.TerritoryID != "" {
					if t := w.TerritoryByID(n.TerritoryID); t != nil {
						terrName = t.Name
					} else {
						terrName = n.TerritoryID
					}
				}

				info := fmt.Sprintf("%s %s", strings.Title(n.NobleRank), n.Name)
				details := []string{fmt.Sprintf("HP: %d", n.HP), fmt.Sprintf("gold: %d", n.GoldCount())}
				if terrName != "" {
					details = append(details, fmt.Sprintf("territory: %s", terrName))
				}

				// Find vassals among alive nobles
				var vassals []*npc.NPC
				for _, vid := range n.VassalIDs {
					if v, ok := nobles[vid]; ok {
						vassals = append(vassals, v)
					}
				}

				if len(vassals) == 0 && (n.NobleRank == "duke" || n.NobleRank == "count" || n.NobleRank == "baron") {
					details = append(details, "0 vassals ⚠")
				} else if n.NobleRank == "knight" {
					details = append(details, fmt.Sprintf("combat: %d", n.Stats.Strength))
				}

				sb.WriteString(fmt.Sprintf("%s%s%s (%s)\n", prefix, connector, info, strings.Join(details, ", ")))

				// Render vassals
				sort.Slice(vassals, func(i, j int) bool {
					ri, rj := rankOrder[vassals[i].NobleRank], rankOrder[vassals[j].NobleRank]
					if ri != rj {
						return ri < rj
					}
					return vassals[i].Name < vassals[j].Name
				})

				childPrefix := prefix
				if prefix != "" {
					if isLast {
						childPrefix += "    "
					} else {
						childPrefix += "│   "
					}
				}

				for i, v := range vassals {
					renderNoble(v, childPrefix, i == len(vassals)-1)
				}
			}

			for i, n := range topLevel {
				renderNoble(n, "", i == len(topLevel)-1)
			}

			return sb.String(), nil
		},
	}
}

func influenceMapTool() *ToolDef {
	return &ToolDef{
		Name:        "influence_map",
		Description: "Show the top 10 most influential NPCs in the realm (or a specific territory). Influence is based on social connections, noble rank, and faction leadership.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"territory_id":{"type":"string","description":"Optional territory ID to filter. Omit for kingdom-wide."}}}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			if ctx.SocialGraph == nil {
				return "Social graph not available.", nil
			}
			var params struct {
				TerritoryID string `json:"territory_id"`
			}
			json.Unmarshal(args, &params)
			result := ctx.SocialGraph.FormatInfluenceMap(params.TerritoryID)

			// Append tier summary if available
			if ctx.NPCTiers != nil {
				tierCounts := map[int]int{}
				for _, t := range ctx.NPCTiers {
					tierCounts[t]++
				}
				result += fmt.Sprintf("\n\nNPC Tiering: %d key figures (T1), %d notable (T2), %d commoners (T3)",
					tierCounts[1], tierCounts[2], tierCounts[3])
			}
			return result, nil
		},
	}
}

func socialClustersTool() *ToolDef {
	return &ToolDef{
		Name:        "social_clusters",
		Description: "Show community clusters (groups of positively-connected NPCs), faction overlap within clusters, and bridge NPCs who connect separate groups.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"territory_id":{"type":"string","description":"Optional territory ID to filter. Omit for kingdom-wide."}}}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			if ctx.SocialGraph == nil {
				return "Social graph not available.", nil
			}
			var params struct {
				TerritoryID string `json:"territory_id"`
			}
			json.Unmarshal(args, &params)
			return ctx.SocialGraph.FormatSocialClusters(params.TerritoryID), nil
		},
	}
}

func tensionAnalysisTool() *ToolDef {
	return &ToolDef{
		Name:        "tension_analysis",
		Description: "Show hostile relationships sorted by severity. Highlights cross-faction conflicts. Use this to identify where to intervene with dreams or omens.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"territory_id":{"type":"string","description":"Optional territory ID to filter. Omit for kingdom-wide."}}}`),
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			if ctx.SocialGraph == nil {
				return "Social graph not available.", nil
			}
			var params struct {
				TerritoryID string `json:"territory_id"`
			}
			json.Unmarshal(args, &params)
			return ctx.SocialGraph.FormatTensionAnalysis(params.TerritoryID), nil
		},
	}
}

func godActTool() *ToolDef {
	return &ToolDef{
		Name:        "god_act",
		Description: `Commit divine actions. Each action needs an action type, reason, and params.
Your influence is INDIRECT ONLY — no direct NPC manipulation.

Available actions and their required params:
- send_dream: {target_name (exact NPC name), dream_content (vivid dream narrative), urgency (low/med/high)}. Dreams become NPC memories that shape decisions.
- send_vision: {target_name (exact NPC name, must have recent prayer), vision_content (symbolic vision), divine_message (optional)}. Consumes the prayer.
- send_omen: {omen_description (atmospheric description)}. All spiritually sensitive NPCs receive this memory.
- spawn_npc: {name (must be a unique proper name, NOT "Stranger"), profession, age}
- spawn_enemy: {enemy_type (one of: wolf, bear, wild_boar, giant_spider, cave_bat, serpent, bandit, robber, pirate, brigand, cultist, or any introduced enemy), location_id (optional)}
- introduce_enemy: {name, target_name (NPC warned via dream, optional), category (beast/human), hp, strength, agility, defense, locations ([location types where it spawns]), items ([loot item names])}
- advance_weather: {weather_type (one of: clear, cloudy, rain, storm)}. Rain/storm fills wells with water.
- replenish_resources: {location_id (use exact location ID), resources: {resource_name: amount}}. Natural resources only: stone, iron_ore, wood, game, berries, herbs, fish, clay.
- create_location: {name (must be a descriptive proper name, NOT "New Place"), type, description}. Natural features only: forest, mine, well, dock, garden.
- introduce_item: {name, target_name (NPC who discovers it via dream), weight, category (food/drink/medicine/material/trade_good/curiosity), slot (weapon/armor/bag, optional), effects ({hunger_restore: N, heal_hp_min: N, heal_hp_max: N, weapon_bonus: N, armor_bonus: N, sobriety: N, stress: N, happiness: N, thirst: N})}
- introduce_technique: {name, target_name (NPC who discovers it via dream), description, bonus_type}
- introduce_profession: {name, target_name (NPC who discovers it via dream), primary_skill, description}
- expand_grid: {amount (2-6)}
- stir_dreams: {territory_id, dream_theme (vivid dream narrative), urgency (low/med/high)}. Queue the same dream to up to 5 random commoners in a territory.
- send_prophecy: {prophecy_text, territory_id (optional — omit for kingdom-wide)}. Creates a shared prophecy event + dreams to spiritually sensitive NPCs.
- do_nothing: {}`,
		Parameters: json.RawMessage(`{"type":"object","properties":{
			"analysis":{"type":"string","description":"2-3 sentences: what is the biggest problem and what will help"},
			"actions":{"type":"array","items":{"type":"object","properties":{
				"action":{"type":"string"},
				"reason":{"type":"string"},
				"params":{"type":"object"}
			},"required":["action","reason"]},"minItems":1}
		},"required":["analysis","actions"]}`),
		IsTerminal: true,
		Handler: func(ctx *AgentContext, args json.RawMessage) (string, error) {
			return "GOD_ACTION_COMMITTED", nil
		},
	}
}
