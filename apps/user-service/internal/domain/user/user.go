package user

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidInput       = errors.New("invalid input data")
	ErrUnauthorized       = errors.New("unauthorized access")
)

// User is the domain aggregate root / entity representing a user in the system.
// It is completely decoupled from any database driver, ORM, or wire format tags.
type User struct {
	ID        string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

// NewUser encapsulates business invariants for user instantiation
func NewUser(name, email, hashedPassword string) (*User, error) {
	trimmedName := strings.TrimSpace(name)
	trimmedEmail := strings.TrimSpace(strings.ToLower(email))

	if trimmedName == "" {
		return nil, ErrInvalidInput
	}
	if !IsValidEmail(trimmedEmail) {
		return nil, ErrInvalidInput
	}
	if strings.TrimSpace(hashedPassword) == "" {
		return nil, ErrInvalidInput
	}

	return &User{
		Name:      trimmedName,
		Email:     trimmedEmail,
		Password:  hashedPassword,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// UpdateProfile encapsulates entity-level business invariants for profile updates
func (u *User) UpdateProfile(name, email *string) error {
	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if trimmedName == "" {
			return ErrInvalidInput
		}
		u.Name = trimmedName
	}

	if email != nil {
		trimmedEmail := strings.TrimSpace(strings.ToLower(*email))
		if !IsValidEmail(trimmedEmail) {
			return ErrInvalidInput
		}
		u.Email = trimmedEmail
	}

	return nil
}

// IsValidEmail verifies standard RFC 5322 email syntax
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
