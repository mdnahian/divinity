package enemy

import "sync"

type EnemyTemplate struct {
	Category  string
	HP        int
	Strength  int
	Agility   int
	Defense   int
	Locations []string
	Items     []string
	BiomeTag  string // "", "desert", "swamp", "forest", "tundra", "mountain", "plains"
}

var templateMu sync.RWMutex

var Templates = map[string]EnemyTemplate{
	// Base enemies (all biomes)
	"wolf":         {Category: "beast", HP: 40, Strength: 45, Agility: 55, Defense: 15, Locations: []string{"forest"}, Items: []string{"hide", "raw meat"}, BiomeTag: ""},
	"bear":         {Category: "beast", HP: 80, Strength: 70, Agility: 30, Defense: 30, Locations: []string{"forest"}, Items: []string{"hide", "raw meat", "raw meat"}, BiomeTag: ""},
	"wild_boar":    {Category: "beast", HP: 50, Strength: 50, Agility: 40, Defense: 20, Locations: []string{"forest", "farm"}, Items: []string{"raw meat", "hide"}, BiomeTag: ""},
	"giant_spider": {Category: "beast", HP: 35, Strength: 35, Agility: 60, Defense: 10, Locations: []string{"forest", "mine"}, Items: []string{"cloth", "herbs"}, BiomeTag: ""},
	"cave_bat":     {Category: "beast", HP: 20, Strength: 20, Agility: 70, Defense: 5, Locations: []string{"mine"}, Items: []string{"herbs"}, BiomeTag: ""},
	"serpent":      {Category: "beast", HP: 30, Strength: 40, Agility: 65, Defense: 8, Locations: []string{"forest", "dock"}, Items: []string{"herbs", "hide"}, BiomeTag: ""},
	"bandit":       {Category: "human", HP: 60, Strength: 50, Agility: 45, Defense: 20, Locations: []string{"forest", "market"}, Items: []string{"gold", "iron sword", "bread"}, BiomeTag: ""},
	"robber":       {Category: "human", HP: 50, Strength: 45, Agility: 50, Defense: 15, Locations: []string{"market", "inn"}, Items: []string{"gold", "gold", "bread"}, BiomeTag: ""},
	"pirate":       {Category: "human", HP: 65, Strength: 55, Agility: 40, Defense: 25, Locations: []string{"dock"}, Items: []string{"gold", "iron sword", "ale"}, BiomeTag: ""},
	"brigand":      {Category: "human", HP: 70, Strength: 60, Agility: 35, Defense: 25, Locations: []string{"forest", "mine"}, Items: []string{"iron axe", "leather armor", "gold"}, BiomeTag: ""},
	"cultist":      {Category: "human", HP: 45, Strength: 35, Agility: 40, Defense: 15, Locations: []string{"shrine", "forest"}, Items: []string{"herbs", "healing potion", "gold"}, BiomeTag: ""},
	// Desert biome
	"scorpion":       {Category: "beast", HP: 30, Strength: 35, Agility: 60, Defense: 20, Locations: []string{"desert", "cave"}, Items: []string{"herbs"}, BiomeTag: "desert"},
	"sand_wurm":      {Category: "beast", HP: 120, Strength: 80, Agility: 20, Defense: 40, Locations: []string{"desert"}, Items: []string{"hide", "hide", "gems"}, BiomeTag: "desert"},
	"desert_raider":  {Category: "human", HP: 65, Strength: 55, Agility: 50, Defense: 20, Locations: []string{"desert", "market"}, Items: []string{"gold", "iron sword", "bread"}, BiomeTag: "desert"},
	"dust_viper":     {Category: "beast", HP: 25, Strength: 30, Agility: 70, Defense: 5, Locations: []string{"desert", "cave"}, Items: []string{"herbs", "hide"}, BiomeTag: "desert"},
	// Swamp biome
	"swamp_troll":    {Category: "beast", HP: 90, Strength: 65, Agility: 20, Defense: 35, Locations: []string{"swamp", "forest"}, Items: []string{"hide", "herbs", "gold"}, BiomeTag: "swamp"},
	"bog_serpent":    {Category: "beast", HP: 45, Strength: 45, Agility: 55, Defense: 12, Locations: []string{"swamp", "dock"}, Items: []string{"herbs", "hide"}, BiomeTag: "swamp"},
	"marsh_witch":    {Category: "human", HP: 55, Strength: 30, Agility: 45, Defense: 15, Locations: []string{"swamp", "shrine"}, Items: []string{"herbs", "healing potion", "gold"}, BiomeTag: "swamp"},
	"leech_swarm":    {Category: "beast", HP: 20, Strength: 15, Agility: 40, Defense: 5, Locations: []string{"swamp"}, Items: []string{"herbs"}, BiomeTag: "swamp"},
	// Mountain/cave biome
	"cave_troll":     {Category: "beast", HP: 100, Strength: 75, Agility: 15, Defense: 40, Locations: []string{"mine", "cave"}, Items: []string{"stone", "iron ore", "gold"}, BiomeTag: "mountain"},
	"rock_golem":     {Category: "beast", HP: 150, Strength: 85, Agility: 10, Defense: 55, Locations: []string{"mine", "cave"}, Items: []string{"stone", "iron ore", "gems"}, BiomeTag: "mountain"},
	"mountain_lion":  {Category: "beast", HP: 55, Strength: 55, Agility: 65, Defense: 15, Locations: []string{"forest", "cave"}, Items: []string{"hide", "raw meat"}, BiomeTag: "mountain"},
	"stone_drake":    {Category: "beast", HP: 130, Strength: 70, Agility: 35, Defense: 45, Locations: []string{"cave"}, Items: []string{"hide", "gems", "gold"}, BiomeTag: "mountain"},
	// Tundra biome
	"ice_wolf":       {Category: "beast", HP: 50, Strength: 50, Agility: 55, Defense: 20, Locations: []string{"tundra", "forest"}, Items: []string{"fur", "raw meat"}, BiomeTag: "tundra"},
	"frost_giant":    {Category: "beast", HP: 200, Strength: 90, Agility: 15, Defense: 50, Locations: []string{"tundra", "cave"}, Items: []string{"fur", "gold", "gold", "iron axe"}, BiomeTag: "tundra"},
	"snow_bear":      {Category: "beast", HP: 95, Strength: 75, Agility: 25, Defense: 35, Locations: []string{"tundra"}, Items: []string{"fur", "raw meat", "raw meat"}, BiomeTag: "tundra"},
	"ice_wraith":     {Category: "beast", HP: 40, Strength: 45, Agility: 70, Defense: 10, Locations: []string{"tundra", "cave"}, Items: []string{"gems", "herbs"}, BiomeTag: "tundra"},
	// Forest biome (enhanced)
	"dire_wolf":      {Category: "beast", HP: 70, Strength: 60, Agility: 55, Defense: 20, Locations: []string{"forest"}, Items: []string{"hide", "raw meat", "raw meat"}, BiomeTag: "forest"},
	"forest_spirit":  {Category: "beast", HP: 60, Strength: 40, Agility: 50, Defense: 25, Locations: []string{"forest", "shrine"}, Items: []string{"herbs", "herbs", "berries"}, BiomeTag: "forest"},
	"treant":         {Category: "beast", HP: 140, Strength: 70, Agility: 5, Defense: 50, Locations: []string{"forest"}, Items: []string{"logs", "logs", "herbs"}, BiomeTag: "forest"},
	// Plains biome
	"plains_lion":    {Category: "beast", HP: 60, Strength: 55, Agility: 60, Defense: 15, Locations: []string{"farm", "forest"}, Items: []string{"hide", "raw meat"}, BiomeTag: "plains"},
	"mounted_bandit": {Category: "human", HP: 75, Strength: 55, Agility: 55, Defense: 25, Locations: []string{"farm", "market"}, Items: []string{"gold", "gold", "iron sword", "leather armor"}, BiomeTag: "plains"},
	"stampede_bull":  {Category: "beast", HP: 80, Strength: 70, Agility: 40, Defense: 25, Locations: []string{"farm"}, Items: []string{"raw meat", "hide", "hide"}, BiomeTag: "plains"},
}

func RegisterEnemy(name string, tmpl EnemyTemplate) {
	templateMu.Lock()
	defer templateMu.Unlock()
	Templates[name] = tmpl
}

func GetTemplate(name string) (EnemyTemplate, bool) {
	templateMu.RLock()
	defer templateMu.RUnlock()
	t, ok := Templates[name]
	return t, ok
}

func AllTemplateNames() []string {
	templateMu.RLock()
	defer templateMu.RUnlock()
	names := make([]string, 0, len(Templates))
	for k := range Templates {
		names = append(names, k)
	}
	return names
}

// TemplatesForBiome returns enemy template names that match the given biome tag.
// If biome is empty, returns all templates with no biome tag (universal enemies).
func TemplatesForBiome(biome string) []string {
	templateMu.RLock()
	defer templateMu.RUnlock()
	var names []string
	for k, t := range Templates {
		if t.BiomeTag == biome || t.BiomeTag == "" {
			names = append(names, k)
		}
	}
	return names
}
