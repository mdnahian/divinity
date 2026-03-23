package graph

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/divinity/core/memory"
	"github.com/divinity/core/world"
)

// NodeMeta holds metadata about a node in the social graph.
type NodeMeta struct {
	Name        string
	Profession  string
	NobleRank   string
	TerritoryID string
	FactionID   string
	IsAlive     bool
}

// TensionEdge represents a hostile relationship between two NPCs.
type TensionEdge struct {
	SourceID   string
	SourceName string
	TargetID   string
	TargetName string
	Sentiment  float64
}

// GraphMetrics holds pre-computed daily graph analytics.
type GraphMetrics struct {
	Centrality   map[string]float64 // weighted degree centrality
	Influence    map[string]float64 // composite influence score
	Clusters     [][]string         // connected components on positive-sentiment subgraph
	Bridges      []string           // articulation points (bridge NPCs)
	TensionEdges []TensionEdge      // sentiment < -0.2, sorted by severity
}

// SocialGraph is an in-memory social graph derived from relationship data.
// It is rebuilt daily from RelationshipStore + World state.
type SocialGraph struct {
	mu        sync.RWMutex
	adjacency map[string]map[string]float64 // npcID → targetID → sentiment
	metadata  map[string]NodeMeta
	metrics   *GraphMetrics
}

func New() *SocialGraph {
	return &SocialGraph{
		adjacency: make(map[string]map[string]float64),
		metadata:  make(map[string]NodeMeta),
	}
}

// Rebuild reconstructs the graph from current world state and relationships.
func (g *SocialGraph) Rebuild(w *world.World, rels *memory.RelationshipStore) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.adjacency = make(map[string]map[string]float64)
	g.metadata = make(map[string]NodeMeta)

	alive := w.AliveNPCs()

	// Populate metadata
	aliveSet := make(map[string]bool)
	for _, n := range alive {
		aliveSet[n.ID] = true
		g.metadata[n.ID] = NodeMeta{
			Name:        n.Name,
			Profession:  n.Profession,
			NobleRank:   n.NobleRank,
			TerritoryID: n.TerritoryID,
			FactionID:   n.FactionID,
			IsAlive:     true,
		}
	}

	// Build adjacency from explicit relationships (|sentiment| > 0.1)
	for _, entry := range rels.AllEntries() {
		if !aliveSet[entry.NpcID] {
			continue
		}
		if math.Abs(entry.Rel.Sentiment) <= 0.1 {
			continue
		}
		g.addEdge(entry.NpcID, entry.Rel.TargetID, entry.Rel.Sentiment)
	}

	// Add implicit edges
	// Faction co-membership: +0.2
	factionMembers := make(map[string][]string)
	for _, n := range alive {
		if n.FactionID != "" {
			factionMembers[n.FactionID] = append(factionMembers[n.FactionID], n.ID)
		}
	}
	for _, members := range factionMembers {
		for i := 0; i < len(members); i++ {
			for j := i + 1; j < len(members); j++ {
				g.addImplicitEdge(members[i], members[j], 0.2)
				g.addImplicitEdge(members[j], members[i], 0.2)
			}
		}
	}

	// Liege-vassal: +0.3
	for _, n := range alive {
		if n.LiegeID != "" && aliveSet[n.LiegeID] {
			g.addImplicitEdge(n.ID, n.LiegeID, 0.3)
			g.addImplicitEdge(n.LiegeID, n.ID, 0.3)
		}
	}

	log.Printf("[Graph] Rebuilt social graph: %d nodes, %d edges", len(g.metadata), g.edgeCount())
}

func (g *SocialGraph) addEdge(from, to string, sentiment float64) {
	if g.adjacency[from] == nil {
		g.adjacency[from] = make(map[string]float64)
	}
	g.adjacency[from][to] = sentiment
}

// addImplicitEdge adds sentiment only if no explicit edge exists already.
func (g *SocialGraph) addImplicitEdge(from, to string, sentiment float64) {
	if g.adjacency[from] == nil {
		g.adjacency[from] = make(map[string]float64)
	}
	if _, exists := g.adjacency[from][to]; !exists {
		g.adjacency[from][to] = sentiment
	}
}

func (g *SocialGraph) edgeCount() int {
	count := 0
	for _, targets := range g.adjacency {
		count += len(targets)
	}
	return count
}

