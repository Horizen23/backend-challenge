package config

import (
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Worker   WorkerConfig
}

type ServerConfig struct {
	Port     string
	GRPCPort string
}

type DatabaseConfig struct {
	MongoURI string
	DBName   string
}

type AuthConfig struct {
	JWTSecret string
	Duration  time.Duration
}

type WorkerConfig struct {
	UserCountInterval time.Duration
}

func LoadConfig() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:     getEnv("PORT", "8080"),
			GRPCPort: getEnv("GRPC_PORT", "50051"),
		},
		Database: DatabaseConfig{
			MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
			DBName:   getEnv("DB_NAME", "user_management_db"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "my-secure-jwt-secret-key-32bytes!"),
			Duration:  24 * time.Hour,
		},
		Worker: WorkerConfig{
			UserCountInterval: 10 * time.Second,
		},
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
