package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	ColWorlds        = "worlds"
	ColNPCs          = "npcs"
	ColLocations     = "locations"
	ColFactions      = "factions"
	ColEnemies       = "enemies"
	ColEvents        = "events"
	ColTechniques    = "techniques"
	ColGroundItems   = "ground_items"
	ColConstructions  = "constructions"
	ColMemories       = "memories"
	ColRelationships  = "relationships"
	ColSharedMemories = "shared_memories"
	ColTrends         = "trends"
	ColTokens         = "tokens"
)

func (c *Client) EnsureIndexes(ctx context.Context) error {
	indexes := map[string][]mongo.IndexModel{
		ColNPCs: {
			{Keys: bson.D{{Key: "alive", Value: 1}}},
			{Keys: bson.D{{Key: "location_id", Value: 1}}},
		},
		ColLocations: {
			{Keys: bson.D{{Key: "type", Value: 1}}},
		},
		ColEnemies: {
			{Keys: bson.D{{Key: "alive", Value: 1}}},
			{Keys: bson.D{{Key: "location_id", Value: 1}}},
		},
		ColEvents: {
			{Keys: bson.D{{Key: "type", Value: 1}}},
		},
		ColMemories: {
			{Keys: bson.D{{Key: "npc_id", Value: 1}}},
		},
		ColConstructions: {
			{Keys: bson.D{{Key: "owner_id", Value: 1}}},
		},
		ColRelationships: {
			{Keys: bson.D{{Key: "npc_id", Value: 1}}},
		},
		ColSharedMemories: {
			{Keys: bson.D{{Key: "location_id", Value: 1}}},
		},
		ColTrends: {
			{Keys: bson.D{{Key: "game_day", Value: 1}}},
		},
	}

	for col, idxModels := range indexes {
		collection := c.Collection(col)
		for _, idx := range idxModels {
			_, err := collection.Indexes().CreateOne(ctx, idx)
			if err != nil {
				return err
			}
		}
	}

	eventsCol := c.Collection(ColEvents)
	cappedOpts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10 * 1024 * 1024)
	_ = c.DB().CreateCollection(ctx, ColEvents, cappedOpts)
	_ = eventsCol

	log.Println("[DB] Indexes ensured")
	return nil
}
