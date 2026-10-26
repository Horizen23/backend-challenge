package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-challenge/apps/gateway-api/internal/config"
	"backend-challenge/apps/common/pkg/logger"
)

// @title User Management API
// @version 1.0.0
// @description RESTful API Gateway for User Management & Authentication.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Paste raw JWT token directly (no "Bearer " prefix required) or "Bearer <token>"
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app, cleanup, err := InitializeServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}
	defer cleanup()

	// Register Swagger UI (only active when compiled with -tags swagger)
	registerSwaggerUI(app.Server, cfg, app.Logger)

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Start Background Worker (User Count Logger every 10s)
	go app.Worker.Start(rootCtx)

	// 2. Start HTTP (Echo) Server
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		app.Logger.Info("User HTTP API Server starting", logger.String("address", addr))
		if err := app.Server.Start(addr); err != nil {
			app.Logger.Info("HTTP server stopped", logger.Err(err))
		}
	}()

	<-rootCtx.Done()
	app.Logger.Info("Received shutdown signal. Stopping HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app.Cleanup(shutdownCtx)
	app.Logger.Info("User HTTP API Server stopped cleanly.")
}
