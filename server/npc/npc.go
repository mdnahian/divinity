package npc

import (
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"

	"github.com/divinity/core/config"
	"github.com/divinity/core/item"
	"github.com/google/uuid"
)

var npcCounter atomic.Int64

type Needs struct {
	Hunger     float64 `json:"hunger"`
	Thirst     float64 `json:"thirst"`
	Fatigue    float64 `json:"fatigue"`
	SocialNeed float64 `json:"socialNeed"`
}

type Stats struct {
	Strength                int `json:"strength"`
	Agility                 int `json:"agility"`
	Health                  int `json:"health"`
	Intelligence            int `json:"intelligence"`
	Wisdom                  int `json:"wisdom"`
	Curiosity               int `json:"curiosity"`
	Creativity              int `json:"creativity"`
	Charisma                int `json:"charisma"`
	Empathy                 int `json:"empathy"`
	Extraversion            int `json:"extraversion"`
	Reputation              int `json:"reputation"`
	Aggression              int `json:"aggression"`
	Virtue                  int `json:"virtue"`
	Greed                   int `json:"greed"`
	Ambition                int `json:"ambition"`
	Conscientiousness       int `json:"conscientiousness"`
	Courage                 int `json:"courage"`
	Endurance               int `json:"endurance"`
	Dexterity               int `json:"dexterity"`
	SpiritualSensitivity    int `json:"spiritualSensitivity"`
	PrayerFrequency         int `json:"prayerFrequency"`
	FaithCapacity           int `json:"faithCapacity"`
	DoubtResistance         int `json:"doubtResistance"`
	MysticalAptitude        int `json:"mysticalAptitude"`
	MiracleReceptivity      int `json:"miracleReceptivity"`
	DivineLuck              int `json:"divineLuck"`
	CurseVulnerability      int `json:"curseVulnerability"`
	PainTolerance           int `json:"painTolerance"`
	DiseaseResistance       int `json:"diseaseResistance"`
	Sanity                  int `json:"sanity"`
	Resilience              int `json:"resilience"`
	Openness                int `json:"openness"`
	Neuroticism             int `json:"neuroticism"`
	Depression              int `json:"depression"`
	Trauma                  int `json:"trauma"`
	Dominance               int `json:"dominance"`
	StrategicThinking       int `json:"strategicThinking"`
	Decisiveness            int `json:"decisiveness"`
	Persuasion              int `json:"persuasion"`
	TeachingDrive           int `json:"teachingDrive"`
	Loyalty                 int `json:"loyalty"`
	Jealousy                int `json:"jealousy"`
	TrustThreshold          int `json:"trustThreshold"`
	Generosity              int `json:"generosity"`
	Conformity              int `json:"conformity"`
	Infamy                  int `json:"infamy"`
	Attractiveness          int `json:"attractiveness"`
	MemoryCapacity          int `json:"memoryCapacity"`
	SocialAwareness         int `json:"socialAwareness"`
	Composure               int `json:"composure"`
	Accountability          int `json:"accountability"`
	PoliticalInstinct       int `json:"politicalInstinct"`
	AddictionSusceptibility int `json:"addictionSusceptibility"`
	MentalFragility         int `json:"mentalFragility"`
	HungerRate              int `json:"hungerRate"`
	SleepNeed               int `json:"sleepNeed"`
	Carpentry               int `json:"carpentry"`
	VisionStat              int `json:"visionStat"`
}

func (s *Stats) ByKey(key string) int {
	m := s.toMap()
	if v, ok := m[key]; ok {
		return v
	}
	return 0
}

func (s *Stats) SetByKey(key string, val int) {
	m := s.toMap()
	m[key] = val
	s.fromMap(m)
}

type EquipmentSlot struct {
	Name       string  `json:"name"`
	Durability float64 `json:"durability"`
}

type Equipment struct {
	Weapon *EquipmentSlot `json:"weapon"`
	Armor  *EquipmentSlot `json:"armor"`
	Bag1   *EquipmentSlot `json:"bag1"`
	Bag2   *EquipmentSlot `json:"bag2"`
}

