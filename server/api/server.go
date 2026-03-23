package api

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/divinity/core/config"
	"github.com/divinity/core/engine"
	"github.com/divinity/core/memory"
)

// GenesisProgress tracks the current state of world generation for the loading screen.
type GenesisProgress struct {
	Phase   atomic.Value // string: current phase name
	Detail  atomic.Value // string: current detail text
	Current atomic.Int32 // current step within phase
	Total   atomic.Int32 // total steps in phase
}

func NewGenesisProgress() *GenesisProgress {
	gp := &GenesisProgress{}
	gp.Phase.Store("Initializing")
	gp.Detail.Store("Starting Divinity...")
	return gp
}

func (gp *GenesisProgress) Set(phase, detail string, current, total int) {
	gp.Phase.Store(phase)
	gp.Detail.Store(detail)
	gp.Current.Store(int32(current))
	gp.Total.Store(int32(total))
	log.Printf("[Genesis] %s — %s (%d/%d)", phase, detail, current, total)
}

func (gp *GenesisProgress) Snapshot() map[string]interface{} {
	return map[string]interface{}{
		"phase":   gp.Phase.Load(),
		"detail":  gp.Detail.Load(),
		"current": gp.Current.Load(),
		"total":   gp.Total.Load(),
	}
}

type Server struct {
	Engine      *engine.Engine
	Config      *config.Config
	Memory      memory.Store
	Genesis     *GenesisProgress
	RateLimiter *RateLimiter
	router      *gin.Engine
	srv         *http.Server
	hub         *Hub
	agentHub    *AgentHub
	ready       atomic.Bool
}

func NewServer(cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	s := &Server{
		Config:      cfg,
		Genesis:     NewGenesisProgress(),
		RateLimiter: NewRateLimiter(&cfg.RateLimit),
		router:      r,
		hub:         NewHub(),
		agentHub:    NewAgentHub(),
	}

	r.Use(s.readinessMiddleware())
	s.registerRoutes()
	return s
}

func (s *Server) SetReady(eng *engine.Engine, mem memory.Store) {
	s.Engine = eng
	s.Memory = mem
	eng.AfterTick = s.BroadcastTick
	eng.OnActionComplete = s.NotifyActionComplete
	s.ready.Store(true)
	log.Println("[API] Server is ready to serve requests")
}

func (s *Server) readinessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/health" {
			c.Next()
			return
		}
		if !s.ready.Load() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "loading",
				"message": "World is being generated, please wait...",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (s *Server) Start(addr string) {
	s.srv = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	go func() {
		log.Printf("[API] Listening on %s", addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[API] Server error: %v", err)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
