package nakama

import (
	"encoding/json"
	"log"
)

type MatchState struct {
	TickCount  int         `json:"tick_count"`
	GameDay    int         `json:"game_day"`
	GameHour   int         `json:"game_hour"`
	GameMinute int         `json:"game_minute"`
	Weather    string      `json:"weather"`
	NPCCount   int        `json:"npc_count"`
	EnemyCount int        `json:"enemy_count"`
	Events     []EventMsg  `json:"events"`
}

type EventMsg struct {
	Text string `json:"text"`
	Type string `json:"type"`
	Time string `json:"time"`
}

func (m *Module) GetMatchState() *MatchState {
	w := m.Engine.World
	eventCount := len(w.EventLog)
	recentStart := eventCount - 20
	if recentStart < 0 {
		recentStart = 0
	}

	events := make([]EventMsg, 0, 20)
	for _, e := range w.EventLog[recentStart:] {
		events = append(events, EventMsg{
			Text: e.Text,
			Type: e.Type,
			Time: e.Time,
		})
	}

	return &MatchState{
		TickCount:  m.Engine.TickCount(),
		GameDay:    w.GameDay,
		GameHour:   w.GameHour,
		GameMinute: w.GameMinute,
		Weather:    w.Weather,
		NPCCount:   len(w.AliveNPCs()),
		EnemyCount: len(w.AliveEnemies()),
		Events:     events,
	}
}

func (m *Module) BroadcastPayload() []byte {
	state := m.GetMatchState()
	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("[Nakama] Marshal error: %v", err)
		return nil
	}
	return data
}
