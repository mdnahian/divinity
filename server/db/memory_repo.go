package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/divinity/core/memory"
)

type memoryDoc struct {
	WorldID    string    `bson:"world_id"`
	NpcID      string    `bson:"npc_id"`
	Text       string    `bson:"text"`
	Time       string    `bson:"time"`
	TS         int64     `bson:"ts"`
	Importance float64   `bson:"importance"`
	Vividness  float64   `bson:"vividness"`
	Category   string    `bson:"category"`
	Tags       []string  `bson:"tags"`
	CreatedAt  time.Time `bson:"created_at"`
}

// MemoryPersister implements memory.Persister using MongoDB.
type MemoryPersister struct {
	client *Client
}

func NewMemoryPersister(c *Client) *MemoryPersister {
	return &MemoryPersister{client: c}
}

func (p *MemoryPersister) SaveMemory(npcID string, entry memory.Entry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := memoryDoc{
		WorldID:    p.client.worldID,
		NpcID:      npcID,
		Text:       entry.Text,
		Time:       entry.Time,
		TS:         entry.TS,
		Importance: entry.Importance,
		Vividness:  entry.Vividness,
		Category:   entry.Category,
		Tags:       entry.Tags,
		CreatedAt:  time.Now(),
	}

	_, err := p.client.Collection(ColMemories).InsertOne(ctx, doc)
	if err != nil {
		log.Printf("[DB] Failed to save memory for %s: %v", npcID, err)
	}
	return err
}

func (p *MemoryPersister) LoadAllMemories() (map[string][]memory.Entry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := p.client.Collection(ColMemories).Find(ctx, p.client.worldFilter())
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string][]memory.Entry)
	for cursor.Next(ctx) {
		var doc memoryDoc
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("[DB] Failed to decode memory doc: %v", err)
			continue
		}
		entry := memory.Entry{
			Text:       doc.Text,
			Time:       doc.Time,
			TS:         doc.TS,
			Importance: doc.Importance,
			Vividness:  doc.Vividness,
			Category:   doc.Category,
			Tags:       doc.Tags,
		}
		result[doc.NpcID] = append(result[doc.NpcID], entry)
	}

	log.Printf("[DB] Loaded memories for %d entities", len(result))
	return result, nil
}

func (p *MemoryPersister) ClearMemories(npcID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.client.Collection(ColMemories).DeleteMany(ctx, p.client.worldFilterWith(bson.M{"npc_id": npcID}))
	if err != nil {
		log.Printf("[DB] Failed to clear memories for %s: %v", npcID, err)
	}
	return err
}