type JournalMeta struct {
	Type         string         `json:"type"` // enemy_warning, resources, skill_notes, trade_info, people, personal
	LocationName string         `json:"locationName"`
	Enemies      []string       `json:"enemies"`
	Resources    map[string]int `json:"resources"`
	Skill        string         `json:"skill"`
	Level        int            `json:"level"`
	Prices       map[string]int `json:"prices"`
	NpcName      string         `json:"npcName"`
	Observation  string         `json:"observation"`
	Reason       string         `json:"reason"`
	WriterName   string         `json:"writerName"`
}

type InventoryItem struct {
	Name        string       `json:"name"`
	Qty         int          `json:"qty"`
	Durability  float64      `json:"durability"`
	JournalMeta *JournalMeta `json:"journalMeta"`
}

type NPC struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Profession     string `json:"profession"`
	HomeID         string `json:"homeId"`
	LocationID     string `json:"locationId"`
	Color          string `json:"color"`
	Alive          bool   `json:"alive"`
	CauseOfDeath   string `json:"causeOfDeath"`
	HP             int    `json:"hp"`
	DeathTimer     int    `json:"deathTimer"`
	BirthDay       int    `json:"birthDay"`
	BaseLifespan   int    `json:"baseLifespan"`
	Fertility      int    `json:"fertility"`
	GeneticQuality int    `json:"geneticQuality"`
	Literacy       int    `json:"literacy"`

	Stats     Stats `json:"stats"`
	Needs     Needs `json:"needs"`
	Happiness int   `json:"happiness"`
	Stress    int   `json:"stress"`
	Hygiene   int   `json:"hygiene"`
	Sobriety  int   `json:"sobriety"`

	Skills        map[string]float64 `json:"skills"`
	Relationships map[string]int     `json:"relationships"`
	Inventory     []InventoryItem    `json:"inventory"`
	Equipment     Equipment          `json:"equipment"`

	FactionID       string  `json:"factionId"`
	LeadershipScore float64 `json:"leadershipScore"`
	HomeBuildingID  string  `json:"homeBuildingId"`

	EmployerID      string `json:"employerId"`
	WorkplaceID     string `json:"workplaceId"`
	Wage            int    `json:"wage"`
	IsBusinessOwner bool   `json:"isBusinessOwner"`
	UnpaidDays      int    `json:"unpaidDays"`

	Apprentices  []string `json:"apprentices"`
	ApprenticeTo string   `json:"apprenticeTo"`

	Busy           bool   `json:"busy"`
	LastAction     string `json:"lastAction"`
	LastDialogue   string `json:"lastDialogue"`
	LastActionTick int    `json:"lastActionTick"`
	CurrentGoal    string `json:"currentGoal"`

	BusyUntilTick     int    `json:"busyUntilTick"`
	PendingActionID   string `json:"pendingActionId"`
	PendingTargetName string `json:"pendingTargetName"`
	PendingReason     string `json:"pendingReason"`
	ResumeActionID    string `json:"resumeActionId"`
	ResumeTicksLeft   int    `json:"resumeTicksLeft"`

	TargetLocationID   string `json:"targetLocationId"`
	TravelStartTick    int    `json:"travelStartTick"`
	TravelArrivalTick  int    `json:"travelArrivalTick"`
	PreviousLocationID string `json:"previousLocationId"`

	Claimed     bool     `json:"claimed"`
	Personality string   `json:"personality"`
	ParentID    string   `json:"parentId,omitempty"`
	ChildIDs    []string `json:"childIds,omitempty"`

	// Noble hierarchy fields
	NobleRank   string   `json:"nobleRank,omitempty"`   // "", "king", "queen", "duke", "count", "baron", "knight"
	TerritoryID string   `json:"territoryId,omitempty"` // territory they rule (if any)
	LiegeID     string   `json:"liegeId,omitempty"`     // feudal superior NPC ID
	VassalIDs   []string `json:"vassalIds,omitempty"`   // NPCs who swear fealty
	SpouseID    string   `json:"spouseId,omitempty"`    // married partner NPC ID

	// Mount ownership
	MountID    string `json:"mountId,omitempty"`    // owned mount ID
	CarriageID string `json:"carriageId,omitempty"` // owned carriage ID
}

