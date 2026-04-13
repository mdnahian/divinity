package world

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"

	"github.com/divinity/core/building"
	"github.com/divinity/core/enemy"
	"github.com/divinity/core/faction"
	"github.com/divinity/core/knowledge"
	"github.com/divinity/core/npc"
)

var WeatherTypes = []string{"clear", "cloudy", "rain", "storm"}
var WeatherWeights = []float64{0.40, 0.30, 0.20, 0.10}

type RoadSegment struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

type GroundItem struct {
	Name       string  `json:"name"`
	Qty        int     `json:"qty"`
	LocationID string  `json:"locationId"`
	DroppedDay int     `json:"droppedDay"`
	Durability float64 `json:"durability"`
}

// Trap represents a hunting trap placed at a location by an NPC.
type Trap struct {
	OwnerID    string `json:"ownerId"`
	LocationID string `json:"locationId"`
	SetDay     int    `json:"setDay"`
	Caught     string `json:"caught"` // item name caught, empty if nothing yet
	CaughtQty  int    `json:"caughtQty"`
}

type ActiveEvent struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	TicksLeft int                    `json:"ticksLeft"`
	Details   map[string]interface{} `json:"details"`
}

type Prayer struct {
	NpcName string  `json:"npcName"`
	NpcID   string  `json:"npcId"`
	Prayer  string  `json:"prayer"`
	Time    string  `json:"time"`
	HP      int     `json:"hp"`
	Stress  int     `json:"stress"`
	Hunger  float64 `json:"hunger"`
}

type EventEntry struct {
	Text        string `json:"text"`
	Type        string `json:"type"`
	Time        string `json:"time"`
	Tick        int64  `json:"tick"`
	NpcID       string `json:"npcId,omitempty"`
	TerritoryID string `json:"territoryId,omitempty"`
}

type SpawnEntry struct {
	NPCID      string `json:"npcId"`
	Name       string `json:"name"`
	Profession string `json:"profession"`
}

