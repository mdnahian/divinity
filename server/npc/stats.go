package npc

import "math"

func generateStats(bias map[string]int) Stats {
	base := func(key string) int {
		b := 0
		if bias != nil {
			b = bias[key]
		}
		return clamp(int(math.Round(gauss(50, 15)))+b, 1, 100)
	}
	lowBase := func(key string) int {
		b := 0
		if bias != nil {
			b = bias[key]
		}
		return clamp(int(math.Round(gauss(25, 12)))+b, 1, 100)
	}
	return Stats{
		Strength:                base("strength"),
		Agility:                 base("agility"),
		Health:                  base("health"),
		Intelligence:            base("intelligence"),
		Wisdom:                  base("wisdom"),
		Curiosity:               base("curiosity"),
		Creativity:              base("creativity"),
		Charisma:                base("charisma"),
		Empathy:                 base("empathy"),
		Extraversion:            base("extraversion"),
		Reputation:              clamp(int(math.Round(gauss(50, 10))), 1, 100),
		Aggression:              base("aggression"),
		Virtue:                  base("virtue"),
		Greed:                   base("greed"),
		Ambition:                base("ambition"),
		Conscientiousness:       base("conscientiousness"),
		Courage:                 base("courage"),
		SpiritualSensitivity:    lowBase("spiritual_sensitivity"),
		PrayerFrequency:         lowBase("prayer_frequency"),
		Endurance:               base("endurance"),
		Dexterity:               base("dexterity"),
		PainTolerance:           base("pain_tolerance"),
		DiseaseResistance:       base("disease_resistance"),
		Sanity:                  clamp(int(math.Round(gauss(75, 10))), 1, 100),
		Resilience:              base("resilience"),
		Openness:                base("openness"),
		Neuroticism:             base("neuroticism"),
		Depression:              clamp(int(math.Round(gauss(20, 15))), 0, 100),
		Trauma:                  clamp(int(math.Round(gauss(5, 10))), 0, 100),
		Dominance:               base("dominance"),
		StrategicThinking:       base("strategic_thinking"),
		Decisiveness:            base("decisiveness"),
		Persuasion:              base("persuasion"),
		TeachingDrive:           base("teaching_drive"),
		Loyalty:                 base("loyalty"),
		Jealousy:                base("jealousy"),
		TrustThreshold:          base("trust_threshold"),
		Generosity:              base("generosity"),
		Conformity:              base("conformity"),
		Infamy:                  clamp(int(math.Round(gauss(10, 10))), 0, 100),
		FaithCapacity:           base("faith_capacity"),
		DoubtResistance:         base("doubt_resistance"),
		DivineLuck:              base("divine_luck"),
		Attractiveness:          base("attractiveness"),
		MemoryCapacity:          base("memory_capacity"),
		SocialAwareness:         base("social_awareness"),
		Composure:               base("composure"),
		Accountability:          base("accountability"),
		PoliticalInstinct:       base("political_instinct"),
		AddictionSusceptibility: base("addiction_susceptibility"),
		MentalFragility:         base("mental_fragility"),
		HungerRate:              clamp(int(math.Round(gauss(50, 12))), 20, 80),
		SleepNeed:               clamp(int(math.Round(gauss(50, 12))), 20, 80),
		Carpentry:               base("carpentry"),
		VisionStat:              base("vision_stat"),
		MysticalAptitude:        base("mystical_aptitude"),
		MiracleReceptivity:      base("miracle_receptivity"),
		CurseVulnerability:      base("curse_vulnerability"),
	}
}

