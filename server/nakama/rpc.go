package nakama

import (
	"encoding/json"
)

type RPCInspectRequest struct {
	NpcID string `json:"npc_id"`
}

func (m *Module) RPCInspectNPC(payload []byte) ([]byte, error) {
	var req RPCInspectRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	n := m.Engine.World.FindNPCByID(req.NpcID)
	if n == nil {
		return json.Marshal(map[string]string{"error": "NPC not found"})
	}

	return json.Marshal(n)
}

func (m *Module) RPCWorldState(_ []byte) ([]byte, error) {
	w := m.Engine.World

	state := map[string]interface{}{
		"grid_w":      w.GridW,
		"grid_h":      w.GridH,
		"game_day":    w.GameDay,
		"game_hour":   w.GameHour,
		"game_minute": w.GameMinute,
		"weather":     w.Weather,
		"treasury":    w.Treasury,
		"npc_count":   len(w.AliveNPCs()),
		"enemy_count": len(w.AliveEnemies()),
		"locations":   w.Locations,
		"factions":    w.Factions,
		"techniques":  w.Techniques,
	}

	return json.Marshal(state)
}

func (m *Module) RPCEngineStatus(_ []byte) ([]byte, error) {
	status := map[string]interface{}{
		"running":    m.Engine.IsRunning(),
		"tick_count": m.Engine.TickCount(),
	}
	return json.Marshal(status)
}
