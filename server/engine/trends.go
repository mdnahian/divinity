package engine

import (
	"fmt"
	"strings"
	"sync"

	"github.com/divinity/core/world"
)

// MaxTrendSnapshots is the number of daily snapshots retained for trend analysis.
const MaxTrendSnapshots = 30

// WorldSnapshot captures key metrics for a single game day.
type WorldSnapshot struct {
	GameDay         int    `bson:"game_day"`
	Population      int    `bson:"population"`
	DeadCount       int    `bson:"dead_count"`
	TotalGold       int    `bson:"total_gold"`
	Treasury        int    `bson:"treasury"`
	HungryCount     int    `bson:"hungry_count"`
	StarvingCount   int    `bson:"starving_count"`
	BrokeCount      int    `bson:"broke_count"`
	EnemyCount      int    `bson:"enemy_count"`
	FactionCount    int    `bson:"faction_count"`
	DepletedResLocs int    `bson:"depleted_res_locs"`
	TotalResLocs    int    `bson:"total_res_locs"`
	AvgHappiness    int    `bson:"avg_happiness"`
	AvgStress       int    `bson:"avg_stress"`
	CrisisLevel     string `bson:"crisis_level"`
}

// TrendTracker stores daily world snapshots for GOD trend analysis.
type TrendTracker struct {
	mu        sync.RWMutex
	snapshots []WorldSnapshot
}

// NewTrendTracker creates a new trend tracker.
func NewTrendTracker() *TrendTracker {
	return &TrendTracker{snapshots: make([]WorldSnapshot, 0, MaxTrendSnapshots)}
}

// Snapshots returns a copy of all snapshots (for persistence).
func (t *TrendTracker) Snapshots() []WorldSnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]WorldSnapshot, len(t.snapshots))
	copy(out, t.snapshots)
	return out
}

// SetSnapshots replaces all snapshots (used for DB hydration).
func (t *TrendTracker) SetSnapshots(snaps []WorldSnapshot) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.snapshots = snaps
	if len(t.snapshots) > MaxTrendSnapshots {
		t.snapshots = t.snapshots[len(t.snapshots)-MaxTrendSnapshots:]
	}
}

// Snapshot captures the current world state as a trend data point.
func (t *TrendTracker) Snapshot(w *world.World) {
	alive := w.AliveNPCs()
	deadCount := len(w.NPCs) - len(alive)

	totalGold := 0
	hungryCount := 0
	starvingCount := 0
	brokeCount := 0
	totalHappiness := 0
	totalStress := 0
	for _, n := range alive {
		gold := n.GoldCount()
		totalGold += gold
		if n.Needs.Hunger < 30 {
			hungryCount++
		}
		if n.Needs.Hunger < 10 {
			starvingCount++
		}
		if gold == 0 {
			brokeCount++
		}
		totalHappiness += n.Happiness
		totalStress += n.Stress
	}

	avgHappiness := 0
	avgStress := 0
	if len(alive) > 0 {
		avgHappiness = totalHappiness / len(alive)
		avgStress = totalStress / len(alive)
	}

	depletedLocs := 0
	totalResLocs := 0
	for _, l := range w.Locations {
		if l.Resources == nil || len(l.Resources) == 0 {
			continue
		}
		totalResLocs++
		for _, v := range l.Resources {
			if v == 0 {
				depletedLocs++
				break
			}
		}
	}

	crisisLevel := "STABLE"
	if len(alive) > 0 {
		hungryPct := hungryCount * 100 / len(alive)
		brokePct := brokeCount * 100 / len(alive)
		if hungryPct >= 50 || brokePct >= 50 {
			crisisLevel = "CRITICAL"
		} else if hungryPct >= 25 || brokePct >= 25 {
			crisisLevel = "CONCERNING"
		}
	}

	snap := WorldSnapshot{
		GameDay:         w.GameDay,
		Population:      len(alive),
		DeadCount:       deadCount,
		TotalGold:       totalGold,
		Treasury:        w.Treasury,
		HungryCount:     hungryCount,
		StarvingCount:   starvingCount,
		BrokeCount:      brokeCount,
		EnemyCount:      len(w.AliveEnemies()),
		FactionCount:    len(w.Factions),
		DepletedResLocs: depletedLocs,
		TotalResLocs:    totalResLocs,
		AvgHappiness:    avgHappiness,
		AvgStress:       avgStress,
		CrisisLevel:     crisisLevel,
	}

	t.mu.Lock()
	t.snapshots = append(t.snapshots, snap)
	if len(t.snapshots) > MaxTrendSnapshots {
		t.snapshots = t.snapshots[len(t.snapshots)-MaxTrendSnapshots:]
	}
	t.mu.Unlock()
}

