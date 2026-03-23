package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

// TerritoryBrief is a pre-computed daily intelligence summary for a territory,
// designed to give the GOD AI immediate context without wasting tool calls.
type TerritoryBrief struct {
	TerritoryID   string
	TerritoryName string

	Population    int
	CrisisNPCs    int
	EnemyPresence int

	MoodDistribution map[string]int // mood label → count
	TopGrievances    []string       // top 3 negative themes from NPC memories
	HostilePairs     []string       // "A hostile toward B" for sentiment < -0.3
	StrongAlliances  []string       // "A allied with B" for sentiment > 0.5
	BrokePercent     int
	RecentDeaths     []string // name + cause from event log
	ResourceSummary  string   // depleted/healthy resource locations
}

// FormatForGod returns a concise text summary suitable for injection into the GOD identity prompt.
func (b *TerritoryBrief) FormatForGod() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("TERRITORY BRIEF: %s (%s)\n", b.TerritoryName, b.TerritoryID))
	sb.WriteString(fmt.Sprintf("Population: %d | Crisis NPCs: %d | Enemies present: %d | Broke: %d%%\n",
		b.Population, b.CrisisNPCs, b.EnemyPresence, b.BrokePercent))

	if len(b.MoodDistribution) > 0 {
		var parts []string
		for mood, count := range b.MoodDistribution {
			parts = append(parts, fmt.Sprintf("%s: %d", mood, count))
		}
		sort.Strings(parts)
		sb.WriteString(fmt.Sprintf("Moods: %s\n", strings.Join(parts, ", ")))
	}

	if len(b.TopGrievances) > 0 {
		sb.WriteString(fmt.Sprintf("Top grievances: %s\n", strings.Join(b.TopGrievances, "; ")))
	}
	if len(b.HostilePairs) > 0 {
		limit := 5
		if len(b.HostilePairs) < limit {
			limit = len(b.HostilePairs)
		}
		sb.WriteString(fmt.Sprintf("Hostile relationships: %s\n", strings.Join(b.HostilePairs[:limit], "; ")))
	}
	if len(b.StrongAlliances) > 0 {
		limit := 5
		if len(b.StrongAlliances) < limit {
			limit = len(b.StrongAlliances)
		}
		sb.WriteString(fmt.Sprintf("Strong alliances: %s\n", strings.Join(b.StrongAlliances[:limit], "; ")))
	}
	if len(b.RecentDeaths) > 0 {
		sb.WriteString(fmt.Sprintf("Recent deaths: %s\n", strings.Join(b.RecentDeaths, "; ")))
	}
	if b.ResourceSummary != "" {
		sb.WriteString(fmt.Sprintf("Resources: %s\n", b.ResourceSummary))
	}

	return sb.String()
}

// computeTerritoryBriefs builds intelligence briefs for all territories.
func (e *Engine) computeTerritoryBriefs() map[string]*TerritoryBrief {
	w := e.World
	briefs := make(map[string]*TerritoryBrief)

	for _, t := range w.Territories {
		briefs[t.ID] = &TerritoryBrief{
			TerritoryID:      t.ID,
			TerritoryName:    t.Name,
			MoodDistribution: make(map[string]int),
		}
	}

	alive := w.AliveNPCs()

	// Gather NPC stats per territory
	for _, n := range alive {
		b, ok := briefs[n.TerritoryID]
		if !ok {
			continue
		}
		b.Population++
		b.MoodDistribution[n.Mood()]++

		if n.HP < 20 || n.Needs.Hunger < 10 {
			b.CrisisNPCs++
		}
		if n.GoldCount() == 0 {
			b.BrokePercent++ // will convert to percentage below
		}
	}

	// Convert broke count to percentage
	for _, b := range briefs {
		if b.Population > 0 {
			b.BrokePercent = b.BrokePercent * 100 / b.Population
		}
	}

	// Enemy presence per territory
	terrLocIDs := buildTerritoryLocationMap(w)
	for _, enemy := range w.AliveEnemies() {
		if tid, ok := terrLocIDs[enemy.LocationID]; ok {
			if b, ok := briefs[tid]; ok {
				b.EnemyPresence++
			}
		}
	}

	// Scan last 100 events for deaths
	start := len(w.EventLog) - 100
	if start < 0 {
		start = 0
	}
	for i := start; i < len(w.EventLog); i++ {
		ev := w.EventLog[i]
		if ev.Type != "death" {
			continue
		}
		text := ev.Text
		if len(text) > 80 {
			text = text[:80]
		}
		if ev.NpcID != "" {
			for _, n := range w.NPCs {
				if n.ID == ev.NpcID {
					if b, ok := briefs[n.TerritoryID]; ok {
						b.RecentDeaths = append(b.RecentDeaths, text)
					}
					break
				}
			}
		}
	}

	// Grievances: scan recent memories for negative keywords
	grievanceKeywords := []string{"starving", "hungry", "attacked", "died", "unpaid", "robbed", "homeless", "sick", "wounded"}
	for _, n := range alive {
		b, ok := briefs[n.TerritoryID]
		if !ok {
			continue
		}
		recent := e.Memory.Recent(n.ID, 3)
		for _, mem := range recent {
			lower := strings.ToLower(mem.Text)
			for _, kw := range grievanceKeywords {
				if strings.Contains(lower, kw) {
					b.TopGrievances = append(b.TopGrievances, kw)
					break // one keyword per memory
				}
			}
		}
	}

	// Deduplicate and limit grievances to top 3 per territory
	for _, b := range briefs {
		b.TopGrievances = topNByFrequency(b.TopGrievances, 3)
	}

	// Relationship extremes per territory
	e.computeRelationshipExtremes(briefs, alive)

	// Resource summary per territory
	computeResourceSummaries(briefs, w)

	return briefs
}