func (s *Stats) toMap() map[string]int {
	return map[string]int{
		"strength": s.Strength, "agility": s.Agility, "health": s.Health,
		"intelligence": s.Intelligence, "wisdom": s.Wisdom, "curiosity": s.Curiosity,
		"creativity": s.Creativity, "charisma": s.Charisma, "empathy": s.Empathy,
		"extraversion": s.Extraversion, "reputation": s.Reputation, "aggression": s.Aggression,
		"virtue": s.Virtue, "greed": s.Greed, "ambition": s.Ambition,
		"conscientiousness": s.Conscientiousness, "courage": s.Courage,
		"spiritual_sensitivity": s.SpiritualSensitivity, "prayer_frequency": s.PrayerFrequency,
		"endurance": s.Endurance, "dexterity": s.Dexterity, "pain_tolerance": s.PainTolerance,
		"disease_resistance": s.DiseaseResistance, "sanity": s.Sanity, "resilience": s.Resilience,
		"openness": s.Openness, "neuroticism": s.Neuroticism, "depression": s.Depression,
		"trauma": s.Trauma, "dominance": s.Dominance, "strategic_thinking": s.StrategicThinking,
		"decisiveness": s.Decisiveness, "persuasion": s.Persuasion, "teaching_drive": s.TeachingDrive,
		"loyalty": s.Loyalty, "jealousy": s.Jealousy, "trust_threshold": s.TrustThreshold,
		"generosity": s.Generosity, "conformity": s.Conformity, "infamy": s.Infamy,
		"faith_capacity": s.FaithCapacity, "doubt_resistance": s.DoubtResistance,
		"divine_luck": s.DivineLuck, "attractiveness": s.Attractiveness,
		"memory_capacity": s.MemoryCapacity, "social_awareness": s.SocialAwareness,
		"composure": s.Composure, "accountability": s.Accountability,
		"political_instinct": s.PoliticalInstinct, "addiction_susceptibility": s.AddictionSusceptibility,
		"mental_fragility": s.MentalFragility, "hunger_rate": s.HungerRate,
		"sleep_need": s.SleepNeed, "carpentry": s.Carpentry, "vision_stat": s.VisionStat,
		"mystical_aptitude": s.MysticalAptitude, "miracle_receptivity": s.MiracleReceptivity,
		"curse_vulnerability": s.CurseVulnerability,
	}
}

func (s *Stats) fromMap(m map[string]int) {
	s.Strength = m["strength"]; s.Agility = m["agility"]; s.Health = m["health"]
	s.Intelligence = m["intelligence"]; s.Wisdom = m["wisdom"]; s.Curiosity = m["curiosity"]
	s.Creativity = m["creativity"]; s.Charisma = m["charisma"]; s.Empathy = m["empathy"]
	s.Extraversion = m["extraversion"]; s.Reputation = m["reputation"]; s.Aggression = m["aggression"]
	s.Virtue = m["virtue"]; s.Greed = m["greed"]; s.Ambition = m["ambition"]
	s.Conscientiousness = m["conscientiousness"]; s.Courage = m["courage"]
	s.SpiritualSensitivity = m["spiritual_sensitivity"]; s.PrayerFrequency = m["prayer_frequency"]
	s.Endurance = m["endurance"]; s.Dexterity = m["dexterity"]; s.PainTolerance = m["pain_tolerance"]
	s.DiseaseResistance = m["disease_resistance"]; s.Sanity = m["sanity"]; s.Resilience = m["resilience"]
	s.Openness = m["openness"]; s.Neuroticism = m["neuroticism"]; s.Depression = m["depression"]
	s.Trauma = m["trauma"]; s.Dominance = m["dominance"]; s.StrategicThinking = m["strategic_thinking"]
	s.Decisiveness = m["decisiveness"]; s.Persuasion = m["persuasion"]; s.TeachingDrive = m["teaching_drive"]
	s.Loyalty = m["loyalty"]; s.Jealousy = m["jealousy"]; s.TrustThreshold = m["trust_threshold"]
	s.Generosity = m["generosity"]; s.Conformity = m["conformity"]; s.Infamy = m["infamy"]
	s.FaithCapacity = m["faith_capacity"]; s.DoubtResistance = m["doubt_resistance"]
	s.DivineLuck = m["divine_luck"]; s.Attractiveness = m["attractiveness"]
	s.MemoryCapacity = m["memory_capacity"]; s.SocialAwareness = m["social_awareness"]
	s.Composure = m["composure"]; s.Accountability = m["accountability"]
	s.PoliticalInstinct = m["political_instinct"]; s.AddictionSusceptibility = m["addiction_susceptibility"]
	s.MentalFragility = m["mental_fragility"]; s.HungerRate = m["hunger_rate"]
	s.SleepNeed = m["sleep_need"]; s.Carpentry = m["carpentry"]; s.VisionStat = m["vision_stat"]
	s.MysticalAptitude = m["mystical_aptitude"]; s.MiracleReceptivity = m["miracle_receptivity"]
	s.CurseVulnerability = m["curse_vulnerability"]
}

func BlendStats(a, b Stats) Stats {
	ma := a.toMap()
	mb := b.toMap()
	result := make(map[string]int)
	for key := range ma {
		avg := float64(ma[key]+mb[key]) / 2
		result[key] = clamp(int(math.Round(avg+gauss(0, 10))), 1, 100)
	}
	result["reputation"] = clamp(int(math.Round(gauss(50, 10))), 1, 100)
	result["infamy"] = 0
	result["depression"] = clamp(int(math.Round(gauss(10, 8))), 0, 100)
	result["trauma"] = 0
	var s Stats
	s.fromMap(result)
	return s
}
