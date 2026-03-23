package memory

import "math"

// Embedder generates vector embeddings for text. Implementations can use
// local models, API-based models, or return nil for no-op behavior.
type Embedder interface {
	// Embed returns a float64 vector for the given text.
	// Returns nil if embedding is not available.
	Embed(text string) []float64
}

// NoOpEmbedder is a placeholder that returns nil embeddings.
// Replace with a real implementation when embedding infrastructure is available.
type NoOpEmbedder struct{}

func (NoOpEmbedder) Embed(string) []float64 { return nil }

// CosineSimilarity computes the cosine similarity between two vectors.
// Returns 0 if either vector is nil or they differ in length.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// SemanticSearch finds the top-k most semantically similar memories to the query.
// Falls back to keyword search if embeddings are not available.
func SemanticSearch(store Store, embedder Embedder, npcID string, query string, k int) []Entry {
	if embedder == nil {
		return store.Search(npcID, query)
	}

	queryVec := embedder.Embed(query)
	if queryVec == nil {
		return store.Search(npcID, query)
	}

	all := store.All(npcID)
	type scored struct {
		entry Entry
		score float64
	}
	var results []scored
	for _, e := range all {
		vec := embedder.Embed(e.Text)
		if vec == nil {
			continue
		}
		sim := CosineSimilarity(queryVec, vec)
		results = append(results, scored{entry: e, score: sim})
	}

	// Sort by similarity descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	out := make([]Entry, 0, k)
	for i := 0; i < len(results) && i < k; i++ {
		out = append(out, results[i].entry)
	}
	return out
}
