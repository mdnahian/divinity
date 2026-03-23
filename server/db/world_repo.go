package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/divinity/core/building"
	"github.com/divinity/core/enemy"
	"github.com/divinity/core/faction"
	"github.com/divinity/core/knowledge"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

type WorldDoc struct {
	ID                string `bson:"_id"`
	GridW             int    `bson:"grid_w"`
	GridH             int    `bson:"grid_h"`
	MinGridW          int    `bson:"min_grid_w"`
	MinGridH          int    `bson:"min_grid_h"`
	GameDay           int    `bson:"game_day"`
	GameHour          int    `bson:"game_hour"`
	GameMinute        int    `bson:"game_minute"`
	Weather           string `bson:"weather"`
	Treasury          int    `bson:"treasury"`
	LastWeatherChange int    `bson:"last_weather_change"`
}

// WorldSummary is returned by ListWorlds for UI/selection.
type WorldSummary struct {
	ID      string `bson:"_id" json:"id"`
	GameDay int    `bson:"game_day" json:"gameDay"`
	Weather string `bson:"weather" json:"weather"`
}

func (c *Client) SaveWorld(ctx context.Context, w *world.World) error {
	doc := WorldDoc{
		ID:                w.ID,
		GridW:             w.GridW,
		GridH:             w.GridH,
		MinGridW:          w.MinGridW,
		MinGridH:          w.MinGridH,
		GameDay:           w.GameDay,
		GameHour:          w.GameHour,
		GameMinute:        w.GameMinute,
		Weather:           w.Weather,
		Treasury:          w.Treasury,
		LastWeatherChange: w.LastWeatherChange,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := c.Collection(ColWorlds).ReplaceOne(ctx, bson.M{"_id": w.ID}, doc, opts)
	if err != nil {
		return fmt.Errorf("save world: %w", err)
	}

	wid := w.ID
	if err := c.saveNPCs(ctx, wid, w.NPCs); err != nil {
		return err
	}
	if err := c.saveLocations(ctx, wid, w.Locations); err != nil {
		return err
	}
	if err := c.saveFactions(ctx, wid, w.Factions); err != nil {
		return err
	}
	if err := c.saveEnemies(ctx, wid, w.Enemies); err != nil {
		return err
	}
	if err := c.saveTechniques(ctx, wid, w.Techniques); err != nil {
		return err
	}
	if err := c.saveGroundItems(ctx, wid, w.GroundItems); err != nil {
		return err
	}
	if err := c.saveConstructions(ctx, wid, w.Constructions); err != nil {
		return err
	}
	if err := c.saveEvents(ctx, wid, w.EventLog); err != nil {
		return err
	}
	if err := c.saveActiveEvents(ctx, wid, w.ActiveEvents); err != nil {
		return err
	}
	if err := c.savePrayers(ctx, wid, w.RecentPrayers); err != nil {
		return err
	}
	if err := c.saveEconomy(ctx, wid, w.Economy); err != nil {
		return err
	}

	return nil
}

func (c *Client) LoadWorld(ctx context.Context, worldID string, basePrices map[string]int, minMult, maxMult float64) (*world.World, error) {
	var doc WorldDoc
	err := c.Collection(ColWorlds).FindOne(ctx, bson.M{"_id": worldID}).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("load world: %w", err)
	}

	w := world.NewWorld(basePrices, minMult, maxMult)
	w.ID = doc.ID
	w.GridW = doc.GridW
	w.GridH = doc.GridH
	w.MinGridW = doc.MinGridW
	w.MinGridH = doc.MinGridH
	w.GameDay = doc.GameDay
	w.GameHour = doc.GameHour
	w.GameMinute = doc.GameMinute
	w.Weather = doc.Weather
	w.Treasury = doc.Treasury
	w.LastWeatherChange = doc.LastWeatherChange
	w.PrevDay = doc.GameDay

	filter := bson.M{"world_id": worldID}

	npcs, err := c.loadNPCs(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.NPCs = npcs

	locations, err := c.loadLocations(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.Locations = locations

	factions, err := c.loadFactions(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.Factions = factions

	enemies, err := c.loadEnemies(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.Enemies = enemies

	techniques, err := c.loadTechniques(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.Techniques = techniques

	groundItems, err := c.loadGroundItems(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.GroundItems = groundItems

	constructions, err := c.loadConstructions(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.Constructions = constructions

	events, err := c.loadEvents(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.EventLog = events

	activeEvents, err := c.loadActiveEvents(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.ActiveEvents = activeEvents

	prayers, err := c.loadPrayers(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.RecentPrayers = prayers

	economy, err := c.loadEconomy(ctx, filter)
	if err != nil {
		return nil, err
	}
	w.Economy = economy

	return w, nil
}

func (c *Client) HasWorld(ctx context.Context, worldID string) bool {
	count, err := c.Collection(ColWorlds).CountDocuments(ctx, bson.M{"_id": worldID})
	return err == nil && count > 0
}

// ListWorlds returns summaries of all stored worlds.
func (c *Client) ListWorlds(ctx context.Context) ([]WorldSummary, error) {
	cursor, err := c.Collection(ColWorlds).Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list worlds: %w", err)
	}
	var summaries []WorldSummary
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, fmt.Errorf("decode worlds: %w", err)
	}
	return summaries, nil
}

// --- sub-collection helpers (all scoped by world_id) ---

func (c *Client) saveNPCs(ctx context.Context, worldID string, npcs []*npc.NPC) error {
	col := c.Collection(ColNPCs)
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear npcs: %w", err)
	}
	if len(npcs) == 0 {
		return nil
	}
	docs := make([]interface{}, len(npcs))
	for i, n := range npcs {
		docs[i] = n
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap npcs: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save npcs: %w", err)
	}
	return nil
}

func (c *Client) loadNPCs(ctx context.Context, filter bson.M) ([]*npc.NPC, error) {
	cursor, err := c.Collection(ColNPCs).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load npcs: %w", err)
	}
	var npcs []*npc.NPC
	if err := cursor.All(ctx, &npcs); err != nil {
		return nil, fmt.Errorf("decode npcs: %w", err)
	}
	if npcs == nil {
		npcs = make([]*npc.NPC, 0)
	}
	return npcs, nil
}

func (c *Client) saveLocations(ctx context.Context, worldID string, locations []*world.Location) error {
	col := c.Collection(ColLocations)
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear locations: %w", err)
	}
	if len(locations) == 0 {
		return nil
	}
	docs := make([]interface{}, len(locations))
	for i, l := range locations {
		docs[i] = l
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap locations: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save locations: %w", err)
	}
	return nil
}

func (c *Client) loadLocations(ctx context.Context, filter bson.M) ([]*world.Location, error) {
	cursor, err := c.Collection(ColLocations).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load locations: %w", err)
	}
	var locations []*world.Location
	if err := cursor.All(ctx, &locations); err != nil {
		return nil, fmt.Errorf("decode locations: %w", err)
	}
	if locations == nil {
		locations = make([]*world.Location, 0)
	}
	return locations, nil
}

func (c *Client) saveFactions(ctx context.Context, worldID string, factions []*faction.Faction) error {
	col := c.Collection(ColFactions)
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear factions: %w", err)
	}
	if len(factions) == 0 {
		return nil
	}
	docs := make([]interface{}, len(factions))
	for i, f := range factions {
		docs[i] = f
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap factions: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save factions: %w", err)
	}
	return nil
}

func (c *Client) loadFactions(ctx context.Context, filter bson.M) ([]*faction.Faction, error) {
	cursor, err := c.Collection(ColFactions).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load factions: %w", err)
	}
	var factions []*faction.Faction
	if err := cursor.All(ctx, &factions); err != nil {
		return nil, fmt.Errorf("decode factions: %w", err)
	}
	if factions == nil {
		factions = make([]*faction.Faction, 0)
	}
	return factions, nil
}

func (c *Client) saveEnemies(ctx context.Context, worldID string, enemies []*enemy.Enemy) error {
	col := c.Collection(ColEnemies)
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear enemies: %w", err)
	}
	if len(enemies) == 0 {
		return nil
	}
	docs := make([]interface{}, len(enemies))
	for i, e := range enemies {
		docs[i] = e
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap enemies: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save enemies: %w", err)
	}
	return nil
}

func (c *Client) loadEnemies(ctx context.Context, filter bson.M) ([]*enemy.Enemy, error) {
	cursor, err := c.Collection(ColEnemies).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load enemies: %w", err)
	}
	var enemies []*enemy.Enemy
	if err := cursor.All(ctx, &enemies); err != nil {
		return nil, fmt.Errorf("decode enemies: %w", err)
	}
	if enemies == nil {
		enemies = make([]*enemy.Enemy, 0)
	}
	return enemies, nil
}

