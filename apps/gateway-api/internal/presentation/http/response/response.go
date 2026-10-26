package response

import (
	"errors"
	"net/http"

	domainUser "backend-challenge/apps/gateway-api/internal/domain/user"
	"backend-challenge/apps/common/pkg/ctxutils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// Response is the standard envelope for generic API responses
type Response[T any] struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`
	Data      T      `json:"data,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ErrorDetail describes specific validation or field-level issues
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// APIError is the standard JSON structure for error responses
type APIError struct {
	RequestID string         `json:"request_id,omitempty"`
	Success   bool           `json:"success"`
	Status    int            `json:"status"`
	Title     string         `json:"title"`
	Detail    string         `json:"detail"`
	Errors    []*ErrorDetail `json:"errors,omitempty"`
}

func (e *APIError) Error() string {
	return e.Detail
}

// Success sends a generic JSON response with data
func Success[T any](c echo.Context, status int, data T) error {
	reqID := ctxutils.GetRequestID(c.Request().Context())
	return c.JSON(status, Response[T]{
		RequestID: reqID,
		Success:   true,
		Data:      data,
	})
}

// Message sends a generic JSON response with a string message (e.g. for deletes)
func Message(c echo.Context, status int, message string) error {
	reqID := ctxutils.GetRequestID(c.Request().Context())
	return c.JSON(status, Response[any]{
		RequestID: reqID,
		Success:   true,
		Message:   message,
	})
}

// MapError translates domain, echo, or validation errors into a standard APIError
func MapError(err error, reqID string) *APIError {
	var valErr validator.ValidationErrors
	if errors.As(err, &valErr) {
		var details []*ErrorDetail
		for _, fe := range valErr {
			details = append(details, &ErrorDetail{
				Field:   fe.Field(),
				Message: fe.Tag(),
				Value:   fe.Value(),
			})
		}
		return &APIError{
			RequestID: reqID,
			Success:   false,
			Status:    http.StatusBadRequest,
			Title:     "Validation Error",
			Detail:    "Request input validation failed",
			Errors:    details,
		}
	}

	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		msg := "unknown error"
		if m, ok := echoErr.Message.(string); ok {
			msg = m
		}
		return &APIError{
			RequestID: reqID,
			Success:   false,
			Status:    echoErr.Code,
			Title:     http.StatusText(echoErr.Code),
			Detail:    msg,
		}
	}

	switch {
	case errors.Is(err, domainUser.ErrUserNotFound):
		return &APIError{RequestID: reqID, Success: false, Status: http.StatusNotFound, Title: "Not Found", Detail: err.Error()}
	case errors.Is(err, domainUser.ErrEmailAlreadyExists):
		return &APIError{RequestID: reqID, Success: false, Status: http.StatusConflict, Title: "Conflict", Detail: err.Error()}
	case errors.Is(err, domainUser.ErrInvalidCredentials):
		return &APIError{RequestID: reqID, Success: false, Status: http.StatusUnauthorized, Title: "Unauthorized", Detail: err.Error()}
	case errors.Is(err, domainUser.ErrInvalidInput):
		return &APIError{RequestID: reqID, Success: false, Status: http.StatusBadRequest, Title: "Bad Request", Detail: err.Error()}
	case errors.Is(err, domainUser.ErrUnauthorized):
		return &APIError{RequestID: reqID, Success: false, Status: http.StatusUnauthorized, Title: "Unauthorized", Detail: err.Error()}
	default:
		return &APIError{RequestID: reqID, Success: false, Status: http.StatusInternalServerError, Title: "Internal Server Error", Detail: err.Error()}
	}
}