type PendingDivineDream struct {
	NPCID      string   `json:"npcId"`
	Text       string   `json:"text"`
	Importance float64  `json:"importance"`
	Vividness  float64  `json:"vividness"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
}

// BiomeOverride allows the GOD AI to dynamically reshape terrain.
type BiomeOverride struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	BiomeType string `json:"biomeType"`
	Radius    int    `json:"radius"`
}

type ChronicleEntry struct {
	Day   int    `json:"day"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type World struct {
	Mu        sync.RWMutex `json:"-"`
	ID        string       `json:"id"`
	GridW     int          `json:"gridW"`
	GridH     int          `json:"gridH"`
	MinGridW  int          `json:"minGridW"`
	MinGridH  int          `json:"minGridH"`
	Locations []*Location    `json:"locations"`
	Roads     []RoadSegment  `json:"roads"`
	NPCs      []*npc.NPC     `json:"npcs"`
	SpawnQueue []SpawnEntry `json:"spawnQueue"`

	// Kingdom structure
	Territories    []*Territory    `json:"territories"`
	Dungeons       []*Dungeon      `json:"dungeons"`
	Mounts         []*Mount        `json:"mounts"`
	Carriages      []*Carriage     `json:"carriages"`
	BiomeOverrides []BiomeOverride `json:"biomeOverrides"`

	GameDay    int    `json:"gameDay"`
	GameHour   int    `json:"gameHour"`
	GameMinute int    `json:"gameMinute"`
	Weather    string `json:"weather"`
	PrevDay    int    `json:"-"`

	LastWeatherChange int `json:"-"`

	EventSeq         int64                      `json:"eventSeq"`
	EventLog         []EventEntry               `json:"eventLog"`
	GroundItems      []GroundItem               `json:"groundItems"`
	Traps            []Trap                     `json:"traps"`
	Factions         []*faction.Faction         `json:"factions"`
	FactionContracts []*faction.FactionContract `json:"factionContracts"`
	Economy          map[string]EconomyEntry    `json:"economy"`
	Constructions    []*building.Construction   `json:"constructions"`
	Techniques       []*knowledge.Technique     `json:"techniques"`
	ActiveEvents     []ActiveEvent              `json:"activeEvents"`
	Enemies             []*enemy.Enemy             `json:"enemies"`
	Treasury            int                        `json:"treasury"`
	RecentPrayers       []Prayer                   `json:"recentPrayers"`
	PendingDivineDreams []PendingDivineDream       `json:"pendingDivineDreams"`
	Chronicles          []ChronicleEntry           `json:"chronicles"`

	BasePrices    map[string]int `json:"-"`
	MinMultiplier float64        `json:"-"`
	MaxMultiplier float64        `json:"-"`

	// Index maps for O(1) lookups — rebuilt on mutation
	locationIndex map[string]*Location `json:"-"`
	npcIndex      map[string]*npc.NPC  `json:"-"`
}

func NewWorld(basePrices map[string]int, minMult, maxMult float64) *World {
	return &World{
		GridW: 11, GridH: 9,
		MinGridW: 11, MinGridH: 9,
		Locations:     make([]*Location, 0),
		Roads:         make([]RoadSegment, 0),
		NPCs:          make([]*npc.NPC, 0),
		SpawnQueue:     make([]SpawnEntry, 0),
		Territories:    make([]*Territory, 0),
		Dungeons:       make([]*Dungeon, 0),
		Mounts:         make([]*Mount, 0),
		Carriages:      make([]*Carriage, 0),
		BiomeOverrides: make([]BiomeOverride, 0),
		GameDay:       1,
		GameHour:      6,
		GameMinute:    0,
		Weather:       "clear",
		PrevDay:       1,
		EventLog:      make([]EventEntry, 0),
		GroundItems:   make([]GroundItem, 0),
		Factions:      make([]*faction.Faction, 0),
		Economy:       make(map[string]EconomyEntry),
		Constructions: make([]*building.Construction, 0),
		Techniques:    make([]*knowledge.Technique, 0),
		ActiveEvents:  make([]ActiveEvent, 0),
		Enemies:       make([]*enemy.Enemy, 0),
		RecentPrayers:       make([]Prayer, 0),
		PendingDivineDreams: make([]PendingDivineDream, 0),
		Chronicles:          make([]ChronicleEntry, 0),
		BasePrices:    basePrices,
		MinMultiplier: minMult,
		MaxMultiplier: maxMult,
		locationIndex: make(map[string]*Location),
		npcIndex:      make(map[string]*npc.NPC),
	}
}

func (w *World) QueueDivineDream(dream PendingDivineDream) {
	w.PendingDivineDreams = append(w.PendingDivineDreams, dream)
}

func (w *World) PopDivineDreams(npcID string) []PendingDivineDream {
	var result, remaining []PendingDivineDream
	for _, d := range w.PendingDivineDreams {
		if d.NPCID == npcID {
			result = append(result, d)
		} else {
			remaining = append(remaining, d)
		}
	}
	w.PendingDivineDreams = remaining
	return result
}

func (w *World) PopSpawnQueue() *SpawnEntry {
	if len(w.SpawnQueue) == 0 {
		return nil
	}
	entry := w.SpawnQueue[0]
	w.SpawnQueue = w.SpawnQueue[1:]
	return &entry
}

// HasClaimedNPCs returns true if any NPC has ever been claimed (alive or dead).
func (w *World) HasClaimedNPCs() bool {
	for _, n := range w.NPCs {
		if n.Claimed {
			return true
		}
	}
	return false
}

// AliveNPCs returns only claimed (agent-controlled) alive NPCs.
// Unclaimed NPCs are dormant spawn-queue profiles and do not participate in the simulation.
func (w *World) AliveNPCs() []*npc.NPC {
	alive := make([]*npc.NPC, 0)
	for _, n := range w.NPCs {
		if n.Alive && n.Claimed {
			alive = append(alive, n)
		}
	}
	return alive
}

// AllAliveNPCs returns all alive NPCs regardless of claimed status.
// Used for the spawn queue and genesis-time operations.
func (w *World) AllAliveNPCs() []*npc.NPC {
	alive := make([]*npc.NPC, 0)
	for _, n := range w.NPCs {
		if n.Alive {
			alive = append(alive, n)
		}
	}
	return alive
}

func (w *World) AliveEnemies() []*enemy.Enemy {
	alive := make([]*enemy.Enemy, 0)
	for _, e := range w.Enemies {
		if e.Alive {
			alive = append(alive, e)
		}
	}
	return alive
}

func (w *World) EnemiesAtLocation(locID string) []*enemy.Enemy {
	result := make([]*enemy.Enemy, 0)
	for _, e := range w.Enemies {
		if e.Alive && e.LocationID == locID {
			result = append(result, e)
		}
	}
	return result
}

func (w *World) IsLocationFull(locID string) bool {
	loc := w.LocationByID(locID)
	if loc == nil || loc.Capacity <= 0 {
		return false
	}
	return len(w.NPCsAtLocation(locID, "")) >= loc.Capacity
}

func (w *World) NPCsAtLocation(locID, excludeID string) []*npc.NPC {
	result := make([]*npc.NPC, 0)
	for _, n := range w.NPCs {
		if n.Alive && n.Claimed && n.LocationID == locID && n.ID != excludeID {
			result = append(result, n)
		}
	}
	return result
}

func (w *World) AdvanceTime(minutes int) {
	w.PrevDay = w.GameDay
	w.GameMinute += minutes
	for w.GameMinute >= 60 {
		w.GameMinute -= 60
		w.GameHour++
	}
	for w.GameHour >= 24 {
		w.GameHour -= 24
		w.GameDay++
	}
	w.maybeChangeWeather()
}

func (w *World) IsNewDay() bool {
	return w.GameDay != w.PrevDay
}

func (w *World) IsNight() bool {
	return w.GameHour >= 22 || w.GameHour < 6
}

func (w *World) TimeString() string {
	return fmt.Sprintf("Day %d, %02d:%02d", w.GameDay, w.GameHour, w.GameMinute)
}

func (w *World) WeatherDescription() string {
	descs := map[string]string{
		"clear":  "The sky is clear and sunny.",
		"cloudy": "Grey clouds hang overhead.",
		"rain":   "Rain falls steadily across the village.",
		"storm":  "A fierce storm rages with thunder and lightning.",
	}
	return descs[w.Weather]
}

func (w *World) maybeChangeWeather() {
	totalMinutes := w.GameDay*1440 + w.GameHour*60 + w.GameMinute
	if totalMinutes-w.LastWeatherChange > 180 && rand.Float64() < 0.15 {
		w.Weather = weightedPick(WeatherTypes, WeatherWeights)
		w.LastWeatherChange = totalMinutes
	}
}

func weightedPick(items []string, weights []float64) string {
	var total float64
	for _, w := range weights {
		total += w
	}
	r := rand.Float64() * total
	for i, w := range weights {
		r -= w
		if r <= 0 {
			return items[i]
		}
	}
	return items[len(items)-1]
}

// RebuildLocationIndex rebuilds the O(1) location lookup map.
// Call after adding/removing locations.
func (w *World) RebuildLocationIndex() {
	w.locationIndex = make(map[string]*Location, len(w.Locations))
	for _, l := range w.Locations {
		w.locationIndex[l.ID] = l
	}
}

// RebuildNPCIndex rebuilds the O(1) NPC lookup map.
// Call after adding/removing NPCs.
func (w *World) RebuildNPCIndex() {
	w.npcIndex = make(map[string]*npc.NPC, len(w.NPCs))
	for _, n := range w.NPCs {
		w.npcIndex[n.ID] = n
	}
}

func (w *World) LocationByID(id string) *Location {
	if w.locationIndex != nil {
		if l, ok := w.locationIndex[id]; ok {
			return l
		}
	}
	// Fallback to linear scan
	for _, l := range w.Locations {
		if l.ID == id {
			return l
		}
	}
	return nil
}

func (w *World) LocationByName(name string) *Location {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, l := range w.Locations {
		if strings.ToLower(l.Name) == name {
			return l
		}
	}
	return nil
}

// NormalizeLocationName strips prefixes, underscores, camelCase joins, parenthetical suffixes, and lowercases.
func NormalizeLocationName(s string) string {
	s = strings.TrimSpace(s)
	// Strip leading @ characters
	s = strings.TrimLeft(s, "@ ")
	// Strip parenthetical suffixes like "(well)" or "(+15 min travel)"
	if idx := strings.Index(s, " ("); idx >= 0 {
		s = s[:idx]
	}
	// Replace underscores with spaces
	s = strings.ReplaceAll(s, "_", " ")
	// Insert spaces before uppercase letters in camelCase (e.g., "WhisperingHollowMarket" -> "Whispering Hollow Market")
	var buf strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := rune(s[i-1])
			if prev >= 'a' && prev <= 'z' {
				buf.WriteRune(' ')
			}
		}
		buf.WriteRune(r)
	}
	s = buf.String()
	// Collapse multiple spaces and lowercase
	parts := strings.Fields(s)
	return strings.ToLower(strings.Join(parts, " "))
}

