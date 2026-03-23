package db

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

// UpsertNPCs saves only the specified NPCs using bulk upsert operations.
func (c *Client) UpsertNPCs(ctx context.Context, worldID string, npcs []*npc.NPC) error {
	if len(npcs) == 0 {
		return nil
	}
	col := c.Collection(ColNPCs)

	models := make([]mongo.WriteModel, 0, len(npcs))
	for _, n := range npcs {
		// Marshal to bson, inject world_id
		raw, err := bson.Marshal(n)
		if err != nil {
			log.Printf("[DB] Failed to marshal NPC %s: %v", n.ID, err)
			continue
		}
		var doc bson.D
		if err := bson.Unmarshal(raw, &doc); err != nil {
			log.Printf("[DB] Failed to unmarshal NPC %s: %v", n.ID, err)
			continue
		}
		doc = append(bson.D{{Key: "world_id", Value: worldID}}, doc...)

		filter := bson.M{"world_id": worldID, "id": n.ID}
		model := mongo.NewReplaceOneModel().
			SetFilter(filter).
			SetReplacement(doc).
			SetUpsert(true)
		models = append(models, model)
	}

	if len(models) == 0 {
		return nil
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := col.BulkWrite(ctx, models, opts)
	if err != nil {
		return fmt.Errorf("upsert npcs (%d): %w", len(models), err)
	}
	return nil
}

// UpsertLocations saves only the specified locations using bulk upsert operations.
func (c *Client) UpsertLocations(ctx context.Context, worldID string, locations []*world.Location) error {
	if len(locations) == 0 {
		return nil
	}
	col := c.Collection(ColLocations)

	models := make([]mongo.WriteModel, 0, len(locations))
	for _, l := range locations {
		raw, err := bson.Marshal(l)
		if err != nil {
			log.Printf("[DB] Failed to marshal location %s: %v", l.ID, err)
			continue
		}
		var doc bson.D
		if err := bson.Unmarshal(raw, &doc); err != nil {
			log.Printf("[DB] Failed to unmarshal location %s: %v", l.ID, err)
			continue
		}
		doc = append(bson.D{{Key: "world_id", Value: worldID}}, doc...)

		filter := bson.M{"world_id": worldID, "id": l.ID}
		model := mongo.NewReplaceOneModel().
			SetFilter(filter).
			SetReplacement(doc).
			SetUpsert(true)
		models = append(models, model)
	}

	if len(models) == 0 {
		return nil
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := col.BulkWrite(ctx, models, opts)
	if err != nil {
		return fmt.Errorf("upsert locations (%d): %w", len(models), err)
	}
	return nil
}
