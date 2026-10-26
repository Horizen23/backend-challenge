package middlewares

import (
	"backend-challenge/apps/common/pkg/ctxutils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequestIDMiddleware extracts the "X-Request-ID" header from the incoming HTTP request.
// If not present, a new UUID is generated and set into both the HTTP response header
// and the request context for distributed tracing and cross-service observability.
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get(string(ctxutils.RequestIDKey))
			if reqID == "" {
				reqID = uuid.New().String()
			}

			// Propagate Request ID back to the client via response header
			c.Response().Header().Set(string(ctxutils.RequestIDKey), reqID)

			// Store Request ID into standard context.Context
			ctx := ctxutils.SetRequestID(c.Request().Context(), reqID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
