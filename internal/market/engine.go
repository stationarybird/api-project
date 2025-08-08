package market

import (
	"context"
	"log"
	"math"
	"math/rand"
	"time"

	"api-project/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TickerView struct {
	Symbol     string
	Drift      float64
	Volatility float64
	StartPrice float64
}


func StartEngine(ctx context.Context, db *mongo.Database) {
	cur, err := db.Collection("tickers").Find(ctx, bson.D{})
	if err != nil {
		log.Println("price engine: failed to load tickers:", err)
		return
	}
	defer cur.Close(ctx)

	type state struct {
		Price float64
		Drift float64
		Sigma float64
	}
	tickerState := map[string]state{}
	for cur.Next(ctx) {
		var t struct {
			Symbol     string  `bson:"symbol"`
			Drift      float64 `bson:"drift"`
			Volatility float64 `bson:"volatility"`
			StartPrice float64 `bson:"start_price"`
		}
		if err := cur.Decode(&t); err != nil {
			continue
		}
		tickerState[t.Symbol] = state{Price: t.StartPrice, Drift: t.Drift, Sigma: t.Volatility}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ticks := db.Collection("price_ticks")

	tick := func() {
		now := time.Now().UTC()
		dt := 1.0 / utils.SECONDS_PER_DAY
		writes := make([]mongo.WriteModel, 0, len(tickerState))
		for sym, s := range tickerState {
			z := r.NormFloat64()
			next := s.Price * math.Exp((s.Drift-0.5*s.Sigma*s.Sigma)*dt+s.Sigma*math.Sqrt(dt)*z)
			if next <= 0 {
				next = s.Price
			}
			tickerState[sym] = state{Price: next, Drift: s.Drift, Sigma: s.Sigma}
			doc := bson.D{{Key: "symbol", Value: sym}, {Key: "ts", Value: now}, {Key: "price", Value: next}}
			writes = append(writes, mongo.NewInsertOneModel().SetDocument(doc))
		}
		if len(writes) > 0 {
			if _, err := ticks.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false)); err != nil {
				log.Println("price engine: bulk write error:", err)
			}
		}
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick()
		}
	}
}
