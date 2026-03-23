package npc

var TraitToBias = map[string]map[string]int{
	"brave":         {"courage": 20},
	"curious":       {"curiosity": 20},
	"stubborn":      {"conscientiousness": 15, "conformity": -15},
	"greedy":        {"greed": 20},
	"kind":          {"empathy": 20, "virtue": 15},
	"aggressive":    {"aggression": 20},
	"wise":          {"wisdom": 20},
	"charming":      {"charisma": 20},
	"spiritual":     {"spiritual_sensitivity": 40, "prayer_frequency": 30},
	"creative":      {"creativity": 20},
	"ambitious":     {"ambition": 20},
	"loyal":         {"loyalty": 20},
	"intelligent":   {"intelligence": 20},
	"clever":        {"intelligence": 15, "agility": 10},
	"strong":        {"strength": 20},
	"cunning":       {"intelligence": 15, "agility": 10, "virtue": -10},
	"gentle":        {"empathy": 15, "aggression": -15},
	"proud":         {"dominance": 15, "ambition": 10},
	"stoic":         {"resilience": 20, "neuroticism": -15},
	"timid":         {"courage": -20},
	"outgoing":      {"extraversion": 20},
	"introverted":   {"extraversion": -20},
	"generous":      {"generosity": 20},
	"pious":         {"spiritual_sensitivity": 45, "prayer_frequency": 35},
	"scholarly":     {"intelligence": 15, "curiosity": 15},
	"ruthless":      {"aggression": 15, "empathy": -15},
	"nurturing":     {"empathy": 15, "teaching_drive": 15},
	"quiet":         {"extraversion": -15},
	"hot_tempered":  {"aggression": 15, "neuroticism": 10},
	"patient":       {"resilience": 15, "conscientiousness": 10},
	"devout":        {"spiritual_sensitivity": 50, "faith_capacity": 20, "prayer_frequency": 40},
	"shrewd":        {"intelligence": 10, "greed": 10},
	"humble":        {"dominance": -15, "virtue": 10},
	"stern":         {"dominance": 15, "empathy": -10},
	"cheerful":      {"extraversion": 10, "empathy": 10},
	"suspicious":    {"trust_threshold": 20, "jealousy": 10},
	"reckless":      {"courage": 15, "conscientiousness": -15},
	"diligent":      {"conscientiousness": 20},
	"lazy":          {"conscientiousness": -20},
	"witty":         {"intelligence": 10, "charisma": 10},
	"fearful":       {"courage": -15, "neuroticism": 15},
	"compassionate": {"empathy": 20, "virtue": 10},
	"selfish":       {"greed": 15, "empathy": -15},
}

func ComputeStatBias(traits []string) map[string]int {
	bias := make(map[string]int)
	for _, trait := range traits {
		if m, ok := TraitToBias[trait]; ok {
			for k, v := range m {
				bias[k] += v
			}
		}
	}
	return bias
}
