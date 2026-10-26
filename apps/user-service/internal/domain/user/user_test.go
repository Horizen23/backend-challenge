package user_test

import (
	"testing"

	domainUser "backend-challenge/apps/user-service/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_NewUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u, err := domainUser.NewUser("Alice", "alice@example.com", "hashed_secret")
		require.NoError(t, err)
		assert.Equal(t, "Alice", u.Name)
		assert.Equal(t, "alice@example.com", u.Email)
		assert.Equal(t, "hashed_secret", u.Password)
		assert.False(t, u.CreatedAt.IsZero())
	})

	t.Run("EmptyName_ReturnsError", func(t *testing.T) {
		_, err := domainUser.NewUser("  ", "alice@example.com", "hashed_secret")
		assert.ErrorIs(t, err, domainUser.ErrInvalidInput)
	})

	t.Run("InvalidEmail_ReturnsError", func(t *testing.T) {
		_, err := domainUser.NewUser("Alice", "invalid-email", "hashed_secret")
		assert.ErrorIs(t, err, domainUser.ErrInvalidInput)
	})

	t.Run("EmptyPassword_ReturnsError", func(t *testing.T) {
		_, err := domainUser.NewUser("Alice", "alice@example.com", "")
		assert.ErrorIs(t, err, domainUser.ErrInvalidInput)
	})
}

func TestUser_UpdateProfile(t *testing.T) {
	u, err := domainUser.NewUser("Alice", "alice@example.com", "hashed_secret")
	require.NoError(t, err)

	t.Run("Success_UpdateNameAndEmail", func(t *testing.T) {
		newName := "Alice Updated"
		newEmail := "alice_new@example.com"
		err := u.UpdateProfile(&newName, &newEmail)
		require.NoError(t, err)
		assert.Equal(t, "Alice Updated", u.Name)
		assert.Equal(t, "alice_new@example.com", u.Email)
	})

	t.Run("InvalidEmail_Fails", func(t *testing.T) {
		badEmail := "bad-format"
		err := u.UpdateProfile(nil, &badEmail)
		assert.ErrorIs(t, err, domainUser.ErrInvalidInput)
	})
}