func (c *Client) saveTechniques(ctx context.Context, worldID string, techniques []*knowledge.Technique) error {
	col := c.Collection(ColTechniques)
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear techniques: %w", err)
	}
	if len(techniques) == 0 {
		return nil
	}
	docs := make([]interface{}, len(techniques))
	for i, t := range techniques {
		docs[i] = t
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap techniques: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save techniques: %w", err)
	}
	return nil
}

func (c *Client) loadTechniques(ctx context.Context, filter bson.M) ([]*knowledge.Technique, error) {
	cursor, err := c.Collection(ColTechniques).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load techniques: %w", err)
	}
	var techniques []*knowledge.Technique
	if err := cursor.All(ctx, &techniques); err != nil {
		return nil, fmt.Errorf("decode techniques: %w", err)
	}
	if techniques == nil {
		techniques = make([]*knowledge.Technique, 0)
	}
	return techniques, nil
}

func (c *Client) saveGroundItems(ctx context.Context, worldID string, items []world.GroundItem) error {
	col := c.Collection(ColGroundItems)
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear ground_items: %w", err)
	}
	if len(items) == 0 {
		return nil
	}
	docs := make([]interface{}, len(items))
	for i, it := range items {
		docs[i] = it
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap ground_items: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save ground_items: %w", err)
	}
	return nil
}