// ComputeMetrics calculates centrality, influence, clusters, bridges, and tension edges.
func (g *SocialGraph) ComputeMetrics() {
	g.mu.Lock()
	defer g.mu.Unlock()

	m := &GraphMetrics{
		Centrality: make(map[string]float64),
		Influence:  make(map[string]float64),
	}

	// Weighted degree centrality
	for nodeID, targets := range g.adjacency {
		var total float64
		for _, sent := range targets {
			total += math.Abs(sent)
		}
		m.Centrality[nodeID] = total
	}

	// Influence score: centrality*0.3 + nobleRank*0.3 + factionLeader*0.2 + reputation-based*0.2
	nobleRankScore := map[string]float64{
		"king": 1.0, "queen": 1.0,
		"prince": 0.8, "princess": 0.8,
		"duke": 0.7, "count": 0.5,
		"baron": 0.3, "knight": 0.2,
	}

	// Normalize centrality for influence calc
	maxCentrality := 0.0
	for _, c := range m.Centrality {
		if c > maxCentrality {
			maxCentrality = c
		}
	}

	for nodeID, meta := range g.metadata {
		normCentrality := 0.0
		if maxCentrality > 0 {
			normCentrality = m.Centrality[nodeID] / maxCentrality
		}

		rankScore := nobleRankScore[meta.NobleRank]

		// Connection count as proxy for reputation component
		connCount := float64(len(g.adjacency[nodeID]))
		normConn := 0.0
		if len(g.metadata) > 1 {
			normConn = connCount / float64(len(g.metadata)-1)
		}

		m.Influence[nodeID] = normCentrality*0.3 + rankScore*0.3 + normConn*0.2 + normCentrality*0.2
	}

	// Connected components on positive-sentiment subgraph (BFS)
	visited := make(map[string]bool)
	for nodeID := range g.metadata {
		if visited[nodeID] {
			continue
		}
		cluster := g.bfsPositive(nodeID, visited)
		if len(cluster) > 1 {
			m.Clusters = append(m.Clusters, cluster)
		}
	}
	// Sort clusters by size descending
	sort.Slice(m.Clusters, func(i, j int) bool {
		return len(m.Clusters[i]) > len(m.Clusters[j])
	})

	// Articulation points (bridges)
	m.Bridges = g.findArticulationPoints()

	// Tension edges (sentiment < -0.2)
	for srcID, targets := range g.adjacency {
		srcMeta := g.metadata[srcID]
		for tgtID, sent := range targets {
			if sent < -0.2 {
				tgtMeta := g.metadata[tgtID]
				m.TensionEdges = append(m.TensionEdges, TensionEdge{
					SourceID:   srcID,
					SourceName: srcMeta.Name,
					TargetID:   tgtID,
					TargetName: tgtMeta.Name,
					Sentiment:  sent,
				})
			}
		}
	}
	// Sort tension edges by severity (most negative first)
	sort.Slice(m.TensionEdges, func(i, j int) bool {
		return m.TensionEdges[i].Sentiment < m.TensionEdges[j].Sentiment
	})

	g.metrics = m
	log.Printf("[Graph] Metrics computed: %d clusters, %d bridges, %d tension edges",
		len(m.Clusters), len(m.Bridges), len(m.TensionEdges))
}

// bfsPositive finds the connected component containing startID on positive edges.
func (g *SocialGraph) bfsPositive(startID string, visited map[string]bool) []string {
	queue := []string{startID}
	visited[startID] = true
	var component []string

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		component = append(component, current)

		for targetID, sent := range g.adjacency[current] {
			if !visited[targetID] && sent > 0 && g.metadata[targetID].IsAlive {
				visited[targetID] = true
				queue = append(queue, targetID)
			}
		}
	}
	return component
}

// findArticulationPoints identifies bridge NPCs using Tarjan's algorithm.
func (g *SocialGraph) findArticulationPoints() []string {
	// Build undirected positive adjacency
	adj := make(map[string][]string)
	for src, targets := range g.adjacency {
		for tgt, sent := range targets {
			if sent > 0 {
				adj[src] = append(adj[src], tgt)
			}
		}
	}

	disc := make(map[string]int)
	low := make(map[string]int)
	parent := make(map[string]string)
	ap := make(map[string]bool)
	timer := 0

	var dfs func(u string)
	dfs = func(u string) {
		children := 0
		disc[u] = timer
		low[u] = timer
		timer++

		for _, v := range adj[u] {
			if _, found := disc[v]; !found {
				children++
				parent[v] = u
				dfs(v)
				if low[v] < low[u] {
					low[u] = low[v]
				}
				// u is an articulation point if:
				// 1. u is root and has 2+ children
				if parent[u] == "" && children > 1 {
					ap[u] = true
				}
				// 2. u is not root and low[v] >= disc[u]
				if parent[u] != "" && low[v] >= disc[u] {
					ap[u] = true
				}
			} else if v != parent[u] {
				if disc[v] < low[u] {
					low[u] = disc[v]
				}
			}
		}
	}

	for nodeID := range g.metadata {
		if _, found := disc[nodeID]; !found {
			parent[nodeID] = ""
			dfs(nodeID)
		}
	}

	var result []string
	for id := range ap {
		result = append(result, id)
	}
	return result
}

// Metrics returns the current pre-computed metrics (read-safe).
func (g *SocialGraph) Metrics() *GraphMetrics {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.metrics
}

