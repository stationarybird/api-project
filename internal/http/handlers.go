package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"api-project/internal/utils"
)

type Collections interface {
	Users() *mongo.Collection
	Tickers() *mongo.Collection
	PriceTicks() *mongo.Collection
}

type Server struct {
	Users      *mongo.Collection
	Tickers    *mongo.Collection
	PriceTicks *mongo.Collection
}

func NewServer(users, tickers, priceTicks *mongo.Collection) *Server {
	return &Server{Users: users, Tickers: tickers, PriceTicks: priceTicks}
}

func (s *Server) RegisterRoutes(e *echo.Echo) {
	e.GET("/api/users", s.getUsers)
	e.POST("/api/users", s.createUser)
	e.GET("/api/users/:id", s.getUser)

	e.GET("/api/tickers", s.getTickers)
	e.GET("/api/tickers/:symbol/price", s.getLatestPrice)
}

func (s *Server) getUsers(c echo.Context) error {
	ctx := context.TODO()
	cur, err := s.Users.Find(ctx, bson.D{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch users"})
	}
	defer cur.Close(ctx)

	var users []utils.User
	if err := cur.All(ctx, &users); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode users"})
	}

	var summaries []utils.UserSummary
	for _, user := range users {
		summaries = append(summaries, utils.UserSummary{
			Name:  user.Name,
			Email: user.Email,
		})
	}

	if summaries == nil {
		summaries = []utils.UserSummary{}
	}
	return c.JSON(http.StatusOK, summaries)
}

func (s *Server) getUser(c echo.Context) error {
	ctx := context.TODO()
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID format"})
	}
	var user utils.User
	if err := s.Users.FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch user"})
	}

	summary := utils.UserSummary{
		Name:  user.Name,
		Email: user.Email,
	}
	return c.JSON(http.StatusOK, summary)
}

func (s *Server) createUser(c echo.Context) error {
	ctx := context.TODO()
	var user utils.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}
	if user.Name == "" || user.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name and email are required"})
	}
	user.CreatedAt = time.Now()
	res, err := s.Users.InsertOne(ctx, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}
	user.ID = res.InsertedID.(primitive.ObjectID)
	return c.JSON(http.StatusCreated, user)
}

func (s *Server) getTickers(c echo.Context) error {
	ctx := context.TODO()
	cur, err := s.Tickers.Find(ctx, bson.D{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch tickers"})
	}
	defer cur.Close(ctx)

	var tickers []utils.Ticker
	if err := cur.All(ctx, &tickers); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode tickers"})
	}

	var summaries []utils.TickerSummary
	for _, ticker := range tickers {
		summaries = append(summaries, utils.TickerSummary{
			Symbol: ticker.Symbol,
			Name:   ticker.Name,
		})
	}

	if summaries == nil {
		summaries = []utils.TickerSummary{}
	}
	return c.JSON(http.StatusOK, summaries)
}

func (s *Server) getLatestPrice(c echo.Context) error {
	ctx := context.TODO()
	sym := c.Param("symbol")
	if sym == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "symbol is required"})
	}
	var pt utils.PriceTick
	if err := s.PriceTicks.FindOne(ctx, bson.M{"symbol": sym}, options.FindOne().SetSort(bson.D{{Key: "ts", Value: -1}})).Decode(&pt); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No price available"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch price"})
	}

	summary := utils.PriceSummary{
		Symbol: pt.Symbol,
		Price:  pt.Price,
		Time:   pt.TS.Format("2006-01-02 15:04:05"),
	}
	return c.JSON(http.StatusOK, summary)
}
