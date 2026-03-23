package db

import (
	"context"
	"fmt"
)

// TrendSnapshotDoc mirrors engine.WorldSnapshot for DB persistence.
type TrendSnapshotDoc struct {
	WorldID         string `bson:"world_id"`
	GameDay         int    `bson:"game_day"`
	Population      int    `bson:"population"`
	DeadCount       int    `bson:"dead_count"`
	TotalGold       int    `bson:"total_gold"`
	Treasury        int    `bson:"treasury"`
	HungryCount     int    `bson:"hungry_count"`
	StarvingCount   int    `bson:"starving_count"`
	BrokeCount      int    `bson:"broke_count"`
	EnemyCount      int    `bson:"enemy_count"`
	FactionCount    int    `bson:"faction_count"`
	DepletedResLocs int    `bson:"depleted_res_locs"`
	TotalResLocs    int    `bson:"total_res_locs"`
	AvgHappiness    int    `bson:"avg_happiness"`
	AvgStress       int    `bson:"avg_stress"`
	CrisisLevel     string `bson:"crisis_level"`
}

func (c *Client) SaveTrends(ctx context.Context, snapshots []TrendSnapshotDoc) error {
	col := c.Collection(ColTrends)
	_, err := col.DeleteMany(ctx, c.worldFilter())
	if err != nil {
		return fmt.Errorf("clear trends: %w", err)
	}
	if len(snapshots) == 0 {
		return nil
	}
	docs := make([]interface{}, len(snapshots))
	for i, s := range snapshots {
		s.WorldID = c.worldID
		docs[i] = s
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save trends: %w", err)
	}
	return nil
}

func (c *Client) LoadTrends(ctx context.Context) ([]TrendSnapshotDoc, error) {
	cursor, err := c.Collection(ColTrends).Find(ctx, c.worldFilter())
	if err != nil {
		return nil, fmt.Errorf("load trends: %w", err)
	}
	var docs []TrendSnapshotDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode trends: %w", err)
	}
	return docs, nil
}