func (c *Client) loadGroundItems(ctx context.Context, filter bson.M) ([]world.GroundItem, error) {
	cursor, err := c.Collection(ColGroundItems).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load ground_items: %w", err)
	}
	var items []world.GroundItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("decode ground_items: %w", err)
	}
	if items == nil {
		items = make([]world.GroundItem, 0)
	}
	return items, nil
}

func (c *Client) saveConstructions(ctx context.Context, worldID string, constructions []*building.Construction) error {
	col := c.Collection(ColConstructions)
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear constructions: %w", err)
	}
	if len(constructions) == 0 {
		return nil
	}
	docs := make([]interface{}, len(constructions))
	for i, cn := range constructions {
		docs[i] = cn
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap constructions: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save constructions: %w", err)
	}
	return nil
}

func (c *Client) loadConstructions(ctx context.Context, filter bson.M) ([]*building.Construction, error) {
	cursor, err := c.Collection(ColConstructions).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load constructions: %w", err)
	}
	var constructions []*building.Construction
	if err := cursor.All(ctx, &constructions); err != nil {
		return nil, fmt.Errorf("decode constructions: %w", err)
	}
	if constructions == nil {
		constructions = make([]*building.Construction, 0)
	}
	return constructions, nil
}

func (c *Client) saveEvents(ctx context.Context, worldID string, events []world.EventEntry) error {
	col := c.Collection(ColEvents)
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear events: %w", err)
	}
	if len(events) == 0 {
		return nil
	}
	docs := make([]interface{}, len(events))
	for i, e := range events {
		docs[i] = e
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap events: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save events: %w", err)
	}
	return nil
}

