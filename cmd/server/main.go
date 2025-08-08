package main

import (
	"context"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	dbpkg "api-project/internal/db"
	httpapi "api-project/internal/http"
	"api-project/internal/market"
)

func main() {
	ctx := context.Background()
	client, err := dbpkg.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Mongo.Disconnect(ctx)

	go market.StartEngine(ctx, client.DB)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	srv := httpapi.NewServer(client.Col.Users, client.Col.Tickers, client.Col.PriceTicks)
	srv.RegisterRoutes(e)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("server listening on :%s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
