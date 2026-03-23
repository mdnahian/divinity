package memory

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

// Relationship tracks the memory-based sentiment between two NPCs.
type Relationship struct {
	TargetID   string   `json:"targetId" bson:"target_id"`
	TargetName string   `json:"targetName" bson:"target_name"`
	Sentiment  float64  `json:"sentiment" bson:"sentiment"` // -1.0 (hostile) to +1.0 (friendly)
	LastUpdate int64    `json:"lastUpdate" bson:"last_update"`
	History    []string `json:"history" bson:"history"` // brief interaction log
}

// RelationshipStore tracks memory-based relationships for all NPCs.
type RelationshipStore struct {
	mu    sync.RWMutex
	rels  map[string]map[string]*Relationship // npcID -> targetID -> Relationship
}

func NewRelationshipStore() *RelationshipStore {
	return &RelationshipStore{
		rels: make(map[string]map[string]*Relationship),
	}
}

// Update adjusts sentiment for an interaction and records it in history.
func (rs *RelationshipStore) Update(npcID, targetID, targetName string, sentimentDelta float64, interaction string, gameDay int64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.rels[npcID] == nil {
		rs.rels[npcID] = make(map[string]*Relationship)
	}
	rel := rs.rels[npcID][targetID]
	if rel == nil {
		rel = &Relationship{TargetID: targetID, TargetName: targetName}
		rs.rels[npcID][targetID] = rel
	}
	rel.Sentiment = math.Max(-1.0, math.Min(1.0, rel.Sentiment+sentimentDelta))
	rel.LastUpdate = gameDay
	if interaction != "" {
		rel.History = append(rel.History, interaction)
		if len(rel.History) > 5 {
			rel.History = rel.History[len(rel.History)-5:]
		}
	}
	if targetName != "" {
		rel.TargetName = targetName
	}
}

// Get returns the relationship between two NPCs (nil if none).
func (rs *RelationshipStore) Get(npcID, targetID string) *Relationship {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if rs.rels[npcID] == nil {
		return nil
	}
	return rs.rels[npcID][targetID]
}

// AllFor returns all relationships for an NPC.
func (rs *RelationshipStore) AllFor(npcID string) []*Relationship {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	var result []*Relationship
	for _, rel := range rs.rels[npcID] {
		result = append(result, rel)
	}
	return result
}

// DecayAll decays sentiment toward 0 for all relationships that haven't been updated recently.
func (rs *RelationshipStore) DecayAll(currentDay int64, decayRate float64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	for _, targets := range rs.rels {
		for _, rel := range targets {
			daysSince := currentDay - rel.LastUpdate
			if daysSince > 3 {
				if rel.Sentiment > 0 {
					rel.Sentiment = math.Max(0, rel.Sentiment-decayRate*float64(daysSince))
				} else if rel.Sentiment < 0 {
					rel.Sentiment = math.Min(0, rel.Sentiment+decayRate*float64(daysSince))
				}
			}
		}
	}
}

// RelationshipEntry is a flattened relationship for persistence.
type RelationshipEntry struct {
	NpcID string
	Rel   *Relationship
}

// AllEntries returns all relationships as a flat slice for persistence.
func (rs *RelationshipStore) AllEntries() []RelationshipEntry {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	var result []RelationshipEntry
	for npcID, targets := range rs.rels {
		for _, rel := range targets {
			result = append(result, RelationshipEntry{NpcID: npcID, Rel: rel})
		}
	}
	return result
}

// Set inserts a relationship directly (used for DB hydration).
func (rs *RelationshipStore) Set(npcID string, rel *Relationship) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.rels[npcID] == nil {
		rs.rels[npcID] = make(map[string]*Relationship)
	}
	rs.rels[npcID][rel.TargetID] = rel
}

// FormatForLLM returns a formatted string of an NPC's relationships for injection into prompts.
func (rs *RelationshipStore) FormatForLLM(npcID string) string {
	rels := rs.AllFor(npcID)
	if len(rels) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Feelings about others:\n")
	for _, rel := range rels {
		if math.Abs(rel.Sentiment) < 0.05 {
			continue
		}
		label := "neutral"
		switch {
		case rel.Sentiment > 0.5:
			label = "friendly"
		case rel.Sentiment > 0.2:
			label = "positive"
		case rel.Sentiment > 0:
			label = "slight positive"
		case rel.Sentiment < -0.5:
			label = "hostile"
		case rel.Sentiment < -0.2:
			label = "negative"
		case rel.Sentiment < 0:
			label = "slight negative"
		}
		historyStr := ""
		if len(rel.History) > 0 {
			historyStr = " (" + strings.Join(rel.History, ", ") + ")"
		}
		name := rel.TargetName
		if name == "" {
			name = rel.TargetID
		}
		sb.WriteString(fmt.Sprintf("- %s: %s%s\n", name, label, historyStr))
	}
	return sb.String()
}