// LocationByNameFuzzy tries exact match, then normalized match, then substring containment.
func (w *World) LocationByNameFuzzy(input string) *Location {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// If input contains " or ", try each part independently
	if strings.Contains(strings.ToLower(input), " or ") {
		parts := strings.SplitN(input, " or ", 2)
		for _, part := range parts {
			if loc := w.LocationByNameFuzzy(strings.TrimSpace(part)); loc != nil {
				return loc
			}
		}
	}

	// 1. Exact case-insensitive match
	lower := strings.ToLower(input)
	for _, l := range w.Locations {
		if strings.ToLower(l.Name) == lower {
			return l
		}
	}

	// 2. Normalized match
	norm := NormalizeLocationName(input)
	for _, l := range w.Locations {
		if NormalizeLocationName(l.Name) == norm {
			return l
		}
	}

	// 3. Substring containment — pick the location whose normalized name is contained in or contains the input
	for _, l := range w.Locations {
		locNorm := NormalizeLocationName(l.Name)
		if strings.Contains(norm, locNorm) || strings.Contains(locNorm, norm) {
			return l
		}
	}

	return nil
}

func (w *World) LocationsByType(t string) []*Location {
	result := make([]*Location, 0)
	for _, l := range w.Locations {
		if l.Type == t {
			result = append(result, l)
		}
	}
	return result
}

