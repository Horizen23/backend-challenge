//go:build !swagger
// +build !swagger

package main

import (
	"backend-challenge/apps/gateway-api/internal/config"
	"backend-challenge/apps/common/pkg/logger"

	"github.com/labstack/echo/v4"
)

// registerSwaggerUI is a no-op when not built with -tags swagger.
func registerSwaggerUI(_ *echo.Echo, _ *config.Config, _ logger.Logger) {}
