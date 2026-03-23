package world

type Location struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Type               string         `json:"type"` // forest, farm, mine, dock, well, inn, market, shrine, forge, mill, home, library, school, palace, castle, manor, stable, arena, cave, dungeon_entrance, desert, swamp, tundra
	X                  int            `json:"x"`
	Y                  int            `json:"y"`
	W                  int            `json:"w"`
	H                  int            `json:"h"`
	Color              string         `json:"color"`
	Description        string         `json:"description"`
	OwnerID            string         `json:"ownerId"`
	Resources          map[string]int `json:"resources"`
	MaxResources       map[string]int `json:"maxResources"`
	Capacity           int            `json:"capacity"`
	BuildingID         string         `json:"buildingId"`
	BuildingType       string         `json:"buildingType"`
	BuildingOwnerID    string         `json:"buildingOwnerId"`
	BuildingDurability float64        `json:"buildingDurability"`
	Tier               int            `json:"tier"`
	TerritoryID        string         `json:"territoryId,omitempty"` // which territory this location belongs to
	CityID             string         `json:"cityId,omitempty"`      // which city cluster this location is part of
}

var CapacityMultipliers = map[string]int{
	"home": 2, "inn": 3, "market": 3,
	"forest": 1, "farm": 1,
	"mine": 2, "dock": 2,
	"forge": 2, "workshop": 2, "mill": 2, "library": 2,
	"shrine": 2, "well": 2, "garden": 2,
	// Kingdom-scale locations
	"palace": 8, "castle": 6, "manor": 4,
	"stable": 3, "arena": 6, "barracks": 4,
	"cave": 2, "dungeon_entrance": 1,
	"desert": 1, "swamp": 1, "tundra": 1,
}

func CalcCapacity(locType string, w, h int) int {
	mult, ok := CapacityMultipliers[locType]
	if !ok {
		mult = 2
	}
	cap := w * h * mult
	if locType == "home" {
		cap = 2
	}
	if cap < 1 {
		cap = 1
	}
	return cap
}

// IsNobleRestricted returns true for locations that only nobles and household staff can enter.
func (l *Location) IsNobleRestricted() bool {
	return l.Type == "palace" || l.Type == "castle" || l.Type == "manor"
}

func (l *Location) GetBuildingID() string { return l.ID }
func (l *Location) GetTier() int {
	tiers := map[string]int{"tent": 1, "shack": 2, "cottage": 3, "house": 4}
	if t, ok := tiers[l.BuildingType]; ok {
		return t
	}
	return 0
}

func (w *World) TravelTicks(fromID, toID string, minutesPerTick, minutesPerUnit int) int {
	from := w.LocationByID(fromID)
	to := w.LocationByID(toID)
	if from == nil || to == nil {
		return 0
	}
	dx := (from.X + from.W/2) - (to.X + to.W/2)
	dy := (from.Y + from.H/2) - (to.Y + to.H/2)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	dist := dx + dy

	if minutesPerUnit <= 0 {
		minutesPerUnit = 1
	}
	gameMinutes := dist * minutesPerUnit
	if minutesPerTick <= 0 {
		minutesPerTick = 5
	}
	ticks := gameMinutes / minutesPerTick
	if ticks < 1 {
		ticks = 1
	}
	return ticks
}

// TravelTicksMounted returns the travel time accounting for mount speed, plus
// whether the NPC is mounted and whether they're in a carriage.
func (w *World) TravelTicksMounted(fromID, toID string, minutesPerTick, minutesPerUnit int, npcID string) (ticks int, mounted bool, inCarriage bool) {
	baseTicks := w.TravelTicks(fromID, toID, minutesPerTick, minutesPerUnit)
	mount := w.MountByOwner(npcID)
	if mount == nil || !mount.Alive {
		return baseTicks, false, false
	}
	// Apply mount speed multiplier (horse=3.0x, war_horse=2.5x, pony=2.0x)
	if mount.Speed > 1.0 {
		reducedTicks := int(float64(baseTicks) / mount.Speed)
		if reducedTicks < 1 {
			reducedTicks = 1
		}
		baseTicks = reducedTicks
	}
	carriage := w.CarriageByOwner(npcID)
	inCarriage = carriage != nil && carriage.HorseID == mount.ID
	return baseTicks, true, inCarriage
}

var ResourceDefaults = map[string]map[string]int{
	"forest": {"berries": 20, "herbs": 8, "game": 12, "wood": 25},
	"farm":   {"wheat": 30, "thatch": 20},
	"mine":   {"stone": 20, "iron_ore": 12},
	"dock":   {"fish": 20, "clay": 10},
	"well":   {"water": 150, "clay": 8},
	// Kingdom-scale locations
	"stable":  {"hay": 20, "water": 30},
	"desert":  {"sand": 15, "cactus_water": 5},
	"swamp":   {"peat": 10, "poison_herbs": 5, "bog_iron": 3},
	"cave":    {"stone": 15, "iron_ore": 8, "gems": 3},
	"tundra":  {"ice": 10, "fur": 5, "frozen_herbs": 3},
	"palace":  {"water": 100},
	"castle":  {"water": 80},
	"barracks": {"water": 40},
}

var ResourceRegen = map[string]int{
	"berries": 2, "herbs": 1, "game": 1, "wood": 1,
	"wheat": 2, "thatch": 1,
	"stone": 1, "iron_ore": 1,
	"fish": 2, "clay": 1, "water": 0,
	// Kingdom-scale resources
	"hay": 1, "sand": 1, "cactus_water": 1,
	"peat": 1, "poison_herbs": 1, "bog_iron": 1,
	"gems": 0, "ice": 1, "fur": 1, "frozen_herbs": 1,
}

func InitLocationResources(loc *Location) {
	if defaults, ok := ResourceDefaults[loc.Type]; ok {
		loc.Resources = make(map[string]int)
		loc.MaxResources = make(map[string]int)
		for k, v := range defaults {
			loc.Resources[k] = v
			loc.MaxResources[k] = v
		}
	} else {
		if loc.Resources == nil {
			loc.Resources = make(map[string]int)
		}
		if loc.MaxResources == nil {
			loc.MaxResources = make(map[string]int)
		}
	}
}
