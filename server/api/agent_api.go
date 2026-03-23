package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/divinity/core/action"
	"github.com/divinity/core/gametools"
	"github.com/divinity/core/god"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

func (s *Server) tokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		npcID, ok := s.Engine.Tokens.Lookup(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unknown token"})
			return
		}
		c.Set("token", token)
		c.Set("npc_id", npcID)
		c.Next()
	}
}

func (s *Server) resolveNPC(c *gin.Context) *npc.NPC {
	npcID, _ := c.Get("npc_id")
	w := s.Engine.World
	n := w.FindNPCByID(npcID.(string))
	if n == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NPC not found"})
	}
	return n
}

func (s *Server) spawnAgent(c *gin.Context) {
	w := s.Engine.World

	w.Mu.Lock()
	entry := w.PopSpawnQueue()
	w.Mu.Unlock()

	if entry == nil {
		log.Println("[Spawn] Queue empty — requesting GOD to create a new NPC")
		var err error
		entry, err = s.godFallbackSpawn(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
	}

	w.Mu.Lock()
	n := w.FindNPCByID(entry.NPCID)
	if n == nil {
		w.Mu.Unlock()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spawn queue NPC not found in world"})
		return
	}
	n.Claimed = true
	token := s.Engine.Tokens.Generate(n.ID)
	w.Mu.Unlock()

	if !s.Engine.IsRunning() {
		s.Engine.Resume()
	}

	loc := w.LocationByID(n.LocationID)
	locName := n.LocationID
	if loc != nil {
		locName = loc.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"token":       token,
		"npc_id":      n.ID,
		"name":        n.Name,
		"profession":  n.Profession,
		"personality": n.PersonalitySummary(),
		"location":    locName,
		"hp":          n.HP,
	})
}

func (s *Server) godFallbackSpawn(ctx context.Context) (*world.SpawnEntry, error) {
	cfg := s.Config
	if cfg.API.OpenRouterKey == "" || !cfg.GodAgent.Enabled {
		return nil, fmt.Errorf("no NPCs available and GOD agent is disabled")
	}

	entry, err := god.SpawnOnDemand(ctx, s.Engine.World, s.Engine.Router, cfg)
	if err != nil {
		log.Printf("[Spawn] GOD fallback failed: %v", err)
		return nil, fmt.Errorf("no NPCs available and GOD failed to create one")
	}

	log.Printf("[Spawn] GOD created %s the %s on demand", entry.Name, entry.Profession)
	return entry, nil
}

func (s *Server) getAgentTools(c *gin.Context) {
	w := s.Engine.World
	w.Mu.RLock()
	n := s.resolveNPC(c)
	if n == nil {
		w.Mu.RUnlock()
		return
	}
	if !n.Alive {
		w.Mu.RUnlock()
		c.JSON(http.StatusGone, gin.H{"error": "NPC is dead"})
		return
	}
	currentTick := s.Engine.TickCount()
	busy := n.BusyUntilTick > currentTick
	w.Mu.RUnlock()

	tools := gametools.NPCTools()
	var schemas []map[string]interface{}
	for _, t := range tools {
		schemas = append(schemas, t.ToOpenAI())
	}

	c.JSON(http.StatusOK, gin.H{
		"npc_id": n.ID,
		"name":   n.Name,
		"busy":   busy,
		"tools":  schemas,
	})
}

