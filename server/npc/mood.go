package npc

type moodEntry struct {
	Name string
	Test func(*NPC) bool
}

var moodTable = []moodEntry{
	{"desperate", func(n *NPC) bool { return n.Needs.Hunger <= 5 || n.Needs.Thirst <= 5 }},
	{"miserable", func(n *NPC) bool { return n.Happiness < 15 && n.Stress > 70 }},
	{"furious", func(n *NPC) bool { return n.Stress > 80 && n.Stats.Aggression > 55 }},
	{"anxious", func(n *NPC) bool { return n.Stress > 60 }},
	{"exhausted", func(n *NPC) bool { return n.Needs.Fatigue > 80 }},
	{"starving", func(n *NPC) bool { return n.Needs.Hunger < 20 }},
	{"parched", func(n *NPC) bool { return n.Needs.Thirst < 20 }},
	{"lonely", func(n *NPC) bool { return n.Needs.SocialNeed > 75 }},
	{"depressed", func(n *NPC) bool { return n.Stats.Depression > 65 }},
	{"melancholy", func(n *NPC) bool { return n.Happiness < 30 }},
	{"restless", func(n *NPC) bool { return n.Stress > 45 && n.Needs.Fatigue < 30 }},
	{"content", func(n *NPC) bool { return n.Happiness > 65 && n.Stress < 30 }},
	{"cheerful", func(n *NPC) bool { return n.Happiness > 75 }},
	{"calm", func(n *NPC) bool { return n.Stress < 20 }},
	{"neutral", func(n *NPC) bool { return true }},
}

func (n *NPC) Mood() string {
	for _, entry := range moodTable {
		if entry.Test(n) {
			return entry.Name
		}
	}
	return "neutral"
}

func (n *NPC) PersonalitySummary() string {
	s := n.Stats
	var traits []string

	if s.Greed > 65 {
		traits = append(traits, "greedy")
	}
	if s.Virtue > 65 {
		traits = append(traits, "virtuous")
	}
	if s.Empathy > 65 {
		traits = append(traits, "empathetic")
	}
	if s.Aggression > 65 {
		traits = append(traits, "aggressive")
	}
	if s.Curiosity > 65 {
		traits = append(traits, "curious")
	}
	if s.Ambition > 65 {
		traits = append(traits, "ambitious")
	}
	if s.Charisma > 65 {
		traits = append(traits, "charming")
	}
	if s.Extraversion > 65 {
		traits = append(traits, "outgoing")
	}
	if s.Courage > 65 {
		traits = append(traits, "brave")
	}
	if s.Conscientiousness > 65 {
		traits = append(traits, "diligent")
	}
	if s.Creativity > 65 {
		traits = append(traits, "creative")
	}
	if s.SpiritualSensitivity > 65 {
		traits = append(traits, "spiritual")
	}
	if s.Intelligence > 70 {
		traits = append(traits, "clever")
	}
	if s.Wisdom > 70 {
		traits = append(traits, "wise")
	}
	if s.Dominance > 65 {
		traits = append(traits, "dominant")
	}
	if s.TeachingDrive > 65 {
		traits = append(traits, "mentor")
	}
	if s.Loyalty > 70 {
		traits = append(traits, "loyal")
	}
	if s.Neuroticism > 65 {
		traits = append(traits, "neurotic")
	}
	if s.Resilience > 65 {
		traits = append(traits, "resilient")
	}
	if s.Generosity > 65 {
		traits = append(traits, "generous")
	}
	if s.Openness > 65 {
		traits = append(traits, "open-minded")
	}
	if s.Jealousy > 65 {
		traits = append(traits, "jealous")
	}
	if s.Conformity > 70 {
		traits = append(traits, "conformist")
	}
	if s.Greed < 30 && s.Generosity < 30 {
		traits = append(traits, "selfless")
	}
	if s.Virtue < 30 {
		traits = append(traits, "unscrupulous")
	}
	if s.Empathy < 30 {
		traits = append(traits, "cold")
	}
	if s.Courage < 30 {
		traits = append(traits, "timid")
	}
	if s.Extraversion < 30 {
		traits = append(traits, "introverted")
	}
	if s.Conformity < 30 {
		traits = append(traits, "rebellious")
	}
	if s.Sanity < 40 {
		traits = append(traits, "unstable")
	}
	if s.AddictionSusceptibility > 70 {
		traits = append(traits, "addictive personality")
	}

	if len(traits) == 0 {
		return "unremarkable"
	}
	if len(traits) > 6 {
		traits = traits[:6]
	}
	result := traits[0]
	for _, t := range traits[1:] {
		result += ", " + t
	}
	return result
}
