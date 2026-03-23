package config

import (
	"os"
	"strconv"
)

type APIConfig struct {
	OpenRouterURL   string
	OpenRouterKey   string
	OpenRouterModel string
	MaxTokens       int
	Temperature     float64
	TimeoutMs       int
}

type GodAgentConfig struct {
	Enabled     bool
	Model       string
	MaxTokens   int
	Temperature float64
}

type GameConfig struct {
	TickIntervalMs     int
	GameMinutesPerTick int
	NPCCount           int
	MaxConcurrentCalls int
	GameDaysPerYear    int
	DeathGraceTicks    int
	GridW              int
	GridH              int
	CityCount          int
	TerritoryCount     int
	LLMBatchSize         int
	GodMaxToolCalls      int
	TravelMinutesPerUnit int
	// NPC AI tiering
	Tier1Count int // Full LLM NPCs (nobles, leaders, crisis)
	Tier2Count int // Batched LLM NPCs
	// Remaining NPCs are Tier 3 (rule-based)
}

type ThresholdConfig struct {
	HungerUrgent   int
	ThirstUrgent   int
	FatigueUrgent  int
	SocialNeedHigh int
	StressHigh     int
}

type DecayConfig struct {
	Hunger     float64
	Thirst     float64
	Fatigue    float64
	SocialNeed float64
	Happiness  float64
	Stress     float64
}

type EconomyConfig struct {
	BasePrices    map[string]int
	MinMultiplier float64
	MaxMultiplier float64
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

type MongoConfig struct {
	URI      string
	Database string
	WorldID  string
}

type Config struct {
	API        APIConfig
	GodAgent   GodAgentConfig
	Game       GameConfig
	Thresholds ThresholdConfig
	Decay      DecayConfig
	Economy    EconomyConfig
	RateLimit  RateLimitConfig
	Mongo      MongoConfig
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func Load() *Config {
	return &Config{
		API: APIConfig{
			OpenRouterURL:   envStr("OPENROUTER_URL", "https://openrouter.ai/api/v1/chat/completions"),
			OpenRouterKey:   envStr("OPENROUTER_KEY", ""),
			OpenRouterModel: envStr("OPENROUTER_MODEL", "qwen/qwen-2.5-7b-instruct"),
			MaxTokens:       envInt("LLM_MAX_TOKENS", 0),
			Temperature:     envFloat("LLM_TEMPERATURE", 0.8),
			TimeoutMs:       envInt("LLM_TIMEOUT_MS", 120000),
		},
		// GOD model: runs ~7 times/day + 1 chronicle/day, so cost is minimal.
		// Budget: qwen/qwen3.5-flash (current) — adequate with strategic brief
		// Better: google/gemini-2.0-flash-001 — better reasoning + narratives
		// Best: anthropic/claude-3.5-haiku — excellent at low cost
		GodAgent: GodAgentConfig{
			Enabled:     envBool("GOD_AGENT_ENABLED", true),
			Model:       envStr("GOD_AGENT_MODEL", "qwen/qwen3.5-flash-02-23"),
			MaxTokens:   envInt("GOD_AGENT_MAX_TOKENS", 2048),
			Temperature: envFloat("GOD_AGENT_TEMPERATURE", 0.7),
		},
		Game: GameConfig{
			TickIntervalMs:     envInt("TICK_INTERVAL_MS", 60000),
			GameMinutesPerTick: envInt("GAME_MINUTES_PER_TICK", 5),
			NPCCount:           envInt("NPC_COUNT", 500),
			MaxConcurrentCalls: envInt("MAX_CONCURRENT_CALLS", 50),
			GameDaysPerYear:    envInt("GAME_DAYS_PER_YEAR", 15),
			DeathGraceTicks:    envInt("DEATH_GRACE_TICKS", 5),
			GridW:              envInt("GRID_W", 300),
			GridH:              envInt("GRID_H", 300),
			CityCount:          envInt("CITY_COUNT", 18),
			TerritoryCount:     envInt("TERRITORY_COUNT", 6),
			LLMBatchSize:       envInt("LLM_BATCH_SIZE", 5),
			GodMaxToolCalls:      envInt("GOD_MAX_TOOL_CALLS", 12),
			TravelMinutesPerUnit: envInt("TRAVEL_MINUTES_PER_UNIT", 1),
			Tier1Count:         envInt("TIER1_COUNT", 50),
			Tier2Count:         envInt("TIER2_COUNT", 100),
		},
		Thresholds: ThresholdConfig{
			HungerUrgent:   25,
			ThirstUrgent:   15,
			FatigueUrgent:  85,
			SocialNeedHigh: 70,
			StressHigh:     75,
		},
		Decay: DecayConfig{
			Hunger:     0.069,  // ~5 game days (100→0) at 5 min/tick, 288 ticks/day
			Thirst:     0.116,  // ~3 game days (100→0)
			Fatigue:    0.10,   // ~3.5 game days (0→100); was 0.174 (~2 days) which caused rest-loop dominance
			SocialNeed: 0.5,    // ~0.7 game days (0→100), talk threshold (~30) in ~5 game hours
			Happiness:  0.107,
			Stress:     -0.072,
		},
		Economy: EconomyConfig{
			BasePrices: map[string]int{
				"bread": 2, "raw meat": 1, "cooked meal": 3, "berries": 1, "wheat": 1,
				"herbs": 2, "hide": 2, "ale": 2, "healing potion": 5, "spice bundle": 4,
				"odd trinket": 2, "pretty stone": 1, "old coin": 3, "carved bone": 2,
				"logs": 2, "stone": 3, "iron ore": 4, "iron ingot": 8, "leather": 3,
				"fish": 2, "thatch": 1, "clay": 1, "flour": 2, "ceramic": 4, "cloth": 2, "rope": 2,
				"iron sword": 20, "iron axe": 15, "pickaxe": 15, "hammer": 12,
				"leather armor": 18, "cloth tunic": 6, "belt pouch": 5, "satchel": 8, "backpack": 14,
				"written journal": 1, "book": 3,
				// Horse & carriage items
				"hay": 1, "horse feed": 3, "saddle": 25, "bridle": 15, "horseshoes": 10,
				// Biome resources
				"sand": 1, "cactus_water": 2, "peat": 2, "poison_herbs": 4, "bog_iron": 5,
				"gems": 15, "crystal": 10, "rare_ore": 12, "ice": 1, "fur": 6, "frozen_herbs": 3,
			},
			MinMultiplier: 0.3,
			MaxMultiplier: 3.0,
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: envFloat("RATE_LIMIT_RPS", 5),
			BurstSize:         envInt("RATE_LIMIT_BURST", 10),
		},
		Mongo: MongoConfig{
			URI:      envStr("MONGODB_URI", "mongodb://localhost:27017"),
			Database: envStr("MONGODB_DATABASE", "divinity"),
			WorldID:  envStr("WORLD_ID", "main"),
		},
	}
}

type TokenUsage struct {
	Requests         int
	PromptTokens     int
	CompletionTokens int
	ActualCost       float64
}

func (t *TokenUsage) TotalTokens() int {
	return t.PromptTokens + t.CompletionTokens
}

func (t *TokenUsage) Record(prompt, completion int, cost float64) {
	t.Requests++
	t.PromptTokens += prompt
	t.CompletionTokens += completion
	t.ActualCost += cost
}

func (t *TokenUsage) Reset() {
	t.Requests = 0
	t.PromptTokens = 0
	t.CompletionTokens = 0
	t.ActualCost = 0
}
