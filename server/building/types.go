package building

type BuildingTypeDef struct {
	MinCarpentry int
	Materials    map[string]int
	Tier         int
	Function     string // shelter, market, workshop, forge, shrine, school, library
}

var Types = map[string]BuildingTypeDef{
	"tent":         {MinCarpentry: 0, Materials: map[string]int{"cloth": 2, "rope": 1}, Tier: 1, Function: "shelter"},
	"shack":        {MinCarpentry: 20, Materials: map[string]int{"logs": 3, "thatch": 3}, Tier: 2, Function: "shelter"},
	"cottage":      {MinCarpentry: 40, Materials: map[string]int{"logs": 5, "stone": 3}, Tier: 3, Function: "shelter"},
	"house":        {MinCarpentry: 55, Materials: map[string]int{"stone": 5, "logs": 4}, Tier: 4, Function: "shelter"},
	"market_stall": {MinCarpentry: 25, Materials: map[string]int{"logs": 3}, Tier: 1, Function: "market"},
	"workshop":     {MinCarpentry: 45, Materials: map[string]int{"logs": 4, "stone": 2}, Tier: 2, Function: "workshop"},
	"forge":        {MinCarpentry: 50, Materials: map[string]int{"stone": 5, "iron ingot": 2}, Tier: 3, Function: "forge"},
	"shrine":       {MinCarpentry: 20, Materials: map[string]int{"logs": 3, "stone": 2}, Tier: 1, Function: "shrine"},
	"school":       {MinCarpentry: 55, Materials: map[string]int{"logs": 5, "stone": 3}, Tier: 2, Function: "school"},
	"library":      {MinCarpentry: 60, Materials: map[string]int{"stone": 5, "logs": 3}, Tier: 3, Function: "library"},
}

type Construction struct {
	ID             string  `json:"id"`
	BuildingType   string  `json:"buildingType"`
	Name           string  `json:"name"`
	OwnerID        string  `json:"ownerId"`
	CommissionerID string  `json:"commissionerId"`
	Progress       float64 `json:"progress"`
	MaxProgress    float64 `json:"maxProgress"`
	Durability     float64 `json:"durability"`
	LocationID     string  `json:"locationId"`
}
