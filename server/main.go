package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/divinity/core/api"
	"github.com/divinity/core/config"
	"github.com/divinity/core/db"
	"github.com/divinity/core/engine"
	"github.com/divinity/core/god"
	"github.com/divinity/core/llm"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/world"
)

func main() {
	cfg := config.Load()
	log.Printf("Divinity starting (tick=%dms, minutes/tick=%d)",
		cfg.Game.TickIntervalMs, cfg.Game.GameMinutesPerTick)

	if cfg.API.OpenRouterKey == "" {
		log.Println("WARNING: No OPENROUTER_KEY set — GOD agent and dreams will be disabled")
	}

	apiAddr := ":8080"
	if v := os.Getenv("API_ADDR"); v != "" {
		apiAddr = v
	}
	apiServer := api.NewServer(cfg)
	apiServer.Start(apiAddr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := llm.NewRouter(cfg)
	var mem *memory.InMemoryStore

	var dbClient *db.Client
	var w *world.World

	dbClient, err := db.Connect(ctx, &cfg.Mongo)
	if err != nil {
		if os.Getenv("ALLOW_MEMORY_ONLY") == "true" {
			log.Printf("[Main] MongoDB unavailable (%v) — running without persistence (ALLOW_MEMORY_ONLY=true)", err)
			dbClient = nil
		} else {
			log.Fatalf("[Main] MongoDB unavailable (%v) — exiting (set ALLOW_MEMORY_ONLY=true to run without persistence)", err)
		}
	} else {
		if err := dbClient.EnsureIndexes(ctx); err != nil {
			log.Printf("[Main] Index creation warning: %v", err)
		}
	}

	// Initialize memory store — with DB persistence if available
	if dbClient != nil {
		persister := db.NewMemoryPersister(dbClient)
		mem = memory.NewInMemoryStoreWithDB(persister)
		if err := mem.HydrateFromDB(); err != nil {
			log.Printf("[Main] Memory hydration warning: %v", err)
		} else {
			log.Println("[Main] Memory store hydrated from MongoDB")
		}
	} else {
		mem = memory.NewInMemoryStore()
	}

	worldID := cfg.Mongo.WorldID
	if dbClient != nil && dbClient.HasWorld(ctx, worldID) {
		log.Printf("[Main] Loading existing world '%s' from MongoDB...", worldID)
		w, err = dbClient.LoadWorld(ctx, worldID, cfg.Economy.BasePrices, cfg.Economy.MinMultiplier, cfg.Economy.MaxMultiplier)
		if err != nil {
			log.Printf("[Main] Load failed (%v) — creating fresh world", err)
			w = nil
		} else {
			log.Printf("[Main] Loaded world — Day %d, %d NPCs, %d locations",
				w.GameDay, len(w.AllAliveNPCs()), len(w.Locations))

			// Rebuild spawn queue from unclaimed NPCs (queue is not persisted in DB)
			for _, n := range w.NPCs {
				if !n.Claimed && n.Alive {
					w.SpawnQueue = append(w.SpawnQueue, world.SpawnEntry{
						NPCID:      n.ID,
						Name:       n.Name,
						Profession: n.Profession,
					})
				}
			}
			log.Printf("[Main] Rebuilt spawn queue: %d unclaimed NPCs available", len(w.SpawnQueue))
		}
	}

	if w == nil {
		w = world.NewWorld(cfg.Economy.BasePrices, cfg.Economy.MinMultiplier, cfg.Economy.MaxMultiplier)
		w.ID = worldID
		log.Printf("[Main] Running GOD genesis for world '%s'...", worldID)
		var lore string
		var err error
		if cfg.Game.TerritoryCount > 1 {
			log.Printf("[Main] Kingdom mode: %d territories", cfg.Game.TerritoryCount)
			lore, err = god.RunKingdomGenesis(ctx, w, router, cfg, func(phase, detail string, current, total int) {
				apiServer.Genesis.Set(phase, detail, current, total)
			})
		} else {
			lore, err = god.RunGenesis(ctx, w, router, cfg)
		}
		if err != nil {
			log.Fatalf("[Main] Genesis failed: %v", err)
		}
		log.Printf("[Main] Genesis complete — %d NPCs, %d locations", len(w.AllAliveNPCs()), len(w.Locations))
		log.Printf("[Main] Lore: %s", lore)

		if dbClient != nil {
			if err := dbClient.SaveWorld(ctx, w); err != nil {
				log.Printf("[Main] Initial save failed: %v", err)
			}
		}
	}

	if cfg.GodAgent.Enabled && cfg.API.OpenRouterKey != "" {
		log.Printf("[Main] GOD agent ENABLED (model=%s, via OpenRouter)", cfg.GodAgent.Model)
	} else if !cfg.GodAgent.Enabled {
		log.Println("[Main] GOD agent DISABLED (GOD_AGENT_ENABLED=false)")
	} else {
		log.Println("[Main] GOD agent DISABLED (no OPENROUTER_KEY set)")
	}
	log.Printf("[Main] LLM batch size: %d | Grid: %dx%d | Cities: %d | Target NPCs: %d",
		cfg.Game.LLMBatchSize, cfg.Game.GridW, cfg.Game.GridH, cfg.Game.CityCount, cfg.Game.NPCCount)

	eng := engine.New(w, cfg, router, mem, dbClient)

	// Hydrate engine-owned stores from MongoDB
	if dbClient != nil {
		if rs, err := dbClient.LoadRelationships(ctx); err != nil {
			log.Printf("[Main] Relationship hydration warning: %v", err)
		} else {
			eng.Relationships = rs
			log.Println("[Main] Relationships hydrated from MongoDB")
		}

		if sm, err := dbClient.LoadSharedMemories(ctx); err != nil {
			log.Printf("[Main] Shared memory hydration warning: %v", err)
		} else {
			eng.SharedMemory = sm
			log.Println("[Main] Shared memories hydrated from MongoDB")
		}

		if docs, err := dbClient.LoadTrends(ctx); err != nil {
			log.Printf("[Main] Trend hydration warning: %v", err)
		} else if len(docs) > 0 {
			snaps := make([]engine.WorldSnapshot, len(docs))
			for i, d := range docs {
				snaps[i] = engine.WorldSnapshot{
					GameDay:         d.GameDay,
					Population:      d.Population,
					DeadCount:       d.DeadCount,
					TotalGold:       d.TotalGold,
					Treasury:        d.Treasury,
					HungryCount:     d.HungryCount,
					StarvingCount:   d.StarvingCount,
					BrokeCount:      d.BrokeCount,
					EnemyCount:      d.EnemyCount,
					FactionCount:    d.FactionCount,
					DepletedResLocs: d.DepletedResLocs,
					TotalResLocs:    d.TotalResLocs,
					AvgHappiness:    d.AvgHappiness,
					AvgStress:       d.AvgStress,
					CrisisLevel:     d.CrisisLevel,
				}
			}
			eng.Trends.SetSnapshots(snaps)
			log.Printf("[Main] Trends hydrated from MongoDB (%d snapshots)", len(snaps))
		}
	}

	// Build social graph from hydrated data so it's available immediately
	eng.SocialGraph.Rebuild(w, eng.Relationships)
	eng.SocialGraph.ComputeMetrics()
	log.Printf("[Main] Social graph built: %d nodes, %d edges", eng.SocialGraph.NodeCount(), eng.SocialGraph.EdgeCount())

	eng.Start(ctx)

	apiServer.SetReady(eng, mem)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("[Main] Shutting down...")
	eng.Pause()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("[Main] API shutdown error: %v", err)
	}

	if dbClient != nil {
		if err := dbClient.SaveWorld(ctx, w); err != nil {
			log.Printf("[Main] Final save failed: %v", err)
		}
		if err := dbClient.SaveRelationships(ctx, eng.Relationships); err != nil {
			log.Printf("[Main] Final relationships save failed: %v", err)
		}
		if err := dbClient.SaveSharedMemories(ctx, eng.SharedMemory); err != nil {
			log.Printf("[Main] Final shared memories save failed: %v", err)
		}
		if err := dbClient.Disconnect(ctx); err != nil {
			log.Printf("[Main] MongoDB disconnect error: %v", err)
		}
	}
}
