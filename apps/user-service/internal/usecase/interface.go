package usecase

import (
	"context"

	domainUser "backend-challenge/apps/user-service/internal/domain/user"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, name, email, password string) (*domainUser.User, error)
	GetUserByID(ctx context.Context, id string) (*domainUser.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domainUser.User, error)
	ListUsers(ctx context.Context) ([]*domainUser.User, error)
	UpdateUser(ctx context.Context, id string, name, email *string) (*domainUser.User, error)
	DeleteUser(ctx context.Context, id string) error
	CountUsers(ctx context.Context) (int64, error)
	VerifyPassword(ctx context.Context, email, password string) (*domainUser.User, bool, error)
}