func (w *World) LogEvent(text, etype string) EventEntry {
	w.EventSeq++
	entry := EventEntry{Text: text, Type: etype, Time: w.TimeString(), Tick: w.EventSeq}
	w.EventLog = append(w.EventLog, entry)
	if len(w.EventLog) > 2000 {
		w.EventLog = w.EventLog[1:]
	}
	return entry
}

func (w *World) LogEventNPC(text, etype, npcID string) EventEntry {
	w.EventSeq++
	entry := EventEntry{Text: text, Type: etype, Time: w.TimeString(), Tick: w.EventSeq, NpcID: npcID}
	w.EventLog = append(w.EventLog, entry)
	if len(w.EventLog) > 2000 {
		w.EventLog = w.EventLog[1:]
	}
	return entry
}

func (w *World) LogEventTerritory(text, etype, territoryID string) EventEntry {
	w.EventSeq++
	entry := EventEntry{Text: text, Type: etype, Time: w.TimeString(), Tick: w.EventSeq, TerritoryID: territoryID}
	w.EventLog = append(w.EventLog, entry)
	if len(w.EventLog) > 2000 {
		w.EventLog = w.EventLog[1:]
	}
	return entry
}

func (w *World) DropItemsOnDeath(n *npc.NPC) {
	count := int(math.Min(3, float64(len(n.Inventory))))
	for i := 0; i < count; i++ {
		it := n.Inventory[i]
		dur := 80.0
		// Gold dropped on the ground scatters and is harder to find over time
		if it.Name == "gold" {
			dur = 30.0
		}
		w.GroundItems = append(w.GroundItems, GroundItem{
			Name:       it.Name,
			Qty:        it.Qty,
			LocationID: n.LocationID,
			DroppedDay: w.GameDay,
			Durability: dur,
		})
	}
	if count > 0 {
		n.Inventory = n.Inventory[count:]
	}
}

func (w *World) GroundItemsAt(locID string) []GroundItem {
	result := make([]GroundItem, 0)
	for _, g := range w.GroundItems {
		if g.LocationID == locID {
			result = append(result, g)
		}
	}
	return result
}

func (w *World) PickUpGroundItem(n *npc.NPC, itemName string) *GroundItem {
	for i, g := range w.GroundItems {
		if g.LocationID == n.LocationID && g.Name == itemName {
			n.AddItem(g.Name, g.Qty)
			picked := w.GroundItems[i]
			w.GroundItems = append(w.GroundItems[:i], w.GroundItems[i+1:]...)
			return &picked
		}
	}
	return nil
}

