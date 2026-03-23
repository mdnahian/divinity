package db

import (
	"context"
	"fmt"

	"github.com/divinity/core/memory"
)

type sharedMemoryDoc struct {
	WorldID    string  `bson:"world_id"`
	LocationID string  `bson:"location_id"`
	Text       string  `bson:"text"`
	Time       string  `bson:"time"`
	GameDay    int64   `bson:"game_day"`
	Vividness  float64 `bson:"vividness"`
}

func (c *Client) SaveSharedMemories(ctx context.Context, sm *memory.SharedMemoryStore) error {
	col := c.Collection(ColSharedMemories)
	_, err := col.DeleteMany(ctx, c.worldFilter())
	if err != nil {
		return fmt.Errorf("clear shared_memories: %w", err)
	}

	all := sm.AllEntries()
	if len(all) == 0 {
		return nil
	}

	docs := make([]interface{}, len(all))
	for i, e := range all {
		docs[i] = sharedMemoryDoc{
			WorldID:    c.worldID,
			LocationID: e.LocationID,
			Text:       e.Mem.Text,
			Time:       e.Mem.Time,
			GameDay:    e.Mem.GameDay,
			Vividness:  e.Mem.Vividness,
		}
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save shared_memories: %w", err)
	}
	return nil
}

func (c *Client) LoadSharedMemories(ctx context.Context) (*memory.SharedMemoryStore, error) {
	sm := memory.NewSharedMemoryStore()

	cursor, err := c.Collection(ColSharedMemories).Find(ctx, c.worldFilter())
	if err != nil {
		return sm, fmt.Errorf("load shared_memories: %w", err)
	}
	var docs []sharedMemoryDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return sm, fmt.Errorf("decode shared_memories: %w", err)
	}

	for _, d := range docs {
		sm.SetLocationMemory(d.LocationID, memory.LocationMemory{
			Text:      d.Text,
			Time:      d.Time,
			GameDay:   d.GameDay,
			Vividness: d.Vividness,
		})
	}
	return sm, nil
}
