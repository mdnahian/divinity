package world

// Mount represents a rideable animal (horse, war horse, pony).
type Mount struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"` // "horse", "war_horse", "pony"
	OwnerID     string  `json:"ownerId"`
	LocationID  string  `json:"locationId"` // stable where it's kept
	HP          int     `json:"hp"`
	MaxHP       int     `json:"maxHp"`
	Speed       float64 `json:"speed"`       // travel multiplier: horse=3.0, war_horse=2.5, pony=2.0
	CarryWeight float64 `json:"carryWeight"` // extra weight capacity
	Hunger      float64 `json:"hunger"`      // 0-100, needs feeding
	Grooming    float64 `json:"grooming"`    // 0-100, affects HP decay
	Price       int     `json:"price"`
	Alive       bool    `json:"alive"`
}

// MountTemplates defines base stats for each mount type.
var MountTemplates = map[string]Mount{
	"horse": {
		Type: "horse", MaxHP: 100, HP: 100,
		Speed: 3.0, CarryWeight: 30, Price: 500,
		Hunger: 100, Grooming: 100, Alive: true,
	},
	"war_horse": {
		Type: "war_horse", MaxHP: 150, HP: 150,
		Speed: 2.5, CarryWeight: 40, Price: 800,
		Hunger: 100, Grooming: 100, Alive: true,
	},
	"pony": {
		Type: "pony", MaxHP: 70, HP: 70,
		Speed: 2.0, CarryWeight: 15, Price: 250,
		Hunger: 100, Grooming: 100, Alive: true,
	},
}

// Carriage is a horse-drawn vehicle for transporting large amounts of cargo.
type Carriage struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	OwnerID     string  `json:"ownerId"`
	HorseID     string  `json:"horseId"`    // must be hitched to a mount
	LocationID  string  `json:"locationId"` // where it's parked
	CargoSlots  int     `json:"cargoSlots"` // number of item slots (e.g., 30)
	CargoWeight float64 `json:"cargoWeight"` // max weight capacity (e.g., 200)
	Durability  float64 `json:"durability"`  // 0-100
	Price       int     `json:"price"`
}

// CarriageTemplates defines base stats for carriage types.
var CarriageTemplates = map[string]Carriage{
	"cart": {
		Name: "Cart", CargoSlots: 15, CargoWeight: 100,
		Durability: 100, Price: 600,
	},
	"wagon": {
		Name: "Wagon", CargoSlots: 30, CargoWeight: 200,
		Durability: 100, Price: 1200,
	},
}
