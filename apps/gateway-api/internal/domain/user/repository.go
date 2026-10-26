package user

import (
	"context"
)

// Repository defines the persistence port for the User aggregate.
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}
