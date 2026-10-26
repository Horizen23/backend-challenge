package config

import (
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Port int
}

type AuthConfig struct {
	JWTSecret string
	Duration  time.Duration
}

type GRPCConfig struct {
	ServerAddr string
}

type Config struct {
	Server ServerConfig
	Auth   AuthConfig
	GRPC   GRPCConfig
}

func LoadConfig() (*Config, error) {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "my-super-secret-jwt-key-2026"
	}

	duration := 24 * time.Hour
	if d := os.Getenv("JWT_EXPIRATION_HOURS"); d != "" {
		if hours, err := strconv.Atoi(d); err == nil {
			duration = time.Duration(hours) * time.Hour
		}
	}

	grpcAddr := os.Getenv("GRPC_SERVER_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	return &Config{
		Server: ServerConfig{
			Port: port,
		},
		Auth: AuthConfig{
			JWTSecret: secret,
			Duration:  duration,
		},
		GRPC: GRPCConfig{
			ServerAddr: grpcAddr,
		},
	}, nil
}