// TopInfluential returns the top N NPCs by influence score, optionally filtered by territory.
func (g *SocialGraph) TopInfluential(n int, territoryID string) []InfluenceEntry {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.metrics == nil {
		return nil
	}

	var entries []InfluenceEntry
	for nodeID, score := range g.metrics.Influence {
		meta := g.metadata[nodeID]
		if territoryID != "" && meta.TerritoryID != territoryID {
			continue
		}
		entries = append(entries, InfluenceEntry{
			ID:         nodeID,
			Name:       meta.Name,
			Profession: meta.Profession,
			NobleRank:  meta.NobleRank,
			FactionID:  meta.FactionID,
			Score:      score,
			Centrality: g.metrics.Centrality[nodeID],
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})

	if len(entries) > n {
		entries = entries[:n]
	}
	return entries
}

// InfluenceEntry is a single entry in the influence ranking.
type InfluenceEntry struct {
	ID         string
	Name       string
	Profession string
	NobleRank  string
	FactionID  string
	Score      float64
	Centrality float64
}

// FormatInfluenceMap formats the top influential NPCs for the GOD AI.
func (g *SocialGraph) FormatInfluenceMap(territoryID string) string {
	entries := g.TopInfluential(10, territoryID)
	if len(entries) == 0 {
		return "No influence data available."
	}

	var sb strings.Builder
	sb.WriteString("=== Influence Map ===\n")
	for i, e := range entries {
		rank := ""
		if e.NobleRank != "" {
			rank = fmt.Sprintf(" [%s]", e.NobleRank)
		}
		faction := ""
		if e.FactionID != "" {
			faction = fmt.Sprintf(" faction:%s", e.FactionID)
		}
		sb.WriteString(fmt.Sprintf("%d. %s (%s)%s%s — influence: %.2f, centrality: %.2f\n",
			i+1, e.Name, e.Profession, rank, faction, e.Score, e.Centrality))
	}
	return sb.String()
}

// FormatSocialClusters formats community clusters for the GOD AI.
func (g *SocialGraph) FormatSocialClusters(territoryID string) string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.metrics == nil || len(g.metrics.Clusters) == 0 {
		return "No social clusters detected."
	}

	var sb strings.Builder
	sb.WriteString("=== Social Clusters ===\n")

	for i, cluster := range g.metrics.Clusters {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("... and %d more clusters\n", len(g.metrics.Clusters)-10))
			break
		}

		// Filter by territory if specified
		var members []string
		factionCounts := make(map[string]int)
		for _, id := range cluster {
			meta := g.metadata[id]
			if territoryID != "" && meta.TerritoryID != territoryID {
				continue
			}
			members = append(members, meta.Name)
			if meta.FactionID != "" {
				factionCounts[meta.FactionID]++
			}
		}
		if len(members) < 2 {
			continue
		}

		sb.WriteString(fmt.Sprintf("\nCluster %d (%d members): %s\n", i+1, len(members), strings.Join(members, ", ")))
		if len(factionCounts) > 0 {
			var parts []string
			for fID, count := range factionCounts {
				parts = append(parts, fmt.Sprintf("%s: %d", fID, count))
			}
			sb.WriteString(fmt.Sprintf("  Faction overlap: %s\n", strings.Join(parts, ", ")))
		}
	}

	// Bridge NPCs
	if len(g.metrics.Bridges) > 0 {
		var bridgeNames []string
		for _, id := range g.metrics.Bridges {
			meta := g.metadata[id]
			if territoryID != "" && meta.TerritoryID != territoryID {
				continue
			}
			bridgeNames = append(bridgeNames, fmt.Sprintf("%s (%s)", meta.Name, meta.Profession))
		}
		if len(bridgeNames) > 0 {
			sb.WriteString(fmt.Sprintf("\nBridge NPCs (connect separate groups): %s\n", strings.Join(bridgeNames, ", ")))
		}
	}

	return sb.String()
}

// FormatTensionAnalysis formats hostile relationships for the GOD AI.
func (g *SocialGraph) FormatTensionAnalysis(territoryID string) string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.metrics == nil || len(g.metrics.TensionEdges) == 0 {
		return "No significant tensions detected."
	}

	var sb strings.Builder
	sb.WriteString("=== Tension Analysis ===\n")

	count := 0
	for _, te := range g.metrics.TensionEdges {
		if territoryID != "" {
			srcMeta := g.metadata[te.SourceID]
			if srcMeta.TerritoryID != territoryID {
				continue
			}
		}
		count++
		if count > 15 {
			sb.WriteString("... (more tensions exist)\n")
			break
		}

		srcMeta := g.metadata[te.SourceID]
		tgtMeta := g.metadata[te.TargetID]

		crossFaction := ""
		if srcMeta.FactionID != "" && tgtMeta.FactionID != "" && srcMeta.FactionID != tgtMeta.FactionID {
			crossFaction = " [CROSS-FACTION]"
		}

		sb.WriteString(fmt.Sprintf("- %s → %s: sentiment %.2f%s\n",
			te.SourceName, te.TargetName, te.Sentiment, crossFaction))
	}

	if count == 0 {
		return "No significant tensions in this territory."
	}

	return sb.String()
}

// NodeCount returns the number of nodes in the graph.
func (g *SocialGraph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.metadata)
}

// EdgeCount returns the number of edges in the graph.
func (g *SocialGraph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.edgeCount()
}