// PartialPickUpGroundItem picks up a partial quantity from the first matching
// ground item stack, leaving the remainder on the ground.
func (w *World) PartialPickUpGroundItem(n *npc.NPC, itemName string, qty int) *GroundItem {
	for i, g := range w.GroundItems {
		if g.LocationID == n.LocationID && g.Name == itemName {
			if qty >= g.Qty {
				// Take the whole stack
				return w.PickUpGroundItem(n, itemName)
			}
			// Take partial amount
			n.AddItem(g.Name, qty)
			picked := GroundItem{
				Name:       g.Name,
				Qty:        qty,
				LocationID: g.LocationID,
				DroppedDay: g.DroppedDay,
				Durability: g.Durability,
			}
			w.GroundItems[i].Qty -= qty
			return &picked
		}
	}
	return nil
}

func (w *World) DecayGroundItems() {
	for i := len(w.GroundItems) - 1; i >= 0; i-- {
		decay := 3.0
		// Gold decays faster on the ground (scavenged/lost coins degrade quickly)
		if w.GroundItems[i].Name == "gold" {
			decay = 15.0
		}
		w.GroundItems[i].Durability -= decay
		if w.GroundItems[i].Durability <= 0 {
			w.GroundItems = append(w.GroundItems[:i], w.GroundItems[i+1:]...)
		}
	}
}

func (w *World) RegenResources() {
	for _, loc := range w.Locations {
		if loc.Resources == nil || loc.MaxResources == nil {
			continue
		}
		for key, maxVal := range loc.MaxResources {
			regen := 1
			if r, ok := ResourceRegen[key]; ok {
				regen = r
			}
			if regen <= 0 {
				continue
			}
			cur := loc.Resources[key]
			loc.Resources[key] = int(math.Min(float64(maxVal), float64(cur+regen)))
		}
	}
}

// RegenWellWater replenishes well water based on current weather.
// Rain and storms fill wells; clear weather does not.
func (w *World) RegenWellWater() {
	regenMap := map[string]int{
		"rain":   10,
		"storm":  20,
		"cloudy": 2,
		"clear":  0,
	}
	regen := regenMap[w.Weather]
	if regen == 0 {
		return
	}
	for _, loc := range w.Locations {
		if loc.Type != "well" || loc.Resources == nil {
			continue
		}
		maxWater := 150
		if loc.MaxResources != nil {
			if mw, ok := loc.MaxResources["water"]; ok && mw > 0 {
				maxWater = mw
			}
		}
		cur := loc.Resources["water"]
		if cur < maxWater {
			loc.Resources["water"] = min(maxWater, cur+regen)
		}
	}
}

func (w *World) GetLocationOwner(locID string) *npc.NPC {
	loc := w.LocationByID(locID)
	if loc == nil || loc.OwnerID == "" {
		return nil
	}
	for _, n := range w.NPCs {
		if n.ID == loc.OwnerID && n.Alive {
			return n
		}
	}
	return nil
}

func (w *World) GetLocationEmployees(locID string) []*npc.NPC {
	result := make([]*npc.NPC, 0)
	for _, n := range w.AliveNPCs() {
		if n.WorkplaceID == locID && n.EmployerID != "" {
			result = append(result, n)
		}
	}
	return result
}

func (w *World) IsWorkerAt(n *npc.NPC, locID string) bool {
	loc := w.LocationByID(locID)
	if loc == nil {
		return false
	}
	if loc.OwnerID == n.ID {
		return true
	}
	if n.WorkplaceID == locID && n.EmployerID != "" {
		return true
	}
	return false
}

func (w *World) ExpandToFitAll() {
	maxX := w.MinGridW
	maxY := w.MinGridH
	for _, loc := range w.Locations {
		if loc.X+loc.W+2 > maxX {
			maxX = loc.X + loc.W + 2
		}
		if loc.Y+loc.H+2 > maxY {
			maxY = loc.Y + loc.H + 2
		}
	}
	w.GridW = maxX
	w.GridH = maxY
}

