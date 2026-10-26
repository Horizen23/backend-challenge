package middlewares

import (
	"time"

	"backend-challenge/apps/common/pkg/ctxutils"
	"backend-challenge/apps/common/pkg/logger"

	"github.com/labstack/echo/v4"
)

// LoggingMiddleware logs HTTP request execution details with structured fields
// (method, uri, status, latency, request_id, user_id).
func LoggingMiddleware(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			req := c.Request()
			res := c.Response()
			reqID := ctxutils.GetRequestID(req.Context())
			userID, email := ctxutils.GetAuthUser(req.Context())

			log.Info("HTTP request completed",
				logger.String("request_id", reqID),
				logger.String("method", req.Method),
				logger.String("uri", req.RequestURI),
				logger.Int("status", res.Status),
				logger.Duration("latency", latency),
				logger.String("user_id", userID),
				logger.String("email", email),
			)

			return err
		}
	}
}
