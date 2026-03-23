package api

import (
	"github.com/divinity/core/building"
	"github.com/divinity/core/config"
	"github.com/divinity/core/enemy"
	"github.com/divinity/core/faction"
	"github.com/divinity/core/knowledge"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

type MemoryView struct {
	Text       string   `json:"text"`
	Time       string   `json:"time"`
	Importance float64  `json:"importance"`
	Vividness  float64  `json:"vividness"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
}

type NPCView struct {
	*npc.NPC

	Mood               string       `json:"mood"`
	PersonalitySummary string       `json:"personalitySummary"`
	LiteracyLevel      string       `json:"literacyLevel"`
	Age                int          `json:"age"`
	LifeStage          string       `json:"lifeStage"`
	EffectiveWisdom    int          `json:"effectiveWisdom"`
	UsedSlots          int          `json:"usedSlots"`
	MaxSlots           int          `json:"maxSlots"`
	UsedWeight         float64      `json:"usedWeight"`
	MaxWeight          float64      `json:"maxWeight"`
	GoldCount          int          `json:"goldCount"`
	RecentMemories     []MemoryView `json:"recentMemories"`
}

func NewNPCView(n *npc.NPC, gameDay int, cfg *config.Config, mem memory.Store) NPCView {
	daysPerYear := cfg.Game.GameDaysPerYear

	var memories []MemoryView
	if mem != nil {
		entries := mem.Recent(n.ID, 8)
		memories = make([]MemoryView, len(entries))
		for i, e := range entries {
			memories[i] = MemoryView{Text: e.Text, Time: e.Time, Importance: e.Importance, Vividness: e.Vividness, Category: e.Category, Tags: e.Tags}
		}
	}
	if memories == nil {
		memories = []MemoryView{}
	}

	return NPCView{
		NPC:                n,
		Mood:               n.Mood(),
		PersonalitySummary: n.PersonalitySummary(),
		LiteracyLevel:      n.LiteracyLevel(),
		Age:                n.GetAge(gameDay, daysPerYear),
		LifeStage:          n.GetLifeStage(gameDay, daysPerYear),
		EffectiveWisdom:    n.GetEffectiveWisdom(gameDay, daysPerYear),
		UsedSlots:          n.UsedSlots(),
		MaxSlots:           n.MaxSlots(),
		UsedWeight:         n.UsedWeight(),
		MaxWeight:          n.MaxWeight(),
		GoldCount:          n.GoldCount(),
		RecentMemories:     memories,
	}
}

type WorldView struct {
	GridW  int `json:"gridW"`
	GridH  int `json:"gridH"`
	MinGridW int `json:"minGridW"`
	MinGridH int `json:"minGridH"`

	TickIntervalMs int `json:"tickIntervalMs"`

	GameDay    int    `json:"gameDay"`
	GameHour   int    `json:"gameHour"`
	GameMinute int    `json:"gameMinute"`
	Weather    string `json:"weather"`
	Treasury   int    `json:"treasury"`

	IsNight    bool   `json:"isNight"`
	TimeString string `json:"timeString"`

	Locations     []*world.Location          `json:"locations"`
	NPCs          []NPCView                  `json:"npcs"`
	AliveEnemies  []*enemy.Enemy             `json:"aliveEnemies"`
	Enemies       []*enemy.Enemy             `json:"enemies"`
	Factions      []*faction.Faction         `json:"factions"`
	Techniques    []*knowledge.Technique     `json:"techniques"`
	Constructions []*building.Construction   `json:"constructions"`
	Mounts        []*world.Mount             `json:"mounts"`
	Carriages     []*world.Carriage          `json:"carriages"`
	GroundItems   []world.GroundItem         `json:"groundItems"`
	ActiveEvents  []world.ActiveEvent        `json:"activeEvents"`
	Economy       map[string]world.EconomyEntry `json:"economy"`
	RecentPrayers []world.Prayer             `json:"recentPrayers"`
	EventLog      []world.EventEntry         `json:"eventLog"`
	Chronicles    []world.ChronicleEntry     `json:"chronicles"`
}

func NewWorldView(w *world.World, cfg *config.Config, mem memory.Store) WorldView {
	var claimedNPCs []*npc.NPC
	for _, n := range w.NPCs {
		if n.Claimed {
			claimedNPCs = append(claimedNPCs, n)
		}
	}
	npcViews := make([]NPCView, len(claimedNPCs))
	for i, n := range claimedNPCs {
		npcViews[i] = NewNPCView(n, w.GameDay, cfg, mem)
	}

	groundItems := w.GroundItems
	if groundItems == nil {
		groundItems = []world.GroundItem{}
	}
	activeEvents := w.ActiveEvents
	if activeEvents == nil {
		activeEvents = []world.ActiveEvent{}
	}
	eventLog := w.EventLog
	if eventLog == nil {
		eventLog = []world.EventEntry{}
	}
	prayers := w.RecentPrayers
	if prayers == nil {
		prayers = []world.Prayer{}
	}

	return WorldView{
		GridW:          w.GridW,
		GridH:          w.GridH,
		MinGridW:       w.MinGridW,
		MinGridH:       w.MinGridH,
		TickIntervalMs: cfg.Game.TickIntervalMs,
		GameDay:        w.GameDay,
		GameHour:      w.GameHour,
		GameMinute:    w.GameMinute,
		Weather:       w.Weather,
		Treasury:      w.Treasury,
		IsNight:       w.IsNight(),
		TimeString:    w.TimeString(),
		Locations:     w.Locations,
		NPCs:          npcViews,
		AliveEnemies:  w.AliveEnemies(),
		Enemies:       w.Enemies,
		Factions:      w.Factions,
		Techniques:    w.Techniques,
		Constructions: w.Constructions,
		Mounts:        w.Mounts,
		Carriages:     w.Carriages,
		GroundItems:   groundItems,
		ActiveEvents:  activeEvents,
		Economy:       w.Economy,
		RecentPrayers: prayers,
		EventLog:      eventLog,
		Chronicles:    w.Chronicles,
	}
}
