package db

import (
	"api-project/internal/utils"
	"context"
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collections struct {
	Users      *mongo.Collection
	Tickers    *mongo.Collection
	PriceTicks *mongo.Collection
}

type Client struct {
	Mongo *mongo.Client
	DB    *mongo.Database
	Col   Collections
}

func Connect(ctx context.Context) (*Client, error) {
	_ = godotenv.Load()
	uri := os.Getenv("MONGO_CONNECTION_URI")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "goAPI"
	}
	database := client.Database(dbName)
	c := &Client{
		Mongo: client,
		DB:    database,
		Col: Collections{
			Users:      database.Collection("users"),
			Tickers:    database.Collection("tickers"),
			PriceTicks: database.Collection("price_ticks"),
		},
	}
	if err := ensureSchema(ctx, c); err != nil {
		return nil, err
	}
	if err := ensureBaseTickers(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func ensureSchema(ctx context.Context, c *Client) error {
	if exists, err := collectionExists(ctx, c.DB, "price_ticks"); err != nil {
		return err
	} else if !exists {
		ts := options.TimeSeries().
			SetTimeField("ts").
			SetMetaField("symbol").
			SetGranularity("seconds")
		if err := c.DB.CreateCollection(ctx, "price_ticks", &options.CreateCollectionOptions{TimeSeriesOptions: ts}); err != nil && !isAlreadyExists(err) {
			return err
		}
	}

	ttlSeconds := int32(7 * 24 * 60 * 60)
	_, err := c.Col.PriceTicks.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "ts", Value: 1}},
		Options: options.Index().
			SetExpireAfterSeconds(ttlSeconds).
			SetPartialFilterExpression(bson.D{{Key: "symbol", Value: bson.D{{Key: "$exists", Value: true}}}}),
	})
	if err != nil && !isIndexAlreadyExists(err) {
		return err
	}

	_, _ = c.Col.Tickers.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "symbol", Value: 1}}, Options: options.Index().SetUnique(true)})
	_, _ = c.Col.Users.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)})
	return nil
}

func collectionExists(ctx context.Context, db *mongo.Database, name string) (bool, error) {
	names, err := db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: name}})
	if err != nil {
		return false, err
	}
	return len(names) > 0, nil
}

func isAlreadyExists(err error) bool {
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) && cmdErr.Code == 48 { // NamespaceExists
		return true
	}
	return false
}

func isIndexAlreadyExists(err error) bool {
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		switch cmdErr.Code {
		case 85, 68:
			return true
		}
	}
	return false
}

func ensureBaseTickers(ctx context.Context, c *Client) error {
	count, err := c.Col.Tickers.CountDocuments(ctx, bson.D{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now().UTC()
	seed := []interface{}{
		utils.Ticker{Symbol: "MNT", Name: "Mountain Dragon", Drift: 0.12, Volatility: 0.35, StartPrice: 85, CreatedAt: now},
		utils.Ticker{Symbol: "INF", Name: "Infernal Dragon", Drift: 0.08, Volatility: 0.45, StartPrice: 420, CreatedAt: now},
		utils.Ticker{Symbol: "ELD", Name: "Elder Dragon", Drift: 0.15, Volatility: 0.60, StartPrice: 69, CreatedAt: now},
		utils.Ticker{Symbol: "CLD", Name: "Cloud Dragon", Drift: 0.25, Volatility: 0.80, StartPrice: 1337, CreatedAt: now},
	}
	_, err = c.Col.Tickers.InsertMany(ctx, seed)
	return err
}
