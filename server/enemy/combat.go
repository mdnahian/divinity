package enemy

import (
	"math"
	"math/rand"
)

type Combatant interface {
	GetStrength() int
	GetAgility() int
	GetWeaponBonus() int
	GetArmorBonus() int
	GetDefense() int
}

func (e *Enemy) GetStrength() int   { return e.Strength }
func (e *Enemy) GetAgility() int    { return e.Agility }
func (e *Enemy) GetWeaponBonus() int { return e.WeaponBonus() }
func (e *Enemy) GetArmorBonus() int  { return e.ArmorBonus() }
func (e *Enemy) GetDefense() int     { return e.Defense }

func CalculateCombatDamage(atkStr, atkAgi, atkWeapon, defStr, defDef int) int {
	atkPower := float64(atkStr) + float64(atkAgi)*0.4 + float64(atkWeapon)
	defPower := float64(defStr)*0.3 + float64(defDef)
	roll := 0.5 + rand.Float64()
	rawDmg := math.Max(1, math.Round((atkPower*roll-defPower)*0.25))
	return int(math.Min(rawDmg, 40))
}

type DroppedItem struct {
	Name       string
	Qty        int
	LocationID string
}

func DropLoot(e *Enemy) []DroppedItem {
	if len(e.Inventory) == 0 {
		return nil
	}
	count := rand.Intn(3) + 1
	if count > len(e.Inventory) {
		count = len(e.Inventory)
	}
	shuffled := make([]InventoryItem, len(e.Inventory))
	copy(shuffled, e.Inventory)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	dropped := make([]DroppedItem, 0, count)
	for i := 0; i < count; i++ {
		dropped = append(dropped, DroppedItem{
			Name:       shuffled[i].Name,
			Qty:        shuffled[i].Qty,
			LocationID: e.LocationID,
		})
	}
	return dropped
}
