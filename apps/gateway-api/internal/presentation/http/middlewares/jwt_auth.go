package middlewares

import (
	"strings"

	"backend-challenge/apps/gateway-api/internal/infrastructure/auth"
	"backend-challenge/apps/common/pkg/ctxutils"

	"github.com/labstack/echo/v4"
)

// JWTAuthMiddleware intercepts incoming requests on protected endpoints,
// validates the "Authorization: Bearer <token>" header using HMAC HS256,
// and injects the authenticated UserID and Email into the request context.
func JWTAuthMiddleware(provider auth.TokenProvider) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := strings.TrimSpace(c.Request().Header.Get("Authorization"))
			if authHeader == "" {
				return echo.NewHTTPError(401, "authorization header is required")
			}

			var tokenStr string
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				tokenStr = strings.TrimSpace(authHeader[7:])
			} else if !strings.Contains(authHeader, " ") {
				// Handle raw token directly (useful for Swagger UI apikey / raw curl)
				tokenStr = authHeader
			} else {
				return echo.NewHTTPError(401, "invalid authorization token format")
			}

			if tokenStr == "" {
				return echo.NewHTTPError(401, "authorization token cannot be empty")
			}

			claims, err := provider.ValidateToken(tokenStr)
			if err != nil {
				return echo.NewHTTPError(401, "invalid or expired token")
			}

			// Store authenticated user claims in context for downstream handlers and gRPC metadata propagation
			ctx := ctxutils.SetAuthUser(c.Request().Context(), claims.UserID, claims.Email)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