type Template struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Age         int                `json:"age"`
	Profession  string             `json:"profession"`
	HomeID      string             `json:"homeId"`
	LocationID  string             `json:"locationId"`
	Personality string             `json:"personality"`
	Items       []string           `json:"items"`
	StatBias    map[string]int     `json:"statBias"`
	Literacy    int                `json:"literacy"`
	Skills      map[string]float64 `json:"skills"`
	StartItems  []InventoryItem    `json:"startItems"`
	Stats       *Stats             `json:"stats"`
	BirthDay    int                `json:"birthDay"`
	Equipment   *Equipment         `json:"equipment"`

	BaseLifespan   int    `json:"baseLifespan"`
	Fertility      int    `json:"fertility"`
	GeneticQuality int    `json:"geneticQuality"`
	Happiness      int    `json:"happiness"`
	Stress         int    `json:"stress"`
	Hygiene        int    `json:"hygiene"`
	Sobriety       int    `json:"sobriety"`
	Needs          *Needs `json:"needs"`

	EmployerID      string `json:"employerId"`
	WorkplaceID     string `json:"workplaceId"`
	Wage            int    `json:"wage"`
	IsBusinessOwner bool   `json:"isBusinessOwner"`
	ParentID        string `json:"parentId,omitempty"`
}

func NewNPC(tmpl Template, index int, worldDay int, cfg *config.Config) *NPC {
	id := tmpl.ID
	if id == "" {
		id = uuid.NewString()
	}

	colors := []string{
		"#e74c3c", "#3498db", "#2ecc71", "#e67e22", "#9b59b6", "#1abc9c",
		"#f39c12", "#e91e63", "#00bcd4", "#8bc34a", "#ff5722", "#607d8b",
		"#795548", "#009688", "#673ab7", "#ff9800", "#03a9f4", "#cddc39",
	}

	age := tmpl.Age
	if age == 0 {
		age = 25
	}
	birthDay := tmpl.BirthDay
	if birthDay == 0 {
		birthDay = worldDay - age*cfg.Game.GameDaysPerYear
	}

	var stats Stats
	if tmpl.Stats != nil {
		stats = *tmpl.Stats
	} else {
		stats = generateStats(tmpl.StatBias)
	}

	baseLifespan := tmpl.BaseLifespan
	if baseLifespan == 0 {
		baseLifespan = max(50, int(math.Round(gauss(70, 10))))
	}
	fertility := tmpl.Fertility
	if fertility == 0 {
		fertility = clamp(int(math.Round(gauss(50, 15))), 1, 100)
	}
	gq := tmpl.GeneticQuality
	if gq == 0 {
		gq = clamp(int(math.Round(gauss(50, 15))), 1, 100)
	}
	literacy := tmpl.Literacy
	if literacy == 0 {
		literacy = clamp(int(math.Round(gauss(30, 20))), 0, 100)
	}

	needs := Needs{
		Hunger:     float64(randInt(55, 80)),
		Thirst:     float64(randInt(55, 80)),
		Fatigue:    float64(randInt(10, 30)),
		SocialNeed: float64(randInt(20, 50)),
	}
	if tmpl.Needs != nil {
		needs = *tmpl.Needs
	}

	happiness := tmpl.Happiness
	if happiness == 0 {
		happiness = randInt(50, 75)
	}
	stress := tmpl.Stress
	if stress == 0 {
		stress = randInt(10, 35)
	}
	hygiene := tmpl.Hygiene
	if hygiene == 0 {
		hygiene = randInt(60, 90)
	}
	sobriety := tmpl.Sobriety
	if sobriety == 0 {
		sobriety = 100
	}

	skills := tmpl.Skills
	if skills == nil {
		skills = map[string]float64{tmpl.Profession: float64(randInt(30, 50))}
	}

	inv := make([]InventoryItem, len(tmpl.StartItems))
	copy(inv, tmpl.StartItems)

	eq := Equipment{}
	if tmpl.Equipment != nil {
		eq = *tmpl.Equipment
	}

	return &NPC{
		ID:              id,
		Name:            tmpl.Name,
		Profession:      tmpl.Profession,
		HomeID:          tmpl.HomeID,
		LocationID:      firstNonEmpty(tmpl.LocationID, tmpl.HomeID),
		Color:           colors[index%len(colors)],
		Alive:           true,
		HP:              100,
		BirthDay:        birthDay,
		BaseLifespan:    baseLifespan,
		Fertility:       fertility,
		GeneticQuality:  gq,
		Literacy:        literacy,
		Stats:           stats,
		Needs:           needs,
		Happiness:       happiness,
		Stress:          stress,
		Hygiene:         hygiene,
		Sobriety:        sobriety,
		Skills:          skills,
		Relationships:   make(map[string]int),
		Inventory:       inv,
		Equipment:       eq,
		EmployerID:      tmpl.EmployerID,
		WorkplaceID:     tmpl.WorkplaceID,
		Wage:            tmpl.Wage,
		IsBusinessOwner: tmpl.IsBusinessOwner,
		Personality:     tmpl.Personality,
		ParentID:        tmpl.ParentID,
	}
}

