package world

// Dungeon represents a multi-floor explorable structure with enemies and loot.
type Dungeon struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	LocationID string         `json:"locationId"` // entrance location on the world map
	BiomeType  string         `json:"biomeType"`  // biome context for enemy/loot generation
	Floors     []DungeonFloor `json:"floors"`
	Difficulty int            `json:"difficulty"` // 1-10
	Discovered bool           `json:"discovered"`
	CreatedDay int            `json:"createdDay"`
}

// DungeonFloor is a single level within a dungeon.
type DungeonFloor struct {
	Level    int           `json:"level"`
	EnemyIDs []string      `json:"enemyIds"`
	Loot     []DungeonLoot `json:"loot"`
	Cleared  bool          `json:"cleared"`
}

// DungeonLoot defines a possible item drop on a dungeon floor.
type DungeonLoot struct {
	ItemName   string  `json:"itemName"`
	DropChance float64 `json:"dropChance"` // 0.0-1.0
	MinQty     int     `json:"minQty"`
	MaxQty     int     `json:"maxQty"`
}
