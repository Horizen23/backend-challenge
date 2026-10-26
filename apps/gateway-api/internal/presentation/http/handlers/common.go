package handlers

import (
	"backend-challenge/apps/gateway-api/internal/presentation/http/response"
	"github.com/labstack/echo/v4"
)

type Response[T any] = response.Response[T]
type APIError = response.APIError
type ErrorDetail = response.ErrorDetail

func MapError(err error, reqID string) *APIError {
	return response.MapError(err, reqID)
}

func SendSuccess[T any](c echo.Context, status int, data T) error {
	return response.Success(c, status, data)
}

func SendMessage(c echo.Context, status int, message string) error {
	return response.Message(c, status, message)
}
