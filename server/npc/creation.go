package npc

import (
	"math"
	"math/rand"
	"strings"

	"github.com/divinity/core/config"
	"github.com/google/uuid"
)

var childNamesM = []string{"Aldric", "Bran", "Cedric", "Dorin", "Edric", "Finn", "Gareth", "Hale", "Ivan", "Jorin"}
var childNamesF = []string{"Anya", "Bessa", "Cora", "Dalia", "Eira", "Fenna", "Gwen", "Hilda", "Iris", "Juna"}

func CreateFromGenesis(data map[string]interface{}, index, worldDay int, cfg *config.Config) *NPC {
	name, _ := data["name"].(string)
	if name == "" {
		name = "Villager"
	}
	profession, _ := data["profession"].(string)
	if profession == "" {
		profession = "farmer"
	}
	age := 25
	if a, ok := data["age"].(float64); ok {
		age = int(a)
	}
	homeID, _ := data["homeId"].(string)
	locID, _ := data["locationId"].(string)

	personality, _ := data["personality"].(string)
	traits := parseTraits(personality)
	statBias := ComputeStatBias(traits)

	var items []InventoryItem
	if rawItems, ok := data["items"].([]interface{}); ok {
		for _, ri := range rawItems {
			iname, _ := ri.(string)
			if iname == "" {
				continue
			}
			qty := randInt(1, 3)
			if iname == "gold" {
				qty = randInt(8, 15)
			}
			items = append(items, InventoryItem{Name: iname, Qty: qty})
		}
	}
	if len(items) == 0 {
		items = []InventoryItem{{Name: "bread", Qty: 2}, {Name: "gold", Qty: 15}}
	}

	lit := randInt(10, 40)
	profDef := GetProfessionDef(profession)
	if profDef.LiteracyBonus > 0 {
		lit = profDef.LiteracyBonus
	}
	primarySkill := profDef.PrimarySkill
	if primarySkill == "" {
		primarySkill = profession
	}

	return NewNPC(Template{
		Name:        name,
		Age:         age,
		Profession:  profession,
		Personality: personality,
		HomeID:      homeID,
		LocationID:  locID,
		StatBias:    statBias,
		StartItems:  items,
		Literacy:    lit,
		Skills:      map[string]float64{primarySkill: float64(randInt(30, 60))},
	}, index, worldDay, cfg)
}

func CreateChild(parentA, parentB *NPC, worldDay int, cfg *config.Config) *NPC {
	idx := int(npcCounter.Add(1))
	isBoy := rand.Float64() < 0.5
	names := childNamesF
	if isBoy {
		names = childNamesM
	}
	name := names[rand.Intn(len(names))]

	stats := BlendStats(parentA.Stats, parentB.Stats)
	isExceptional := rand.Float64() < 0.05
	if isExceptional {
		m := stats.toMap()
		for k, v := range m {
			m[k] = clamp(v+20, 1, 100)
		}
		stats.fromMap(m)
	}

	professions := []string{parentA.Profession, parentB.Profession, "farmer", "hunter", "merchant"}
	profession := professions[rand.Intn(len(professions))]

	homeID := parentA.HomeID
	if rand.Float64() < 0.5 {
		homeID = parentB.HomeID
	}

	baseLifespan := max(50, int(math.Round(
		float64(parentA.BaseLifespan+parentB.BaseLifespan)/2+gauss(0, 5),
	)))
	fertility := clamp(int(math.Round(
		float64(parentA.Fertility+parentB.Fertility)/2+gauss(0, 10),
	)), 1, 100)
	gq := clamp(int(math.Round(
		float64(parentA.GeneticQuality+parentB.GeneticQuality)/2+gauss(0, 8),
	)), 1, 100)
	lit := int(math.Round(float64(parentA.Literacy+parentB.Literacy) * 0.15))

	child := NewNPC(Template{
		ID:             uuid.NewString(),
		Name:           name,
		Profession:     profession,
		HomeID:         homeID,
		BirthDay:       worldDay,
		Stats:          &stats,
		BaseLifespan:   baseLifespan,
		Fertility:      fertility,
		GeneticQuality: gq,
		Literacy:       lit,
		Skills:         map[string]float64{profession: float64(randInt(5, 15))},
		StartItems:     []InventoryItem{},
		Needs:          &Needs{Hunger: 80, Thirst: 80, Fatigue: 10, SocialNeed: 20},
		Happiness:      70,
		Stress:         10,
		ParentID:       parentA.ID,
	}, idx, worldDay, cfg)
	child.Claimed = true
	return child
}

func parseTraits(personality string) []string {
	if personality == "" {
		return nil
	}
	parts := strings.Split(personality, ",")
	traits := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(strings.ToLower(p))
		t = strings.ReplaceAll(t, " ", "_")
		t = strings.ReplaceAll(t, "-", "_")
		if t != "" {
			traits = append(traits, t)
		}
	}
	return traits
}
