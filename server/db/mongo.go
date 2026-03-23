package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/divinity/core/config"
)

type Client struct {
	client   *mongo.Client
	database *mongo.Database
	worldID  string
}

func Connect(ctx context.Context, cfg *config.MongoConfig) (*Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	log.Printf("[DB] Connected to MongoDB at %s (database: %s, world: %s)", cfg.URI, cfg.Database, cfg.WorldID)
	return &Client{
		client:   client,
		database: client.Database(cfg.Database),
		worldID:  cfg.WorldID,
	}, nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) DB() *mongo.Database {
	return c.database
}

func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// WorldID returns the world ID this client is scoped to.
func (c *Client) WorldID() string {
	return c.worldID
}

// worldFilter returns a bson filter scoped to this client's world.
func (c *Client) worldFilter() bson.M {
	return bson.M{"world_id": c.worldID}
}

// worldFilterWith returns a bson filter scoped to this world plus additional fields.
func (c *Client) worldFilterWith(extra bson.M) bson.M {
	f := bson.M{"world_id": c.worldID}
	for k, v := range extra {
		f[k] = v
	}
	return f
}

