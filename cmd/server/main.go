package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Ramsi97/flowra-back-end/config"
	authHandler "github.com/Ramsi97/flowra-back-end/internal/auth/delivery/http"
	authRepo "github.com/Ramsi97/flowra-back-end/internal/auth/repository/mongo"
	authUseCase "github.com/Ramsi97/flowra-back-end/internal/auth/usecase"
	"github.com/Ramsi97/flowra-back-end/internal/middleware"
	schedHandler "github.com/Ramsi97/flowra-back-end/internal/schedule/delivery/http"
	schedRepo "github.com/Ramsi97/flowra-back-end/internal/schedule/repository/mongo"
	schedUseCase "github.com/Ramsi97/flowra-back-end/internal/schedule/usecase"
	taskHandler "github.com/Ramsi97/flowra-back-end/internal/task/delivery/http"
	taskRepo "github.com/Ramsi97/flowra-back-end/internal/task/repository/mongo"
	taskUseCase "github.com/Ramsi97/flowra-back-end/internal/task/usecase"
	"github.com/Ramsi97/flowra-back-end/pkg/cloudinary"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.LoadConfig()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(cfg.DBName)

	// Initialize Cloudinary
	cld, err := cloudinary.NewClient(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
		cfg.CloudinaryFolder,
	)
	if err != nil {
		log.Printf("Warning: Cloudinary client failed to initialize: %v", err)
	}

	// Initialize Repositories
	ar := authRepo.NewAuthMongoRepo(db)
	tr := taskRepo.NewTaskMongoRepo(db)
	ex := taskRepo.NewExceptionMongoRepo(db)
	sr := schedRepo.NewScheduleMongoRepo(db)

	// Initialize UseCases
	au := authUseCase.NewAuthUseCase(ar, cfg.JWTSecret)
	tu := taskUseCase.NewTaskUseCase(tr)
	su := schedUseCase.NewScheduleUseCase(sr, tr, ex, ar)

	// Initialize Handlers
	ah := authHandler.NewAuthHandler(au, cld)
	th := taskHandler.NewTaskHandler(tu)
	sh := schedHandler.NewScheduleHandler(su)

	// Setup Router
	r := gin.Default()

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Middleware
	jwtMid := middleware.JWTMiddleware(cfg.JWTSecret)

	// Setup Routes
	authHandler.SetupRoutes(r, ah, jwtMid)
	taskHandler.SetupRoutes(r, th, jwtMid)
	schedHandler.SetupRoutes(r, sh, jwtMid)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