// IsNoble returns true if the NPC holds any noble rank.
func (n *NPC) IsNoble() bool {
	switch n.NobleRank {
	case "king", "queen", "duke", "count", "baron", "knight", "prince", "princess":
		return true
	}
	return false
}

// IsHighNoble returns true for nobles who should NOT perform manual labor.
func (n *NPC) IsHighNoble() bool {
	switch n.NobleRank {
	case "king", "queen", "duke", "count", "prince", "princess":
		return true
	}
	return false
}

var nobleHouseholdProfs = map[string]bool{
	"steward": true, "guard": true, "stablehand": true, "healer": true,
}

// CanAccessNobleArea returns true if this NPC is allowed in restricted noble locations.
func (n *NPC) CanAccessNobleArea() bool {
	return n.IsNoble() || nobleHouseholdProfs[n.Profession]
}

func (n *NPC) GetAge(worldDay int, daysPerYear int) int {
	return max(0, (worldDay-n.BirthDay)/daysPerYear)
}

func (n *NPC) GetLifeStage(worldDay, daysPerYear int) string {
	age := n.GetAge(worldDay, daysPerYear)
	switch {
	case age < 5:
		return "infant"
	case age < 12:
		return "child"
	case age < 18:
		return "adolescent"
	case age < 55:
		return "adult"
	case age < 70:
		return "elder"
	default:
		return "ancient"
	}
}

func (n *NPC) GetEffectiveWisdom(worldDay, daysPerYear int) int {
	stage := n.GetLifeStage(worldDay, daysPerYear)
	mult := 1.0
	switch stage {
	case "ancient":
		mult = 3.0
	case "elder":
		mult = 1.5
	}
	return min(100, int(math.Round(float64(n.Stats.Wisdom)*mult)))
}

func (n *NPC) GetRelationship(otherID string) int {
	return n.Relationships[otherID]
}

func (n *NPC) AdjustRelationship(otherID string, delta int) {
	current := n.Relationships[otherID]
	n.Relationships[otherID] = clamp(current+delta, -100, 100)
}

func (n *NPC) GetSkillLevel(name string) float64 {
	return n.Skills[name]
}

func (n *NPC) GainSkill(name string, amount float64) {
	n.Skills[name] = math.Min(100, n.Skills[name]+amount)
}

func (n *NPC) GoldCount() int {
	for _, it := range n.Inventory {
		if it.Name == "gold" && it.Qty > 0 {
			return it.Qty
		}
	}
	return 0
}

func (n *NPC) HasItem(name string) *InventoryItem {
	for i := range n.Inventory {
		if n.Inventory[i].Name == name && n.Inventory[i].Qty > 0 {
			return &n.Inventory[i]
		}
	}
	return nil
}

