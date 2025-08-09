package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"api-project/internal/auth"
	"api-project/internal/utils"
)

func (s *Server) register(c echo.Context) error {
	ctx := context.TODO()

	var req utils.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON",
		})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Name, email, and password are required",
		})
	}

	if len(req.Password) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Password must be at least 6 characters",
		})
	}

	var existingUser utils.User
	err := s.Users.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "User with this email already exists",
		})
	} else if err != mongo.ErrNoDocuments {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to check existing user",
		})
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to hash password",
		})
	}

	user := utils.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	result, err := s.Users.InsertOne(ctx, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	token, err := auth.GenerateJWT(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	response := utils.LoginResponse{
		Token: token,
		User: utils.UserSummary{
			Name:  user.Name,
			Email: user.Email,
		},
	}

	return c.JSON(http.StatusCreated, response)
}

func (s *Server) login(c echo.Context) error {
	ctx := context.TODO()

	var req utils.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email and password are required",
		})
	}

	var user utils.User
	err := s.Users.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid email or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to find user",
		})
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid email or password",
		})
	}

	token, err := auth.GenerateJWT(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	response := utils.LoginResponse{
		Token: token,
		User: utils.UserSummary{
			Name:  user.Name,
			Email: user.Email,
		},
	}

	return c.JSON(http.StatusOK, response)
}

func (s *Server) me(c echo.Context) error {
	ctx := context.TODO()

	userID, ok := c.Get("userID").(primitive.ObjectID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid token",
		})
	}

	var user utils.User
	err := s.Users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch user",
		})
	}

	summary := utils.UserSummary{
		Name:  user.Name,
		Email: user.Email,
	}

	return c.JSON(http.StatusOK, summary)
}

func (s *Server) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header required",
				})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Bearer token required",
				})
			}

			userID, err := auth.GetUserIDFromToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

			c.Set("userID", userID)
			return next(c)
		}
	}
}
