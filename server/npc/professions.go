package npc

import "sync"

// ProfessionDef describes the mechanical properties of a profession.
type ProfessionDef struct {
	PrimarySkill  string // skill key gained by working in this profession
	LiteracyBonus int    // base literacy for newly created NPCs of this profession
	Description   string // shown in GOD context
}

var profMu sync.RWMutex

// ProfessionRegistry maps profession name → definition.
// Pre-populated with all built-in professions; new entries added at runtime via RegisterProfession.
var ProfessionRegistry = map[string]ProfessionDef{
	// Common professions
	"farmer":     {PrimarySkill: "farmer", LiteracyBonus: 0, Description: "Grows crops and tends the land."},
	"hunter":     {PrimarySkill: "hunter", LiteracyBonus: 0, Description: "Tracks and hunts wild animals."},
	"blacksmith": {PrimarySkill: "blacksmith", LiteracyBonus: 5, Description: "Forges metal tools and weapons."},
	"carpenter":  {PrimarySkill: "carpenter", LiteracyBonus: 0, Description: "Works with wood to build structures."},
	"merchant":   {PrimarySkill: "merchant", LiteracyBonus: 55, Description: "Trades goods for profit."},
	"herbalist":  {PrimarySkill: "herbalist", LiteracyBonus: 45, Description: "Gathers and uses medicinal herbs."},
	"healer":     {PrimarySkill: "healer", LiteracyBonus: 50, Description: "Tends to the sick and injured."},
	"scribe":     {PrimarySkill: "writing", LiteracyBonus: 75, Description: "Records knowledge and correspondence."},
	"miner":      {PrimarySkill: "miner", LiteracyBonus: 0, Description: "Extracts ore and minerals from the earth."},
	"fisher":     {PrimarySkill: "fisher", LiteracyBonus: 0, Description: "Catches fish from rivers and lakes."},
	"tailor":     {PrimarySkill: "tailor", LiteracyBonus: 10, Description: "Crafts and repairs clothing."},
	"potter":     {PrimarySkill: "potter", LiteracyBonus: 0, Description: "Shapes clay into vessels and goods."},
	"barmaid":    {PrimarySkill: "innkeeping", LiteracyBonus: 10, Description: "Serves patrons at the inn."},
	"elder":      {PrimarySkill: "wisdom", LiteracyBonus: 60, Description: "Respected village elder with deep knowledge."},
	"guard":      {PrimarySkill: "combat", LiteracyBonus: 5, Description: "Protects the village from threats."},
	"baker":      {PrimarySkill: "cooking", LiteracyBonus: 5, Description: "Bakes bread and other goods."},
	"cook":       {PrimarySkill: "cooking", LiteracyBonus: 5, Description: "Prepares meals for the village."},
	"innkeeper":  {PrimarySkill: "innkeeping", LiteracyBonus: 20, Description: "Runs the village inn."},
	// Noble professions
	"king":             {PrimarySkill: "leadership", LiteracyBonus: 90, Description: "Supreme ruler of the kingdom."},
	"queen":            {PrimarySkill: "diplomacy", LiteracyBonus: 90, Description: "Royal consort and co-ruler of the kingdom."},
	"prince":           {PrimarySkill: "leadership", LiteracyBonus: 80, Description: "Royal heir to the throne."},
	"princess":         {PrimarySkill: "diplomacy", LiteracyBonus: 80, Description: "Royal princess of the realm."},
	"duke":             {PrimarySkill: "governance", LiteracyBonus: 85, Description: "Rules a territory on behalf of the crown."},
	"count":            {PrimarySkill: "governance", LiteracyBonus: 75, Description: "Manages a county within a duchy."},
	"baron":            {PrimarySkill: "governance", LiteracyBonus: 70, Description: "Oversees a barony and its lands."},
	"knight":           {PrimarySkill: "combat", LiteracyBonus: 50, Description: "Sworn warrior in service to a lord."},
	"steward":          {PrimarySkill: "management", LiteracyBonus: 70, Description: "Manages a noble household's affairs."},
	// Horse & transport professions
	"stablehand":       {PrimarySkill: "animal_care", LiteracyBonus: 10, Description: "Cares for horses and manages stables."},
	"carriage_driver":  {PrimarySkill: "driving", LiteracyBonus: 15, Description: "Drives carriages between towns."},
}

// RegisterProfession adds or updates a profession in the registry at runtime.
func RegisterProfession(name, primarySkill string, literacyBonus int, description string) {
	profMu.Lock()
	defer profMu.Unlock()
	ProfessionRegistry[name] = ProfessionDef{
		PrimarySkill:  primarySkill,
		LiteracyBonus: literacyBonus,
		Description:   description,
	}
}

// GetProfessionDef returns the definition for a profession, or a default if unknown.
func GetProfessionDef(profession string) ProfessionDef {
	profMu.RLock()
	defer profMu.RUnlock()
	if def, ok := ProfessionRegistry[profession]; ok {
		return def
	}
	return ProfessionDef{PrimarySkill: profession, LiteracyBonus: 0}
}
