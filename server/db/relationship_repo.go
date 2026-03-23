package db

import (
	"context"
	"fmt"

	"github.com/divinity/core/memory"
)

type relationshipDoc struct {
	WorldID    string   `bson:"world_id"`
	NpcID      string   `bson:"npc_id"`
	TargetID   string   `bson:"target_id"`
	TargetName string   `bson:"target_name"`
	Sentiment  float64  `bson:"sentiment"`
	LastUpdate int64    `bson:"last_update"`
	History    []string `bson:"history"`
}

func (c *Client) SaveRelationships(ctx context.Context, rs *memory.RelationshipStore) error {
	col := c.Collection(ColRelationships)
	_, err := col.DeleteMany(ctx, c.worldFilter())
	if err != nil {
		return fmt.Errorf("clear relationships: %w", err)
	}

	all := rs.AllEntries()
	if len(all) == 0 {
		return nil
	}

	docs := make([]interface{}, len(all))
	for i, e := range all {
		docs[i] = relationshipDoc{
			WorldID:    c.worldID,
			NpcID:      e.NpcID,
			TargetID:   e.Rel.TargetID,
			TargetName: e.Rel.TargetName,
			Sentiment:  e.Rel.Sentiment,
			LastUpdate: e.Rel.LastUpdate,
			History:    e.Rel.History,
		}
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save relationships: %w", err)
	}
	return nil
}

func (c *Client) LoadRelationships(ctx context.Context) (*memory.RelationshipStore, error) {
	rs := memory.NewRelationshipStore()

	cursor, err := c.Collection(ColRelationships).Find(ctx, c.worldFilter())
	if err != nil {
		return rs, fmt.Errorf("load relationships: %w", err)
	}
	var docs []relationshipDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return rs, fmt.Errorf("decode relationships: %w", err)
	}

	for _, d := range docs {
		rs.Set(d.NpcID, &memory.Relationship{
			TargetID:   d.TargetID,
			TargetName: d.TargetName,
			Sentiment:  d.Sentiment,
			LastUpdate: d.LastUpdate,
			History:    d.History,
		})
	}
	return rs, nil
}
