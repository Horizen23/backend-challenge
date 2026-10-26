package handlers

import (
	"net/http"

	"backend-challenge/apps/gateway-api/internal/presentation/http/dto"
	presentationMapper "backend-challenge/apps/gateway-api/internal/presentation/http/mapper"
	"backend-challenge/apps/gateway-api/internal/presentation/http/response"
	"backend-challenge/apps/gateway-api/internal/usecase"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	useCase usecase.UserUseCase
	mapper  presentationMapper.PresentationMapper
}

func NewUserHandler(useCase usecase.UserUseCase, mapper presentationMapper.PresentationMapper) *UserHandler {
	return &UserHandler{
		useCase: useCase,
		mapper:  mapper,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account and obtain a JWT bearer token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User Registration Info"
// @Success 201 {object} response.Response[dto.AuthResponse]
// @Failure 400 {object} response.APIError
// @Failure 409 {object} response.APIError
// @Failure 500 {object} response.APIError
// @Router /api/auth/register [post]
func (h *UserHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, token, err := h.useCase.Register(c.Request().Context(), req)
	if err != nil {
		return err
	}

	resp := dto.AuthResponse{
		Token: token,
		User:  h.mapper.ToUserResponse(user),
	}
	return response.Success(c, http.StatusCreated, resp)
}

// Login godoc
// @Summary User Login
// @Description Authenticate user credentials and return a signed JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Credentials"
// @Success 200 {object} response.Response[dto.AuthResponse]
// @Failure 400 {object} response.APIError
// @Failure 401 {object} response.APIError
// @Failure 500 {object} response.APIError
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, token, err := h.useCase.Login(c.Request().Context(), req)
	if err != nil {
		return err
	}

	resp := dto.AuthResponse{
		Token: token,
		User:  h.mapper.ToUserResponse(user),
	}
	return response.Success(c, http.StatusOK, resp)
}

// CreateUser godoc
// @Summary Create a user
// @Description Create a user record directly (Protected endpoint)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "User creation payload"
// @Success 201 {object} response.Response[dto.UserResponse]
// @Failure 400 {object} response.APIError
// @Failure 401 {object} response.APIError
// @Failure 409 {object} response.APIError
// @Failure 500 {object} response.APIError
// @Router /api/users [post]
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req dto.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.useCase.CreateUser(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusCreated, h.mapper.ToUserResponse(user))
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Retrieve single user by their unique identifier (Protected endpoint)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response[dto.UserResponse]
// @Failure 400 {object} response.APIError
// @Failure 401 {object} response.APIError
// @Failure 404 {object} response.APIError
// @Failure 500 {object} response.APIError
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUserByID(c echo.Context) error {
	id := c.Param("id")
	user, err := h.useCase.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusOK, h.mapper.ToUserResponse(user))
}

// ListUsers godoc
// @Summary List all users
// @Description Retrieve a list of all existing users (Protected endpoint)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response[[]dto.UserResponse]
// @Failure 401 {object} response.APIError
// @Failure 500 {object} response.APIError
// @Router /api/users [get]
func (h *UserHandler) ListUsers(c echo.Context) error {
	users, err := h.useCase.ListUsers(c.Request().Context())
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusOK, h.mapper.ToUserResponseList(users))
}

// UpdateUser godoc
// @Summary Update an existing user
// @Description Modify name or email for a user (Protected endpoint)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "Update user payload"
// @Success 200 {object} response.Response[dto.UserResponse]
// @Failure 400 {object} response.APIError
// @Failure 401 {object} response.APIError
// @Failure 404 {object} response.APIError
// @Failure 409 {object} response.APIError
// @Failure 500 {object} response.APIError
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c echo.Context) error {
	id := c.Param("id")
	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.useCase.UpdateUser(c.Request().Context(), id, req)
	if err != nil {
		return err
	}

	return response.Success(c, http.StatusOK, h.mapper.ToUserResponse(user))
}

// DeleteUser godoc
// @Summary Delete user by ID
// @Description Delete a user record permanently (Protected endpoint)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response[any]
// @Failure 400 {object} response.APIError
// @Failure 401 {object} response.APIError
// @Failure 404 {object} response.APIError
// @Failure 500 {object} response.APIError
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(c echo.Context) error {
	id := c.Param("id")
	if err := h.useCase.DeleteUser(c.Request().Context(), id); err != nil {
		return err
	}

	return response.Message(c, http.StatusOK, "user deleted successfully")
}
