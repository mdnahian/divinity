package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/divinity/core/config"
	"github.com/divinity/core/db"
	"github.com/divinity/core/graph"
	"github.com/divinity/core/llm"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/world"
)

type Engine struct {
	World         *world.World
	Config        *config.Config
	Router        *llm.Router
	Memory        memory.Store
	Relationships *memory.RelationshipStore
	SharedMemory  *memory.SharedMemoryStore
	Trends        *TrendTracker
	DB            *db.Client
	Tokens           *TokenRegistry
	TerritoryBriefs  map[string]*TerritoryBrief
	SocialGraph      *graph.SocialGraph
	NPCTiers         map[string]NPCTier // NPC ID → tier classification (refreshed daily)

	AfterTick       func()
	OnActionComplete func(npcID, actionID, result string)

	mu                 sync.RWMutex
	running            bool
	parentCtx          context.Context
	cancel             context.CancelFunc
	tickNum            int
	lastGodSlot        int
	lastEnemySpawnTick int

	focusTerritoryIdx int

	dirtyNPCs      map[string]bool // NPC IDs changed since last save
	dirtyLocations map[string]bool // Location IDs changed since last save

	npcMu   sync.Mutex            // guards npcLocks map
	npcLocks map[string]*sync.Mutex // per-NPC lock for agent submissions
}

func New(w *world.World, cfg *config.Config, router *llm.Router, mem memory.Store, dbClient *db.Client) *Engine {
	return &Engine{
		World:         w,
		Config:        cfg,
		Router:        router,
		Memory:        mem,
		Relationships: memory.NewRelationshipStore(),
		SharedMemory:  memory.NewSharedMemoryStore(),
		Trends:        NewTrendTracker(),
		DB:            dbClient,
		Tokens:        NewTokenRegistry(),
		SocialGraph:    graph.New(),
		NPCTiers:       make(map[string]NPCTier),
		lastGodSlot:    -1,
		dirtyNPCs:      make(map[string]bool),
		dirtyLocations: make(map[string]bool),
		npcLocks:       make(map[string]*sync.Mutex),
	}
}

// MarkNPCDirty marks an NPC as needing to be saved.
func (e *Engine) MarkNPCDirty(npcID string) {
	e.dirtyNPCs[npcID] = true
}

// MarkLocationDirty marks a location as needing to be saved.
func (e *Engine) MarkLocationDirty(locID string) {
	e.dirtyLocations[locID] = true
}

// MarkAllNPCsDirty marks all alive NPCs as dirty (used for full-save ticks).
func (e *Engine) MarkAllNPCsDirty() {
	for _, n := range e.World.AliveNPCs() {
		e.dirtyNPCs[n.ID] = true
	}
}

// NPCTierLabel returns a human-readable label for the given tier.
func NPCTierLabel(t NPCTier) string {
	switch t {
	case Tier1FullLLM:
		return "Tier 1 (key figure)"
	case Tier2BatchLLM:
		return "Tier 2 (notable)"
	case Tier3RuleBased:
		return "Tier 3 (commoner)"
	default:
		return "unclassified"
	}
}

func (e *Engine) Start(ctx context.Context) {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.parentCtx = ctx
	ctx, e.cancel = context.WithCancel(ctx)
	e.mu.Unlock()

	log.Println("[Engine] Started")
	go e.loop(ctx)
}

func (e *Engine) Resume() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	if e.parentCtx == nil {
		e.mu.Unlock()
		log.Println("[Engine] Cannot resume: engine was never started")
		return
	}
	e.running = true
	ctx, cancel := context.WithCancel(e.parentCtx)
	e.cancel = cancel
	e.mu.Unlock()

	log.Println("[Engine] Resumed")
	go e.loop(ctx)
}

func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running && e.cancel != nil {
		e.cancel()
		e.running = false
		log.Println("[Engine] Paused")
	}
}

func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

func (e *Engine) TickCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.tickNum
}

func (e *Engine) loop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(e.Config.Game.TickIntervalMs) * time.Millisecond)
	defer ticker.Stop()
	defer func() {
		e.mu.Lock()
		e.running = false
		e.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.tick(ctx)
		}
	}
}
