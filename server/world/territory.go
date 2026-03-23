package world

import (
	"fmt"
	"math"
)

// Territory represents a major region of the kingdom.
// There are 6 territories: 1 royal domain (center) and 5 duchies.
type Territory struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`    // "royal_domain" or "duchy"
	CenterX int    `json:"centerX"` // Grid center of this territory
	CenterY int    `json:"centerY"`
	Radius  int    `json:"radius"` // Approximate control radius in tiles

	RulerID   string `json:"rulerId"`   // NPC ID of king/duke
	RulerName string `json:"rulerName"` // Display name
	BiomeHint string `json:"biomeHint"` // Dominant biome: desert, swamp, forest, tundra, plains, mountain

	CityIDs []string `json:"cityIds"` // Location IDs of city centers within this territory

	TaxRate  float64  `json:"taxRate"`
	Laws     []string `json:"laws"`
	Treasury int      `json:"treasury"`

	// Inter-territory relationships
	Allies  []string `json:"allies,omitempty"`  // Territory IDs
	Enemies []string `json:"enemies,omitempty"` // Territory IDs
}

// ValidBiomeHints lists all allowed biome hint values for territories.
var ValidBiomeHints = []string{
	"desert", "swamp", "forest", "tundra", "plains", "mountain",
}

const TerritoryTypeRoyalDomain = "royal_domain"
const TerritoryTypeDuchy = "duchy"
const DefaultTerritoryCount = 6

// PlaceTerritoryCenters distributes territory centers on a hex-ish pattern.
// Center territory at grid center; 5 outer territories evenly spaced in a ring.
func PlaceTerritoryCenters(gridW, gridH, count int) []Territory {
	centerX := gridW / 2
	centerY := gridH / 2

	// Radius from center to outer territory centers (~40% of half-grid)
	ringRadius := min(gridW, gridH) * 2 / 5

	territories := make([]Territory, 0, count)

	// Center territory (royal domain)
	territories = append(territories, Territory{
		ID:      "territory_1",
		Type:    TerritoryTypeRoyalDomain,
		CenterX: centerX,
		CenterY: centerY,
		Radius:  ringRadius / 2,
		TaxRate: 0.10,
		CityIDs: make([]string, 0),
		Laws:    make([]string, 0),
	})

	// Outer territories (duchies) in a ring
	outerCount := count - 1
	for i := 0; i < outerCount; i++ {
		angle := float64(i) * (2 * math.Pi / float64(outerCount))
		cx := centerX + int(float64(ringRadius)*math.Cos(angle))
		cy := centerY + int(float64(ringRadius)*math.Sin(angle))

		// Clamp to grid bounds with margin
		margin := 10
		cx = clampInt(cx, margin, gridW-margin)
		cy = clampInt(cy, margin, gridH-margin)

		territories = append(territories, Territory{
			ID:      fmt.Sprintf("territory_%d", i+2),
			Type:    TerritoryTypeDuchy,
			CenterX: cx,
			CenterY: cy,
			Radius:  ringRadius / 2,
			TaxRate: 0.10,
			CityIDs: make([]string, 0),
			Laws:    make([]string, 0),
		})
	}

	return territories
}

// ClosestTerritory returns the territory whose center is closest to (x, y).
func ClosestTerritory(territories []*Territory, x, y int) *Territory {
	if len(territories) == 0 {
		return nil
	}
	best := territories[0]
	bestDist := manhattanDist(x, y, best.CenterX, best.CenterY)
	for _, t := range territories[1:] {
		d := manhattanDist(x, y, t.CenterX, t.CenterY)
		if d < bestDist {
			bestDist = d
			best = t
		}
	}
	return best
}

func manhattanDist(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	dy := y1 - y2
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