func (n *NPC) RemoveItem(name string, qty int) bool {
	for i := range n.Inventory {
		if n.Inventory[i].Name == name && n.Inventory[i].Qty > 0 {
			n.Inventory[i].Qty -= qty
			if n.Inventory[i].Qty <= 0 {
				n.Inventory = append(n.Inventory[:i], n.Inventory[i+1:]...)
			}
			return true
		}
	}
	return false
}

func (n *NPC) AddItem(name string, qty int) {
	for i := range n.Inventory {
		if n.Inventory[i].Name == name {
			n.Inventory[i].Qty += qty
			return
		}
	}
	info := item.GetInfo(name)
	n.Inventory = append(n.Inventory, InventoryItem{
		Name:       name,
		Qty:        qty,
		Durability: float64(info.DurabilityBase),
	})
}

func (n *NPC) AddItemSafe(name string, qty int, dropFn func(string, int)) bool {
	if n.CanCarry(name, qty) {
		n.AddItem(name, qty)
		return true
	}
	if dropFn != nil {
		dropFn(name, qty)
	}
	return false
}

func (n *NPC) MaxSlots() int {
	base := 6
	for _, eq := range []*EquipmentSlot{n.Equipment.Bag1, n.Equipment.Bag2} {
		if eq != nil {
			info := item.GetInfo(eq.Name)
			base += info.BagSlots
		}
	}
	return base
}

func (n *NPC) MaxWeight() float64 {
	base := 15.0 + math.Max(0, float64(n.Stats.Strength-50))*0.3
	for _, eq := range []*EquipmentSlot{n.Equipment.Bag1, n.Equipment.Bag2} {
		if eq != nil {
			info := item.GetInfo(eq.Name)
			base += float64(info.BagWeight)
		}
	}
	return base
}

func (n *NPC) UsedSlots() int {
	return len(n.Inventory)
}

func (n *NPC) UsedWeight() float64 {
	var w float64
	for _, it := range n.Inventory {
		w += item.GetWeight(it.Name, it.Qty)
	}
	return math.Round(w*100) / 100
}

func (n *NPC) CanCarry(itemName string, qty int) bool {
	info := item.GetInfo(itemName)
	addWeight := info.Weight * float64(qty)
	if n.UsedWeight()+addWeight > n.MaxWeight() {
		return false
	}
	existing := n.HasItem(itemName)
	if existing == nil {
		return n.UsedSlots() < n.MaxSlots()
	}
	return true
}

func (n *NPC) EquippedWeaponBonus() int {
	w := n.Equipment.Weapon
	if w == nil {
		return 0
	}
	def := item.GetInfo(w.Name)
	if b := def.Effects["weapon_bonus"]; b > 0 {
		return int(b)
	}
	return 5
}

func (n *NPC) EquippedArmorBonus() int {
	a := n.Equipment.Armor
	if a == nil {
		return 0
	}
	def := item.GetInfo(a.Name)
	if b := def.Effects["armor_bonus"]; b > 0 {
		return int(b)
	}
	return 2
}

func (n *NPC) HasItemOfCategory(cat string) *InventoryItem {
	for i := range n.Inventory {
		if n.Inventory[i].Qty > 0 {
			def := item.GetInfo(n.Inventory[i].Name)
			if def.Category == cat {
				return &n.Inventory[i]
			}
		}
	}
	return nil
}

func (n *NPC) HasProfessionOrSkill(profName string, skillKey string, minSkill int) bool {
	if n.Profession == profName {
		return true
	}
	def := GetProfessionDef(n.Profession)
	if def.PrimarySkill == skillKey {
		return true
	}
	return int(n.GetSkillLevel(skillKey)) >= minSkill
}