// FormatForGod produces a trend summary for the GOD prompt.
// Shows the trajectory of key metrics over recent days.
func (t *TrendTracker) FormatForGod() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.snapshots) < 2 {
		return "Insufficient data for trend analysis (need 2+ days)."
	}

	// Compare current to the snapshot from ~5 days ago (or earliest available)
	current := t.snapshots[len(t.snapshots)-1]
	compareIdx := len(t.snapshots) - 6
	if compareIdx < 0 {
		compareIdx = 0
	}
	past := t.snapshots[compareIdx]
	daySpan := current.GameDay - past.GameDay
	if daySpan == 0 {
		daySpan = 1
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Trends over last %d days (Day %d → Day %d):", daySpan, past.GameDay, current.GameDay))
	lines = append(lines, fmt.Sprintf("  Population: %d → %d (%s)", past.Population, current.Population, trendArrow(current.Population-past.Population)))
	lines = append(lines, fmt.Sprintf("  Deaths (total): %d → %d", past.DeadCount, current.DeadCount))
	lines = append(lines, fmt.Sprintf("  Gold in circulation: %d → %d (%s)", past.TotalGold, current.TotalGold, trendArrow(current.TotalGold-past.TotalGold)))
	lines = append(lines, fmt.Sprintf("  Treasury: %d → %d (%s)", past.Treasury, current.Treasury, trendArrow(current.Treasury-past.Treasury)))
	lines = append(lines, fmt.Sprintf("  Hungry NPCs: %d → %d (%s)", past.HungryCount, current.HungryCount, trendArrowInverse(current.HungryCount-past.HungryCount)))
	lines = append(lines, fmt.Sprintf("  Broke NPCs: %d → %d (%s)", past.BrokeCount, current.BrokeCount, trendArrowInverse(current.BrokeCount-past.BrokeCount)))
	lines = append(lines, fmt.Sprintf("  Enemies: %d → %d", past.EnemyCount, current.EnemyCount))
	lines = append(lines, fmt.Sprintf("  Avg happiness: %d → %d (%s)", past.AvgHappiness, current.AvgHappiness, trendArrow(current.AvgHappiness-past.AvgHappiness)))
	lines = append(lines, fmt.Sprintf("  Avg stress: %d → %d (%s)", past.AvgStress, current.AvgStress, trendArrowInverse(current.AvgStress-past.AvgStress)))
	lines = append(lines, fmt.Sprintf("  Depleted resource locations: %d/%d → %d/%d", past.DepletedResLocs, past.TotalResLocs, current.DepletedResLocs, current.TotalResLocs))
	lines = append(lines, fmt.Sprintf("  Crisis level: %s → %s", past.CrisisLevel, current.CrisisLevel))

	return strings.Join(lines, "\n")
}

// trendArrow returns a direction indicator. Positive = improving.
func trendArrow(delta int) string {
	switch {
	case delta > 0:
		return "UP"
	case delta < 0:
		return "DOWN"
	default:
		return "FLAT"
	}
}

// trendArrowInverse returns a direction indicator where lower = better.
func trendArrowInverse(delta int) string {
	switch {
	case delta > 0:
		return "WORSE"
	case delta < 0:
		return "BETTER"
	default:
		return "FLAT"
	}
}
