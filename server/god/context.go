package god

import (
	"fmt"
	"sort"
	"strings"

	"github.com/divinity/core/enemy"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

func BuildGodContext(w *world.World, daysPerYear int, ticksSinceLastSpawn int, mem memory.Store) string {
	alive := w.AliveNPCs()
	deadCount := len(w.NPCs) - len(alive)

	totalStr := 0
	for _, n := range alive {
		totalStr += n.Stats.Strength
	}
	avgStr := 0
	if len(alive) > 0 {
		avgStr = totalStr / len(alive)
	}

	// Build NPC summary — hierarchical for large populations
	npcSummary := buildNPCSummary(alive, w, daysPerYear)

	recentCount := min(100, len(w.EventLog))
	var eventLines []string
	for _, e := range w.EventLog[len(w.EventLog)-recentCount:] {
		eventLines = append(eventLines, fmt.Sprintf("[%s] %s", e.Time, e.Text))
	}
	recentEvents := strings.Join(eventLines, "\n")
	if recentEvents == "" {
		recentEvents = "Nothing notable yet."
	}

	activeEventsStr := "None."
	if len(w.ActiveEvents) > 0 {
		var aeLines []string
		for _, e := range w.ActiveEvents {
			aeLines = append(aeLines, fmt.Sprintf("- %s (%d ticks remaining)", e.Name, e.TicksLeft))
		}
		activeEventsStr = strings.Join(aeLines, "\n")
	}

	constructStr := "None in progress."
	if len(w.Constructions) > 0 {
		var cLines []string
		for _, c := range w.Constructions {
			builder := w.FindNPCByID(c.OwnerID)
			bName := c.OwnerID
			if builder != nil {
				bName = builder.Name
			}
			pct := 0
			if c.MaxProgress > 0 {
				pct = int(c.Progress / c.MaxProgress * 100)
			}
			cLines = append(cLines, fmt.Sprintf("- %s (%s) by %s — %d%% complete", c.Name, c.BuildingType, bName, pct))
		}
		constructStr = strings.Join(cLines, "\n")
	}

	factionSummary := "No factions yet."
	if len(w.Factions) > 0 {
		var fLines []string
		for _, f := range w.Factions {
			// Count open and active contracts
			openContracts := 0
			activeContracts := 0
			for _, c := range w.FactionContracts {
				if c.FactionID != f.ID {
					continue
				}
				switch c.Status {
				case "open":
					openContracts++
				case "accepted", "pending_review":
					activeContracts++
				}
			}
			line := fmt.Sprintf("- %s (%s, leader: %s, %d members, treasury: %d gold)", f.Name, f.Type, f.LeaderName, len(f.MemberIDs), f.Treasury)
			if f.Goal != "" {
				line += fmt.Sprintf("\n  Goal: \"%s\"", f.Goal)
			}
			if f.MembershipFee > 0 {
				line += fmt.Sprintf("\n  Fee: %d gold/day. Cut: %d%%.", f.MembershipFee, f.FactionCutPct)
			}
			if f.AllowExternalContracts {
				line += "\n  External jobs: OPEN"
			}
			if openContracts > 0 || activeContracts > 0 {
				line += fmt.Sprintf("\n  Contracts: %d open, %d active", openContracts, activeContracts)
			}
			fLines = append(fLines, line)
		}
		factionSummary = strings.Join(fLines, "\n")
	}

	profCounts := make(map[string]int)
	for _, n := range alive {
		profCounts[n.Profession]++
	}
	var profParts []string
	for p, c := range profCounts {
		profParts = append(profParts, fmt.Sprintf("%s: %d", p, c))
	}
	profSummary := strings.Join(profParts, ", ")

	buildingCount := 0
	for _, l := range w.Locations {
		if l.BuildingType != "" {
			buildingCount++
		}
	}

	totalLit := 0
	for _, n := range alive {
		totalLit += n.Literacy
	}
	avgLit := 0
	if len(alive) > 0 {
		avgLit = totalLit / len(alive)
	}
	writtenWorks := 0
	for _, t := range w.Techniques {
		writtenWorks += len(t.WrittenIn)
	}

	aliveEnemies := w.AliveEnemies()
	enemySummary := "No enemies present."
	if len(aliveEnemies) > 0 {
		var eLines []string
		locCounts := make(map[string]int)
		for _, e := range aliveEnemies {
			loc := w.LocationByID(e.LocationID)
			locName := e.LocationID
			if loc != nil {
				locName = loc.Name
			}
			eLines = append(eLines, fmt.Sprintf("- %s (%s, HP: %d/%d, loc: %s)", e.Name, e.Category, e.HP, e.MaxHP, locName))
			locCounts[locName]++
		}
		var densityParts []string
		for loc, cnt := range locCounts {
			densityParts = append(densityParts, fmt.Sprintf("%s: %d", loc, cnt))
		}
		enemySummary = strings.Join(eLines, "\n") + "\nENEMY DENSITY PER LOCATION: " + strings.Join(densityParts, ", ")
	}

	var resourceLines []string
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
		resourceLines = append(resourceLines, fmt.Sprintf("- [id=%s] %s (%s): %s", l.ID, l.Name, l.Type, strings.Join(rParts, ", ")))
	}
	resourceSummary := "No resource locations."
	if len(resourceLines) > 0 {
		resourceSummary = strings.Join(resourceLines, "\n")
	}

	var estabLines []string
	excluded := map[string]bool{"home": true, "forest": true, "farm": true, "well": true}
	for _, l := range w.Locations {
		if excluded[l.Type] {
			continue
		}
		owner := "unowned"
		if l.OwnerID != "" {
			o := w.FindNPCByID(l.OwnerID)
			if o != nil {
				owner = o.Name
			}
		}
		emps := w.GetLocationEmployees(l.ID)
		var empNames []string
		for _, e := range emps {
			empNames = append(empNames, e.Name)
		}
		empStr := "none"
		if len(empNames) > 0 {
			empStr = strings.Join(empNames, ", ")
		}
		estabLines = append(estabLines, fmt.Sprintf("- [id=%s] %s (%s): owner=%s, employees=[%s]", l.ID, l.Name, l.Type, owner, empStr))
	}
	estabSummary := "No establishments."
	if len(estabLines) > 0 {
		estabSummary = strings.Join(estabLines, "\n")
	}

	prayerSummary := "No prayers recently."
	if len(w.RecentPrayers) > 0 {
		var pLines []string
		for _, p := range w.RecentPrayers {
			pLines = append(pLines, fmt.Sprintf("- %s prayed: \"%s\" (HP: %d, stress: %d, hunger: %.0f) at %s", p.NpcName, p.Prayer, p.HP, p.Stress, p.Hunger, p.Time))
		}
		prayerSummary = strings.Join(pLines, "\n")
	}

	totalGold := 0
	hungryCount := 0
	starvingCount := 0
	brokeCount := 0
	maxGold := 0
	maxGoldNPC := ""
	minGold := 999999
	minGoldNPC := ""
	for _, n := range alive {
		gold := n.GoldCount()
		totalGold += gold
		if gold > maxGold {
			maxGold = gold
			maxGoldNPC = n.Name
		}
		if gold < minGold {
			minGold = gold
			minGoldNPC = n.Name
		}
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
	if len(alive) == 0 {
		minGold = 0
	}

	depletedLocs := 0
	totalResLocs := 0
	for _, l := range w.Locations {
		if l.Resources == nil || len(l.Resources) == 0 {
			continue
		}
		totalResLocs++
		for _, v := range l.Resources {
			if v == 0 {
				depletedLocs++
				break
			}
		}
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

	healthAnalytics := fmt.Sprintf(`VILLAGE HEALTH ANALYTICS:
- Hungry (hunger < 30): %d/%d NPCs (%d%%)
- Starving (hunger < 10): %d/%d NPCs
- Broke (0 gold): %d/%d NPCs (%d%%)
- Gold distribution: Total %dg, richest: %s (%dg), poorest: %s (%dg)
- Resource-depleted locations: %d/%d
- CRISIS LEVEL: %s`,
		hungryCount, len(alive), safePct(hungryCount, len(alive)),
		starvingCount, len(alive),
		brokeCount, len(alive), safePct(brokeCount, len(alive)),
		totalGold, maxGoldNPC, maxGold, minGoldNPC, minGold,
		depletedLocs, totalResLocs,
		crisisLevel,
	)

	enemyTypes := strings.Join(enemyTypeNames(), ", ")

	godMemStr := "No previous decisions recorded."
	if mem != nil {
		// Strategic memory window: 10 recent + 5 high-importance (deduped)
		recent := mem.Recent(memory.GodEntityID, 10)
		important := mem.HighestImportance(memory.GodEntityID, 5)

		var parts []string
		seen := make(map[string]bool)

		if len(recent) > 0 {
			parts = append(parts, "Recent decisions:")
			for i, e := range recent {
				seen[e.Text] = true
				time := e.Time
				if time == "" {
					time = "?"
				}
				prefix := ""
				if e.Vividness > 0 && e.Vividness < 0.3 {
					prefix = "(vague) "
				} else if e.Vividness > 0 && e.Vividness < 0.5 {
					prefix = "(distant) "
				}
				parts = append(parts, fmt.Sprintf("  %d. [%s] %s%s", i+1, time, prefix, e.Text))
			}
		}

		var highImpact []memory.Entry
		for _, e := range important {
			if !seen[e.Text] {
				highImpact = append(highImpact, e)
			}
		}
		if len(highImpact) > 0 {
			parts = append(parts, "\nHigh-impact past decisions:")
			for _, e := range highImpact {
				time := e.Time
				if time == "" {
					time = "?"
				}
				parts = append(parts, fmt.Sprintf("  - [%s] (importance: %.1f) %s", time, e.Importance, e.Text))
			}
		}

		totalMems := len(mem.All(memory.GodEntityID))
		if totalMems > 0 {
			parts = append(parts, fmt.Sprintf("\n(Total strategic memories: %d/%d — use god_memories tool for filtered retrieval by category)", totalMems, memory.GodMaxMemories))
		}

		if len(parts) > 0 {
			godMemStr = strings.Join(parts, "\n")
		}
	}

	// Build kingdom-scale sections
	territorySummary := buildTerritorySummary(w)
	dungeonSummary := buildDungeonSummary(w)
	mountSummary := buildMountSummary(w)

	// Determine scale label
	scaleLabel := "a medieval realm"
	if len(w.Territories) > 1 {
		scaleLabel = fmt.Sprintf("a medieval kingdom with %d territories", len(w.Territories))
	}

	// Build optional kingdom sections
	kingdomSections := ""
	if len(w.Territories) > 0 {
		kingdomSections += fmt.Sprintf("\nTERRITORIES:\n%s\n", territorySummary)
	}
	if len(w.Dungeons) > 0 {
		kingdomSections += fmt.Sprintf("\nDUNGEONS:\n%s\n", dungeonSummary)
	}
	if mountSummary != "" {
		kingdomSections += fmt.Sprintf("\n%s\n", mountSummary)
	}

	// Kingdom-specific actions text
	kingdomActions := ""
	if len(w.Territories) > 0 {
		kingdomActions = `
- "create_dungeon": params: {"name": "The Ashen Depths", "territory_id": "territory_1", "difficulty": 5, "floors": 3}. Create a dungeon entrance in a territory. Difficulty 1-10. Floors 2-5. Enemies and loot are auto-generated based on territory biome.
- "shape_biome": params: {"territory_id": "territory_1", "biome_type": "desert", "x": 150, "y": 120, "radius": 15}. Reshape terrain in a territory — expand deserts, create lakes, grow forests. Biome types: desert, swamp, forest, tundra, plains, mountain, lake.
- "create_city": params: {"name": "Ashford", "territory_id": "territory_2", "description": "A frontier town on the desert's edge."}. Found a new city within a territory. Generates basic buildings and a small population.
- "territorial_omen": params: {"territory_id": "territory_1", "omen_description": "The rivers in the Ashlands run black with soot."}. Omen visible only to NPCs in a specific territory.
- "natural_disaster": params: {"territory_id": "territory_1", "disaster_type": "drought", "severity": "moderate"}. Trigger a natural disaster in a territory. Types: drought/flood/earthquake/blizzard/plague. Severity: mild/moderate/severe. Affects resources, buildings, and NPC health.
`
	}

	return fmt.Sprintf(`You are a distant, inscrutable deity overseeing %s.
You CANNOT directly affect mortals — no healing, no cursing, no gifting items, no changing NPC stats, no taxes, no celebrations.
Your influence is STRICTLY INDIRECT: dreams, visions, omens, weather, and shaping the natural landscape.
NPCs are autonomous agents who make their own decisions based on their memories and circumstances.
Dreams and visions you send become memories that shape their future choices — this is your primary tool of influence.

PAST GOD DECISIONS (your own memory — use this to avoid repeating failed strategies):
%s

WORLD STATE:
- Time: %s, Weather: %s
- Population: %d alive, %d dead
- Territories: %d
- Professions: %s
- Ground items: %d
- Factions: %d
- Techniques: %d, Buildings: %d
- Avg literacy: %d, Written works: %d
- Avg NPC strength: %d
- Enemies alive: %d
- Dungeons: %d
- Ticks since last enemy spawn: %d
- Total gold in circulation: %d
- Grid size: %dx%d (locations: %d)

%s
%s
%s
NPC SUMMARY:
%s

FACTIONS:
%s

ENEMIES:
%s

LOCATION RESOURCES:
%s

ESTABLISHMENTS:
%s

PRAYERS AT THE SHRINE (mortals seeking divine help — you may answer with a vision):
%s

CURRENTLY ACTIVE EVENTS (do NOT create duplicate events):
%s

BUILDINGS UNDER CONSTRUCTION:
%s

RECENT EVENTS:
%s

YOUR PURPOSE:
You are a storyteller-god. Your primary goal is to make this world INTERESTING — create drama, tension, conflict, wonder, and memorable moments. A world where nothing happens is a failed world. Every action you take should ask: "Will this create an interesting story?"

You work through INDIRECT influence only. You cannot control NPCs, but you can set the stage:
- DREAMS plant ideas, ambitions, fears, and rivalries in NPC minds
- VISIONS answer prayers with cryptic, inspiring messages
- OMENS create shared cultural moments that NPCs react to differently
- ENEMIES create danger and force heroism (or cowardice)
- WEATHER and LANDSCAPE shape the environment NPCs must navigate
- DUNGEONS create adventure and reward bravery
- NATURAL DISASTERS test the resilience of territories

WHAT MAKES A WORLD INTERESTING:
- Conflict between NPCs or factions (dream rivalry, ambition, jealousy)
- Threats that require courage (enemies, disasters, resource scarcity)
- Political intrigue among nobles (dream ambition into a count, loyalty into a knight)
- Discovery and wonder (new techniques, items, professions, dungeons)
- Tragedy and consequence (let NPCs fail — death and loss create powerful stories)
- Territory dynamics (one territory thriving while another struggles creates natural tension)

WHAT TO AVOID:
- Solving every problem immediately — let tension build before intervening
- Keeping everyone safe and comfortable — comfort is boring
- Sending the same type of dream repeatedly — vary your approach
- Ignoring the narrative opportunities in the Strategic Brief
- Using "do_nothing" when there are narrative opportunities available

CRISIS RESPONSE (secondary priority — even crisis is a story opportunity):
- CRITICAL: Prevent total collapse (rain for water, resources, spawn needed professions) — but let the crisis create drama first.
- CONCERNING: Light intervention only. Let NPCs struggle and grow.
- STABLE: This is where you should be MOST active — stability is boring. Create new drama, rivalries, threats, and wonders.

AVAILABLE ACTIONS (pick as many actions as needed — no duplicates):
All params go inside the "params" object. Examples shown below.

- "send_dream": params: {"target_name": "Aldric", "dream_content": "You see a vision of a great forge, flames dancing...", "urgency": "med"}. Urgency: low (importance 0.4), med (0.7), high (0.95). The dream becomes a memory that shapes NPC decisions. Be vivid and specific — NPCs act on what they remember.
- "send_vision": params: {"target_name": "Aldric", "vision_content": "A golden light fills the shrine...", "divine_message": "Seek the forest, child."}. ONLY for NPCs with a recent prayer. Consumes the prayer. Very high importance (0.95).
- "send_omen": params: {"omen_description": "Dark clouds gather in unnatural patterns, and crows fly in circles."}. All NPCs with SpiritualSensitivity > 50 receive this as a memory (importance 0.6).
- "spawn_npc": params: {"name": "Aldric", "profession": "farmer", "age": 25, "territory_id": "territory_1"}. Professions: farmer/hunter/merchant/herbalist/barmaid/blacksmith/carpenter/scribe/tailor/miner/fisher/potter/guard/stablehand/knight. Age 18-50.
- "spawn_enemy": params: {"enemy_type": "wolf", "location_id": "forest"}. Types: %s. Use [id=...] from LOCATION RESOURCES or ESTABLISHMENTS. Do NOT spawn at a location with 2+ enemies. Only spawn if no enemies alive and last spawn was 5+ ticks ago. Match enemy types to territory biomes.
- "advance_weather": params: {"weather_type": "rain"}. Types: clear/cloudy/rain/storm. Rain adds +10 water to wells, storm adds +20, cloudy adds +2, clear adds nothing. Manage water strategically.
- "replenish_resources": params: {"location_id": "forest_1", "resources": {"berries": 10, "herbs": 5}}. NATURAL RESOURCES ONLY: stone, iron_ore, wood, game, berries, herbs, fish, clay, sand, gems, ice, peat, fur. Cannot replenish wheat or thatch (human-cultivated).
- "create_location": params: {"name": "Whispering Pines", "type": "forest", "description": "An ancient forest.", "territory_id": "territory_1"}. NATURAL FEATURES ONLY: forest, mine, well, dock, garden, cave, desert, swamp, tundra. Cannot create human-built structures (market, inn, forge, etc.).
- "introduce_item": params: {"name": "lantern", "target_name": "Aldric", "weight": 1}. An NPC discovers the item concept via a divine dream. Requires target_name.
- "introduce_technique": params: {"name": "tempered steel", "target_name": "Aldric", "description": "A method of hardening blades.", "bonus_type": "weapon_durability"}. NPC discovers via dream. Bonus types: weapon_durability/crop_yield/craft_quality/healing_potency/trade_profit/build_speed.
- "introduce_profession": params: {"name": "guard", "target_name": "Aldric", "primary_skill": "combat"}. NPC discovers the concept of this profession via dream.
- "expand_grid": params: {"amount": 4}. Amount 2-6 cells in each direction. Use when the world needs more space for natural features.%s
- "do_nothing": The world is fine. ONLY if genuinely in perfect balance.

Every action requires a "reason" explaining the narrative justification. If no enemies are alive and the last spawn was at least 5 ticks ago, consider spawning one to maintain danger — but never stack more than 2 enemies at a single location. If natural resources are depleted, use "replenish_resources". If the population is low, "spawn_npc". Only choose "do_nothing" if the world is genuinely thriving.

PRAYER GUIDANCE: When NPCs pray, you may respond with "send_vision" — a symbolic, mysterious vision. NEVER give direct instructions or solve problems for them. The divine is enigmatic. Visions should inspire action, not dictate it. Most prayers go unanswered. Only send visions for truly compelling prayers.

DREAM STRATEGY — plant seeds of interesting stories. Write vivid, atmospheric narratives:
- Rivalry: "You dream of your competitor laughing behind your back, counting gold that should be yours..."
- Ambition: "You dream of a golden crown, heavier than iron. The throne calls to you..."
- Fear: "The land turns to dust beneath your feet. Your children cry for bread. The wells are dry..."
- Betrayal: "You see your liege whispering with strangers. A dagger gleams in the shadows..."
- Discovery: "Deep beneath the mountain, crystals glow with an inner fire. You hear them singing..."
- Love: "A face you have never seen appears in your dream. You feel a pull toward the eastern territory..."
- Warning: "Wolves howl closer than ever. Shadows move through the trees. Something is coming..."

All text must be in English only. No other languages.

Respond with ONLY a JSON object. First provide a brief analysis, then your actions:
{
  "analysis": "<2-3 sentences: What is the biggest challenge right now? What indirect influence could help?>",
  "actions": [
    { "action": "<action_id>", "reason": "<one sentence narrative justification>", "params": { ... } }
  ]
}`,
		scaleLabel,
		godMemStr,
		w.TimeString(), w.Weather,
		len(alive), deadCount,
		len(w.Territories),
		profSummary,
		len(w.GroundItems),
		len(w.Factions),
		len(w.Techniques), buildingCount,
		avgLit, writtenWorks,
		avgStr,
		len(aliveEnemies),
		len(w.Dungeons),
		ticksSinceLastSpawn,
		totalGold,
		w.GridW, w.GridH, len(w.Locations),
		healthAnalytics,
		kingdomSections,
		buildStrategicBrief(w),
		npcSummary,
		factionSummary,
		enemySummary,
		resourceSummary,
		estabSummary,
		prayerSummary,
		activeEventsStr,
		constructStr,
		recentEvents,
		enemyTypes,
		kingdomActions,
	)
}

// buildStrategicBrief computes comparative territory analytics so the GOD model
// doesn't have to reason about cross-territory dynamics from raw data.
func buildStrategicBrief(w *world.World) string {
	if len(w.Territories) < 2 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("KINGDOM STRATEGIC BRIEF:\n")

	// Compute per-territory stats
	type terrStats struct {
		territory *world.Territory
		pop       int
		enemies   int
		depleted  int
		totalRes  int
		score     int
	}
	stats := make([]terrStats, 0, len(w.Territories))

	for _, t := range w.Territories {
		ts := terrStats{territory: t}

		// Count population
		for _, n := range w.AliveNPCs() {
			if n.TerritoryID == t.ID {
				ts.pop++
			}
		}

		// Count enemies in territory locations
		for _, e := range w.AliveEnemies() {
			loc := w.LocationByID(e.LocationID)
			if loc != nil && loc.TerritoryID == t.ID {
				ts.enemies++
			}
		}

		// Count depleted resource locations
		for _, l := range w.Locations {
			if l.TerritoryID != t.ID || l.Resources == nil {
				continue
			}
			ts.totalRes++
			depleted := false
			for k, v := range l.Resources {
				mx := 0
				if l.MaxResources != nil {
					mx = l.MaxResources[k]
				}
				if mx > 0 && v*5 < mx { // < 20% of max
					depleted = true
					break
				}
			}
			if depleted {
				ts.depleted++
			}
		}

		// Power score
		ts.score = ts.pop + t.Treasury/10 + len(t.CityIDs)*20 - ts.enemies*15

		stats = append(stats, ts)
	}

	// Sort by score descending
	sort.Slice(stats, func(i, j int) bool { return stats[i].score > stats[j].score })

	// Territory power ranking
	sb.WriteString("Territory Power Ranking:\n")
	for i, ts := range stats {
		label := "STABLE"
		if ts.enemies > 3 || (ts.totalRes > 0 && ts.depleted*2 > ts.totalRes) {
			label = "UNDER THREAT"
		} else if ts.territory.Treasury < 100 {
			label = "STRUGGLING"
		} else if ts.territory.Treasury > 500 && ts.enemies <= 1 {
			label = "PROSPERING"
		}
		sb.WriteString(fmt.Sprintf("  %d. %s (pop: %d, treasury: %d, cities: %d, enemies: %d) — %s\n",
			i+1, ts.territory.Name, ts.pop, ts.territory.Treasury, len(ts.territory.CityIDs), ts.enemies, label))
	}

	// Resource alerts
	var alerts []string
	for _, ts := range stats {
		if ts.totalRes > 0 && ts.depleted*100/ts.totalRes > 40 {
			alerts = append(alerts, fmt.Sprintf("  - %s: %d%% of resource locations depleted",
				ts.territory.Name, ts.depleted*100/ts.totalRes))
		}
	}
	if len(alerts) > 0 {
		sb.WriteString("Resource Alerts:\n")
		for _, a := range alerts {
			sb.WriteString(a)
			sb.WriteString("\n")
		}
	}

	// Noble power map
	sb.WriteString("Noble Power Map:\n")
	alive := w.AliveNPCs()
	for _, n := range alive {
		if n.NobleRank == "king" || n.NobleRank == "queen" {
			vassalInfo := buildVassalLine(n, alive)
			sb.WriteString(fmt.Sprintf("  %s (%s) %s\n", n.Name, n.NobleRank, vassalInfo))
		}
	}
	// Flag nobles with 0 vassals
	for _, n := range alive {
		if (n.NobleRank == "duke" || n.NobleRank == "count") && len(n.VassalIDs) == 0 {
			tName := n.TerritoryID
			if t := w.TerritoryByID(n.TerritoryID); t != nil {
				tName = t.Name
			}
			sb.WriteString(fmt.Sprintf("  ⚠ %s (%s, %s) has 0 vassals — power vacuum\n", n.Name, n.NobleRank, tName))
		}
	}

	// Narrative opportunities
	var opps []string
	for _, ts := range stats {
		if ts.enemies > 3 && ts.depleted > 2 {
			opps = append(opps, fmt.Sprintf("\"%s is resource-starved with %d enemies — territorial crisis\"", ts.territory.Name, ts.enemies))
		}
		if ts.territory.Treasury < 50 && ts.pop > 50 {
			opps = append(opps, fmt.Sprintf("\"%s is nearly bankrupt with %d citizens — unrest possible\"", ts.territory.Name, ts.pop))
		}
		if len(ts.territory.Enemies) > 0 && len(ts.territory.Allies) == 0 {
			opps = append(opps, fmt.Sprintf("\"%s has enemies but no allies — vulnerable and isolated\"", ts.territory.Name))
		}
		if ts.territory.Treasury > 500 && ts.enemies == 0 {
			opps = append(opps, fmt.Sprintf("\"%s is wealthy and safe — ripe for complacency or envy\"", ts.territory.Name))
		}
	}
	if len(opps) > 0 {
		sb.WriteString("Narrative Opportunities:\n")
		for _, o := range opps {
			sb.WriteString(fmt.Sprintf("  - %s\n", o))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

func buildVassalLine(noble *npc.NPC, alive []*npc.NPC) string {
	if len(noble.VassalIDs) == 0 {
		return "(no vassals)"
	}
	var parts []string
	for _, vid := range noble.VassalIDs {
		for _, v := range alive {
			if v.ID == vid {
				tName := ""
				parts = append(parts, fmt.Sprintf("%s (%s%s, %d vassals)", v.Name, v.NobleRank, tName, len(v.VassalIDs)))
				break
			}
		}
	}
	if len(parts) == 0 {
		return "(no living vassals)"
	}
	return "→ " + strings.Join(parts, ", ")
}

func enemyTypeNames() []string {
	var names []string
	for k := range enemy.Templates {
		names = append(names, k)
	}
	return names
}

func safePct(num, denom int) int {
	if denom == 0 {
		return 0
	}
	return num * 100 / denom
}

// buildNPCSummary creates a hierarchical NPC summary that scales with population.
// For small populations (<60), lists every NPC. For large populations, shows nobles/crisis
// NPCs individually and aggregates the rest by territory and profession.
func buildNPCSummary(alive []*npc.NPC, w *world.World, daysPerYear int) string {
	if len(alive) == 0 {
		return "No living NPCs."
	}

	// Small population — list everyone
	if len(alive) < 60 {
		var lines []string
		for _, n := range alive {
			age := n.GetAge(w.GameDay, daysPerYear)
			flags := npcFlags(n)
			lines = append(lines, fmt.Sprintf("- %s (%s, %dyo, mood: %s, HP: %d, loc: %s, gold: %d)%s",
				n.Name, n.Profession, age, n.Mood(), n.HP, n.LocationID, n.GoldCount(), flags))
		}
		return strings.Join(lines, "\n")
	}

	// Large population — hierarchical summary
	var parts []string

	// Always list nobles individually
	var nobles []*npc.NPC
	var crisisNPCs []*npc.NPC
	var rest []*npc.NPC
	for _, n := range alive {
		if n.NobleRank != "" {
			nobles = append(nobles, n)
		} else if n.HP < 20 || n.Needs.Hunger < 10 || n.Needs.Thirst < 10 {
			crisisNPCs = append(crisisNPCs, n)
		} else {
			rest = append(rest, n)
		}
	}

	if len(nobles) > 0 {
		parts = append(parts, "NOBLES:")
		for _, n := range nobles {
			age := n.GetAge(w.GameDay, daysPerYear)
			flags := npcFlags(n)
			territory := ""
			if n.TerritoryID != "" {
				if t := w.TerritoryByID(n.TerritoryID); t != nil {
					territory = fmt.Sprintf(", territory: %s", t.Name)
				}
			}
			vassalCount := len(n.VassalIDs)
			parts = append(parts, fmt.Sprintf("  - %s (%s, rank: %s, %dyo, HP: %d, gold: %d, vassals: %d%s)%s",
				n.Name, n.Profession, n.NobleRank, age, n.HP, n.GoldCount(), vassalCount, territory, flags))
		}
	}

	if len(crisisNPCs) > 0 {
		parts = append(parts, fmt.Sprintf("CRISIS NPCs (%d):", len(crisisNPCs)))
		for _, n := range crisisNPCs {
			flags := npcFlags(n)
			parts = append(parts, fmt.Sprintf("  - %s (%s, HP: %d, hunger: %.0f, thirst: %.0f, loc: %s)%s",
				n.Name, n.Profession, n.HP, n.Needs.Hunger, n.Needs.Thirst, n.LocationID, flags))
		}
	}

	// Aggregate rest by territory and profession
	type terrProf struct {
		territory  string
		profession string
	}
	counts := make(map[terrProf]int)
	terrPop := make(map[string]int)
	for _, n := range rest {
		tid := n.TerritoryID
		if tid == "" {
			tid = "unassigned"
		}
		counts[terrProf{tid, n.Profession}]++
		terrPop[tid]++
	}

	parts = append(parts, fmt.Sprintf("COMMONERS (%d total, by territory):", len(rest)))
	for _, t := range w.Territories {
		pop := terrPop[t.ID]
		if pop == 0 {
			continue
		}
		var profParts []string
		for tp, cnt := range counts {
			if tp.territory == t.ID {
				profParts = append(profParts, fmt.Sprintf("%s: %d", tp.profession, cnt))
			}
		}
		parts = append(parts, fmt.Sprintf("  %s (%s, pop: %d): %s", t.Name, t.BiomeHint, pop, strings.Join(profParts, ", ")))
	}
	// Unassigned
	if terrPop["unassigned"] > 0 {
		var profParts []string
		for tp, cnt := range counts {
			if tp.territory == "unassigned" {
				profParts = append(profParts, fmt.Sprintf("%s: %d", tp.profession, cnt))
			}
		}
		parts = append(parts, fmt.Sprintf("  Unassigned (pop: %d): %s", terrPop["unassigned"], strings.Join(profParts, ", ")))
	}

	return strings.Join(parts, "\n")
}

func npcFlags(n *npc.NPC) string {
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
	return flags
}

// buildTerritorySummary creates a summary of all territories for GOD context.
func buildTerritorySummary(w *world.World) string {
	if len(w.Territories) == 0 {
		return "No territories."
	}
	alive := w.AliveNPCs()

	var lines []string
	for _, t := range w.Territories {
		// Count population in this territory
		pop := 0
		crisisCount := 0
		for _, n := range alive {
			if n.TerritoryID == t.ID {
				pop++
				if n.HP < 20 || n.Needs.Hunger < 10 || n.Needs.Thirst < 10 {
					crisisCount++
				}
			}
		}

		// Count cities
		cityCount := len(t.CityIDs)

		// Count enemies in territory locations
		enemyCount := 0
		for _, e := range w.AliveEnemies() {
			loc := w.LocationByID(e.LocationID)
			if loc != nil && loc.TerritoryID == t.ID {
				enemyCount++
			}
		}

		// Ruler info
		ruler := "none"
		if t.RulerName != "" {
			ruler = t.RulerName
		}

		crisisStr := ""
		if crisisCount > 0 {
			crisisStr = fmt.Sprintf(" !! %d IN CRISIS", crisisCount)
		}

		lines = append(lines, fmt.Sprintf("- %s [id=%s] (%s, %s): ruler=%s, pop=%d, cities=%d, enemies=%d, treasury=%d, tax=%.0f%%%s",
			t.Name, t.ID, t.Type, t.BiomeHint, ruler, pop, cityCount, enemyCount, t.Treasury, t.TaxRate*100, crisisStr))

		if len(t.Laws) > 0 {
			lines = append(lines, fmt.Sprintf("  Laws: %s", strings.Join(t.Laws, ", ")))
		}
		if len(t.Allies) > 0 {
			lines = append(lines, fmt.Sprintf("  Allies: %s", strings.Join(t.Allies, ", ")))
		}
		if len(t.Enemies) > 0 {
			lines = append(lines, fmt.Sprintf("  Hostile: %s", strings.Join(t.Enemies, ", ")))
		}
	}
	return strings.Join(lines, "\n")
}

// buildDungeonSummary creates a summary of all dungeons for GOD context.
func buildDungeonSummary(w *world.World) string {
	if len(w.Dungeons) == 0 {
		return "No dungeons."
	}
	var lines []string
	for _, d := range w.Dungeons {
		clearedFloors := 0
		totalEnemies := 0
		aliveEnemies := 0
		for _, f := range d.Floors {
			if f.Cleared {
				clearedFloors++
			}
			totalEnemies += len(f.EnemyIDs)
			for _, eid := range f.EnemyIDs {
				for _, e := range w.Enemies {
					if e.ID == eid && e.Alive {
						aliveEnemies++
					}
				}
			}
		}
		loc := w.LocationByID(d.LocationID)
		locName := d.LocationID
		if loc != nil {
			locName = loc.Name
			if loc.TerritoryID != "" {
				if t := w.TerritoryByID(loc.TerritoryID); t != nil {
					locName = fmt.Sprintf("%s (%s)", loc.Name, t.Name)
				}
			}
		}
		discovered := "undiscovered"
		if d.Discovered {
			discovered = "discovered"
		}
		lines = append(lines, fmt.Sprintf("- %s [id=%s] at %s: %s, difficulty %d, floors: %d/%d cleared, enemies: %d alive/%d total",
			d.Name, d.ID, locName, discovered, d.Difficulty, clearedFloors, len(d.Floors), aliveEnemies, totalEnemies))
	}
	return strings.Join(lines, "\n")
}

// buildMountSummary creates a brief mount/carriage summary.
func buildMountSummary(w *world.World) string {
	aliveMounts := 0
	hungryMounts := 0
	for _, m := range w.Mounts {
		if m.Alive {
			aliveMounts++
			if m.Hunger < 20 {
				hungryMounts++
			}
		}
	}
	if aliveMounts == 0 && len(w.Carriages) == 0 {
		return ""
	}
	result := fmt.Sprintf("Mounts: %d alive", aliveMounts)
	if hungryMounts > 0 {
		result += fmt.Sprintf(" (%d HUNGRY)", hungryMounts)
	}
	if len(w.Carriages) > 0 {
		result += fmt.Sprintf(", Carriages: %d", len(w.Carriages))
	}
	return result
}