func (n *NPC) EquipItem(itemName string) bool {
	slot := n.HasItem(itemName)
	if slot == nil {
		return false
	}
	info := item.GetInfo(itemName)
	if info.Slot == "" {
		return false
	}
	switch info.Slot {
	case "weapon":
		if n.Equipment.Weapon != nil {
			n.AddItem(n.Equipment.Weapon.Name, 1)
		}
		n.RemoveItem(itemName, 1)
		n.Equipment.Weapon = &EquipmentSlot{Name: itemName, Durability: float64(info.DurabilityBase)}
	case "armor":
		if n.Equipment.Armor != nil {
			n.AddItem(n.Equipment.Armor.Name, 1)
		}
		n.RemoveItem(itemName, 1)
		n.Equipment.Armor = &EquipmentSlot{Name: itemName, Durability: float64(info.DurabilityBase)}
	case "bag":
		if n.Equipment.Bag1 == nil {
			n.RemoveItem(itemName, 1)
			n.Equipment.Bag1 = &EquipmentSlot{Name: itemName, Durability: float64(info.DurabilityBase)}
		} else if n.Equipment.Bag2 == nil {
			n.RemoveItem(itemName, 1)
			n.Equipment.Bag2 = &EquipmentSlot{Name: itemName, Durability: float64(info.DurabilityBase)}
		} else {
			return false
		}
	default:
		return false
	}
	return true
}

func (n *NPC) UnequipItem(slotName string) bool {
	var eq **EquipmentSlot
	switch slotName {
	case "weapon":
		eq = &n.Equipment.Weapon
	case "armor":
		eq = &n.Equipment.Armor
	case "bag_1":
		eq = &n.Equipment.Bag1
	case "bag_2":
		eq = &n.Equipment.Bag2
	default:
		return false
	}
	if *eq == nil {
		return false
	}
	if !n.CanCarry((*eq).Name, 1) {
		return false
	}
	n.AddItem((*eq).Name, 1)
	*eq = nil
	return true
}

func (n *NPC) DecayNeeds(cfg *config.Config) {
	if !n.Alive {
		return
	}
	d := cfg.Decay
	sleeping := n.PendingActionID == "sleep"
	if sleeping {
		// While sleeping: halve hunger/thirst drain, reduce fatigue instead of increasing it
		n.Needs.Hunger = clampF(n.Needs.Hunger-d.Hunger*0.5, 0, 100)
		n.Needs.Thirst = clampF(n.Needs.Thirst-d.Thirst*0.5, 0, 100)
		n.Needs.Fatigue = clampF(n.Needs.Fatigue-0.5, 0, 100)
	} else {
		n.Needs.Hunger = clampF(n.Needs.Hunger-d.Hunger, 0, 100)
		n.Needs.Thirst = clampF(n.Needs.Thirst-d.Thirst, 0, 100)
		n.Needs.Fatigue = clampF(n.Needs.Fatigue+d.Fatigue, 0, 100)
	}
	n.Needs.SocialNeed = clampF(n.Needs.SocialNeed+d.SocialNeed, 0, 100)
	// Happiness decays down over time; use probabilistic rounding since rate < 1
	if d.Happiness > 0 && rand.Float64() < d.Happiness {
		n.Happiness = clamp(n.Happiness-1, 0, 100)
	}
	// Stress decays toward zero over time; d.Stress is negative (e.g. -0.072)
	if d.Stress < 0 && rand.Float64() < -d.Stress {
		n.Stress = clamp(n.Stress-1, 0, 100)
	}

	if n.Needs.Hunger < float64(cfg.Thresholds.HungerUrgent) {
		n.Stress = clamp(n.Stress+2, 0, 100)
		n.Happiness = clamp(n.Happiness-1, 0, 100)
	}
	if n.Needs.Thirst < float64(cfg.Thresholds.ThirstUrgent) {
		n.Stress = clamp(n.Stress+3, 0, 100)
	}
	if n.Needs.Fatigue > float64(cfg.Thresholds.FatigueUrgent) {
		n.Happiness = clamp(n.Happiness-1, 0, 100)
	}
	if n.HP < 100 {
		n.HP = min(100, n.HP+1)
	}
}

