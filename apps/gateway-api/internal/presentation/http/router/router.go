package router

import (
	"backend-challenge/apps/gateway-api/internal/presentation/http/handlers"

	"github.com/labstack/echo/v4"
)

type Router struct {
	userHandler *handlers.UserHandler
}

func NewRouter(userHandler *handlers.UserHandler) *Router {
	return &Router{
		userHandler: userHandler,
	}
}

func (r *Router) RegisterAll(e *echo.Echo, jwtMw echo.MiddlewareFunc) {
	// Public Healthcheck
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	api := e.Group("/api")

	// Public Auth
	authGroup := api.Group("/auth")
	authGroup.POST("/register", r.userHandler.Register)
	authGroup.POST("/login", r.userHandler.Login)

	// Protected User Endpoints
	usersGroup := api.Group("/users", jwtMw)
	usersGroup.POST("", r.userHandler.CreateUser)
	usersGroup.GET("", r.userHandler.ListUsers)
	usersGroup.GET("/:id", r.userHandler.GetUserByID)
	usersGroup.PUT("/:id", r.userHandler.UpdateUser)
	usersGroup.DELETE("/:id", r.userHandler.DeleteUser)
}