// computeRelationshipExtremes populates HostilePairs and StrongAlliances on briefs.
func (e *Engine) computeRelationshipExtremes(briefs map[string]*TerritoryBrief, alive []*npc.NPC) {
	npcNames := make(map[string]string)
	npcTerr := make(map[string]string)
	for _, n := range alive {
		npcNames[n.ID] = n.Name
		npcTerr[n.ID] = n.TerritoryID
	}

	for _, entry := range e.Relationships.AllEntries() {
		tid := npcTerr[entry.NpcID]
		b, ok := briefs[tid]
		if !ok {
			continue
		}
		srcName := npcNames[entry.NpcID]
		if srcName == "" {
			continue
		}

		if entry.Rel.Sentiment < -0.3 {
			b.HostilePairs = append(b.HostilePairs,
				fmt.Sprintf("%s hostile toward %s (%.1f)", srcName, entry.Rel.TargetName, entry.Rel.Sentiment))
		}
		if entry.Rel.Sentiment > 0.5 {
			b.StrongAlliances = append(b.StrongAlliances,
				fmt.Sprintf("%s allied with %s (%.1f)", srcName, entry.Rel.TargetName, entry.Rel.Sentiment))
		}
	}
}

func computeResourceSummaries(briefs map[string]*TerritoryBrief, w *world.World) {
	for _, loc := range w.Locations {
		b, ok := briefs[loc.TerritoryID]
		if !ok || loc.Resources == nil {
			continue
		}

		depleted := 0
		total := 0
		for res, qty := range loc.Resources {
			total++
			maxQty := 0
			if loc.MaxResources != nil {
				maxQty = loc.MaxResources[res]
			}
			if maxQty > 0 && qty*100/maxQty < 20 {
				depleted++
			}
		}

		if total > 0 {
			if b.ResourceSummary != "" {
				b.ResourceSummary += "; "
			}
			if depleted > 0 {
				b.ResourceSummary += fmt.Sprintf("%s: %d/%d depleted", loc.Name, depleted, total)
			} else {
				b.ResourceSummary += fmt.Sprintf("%s: healthy", loc.Name)
			}
		}
	}
}

// buildTerritoryLocationMap maps location ID → territory ID.
func buildTerritoryLocationMap(w *world.World) map[string]string {
	m := make(map[string]string)
	for _, loc := range w.Locations {
		if loc.TerritoryID != "" {
			m[loc.ID] = loc.TerritoryID
		}
	}
	return m
}

// topNByFrequency counts occurrences and returns the top N most frequent strings.
func topNByFrequency(items []string, n int) []string {
	if len(items) == 0 {
		return nil
	}
	counts := make(map[string]int)
	for _, s := range items {
		counts[s]++
	}
	type kv struct {
		key   string
		count int
	}
	var sorted []kv
	for k, v := range counts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	result := make([]string, len(sorted))
	for i, kv := range sorted {
		result[i] = kv.key
	}
	return result
}
