package enemy

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync/atomic"

	"github.com/divinity/core/item"
)

var enemyCounter atomic.Int64

type InventoryItem struct {
	Name string `json:"name"`
	Qty  int    `json:"qty"`
}

type Enemy struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Category   string          `json:"category"`
	LocationID string          `json:"locationId"`
	Alive      bool            `json:"alive"`
	HP         int             `json:"hp"`
	MaxHP      int             `json:"maxHp"`
	Strength   int             `json:"strength"`
	Agility    int             `json:"agility"`
	Defense    int             `json:"defense"`
	Inventory  []InventoryItem `json:"inventory"`
}

func Create(enemyType, locationID string) *Enemy {
	tmpl, ok := GetTemplate(enemyType)
	if !ok {
		tmpl, _ = GetTemplate("wolf")
		enemyType = "wolf"
	}
	id := fmt.Sprintf("enemy_%d", enemyCounter.Add(1))
	hp := tmpl.HP + rand.Intn(16) - 5

	items := make([]InventoryItem, 0, len(tmpl.Items))
	for _, name := range tmpl.Items {
		qty := 1
		if name == "gold" {
			qty = rand.Intn(5) + 2
		}
		items = append(items, InventoryItem{Name: name, Qty: qty})
	}

	return &Enemy{
		ID:         id,
		Type:       enemyType,
		Name:       strings.ReplaceAll(enemyType, "_", " "),
		Category:   tmpl.Category,
		LocationID: locationID,
		Alive:      true,
		HP:         hp,
		MaxHP:      hp,
		Strength:   tmpl.Strength + rand.Intn(11) - 5,
		Agility:    tmpl.Agility + rand.Intn(11) - 5,
		Defense:    tmpl.Defense + rand.Intn(7) - 3,
		Inventory:  items,
	}
}

func CreateScaled(enemyType, locationID string, avgNPCStrength int) *Enemy {
	tmpl, ok := GetTemplate(enemyType)
	if !ok {
		tmpl, _ = GetTemplate("wolf")
		enemyType = "wolf"
	}

	scale := math.Max(0.5, math.Min(1.5, float64(avgNPCStrength)/50.0))

	id := fmt.Sprintf("enemy_%d", enemyCounter.Add(1))
	hp := int(float64(tmpl.HP)*scale) + rand.Intn(16) - 5

	items := make([]InventoryItem, 0, len(tmpl.Items))
	for _, name := range tmpl.Items {
		qty := 1
		if name == "gold" {
			qty = rand.Intn(5) + 2
		}
		items = append(items, InventoryItem{Name: name, Qty: qty})
	}

	return &Enemy{
		ID:         id,
		Type:       enemyType,
		Name:       strings.ReplaceAll(enemyType, "_", " "),
		Category:   tmpl.Category,
		LocationID: locationID,
		Alive:      true,
		HP:         hp,
		MaxHP:      hp,
		Strength:   int(float64(tmpl.Strength)*scale) + rand.Intn(11) - 5,
		Agility:    int(float64(tmpl.Agility)*scale) + rand.Intn(11) - 5,
		Defense:    int(float64(tmpl.Defense)*scale) + rand.Intn(7) - 3,
		Inventory:  items,
	}
}

func (e *Enemy) WeaponBonus() int {
	for _, it := range e.Inventory {
		def := item.GetInfo(it.Name)
		if b := def.Effects["weapon_bonus"]; b > 0 {
			return int(b)
		}
	}
	return 0
}

func (e *Enemy) ArmorBonus() int {
	for _, it := range e.Inventory {
		def := item.GetInfo(it.Name)
		if b := def.Effects["armor_bonus"]; b > 0 {
			return int(b)
		}
	}
	return 0
}

func ValidTypesForLocation(locationType string) []string {
	templateMu.RLock()
	defer templateMu.RUnlock()
	var types []string
	for name, tmpl := range Templates {
		for _, loc := range tmpl.Locations {
			if loc == locationType {
				types = append(types, name)
				break
			}
		}
	}
	return types
}
