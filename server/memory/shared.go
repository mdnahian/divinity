package memory

import (
	"fmt"
	"sync"
)

// LocationMemory stores significant events at a location that any NPC visiting can perceive.
type LocationMemory struct {
	Text      string `json:"text" bson:"text"`
	Time      string `json:"time" bson:"time"`
	GameDay   int64  `json:"gameDay" bson:"game_day"`
	Vividness float64 `json:"vividness" bson:"vividness"`
}

// SharedMemoryStore manages location-based memories and rumor propagation.
type SharedMemoryStore struct {
	mu       sync.RWMutex
	locMems  map[string][]LocationMemory // locationID -> memories
}

func NewSharedMemoryStore() *SharedMemoryStore {
	return &SharedMemoryStore{
		locMems: make(map[string][]LocationMemory),
	}
}

// AddLocationMemory records a significant event at a location.
func (s *SharedMemoryStore) AddLocationMemory(locationID string, text, time string, gameDay int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mems := s.locMems[locationID]
	// Deduplicate: refresh existing identical memory instead of adding
	for i, m := range mems {
		if m.Text == text {
			mems[i].Vividness = 1.0
			mems[i].Time = time
			mems[i].GameDay = gameDay
			s.locMems[locationID] = mems
			return
		}
	}
	mems = append(mems, LocationMemory{
		Text:      text,
		Time:      time,
		GameDay:   gameDay,
		Vividness: 1.0,
	})
	// Keep max 5 location memories
	if len(mems) > 5 {
		mems = mems[len(mems)-5:]
	}
	s.locMems[locationID] = mems
}

// LocationMemories returns recent memories at a location.
func (s *SharedMemoryStore) LocationMemories(locationID string) []LocationMemory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]LocationMemory, len(s.locMems[locationID]))
	copy(out, s.locMems[locationID])
	return out
}

// SharedMemoryEntry is a flattened location memory for persistence.
type SharedMemoryEntry struct {
	LocationID string
	Mem        LocationMemory
}

// AllEntries returns all location memories as a flat slice for persistence.
func (s *SharedMemoryStore) AllEntries() []SharedMemoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []SharedMemoryEntry
	for locID, mems := range s.locMems {
		for _, m := range mems {
			result = append(result, SharedMemoryEntry{LocationID: locID, Mem: m})
		}
	}
	return result
}

// SetLocationMemory inserts a location memory directly (used for DB hydration).
func (s *SharedMemoryStore) SetLocationMemory(locationID string, mem LocationMemory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.locMems[locationID] = append(s.locMems[locationID], mem)
}

// DecayLocationMemories reduces vividness of all location memories.
func (s *SharedMemoryStore) DecayLocationMemories() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for locID, mems := range s.locMems {
		kept := make([]LocationMemory, 0, len(mems))
		for i := range mems {
			mems[i].Vividness *= 0.7
			if mems[i].Vividness >= 0.1 {
				kept = append(kept, mems[i])
			}
		}
		s.locMems[locID] = kept
	}
}

// PropagateRumor creates a "heard" memory in the listener's store when two NPCs interact.
// Returns the rumor entry to be added to the listener's memory store.
func PropagateRumor(sourceMemory Entry, sourceName string) Entry {
	return Entry{
		Text:       fmt.Sprintf("(heard from %s) %s", sourceName, sourceMemory.Text),
		Time:       sourceMemory.Time,
		TS:         sourceMemory.TS,
		Importance: sourceMemory.Importance * 0.5, // reduced importance for secondhand info
		Vividness:  0.6,                            // reduced vividness
		Category:   CatHeard,
		Tags:       sourceMemory.Tags,
	}
}
