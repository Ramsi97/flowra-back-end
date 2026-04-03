package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI            string
	DBName              string
	JWTSecret           string
	Port                string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	CloudinaryFolder    string
	GeminiAPIKey        string // ready for real AI integration
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	return &Config{
		MongoURI:            getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:              getEnv("DB_NAME", "flowra"),
		JWTSecret:           getEnv("JWT_SECRET", "changeme"),
		Port:                getEnv("PORT", "8080"),
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),
		CloudinaryFolder:    getEnv("CLOUDINARY_FOLDER", "flowra"),
		GeminiAPIKey:        getEnv("GEMINI_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
