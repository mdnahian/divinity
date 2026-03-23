package memory

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

const (
	MaxMemories    = 20
	GodMaxMemories = 100 // GOD manages an entire world and needs much more history
	DefaultDecay   = 0.05
	ImportantDecay = 0.02
	TraumaticDecay = 0.01
	DreamPrefix       = "DREAM: "
	DivineDreamPrefix = "DIVINE DREAM: "
	VisionPrefix      = "DIVINE VISION: "
	OmenPrefix        = "OMEN: "
	VividThreshold = 0.1
	GodEntityID    = "__god__"
)

// Categories for memory entries.
const (
	CatCombat     = "combat"
	CatEconomic   = "economic"
	CatSocial     = "social"
	CatEmployment = "employment"
	CatMiracle    = "miracle"
	CatDream      = "dream"
	CatRoutine    = "routine"
	CatEducation  = "education"
	CatFaction    = "faction"
	CatReflection = "reflection"
	CatHeard      = "heard"
	CatGod = "god"

	// GOD subcategories for filtered retrieval
	CatGodSpawn    = "god:spawn"
	CatGodMiracle  = "god:miracle"
	CatGodEvent    = "god:event"
	CatGodResource = "god:resource"
	CatGodCrisis   = "god:crisis"
	CatGodGrowth   = "god:growth"
	CatGodWeather  = "god:weather"
	CatGodCurse    = "god:curse"
	CatGodDream    = "god:dream"
	CatGodVision   = "god:vision"
	CatGodOmen     = "god:omen"
)

type Entry struct {
	Text       string   `json:"text" bson:"text"`
	Time       string   `json:"time" bson:"time"`
	TS         int64    `json:"ts" bson:"ts"`
	Importance float64  `json:"importance" bson:"importance"`
	Vividness  float64  `json:"vividness" bson:"vividness"`
	Category   string   `json:"category" bson:"category"`
	Tags       []string `json:"tags" bson:"tags"`
}

// EffectiveCap computes the per-NPC memory cap based on MemoryCapacity stat (1-100).
// Range: 11-20 memories.
func EffectiveCap(memoryCapacity int) int {
	return 10 + max(1, memoryCapacity/10)
}

// DecayRate returns the per-day vividness decay rate based on importance.
func (e *Entry) DecayRate() float64 {
	switch {
	case e.Importance >= 0.8:
		return TraumaticDecay
	case e.Importance >= 0.5:
		return ImportantDecay
	default:
		return DefaultDecay
	}
}

type Store interface {
	// Add appends a memory entry for the given entity. capOverride <= 0 uses MaxMemories.
	Add(npcID string, entry Entry)
	// AddWithCap appends a memory entry with a custom capacity limit (from MemoryCapacity stat).
	AddWithCap(npcID string, entry Entry, cap int)
	Recent(npcID string, count int) []Entry
	All(npcID string) []Entry
	FormatForLLM(npcID string, count int) string
	Clear(npcID string)

	// Filtered retrieval
	ByCategory(npcID string, category string) []Entry
	ByTag(npcID string, tag string) []Entry
	Search(npcID string, keyword string) []Entry
	HighestImportance(npcID string, count int) []Entry

	// Decay: apply daily vividness decay and evict faded memories.
	ApplyDecay(npcID string)

	// Blended retrieval for NPC behavioral prompts
	FormatBlendedForLLM(npcID string) string

	// Persistence
	Save(npcID string, entry Entry) error
	LoadAll() (map[string][]Entry, error)
	ClearPersisted(npcID string) error
}

type InMemoryStore struct {
	mu     sync.RWMutex
	stores map[string][]Entry
	db     Persister // nil if no DB available
}

// Persister is the interface for MongoDB persistence of memories.
type Persister interface {
	SaveMemory(npcID string, entry Entry) error
	LoadAllMemories() (map[string][]Entry, error)
	ClearMemories(npcID string) error
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{stores: make(map[string][]Entry)}
}

func NewInMemoryStoreWithDB(p Persister) *InMemoryStore {
	return &InMemoryStore{stores: make(map[string][]Entry), db: p}
}

// HydrateFromDB loads all persisted memories into the in-memory store.
func (m *InMemoryStore) HydrateFromDB() error {
	if m.db == nil {
		return nil
	}
	all, err := m.db.LoadAllMemories()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for npcID, entries := range all {
		m.stores[npcID] = entries
	}
	return nil
}

func (m *InMemoryStore) Add(npcID string, entry Entry) {
	cap := MaxMemories
	if npcID == GodEntityID {
		cap = GodMaxMemories
	}
	m.AddWithCap(npcID, entry, cap)
}

func (m *InMemoryStore) AddWithCap(npcID string, entry Entry, cap int) {
	if cap <= 0 {
		cap = MaxMemories
	}
	// Default vividness to 1.0 if not set
	if entry.Vividness == 0 {
		entry.Vividness = 1.0
	}
	// Default importance to routine if not set
	if entry.Importance == 0 {
		entry.Importance = 0.1
	}

	m.mu.Lock()
	s := m.stores[npcID]
	s = append(s, entry)
	if len(s) > cap {
		s = evictLowest(s, cap)
	}
	m.stores[npcID] = s
	m.mu.Unlock()

	// Persist asynchronously (best-effort)
	if m.db != nil {
		go m.db.SaveMemory(npcID, entry)
	}
}