func (w *World) FindNPCByID(id string) *npc.NPC {
	if w.npcIndex != nil {
		if n, ok := w.npcIndex[id]; ok {
			return n
		}
	}
	for _, n := range w.NPCs {
		if n.ID == id {
			return n
		}
	}
	return nil
}

func (w *World) FindNPCByName(name string) *npc.NPC {
	for _, n := range w.NPCs {
		if n.Alive && n.Name == name {
			return n
		}
	}
	return nil
}

// FindNPCByNameAtLocation finds an alive NPC by name, preferring NPCs at the
// given location. This avoids resolving the wrong NPC when multiple share a
// name across the world.
func (w *World) FindNPCByNameAtLocation(name, locationID string) *npc.NPC {
	// First pass: look for a match at the specified location
	for _, n := range w.NPCs {
		if n.Alive && n.Name == name && n.LocationID == locationID {
			return n
		}
	}
	// Fallback: global search (same as FindNPCByName)
	for _, n := range w.NPCs {
		if n.Alive && n.Name == name {
			return n
		}
	}
	return nil
}

func (w *World) FindNPCByNameAtLocationExcluding(name, locationID, excludeID string) *npc.NPC {
	// First pass: look for a match at the specified location, excluding one NPC
	for _, n := range w.NPCs {
		if n.Alive && n.Name == name && n.LocationID == locationID && n.ID != excludeID {
			return n
		}
	}
	// Fallback: global search excluding one NPC
	for _, n := range w.NPCs {
		if n.Alive && n.Name == name && n.ID != excludeID {
			return n
		}
	}
	return nil
}

func IsWorkerAtType(n *npc.NPC, locType string, w *World) bool {
	for _, loc := range w.LocationsByType(locType) {
		if w.IsWorkerAt(n, loc.ID) {
			return true
		}
	}
	return false
}

func (w *World) ExpandToFit(x, y, tw, th int) {
	if x+tw+2 > w.GridW {
		w.GridW = x + tw + 2
	}
	if y+th+2 > w.GridH {
		w.GridH = y + th + 2
	}
}

func (w *World) CompleteConstruction(c *building.Construction) *Location {
	bt, ok := building.Types[c.BuildingType]
	if !ok {
		return nil
	}
	locType := bt.Function
	if locType == "shelter" {
		locType = "home"
	}

	x := w.GridW - 2
	y := rand.Intn(max(1, w.GridH-2))
	bw, bh := 2, 1
	w.ExpandToFit(x, y, bw, bh)

	colors := map[string]string{
		"shelter": "#cd853f", "market": "#b8860b", "workshop": "#cd853f",
		"forge": "#d35400", "shrine": "#9370db", "school": "#4169e1", "library": "#4169e1",
	}
	color := colors[bt.Function]
	if color == "" {
		color = "#888888"
	}

	loc := &Location{
		ID:                 c.ID,
		Name:               c.Name,
		Type:               locType,
		X:                  x,
		Y:                  y,
		W:                  bw,
		H:                  bh,
		Color:              color,
		Description:        fmt.Sprintf("A newly built %s.", c.BuildingType),
		Capacity:           CalcCapacity(locType, bw, bh),
		BuildingID:         c.ID,
		BuildingType:       c.BuildingType,
		BuildingDurability: 100,
		BuildingOwnerID:    c.OwnerID,
		Tier:               bt.Tier,
	}
	w.Locations = append(w.Locations, loc)
	return loc
}

func (w *World) DecayBuildings() *Location {
	for i, loc := range w.Locations {
		if loc.BuildingDurability == 0 && loc.BuildingType == "" {
			continue
		}
		if loc.BuildingDurability == 0 {
			continue
		}
		owned := false
		if loc.BuildingOwnerID != "" {
			for _, n := range w.NPCs {
				if n.ID == loc.BuildingOwnerID && n.Alive {
					owned = true
					break
				}
			}
		}
		if owned {
			loc.BuildingDurability -= 0.2
		} else {
			loc.BuildingDurability -= 1
		}
		if loc.BuildingDurability <= 0 {
			collapsed := w.Locations[i]
			w.Locations = append(w.Locations[:i], w.Locations[i+1:]...)
			return collapsed
		}
	}
	return nil
}

// --- Territory helpers ---