func (c *Client) loadEvents(ctx context.Context, filter bson.M) ([]world.EventEntry, error) {
	cursor, err := c.Collection(ColEvents).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load events: %w", err)
	}
	var events []world.EventEntry
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("decode events: %w", err)
	}
	if events == nil {
		events = make([]world.EventEntry, 0)
	}
	return events, nil
}

func (c *Client) saveActiveEvents(ctx context.Context, worldID string, events []world.ActiveEvent) error {
	col := c.Collection("active_events")
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear active_events: %w", err)
	}
	if len(events) == 0 {
		return nil
	}
	docs := make([]interface{}, len(events))
	for i, e := range events {
		docs[i] = e
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap active_events: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save active_events: %w", err)
	}
	return nil
}

func (c *Client) loadActiveEvents(ctx context.Context, filter bson.M) ([]world.ActiveEvent, error) {
	cursor, err := c.Collection("active_events").Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load active_events: %w", err)
	}
	var events []world.ActiveEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("decode active_events: %w", err)
	}
	if events == nil {
		events = make([]world.ActiveEvent, 0)
	}
	return events, nil
}

func (c *Client) savePrayers(ctx context.Context, worldID string, prayers []world.Prayer) error {
	col := c.Collection("prayers")
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear prayers: %w", err)
	}
	if len(prayers) == 0 {
		return nil
	}
	docs := make([]interface{}, len(prayers))
	for i, p := range prayers {
		docs[i] = p
	}
	docs, err = c.wrapDocsWithID(worldID, docs)
	if err != nil {
		return fmt.Errorf("wrap prayers: %w", err)
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save prayers: %w", err)
	}
	return nil
}

func (c *Client) loadPrayers(ctx context.Context, filter bson.M) ([]world.Prayer, error) {
	cursor, err := c.Collection("prayers").Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load prayers: %w", err)
	}
	var prayers []world.Prayer
	if err := cursor.All(ctx, &prayers); err != nil {
		return nil, fmt.Errorf("decode prayers: %w", err)
	}
	if prayers == nil {
		prayers = make([]world.Prayer, 0)
	}
	return prayers, nil
}

func (c *Client) saveEconomy(ctx context.Context, worldID string, economy map[string]world.EconomyEntry) error {
	col := c.Collection("economy")
	_, err := col.DeleteMany(ctx, bson.M{"world_id": worldID})
	if err != nil {
		return fmt.Errorf("clear economy: %w", err)
	}
	if len(economy) == 0 {
		return nil
	}
	type econDoc struct {
		WorldID  string             `bson:"world_id"`
		ItemName string             `bson:"item_name"`
		Entry    world.EconomyEntry `bson:"entry"`
	}
	docs := make([]interface{}, 0, len(economy))
	for name, entry := range economy {
		docs = append(docs, econDoc{WorldID: worldID, ItemName: name, Entry: entry})
	}
	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("save economy: %w", err)
	}
	return nil
}

func (c *Client) loadEconomy(ctx context.Context, filter bson.M) (map[string]world.EconomyEntry, error) {
	type econDoc struct {
		WorldID  string             `bson:"world_id"`
		ItemName string             `bson:"item_name"`
		Entry    world.EconomyEntry `bson:"entry"`
	}
	cursor, err := c.Collection("economy").Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load economy: %w", err)
	}
	var docs []econDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode economy: %w", err)
	}
	economy := make(map[string]world.EconomyEntry)
	for _, d := range docs {
		economy[d.ItemName] = d.Entry
	}
	return economy, nil
}

// wrapDocsWithID injects world_id into each document using bson marshal/unmarshal.
func (c *Client) wrapDocsWithID(worldID string, items []interface{}) ([]interface{}, error) {
	out := make([]interface{}, len(items))
	for i, item := range items {
		raw, err := bson.Marshal(item)
		if err != nil {
			return nil, err
		}
		var doc bson.D
		if err := bson.Unmarshal(raw, &doc); err != nil {
			return nil, err
		}
		out[i] = append(bson.D{{Key: "world_id", Value: worldID}}, doc...)
	}
	return out, nil
}
