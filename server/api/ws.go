package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	conn.Close()
}

func (h *Hub) Broadcast(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[WS] write error: %v", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

func (h *Hub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

func (s *Server) handleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] upgrade error: %v", err)
		return
	}
	s.hub.Add(conn)
	log.Printf("[WS] client connected (%d total)", s.hub.Count())

	// Read loop — just waits for close
	go func() {
		defer s.hub.Remove(conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
		log.Printf("[WS] client disconnected (%d total)", s.hub.Count())
	}()
}

// AgentHub manages per-NPC WebSocket connections for push notifications.
type AgentHub struct {
	mu    sync.Mutex
	conns map[string]*websocket.Conn // npcID → conn
}

func NewAgentHub() *AgentHub {
	return &AgentHub{conns: make(map[string]*websocket.Conn)}
}

func (h *AgentHub) Set(npcID string, conn *websocket.Conn) {
	h.mu.Lock()
	if old, ok := h.conns[npcID]; ok {
		old.Close()
	}
	h.conns[npcID] = conn
	h.mu.Unlock()
}

func (h *AgentHub) Remove(npcID string) {
	h.mu.Lock()
	if conn, ok := h.conns[npcID]; ok {
		conn.Close()
		delete(h.conns, npcID)
	}
	h.mu.Unlock()
}

// Notify sends a JSON message to the agent connected for the given NPC.
func (h *AgentHub) Notify(npcID string, payload interface{}) {
	h.mu.Lock()
	conn, ok := h.conns[npcID]
	h.mu.Unlock()
	if !ok {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("[AgentWS] write error for %s: %v", npcID, err)
		h.Remove(npcID)
	}
}

func (s *Server) handleAgentWS(c *gin.Context) {
	// Authenticate via query param since WS can't use headers easily
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token query param"})
		return
	}
	npcID, ok := s.Engine.Tokens.Lookup(token)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unknown token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[AgentWS] upgrade error: %v", err)
		return
	}
	s.agentHub.Set(npcID, conn)
	log.Printf("[AgentWS] %s connected", npcID)

	// Read loop — wait for close
	go func() {
		defer func() {
			s.agentHub.Remove(npcID)
			log.Printf("[AgentWS] %s disconnected", npcID)
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// NotifyActionComplete sends an action_completed event to the NPC's agent WebSocket.
func (s *Server) NotifyActionComplete(npcID, actionID, result string) {
	s.agentHub.Notify(npcID, map[string]interface{}{
		"type":      "action_completed",
		"npc_id":    npcID,
		"action_id": actionID,
		"result":    result,
	})
}

// BroadcastTick builds the tick payload and sends it to all connected clients.
func (s *Server) BroadcastTick() {
	if s.hub.Count() == 0 {
		return
	}

	w := s.Engine.World
	w.Mu.RLock()

	worldView := NewWorldView(w, s.Config, s.Memory)

	events := w.EventLog
	totalCount := w.EventSeq
	limit := 100
	if len(events) > limit {
		events = events[len(events)-limit:]
	}

	usage := s.Engine.Router.Usage()
	stats := map[string]any{
		"tickCount":        s.Engine.TickCount(),
		"running":          s.Engine.IsRunning(),
		"npcCount":         len(w.AliveNPCs()),
		"enemyCount":       len(w.AliveEnemies()),
		"gameDay":          w.GameDay,
		"llmRequests":      usage.Requests,
		"promptTokens":     usage.PromptTokens,
		"completionTokens": usage.CompletionTokens,
		"totalTokens":      usage.TotalTokens(),
		"totalCost":        usage.ActualCost,
	}

	w.Mu.RUnlock()

	payload := map[string]any{
		"world":      worldView,
		"events":     events,
		"totalCount": totalCount,
		"stats":      stats,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WS] marshal error: %v", err)
		return
	}
	s.hub.Broadcast(data)
}
