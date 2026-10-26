package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-challenge/apps/gateway-api/internal/config"
	domainUser "backend-challenge/apps/gateway-api/internal/domain/user"
	"backend-challenge/apps/gateway-api/internal/infrastructure/auth"
	"backend-challenge/apps/common/pkg/logger"
	httpPresentation "backend-challenge/apps/gateway-api/internal/presentation/http"
	"backend-challenge/apps/gateway-api/internal/presentation/http/dto"
	"backend-challenge/apps/gateway-api/internal/presentation/http/handlers"
	"backend-challenge/apps/gateway-api/internal/presentation/http/mapper"
	"backend-challenge/apps/gateway-api/internal/presentation/http/router"
	mockUseCase "backend-challenge/apps/gateway-api/tests/mocks/usecase"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestServer(t *testing.T) (*echo.Echo, *mockUseCase.MockUserUseCase, auth.TokenProvider) {
	cfg := &config.Config{}
	log := logger.NewLogger()
	jwtProvider := auth.NewJWTProvider("my-test-secret-32-bytes-long-key!", time.Hour)
	uc := mockUseCase.NewMockUserUseCase(t)
	presMapper := &mapper.PresentationMapperImpl{}

	userHandler := handlers.NewUserHandler(uc, presMapper)
	r := router.NewRouter(userHandler)

	e := httpPresentation.NewServer(cfg, r, log, jwtProvider)
	return e, uc, jwtProvider
}

func TestAuthEndpoints(t *testing.T) {
	e, uc, _ := setupTestServer(t)

	t.Run("Register Success", func(t *testing.T) {
		uc.EXPECT().Register(mock.Anything, dto.RegisterRequest{
			Name:     "Tester",
			Email:    "test@example.com",
			Password: "password123",
		}).Return(&domainUser.User{
			ID:        "user-100",
			Name:      "Tester",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
		}, "sample-jwt-token", nil).Once()

		body := `{"name":"Tester","email":"test@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "sample-jwt-token", data["token"])
	})

	t.Run("Login Success", func(t *testing.T) {
		uc.EXPECT().Login(mock.Anything, dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}).Return(&domainUser.User{
			ID:    "user-100",
			Email: "test@example.com",
		}, "sample-jwt-token", nil).Once()

		body := `{"email":"test@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestProtectedUserRoutes(t *testing.T) {
	e, uc, jwtProvider := setupTestServer(t)

	t.Run("Unauthorized without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Authorized with valid token", func(t *testing.T) {
		uc.EXPECT().ListUsers(mock.Anything).Return([]*domainUser.User{
			{ID: "user-1", Name: "User 1", Email: "u1@example.com"},
		}, nil).Once()

		token, _ := jwtProvider.GenerateToken("user-1", "test@example.com")
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