func (w *World) TerritoryByID(id string) *Territory {
	for _, t := range w.Territories {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (w *World) TerritoryForLocation(locID string) *Territory {
	loc := w.LocationByID(locID)
	if loc == nil {
		return nil
	}
	if loc.TerritoryID != "" {
		return w.TerritoryByID(loc.TerritoryID)
	}
	return ClosestTerritory(w.Territories, loc.X, loc.Y)
}

// --- Mount helpers ---

func (w *World) MountByID(id string) *Mount {
	for _, m := range w.Mounts {
		if m.ID == id {
			return m
		}
	}
	return nil
}

func (w *World) MountByOwner(npcID string) *Mount {
	for _, m := range w.Mounts {
		if m.OwnerID == npcID && m.Alive {
			return m
		}
	}
	return nil
}

func (w *World) MountsAtLocation(locID string) []*Mount {
	result := make([]*Mount, 0)
	for _, m := range w.Mounts {
		if m.LocationID == locID && m.Alive {
			result = append(result, m)
		}
	}
	return result
}

// --- Carriage helpers ---

func (w *World) CarriageByID(id string) *Carriage {
	for _, c := range w.Carriages {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (w *World) CarriageByOwner(npcID string) *Carriage {
	for _, c := range w.Carriages {
		if c.OwnerID == npcID {
			return c
		}
	}
	return nil
}

// --- Dungeon helpers ---

func (w *World) DungeonByID(id string) *Dungeon {
	for _, d := range w.Dungeons {
		if d.ID == id {
			return d
		}
	}
	return nil
}

func (w *World) DungeonAtLocation(locID string) *Dungeon {
	for _, d := range w.Dungeons {
		if d.LocationID == locID {
			return d
		}
	}
	return nil
}

// --- Mount tick helpers ---

// DecayMounts reduces mount hunger and grooming daily, damages starving mounts.
func (w *World) DecayMounts() {
	for _, m := range w.Mounts {
		if !m.Alive {
			continue
		}
		m.Hunger -= 5
		if m.Hunger < 0 {
			m.Hunger = 0
		}
		m.Grooming -= 3
		if m.Grooming < 0 {
			m.Grooming = 0
		}
		if m.Hunger < 10 {
			m.HP -= 5
			if m.HP <= 0 {
				m.HP = 0
				m.Alive = false
				// Unhitch any carriage
				for _, c := range w.Carriages {
					if c.HorseID == m.ID {
						c.HorseID = ""
					}
				}
			}
		}
	}
}

// TrapsAt returns all traps placed at a given location.
func (w *World) TrapsAt(locID string) []Trap {
	var result []Trap
	for _, t := range w.Traps {
		if t.LocationID == locID {
			result = append(result, t)
		}
	}
	return result
}

// TrapsByOwner returns all traps owned by an NPC.
func (w *World) TrapsByOwner(ownerID string) []Trap {
	var result []Trap
	for _, t := range w.Traps {
		if t.OwnerID == ownerID {
			result = append(result, t)
		}
	}
	return result
}

// RemoveTrap removes a trap owned by ownerID at locationID. Returns true if removed.
func (w *World) RemoveTrap(ownerID, locationID string) bool {
	for i, t := range w.Traps {
		if t.OwnerID == ownerID && t.LocationID == locationID {
			w.Traps = append(w.Traps[:i], w.Traps[i+1:]...)
			return true
		}
	}
	return false
}

// TickTraps processes trap catches. Called during daily tick.
func (w *World) TickTraps() {
	for i := range w.Traps {
		t := &w.Traps[i]
		if t.Caught != "" {
			continue // already caught something
		}
		loc := w.LocationByID(t.LocationID)
		if loc == nil || loc.Type != "forest" {
			continue
		}
		// 40% chance per day to catch game
		if rand.Float64() < 0.40 {
			avail := 99
			if loc.Resources != nil {
				avail = loc.Resources["game"]
			}
			if avail > 0 {
				caught := 1
				if avail >= 2 && rand.Float64() < 0.3 {
					caught = 2
				}
				if loc.Resources != nil {
					loc.Resources["game"] = max(0, loc.Resources["game"]-caught)
				}
				t.Caught = "raw meat"
				t.CaughtQty = caught
			}
		}
	}
}
