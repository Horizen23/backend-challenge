package usecase

import (
	"context"
	"strings"

	domainUser "backend-challenge/apps/user-service/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo domainUser.Repository
}

func NewUserService(repo domainUser.Repository) UserUseCase {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context, name, email, password string) (*domainUser.User, error) {
	if len(password) < 6 {
		return nil, domainUser.ErrInvalidInput
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Instantiate Domain Entity enforcing all business validation invariants
	user, err := domainUser.NewUser(name, email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByEmail(ctx, user.Email)
	if err == nil && existing != nil {
		return nil, domainUser.ErrEmailAlreadyExists
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*domainUser.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domainUser.ErrInvalidInput
	}
	return s.repo.GetByID(ctx, id)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !domainUser.IsValidEmail(email) {
		return nil, domainUser.ErrInvalidInput
	}
	return s.repo.GetByEmail(ctx, email)
}

func (s *userService) ListUsers(ctx context.Context) ([]*domainUser.User, error) {
	return s.repo.List(ctx)
}

func (s *userService) UpdateUser(ctx context.Context, id string, name, email *string) (*domainUser.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domainUser.ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check email uniqueness if email is being updated
	if email != nil {
		trimmedEmail := strings.TrimSpace(strings.ToLower(*email))
		if trimmedEmail != user.Email {
			existing, err := s.repo.GetByEmail(ctx, trimmedEmail)
			if err == nil && existing != nil && existing.ID != user.ID {
				return nil, domainUser.ErrEmailAlreadyExists
			}
		}
	}

	// Delegate profile invariants & validation to Domain Entity
	if err := user.UpdateProfile(name, email); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return domainUser.ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}

func (s *userService) CountUsers(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

func (s *userService) VerifyPassword(ctx context.Context, email, password string) (*domainUser.User, bool, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !domainUser.IsValidEmail(email) || password == "" {
		return nil, false, domainUser.ErrInvalidCredentials
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, false, domainUser.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, false, nil
	}

	return user, true, nil
}
