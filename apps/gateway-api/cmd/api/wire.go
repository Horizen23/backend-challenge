//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"backend-challenge/apps/gateway-api/internal/config"
	"backend-challenge/apps/common/pkg/logger"
	"backend-challenge/apps/gateway-api/internal/presentation/worker"

	"github.com/google/wire"
	"github.com/labstack/echo/v4"
)

type ServerApp struct {
	Config *config.Config
	Server *echo.Echo
	Worker *worker.UserCountWorker
	Logger logger.Logger
}

func (app *ServerApp) Cleanup(ctx context.Context) {
	if app.Server != nil {
		_ = app.Server.Shutdown(ctx)
	}
	if app.Logger != nil {
		app.Logger.Info("HTTP server closed gracefully.")
	}
}

func InitializeServer(cfg *config.Config) (*ServerApp, func(), error) {
	wire.Build(
		CommonSet,
		wire.Struct(new(ServerApp), "*"),
	)
	return nil, nil, nil
}
