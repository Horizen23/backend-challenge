//go:build swagger
// +build swagger

package main

import (
	_ "backend-challenge/apps/gateway-api/docs"
	"backend-challenge/apps/gateway-api/internal/config"
	"backend-challenge/apps/common/pkg/logger"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// registerSwaggerUI registers the Swagger UI route on the Echo instance.
// Only compiled when built with build tag: -tags swagger.
func registerSwaggerUI(e *echo.Echo, cfg *config.Config, log logger.Logger) {
	if os.Getenv("APP_ENV") == "production" || os.Getenv("ENV") == "production" {
		log.Warn("Swagger UI is disabled in production")
		return
	}

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	log.Info("Swagger UI enabled",
		logger.String("url", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)),
	)
}
