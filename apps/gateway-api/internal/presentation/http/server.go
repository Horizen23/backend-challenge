package http

import (
	"backend-challenge/apps/gateway-api/internal/config"
	"backend-challenge/apps/gateway-api/internal/infrastructure/auth"
	"backend-challenge/apps/common/pkg/ctxutils"
	"backend-challenge/apps/common/pkg/logger"
	"backend-challenge/apps/gateway-api/internal/presentation/http/handlers"
	"backend-challenge/apps/gateway-api/internal/presentation/http/middlewares"
	"backend-challenge/apps/gateway-api/internal/presentation/http/router"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func NewServer(cfg *config.Config, r *router.Router, log logger.Logger, tokenProvider auth.TokenProvider) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middlewares
	e.Use(echoMiddleware.Recover())
	e.Use(middlewares.RequestIDMiddleware())
	e.Use(middlewares.LoggingMiddleware(log))

	// Global Custom Error Handler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		reqID := ctxutils.GetRequestID(c.Request().Context())
		apiErr := handlers.MapError(err, reqID)
		_ = c.JSON(apiErr.Status, apiErr)
	}

	jwtMw := middlewares.JWTAuthMiddleware(tokenProvider)
	r.RegisterAll(e, jwtMw)

	return e
}
