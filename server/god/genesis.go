package god

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"

	"github.com/divinity/core/config"
	"github.com/divinity/core/faction"
	"github.com/divinity/core/llm"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

func buildGenesisPrompt(npcCount int) string {
	return fmt.Sprintf(`You are the Supreme GOD creating a small medieval village. This is a humble beginning — one settlement that may grow over time.

Create a JSON object with:

1. "locations": array of exactly %d locations for a single village. Each must have:
   - "name": unique creative location name
   - "type": one of: market, inn, well, farm, forest, shrine, library, workshop, dock, garden, forge, mine, mill
   - "description": one evocative sentence describing this place
   REQUIRED: exactly 1 of each: market, inn, well, farm, forest, shrine, forge, mine, mill, dock, library, workshop. Add a couple of wells near outlying resource locations (farm, forest, mine) so NPCs have water access everywhere.

2. "npcs": array of exactly %d villagers — one per profession. Each must have:
   - "name": unique medieval-style name
   - "age": between 18-70
   - "profession": one of: farmer, hunter, merchant, herbalist, barmaid, blacksmith, carpenter, scribe, healer, fisher, miner, tailor, potter, potter
   - "personality": 2-4 comma-separated trait words (e.g. "brave, curious, stubborn")
   - "items": array of 2-4 starting item names chosen from: bread, gold, wheat, herbs, raw meat, hide, ale, healing potion, berries, spice bundle, walking stick, iron ore, iron ingot, leather, logs, cloth, rope, clay, fish
   REQUIRED professions (exactly 1 each): farmer, hunter, merchant, herbalist, barmaid, blacksmith, carpenter, scribe, healer, fisher, miner, tailor, potter. Remaining slots can be any profession.

3. "weather": starting weather, one of: clear, cloudy, rain

4. "lore": one sentence of world backstory/lore

IMPORTANT:
- This is a SMALL village, not a sprawling world. Keep it intimate and cozy.
- Every NPC must have "gold" in their items
- Be creative and evocative with names and descriptions

Respond with ONLY valid JSON. No markdown fences, no explanation.`,
		max(14, npcCount),
		max(14, npcCount),
	)
}

var RequiredTypes = []string{"market", "inn", "well", "farm", "forest", "shrine", "forge", "mill", "mine", "library", "dock"}

var TypeSizes = map[string][2]int{
	"market": {3, 2}, "inn": {3, 2}, "well": {1, 1},
	"farm": {4, 3}, "forest": {4, 4}, "shrine": {1, 1},
	"library": {2, 2}, "workshop": {2, 2}, "dock": {3, 2},
	"garden": {2, 2}, "home": {1, 1},
	"forge": {2, 2}, "mine": {3, 3}, "mill": {2, 2},
	// Kingdom-scale locations
	"palace": {16, 14}, "castle": {10, 9}, "manor": {5, 4},
	"stable": {3, 2}, "arena": {4, 3}, "barracks": {3, 2},
	"cave": {2, 2}, "dungeon_entrance": {2, 2},
}

var TypeColors = map[string]string{
	"market": "#b8860b", "inn": "#cd853f", "well": "#4682b4",
	"farm": "#228b22", "forest": "#2e8b57", "shrine": "#9370db",
	"library": "#4169e1", "workshop": "#cd853f", "dock": "#4682b4",
	"garden": "#3cb371", "home": "#8b7355", "forge": "#d35400",
	"mine": "#708090", "mill": "#a0522d", "school": "#4169e1",
	// Kingdom-scale locations
	"palace": "#ffd700", "castle": "#808080", "manor": "#a0522d",
	"stable": "#8b4513", "arena": "#b22222", "barracks": "#696969",
	"cave": "#4a4a4a", "dungeon_entrance": "#2f1f1f",
}

func locationColor(locType string) string {
	if c, ok := TypeColors[locType]; ok {
		return c
	}
	return "#888888"
}

func scaledRequiredProfs(npcCount int) map[string]int {
	return map[string]int{
		"farmer":     max(1, npcCount/20),
		"merchant":   max(1, npcCount/30),
		"hunter":     max(1, npcCount/30),
		"carpenter":  max(1, npcCount/30),
		"blacksmith": max(1, npcCount/30),
		"barmaid":    max(1, npcCount/30),
		"miner":      max(1, npcCount/30),
		"stablehand": max(1, npcCount/50),
		"guard":      max(1, npcCount/25),
	}
}

var RequiredProfs = map[string]int{
	"farmer": 1, "merchant": 1, "hunter": 1, "carpenter": 1, "blacksmith": 1, "barmaid": 1, "miner": 1,
}

var ProfLocMap = map[string]string{
	"barmaid": "inn", "blacksmith": "forge", "merchant": "market",
	"scribe": "library", "farmer": "farm", "carpenter": "workshop",
	"fisher": "dock", "miner": "mine",
	"stablehand": "stable", "guard": "barracks", "steward": "castle",
}

func RunGenesis(ctx context.Context, w *world.World, router *llm.Router, cfg *config.Config) (string, error) {
	const genesisSystem = `You are the Supreme GOD creating a medieval village. Output ONLY valid JSON. No markdown fences. No explanation. ALL text MUST be in English.`
	genesisPrompt := buildGenesisPrompt(cfg.Game.NPCCount)
	resp, err := router.CallGod(ctx, genesisSystem, genesisPrompt, cfg.GodAgent.Model, 0, cfg.GodAgent.Temperature)
	if err != nil {
		return "", fmt.Errorf("genesis LLM call: %w", err)
	}
	log.Printf("[Genesis] LLM response: %d chars, %d prompt tokens, %d completion tokens",
		len(resp.Content), resp.PromptTokens, resp.CompletionTokens)

	data, err := llm.ExtractJSON(resp.Content)
	if err != nil {
		return "", fmt.Errorf("genesis parse: %w", err)
	}

	validateGenesis(data, cfg.Game.NPCCount)

	locs := extractLocations(data)
	npcs := extractNPCs(data)
	weather, _ := data["weather"].(string)
	if weather == "" {
		weather = "clear"
	}
	lore, _ := data["lore"].(string)
	if lore == "" {
		lore = "A new world awakens from the void."
	}

	targetNPCCount := cfg.Game.NPCCount
	if targetNPCCount < len(npcs) {
		targetNPCCount = len(npcs)
	}

	layout := autoLayoutLocations(locs, targetNPCCount, cfg.Game.GridW, cfg.Game.GridH, cfg.Game.CityCount)
	w.Locations = layout.locations
	w.GridW = layout.gridW
	w.GridH = layout.gridH
	w.Roads = layout.roads
	w.Weather = weather

	for _, loc := range w.Locations {
		world.InitLocationResources(loc)
	}

	allNPCData := npcs
	if len(allNPCData) < targetNPCCount {
		extra := generateFillerNPCs(targetNPCCount - len(allNPCData))
		allNPCData = append(allNPCData, extra...)
	}

	homeLocations := make([]*world.Location, 0)
	for _, l := range w.Locations {
		if l.Type == "home" {
			homeLocations = append(homeLocations, l)
		}
	}
	var scatterLocs []*world.Location
	for _, l := range w.Locations {
		if l.Type != "inn" && l.Type != "home" {
			scatterLocs = append(scatterLocs, l)
		}
	}
	if len(scatterLocs) == 0 {
		scatterLocs = w.Locations
	}
	if len(scatterLocs) == 0 {
		log.Printf("[Genesis] WARNING: no locations available for NPC placement")
		return "no locations for NPC placement", nil
	}

	indices := make([]int, len(allNPCData))
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) { indices[i], indices[j] = indices[j], indices[i] })

	homeIdx := 0
	for i, npcData := range allNPCData {
		shuffledPos := indexOf(indices, i)
		getsHome := shuffledPos < len(homeLocations) && homeIdx < len(homeLocations)

		if getsHome {
			home := homeLocations[homeIdx]
			homeIdx++
			npcData["homeId"] = home.ID
			npcData["locationId"] = home.ID
			home.Name = fmt.Sprintf("%s's Home", npcData["name"])
			home.BuildingType = "shack"
			home.BuildingDurability = 80
		} else {
			pick := scatterLocs[rand.Intn(len(scatterLocs))]
			npcData["homeId"] = ""
			npcData["locationId"] = pick.ID
		}

		newNPC := npc.CreateFromGenesis(npcData, i, w.GameDay, cfg)
		if getsHome && homeIdx > 0 {
			homeLocations[homeIdx-1].BuildingOwnerID = newNPC.ID
			newNPC.HomeBuildingID = homeLocations[homeIdx-1].ID
		}
		w.NPCs = append(w.NPCs, newNPC)
	}

	for _, n := range w.NPCs {
		switch n.Profession {
		case "blacksmith":
			n.AddItem("iron ore", 4)
			n.AddItem("iron ingot", 2)
		case "miner":
			n.AddItem("pickaxe", 1)
			n.EquipItem("pickaxe")
		case "carpenter":
			n.AddItem("iron axe", 1)
			n.EquipItem("iron axe")
			n.AddItem("logs", 2)
			n.AddItem("rope", 2)
			n.AddItem("cloth", 2)
		case "tailor":
			n.AddItem("leather", 2)
			n.AddItem("cloth", 2)
		}
	}

	for _, n := range w.NPCs {
		locType, ok := ProfLocMap[n.Profession]
		if !ok {
			continue
		}
		for _, l := range w.Locations {
			if l.Type == locType && l.OwnerID == "" {
				l.OwnerID = n.ID
				n.IsBusinessOwner = true
				n.WorkplaceID = l.ID
				break
			}
		}
	}

	for _, n := range w.NPCs {
		if n.IsBusinessOwner || n.WorkplaceID != "" {
			continue
		}
		locType, ok := ProfLocMap[n.Profession]
		if !ok {
			continue
		}
		for _, l := range w.Locations {
			if l.Type == locType && l.OwnerID != "" {
				owner := w.FindNPCByID(l.OwnerID)
				if owner != nil {
					n.EmployerID = owner.ID
					n.WorkplaceID = l.ID
					n.Wage = 2
					break
				}
			}
		}
	}

	for _, n := range w.NPCs {
		w.SpawnQueue = append(w.SpawnQueue, world.SpawnEntry{
			NPCID:      n.ID,
			Name:       n.Name,
			Profession: n.Profession,
		})
	}

	log.Printf("[Genesis] Created world with %d locations, %d NPCs (%d in spawn queue)",
		len(w.Locations), len(w.NPCs), len(w.SpawnQueue))
	return lore, nil
}

