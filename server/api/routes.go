package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/divinity/core/world"
)

func (s *Server) registerRoutes() {
	api := s.router.Group("/api")

	api.GET("/world", s.getWorld)
	api.GET("/npcs", s.getNPCs)
	api.GET("/npcs/:id", s.getNPC)
	api.GET("/events", s.getEvents)
	api.GET("/stats", s.getStats)

	api.GET("/ws", s.handleWS)
	api.GET("/agent/ws", s.handleAgentWS)

	api.POST("/agent/spawn", s.spawnAgent)

	authed := api.Group("/agent")
	authed.Use(s.tokenAuth())
	authed.Use(s.RateLimiter.Middleware())
	authed.GET("/tools", s.getAgentTools)
	authed.POST("/call", s.callAgentTool)
	authed.POST("/commit", s.commitAgentAction)
	authed.GET("/state", s.getAgentState)
	authed.GET("/prompt", s.getAgentPrompt)

	s.router.GET("/health", func(c *gin.Context) {
		if s.ready.Load() {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "loading",
				"progress": s.Genesis.Snapshot(),
			})
		}
	})
}

func (s *Server) getWorld(c *gin.Context) {
	w := s.Engine.World
	w.Mu.RLock()
	view := NewWorldView(w, s.Config, s.Memory)
	w.Mu.RUnlock()
	c.JSON(http.StatusOK, view)
}

func (s *Server) getNPCs(c *gin.Context) {
	w := s.Engine.World
	w.Mu.RLock()
	claimed := w.AliveNPCs()
	views := make([]NPCView, len(claimed))
	for i, n := range claimed {
		views[i] = NewNPCView(n, w.GameDay, s.Config, s.Memory)
	}
	w.Mu.RUnlock()
	c.JSON(http.StatusOK, views)
}

func (s *Server) getNPC(c *gin.Context) {
	id := c.Param("id")
	w := s.Engine.World
	w.Mu.RLock()
	n := w.FindNPCByID(id)
	if n == nil {
		w.Mu.RUnlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "NPC not found"})
		return
	}
	view := NewNPCView(n, w.GameDay, s.Config, s.Memory)
	w.Mu.RUnlock()
	c.JSON(http.StatusOK, view)
}

func (s *Server) getEvents(c *gin.Context) {
	w := s.Engine.World

	// Query params: ?after_seq=N&limit=50&territory_id=X
	afterSeq := int64(0)
	if v := c.Query("after_seq"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			afterSeq = n
		}
	}
	limit := 100
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	territoryID := c.Query("territory_id")

	w.Mu.RLock()
	totalCount := w.EventSeq
	var filtered []world.EventEntry
	for i := len(w.EventLog) - 1; i >= 0 && len(filtered) < limit; i-- {
		e := w.EventLog[i]
		if afterSeq > 0 && e.Tick <= afterSeq {
			break
		}
		if territoryID != "" && e.TerritoryID != territoryID && e.TerritoryID != "" {
			continue
		}
		filtered = append(filtered, e)
	}
	w.Mu.RUnlock()

	// Reverse to chronological order
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	c.JSON(http.StatusOK, gin.H{
		"events":     filtered,
		"totalCount": totalCount,
	})
}

func (s *Server) getStats(c *gin.Context) {
	w := s.Engine.World
	w.Mu.RLock()
	data := gin.H{
		"tickCount":  s.Engine.TickCount(),
		"running":    s.Engine.IsRunning(),
		"npcCount":   len(w.AliveNPCs()),
		"enemyCount": len(w.AliveEnemies()),
		"gameDay":    w.GameDay,
	}
	w.Mu.RUnlock()
	c.JSON(http.StatusOK, data)
}
