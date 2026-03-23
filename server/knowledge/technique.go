package knowledge

import (
	"fmt"
	"math/rand"
	"sync/atomic"
)

var techCounter atomic.Int64

type BonusType struct {
	Label      string
	Multiplier float64
}

var BonusTypes = map[string]BonusType{
	"weapon_durability": {Label: "+15% weapon durability", Multiplier: 0.15},
	"crop_yield":        {Label: "+20% crop yield", Multiplier: 0.20},
	"craft_quality":     {Label: "+10% craft quality", Multiplier: 0.10},
	"healing_potency":   {Label: "+15% healing potency", Multiplier: 0.15},
	"trade_profit":      {Label: "+10% trade profit", Multiplier: 0.10},
	"build_speed":       {Label: "+15% build speed", Multiplier: 0.15},
	"fish_catch":        {Label: "+20% fish catch rate", Multiplier: 0.20},
	"smelt_efficiency":  {Label: "+10% smelt efficiency", Multiplier: 0.10},
}

// GetTechniqueBonus returns the cumulative multiplier for a given bonusType
// that applies to a specific NPC (by ID) from techniques they know.
func GetTechniqueBonus(npcID string, bonusType string, techniques []*Technique) float64 {
	var total float64
	for _, t := range techniques {
		if t.BonusType != bonusType {
			continue
		}
		for _, knower := range t.KnownBy {
			if knower == npcID {
				if bt, ok := BonusTypes[bonusType]; ok {
					total += bt.Multiplier
				}
				break
			}
		}
	}
	return total
}

type WrittenRecord struct {
	AuthorID    string `json:"authorId"`
	Day         int    `json:"day"`
	AuthorAlive bool   `json:"authorAlive"`
}

type Technique struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	BonusType     string          `json:"bonusType"`
	BonusLabel    string          `json:"bonusLabel"`
	KnownBy       []string        `json:"knownBy"`
	WrittenIn     []WrittenRecord `json:"writtenIn"`
	DiscoveredDay int             `json:"discoveredDay"`
}

var discoveryPrefixes = []string{"Improved", "Advanced", "Master", "Refined", "Ancient", "Clever"}

func TryDiscoverTechnique(npcID, npcName string, skillName string, skillLevel, creativity int, existingNames []string, worldDay int) *Technique {
	if skillLevel < 70 || creativity < 60 {
		return nil
	}
	if rand.Float64() > 0.02 {
		return nil
	}

	prefix := discoveryPrefixes[rand.Intn(len(discoveryPrefixes))]
	name := fmt.Sprintf("%s %s", prefix, skillName)
	for _, existing := range existingNames {
		if existing == name {
			return nil
		}
	}

	bonusKeys := make([]string, 0, len(BonusTypes))
	for k := range BonusTypes {
		bonusKeys = append(bonusKeys, k)
	}
	bonusType := bonusKeys[rand.Intn(len(bonusKeys))]

	return CreateTechnique(
		name,
		fmt.Sprintf("A technique for %s discovered by %s.", skillName, npcName),
		bonusType,
		npcID,
		worldDay,
	)
}

func CreateTechnique(name, description, bonusType, discoveredBy string, worldDay int) *Technique {
	label := bonusType
	if bt, ok := BonusTypes[bonusType]; ok {
		label = bt.Label
	}
	knownBy := []string{}
	if discoveredBy != "" {
		knownBy = []string{discoveredBy}
	}
	return &Technique{
		ID:            fmt.Sprintf("tech_%d", techCounter.Add(1)),
		Name:          name,
		Description:   description,
		BonusType:     bonusType,
		BonusLabel:    label,
		KnownBy:       knownBy,
		WrittenIn:     []WrittenRecord{},
		DiscoveredDay: worldDay,
	}
}
