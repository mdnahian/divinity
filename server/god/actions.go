package god

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/divinity/core/config"
	"github.com/divinity/core/enemy"
	"github.com/divinity/core/item"
	"github.com/divinity/core/knowledge"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

type GodResult struct {
	Text   string
	Type   string
	Logged bool
}

// Natural resource types that GOD can replenish (not wheat/thatch — those are human-cultivated).
var allowedResources = map[string]bool{
	"stone": true, "iron_ore": true, "wood": true, "game": true,
	"berries": true, "herbs": true, "fish": true, "clay": true,
	"sand": true, "gems": true, "ice": true, "peat": true,
	"fur": true, "cactus_water": true, "poison_herbs": true,
	"bog_iron": true, "frozen_herbs": true, "rare_ore": true,
}

// Natural location types that GOD can create (not human-built structures).
var allowedLocationTypes = map[string]bool{
	"forest": true, "mine": true, "well": true, "dock": true, "garden": true,
	"cave": true, "dungeon_entrance": true, "desert": true, "swamp": true, "tundra": true,
}

// ValidateGodActions parses god_act args and checks required params. Returns validation errors and false if any action is invalid.
func ValidateGodActions(argsJSON string, w *world.World) (errs []string, valid bool) {
	var params struct {
		Actions []struct {
			Action string                 `json:"action"`
			Params map[string]interface{} `json:"params"`
		} `json:"actions"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
		if strings.Contains(err.Error(), "unexpected end of JSON input") {
			return []string{"Your JSON was truncated (output too long). Keep 'analysis' under 1 sentence and minimize 'reason' fields. Retry god_act with shorter text."}, false
		}
		return []string{fmt.Sprintf("Invalid JSON: %v", err)}, false
	}
	aliveNames := make(map[string]string) // lower name -> original name
	for _, n := range w.AliveNPCs() {
		aliveNames[strings.ToLower(n.Name)] = n.Name
	}
	genericNames := map[string]bool{"stranger": true, "villager": true, "npc": true, "person": true, "someone": true, "": true}
	genericLocNames := map[string]bool{"new place": true, "location": true, "place": true, "": true}

	for i, a := range params.Actions {
		if a.Action == "" {
			continue
		}
		p := a.Params
		if p == nil {
			p = make(map[string]interface{})
		}
		switch a.Action {
		case "introduce_profession":
			if name, _ := p["name"].(string); strings.TrimSpace(name) == "" {
				errs = append(errs, fmt.Sprintf("action %d (introduce_profession): required param \"name\" is missing or empty", i+1))
			}
			targetName, _ := p["target_name"].(string)
			targetName = strings.TrimSpace(targetName)
			if targetName == "" {
				errs = append(errs, fmt.Sprintf("action %d (introduce_profession): required param \"target_name\" is missing — an NPC must discover this profession via dream", i+1))
			} else if _, ok := aliveNames[strings.ToLower(targetName)]; !ok {
				var names []string
				for _, n := range aliveNames {
					names = append(names, n)
				}
				errs = append(errs, fmt.Sprintf("action %d (introduce_profession): no alive NPC named %q. Use exact name from: %s", i+1, targetName, strings.Join(names, ", ")))
			}
		case "introduce_technique":
			if name, _ := p["name"].(string); strings.TrimSpace(name) == "" {
				errs = append(errs, fmt.Sprintf("action %d (introduce_technique): required param \"name\" is missing or empty", i+1))
			}
			targetName, _ := p["target_name"].(string)
			targetName = strings.TrimSpace(targetName)
			if targetName == "" {
				errs = append(errs, fmt.Sprintf("action %d (introduce_technique): required param \"target_name\" is missing — an NPC must discover this technique via dream", i+1))
			} else if _, ok := aliveNames[strings.ToLower(targetName)]; !ok {
				var names []string
				for _, n := range aliveNames {
					names = append(names, n)
				}
				errs = append(errs, fmt.Sprintf("action %d (introduce_technique): no alive NPC named %q. Use exact name from: %s", i+1, targetName, strings.Join(names, ", ")))
			}
		case "introduce_item":
			if name, _ := p["name"].(string); strings.TrimSpace(name) == "" {
				errs = append(errs, fmt.Sprintf("action %d (introduce_item): required param \"name\" is missing or empty", i+1))
			}
			targetName, _ := p["target_name"].(string)
			targetName = strings.TrimSpace(targetName)
			if targetName == "" {
				errs = append(errs, fmt.Sprintf("action %d (introduce_item): required param \"target_name\" is missing — an NPC must discover this item via dream", i+1))
			} else if _, ok := aliveNames[strings.ToLower(targetName)]; !ok {
				var names []string
				for _, n := range aliveNames {
					names = append(names, n)
				}
				errs = append(errs, fmt.Sprintf("action %d (introduce_item): no alive NPC named %q. Use exact name from: %s", i+1, targetName, strings.Join(names, ", ")))
			}
		case "send_dream":
			targetName, _ := p["target_name"].(string)
			targetName = strings.TrimSpace(targetName)
			if targetName == "" {
				errs = append(errs, fmt.Sprintf("action %d (send_dream): required param \"target_name\" is missing or empty", i+1))
			} else if _, ok := aliveNames[strings.ToLower(targetName)]; !ok {
				var names []string
				for _, n := range aliveNames {
					names = append(names, n)
				}
				errs = append(errs, fmt.Sprintf("action %d (send_dream): no alive NPC named %q. Use exact name from: %s", i+1, targetName, strings.Join(names, ", ")))
			}
			dreamContent, _ := p["dream_content"].(string)
			if strings.TrimSpace(dreamContent) == "" {
				errs = append(errs, fmt.Sprintf("action %d (send_dream): required param \"dream_content\" is missing or empty", i+1))
			}
		case "send_vision":
			targetName, _ := p["target_name"].(string)
			targetName = strings.TrimSpace(targetName)
			if targetName == "" {
				errs = append(errs, fmt.Sprintf("action %d (send_vision): required param \"target_name\" is missing or empty", i+1))
			} else if _, ok := aliveNames[strings.ToLower(targetName)]; !ok {
				var names []string
				for _, n := range aliveNames {
					names = append(names, n)
				}
				errs = append(errs, fmt.Sprintf("action %d (send_vision): no alive NPC named %q. Use exact name from: %s", i+1, targetName, strings.Join(names, ", ")))
			}
			visionContent, _ := p["vision_content"].(string)
			if strings.TrimSpace(visionContent) == "" {
				errs = append(errs, fmt.Sprintf("action %d (send_vision): required param \"vision_content\" is missing or empty", i+1))
			}
		case "send_omen":
			omenDesc, _ := p["omen_description"].(string)
			if strings.TrimSpace(omenDesc) == "" {
				errs = append(errs, fmt.Sprintf("action %d (send_omen): required param \"omen_description\" is missing or empty", i+1))
			}
		case "spawn_enemy":
			enemyType, _ := p["enemy_type"].(string)
			resolved := resolveEnemyType(enemyType)
			if resolved == "" {
				errs = append(errs, fmt.Sprintf("action %d (spawn_enemy): unknown enemy_type %q. Valid types: %s", i+1, enemyType, strings.Join(enemy.AllTemplateNames(), ", ")))
			}
		case "introduce_enemy":
			if name, _ := p["name"].(string); strings.TrimSpace(name) == "" {
				errs = append(errs, fmt.Sprintf("action %d (introduce_enemy): required param \"name\" is missing or empty", i+1))
			}
		case "spawn_npc":
			name, _ := p["name"].(string)
			if genericNames[strings.ToLower(strings.TrimSpace(name))] {
				errs = append(errs, fmt.Sprintf("action %d (spawn_npc): \"name\" must be a proper unique name, not %q", i+1, name))
			}
		case "create_location":
			name, _ := p["name"].(string)
			if genericLocNames[strings.ToLower(strings.TrimSpace(name))] {
				errs = append(errs, fmt.Sprintf("action %d (create_location): \"name\" must be a proper descriptive name, not %q", i+1, name))
			}
			locType, _ := p["type"].(string)
			if locType != "" && !allowedLocationTypes[locType] {
				var allowed []string
				for k := range allowedLocationTypes {
					allowed = append(allowed, k)
				}
				errs = append(errs, fmt.Sprintf("action %d (create_location): type %q is not allowed. GOD can only create natural features: %s", i+1, locType, strings.Join(allowed, ", ")))
			}
		case "replenish_resources":
			locID, _ := p["location_id"].(string)
			if locID != "" {
				loc := w.LocationByID(locID)
				if loc == nil {
					loc = w.LocationByName(locID)
				}
				if loc == nil {
					loc = w.LocationByNameFuzzy(locID)
				}
				if loc == nil {
					var locNames []string
					for _, l := range w.Locations {
						locNames = append(locNames, fmt.Sprintf("%s (id=%s)", l.Name, l.ID))
					}
					errs = append(errs, fmt.Sprintf("action %d (replenish_resources): unknown location %q. Known locations: %s", i+1, locID, strings.Join(locNames, ", ")))
				}
			}
			res, _ := p["resources"].(map[string]interface{})
			for key := range res {
				if !allowedResources[key] {
					var allowed []string
					for k := range allowedResources {
						allowed = append(allowed, k)
					}
					errs = append(errs, fmt.Sprintf("action %d (replenish_resources): resource %q is not allowed. GOD can only replenish natural resources: %s", i+1, key, strings.Join(allowed, ", ")))
				}
			}
		case "create_dungeon":
			if name, _ := p["name"].(string); strings.TrimSpace(name) == "" {
				errs = append(errs, fmt.Sprintf("action %d (create_dungeon): required param \"name\" is missing or empty", i+1))
			}
		case "shape_biome":
			biomeType, _ := p["biome_type"].(string)
			validBiomes := map[string]bool{"desert": true, "swamp": true, "forest": true, "tundra": true, "plains": true, "mountain": true, "lake": true}
			if !validBiomes[biomeType] {
				errs = append(errs, fmt.Sprintf("action %d (shape_biome): invalid biome_type %q. Valid: desert, swamp, forest, tundra, plains, mountain, lake", i+1, biomeType))
			}
		case "create_city":
			if name, _ := p["name"].(string); strings.TrimSpace(name) == "" {
				errs = append(errs, fmt.Sprintf("action %d (create_city): required param \"name\" is missing or empty", i+1))
			}
		case "territorial_omen":
			territoryID, _ := p["territory_id"].(string)
			if w.TerritoryByID(territoryID) == nil {
				var tNames []string
				for _, t := range w.Territories {
					tNames = append(tNames, fmt.Sprintf("%s (id=%s)", t.Name, t.ID))
				}
				errs = append(errs, fmt.Sprintf("action %d (territorial_omen): unknown territory %q. Known: %s", i+1, territoryID, strings.Join(tNames, ", ")))
			}
			if omen, _ := p["omen_description"].(string); strings.TrimSpace(omen) == "" {
				errs = append(errs, fmt.Sprintf("action %d (territorial_omen): required param \"omen_description\" is missing or empty", i+1))
			}
		case "natural_disaster":
			territoryID, _ := p["territory_id"].(string)
			if w.TerritoryByID(territoryID) == nil {
				var tNames []string
				for _, t := range w.Territories {
					tNames = append(tNames, fmt.Sprintf("%s (id=%s)", t.Name, t.ID))
				}
				errs = append(errs, fmt.Sprintf("action %d (natural_disaster): unknown territory %q. Known: %s", i+1, territoryID, strings.Join(tNames, ", ")))
			}
			disasterType, _ := p["disaster_type"].(string)
			validDisasters := map[string]bool{"drought": true, "flood": true, "earthquake": true, "blizzard": true, "plague": true}
			if !validDisasters[disasterType] {
				errs = append(errs, fmt.Sprintf("action %d (natural_disaster): invalid disaster_type %q. Valid: drought, flood, earthquake, blizzard, plague", i+1, disasterType))
			}
		case "stir_dreams":
			territoryID, _ := p["territory_id"].(string)
			if w.TerritoryByID(territoryID) == nil {
				var tNames []string
				for _, t := range w.Territories {
					tNames = append(tNames, fmt.Sprintf("%s (id=%s)", t.Name, t.ID))
				}
				errs = append(errs, fmt.Sprintf("action %d (stir_dreams): unknown territory %q. Known: %s", i+1, territoryID, strings.Join(tNames, ", ")))
			}
			if theme, _ := p["dream_theme"].(string); strings.TrimSpace(theme) == "" {
				errs = append(errs, fmt.Sprintf("action %d (stir_dreams): required param \"dream_theme\" is missing or empty", i+1))
			}
			if urgency, _ := p["urgency"].(string); urgency != "low" && urgency != "med" && urgency != "high" {
				errs = append(errs, fmt.Sprintf("action %d (stir_dreams): urgency must be low, med, or high (got %q)", i+1, urgency))
			}
		case "send_prophecy":
			if text, _ := p["prophecy_text"].(string); strings.TrimSpace(text) == "" {
				errs = append(errs, fmt.Sprintf("action %d (send_prophecy): required param \"prophecy_text\" is missing or empty", i+1))
			}
			if territoryID, _ := p["territory_id"].(string); territoryID != "" {
				if w.TerritoryByID(territoryID) == nil {
					var tNames []string
					for _, t := range w.Territories {
						tNames = append(tNames, fmt.Sprintf("%s (id=%s)", t.Name, t.ID))
					}
					errs = append(errs, fmt.Sprintf("action %d (send_prophecy): unknown territory %q. Known: %s", i+1, territoryID, strings.Join(tNames, ", ")))
				}
			}
		}
	}
	return errs, len(errs) == 0
}

func ExecuteGodAction(parsed map[string]interface{}, w *world.World, mem memory.Store, cfg *config.Config) GodResult {
	action, _ := parsed["action"].(string)
	params, _ := parsed["params"].(map[string]interface{})
	if params == nil {
		params = parsed
	}
	reason, _ := parsed["reason"].(string)

	switch action {
	case "do_nothing":
		return GodResult{Text: fmt.Sprintf("The GOD surveys the world and is satisfied. \"%s\"", reason), Type: "god"}

	case "spawn_npc":
		return spawnNPC(params, reason, w, cfg)

	case "spawn_enemy":
		return spawnEnemy(params, reason, w)

	case "send_dream":
		return sendDream(params, reason, w, mem)

	case "send_vision":
		return sendVision(params, reason, w, mem)

	case "send_omen":
		return sendOmen(params, reason, w, mem)

	case "advance_weather":
		wt, _ := params["weather_type"].(string)
		if wt == "" || wt == "fog" {
			wt = "clear"
		}
		w.Weather = wt
		return GodResult{Text: fmt.Sprintf("The GOD shifts the weather to %s. \"%s\"", wt, reason), Type: "god"}

	case "replenish_resources":
		return replenishResources(params, reason, w)

	case "introduce_item":
		return introduceItem(params, reason, w, mem)

	case "introduce_technique":
		return introduceTechnique(params, reason, w, mem)

	case "create_location":
		return createLocation(params, reason, w)

	case "expand_grid":
		amount := getInt(params, "amount", 4)
		amount = max(2, min(6, amount))
		w.GridW += amount
		w.GridH += amount
		return GodResult{Text: fmt.Sprintf("The GOD expands the world boundaries by %d cells. \"%s\"", amount, reason), Type: "god"}

	case "introduce_profession":
		return introduceProfession(params, reason, w, mem)

	case "introduce_enemy":
		return introduceEnemy(params, reason, w, mem)

	case "create_dungeon":
		return createDungeon(params, reason, w)

	case "shape_biome":
		return shapeBiome(params, reason, w)

	case "create_city":
		return createCity(params, reason, w, cfg)

	case "territorial_omen":
		return territorialOmen(params, reason, w, mem)

	case "natural_disaster":
		return naturalDisaster(params, reason, w, mem)

	case "stir_dreams":
		return stirDreams(params, reason, w)

	case "send_prophecy":
		return sendProphecy(params, reason, w)

	default:
		return GodResult{Text: "The GOD contemplates in silence.", Type: "god"}
	}
}

func spawnNPC(params map[string]interface{}, reason string, w *world.World, cfg *config.Config) GodResult {
	name, _ := params["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" || strings.EqualFold(name, "stranger") || strings.EqualFold(name, "villager") {
		return GodResult{Text: "The GOD tried to spawn an NPC but provided no proper name.", Type: "god"}
	}
	profession, _ := params["profession"].(string)
	if profession == "" {
		profession = "farmer"
	}
	age := getInt(params, "age", 25)

	var fallbackID string
	inns := w.LocationsByType("inn")
	if len(inns) > 0 {
		fallbackID = inns[0].ID
	} else if len(w.Locations) > 0 {
		fallbackID = w.Locations[0].ID
	}

	tmpl := npc.Template{
		Name:       name,
		Profession: profession,
		Age:        age,
		HomeID:     fallbackID,
		StartItems: []npc.InventoryItem{{Name: "bread", Qty: 2}, {Name: "gold", Qty: 15}},
		Skills:     map[string]float64{profession: float64(30 + rand.Intn(20))},
	}
	newNPC := npc.NewNPC(tmpl, len(w.NPCs), w.GameDay, cfg)
	markets := w.LocationsByType("market")
	if len(markets) > 0 {
		newNPC.LocationID = markets[0].ID
	}
	w.NPCs = append(w.NPCs, newNPC)
	w.SpawnQueue = append(w.SpawnQueue, world.SpawnEntry{
		NPCID:      newNPC.ID,
		Name:       newNPC.Name,
		Profession: newNPC.Profession,
	})
	return GodResult{Text: fmt.Sprintf("A new villager arrives: %s the %s, age %d. \"%s\"", name, profession, age, reason), Type: "god"}
}

func spawnEnemy(params map[string]interface{}, reason string, w *world.World) GodResult {
	enemyType, _ := params["enemy_type"].(string)
	enemyType = resolveEnemyType(enemyType)
	if enemyType == "" {
		return GodResult{Text: fmt.Sprintf("The GOD tried to summon an unknown creature \"%s\".", params["enemy_type"]), Type: "god"}
	}
	locID, _ := params["location_id"].(string)
	loc := w.LocationByID(locID)
	if loc == nil {
		tmpl, _ := enemy.GetTemplate(enemyType)
		var candidates []*world.Location
		for _, l := range w.Locations {
			for _, lt := range tmpl.Locations {
				if l.Type == lt {
					candidates = append(candidates, l)
					break
				}
			}
		}
		if len(candidates) == 0 {
			return GodResult{Text: fmt.Sprintf("No suitable location for %s.", enemyType), Type: "god"}
		}
		loc = candidates[rand.Intn(len(candidates))]
		locID = loc.ID
	}

	alive := w.AliveNPCs()
	avgStr := 0
	if len(alive) > 0 {
		total := 0
		for _, n := range alive {
			total += n.Stats.Strength
		}
		avgStr = total / len(alive)
	}

	e := enemy.CreateScaled(enemyType, locID, avgStr)
	w.Enemies = append(w.Enemies, e)
	locName := locID
	if loc != nil {
		locName = loc.Name
	}
	w.ActiveEvents = append(w.ActiveEvents, world.ActiveEvent{Name: fmt.Sprintf("A %s has been spotted at %s!", e.Name, locName), TicksLeft: 20})
	return GodResult{Text: fmt.Sprintf("The GOD unleashes a %s at %s. \"%s\"", e.Name, locName, reason), Type: "god"}
}

// sendDream injects a divine dream into an NPC's memory. The dream becomes a high-importance
// memory that influences the NPC's future decisions through the memory → prompt → LLM pipeline.
// No direct stat changes — influence is purely through information.
func sendDream(params map[string]interface{}, reason string, w *world.World, mem memory.Store) GodResult {
	targetName, _ := params["target_name"].(string)
	var target *npc.NPC
	for _, n := range w.AliveNPCs() {
		if strings.EqualFold(n.Name, targetName) {
			target = n
			break
		}
	}
	if target == nil {
		return GodResult{Text: "The divine dream drifts into the void — no such soul found.", Type: "god"}
	}

	dreamContent, _ := params["dream_content"].(string)
	if strings.TrimSpace(dreamContent) == "" {
		return GodResult{Text: "The GOD tried to send a dream but it had no content.", Type: "god"}
	}

	urgency, _ := params["urgency"].(string)
	importance := 0.7
	switch urgency {
	case "low":
		importance = 0.4
	case "high":
		importance = 0.95
	}

	w.QueueDivineDream(world.PendingDivineDream{
		NPCID:      target.ID,
		Text:       memory.DivineDreamPrefix + dreamContent,
		Importance: importance,
		Vividness:  1.0,
		Category:   memory.CatGodDream,
		Tags:       []string{target.ID, "divine"},
	})

	return GodResult{Text: fmt.Sprintf("The GOD sends a divine dream to %s: \"%s\" (urgency: %s, will arrive during sleep). \"%s\"", target.Name, dreamContent, urgency, reason), Type: "god"}
}

// sendVision sends a divine vision to an NPC who has recently prayed. Visions are the most
// powerful form of divine communication — high importance, vivid memories. The NPC's prayer
// is consumed (cleared from RecentPrayers).
func sendVision(params map[string]interface{}, reason string, w *world.World, mem memory.Store) GodResult {
	targetName, _ := params["target_name"].(string)
	var target *npc.NPC
	for _, n := range w.AliveNPCs() {
		if strings.EqualFold(n.Name, targetName) {
			target = n
			break
		}
	}
	if target == nil {
		return GodResult{Text: "The divine vision finds no soul to receive it.", Type: "god"}
	}

	// Check if this NPC has a recent prayer
	hasPrayer := false
	for _, p := range w.RecentPrayers {
		if p.NpcID == target.ID {
			hasPrayer = true
			break
		}
	}
	if !hasPrayer {
		return GodResult{Text: fmt.Sprintf("The GOD tried to send a vision to %s, but they have not prayed recently. Visions require a recent prayer.", target.Name), Type: "god"}
	}

	visionContent, _ := params["vision_content"].(string)
	if strings.TrimSpace(visionContent) == "" {
		return GodResult{Text: "The GOD tried to send a vision but it had no content.", Type: "god"}
	}

	divineMessage, _ := params["divine_message"].(string)
	fullVision := visionContent
	if divineMessage != "" {
		fullVision = fmt.Sprintf("%s A divine voice speaks: \"%s\"", visionContent, divineMessage)
	}

	mem.Add(target.ID, memory.Entry{
		Text:       memory.VisionPrefix + fullVision,
		Time:       w.TimeString(),
		Importance: 0.95,
		Vividness:  1.0,
		Category:   memory.CatGodVision,
		Tags:       []string{target.ID, "divine", "vision"},
	})

	// Clear the target's prayer from RecentPrayers
	filtered := make([]world.Prayer, 0, len(w.RecentPrayers))
	for _, p := range w.RecentPrayers {
		if p.NpcID != target.ID {
			filtered = append(filtered, p)
		}
	}
	w.RecentPrayers = filtered

	w.ActiveEvents = append(w.ActiveEvents, world.ActiveEvent{
		Name:      fmt.Sprintf("A divine vision appears to %s at the shrine!", target.Name),
		TicksLeft: 15,
	})
	return GodResult{Text: fmt.Sprintf("The GOD sends a divine vision to %s: \"%s\". \"%s\"", target.Name, visionContent, reason), Type: "god"}
}

// sendOmen broadcasts an ominous sign to all spiritually sensitive NPCs (SpiritualSensitivity > 50).
// Omens are moderate-importance memories that create a shared sense of foreboding or hope.
func sendOmen(params map[string]interface{}, reason string, w *world.World, mem memory.Store) GodResult {
	omenDesc, _ := params["omen_description"].(string)
	if strings.TrimSpace(omenDesc) == "" {
		return GodResult{Text: "The GOD tried to send an omen but gave no description.", Type: "god"}
	}

	alive := w.AliveNPCs()
	affected := 0
	for _, n := range alive {
		if n.Stats.SpiritualSensitivity > 50 {
			mem.Add(n.ID, memory.Entry{
				Text:       memory.OmenPrefix + omenDesc,
				Time:       w.TimeString(),
				Importance: 0.6,
				Vividness:  1.0,
				Category:   memory.CatGodOmen,
				Tags:       []string{"omen", "divine"},
			})
			affected++
		}
	}

	w.ActiveEvents = append(w.ActiveEvents, world.ActiveEvent{
		Name:      fmt.Sprintf("Strange omens in the sky: %s", omenDesc),
		TicksLeft: 20,
	})
	return GodResult{Text: fmt.Sprintf("The GOD sends an omen: \"%s\" (%d NPCs sensed it). \"%s\"", omenDesc, affected, reason), Type: "god"}
}

func replenishResources(params map[string]interface{}, reason string, w *world.World) GodResult {
	locID, _ := params["location_id"].(string)
	loc := w.LocationByID(locID)
	if loc == nil {
		loc = w.LocationByName(locID)
	}
	if loc == nil {
		loc = w.LocationByNameFuzzy(locID)
	}
	if loc == nil {
		return GodResult{Text: fmt.Sprintf("The GOD tried to replenish resources at an unknown location %q.", locID), Type: "god"}
	}
	if loc.Resources == nil {
		loc.Resources = make(map[string]int)
	}
	if loc.MaxResources == nil {
		loc.MaxResources = make(map[string]int)
	}
	res, _ := params["resources"].(map[string]interface{})
	var added []string
	var blocked []string
	for key, val := range res {
		if !allowedResources[key] {
			blocked = append(blocked, key)
			continue
		}
		amount := 0
		switch v := val.(type) {
		case float64:
			amount = int(math.Min(50, math.Max(0, v)))
		case int:
			amount = min(50, max(0, v))
		}
		loc.Resources[key] += amount
		if loc.Resources[key] > loc.MaxResources[key] {
			loc.MaxResources[key] = loc.Resources[key]
		}
		added = append(added, fmt.Sprintf("%s: +%d", key, amount))
	}
	result := fmt.Sprintf("The GOD replenishes %s: %s. \"%s\"", loc.Name, strings.Join(added, ", "), reason)
	if len(blocked) > 0 {
		result += fmt.Sprintf(" (blocked non-natural resources: %s)", strings.Join(blocked, ", "))
	}
	return GodResult{Text: result, Type: "god"}
}

func introduceItem(params map[string]interface{}, reason string, w *world.World, mem memory.Store) GodResult {
	name, _ := params["name"].(string)
	if name == "" {
		return GodResult{Text: "The GOD tried to create an item but gave no name.", Type: "god"}
	}
	if _, ok := item.Registry[name]; ok {
		return GodResult{Text: fmt.Sprintf("The item \"%s\" already exists.", name), Type: "god"}
	}

	targetName, _ := params["target_name"].(string)
	var target *npc.NPC
	for _, n := range w.AliveNPCs() {
		if strings.EqualFold(n.Name, targetName) {
			target = n
			break
		}
	}
	if target == nil {
		return GodResult{Text: "The GOD tried to introduce an item but no NPC was found to discover it.", Type: "god"}
	}

	wt := 0.5
	if wf, ok := params["weight"].(float64); ok {
		wt = wf
	}
	category, _ := params["category"].(string)
	slot, _ := params["slot"].(string)
	stackable := true
	if slot != "" {
		stackable = false
	}

	var effects map[string]float64
	if rawEffects, ok := params["effects"].(map[string]interface{}); ok && len(rawEffects) > 0 {
		effects = make(map[string]float64, len(rawEffects))
		for k, v := range rawEffects {
			if fv, ok := v.(float64); ok {
				effects[k] = fv
			}
		}
	}

	item.Registry[name] = item.ItemDef{
		Weight:         wt,
		Stackable:      stackable,
		Slot:           slot,
		DurabilityBase: 50,
		DecayRate:      1,
		Category:       category,
		Effects:        effects,
	}

	// Queue discovery dream for delivery during sleep
	w.QueueDivineDream(world.PendingDivineDream{
		NPCID:      target.ID,
		Text:       memory.DivineDreamPrefix + fmt.Sprintf("In a vivid dream, I saw how to create a new thing: \"%s\". The knowledge burns bright in my mind.", name),
		Importance: 0.8,
		Vividness:  1.0,
		Category:   memory.CatGodDream,
		Tags:       []string{target.ID, "divine", "discovery", name},
	})

	return GodResult{Text: fmt.Sprintf("The GOD introduces a new item: \"%s\" — discovered by %s in a dream. %s", name, target.Name, reason), Type: "god"}
}

func introduceTechnique(params map[string]interface{}, reason string, w *world.World, mem memory.Store) GodResult {
	name, _ := params["name"].(string)
	if name == "" {
		return GodResult{Text: "The GOD tried to introduce a technique but gave no name.", Type: "god"}
	}

	targetName, _ := params["target_name"].(string)
	var target *npc.NPC
	for _, n := range w.AliveNPCs() {
		if strings.EqualFold(n.Name, targetName) {
			target = n
			break
		}
	}
	if target == nil {
		return GodResult{Text: "The GOD tried to introduce a technique but no NPC was found to discover it.", Type: "god"}
	}

	desc, _ := params["description"].(string)
	bonusType, _ := params["bonus_type"].(string)
	if bonusType == "" {
		bonusType = "craft_quality"
	}
	tech := knowledge.CreateTechnique(name, desc, bonusType, "", w.GameDay)
	w.Techniques = append(w.Techniques, tech)

	// Queue discovery dream for delivery during sleep
	dreamDesc := desc
	if dreamDesc == "" {
		dreamDesc = name
	}
	w.QueueDivineDream(world.PendingDivineDream{
		NPCID:      target.ID,
		Text:       memory.DivineDreamPrefix + fmt.Sprintf("In a dream, I discovered a new technique: \"%s\" — %s. I must try this!", name, dreamDesc),
		Importance: 0.8,
		Vividness:  1.0,
		Category:   memory.CatGodDream,
		Tags:       []string{target.ID, "divine", "discovery", name},
	})

	return GodResult{Text: fmt.Sprintf("The GOD seeds the technique \"%s\" (%s) — discovered by %s in a dream. %s", name, tech.BonusLabel, target.Name, reason), Type: "god"}
}

func introduceProfession(params map[string]interface{}, reason string, w *world.World, mem memory.Store) GodResult {
	name, _ := params["name"].(string)
	if name == "" {
		return GodResult{Text: "The GOD tried to introduce a profession but gave no name.", Type: "god"}
	}

	targetName, _ := params["target_name"].(string)
	var target *npc.NPC
	for _, n := range w.AliveNPCs() {
		if strings.EqualFold(n.Name, targetName) {
			target = n
			break
		}
	}
	if target == nil {
		return GodResult{Text: "The GOD tried to introduce a profession but no NPC was found to discover it.", Type: "god"}
	}

	skill, _ := params["primary_skill"].(string)
	if skill == "" {
		skill = name
	}
	litBonus := getInt(params, "literacy_bonus", 0)
	description, _ := params["description"].(string)
	npc.RegisterProfession(name, skill, litBonus, description)

	// Queue discovery dream for delivery during sleep
	dreamDesc := description
	if dreamDesc == "" {
		dreamDesc = "a new way of working"
	}
	w.QueueDivineDream(world.PendingDivineDream{
		NPCID:      target.ID,
		Text:       memory.DivineDreamPrefix + fmt.Sprintf("In a dream, I envisioned a new calling: \"%s\" — %s. Perhaps someone could pursue this path.", name, dreamDesc),
		Importance: 0.8,
		Vividness:  1.0,
		Category:   memory.CatGodDream,
		Tags:       []string{target.ID, "divine", "discovery", name},
	})

	return GodResult{Text: fmt.Sprintf("The GOD introduces a new profession: \"%s\" (primary skill: %s) — discovered by %s in a dream. \"%s\"", name, skill, target.Name, reason), Type: "god"}
}

func introduceEnemy(params map[string]interface{}, reason string, w *world.World, mem memory.Store) GodResult {
	name, _ := params["name"].(string)
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return GodResult{Text: "The GOD tried to introduce an enemy but gave no name.", Type: "god"}
	}
	name = strings.ReplaceAll(name, " ", "_")

	if _, exists := enemy.GetTemplate(name); exists {
		return GodResult{Text: fmt.Sprintf("Enemy type %q already exists.", name), Type: "god"}
	}

	category, _ := params["category"].(string)
	if category != "beast" && category != "human" {
		category = "beast"
	}

	hp := getInt(params, "hp", 40)
	strength := getInt(params, "strength", 40)
	agility := getInt(params, "agility", 40)
	defense := getInt(params, "defense", 15)

	var locations []string
	if locs, ok := params["locations"].([]interface{}); ok {
		for _, l := range locs {
			if s, ok := l.(string); ok {
				locations = append(locations, s)
			}
		}
	}
	if len(locations) == 0 {
		locations = []string{"forest"}
	}

	var items []string
	if itemList, ok := params["items"].([]interface{}); ok {
		for _, it := range itemList {
			if s, ok := it.(string); ok {
				items = append(items, s)
			}
		}
	}

	enemy.RegisterEnemy(name, enemy.EnemyTemplate{
		Category:  category,
		HP:        hp,
		Strength:  strength,
		Agility:   agility,
		Defense:   defense,
		Locations: locations,
		Items:     items,
	})

	displayName := strings.ReplaceAll(name, "_", " ")

	targetName, _ := params["target_name"].(string)
	var target *npc.NPC
	for _, n := range w.AliveNPCs() {
		if strings.EqualFold(n.Name, targetName) {
			target = n
			break
		}
	}
	if target != nil {
		w.QueueDivineDream(world.PendingDivineDream{
			NPCID:      target.ID,
			Text:       memory.DivineDreamPrefix + fmt.Sprintf("In a vision, I saw a terrible creature — a %s. I must warn others.", displayName),
			Importance: 0.8,
			Vividness:  1.0,
			Category:   memory.CatGodDream,
			Tags:       []string{target.ID, "divine", "discovery", name},
		})
	}

	return GodResult{Text: fmt.Sprintf("A new creature enters the world: the %s. \"%s\"", displayName, reason), Type: "god"}
}

func createLocation(params map[string]interface{}, reason string, w *world.World) GodResult {
	name, _ := params["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" || strings.EqualFold(name, "new place") || strings.EqualFold(name, "location") {
		return GodResult{Text: "The GOD tried to create a location but provided no proper name.", Type: "god"}
	}
	locType, _ := params["type"].(string)
	if locType == "" {
		locType = "forest"
	}
	if !allowedLocationTypes[locType] {
		var allowed []string
		for k := range allowedLocationTypes {
			allowed = append(allowed, k)
		}
		return GodResult{Text: fmt.Sprintf("The GOD cannot create a %s — only natural features are allowed: %s.", locType, strings.Join(allowed, ", ")), Type: "god"}
	}
	size := TypeSizes[locType]
	if size == [2]int{} {
		size = [2]int{2, 1}
	}
	x := w.GridW
	y := rand.Intn(max(1, w.GridH-size[1]))
	needW := x + size[0] + 2
	needH := y + size[1] + 2
	if needW > w.GridW {
		w.GridW = needW
	}
	if needH > w.GridH {
		w.GridH = needH
	}
	id := fmt.Sprintf("god_loc_%d", len(w.Locations))
	desc, _ := params["description"].(string)
	if desc == "" {
		desc = fmt.Sprintf("A %s shaped by divine will.", locType)
	}
	newLoc := &world.Location{
		ID: id, Name: name, Type: locType,
		X: x, Y: y, W: size[0], H: size[1],
		Color: locationColor(locType), Description: desc,
	}
	world.InitLocationResources(newLoc)
	w.Locations = append(w.Locations, newLoc)
	return GodResult{Text: fmt.Sprintf("The GOD creates a new location: \"%s\" (%s). %s", name, locType, reason), Type: "god"}
}

func getInt(m map[string]interface{}, key string, fallback int) int {
	v, ok := m[key]
	if !ok {
		return fallback
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	}
	return fallback
}

// createDungeon creates a new dungeon at a cave/dungeon_entrance in a territory.
func createDungeon(params map[string]interface{}, reason string, w *world.World) GodResult {
	name, _ := params["name"].(string)
	if strings.TrimSpace(name) == "" {
		name = "The Depths"
	}
	territoryID, _ := params["territory_id"].(string)
	territory := w.TerritoryByID(territoryID)
	if territory == nil && len(w.Territories) > 0 {
		territory = w.Territories[rand.Intn(len(w.Territories))]
	}

	difficulty := getInt(params, "difficulty", 3)
	difficulty = max(1, min(10, difficulty))
	floorCount := getInt(params, "floors", 3)
	floorCount = max(2, min(5, floorCount))

	// Find or create a cave entrance in the territory
	var entrance *world.Location
	for _, l := range w.Locations {
		if (l.Type == "cave" || l.Type == "dungeon_entrance") && w.DungeonAtLocation(l.ID) == nil {
			if territory == nil || l.TerritoryID == territory.ID {
				entrance = l
				break
			}
		}
	}
	if entrance == nil {
		// Create a cave entrance
		id := fmt.Sprintf("god_cave_%d", len(w.Locations))
		size := TypeSizes["cave"]
		if size == [2]int{} {
			size = [2]int{2, 2}
		}
		x := rand.Intn(max(1, w.GridW-size[0]))
		y := rand.Intn(max(1, w.GridH-size[1]))
		if territory != nil {
			x = territory.CenterX + rand.Intn(40) - 20
			y = territory.CenterY + rand.Intn(40) - 20
			x = max(0, min(w.GridW-size[0], x))
			y = max(0, min(w.GridH-size[1], y))
		}
		entrance = &world.Location{
			ID: id, Name: name + " Entrance", Type: "cave",
			X: x, Y: y, W: size[0], H: size[1],
			Color: locationColor("cave"), Description: "A dark cave entrance.",
		}
		if territory != nil {
			entrance.TerritoryID = territory.ID
		}
		world.InitLocationResources(entrance)
		w.Locations = append(w.Locations, entrance)
	}

	// Determine biome for enemy selection
	biome := "forest"
	if territory != nil {
		biome = territory.BiomeHint
	}

	// Build floors with biome-appropriate enemies
	dungeonID := fmt.Sprintf("dungeon_%d", len(w.Dungeons))
	var floors []world.DungeonFloor
	for i := 0; i < floorCount; i++ {
		enemyCount := 2 + rand.Intn(difficulty)
		var enemyIDs []string
		biomeTemplates := enemy.TemplatesForBiome(biome)
		if len(biomeTemplates) == 0 {
			biomeTemplates = enemy.AllTemplateNames()
		}

		alive := w.AliveNPCs()
		avgStr := 30
		if len(alive) > 0 {
			total := 0
			for _, n := range alive {
				total += n.Stats.Strength
			}
			avgStr = total / len(alive)
		}

		for j := 0; j < enemyCount; j++ {
			tmplName := biomeTemplates[rand.Intn(len(biomeTemplates))]
			e := enemy.CreateScaled(tmplName, entrance.ID, avgStr+difficulty*3)
			w.Enemies = append(w.Enemies, e)
			enemyIDs = append(enemyIDs, e.ID)
		}

		// Floor loot scales with difficulty
		var loot []world.DungeonLoot
		loot = append(loot, world.DungeonLoot{ItemName: "gold", DropChance: 0.8, MinQty: 5 * difficulty, MaxQty: 15 * difficulty})
		if difficulty >= 3 {
			loot = append(loot, world.DungeonLoot{ItemName: "iron_ore", DropChance: 0.5, MinQty: 1, MaxQty: 3})
		}
		if difficulty >= 5 {
			loot = append(loot, world.DungeonLoot{ItemName: "gems", DropChance: 0.3, MinQty: 1, MaxQty: 2})
		}
		if i == floorCount-1 && difficulty >= 7 {
			loot = append(loot, world.DungeonLoot{ItemName: "enchanted_blade", DropChance: 0.2, MinQty: 1, MaxQty: 1})
		}

		floors = append(floors, world.DungeonFloor{
			Level:    i + 1,
			EnemyIDs: enemyIDs,
			Loot:     loot,
			Cleared:  false,
		})
	}

	dungeon := &world.Dungeon{
		ID: dungeonID, Name: name,
		LocationID: entrance.ID,
		BiomeType:  biome,
		Floors:     floors,
		Difficulty: difficulty,
		Discovered: true,
		CreatedDay: w.GameDay,
	}
	w.Dungeons = append(w.Dungeons, dungeon)

	w.ActiveEvents = append(w.ActiveEvents, world.ActiveEvent{
		Name:      fmt.Sprintf("A dark dungeon has been revealed: %s!", name),
		TicksLeft: 30,
	})

	territoryName := "the wilderness"
	if territory != nil {
		territoryName = territory.Name
	}
	return GodResult{Text: fmt.Sprintf("The GOD reveals a dungeon: %s (difficulty %d, %d floors) in %s. \"%s\"", name, difficulty, floorCount, territoryName, reason), Type: "god"}
}

// shapeBiome adds a biome override that reshapes terrain in a territory.
func shapeBiome(params map[string]interface{}, reason string, w *world.World) GodResult {
	territoryID, _ := params["territory_id"].(string)
	biomeType, _ := params["biome_type"].(string)
	validBiomes := map[string]bool{"desert": true, "swamp": true, "forest": true, "tundra": true, "plains": true, "mountain": true, "lake": true}
	if !validBiomes[biomeType] {
		return GodResult{Text: fmt.Sprintf("Unknown biome type %q.", biomeType), Type: "god"}
	}

	x := getInt(params, "x", 0)
	y := getInt(params, "y", 0)
	radius := getInt(params, "radius", 10)
	radius = max(5, min(30, radius))

	// If territory specified, default to its center
	if territory := w.TerritoryByID(territoryID); territory != nil && x == 0 && y == 0 {
		x = territory.CenterX + rand.Intn(40) - 20
		y = territory.CenterY + rand.Intn(40) - 20
	}

	w.BiomeOverrides = append(w.BiomeOverrides, world.BiomeOverride{
		X: x, Y: y, BiomeType: biomeType, Radius: radius,
	})

	return GodResult{Text: fmt.Sprintf("The GOD reshapes the land — a %s spreads at (%d,%d) radius %d. \"%s\"", biomeType, x, y, radius, reason), Type: "god"}
}

// createCity founds a new city within a territory.
func createCity(params map[string]interface{}, reason string, w *world.World, cfg *config.Config) GodResult {
	name, _ := params["name"].(string)
	if strings.TrimSpace(name) == "" {
		return GodResult{Text: "The GOD tried to found a city but gave no name.", Type: "god"}
	}
	territoryID, _ := params["territory_id"].(string)
	territory := w.TerritoryByID(territoryID)
	if territory == nil && len(w.Territories) > 0 {
		territory = w.Territories[rand.Intn(len(w.Territories))]
	}

	desc, _ := params["description"].(string)
	if desc == "" {
		desc = "A new settlement founded by divine providence."
	}

	// Place city center near territory center
	x := rand.Intn(max(1, w.GridW-4))
	y := rand.Intn(max(1, w.GridH-4))
	if territory != nil {
		x = territory.CenterX + rand.Intn(60) - 30
		y = territory.CenterY + rand.Intn(60) - 30
		x = max(0, min(w.GridW-4, x))
		y = max(0, min(w.GridH-4, y))
	}

	// Create the city center location
	cityID := fmt.Sprintf("god_city_%d", len(w.Locations))
	cityCenter := &world.Location{
		ID: cityID, Name: name, Type: "city_center",
		X: x, Y: y, W: 4, H: 3,
		Color: locationColor("city_center"), Description: desc,
	}
	if territory != nil {
		cityCenter.TerritoryID = territory.ID
		territory.CityIDs = append(territory.CityIDs, cityID)
	}
	world.InitLocationResources(cityCenter)
	w.Locations = append(w.Locations, cityCenter)

	// Create basic buildings around the city
	basicBuildings := []struct {
		suffix string
		lType  string
	}{
		{"Market", "market"},
		{"Inn", "inn"},
		{"Well", "well"},
	}
	for _, bb := range basicBuildings {
		bID := fmt.Sprintf("%s_%s", cityID, bb.lType)
		bx := x + rand.Intn(8) - 4
		by := y + rand.Intn(8) - 4
		bx = max(0, min(w.GridW-2, bx))
		by = max(0, min(w.GridH-2, by))
		size := TypeSizes[bb.lType]
		if size == [2]int{} {
			size = [2]int{2, 1}
		}
		loc := &world.Location{
			ID: bID, Name: fmt.Sprintf("%s %s", name, bb.suffix), Type: bb.lType,
			X: bx, Y: by, W: size[0], H: size[1],
			Color: locationColor(bb.lType), CityID: cityID,
		}
		if territory != nil {
			loc.TerritoryID = territory.ID
		}
		world.InitLocationResources(loc)
		w.Locations = append(w.Locations, loc)
	}

	// Spawn a few NPCs for the new city
	profs := []string{"farmer", "merchant", "guard"}
	for _, prof := range profs {
		newNPC := npc.NewNPC(npc.Template{
			Name:       fmt.Sprintf("%s_%s", name, prof),
			Profession: prof,
			Age:        20 + rand.Intn(25),
			HomeID:     cityID,
			StartItems: []npc.InventoryItem{{Name: "bread", Qty: 3}, {Name: "gold", Qty: 20}},
			Skills:     map[string]float64{prof: float64(25 + rand.Intn(25))},
		}, len(w.NPCs), w.GameDay, cfg)
		newNPC.LocationID = cityID
		if territory != nil {
			newNPC.TerritoryID = territory.ID
		}
		w.NPCs = append(w.NPCs, newNPC)
	}

	w.ActiveEvents = append(w.ActiveEvents, world.ActiveEvent{
		Name:      fmt.Sprintf("A new city has been founded: %s!", name),
		TicksLeft: 30,
	})

	territoryName := "the wilderness"
	if territory != nil {
		territoryName = territory.Name
	}
	return GodResult{Text: fmt.Sprintf("The GOD founds a new city: %s in %s. \"%s\"", name, territoryName, reason), Type: "god"}
}

// territorialOmen sends an omen visible only to NPCs in a specific territory.
func territorialOmen(params map[string]interface{}, reason string, w *world.World, mem memory.Store) GodResult {
	territoryID, _ := params["territory_id"].(string)
	territory := w.TerritoryByID(territoryID)
	if territory == nil {
		return GodResult{Text: fmt.Sprintf("Unknown territory %q.", territoryID), Type: "god"}
	}

	omenDesc, _ := params["omen_description"].(string)
	if strings.TrimSpace(omenDesc) == "" {
		return GodResult{Text: "The GOD tried to send a territorial omen but gave no description.", Type: "god"}
	}

	alive := w.AliveNPCs()
	affected := 0
	for _, n := range alive {
		if n.TerritoryID == territory.ID && n.Stats.SpiritualSensitivity > 40 {
			mem.Add(n.ID, memory.Entry{
				Text:       memory.OmenPrefix + fmt.Sprintf("[In %s] %s", territory.Name, omenDesc),
				Time:       w.TimeString(),
				Importance: 0.65,
				Vividness:  1.0,
				Category:   memory.CatGodOmen,
				Tags:       []string{"omen", "divine", territory.ID},
			})
			affected++
		}
	}

	w.ActiveEvents = append(w.ActiveEvents, world.ActiveEvent{
		Name:      fmt.Sprintf("Strange omens in %s: %s", territory.Name, omenDesc),
		TicksLeft: 20,
	})
	return GodResult{Text: fmt.Sprintf("The GOD sends an omen to %s: \"%s\" (%d NPCs sensed it). \"%s\"", territory.Name, omenDesc, affected, reason), Type: "god"}
}

// naturalDisaster triggers a natural disaster in a territory, affecting resources and NPCs.
func naturalDisaster(params map[string]interface{}, reason string, w *world.World, mem memory.Store) GodResult {
	territoryID, _ := params["territory_id"].(string)
	territory := w.TerritoryByID(territoryID)
	if territory == nil {
		return GodResult{Text: fmt.Sprintf("Unknown territory %q.", territoryID), Type: "god"}
	}

	disasterType, _ := params["disaster_type"].(string)
	validDisasters := map[string]bool{"drought": true, "flood": true, "earthquake": true, "blizzard": true, "plague": true}
	if !validDisasters[disasterType] {
		return GodResult{Text: fmt.Sprintf("Unknown disaster type %q.", disasterType), Type: "god"}
	}

	severity, _ := params["severity"].(string)
	severityMult := 1.0
	switch severity {
	case "mild":
		severityMult = 0.5
	case "severe":
		severityMult = 2.0
	default:
		severity = "moderate"
	}

	// Apply effects to territory locations
	affectedLocs := 0
	for _, l := range w.Locations {
		if l.TerritoryID != territory.ID {
			continue
		}
		affectedLocs++
		if l.Resources == nil {
			continue
		}
		switch disasterType {
		case "drought":
			// Remove water, reduce crops
			delete(l.Resources, "water")
			if v, ok := l.Resources["wheat"]; ok {
				l.Resources["wheat"] = max(0, v-int(float64(10)*severityMult))
			}
			if v, ok := l.Resources["berries"]; ok {
				l.Resources["berries"] = max(0, v-int(float64(5)*severityMult))
			}
		case "flood":
			// Destroy crops and some resources, but add water
			if v, ok := l.Resources["wheat"]; ok {
				l.Resources["wheat"] = max(0, v-int(float64(15)*severityMult))
			}
			l.Resources["water"] = min(100, l.Resources["water"]+int(30*severityMult))
		case "earthquake":
			// Reduce stone/ore, damage buildings
			if v, ok := l.Resources["stone"]; ok {
				l.Resources["stone"] = max(0, v-int(float64(10)*severityMult))
			}
			if v, ok := l.Resources["iron_ore"]; ok {
				l.Resources["iron_ore"] = max(0, v-int(float64(5)*severityMult))
			}
		case "blizzard":
			// Remove herbs, reduce game
			delete(l.Resources, "herbs")
			if v, ok := l.Resources["game"]; ok {
				l.Resources["game"] = max(0, v-int(float64(8)*severityMult))
			}
		case "plague":
			// Reduce food resources
			for _, key := range []string{"wheat", "berries", "game", "fish"} {
				if v, ok := l.Resources[key]; ok {
					l.Resources[key] = max(0, v-int(float64(8)*severityMult))
				}
			}
		}
	}

	// Damage NPCs in the territory
	alive := w.AliveNPCs()
	affectedNPCs := 0
	for _, n := range alive {
		if n.TerritoryID != territory.ID {
			continue
		}
		affectedNPCs++
		dmg := int(float64(5) * severityMult)
		n.HP = max(1, n.HP-dmg)
		n.Stress = min(100, n.Stress+int(float64(10)*severityMult))
		mem.Add(n.ID, memory.Entry{
			Text:       fmt.Sprintf("A terrible %s struck %s! The damage is %s.", disasterType, territory.Name, severity),
			Time:       w.TimeString(),
			Importance: 0.8,
			Category:   memory.CatRoutine,
			Tags:       []string{"disaster", disasterType, territory.ID},
		})
	}

	tickDuration := int(20 * severityMult)
	w.ActiveEvents = append(w.ActiveEvents, world.ActiveEvent{
		Name:      fmt.Sprintf("A %s %s devastates %s!", severity, disasterType, territory.Name),
		TicksLeft: tickDuration,
	})

	return GodResult{Text: fmt.Sprintf("The GOD unleashes a %s %s upon %s — %d locations and %d NPCs affected. \"%s\"",
		severity, disasterType, territory.Name, affectedLocs, affectedNPCs, reason), Type: "god"}
}

// resolveEnemyType tries to match an enemy type string to a known template key.
// Handles plurals (wolves->wolf) and partial matches.
func resolveEnemyType(input string) string {
	if input == "" {
		return "wolf"
	}
	input = strings.ToLower(strings.TrimSpace(input))
	// Exact match
	if _, ok := enemy.GetTemplate(input); ok {
		return input
	}
	// Try singularizing: strip trailing "es" then "s"
	singular := input
	if strings.HasSuffix(singular, "ves") {
		// wolves -> wolf
		singular = strings.TrimSuffix(singular, "ves") + "f"
	} else if strings.HasSuffix(singular, "es") {
		singular = strings.TrimSuffix(singular, "es")
	} else if strings.HasSuffix(singular, "s") {
		singular = strings.TrimSuffix(singular, "s")
	}
	if _, ok := enemy.GetTemplate(singular); ok {
		return singular
	}
	// Substring containment
	for _, k := range enemy.AllTemplateNames() {
		if strings.Contains(k, input) || strings.Contains(input, k) {
			return k
		}
	}
	return ""
}

// stirDreams queues the same dream to up to 5 random non-noble alive NPCs in a territory.
func stirDreams(params map[string]interface{}, reason string, w *world.World) GodResult {
	territoryID, _ := params["territory_id"].(string)
	territory := w.TerritoryByID(territoryID)
	if territory == nil {
		return GodResult{Text: fmt.Sprintf("Unknown territory %q.", territoryID), Type: "god"}
	}

	dreamTheme, _ := params["dream_theme"].(string)
	if strings.TrimSpace(dreamTheme) == "" {
		return GodResult{Text: "The GOD tried to stir dreams but provided no theme.", Type: "god"}
	}

	urgency, _ := params["urgency"].(string)
	importance := 0.7
	switch urgency {
	case "low":
		importance = 0.4
	case "high":
		importance = 0.95
	}

	// Collect non-noble alive NPCs in the territory
	var candidates []*npc.NPC
	for _, n := range w.AliveNPCs() {
		if n.TerritoryID == territoryID && !n.IsNoble() {
			candidates = append(candidates, n)
		}
	}
	if len(candidates) == 0 {
		return GodResult{Text: fmt.Sprintf("No eligible commoners in %s to receive dreams.", territory.Name), Type: "god"}
	}

	// Shuffle and take up to 5
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	if len(candidates) > 5 {
		candidates = candidates[:5]
	}

	dreamText := memory.DivineDreamPrefix + dreamTheme
	for _, n := range candidates {
		w.QueueDivineDream(world.PendingDivineDream{
			NPCID:      n.ID,
			Text:       dreamText,
			Importance: importance,
			Vividness:  1.0,
			Category:   memory.CatGodDream,
			Tags:       []string{n.ID, "divine", "stir_dreams"},
		})
	}

	var names []string
	for _, n := range candidates {
		names = append(names, n.Name)
	}
	return GodResult{Text: fmt.Sprintf("The GOD stirs dreams in %s — %d commoners will dream of: \"%s\" (urgency: %s). Recipients: %s. \"%s\"",
		territory.Name, len(candidates), dreamTheme, urgency, strings.Join(names, ", "), reason), Type: "god"}
}

// sendProphecy creates a shared prophecy event and queues dreams to spiritually sensitive NPCs.
func sendProphecy(params map[string]interface{}, reason string, w *world.World) GodResult {
	prophecyText, _ := params["prophecy_text"].(string)
	if strings.TrimSpace(prophecyText) == "" {
		return GodResult{Text: "The GOD tried to send a prophecy but provided no text.", Type: "god"}
	}

	territoryID, _ := params["territory_id"].(string)
	var territoryName string
	if territoryID != "" {
		territory := w.TerritoryByID(territoryID)
		if territory == nil {
			return GodResult{Text: fmt.Sprintf("Unknown territory %q.", territoryID), Type: "god"}
		}
		territoryName = territory.Name
	}

	// Create the active event
	w.ActiveEvents = append(w.ActiveEvents, world.ActiveEvent{
		Name:      "A prophecy echoes: " + prophecyText,
		TicksLeft: 40,
	})

	// Queue dreams to spiritually sensitive NPCs
	dreamText := memory.DivineDreamPrefix + "A prophetic vision fills your mind: " + prophecyText
	affected := 0
	for _, n := range w.AliveNPCs() {
		if n.Stats.SpiritualSensitivity <= 30 {
			continue
		}
		if territoryID != "" && n.TerritoryID != territoryID {
			continue
		}
		w.QueueDivineDream(world.PendingDivineDream{
			NPCID:      n.ID,
			Text:       dreamText,
			Importance: 0.85,
			Vividness:  1.0,
			Category:   memory.CatGodDream,
			Tags:       []string{n.ID, "divine", "prophecy"},
		})
		affected++
	}

	scope := "the entire kingdom"
	if territoryName != "" {
		scope = territoryName
	}
	return GodResult{Text: fmt.Sprintf("The GOD sends a prophecy across %s: \"%s\" (%d spiritually sensitive NPCs will dream of it). \"%s\"",
		scope, prophecyText, affected, reason), Type: "god"}
}