func (s *Server) callAgentTool(c *gin.Context) {
	w := s.Engine.World
	w.Mu.RLock()
	n := s.resolveNPC(c)
	if n == nil {
		w.Mu.RUnlock()
		return
	}
	if !n.Alive {
		w.Mu.RUnlock()
		c.JSON(http.StatusGone, gin.H{"error": "NPC is dead"})
		return
	}
	w.Mu.RUnlock()

	var req struct {
		Tool string          `json:"tool"`
		Args json.RawMessage `json:"args"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	tools := gametools.NPCTools()
	var tool *gametools.ToolDef
	for _, t := range tools {
		if t.Name == req.Tool {
			tool = t
			break
		}
	}
	if tool == nil {
		log.Printf("[NPC:%s] Unknown tool call: %q", n.Name, req.Tool)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown tool %q", req.Tool)})
		return
	}

	if tool.IsTerminal {
		c.JSON(http.StatusBadRequest, gin.H{"error": "use /commit endpoint for terminal actions"})
		return
	}

	log.Printf("[NPC:%s] Tool call: %s | args: %s", n.Name, req.Tool, truncateStr(string(req.Args), 500))

	agentCtx := &gametools.AgentContext{
		NPC:           n,
		World:         w,
		Memory:        s.Memory,
		Relationships: s.Engine.Relationships,
		SharedMemory:  s.Engine.SharedMemory,
		Config:        s.Config,
	}

	w.Mu.RLock()
	result, err := tool.Handler(agentCtx, req.Args)
	w.Mu.RUnlock()
	if err != nil {
		log.Printf("[NPC:%s] Tool error: %s | %v", n.Name, req.Tool, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[NPC:%s] Tool result: %s | %s", n.Name, req.Tool, truncateStr(result, 500))

	c.JSON(http.StatusOK, gin.H{
		"tool":   req.Tool,
		"result": result,
	})
}

func (s *Server) commitAgentAction(c *gin.Context) {
	w := s.Engine.World
	w.Mu.RLock()
	n := s.resolveNPC(c)
	if n == nil {
		w.Mu.RUnlock()
		return
	}
	if !n.Alive {
		w.Mu.RUnlock()
		c.JSON(http.StatusGone, gin.H{"error": "NPC is dead"})
		return
	}
	currentTick := s.Engine.TickCount()
	w.Mu.RUnlock()

	var req struct {
		ActionID string `json:"action_id"`
		Target   string `json:"target"`
		Dialogue string `json:"dialogue"`
		Goal     string `json:"goal"`
		Location string `json:"location"`
		Reason   string `json:"reason"`
		Force    bool   `json:"force"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if n.BusyUntilTick > currentTick && !req.Force {
		c.JSON(http.StatusConflict, gin.H{"error": "NPC is busy", "busy_until_tick": n.BusyUntilTick})
		return
	}

	// Capture pre-interrupt state for response metadata
	wasInterrupted := req.Force && n.BusyUntilTick > currentTick
	interruptedAction := ""
	interruptTicksLeft := 0
	if wasInterrupted {
		interruptedAction = n.PendingActionID
		interruptTicksLeft = n.BusyUntilTick - currentTick
	}

	log.Printf("[NPC:%s] Commit action: %s | target: %q | location: %q | dialogue: %q | goal: %q | reason: %q | force: %v",
		n.Name, req.ActionID, req.Target, req.Location, truncateStr(req.Dialogue, 200), truncateStr(req.Goal, 200), truncateStr(req.Reason, 200), req.Force)

	args, _ := json.Marshal(req)
	agentCtx := &gametools.AgentContext{
		NPC:           n,
		World:         w,
		Memory:        s.Memory,
		Relationships: s.Engine.Relationships,
		SharedMemory:  s.Engine.SharedMemory,
		Config:        s.Config,
	}

	tools := gametools.NPCTools()
	var commitTool *gametools.ToolDef
	for _, t := range tools {
		if t.Name == "commit_action" {
			commitTool = t
			break
		}
	}

	w.Mu.RLock()
	result, _ := commitTool.Handler(agentCtx, args)
	w.Mu.RUnlock()

	if result != "ACTION_COMMITTED" {
		log.Printf("[NPC:%s] Commit REJECTED: %s | reason: %s", n.Name, req.ActionID, result)
		c.JSON(http.StatusBadRequest, gin.H{"error": result})
		return
	}

	s.Engine.SubmitExternalAction(n, req.ActionID, req.Target, req.Dialogue, req.Goal, req.Location, req.Reason, req.Force)
	log.Printf("[NPC:%s] Action committed: %s", n.Name, req.ActionID)

	response := gin.H{
		"status":    "committed",
		"action_id": req.ActionID,
		"npc_id":    n.ID,
	}
	if wasInterrupted {
		response["interrupted_action"] = interruptedAction
		response["resumable"] = true
		response["resume_ticks_left"] = interruptTicksLeft
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) getAgentState(c *gin.Context) {
	w := s.Engine.World

	// Snapshot NPC state under RLock, release before formatting memories/relationships
	w.Mu.RLock()
	n := s.resolveNPC(c)
	if n == nil {
		w.Mu.RUnlock()
		return
	}
	currentTick := s.Engine.TickCount()
	busy := n.BusyUntilTick > currentTick

	loc := w.LocationByID(n.LocationID)
	locName := n.LocationID
	if loc != nil {
		locName = loc.Name
	}

	// Snapshot scalar NPC fields while holding lock
	npcID := n.ID
	npcName := n.Name
	alive := n.Alive
	busyUntilTick := n.BusyUntilTick
	pendingAction := n.PendingActionID
	resumeAction := n.ResumeActionID
	resumeTicksLeft := n.ResumeTicksLeft
	hp := n.HP
	hunger := n.Needs.Hunger
	thirst := n.Needs.Thirst
	fatigue := n.Needs.Fatigue
	w.Mu.RUnlock()

	// Memory and relationship stores have their own locks — no world lock needed
	memoryContext := ""
	if s.Memory != nil {
		memoryContext = s.Memory.FormatBlendedForLLM(npcID)
	}
	relationshipContext := ""
	if s.Engine.Relationships != nil {
		relationshipContext = s.Engine.Relationships.FormatForLLM(npcID)
	}

	c.JSON(http.StatusOK, gin.H{
		"npc_id":               npcID,
		"name":                 npcName,
		"alive":                alive,
		"busy":                 busy,
		"busy_until_tick":      busyUntilTick,
		"current_tick":         currentTick,
		"pending_action":       pendingAction,
		"resume_action":        resumeAction,
		"resume_ticks_left":    resumeTicksLeft,
		"location":             locName,
		"hp":                   hp,
		"hunger":               hunger,
		"thirst":               thirst,
		"fatigue":              fatigue,
		"memory_context":       memoryContext,
		"relationship_context": relationshipContext,
	})
}

func (s *Server) getAgentPrompt(c *gin.Context) {
	w := s.Engine.World
	w.Mu.RLock()
	n := s.resolveNPC(c)
	if n == nil {
		w.Mu.RUnlock()
		return
	}
	if !n.Alive {
		w.Mu.RUnlock()
		c.JSON(http.StatusGone, gin.H{"error": "NPC is dead"})
		return
	}

	loc := w.LocationByID(n.LocationID)
	locName := n.LocationID
	if loc != nil {
		locName = loc.Name
	}

	agentCtx := &gametools.AgentContext{
		NPC:           n,
		World:         w,
		Memory:        s.Memory,
		Relationships: s.Engine.Relationships,
		SharedMemory:  s.Engine.SharedMemory,
		Config:        s.Config,
	}

	systemPrompt, userPrompt := gametools.BuildNPCPrompt(agentCtx, locName)
	w.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"npc_id":        n.ID,
		"name":          n.Name,
		"system_prompt": systemPrompt,
		"user_prompt":   userPrompt,
	})
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (s *Server) getNPCPrompt(c *gin.Context) {
	id := c.Param("id")
	w := s.Engine.World
	w.Mu.RLock()
	n := w.FindNPCByID(id)
	if n == nil {
		w.Mu.RUnlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "NPC not found"})
		return
	}

	loc := w.LocationByID(n.LocationID)
	locName := n.LocationID
	if loc != nil {
		locName = loc.Name
	}

	agentCtx := &gametools.AgentContext{
		NPC:           n,
		World:         w,
		Memory:        s.Memory,
		Relationships: s.Engine.Relationships,
		SharedMemory:  s.Engine.SharedMemory,
		Config:        s.Config,
	}

	systemPrompt, userPrompt := gametools.BuildNPCPrompt(agentCtx, locName)

	tools := gametools.NPCTools()
	var toolSchemas []map[string]interface{}
	for _, t := range tools {
		toolSchemas = append(toolSchemas, t.ToOpenAI())
	}

	mpt := s.Config.Game.GameMinutesPerTick
	travelRate := s.Config.Game.TravelMinutesPerUnit

	available := action.GetAvailableActions(n, w)
	type candidateView struct {
		Name          string `json:"name"`
		TravelMin     int    `json:"travelMin"`
		FootTravelMin int    `json:"footTravelMin,omitempty"`
		TravelMode    string `json:"travelMode,omitempty"`
		Here          bool   `json:"here,omitempty"`
	}
	type actionView struct {
		ID         string          `json:"id"`
		Label      string          `json:"label"`
		Category   string          `json:"category"`
		ActionMin  int             `json:"actionMin"`
		TravelMin  int             `json:"travelMin"`
		Skill      string          `json:"skillKey,omitempty"`
		Dest       string          `json:"destination,omitempty"`
		Candidates []candidateView `json:"candidates,omitempty"`
	}
	actions := make([]actionView, len(available))
	for i, a := range available {
		_, mins := a.Duration(n, mpt)
		av := actionView{
			ID:        a.ID,
			Label:     a.Label,
			Category:  a.Category,
			ActionMin: mins,
			Skill:     a.SkillKey,
		}

		if a.Candidates != nil {
			candidates := a.Candidates(n, w)
			for _, c := range candidates {
				cv := candidateView{Name: c.Name}
				if c.ID == n.LocationID {
					cv.Here = true
				} else {
					footTicks := w.TravelTicks(n.LocationID, c.ID, mpt, travelRate)
					footMins := footTicks * mpt
					mountedTicks, hasMnt, hasCarr := w.TravelTicksMounted(n.LocationID, c.ID, mpt, travelRate, n.ID)
					mountedMins := mountedTicks * mpt
					if hasMnt {
						cv.TravelMin = mountedMins
						cv.FootTravelMin = footMins
						if hasCarr {
							cv.TravelMode = "carriage"
						} else {
							cv.TravelMode = "riding"
						}
					} else {
						cv.TravelMin = footMins
					}
				}
				av.Candidates = append(av.Candidates, cv)
			}
		} else if a.Destination != nil {
			destID := a.Destination(n, w)
			if destID != "" && destID != n.LocationID {
				footTicks := w.TravelTicks(n.LocationID, destID, mpt, travelRate)
				footMins := footTicks * mpt
				mountedTicks, hasMnt, hasCarr := w.TravelTicksMounted(n.LocationID, destID, mpt, travelRate, n.ID)
				mountedMins := mountedTicks * mpt
				if hasMnt {
					av.TravelMin = mountedMins
					if dest := w.LocationByID(destID); dest != nil {
						if hasCarr {
							av.Dest = fmt.Sprintf("%s (by carriage, %d min on foot)", dest.Name, footMins)
						} else {
							av.Dest = fmt.Sprintf("%s (riding, %d min on foot)", dest.Name, footMins)
						}
					}
				} else {
					av.TravelMin = footMins
					if dest := w.LocationByID(destID); dest != nil {
						av.Dest = dest.Name
					}
				}
			}
		}

		actions[i] = av
	}
	w.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"npc_id":        n.ID,
		"name":          n.Name,
		"system_prompt": systemPrompt,
		"user_prompt":   userPrompt,
		"tools":         toolSchemas,
		"actions":       actions,
	})
}
