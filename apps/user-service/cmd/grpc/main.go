package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-challenge/apps/user-service/internal/config"
	"backend-challenge/apps/common/pkg/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app, cleanup, err := InitializeServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize gRPC server: %v", err)
	}
	defer cleanup()

	addr := fmt.Sprintf(":%s", cfg.Server.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		app.Logger.Info("User gRPC Service starting", logger.String("address", addr))
		if err := app.GRPCServer.Serve(lis); err != nil {
			app.Logger.Info("gRPC server stopped", logger.Err(err))
		}
	}()

	<-rootCtx.Done()
	app.Logger.Info("Received shutdown signal. Stopping gRPC server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app.Cleanup(shutdownCtx)
	app.Logger.Info("User gRPC Service stopped cleanly.")
}