func validateGenesis(data map[string]interface{}, npcCount ...int) {
	locsRaw, _ := data["locations"].([]interface{})
	types := make(map[string]bool)
	typeCounts := make(map[string]int)
	for _, lr := range locsRaw {
		locMap, _ := lr.(map[string]interface{})
		if locMap == nil {
			continue
		}
		t, _ := locMap["type"].(string)
		types[t] = true
		typeCounts[t]++
	}
	for _, req := range RequiredTypes {
		if !types[req] {
			locsRaw = append(locsRaw, map[string]interface{}{
				"name":        strings.ToUpper(req[:1]) + req[1:],
				"type":        req,
				"description": fmt.Sprintf("A %s area.", req),
			})
			typeCounts[req]++
		}
	}

	nc := 14
	if len(npcCount) > 0 && npcCount[0] > 0 {
		nc = npcCount[0]
	}
	minWells := max(2, nc/10)
	for typeCounts["well"] < minWells {
		wellNames := []string{"Village Well", "Spring Well", "Stone Well", "Mossy Well", "Old Well", "Market Well", "Garden Well", "Hill Well"}
		idx := typeCounts["well"]
		name := fmt.Sprintf("Well %d", idx+1)
		if idx < len(wellNames) {
			name = wellNames[idx]
		}
		locsRaw = append(locsRaw, map[string]interface{}{
			"name":        name,
			"type":        "well",
			"description": "A freshwater well providing clean drinking water.",
		})
		typeCounts["well"]++
	}

	data["locations"] = locsRaw

	npcsRaw, _ := data["npcs"].([]interface{})
	profCounts := make(map[string]int)
	for _, nr := range npcsRaw {
		nm, _ := nr.(map[string]interface{})
		if nm == nil {
			continue
		}
		prof, _ := nm["profession"].(string)
		if prof == "" {
			nm["profession"] = "farmer"
			prof = "farmer"
		}
		profCounts[prof]++
		if nm["name"] == nil || nm["name"] == "" {
			nm["name"] = "Stranger"
		}
		items, _ := nm["items"].([]interface{})
		hasGold := false
		for _, it := range items {
			if s, ok := it.(string); ok && s == "gold" {
				hasGold = true
			}
		}
		if !hasGold {
			nm["items"] = append(items, "gold")
		}
	}

	// Scale required professions based on actual NPC count (not target),
	// since LLM may return fewer NPCs than the full target
	actualNPCCount := len(npcsRaw)
	reqProfs := RequiredProfs
	if actualNPCCount > len(RequiredProfs) {
		reqProfs = scaledRequiredProfs(actualNPCCount)
	}
	for prof, minCount := range reqProfs {
		iterations := 0
		for profCounts[prof] < minCount {
			iterations++
			if iterations > 100 {
				break
			}
			reassigned := false
			for _, nr := range npcsRaw {
				nm, _ := nr.(map[string]interface{})
				if nm == nil {
					continue
				}
				p, _ := nm["profession"].(string)
				if profCounts[p] > RequiredProfs[p]+1 {
					profCounts[p]--
					nm["profession"] = prof
					profCounts[prof]++
					reassigned = true
					break
				}
			}
			if !reassigned {
				break
			}
		}
	}
}

