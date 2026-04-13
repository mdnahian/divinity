package item

type ItemDef struct {
	Weight         float64
	Stackable      bool
	Slot           string             // "", "weapon", "armor", "bag"
	DurabilityBase int
	DecayRate      float64
	BagSlots       int
	BagWeight      float64
	Category       string             // "food", "drink", "medicine", "material", "currency", "trade_good", "readable", "curiosity"
	Effects        map[string]float64 // e.g. {"hunger_restore": 25, "weapon_bonus": 15, "armor_bonus": 10}
}

var Registry = map[string]ItemDef{
	// Food items
	"bread":       {Weight: 0.3, Stackable: true, DurabilityBase: 60, DecayRate: 2, Category: "food", Effects: map[string]float64{"hunger_restore": 25}},
	"raw meat":    {Weight: 0.5, Stackable: true, DurabilityBase: 40, DecayRate: 4, Category: "food", Effects: map[string]float64{"hunger_restore": 15}},
	"cooked meal": {Weight: 0.4, Stackable: true, DurabilityBase: 50, DecayRate: 3, Category: "food", Effects: map[string]float64{"hunger_restore": 40}},
	"berries":     {Weight: 0.1, Stackable: true, DurabilityBase: 30, DecayRate: 5, Category: "food", Effects: map[string]float64{"hunger_restore": 10}},
	"wheat":       {Weight: 0.3, Stackable: true, DurabilityBase: 80, DecayRate: 1, Category: "food", Effects: map[string]float64{"hunger_restore": 8}},
	"fish":        {Weight: 0.4, Stackable: true, DurabilityBase: 25, DecayRate: 6, Category: "food", Effects: map[string]float64{"hunger_restore": 12}},

	// Drink items
	"ale": {Weight: 0.5, Stackable: true, DurabilityBase: 100, DecayRate: 0, Category: "drink", Effects: map[string]float64{"sobriety": -20, "stress": -10, "happiness": 5, "thirst": 10}},

	// Medicine items
	"herbs":          {Weight: 0.2, Stackable: true, DurabilityBase: 50, DecayRate: 2, Category: "medicine", Effects: map[string]float64{"heal_hp_min": 10, "heal_hp_max": 18}},
	"healing potion": {Weight: 0.3, Stackable: true, DurabilityBase: 100, DecayRate: 0, Category: "medicine", Effects: map[string]float64{"heal_hp_min": 20, "heal_hp_max": 35}},

	// Currency
	"gold": {Weight: 0.01, Stackable: true, DurabilityBase: 999, DecayRate: 0, Category: "currency"},

	// Readable items
	"journal entry":   {Weight: 0.1, Stackable: true, DurabilityBase: 80, DecayRate: 1, Category: "readable"},
	"written journal": {Weight: 0.1, Stackable: false, DurabilityBase: 80, DecayRate: 1, Category: "readable"},
	"book":            {Weight: 0.3, Stackable: true, DurabilityBase: 90, DecayRate: 0.5, Category: "readable"},

	// Curiosity / trade items
	"spice bundle": {Weight: 0.2, Stackable: true, DurabilityBase: 60, DecayRate: 2, Category: "curiosity"},
	"odd trinket":  {Weight: 0.1, Stackable: true, DurabilityBase: 100, DecayRate: 0, Category: "curiosity"},
	"pretty stone": {Weight: 0.2, Stackable: true, DurabilityBase: 999, DecayRate: 0, Category: "curiosity"},
	"old coin":     {Weight: 0.05, Stackable: true, DurabilityBase: 999, DecayRate: 0, Category: "curiosity"},
	"carved bone":  {Weight: 0.1, Stackable: true, DurabilityBase: 100, DecayRate: 0, Category: "curiosity"},

	// Weapons
	"iron sword":    {Weight: 3.0, Stackable: false, Slot: "weapon", DurabilityBase: 100, DecayRate: 0.5, Effects: map[string]float64{"weapon_bonus": 15}},
	"iron axe":      {Weight: 2.5, Stackable: false, Slot: "weapon", DurabilityBase: 100, DecayRate: 0.5, Effects: map[string]float64{"weapon_bonus": 12}},
	"pickaxe":       {Weight: 2.5, Stackable: false, Slot: "weapon", DurabilityBase: 100, DecayRate: 0.5, Effects: map[string]float64{"weapon_bonus": 5}},
	"hammer":        {Weight: 2.0, Stackable: false, Slot: "weapon", DurabilityBase: 100, DecayRate: 0.5, Effects: map[string]float64{"weapon_bonus": 8}},
	"wooden club":   {Weight: 1.5, Stackable: false, Slot: "weapon", DurabilityBase: 60, DecayRate: 1, Effects: map[string]float64{"weapon_bonus": 5}},
	"walking stick": {Weight: 1.0, Stackable: false, Slot: "weapon", DurabilityBase: 40, DecayRate: 1, Effects: map[string]float64{"weapon_bonus": 2}},

	// Armor
	"leather armor": {Weight: 4.0, Stackable: false, Slot: "armor", DurabilityBase: 80, DecayRate: 0.5, Effects: map[string]float64{"armor_bonus": 10}},
	"cloth tunic":   {Weight: 1.0, Stackable: false, Slot: "armor", DurabilityBase: 50, DecayRate: 1, Effects: map[string]float64{"armor_bonus": 3}},

	// Bags
	"belt pouch": {Weight: 0.3, Stackable: false, Slot: "bag", DurabilityBase: 70, DecayRate: 0.5, BagSlots: 3, BagWeight: 2},
	"satchel":    {Weight: 0.5, Stackable: false, Slot: "bag", DurabilityBase: 70, DecayRate: 0.5, BagSlots: 6, BagWeight: 5},
	"backpack":   {Weight: 1.0, Stackable: false, Slot: "bag", DurabilityBase: 80, DecayRate: 0.5, BagSlots: 12, BagWeight: 12},

	// Raw materials
	"hide":       {Weight: 1.0, Stackable: true, DurabilityBase: 70, DecayRate: 1, Category: "material"},
	"logs":       {Weight: 2.0, Stackable: true, DurabilityBase: 100, DecayRate: 0.5, Category: "material"},
	"stone":      {Weight: 3.0, Stackable: true, DurabilityBase: 999, DecayRate: 0, Category: "material"},
	"iron ore":   {Weight: 2.0, Stackable: true, DurabilityBase: 999, DecayRate: 0, Category: "material"},
	"iron ingot": {Weight: 1.5, Stackable: true, DurabilityBase: 999, DecayRate: 0, Category: "material"},
	"leather":    {Weight: 0.8, Stackable: true, DurabilityBase: 80, DecayRate: 0.5, Category: "material"},
	"thatch":     {Weight: 0.5, Stackable: true, DurabilityBase: 60, DecayRate: 1, Category: "material"},
	"clay":       {Weight: 1.0, Stackable: true, DurabilityBase: 999, DecayRate: 0, Category: "material"},
	"flour":      {Weight: 0.3, Stackable: true, DurabilityBase: 70, DecayRate: 1, Category: "material"},
	"cloth":      {Weight: 0.3, Stackable: true, DurabilityBase: 60, DecayRate: 1, Category: "material"},
	"rope":       {Weight: 0.4, Stackable: true, DurabilityBase: 70, DecayRate: 0.5, Category: "material"},

	// Trade goods
	"ceramic": {Weight: 0.5, Stackable: true, DurabilityBase: 60, DecayRate: 0, Category: "trade_good"},

	// Tools (craftable without forge)
	"snare":   {Weight: 0.3, Stackable: true, DurabilityBase: 40, DecayRate: 2, Category: "tool"},
	"canteen": {Weight: 0.4, Stackable: false, DurabilityBase: 80, DecayRate: 0.5, Category: "tool", Effects: map[string]float64{"water_capacity": 3}},
	"filled canteen": {Weight: 0.8, Stackable: false, DurabilityBase: 80, DecayRate: 0.5, Category: "tool", Effects: map[string]float64{"thirst_restore": 30, "water_charges": 3}},

	// Shelter items
	"lean-to frame": {Weight: 0, Stackable: false, DurabilityBase: 30, DecayRate: 3, Category: "structure"},
}

func GetInfo(name string) ItemDef {
	if def, ok := Registry[name]; ok {
		return def
	}
	return ItemDef{Weight: 0.5, Stackable: true, DurabilityBase: 50, DecayRate: 1}
}

func GetWeight(name string, qty int) float64 {
	return GetInfo(name).Weight * float64(qty)
}

func ItemsByCategory(cat string) []string {
	var names []string
	for name, def := range Registry {
		if def.Category == cat {
			names = append(names, name)
		}
	}
	return names
}
