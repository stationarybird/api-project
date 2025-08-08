package utils

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Email     string             `json:"email" bson:"email"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type Ticker struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Symbol     string             `json:"symbol" bson:"symbol"`
	Name       string             `json:"name" bson:"name"`
	Drift      float64            `json:"drift" bson:"drift"`
	Volatility float64            `json:"volatility" bson:"volatility"`
	StartPrice float64            `json:"start_price" bson:"start_price"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
}

type PriceTick struct {
	Symbol string    `json:"symbol" bson:"symbol"`
	TS     time.Time `json:"ts" bson:"ts"`
	Price  float64   `json:"price" bson:"price"`
}

type UserSummary struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TickerSummary struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type PriceSummary struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Time   string  `json:"time"`
}

const SECONDS_PER_DAY = 86400.0