func extractLocations(data map[string]interface{}) []map[string]interface{} {
	locsRaw, _ := data["locations"].([]interface{})
	var result []map[string]interface{}
	for _, lr := range locsRaw {
		if m, ok := lr.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

func extractNPCs(data map[string]interface{}) []map[string]interface{} {
	npcsRaw, _ := data["npcs"].([]interface{})
	var result []map[string]interface{}
	for _, nr := range npcsRaw {
		if m, ok := nr.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

type layoutResult struct {
	locations []*world.Location
	gridW     int
	gridH     int
	roads     []world.RoadSegment
}

var townTypes = map[string]bool{
	"inn": true, "market": true, "well": true, "shrine": true,
	"forge": true, "mill": true, "library": true, "workshop": true,
	"garden": true,
}

var outerTypes = map[string]bool{
	"forest": true, "farm": true, "mine": true, "dock": true,
}

var fillerFirstNames = []string{
	"Aldric", "Bran", "Cedric", "Dorin", "Edric", "Finn", "Gareth", "Hale",
	"Ivan", "Jorin", "Kael", "Loric", "Magnus", "Nolan", "Osric", "Perrin",
	"Quinn", "Rowan", "Soren", "Torin", "Ulric", "Varn", "Wren", "Yorick",
	"Anya", "Bessa", "Cora", "Dalia", "Eira", "Fenna", "Gwen", "Hilda",
	"Iris", "Juna", "Kira", "Lena", "Mira", "Nessa", "Opal", "Petra",
	"Reva", "Sable", "Tessa", "Una", "Vala", "Willa", "Yara", "Zara",
}
var fillerLastNames = []string{
	"Thorne", "Greenleaf", "Ironhand", "Deeproot", "Swiftfoot", "Woodcarver",
	"Stoneheart", "Brightwater", "Ashford", "Blackthorn", "Coldstream",
	"Duskwalker", "Ember", "Frostborn", "Goldweaver", "Hillcrest",
	"Kettleblack", "Longbow", "Meadowlark", "Nightingale", "Oakshield",
	"Pinecroft", "Ravenswood", "Silverbrook", "Tanglewood", "Underhill",
}
var fillerProfessions = []string{
	"farmer", "farmer", "farmer", "hunter", "merchant", "herbalist",
	"blacksmith", "carpenter", "fisher", "miner", "tailor", "potter",
	"healer", "scribe",
}
var fillerPersonalities = []string{
	"brave", "curious", "stubborn", "kind", "cautious", "cheerful",
	"grumpy", "quiet", "ambitious", "loyal", "resourceful", "gentle",
	"fierce", "patient", "witty", "stoic", "generous", "shrewd",
}
var fillerItems = [][]string{
	{"bread", "gold"}, {"berries", "gold", "herbs"},
	{"gold", "wheat"}, {"gold", "fish"}, {"gold", "leather"},
	{"gold", "bread", "ale"}, {"gold", "iron ore"},
}

func generateFillerNPCs(count int) []map[string]interface{} {
	usedNames := make(map[string]bool)
	result := make([]map[string]interface{}, 0, count)
	for i := 0; i < count; i++ {
		var name string
		for attempts := 0; attempts < 50; attempts++ {
			first := fillerFirstNames[rand.Intn(len(fillerFirstNames))]
			last := fillerLastNames[rand.Intn(len(fillerLastNames))]
			name = first + " " + last
			if !usedNames[name] {
				usedNames[name] = true
				break
			}
		}
		prof := fillerProfessions[rand.Intn(len(fillerProfessions))]
		p1 := fillerPersonalities[rand.Intn(len(fillerPersonalities))]
		p2 := fillerPersonalities[rand.Intn(len(fillerPersonalities))]
		for p2 == p1 {
			p2 = fillerPersonalities[rand.Intn(len(fillerPersonalities))]
		}
		items := fillerItems[rand.Intn(len(fillerItems))]
		itemsIface := make([]interface{}, len(items))
		for j, it := range items {
			itemsIface[j] = it
		}
		result = append(result, map[string]interface{}{
			"name":        name,
			"age":         float64(18 + rand.Intn(52)),
			"profession":  prof,
			"personality": p1 + ", " + p2,
			"items":       itemsIface,
		})
	}
	return result
}

type cityCenter struct {
	x, y int
}

func autoLayoutLocations(locs []map[string]interface{}, npcCount, gridW, gridH, cityCount int) layoutResult {
	scaledMin := int(math.Sqrt(float64(len(locs)+npcCount)) * 3)
	gw := max(gridW, scaledMin)
	gh := max(gridH, scaledMin*3/4)
	if cityCount < 1 {
		cityCount = 1
	}

	grid := make([][]bool, gh)
	for i := range grid {
		grid[i] = make([]bool, gw)
	}

	rebuildGrid := func(placed []*world.Location) {
		grid = make([][]bool, gh)
		for i := range grid {
			grid[i] = make([]bool, gw)
		}
		for _, loc := range placed {
			x0 := max(0, loc.X-1)
			y0 := max(0, loc.Y-1)
			x1 := min(gw, loc.X+loc.W+1)
			y1 := min(gh, loc.Y+loc.H+1)
			for dy := y0; dy < y1; dy++ {
				for dx := x0; dx < x1; dx++ {
					grid[dy][dx] = true
				}
			}
		}
	}

	fits := func(x, y, w, h int) bool {
		// Check with 1-cell buffer to match markPlaced buffer
		bx0 := x - 1
		by0 := y - 1
		bx1 := x + w + 1
		by1 := y + h + 1
		if bx0 < 0 || by0 < 0 || bx1 > gw || by1 > gh {
			// At least ensure the core footprint fits
			if x < 0 || y < 0 || x+w > gw || y+h > gh {
				return false
			}
		}
		// Check the buffered area for any occupied cells
		cx0 := max(0, bx0)
		cy0 := max(0, by0)
		cx1 := min(gw, bx1)
		cy1 := min(gh, by1)
		for dy := cy0; dy < cy1; dy++ {
			for dx := cx0; dx < cx1; dx++ {
				if grid[dy][dx] {
					return false
				}
			}
		}
		return true
	}

	markPlaced := func(x, y, w, h int) {
		x0 := max(0, x-1)
		y0 := max(0, y-1)
		x1 := min(gw, x+w+1)
		y1 := min(gh, y+h+1)
		for dy := y0; dy < y1; dy++ {
			for dx := x0; dx < x1; dx++ {
				grid[dy][dx] = true
			}
		}
	}

	centers := pickCityCenters(gw, gh, cityCount)

	// Generate town roads from city centers
	roadGrid := make([][]bool, gh)
	for i := range roadGrid {
		roadGrid[i] = make([]bool, gw)
	}
	var roads []world.RoadSegment

	for _, c := range centers {
		// Main road (horizontal)
		roadLen := 10 + rand.Intn(5) // 10-14 tiles
		x1 := max(0, c.x-roadLen/2)
		x2 := min(gw-1, c.x+roadLen/2)
		for x := x1; x <= x2; x++ {
			roadGrid[c.y][x] = true
		}
		roads = append(roads, world.RoadSegment{X1: x1, Y1: c.y, X2: x2, Y2: c.y})

		// Cross road (vertical)
		crossLen := 8 + rand.Intn(5) // 8-12 tiles
		y1 := max(0, c.y-crossLen/2)
		y2 := min(gh-1, c.y+crossLen/2)
		for y := y1; y <= y2; y++ {
			roadGrid[y][c.x] = true
		}
		roads = append(roads, world.RoadSegment{X1: c.x, Y1: y1, X2: c.x, Y2: y2})
	}

	findAlongRoad := func(cx, cy, w, h, maxR int) (int, int, bool) {
		type candidate struct{ x, y, score int }
		var candidates []candidate
		for y := max(0, cy-maxR); y <= min(gh-h, cy+maxR); y++ {
			for x := max(0, cx-maxR); x <= min(gw-w, cx+maxR); x++ {
				if !fits(x, y, w, h) {
					continue
				}
				// Score by adjacency to road cells
				score := 0
				for dy := -1; dy <= h; dy++ {
					for dx := -1; dx <= w; dx++ {
						ry := y + dy
						rx := x + dx
						if ry >= 0 && ry < len(roadGrid) && rx >= 0 && rx < len(roadGrid[ry]) && roadGrid[ry][rx] {
							score++
						}
					}
				}
				if score > 0 {
					candidates = append(candidates, candidate{x, y, score})
				}
			}
		}
		if len(candidates) == 0 {
			return 0, 0, false
		}
		// Pick randomly from top-scoring candidates
		maxScore := 0
		for _, c := range candidates {
			if c.score > maxScore {
				maxScore = c.score
			}
		}
		var best []candidate
		for _, c := range candidates {
			if c.score >= maxScore-1 {
				best = append(best, c)
			}
		}
		pick := best[rand.Intn(len(best))]
		return pick.x, pick.y, true
	}

	findNear := func(cx, cy, w, h, minR, maxR int) (int, int, bool) {
		type candidate struct{ x, y int }
		var candidates []candidate
		for y := max(0, cy-maxR); y <= min(gh-h, cy+maxR); y++ {
			for x := max(0, cx-maxR); x <= min(gw-w, cx+maxR); x++ {
				dx := float64(x+w/2) - float64(cx)
				dy := float64(y+h/2) - float64(cy)
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist >= float64(minR) && dist <= float64(maxR) && fits(x, y, w, h) {
					candidates = append(candidates, candidate{x, y})
				}
			}
		}
		if len(candidates) == 0 {
			return 0, 0, false
		}
		c := candidates[rand.Intn(len(candidates))]
		return c.x, c.y, true
	}

	findFarFromAll := func(w, h int, minDist float64) (int, int, bool) {
		type candidate struct{ x, y int }
		var candidates []candidate
		for y := 0; y <= gh-h; y++ {
			for x := 0; x <= gw-w; x++ {
				if !fits(x, y, w, h) {
					continue
				}
				closest := math.MaxFloat64
				for _, c := range centers {
					dx := float64(x+w/2) - float64(c.x)
					dy := float64(y+h/2) - float64(c.y)
					d := math.Sqrt(dx*dx + dy*dy)
					if d < closest {
						closest = d
					}
				}
				if closest >= minDist {
					candidates = append(candidates, candidate{x, y})
				}
			}
		}
		if len(candidates) == 0 {
			return 0, 0, false
		}
		c := candidates[rand.Intn(len(candidates))]
		return c.x, c.y, true
	}

	usedIDs := make(map[string]bool)
	genID := func(t string) string {
		if !usedIDs[t] {
			usedIDs[t] = true
			return t
		}
		for i := 2; ; i++ {
			id := fmt.Sprintf("%s_%d", t, i)
			if !usedIDs[id] {
				usedIDs[id] = true
				return id
			}
		}
	}

	var townLocs, outerLocs, wellLocs []map[string]interface{}
	for _, l := range locs {
		t, _ := l["type"].(string)
		if outerTypes[t] {
			outerLocs = append(outerLocs, l)
		} else if t == "well" {
			wellLocs = append(wellLocs, l)
		} else {
			townLocs = append(townLocs, l)
		}
	}

	var placed []*world.Location

	for i, l := range townLocs {
		t, _ := l["type"].(string)
		name, _ := l["name"].(string)
		desc, _ := l["description"].(string)
		size := TypeSizes[t]
		if size == [2]int{} {
			size = [2]int{2, 1}
		}
		c := centers[i%len(centers)]
		// Try road-adjacent first
		x, y, ok := findAlongRoad(c.x, c.y, size[0], size[1], 6)
		if !ok {
			x, y, ok = findNear(c.x, c.y, size[0], size[1], 0, 5)
		}
		if !ok {
			x, y, ok = findNear(c.x, c.y, size[0], size[1], 0, 10)
		}
		if !ok {
			gw += 4
			gh += 4
			rebuildGrid(placed)
			// Rebuild road grid too
			newRoadGrid := make([][]bool, gh)
			for ri := range newRoadGrid {
				newRoadGrid[ri] = make([]bool, gw)
			}
			for ry := 0; ry < len(roadGrid) && ry < gh; ry++ {
				for rx := 0; rx < len(roadGrid[ry]) && rx < gw; rx++ {
					newRoadGrid[ry][rx] = roadGrid[ry][rx]
				}
			}
			roadGrid = newRoadGrid
			x, y, ok = findNear(c.x, c.y, size[0], size[1], 0, 10)
			if !ok {
				continue
			}
		}
		loc := &world.Location{
			ID: genID(t), Name: name, Type: t,
			X: x, Y: y, W: size[0], H: size[1],
			Color: locationColor(t), Description: desc,
			Capacity: world.CalcCapacity(t, size[0], size[1]),
		}
		placed = append(placed, loc)
		markPlaced(x, y, size[0], size[1])
	}

	// Place wells: one near each city center, remaining spread evenly across the grid.
	spreadDist := math.Sqrt(float64(gw*gh) / float64(max(1, len(wellLocs))))
	for i, wl := range wellLocs {
		name, _ := wl["name"].(string)
		desc, _ := wl["description"].(string)
		var x, y int
		var ok bool
		if i < len(centers) {
			c := centers[i]
			x, y, ok = findAlongRoad(c.x, c.y, 1, 1, 4)
			if !ok {
				x, y, ok = findNear(c.x, c.y, 1, 1, 0, 6)
			}
			if !ok {
				x, y, ok = findNear(c.x, c.y, 1, 1, 0, 12)
			}
		} else {
			// Spread remaining wells far from all existing placed locations
			x, y, ok = findFarFromAll(1, 1, spreadDist)
			if !ok {
				x, y, ok = findFarFromAll(1, 1, spreadDist*0.6)
			}
			if !ok {
				// Last resort: anywhere that fits
				x, y, ok = findNear(gw/2, gh/2, 1, 1, 0, max(gw, gh))
			}
		}
		if !ok {
			continue
		}
		loc := &world.Location{
			ID: genID("well"), Name: name, Type: "well",
			X: x, Y: y, W: 1, H: 1,
			Color: locationColor("well"), Description: desc,
			Capacity: world.CalcCapacity("well", 1, 1),
		}
		placed = append(placed, loc)
		markPlaced(x, y, 1, 1)
	}

	minOuterDist := float64(min(gw, gh)) * 0.15
	for _, l := range outerLocs {
		t, _ := l["type"].(string)
		name, _ := l["name"].(string)
		desc, _ := l["description"].(string)
		size := TypeSizes[t]
		if size == [2]int{} {
			size = [2]int{2, 1}
		}

		var x, y int
		var ok bool
		switch t {
		case "dock":
			margin := 4
			type candidate struct{ x, y int }
			var candidates []candidate
			for py := 0; py <= gh-size[1]; py++ {
				for px := 0; px <= gw-size[0]; px++ {
					if !fits(px, py, size[0], size[1]) {
						continue
					}
					if px < margin || px+size[0] > gw-margin || py < margin || py+size[1] > gh-margin {
						candidates = append(candidates, candidate{px, py})
					}
				}
			}
			if len(candidates) > 0 {
				c := candidates[rand.Intn(len(candidates))]
				x, y, ok = c.x, c.y, true
			}
		case "farm":
			nearestCenter := centers[rand.Intn(len(centers))]
			x, y, ok = findNear(nearestCenter.x, nearestCenter.y, size[0], size[1], 6, 15)
		default:
			x, y, ok = findFarFromAll(size[0], size[1], minOuterDist)
		}
		if !ok {
			gw += 4
			gh += 4
			rebuildGrid(placed)
			x, y, ok = findFarFromAll(size[0], size[1], 0)
			if !ok {
				continue
			}
		}
		loc := &world.Location{
			ID: genID(t), Name: name, Type: t,
			X: x, Y: y, W: size[0], H: size[1],
			Color: locationColor(t), Description: desc,
			Capacity: world.CalcCapacity(t, size[0], size[1]),
		}
		placed = append(placed, loc)
		markPlaced(x, y, size[0], size[1])
	}

	rebuildGrid(placed)
	homeCount := max(2, int(float64(npcCount)*(0.4+rand.Float64()*0.2)))
	for i := 0; i < homeCount; i++ {
		homeID := fmt.Sprintf("home_%d", i)
		usedIDs[homeID] = true
		c := centers[i%len(centers)]
		x, y, ok := findAlongRoad(c.x, c.y, 1, 1, 8)
		if !ok {
			x, y, ok = findNear(c.x, c.y, 1, 1, 1, 5)
		}
		if !ok {
			x, y, ok = findNear(c.x, c.y, 1, 1, 0, 10)
		}
		if !ok {
			gw += 2
			gh += 2
			rebuildGrid(placed)
			x, y, ok = findNear(c.x, c.y, 1, 1, 0, 12)
			if !ok {
				continue
			}
		}
		loc := &world.Location{
			ID: homeID, Name: fmt.Sprintf("Home %d", i+1), Type: "home",
			X: x, Y: y, W: 1, H: 1,
			Color:    locationColor("home"),
			Capacity: 2,
		}
		placed = append(placed, loc)
		markPlaced(x, y, 1, 1)
	}

	return layoutResult{locations: placed, gridW: gw, gridH: gh, roads: roads}
}

func pickCityCenters(gw, gh, count int) []cityCenter {
	if count <= 1 {
		return []cityCenter{{gw / 2, gh / 2}}
	}
	margin := 8
	var centers []cityCenter
	for i := 0; i < count; i++ {
		var best cityCenter
		bestDist := -1.0
		for attempt := 0; attempt < 100; attempt++ {
			cx := margin + rand.Intn(gw-2*margin)
			cy := margin + rand.Intn(gh-2*margin)
			minDist := math.MaxFloat64
			for _, c := range centers {
				dx := float64(cx - c.x)
				dy := float64(cy - c.y)
				d := math.Sqrt(dx*dx + dy*dy)
				if d < minDist {
					minDist = d
				}
			}
			if len(centers) == 0 {
				minDist = float64(gw + gh)
			}
			if minDist > bestDist {
				bestDist = minDist
				best = cityCenter{cx, cy}
			}
		}
		centers = append(centers, best)
	}
	return centers
}

func indexOf(s []int, val int) int {
	for i, v := range s {
		if v == val {
			return i
		}
	}
	return -1
}

// --- Kingdom-Scale Genesis ---

func buildKingdomMetaPrompt(territoryCount int) string {
	return fmt.Sprintf(`You are the Supreme GOD creating a vast medieval kingdom with %d territories.

Create a JSON object with:

1. "kingdom_name": a creative name for this kingdom
2. "lore": one sentence of world backstory

3. "territories": array of exactly %d territories. The FIRST territory is the Royal Domain (center). The remaining %d are duchies.
   Each territory must have:
   - "name": unique territory name (e.g. "The Ashlands", "Verdant Reach")
   - "biome": one of: desert, swamp, forest, tundra, plains, mountain (each duchy should have a DIFFERENT biome — variety is key)
   - "description": one evocative sentence
   - "ruler_name": for the royal domain, the king's name; for duchies, the duke's name
   - "queen_name": ONLY for the royal domain — the queen's name

4. "weather": starting weather, one of: clear, cloudy, rain

Respond with ONLY valid JSON. No markdown fences, no explanation.`,
		territoryCount, territoryCount, territoryCount-1)
}

func buildTerritoryGenesisPrompt(territory world.Territory, npcsPerTerritory int) string {
	locCount := max(14, npcsPerTerritory/3)
	if locCount > 30 {
		locCount = 30 // Cap to keep JSON under token limits; layout will add homes
	}
	// Cap LLM-generated NPCs — the rest will be filled procedurally by generateFillerNPCs
	llmNPCCount := npcsPerTerritory
	if llmNPCCount > 25 {
		llmNPCCount = 25
	}
	extra := ""
	if territory.Type == world.TerritoryTypeRoyalDomain {
		extra = `
   IMPORTANT: This is the ROYAL DOMAIN. Include a "palace" type location for the royal family. Include a "barracks" and a "stable".`
	} else {
		extra = fmt.Sprintf(`
   IMPORTANT: This is the duchy of %s. Include a "castle" type location for the duke, a "barracks" for guards, and a "stable" for horses.
   The dominant biome is %s — reflect this in location names and descriptions.`, territory.Name, territory.BiomeHint)
	}

	return fmt.Sprintf(`You are the Supreme GOD creating the territory of "%s" within a medieval kingdom.
The dominant biome here is: %s.

Create a JSON object with:

1. "locations": array of exactly %d locations for 3 cities/towns in this territory. Each must have:
   - "name": unique creative location name reflecting the %s biome
   - "type": one of: market, inn, well, farm, forest, shrine, library, workshop, dock, garden, forge, mine, mill, stable, barracks, palace, castle, manor, cave
   - "description": one evocative sentence
   REQUIRED: at least 1 of each: market, inn, well, farm, forest, shrine, forge, mine. Add 3-5 wells spread across the territory. Add at least 1 stable.%s

2. "npcs": array of exactly %d people. Each must have:
   - "name": unique medieval-style name
   - "age": between 18-70
   - "profession": one of: farmer, hunter, merchant, herbalist, barmaid, blacksmith, carpenter, scribe, healer, fisher, miner, tailor, potter, guard, baker, cook, innkeeper, stablehand
   - "personality": 2-4 comma-separated trait words
   - "items": array of 2-4 starting item names (must include "gold")
   REQUIRED professions (at least 1 each): farmer, hunter, merchant, blacksmith, carpenter, barmaid, miner, guard, stablehand

3. "cities": array of 3 city names for the 3 towns in this territory

Respond with ONLY valid JSON. No markdown fences, no explanation.`,
		territory.Name, territory.BiomeHint,
		locCount, territory.BiomeHint,
		extra, llmNPCCount)
}

// ProgressFunc is called by genesis to report progress to the loading screen.
// Parameters: phase name, detail text, current step, total steps.
type ProgressFunc func(phase, detail string, current, total int)

// RunKingdomGenesis generates a full kingdom with 6 territories, each with multiple cities.
func RunKingdomGenesis(ctx context.Context, w *world.World, router *llm.Router, cfg *config.Config, onProgress ...ProgressFunc) (string, error) {
	progress := func(phase, detail string, current, total int) {
		if len(onProgress) > 0 && onProgress[0] != nil {
			onProgress[0](phase, detail, current, total)
		}
	}
	const genesisSystem = `You are the Supreme GOD creating a medieval kingdom. Output ONLY valid JSON. No markdown fences. No explanation. ALL text MUST be in English.`

	territoryCount := cfg.Game.TerritoryCount
	if territoryCount < 2 {
		territoryCount = 6
	}

	// Step 1: Generate kingdom metadata (territory names, biomes, rulers)
	progress("Conjuring the Realm", "The gods deliberate over the shape of the kingdom...", 0, territoryCount+4)
	log.Printf("[KingdomGenesis] Generating kingdom metadata for %d territories...", territoryCount)
	metaPrompt := buildKingdomMetaPrompt(territoryCount)
	metaResp, err := router.CallGod(ctx, genesisSystem, metaPrompt, cfg.GodAgent.Model, 0, cfg.GodAgent.Temperature)
	if err != nil {
		return "", fmt.Errorf("kingdom meta LLM call: %w", err)
	}
	metaData, err := llm.ExtractJSON(metaResp.Content)
	if err != nil {
		return "", fmt.Errorf("kingdom meta parse: %w", err)
	}

	lore, _ := metaData["lore"].(string)
	if lore == "" {
		lore = "A vast kingdom rises from the ancient lands."
	}
	weather, _ := metaData["weather"].(string)
	if weather == "" {
		weather = "clear"
	}

	// Step 2: Create territory structures from placement algorithm + LLM metadata
	gridW := cfg.Game.GridW
	gridH := cfg.Game.GridH
	territoryPlacements := world.PlaceTerritoryCenters(gridW, gridH, territoryCount)

	// Parse LLM territory metadata
	terrMeta, _ := metaData["territories"].([]interface{})

	territories := make([]*world.Territory, 0, territoryCount)
	for i, tp := range territoryPlacements {
		t := &world.Territory{
			ID:      tp.ID,
			Type:    tp.Type,
			CenterX: tp.CenterX,
			CenterY: tp.CenterY,
			Radius:  tp.Radius,
			TaxRate: tp.TaxRate,
			CityIDs: make([]string, 0),
			Laws:    make([]string, 0),
		}
		// Apply LLM metadata
		if i < len(terrMeta) {
			tm, _ := terrMeta[i].(map[string]interface{})
			if tm != nil {
				if name, ok := tm["name"].(string); ok && name != "" {
					t.Name = name
				}
				if biome, ok := tm["biome"].(string); ok && biome != "" {
					t.BiomeHint = biome
				}
				if ruler, ok := tm["ruler_name"].(string); ok && ruler != "" {
					t.RulerName = ruler
				}
			}
		}
		if t.Name == "" {
			t.Name = fmt.Sprintf("Territory %d", i+1)
		}
		if t.BiomeHint == "" {
			biomes := world.ValidBiomeHints
			t.BiomeHint = biomes[i%len(biomes)]
		}
		territories = append(territories, t)
	}

	w.Territories = territories
	w.GridW = gridW
	w.GridH = gridH
	w.Weather = weather

	progress("Shaping the Territories", fmt.Sprintf("Forging %d territories in parallel...", len(territories)), 1, territoryCount+4)

	// Step 3: Generate each territory's locations and NPCs
	npcsPerTerritory := cfg.Game.NPCCount / territoryCount
	citiesPerTerritory := cfg.Game.CityCount / territoryCount
	if citiesPerTerritory < 2 {
		citiesPerTerritory = 3
	}

	globalNPCIdx := 0
	allRoads := make([]world.RoadSegment, 0)

	// Fire all territory LLM calls in parallel
	type terrResult struct {
		idx      int
		territory *world.Territory
		data     map[string]interface{}
		err      error
	}
	results := make([]terrResult, len(territories))
	var wg sync.WaitGroup
	for i, territory := range territories {
		wg.Add(1)
		go func(idx int, t *world.Territory) {
			defer wg.Done()
			log.Printf("[KingdomGenesis] Generating territory: %s (biome: %s)...", t.Name, t.BiomeHint)
			terrPrompt := buildTerritoryGenesisPrompt(*t, npcsPerTerritory)

			const maxRetries = 3
			var lastErr error
			for attempt := 1; attempt <= maxRetries; attempt++ {
				terrResp, err := router.CallGod(ctx, genesisSystem, terrPrompt, cfg.GodAgent.Model, 0, cfg.GodAgent.Temperature)
				if err != nil {
					lastErr = fmt.Errorf("LLM call: %w", err)
					log.Printf("[KingdomGenesis] %s attempt %d/%d failed: %v", t.Name, attempt, maxRetries, lastErr)
					continue
				}
				terrData, err := llm.ExtractJSON(terrResp.Content)
				if err != nil {
					lastErr = fmt.Errorf("JSON parse: %w", err)
					log.Printf("[KingdomGenesis] %s attempt %d/%d failed: %v", t.Name, attempt, maxRetries, lastErr)
					continue
				}
				results[idx] = terrResult{idx: idx, territory: t, data: terrData}
				if attempt > 1 {
					log.Printf("[KingdomGenesis] %s succeeded on attempt %d", t.Name, attempt)
				}
				return
			}
			results[idx] = terrResult{idx: idx, territory: t, err: lastErr}
		}(i, territory)
	}
	wg.Wait()
	log.Printf("[KingdomGenesis] All %d territory LLM calls complete", len(territories))

	// Process results sequentially (modifies shared world state)
	for ri, res := range results {
		territory := res.territory
		progress("Populating the Land", fmt.Sprintf("Building %s...", territory.Name), 2+ri, territoryCount+4)
		if res.err != nil {
			log.Printf("[KingdomGenesis] %s failed (%v), using procedural fallback", territory.Name, res.err)
			globalNPCIdx = generateProceduralTerritory(w, territory, npcsPerTerritory, citiesPerTerritory, globalNPCIdx, cfg)
			continue
		}

		terrData := res.data
		validateGenesis(terrData, npcsPerTerritory)

		locs := extractLocations(terrData)
		npcs := extractNPCs(terrData)
		log.Printf("[KingdomGenesis] %s: %d locations, %d LLM NPCs (filling to %d)", territory.Name, len(locs), len(npcs), npcsPerTerritory)

		// Parse city names
		cityNames := []string{}
		if cities, ok := terrData["cities"].([]interface{}); ok {
			for _, c := range cities {
				if s, ok := c.(string); ok {
					cityNames = append(cityNames, s)
				}
			}
		}
		for len(cityNames) < citiesPerTerritory {
			cityNames = append(cityNames, fmt.Sprintf("%s Town %d", territory.Name, len(cityNames)+1))
		}

		// Layout locations within this territory's area
		terrLayout := layoutTerritoryLocations(locs, territory, citiesPerTerritory, npcsPerTerritory, gridW, gridH, w.Locations)

		// Tag all locations with territory ID and add to world
		for _, loc := range terrLayout.locations {
			loc.TerritoryID = territory.ID
			w.Locations = append(w.Locations, loc)
		}
		allRoads = append(allRoads, terrLayout.roads...)

		// Add filler NPCs if needed
		allNPCData := npcs
		if len(allNPCData) < npcsPerTerritory {
			extra := generateFillerNPCs(npcsPerTerritory - len(allNPCData))
			allNPCData = append(allNPCData, extra...)
		}

		// Assign NPCs to locations
		homeLocations := make([]*world.Location, 0)
		var scatterLocs []*world.Location
		for _, l := range terrLayout.locations {
			if l.Type == "home" {
				homeLocations = append(homeLocations, l)
			} else if l.Type != "inn" {
				scatterLocs = append(scatterLocs, l)
			}
		}
		if len(scatterLocs) == 0 {
			scatterLocs = terrLayout.locations
		}
		if len(scatterLocs) == 0 {
			// Last resort: use all world locations
			scatterLocs = w.Locations
		}
		if len(scatterLocs) == 0 {
			log.Printf("[Genesis] WARNING: no locations available for territory %s, skipping NPC placement", territory.Name)
			globalNPCIdx += len(allNPCData)
			continue
		}

		homeIdx := 0
		for i, npcData := range allNPCData {
			getsHome := homeIdx < len(homeLocations) && i < len(homeLocations)

			if getsHome {
				home := homeLocations[homeIdx]
				homeIdx++
				npcData["homeId"] = home.ID
				npcData["locationId"] = home.ID
				home.Name = fmt.Sprintf("%s's Home", npcData["name"])
				home.BuildingType = "shack"
				home.BuildingDurability = 80
			} else {
				pick := scatterLocs[rand.Intn(len(scatterLocs))]
				npcData["homeId"] = ""
				npcData["locationId"] = pick.ID
			}

			newNPC := npc.CreateFromGenesis(npcData, globalNPCIdx+i, w.GameDay, cfg)
			newNPC.TerritoryID = territory.ID
			if getsHome && homeIdx > 0 {
				homeLocations[homeIdx-1].BuildingOwnerID = newNPC.ID
				newNPC.HomeBuildingID = homeLocations[homeIdx-1].ID
			}
			w.NPCs = append(w.NPCs, newNPC)
		}
		log.Printf("[KingdomGenesis] %s: placed %d NPCs", territory.Name, len(allNPCData))
		globalNPCIdx += len(allNPCData)
	}

	w.Roads = allRoads

	progress("Carving Roads", "Connecting the territories with roads and highways...", territoryCount+2, territoryCount+4)
	log.Printf("[KingdomGenesis] Generating inter-territory roads...")

	// Add inter-territory roads
	w.Roads = append(w.Roads, generateInterTerritoryRoads(w)...)

	// Initialize resources for all locations
	for _, loc := range w.Locations {
		world.InitLocationResources(loc)
	}

	progress("Crowning Rulers", "Establishing the noble hierarchy...", territoryCount+3, territoryCount+4)
	log.Printf("[KingdomGenesis] Creating noble hierarchy...")

	// Step 4: Create noble hierarchy
	createNobleHierarchy(w, territories, metaData, cfg)

	// Assign profession-based workplaces
	for _, n := range w.NPCs {
		locType, ok := ProfLocMap[n.Profession]
		if !ok {
			continue
		}
		for _, l := range w.Locations {
			if l.Type == locType && l.OwnerID == "" && l.TerritoryID == n.TerritoryID {
				l.OwnerID = n.ID
				n.IsBusinessOwner = true
				n.WorkplaceID = l.ID
				break
			}
		}
	}

	// Assign workers to owned workplaces
	for _, n := range w.NPCs {
		if n.IsBusinessOwner || n.WorkplaceID != "" {
			continue
		}
		locType, ok := ProfLocMap[n.Profession]
		if !ok {
			continue
		}
		for _, l := range w.Locations {
			if l.Type == locType && l.OwnerID != "" && l.TerritoryID == n.TerritoryID {
				owner := w.FindNPCByID(l.OwnerID)
				if owner != nil {
					n.EmployerID = owner.ID
					n.WorkplaceID = l.ID
					n.Wage = 2
					break
				}
			}
		}
	}

	// Give profession-specific starting items
	for _, n := range w.NPCs {
		switch n.Profession {
		case "blacksmith":
			n.AddItem("iron ore", 4)
			n.AddItem("iron ingot", 2)
		case "miner":
			n.AddItem("pickaxe", 1)
			n.EquipItem("pickaxe")
		case "carpenter":
			n.AddItem("iron axe", 1)
			n.EquipItem("iron axe")
			n.AddItem("logs", 2)
			n.AddItem("rope", 2)
		case "tailor":
			n.AddItem("leather", 2)
			n.AddItem("cloth", 2)
		case "guard", "knight":
			n.AddItem("iron sword", 1)
			n.EquipItem("iron sword")
			n.AddItem("leather armor", 1)
			n.EquipItem("leather armor")
		case "stablehand":
			n.AddItem("hay", 5)
		}
	}

	// Give merchants a starting horse and cart for trade transport
	for _, n := range w.NPCs {
		if n.Profession != "merchant" || n.MountID != "" {
			continue
		}
		mountID := fmt.Sprintf("mount_merchant_%s_%d", n.ID, len(w.Mounts))
		tmpl := world.MountTemplates["horse"]
		mount := &world.Mount{
			ID: mountID, Name: fmt.Sprintf("%s's horse", n.Name),
			Type: tmpl.Type, OwnerID: n.ID, LocationID: n.LocationID,
			HP: tmpl.HP, MaxHP: tmpl.MaxHP, Speed: tmpl.Speed,
			CarryWeight: tmpl.CarryWeight, Hunger: 100, Grooming: 100,
			Price: tmpl.Price, Alive: true,
		}
		w.Mounts = append(w.Mounts, mount)
		n.MountID = mountID

		carriageID := fmt.Sprintf("carriage_merchant_%s_%d", n.ID, len(w.Carriages))
		cartTmpl := world.CarriageTemplates["cart"]
		carriage := &world.Carriage{
			ID: carriageID, Name: fmt.Sprintf("%s's cart", n.Name),
			OwnerID: n.ID, HorseID: mountID, LocationID: n.LocationID,
			CargoSlots: cartTmpl.CargoSlots, CargoWeight: cartTmpl.CargoWeight,
			Durability: 100, Price: cartTmpl.Price,
		}
		w.Carriages = append(w.Carriages, carriage)
		n.CarriageID = carriageID
	}

	// Add all NPCs to spawn queue, nobles first
	nobleQueue := make([]world.SpawnEntry, 0)
	commonerQueue := make([]world.SpawnEntry, 0)
	for _, n := range w.NPCs {
		entry := world.SpawnEntry{
			NPCID:      n.ID,
			Name:       n.Name,
			Profession: n.Profession,
		}
		if n.IsNoble() {
			nobleQueue = append(nobleQueue, entry)
		} else {
			commonerQueue = append(commonerQueue, entry)
		}
	}
	w.SpawnQueue = append(nobleQueue, commonerQueue...)
	log.Printf("[KingdomGenesis] Spawn queue: %d nobles first, then %d commoners", len(nobleQueue), len(commonerQueue))

	// Rebuild indexes
	w.RebuildLocationIndex()
	w.RebuildNPCIndex()

	progress("Genesis Complete", fmt.Sprintf("%d territories, %d locations, %d NPCs", len(w.Territories), len(w.Locations), len(w.NPCs)), territoryCount+4, territoryCount+4)
	log.Printf("[KingdomGenesis] Created kingdom with %d territories, %d locations, %d NPCs",
		len(w.Territories), len(w.Locations), len(w.NPCs))
	return lore, nil
}

// layoutTerritoryLocations places locations within a territory's region of the grid.
func layoutTerritoryLocations(locs []map[string]interface{}, territory *world.Territory, citiesPerTerritory, npcsPerTerritory, gridW, gridH int, existingLocs []*world.Location) layoutResult {
	// Define a sub-grid region for this territory
	radius := territory.Radius
	minX := max(2, territory.CenterX-radius)
	maxX := min(gridW-2, territory.CenterX+radius)
	minY := max(2, territory.CenterY-radius)
	maxY := min(gridH-2, territory.CenterY+radius)

	subW := maxX - minX
	subH := maxY - minY

	// Pick city centers within this territory
	cityCenters := make([]cityCenter, 0, citiesPerTerritory)
	// First city at territory center
	cityCenters = append(cityCenters, cityCenter{territory.CenterX, territory.CenterY})
	// Additional cities spread around
	for i := 1; i < citiesPerTerritory; i++ {
		angle := float64(i) * (2 * math.Pi / float64(citiesPerTerritory))
		cx := territory.CenterX + int(float64(radius/2)*math.Cos(angle))
		cy := territory.CenterY + int(float64(radius/2)*math.Sin(angle))
		cx = max(minX+3, min(maxX-3, cx))
		cy = max(minY+3, min(maxY-3, cy))
		cityCenters = append(cityCenters, cityCenter{cx, cy})
	}

	// Build occupation grid including existing locations
	grid := make([][]bool, gridH)
	for i := range grid {
		grid[i] = make([]bool, gridW)
	}
	for _, loc := range existingLocs {
		x0 := max(0, loc.X-1)
		y0 := max(0, loc.Y-1)
		x1 := min(gridW, loc.X+loc.W+1)
		y1 := min(gridH, loc.Y+loc.H+1)
		for dy := y0; dy < y1; dy++ {
			for dx := x0; dx < x1; dx++ {
				grid[dy][dx] = true
			}
		}
	}

	fits := func(x, y, w, h int) bool {
		if x < minX || y < minY || x+w > maxX || y+h > maxY {
			return false
		}
		bx0 := max(0, x-1)
		by0 := max(0, y-1)
		bx1 := min(gridW, x+w+1)
		by1 := min(gridH, y+h+1)
		for dy := by0; dy < by1; dy++ {
			for dx := bx0; dx < bx1; dx++ {
				if grid[dy][dx] {
					return false
				}
			}
		}
		return true
	}

	markPlaced := func(x, y, w, h int) {
		x0 := max(0, x-1)
		y0 := max(0, y-1)
		x1 := min(gridW, x+w+1)
		y1 := min(gridH, y+h+1)
		for dy := y0; dy < y1; dy++ {
			for dx := x0; dx < x1; dx++ {
				grid[dy][dx] = true
			}
		}
	}

	findNear := func(cx, cy, w, h, minR, maxR int) (int, int, bool) {
		// Reservoir sampling: pick a random valid spot without collecting all candidates
		bestX, bestY := 0, 0
		found := 0
		minR2 := minR * minR
		maxR2 := maxR * maxR
		for y := max(minY, cy-maxR); y <= min(maxY-h, cy+maxR); y++ {
			for x := max(minX, cx-maxR); x <= min(maxX-w, cx+maxR); x++ {
				dx := x + w/2 - cx
				dy := y + h/2 - cy
				d2 := dx*dx + dy*dy
				if d2 < minR2 || d2 > maxR2 {
					continue
				}
				if !fits(x, y, w, h) {
					continue
				}
				found++
				if rand.Intn(found) == 0 {
					bestX, bestY = x, y
				}
			}
		}
		if found == 0 {
			return 0, 0, false
		}
		return bestX, bestY, true
	}

	usedIDs := make(map[string]bool)
	for _, loc := range existingLocs {
		usedIDs[loc.ID] = true
	}
	genID := func(t string) string {
		prefix := territory.ID + "_" + t
		if !usedIDs[prefix] {
			usedIDs[prefix] = true
			return prefix
		}
		for i := 2; ; i++ {
			id := fmt.Sprintf("%s_%d", prefix, i)
			if !usedIDs[id] {
				usedIDs[id] = true
				return id
			}
		}
	}

	// Generate roads from city centers
	roadGrid := make([][]bool, gridH)
	for i := range roadGrid {
		roadGrid[i] = make([]bool, gridW)
	}
	var roads []world.RoadSegment
	for _, c := range cityCenters {
		// Horizontal road
		roadLen := 10 + rand.Intn(5)
		x1 := max(minX, c.x-roadLen/2)
		x2 := min(maxX-1, c.x+roadLen/2)
		for x := x1; x <= x2; x++ {
			if c.y >= 0 && c.y < gridH {
				roadGrid[c.y][x] = true
			}
		}
		roads = append(roads, world.RoadSegment{X1: x1, Y1: c.y, X2: x2, Y2: c.y})

		// Vertical road
		crossLen := 8 + rand.Intn(5)
		y1 := max(minY, c.y-crossLen/2)
		y2 := min(maxY-1, c.y+crossLen/2)
		for y := y1; y <= y2; y++ {
			if c.x >= 0 && c.x < gridW {
				roadGrid[y][c.x] = true
			}
		}
		roads = append(roads, world.RoadSegment{X1: c.x, Y1: y1, X2: c.x, Y2: y2})
	}

	var placed []*world.Location

	// Place each location
	for i, l := range locs {
		t, _ := l["type"].(string)
		name, _ := l["name"].(string)
		desc, _ := l["description"].(string)
		size := TypeSizes[t]
		if size == [2]int{} {
			size = [2]int{2, 1}
		}

		c := cityCenters[i%len(cityCenters)]
		x, y, ok := findNear(c.x, c.y, size[0], size[1], 0, 8)
		if !ok {
			x, y, ok = findNear(c.x, c.y, size[0], size[1], 0, 15)
		}
		if !ok {
			x, y, ok = findNear(territory.CenterX, territory.CenterY, size[0], size[1], 0, radius)
		}
		if !ok {
			continue
		}
		loc := &world.Location{
			ID: genID(t), Name: name, Type: t,
			X: x, Y: y, W: size[0], H: size[1],
			Color: locationColor(t), Description: desc,
			Capacity: world.CalcCapacity(t, size[0], size[1]),
		}
		placed = append(placed, loc)
		markPlaced(x, y, size[0], size[1])
	}

	// Place homes
	homeCount := max(2, int(float64(npcsPerTerritory)*0.4))
	for i := 0; i < homeCount; i++ {
		homeID := genID("home")
		c := cityCenters[i%len(cityCenters)]
		x, y, ok := findNear(c.x, c.y, 1, 1, 1, 6)
		if !ok {
			x, y, ok = findNear(c.x, c.y, 1, 1, 0, 12)
		}
		if !ok {
			continue
		}
		loc := &world.Location{
			ID: homeID, Name: fmt.Sprintf("Home %d", i+1), Type: "home",
			X: x, Y: y, W: 1, H: 1,
			Color: locationColor("home"), Capacity: 2,
		}
		placed = append(placed, loc)
		markPlaced(x, y, 1, 1)
	}

	_ = subW
	_ = subH
	_ = roadGrid

	return layoutResult{locations: placed, gridW: gridW, gridH: gridH, roads: roads}
}

// generateProceduralTerritory creates a territory's content without LLM.
func generateProceduralTerritory(w *world.World, territory *world.Territory, npcCount, cityCount, startIdx int, cfg *config.Config) int {
	// Generate minimal required locations procedurally
	var locs []map[string]interface{}
	for _, req := range RequiredTypes {
		locs = append(locs, map[string]interface{}{
			"name":        fmt.Sprintf("%s %s", territory.Name, strings.Title(req)),
			"type":        req,
			"description": fmt.Sprintf("A %s in %s.", req, territory.Name),
		})
	}
	// Add territory-specific buildings
	if territory.Type == world.TerritoryTypeRoyalDomain {
		locs = append(locs, map[string]interface{}{
			"name": "Royal Palace", "type": "palace",
			"description": "The magnificent palace of the royal family.",
		})
	} else {
		locs = append(locs, map[string]interface{}{
			"name": fmt.Sprintf("%s Castle", territory.Name), "type": "castle",
			"description": fmt.Sprintf("The fortified castle of the Duke of %s.", territory.Name),
		})
	}
	locs = append(locs, map[string]interface{}{
		"name": fmt.Sprintf("%s Stables", territory.Name), "type": "stable",
		"description": "A stable for horses and carriages.",
	})
	locs = append(locs, map[string]interface{}{
		"name": fmt.Sprintf("%s Barracks", territory.Name), "type": "barracks",
		"description": "Quarters for the territory's guard force.",
	})
	// Add extra wells
	for i := 0; i < 3; i++ {
		locs = append(locs, map[string]interface{}{
			"name": fmt.Sprintf("%s Well %d", territory.Name, i+1), "type": "well",
			"description": "A freshwater well.",
		})
	}

	layout := layoutTerritoryLocations(locs, territory, cityCount, npcCount, w.GridW, w.GridH, w.Locations)
	for _, loc := range layout.locations {
		loc.TerritoryID = territory.ID
		w.Locations = append(w.Locations, loc)
	}
	w.Roads = append(w.Roads, layout.roads...)

	// Generate filler NPCs
	fillerNPCs := generateFillerNPCs(npcCount)
	var scatterLocs []*world.Location
	for _, l := range layout.locations {
		if l.Type != "home" && l.Type != "inn" {
			scatterLocs = append(scatterLocs, l)
		}
	}
	if len(scatterLocs) == 0 {
		scatterLocs = layout.locations
	}
	if len(scatterLocs) == 0 {
		scatterLocs = w.Locations
	}

	for i, npcData := range fillerNPCs {
		if len(scatterLocs) == 0 {
			break
		}
		pick := scatterLocs[rand.Intn(len(scatterLocs))]
		npcData["locationId"] = pick.ID
		newNPC := npc.CreateFromGenesis(npcData, startIdx+i, w.GameDay, cfg)
		newNPC.TerritoryID = territory.ID
		w.NPCs = append(w.NPCs, newNPC)
	}

	return startIdx + len(fillerNPCs)
}

// generateInterTerritoryRoads creates road segments connecting adjacent territories.
func generateInterTerritoryRoads(w *world.World) []world.RoadSegment {
	var roads []world.RoadSegment
	if len(w.Territories) < 2 {
		return roads
	}

	// Connect each outer territory to the center territory
	center := w.Territories[0]
	for _, t := range w.Territories[1:] {
		roads = append(roads, world.RoadSegment{
			X1: center.CenterX, Y1: center.CenterY,
			X2: t.CenterX, Y2: t.CenterY,
		})
	}

	// Connect adjacent outer territories (ring)
	outer := w.Territories[1:]
	for i := 0; i < len(outer); i++ {
		next := (i + 1) % len(outer)
		roads = append(roads, world.RoadSegment{
			X1: outer[i].CenterX, Y1: outer[i].CenterY,
			X2: outer[next].CenterX, Y2: outer[next].CenterY,
		})
	}

	return roads
}

// createNobleHierarchy creates the royal family, dukes, and subordinate nobles.
func createNobleHierarchy(w *world.World, territories []*world.Territory, metaData map[string]interface{}, cfg *config.Config) {
	npcIdx := len(w.NPCs)

	// Find the royal domain
	var royalTerritory *world.Territory
	for _, t := range territories {
		if t.Type == world.TerritoryTypeRoyalDomain {
			royalTerritory = t
			break
		}
	}
	if royalTerritory == nil {
		return
	}

	// Find palace location
	palaceLoc := ""
	for _, l := range w.Locations {
		if l.Type == "palace" && l.TerritoryID == royalTerritory.ID {
			palaceLoc = l.ID
			break
		}
	}
	if palaceLoc == "" {
		// Fallback: use any castle or first location in royal territory
		for _, l := range w.Locations {
			if l.TerritoryID == royalTerritory.ID && (l.Type == "castle" || l.Type == "market") {
				palaceLoc = l.ID
				break
			}
		}
	}

	// Get queen name from meta
	queenName := "Queen Elara"
	terrMeta, _ := metaData["territories"].([]interface{})
	if len(terrMeta) > 0 {
		if tm, ok := terrMeta[0].(map[string]interface{}); ok {
			if qn, ok := tm["queen_name"].(string); ok && qn != "" {
				queenName = qn
			}
		}
	}

	// Create King
	kingName := royalTerritory.RulerName
	if kingName == "" {
		kingName = "King Aldric"
	}
	king := createNobleNPC(npcIdx, kingName, "king", "king", palaceLoc, royalTerritory.ID, cfg, w.GameDay)
	w.NPCs = append(w.NPCs, king)
	royalTerritory.RulerID = king.ID
	npcIdx++

	// Create Queen
	queen := createNobleNPC(npcIdx, queenName, "queen", "queen", palaceLoc, royalTerritory.ID, cfg, w.GameDay)
	queen.SpouseID = king.ID
	king.SpouseID = queen.ID
	w.NPCs = append(w.NPCs, queen)
	npcIdx++

	// Create 1-2 royal children
	childCount := 1 + rand.Intn(2)
	childProfs := []string{"prince", "princess"}
	for i := 0; i < childCount; i++ {
		prof := childProfs[i%2]
		childName := fillerFirstNames[rand.Intn(len(fillerFirstNames))]
		child := createNobleNPC(npcIdx, childName, prof, prof, palaceLoc, royalTerritory.ID, cfg, w.GameDay)
		child.ParentID = king.ID
		king.ChildIDs = append(king.ChildIDs, child.ID)
		child.LiegeID = king.ID
		w.NPCs = append(w.NPCs, child)
		npcIdx++
	}

	// Create Dukes for each duchy
	for _, territory := range territories {
		if territory.Type != world.TerritoryTypeDuchy {
			continue
		}
		// Find castle location
		castleLoc := ""
		for _, l := range w.Locations {
			if l.Type == "castle" && l.TerritoryID == territory.ID {
				castleLoc = l.ID
				break
			}
		}
		if castleLoc == "" {
			for _, l := range w.Locations {
				if l.TerritoryID == territory.ID {
					castleLoc = l.ID
					break
				}
			}
		}

		dukeName := territory.RulerName
		if dukeName == "" {
			dukeName = fillerFirstNames[rand.Intn(len(fillerFirstNames))] + " " + fillerLastNames[rand.Intn(len(fillerLastNames))]
		}
		duke := createNobleNPC(npcIdx, dukeName, "duke", "duke", castleLoc, territory.ID, cfg, w.GameDay)
		duke.LiegeID = king.ID
		king.VassalIDs = append(king.VassalIDs, duke.ID)
		w.NPCs = append(w.NPCs, duke)
		territory.RulerID = duke.ID
		npcIdx++

		// Create 1-2 Counts under each Duke
		countCount := 1 + rand.Intn(2)
		for j := 0; j < countCount; j++ {
			countName := fillerFirstNames[rand.Intn(len(fillerFirstNames))] + " " + fillerLastNames[rand.Intn(len(fillerLastNames))]
			// Find a manor or use castle
			manorLoc := castleLoc
			for _, l := range w.Locations {
				if l.Type == "manor" && l.TerritoryID == territory.ID && l.OwnerID == "" {
					manorLoc = l.ID
					break
				}
			}
			count := createNobleNPC(npcIdx, countName, "count", "count", manorLoc, territory.ID, cfg, w.GameDay)
			count.LiegeID = duke.ID
			duke.VassalIDs = append(duke.VassalIDs, count.ID)
			w.NPCs = append(w.NPCs, count)
			npcIdx++
		}

		// Create 2-3 Barons under each Duke
		baronCount := 2 + rand.Intn(2)
		for j := 0; j < baronCount; j++ {
			baronName := fillerFirstNames[rand.Intn(len(fillerFirstNames))] + " " + fillerLastNames[rand.Intn(len(fillerLastNames))]
			baron := createNobleNPC(npcIdx, baronName, "baron", "baron", castleLoc, territory.ID, cfg, w.GameDay)
			baron.LiegeID = duke.ID
			duke.VassalIDs = append(duke.VassalIDs, baron.ID)
			w.NPCs = append(w.NPCs, baron)
			npcIdx++
		}

		// Create 2-4 Knights per territory
		knightCount := 2 + rand.Intn(3)
		barracksLoc := castleLoc
		for _, l := range w.Locations {
			if l.Type == "barracks" && l.TerritoryID == territory.ID {
				barracksLoc = l.ID
				break
			}
		}
		for j := 0; j < knightCount; j++ {
			knightName := fillerFirstNames[rand.Intn(len(fillerFirstNames))] + " " + fillerLastNames[rand.Intn(len(fillerLastNames))]
			knight := createNobleNPC(npcIdx, knightName, "knight", "knight", barracksLoc, territory.ID, cfg, w.GameDay)
			knight.LiegeID = duke.ID
			knight.AddItem("iron sword", 1)
			knight.EquipItem("iron sword")
			knight.AddItem("leather armor", 1)
			knight.EquipItem("leather armor")
			w.NPCs = append(w.NPCs, knight)
			npcIdx++
		}
	}

	// Create Royal Court faction — all nobles belong to it from the start
	var courtMembers []string
	var kingNPC *npc.NPC
	for _, n := range w.NPCs {
		if n.IsNoble() {
			courtMembers = append(courtMembers, n.ID)
			if n.NobleRank == "king" {
				kingNPC = n
			}
		}
	}
	if kingNPC != nil && len(courtMembers) > 0 {
		court := faction.NewFaction("political", kingNPC.ID, kingNPC.Name, courtMembers, w.GameDay)
		court.ID = "faction_royal_court"
		court.Name = "The Royal Court"
		w.Factions = append(w.Factions, court)
		for _, n := range w.NPCs {
			if n.IsNoble() {
				n.FactionID = court.ID
			}
		}
		log.Printf("[Genesis] Created Royal Court faction with %d members, led by %s", len(courtMembers), kingNPC.Name)
	}

	// Give all nobles starting mounts — nobles would never walk on foot
	mountCount := 0
	for _, n := range w.NPCs {
		if !n.IsNoble() || n.MountID != "" {
			continue
		}
		// Pick mount type by rank
		var mountType string
		switch n.NobleRank {
		case "king", "queen", "duke", "prince", "princess":
			mountType = "war_horse"
		default:
			mountType = "horse"
		}
		tmpl := world.MountTemplates[mountType]
		mountID := fmt.Sprintf("mount_noble_%s_%d", n.ID, len(w.Mounts))
		mount := &world.Mount{
			ID:          mountID,
			Name:        fmt.Sprintf("%s's %s", n.Name, mountType),
			Type:        tmpl.Type,
			OwnerID:     n.ID,
			LocationID:  n.LocationID,
			HP:          tmpl.HP,
			MaxHP:       tmpl.MaxHP,
			Speed:       tmpl.Speed,
			CarryWeight: tmpl.CarryWeight,
			Hunger:      100,
			Grooming:    100,
			Price:       tmpl.Price,
			Alive:       true,
		}
		w.Mounts = append(w.Mounts, mount)
		n.MountID = mountID
		mountCount++
	}
	if mountCount > 0 {
		log.Printf("[Genesis] Gave %d nobles their starting mounts", mountCount)
	}
}

// createNobleNPC creates an NPC with boosted noble-appropriate stats.
func createNobleNPC(idx int, name, profession, rank, locationID, territoryID string, cfg *config.Config, gameDay int) *npc.NPC {
	data := map[string]interface{}{
		"name":        name,
		"age":         float64(25 + rand.Intn(35)),
		"profession":  profession,
		"personality": "authoritative, intelligent, ambitious",
		"items":       []interface{}{"gold", "gold", "gold", "bread"},
		"locationId":  locationID,
	}
	n := npc.CreateFromGenesis(data, idx, gameDay, cfg)
	n.NobleRank = rank
	n.TerritoryID = territoryID
	n.LocationID = locationID

	// Boost noble-appropriate stats
	n.Stats.Intelligence = max(n.Stats.Intelligence, 65+rand.Intn(20))
	n.Stats.Charisma = max(n.Stats.Charisma, 60+rand.Intn(20))
	n.Stats.Dominance = max(n.Stats.Dominance, 55+rand.Intn(25))
	n.Stats.PoliticalInstinct = max(n.Stats.PoliticalInstinct, 60+rand.Intn(20))
	n.Stats.StrategicThinking = max(n.Stats.StrategicThinking, 55+rand.Intn(25))
	n.Stats.Ambition = max(n.Stats.Ambition, 60+rand.Intn(20))
	n.Stats.Wisdom = max(n.Stats.Wisdom, 50+rand.Intn(25))
	n.Literacy = max(n.Literacy, 70)

	// Nobles start with more gold
	switch rank {
	case "king", "queen":
		n.AddItem("gold", 500)
	case "duke":
		n.AddItem("gold", 200)
	case "count":
		n.AddItem("gold", 100)
	case "baron":
		n.AddItem("gold", 50)
	case "knight":
		n.AddItem("gold", 30)
	}

	return n
}
