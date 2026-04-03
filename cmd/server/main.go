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
	schedhttp "github.com/Ramsi97/flowra-back-end/internal/schedule/delivery/http"
	schedmongo "github.com/Ramsi97/flowra-back-end/internal/schedule/repository/mongo"
	schedusecase "github.com/Ramsi97/flowra-back-end/internal/schedule/usecase"
	taskhttp "github.com/Ramsi97/flowra-back-end/internal/task/delivery/http"
	taskmongo "github.com/Ramsi97/flowra-back-end/internal/task/repository/mongo"
	taskusecase "github.com/Ramsi97/flowra-back-end/internal/task/usecase"
	"github.com/Ramsi97/flowra-back-end/pkg/cloudinary"
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

	// 3. Initialize Cloudinary
	cld, err := cloudinary.NewClient(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
		cfg.CloudinaryFolder,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}

	// 4. Auth layer
	authRepo := mongo.NewAuthMongoRepo(db)
	authUseCase := usecase.NewAuthUseCase(authRepo, cfg.JWTSecret)
	authHandler := authhttp.NewAuthHandler(authUseCase, cld)

	// 5. Task layer
	taskRepo := taskmongo.NewTaskMongoRepo(db)
	taskUseCase := taskusecase.NewTaskUseCase(taskRepo)
	taskHandler := taskhttp.NewTaskHandler(taskUseCase)

	// 6. Schedule layer
	schedRepo := schedmongo.NewScheduleMongoRepo(db)
	schedUseCase := schedusecase.NewScheduleUseCase(schedRepo, taskRepo)
	schedHandler := schedhttp.NewScheduleHandler(schedUseCase)

	// 7. Router + middleware
	router := gin.Default()
	jwtMW := middleware.JWTMiddleware(cfg.JWTSecret)

	authhttp.SetupRoutes(router, authHandler, jwtMW)
	taskhttp.SetupRoutes(router, taskHandler, jwtMW)
	schedhttp.SetupRoutes(router, schedHandler, jwtMW)

	// 8. Start server
	log.Printf("Server running on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