// evictLowest removes the entry with the lowest eviction score to bring the slice to cap size.
// evictionScore = importance * (0.3 + 0.7 * vividness)
func evictLowest(entries []Entry, cap int) []Entry {
	for len(entries) > cap {
		minIdx := 0
		minScore := math.MaxFloat64
		for i, e := range entries {
			score := e.Importance * (0.3 + 0.7*e.Vividness)
			if score < minScore {
				minScore = score
				minIdx = i
			}
		}
		entries = append(entries[:minIdx], entries[minIdx+1:]...)
	}
	return entries
}

func (m *InMemoryStore) Recent(npcID string, count int) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.stores[npcID]
	if len(s) <= count {
		out := make([]Entry, len(s))
		copy(out, s)
		return out
	}
	out := make([]Entry, count)
	copy(out, s[len(s)-count:])
	return out
}

func (m *InMemoryStore) All(npcID string) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.stores[npcID]
	out := make([]Entry, len(s))
	copy(out, s)
	return out
}

func (m *InMemoryStore) FormatForLLM(npcID string, count int) string {
	entries := m.Recent(npcID, count)
	if len(entries) == 0 {
		return "No recent memories."
	}
	result := ""
	for i, e := range entries {
		time := e.Time
		if time == "" {
			time = "?"
		}
		prefix := ""
		if e.Vividness < 0.3 {
			prefix = "(vague) "
		} else if e.Vividness < 0.5 {
			prefix = "(distant memory) "
		}
		result += fmt.Sprintf("%d. [%s] %s%s\n", i+1, time, prefix, e.Text)
	}
	return result
}

func (m *InMemoryStore) Clear(npcID string) {
	m.mu.Lock()
	m.stores[npcID] = nil
	m.mu.Unlock()

	if m.db != nil {
		go m.db.ClearMemories(npcID)
	}
}

// ByCategory returns all memories of a given category for the entity.
func (m *InMemoryStore) ByCategory(npcID string, category string) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []Entry
	for _, e := range m.stores[npcID] {
		if e.Category == category {
			result = append(result, e)
		}
	}
	return result
}

// ByTag returns all memories containing the given tag.
func (m *InMemoryStore) ByTag(npcID string, tag string) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tag = strings.ToLower(tag)
	var result []Entry
	for _, e := range m.stores[npcID] {
		for _, t := range e.Tags {
			if strings.ToLower(t) == tag {
				result = append(result, e)
				break
			}
		}
	}
	return result
}

// Search returns memories whose text contains the keyword (case-insensitive).
func (m *InMemoryStore) Search(npcID string, keyword string) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keyword = strings.ToLower(keyword)
	var result []Entry
	for _, e := range m.stores[npcID] {
		if strings.Contains(strings.ToLower(e.Text), keyword) {
			result = append(result, e)
		}
	}
	return result
}

// HighestImportance returns the top N memories by importance score.
func (m *InMemoryStore) HighestImportance(npcID string, count int) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]Entry, len(m.stores[npcID]))
	copy(all, m.stores[npcID])
	sort.Slice(all, func(i, j int) bool {
		return all[i].Importance > all[j].Importance
	})
	if len(all) > count {
		all = all[:count]
	}
	return all
}

// ApplyDecay reduces vividness on all memories for the entity and evicts those below threshold.
func (m *InMemoryStore) ApplyDecay(npcID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entries := m.stores[npcID]
	kept := make([]Entry, 0, len(entries))
	for i := range entries {
		entries[i].Vividness *= (1.0 - entries[i].DecayRate())
		if entries[i].Vividness >= VividThreshold {
			kept = append(kept, entries[i])
		}
	}
	m.stores[npcID] = kept
}

// Persistence methods delegate to Persister.

func (m *InMemoryStore) Save(npcID string, entry Entry) error {
	if m.db == nil {
		return nil
	}
	return m.db.SaveMemory(npcID, entry)
}

func (m *InMemoryStore) LoadAll() (map[string][]Entry, error) {
	if m.db == nil {
		return nil, nil
	}
	return m.db.LoadAllMemories()
}

func (m *InMemoryStore) ClearPersisted(npcID string) error {
	if m.db == nil {
		return nil
	}
	return m.db.ClearMemories(npcID)
}

// FormatBlendedForLLM returns a blended memory prompt for NPC decision-making:
// - Last 3 memories (immediate context)
// - Top 2 most important memories (defining experiences)
// - Excludes duplicates
func (m *InMemoryStore) FormatBlendedForLLM(npcID string) string {
	recent := m.Recent(npcID, 3)
	important := m.HighestImportance(npcID, 2)

	seen := make(map[string]bool)
	var all []Entry
	for _, e := range recent {
		if !seen[e.Text] {
			seen[e.Text] = true
			all = append(all, e)
		}
	}
	for _, e := range important {
		if !seen[e.Text] {
			seen[e.Text] = true
			all = append(all, e)
		}
	}

	if len(all) == 0 {
		return "No memories yet."
	}

	var sb strings.Builder
	sb.WriteString("Recent:\n")
	for i, e := range recent {
		time := e.Time
		if time == "" {
			time = "?"
		}
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, time, e.Text))
	}

	// Add defining moments if any important memories weren't in recent
	var defining []Entry
	for _, e := range important {
		found := false
		for _, r := range recent {
			if r.Text == e.Text {
				found = true
				break
			}
		}
		if !found {
			defining = append(defining, e)
		}
	}
	if len(defining) > 0 {
		sb.WriteString("\nDefining moments:\n")
		for _, e := range defining {
			vivid := ""
			if e.Vividness > 0.7 {
				vivid = "(vivid) "
			}
			sb.WriteString(fmt.Sprintf("- [%s] %s%s\n", e.Time, vivid, e.Text))
		}
	}

	return sb.String()
}
