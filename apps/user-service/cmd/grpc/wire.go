//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"backend-challenge/apps/user-service/internal/config"
	"backend-challenge/apps/common/pkg/logger"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

type ServerApp struct {
	Config      *config.Config
	GRPCServer  *grpc.Server
	Logger      logger.Logger
	MongoClient *mongo.Client
}

func (app *ServerApp) Cleanup(ctx context.Context) {
	if app.GRPCServer != nil {
		app.GRPCServer.GracefulStop()
	}
	if app.MongoClient != nil {
		_ = app.MongoClient.Disconnect(ctx)
	}
	if app.Logger != nil {
		app.Logger.Info("gRPC server and connections closed gracefully.")
	}
}

func InitializeServer(cfg *config.Config) (*ServerApp, func(), error) {
	wire.Build(
		CommonSet,
		wire.Struct(new(ServerApp), "*"),
	)
	return nil, nil, nil
}
