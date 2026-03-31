package main

import (
	"context"
	"log"
	"time"

	"github.com/Ramsi97/flowra-back-end/config"
	authhttp "github.com/Ramsi97/flowra-back-end/internal/auth/delivery/http"
	"github.com/Ramsi97/flowra-back-end/internal/auth/repository/mongo"
	"github.com/Ramsi97/flowra-back-end/internal/auth/usecase"
	"github.com/Ramsi97/flowra-back-end/internal/middleware"
	"github.com/gin-gonic/gin"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := driver.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}
	log.Println("Connected to MongoDB")

	db := client.Database(cfg.DBName)

	// 3. Build dependency graph: repo → usecase → handler
	authRepo := mongo.NewAuthMongoRepo(db)
	authUseCase := usecase.NewAuthUseCase(authRepo, cfg.JWTSecret)
	authHandler := authhttp.NewAuthHandler(authUseCase)

	// 4. Set up Gin router and routes
	router := gin.Default()
	jwtMW := middleware.JWTMiddleware(cfg.JWTSecret)
	authhttp.SetupRoutes(router, authHandler, jwtMW)

	// 5. Start the server
	log.Printf("Server running on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