func (n *NPC) DecayInventoryDurability() {
	if !n.Alive {
		return
	}
	careful := n.Stats.Conscientiousness > 60
	for i := len(n.Inventory) - 1; i >= 0; i-- {
		it := &n.Inventory[i]
		if it.Durability == 0 {
			continue
		}
		info := item.GetInfo(it.Name)
		rate := info.DecayRate
		if careful {
			rate *= 0.5
		}
		it.Durability = math.Max(0, it.Durability-rate)
		if it.Durability <= 0 && it.Name != "gold" {
			n.Inventory = append(n.Inventory[:i], n.Inventory[i+1:]...)
		}
	}
	for _, slot := range []**EquipmentSlot{&n.Equipment.Weapon, &n.Equipment.Armor, &n.Equipment.Bag1, &n.Equipment.Bag2} {
		eq := *slot
		if eq == nil {
			continue
		}
		info := item.GetInfo(eq.Name)
		rate := info.DecayRate
		if careful {
			rate *= 0.5
		}
		eq.Durability = math.Max(0, eq.Durability-rate)
		if eq.Durability <= 0 {
			*slot = nil
		}
	}
}

func (n *NPC) DailyHygieneSobriety() {
	if !n.Alive {
		return
	}
	n.Hygiene = clamp(n.Hygiene-3, 0, 100)
	n.Sobriety = clamp(n.Sobriety+5, 0, 100)

	if n.Hygiene < 20 {
		n.Stats.DiseaseResistance = clamp(n.Stats.DiseaseResistance-2, 0, 100)
		n.Stats.Reputation = clamp(n.Stats.Reputation-2, 0, 100)
	}
	if n.Sobriety < 30 {
		n.Stats.Aggression = clamp(n.Stats.Aggression+2, 0, 100)
		n.Stress = clamp(n.Stress+5, 0, 100)
	}
}

func (n *NPC) CheckDeath(graceTicks int) string {
	if !n.Alive {
		return ""
	}
	starving := n.Needs.Hunger <= 0
	dehydrated := n.Needs.Thirst <= 0
	dead := n.HP <= 0

	if starving || dehydrated || dead {
		n.DeathTimer++
	} else {
		n.DeathTimer = 0
		return ""
	}

	if n.DeathTimer < graceTicks && !dead {
		return ""
	}

	switch {
	case dead:
		n.Alive = false
		n.CauseOfDeath = "injuries"
	case starving && dehydrated:
		n.Alive = false
		n.CauseOfDeath = "starvation and dehydration"
	case starving:
		n.Alive = false
		n.CauseOfDeath = "starvation"
	default:
		n.Alive = false
		n.CauseOfDeath = "dehydration"
	}
	return n.CauseOfDeath
}

func (n *NPC) CheckOldAge(worldDay, daysPerYear int) string {
	if !n.Alive {
		return ""
	}
	age := n.GetAge(worldDay, daysPerYear)
	if age >= n.BaseLifespan {
		chance := float64(age-n.BaseLifespan+1) * 0.15
		if rand.Float64() < chance {
			n.Alive = false
			n.CauseOfDeath = fmt.Sprintf("old age (%d years)", age)
			return n.CauseOfDeath
		}
	}
	return ""
}

func (n *NPC) GetHomeTier(locations []interface {
	GetBuildingID() string
	GetTier() int
}) int {
	if n.HomeBuildingID == "" {
		return 0
	}
	for _, loc := range locations {
		if loc.GetBuildingID() == n.HomeBuildingID {
			return loc.GetTier()
		}
	}
	return 0
}

func (n *NPC) HomeAmbitionTarget() int {
	ambition := n.Stats.Ambition
	switch {
	case ambition < 30:
		return 1
	case ambition < 50:
		return 2
	case ambition < 70:
		return 3
	default:
		return 4
	}
}

func (n *NPC) LiteracyLevel() string {
	switch {
	case n.Literacy < 10:
		return "illiterate"
	case n.Literacy < 31:
		return "symbol recognition"
	case n.Literacy < 61:
		return "functional"
	case n.Literacy < 86:
		return "educated"
	default:
		return "scholar"
	}
}

func (n *NPC) IsInjured() bool {
	return n.HP < 70
}

// Helpers

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func randInt(lo, hi int) int {
	if lo >= hi {
		return lo
	}
	return lo + rand.Intn(hi-lo+1)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func gauss(mean, std float64) float64 {
	return mean + rand.NormFloat64()*std
}
