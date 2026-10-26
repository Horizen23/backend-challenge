package usecase_test

import (
	"context"
	"testing"
	"time"

	domainUser "backend-challenge/apps/user-service/internal/domain/user"
	"backend-challenge/apps/user-service/internal/usecase"
	mockUser "backend-challenge/apps/user-service/tests/mocks/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_CreateUser", func(t *testing.T) {
		repo := mockUser.NewMockRepository(t)
		svc := usecase.NewUserService(repo)

		repo.EXPECT().
			GetByEmail(ctx, "test@example.com").
			Return(nil, domainUser.ErrUserNotFound).
			Once()

		repo.EXPECT().
			Create(ctx, mock.MatchedBy(func(u *domainUser.User) bool {
				return u.Email == "test@example.com" && u.Name == "Test User"
			})).
			Run(func(_ context.Context, u *domainUser.User) {
				u.ID = "generated-id-123"
				u.CreatedAt = time.Now().UTC()
			}).
			Return(nil).
			Once()

		user, err := svc.CreateUser(ctx, "Test User", "test@example.com", "password123")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
	})

	t.Run("Duplicate_Email_Returns_ErrEmailAlreadyExists", func(t *testing.T) {
		repo := mockUser.NewMockRepository(t)
		svc := usecase.NewUserService(repo)

		repo.EXPECT().
			GetByEmail(ctx, "dup@example.com").
			Return(&domainUser.User{ID: "existing-id", Email: "dup@example.com"}, nil).
			Once()

		user, err := svc.CreateUser(ctx, "Test User", "dup@example.com", "password123")
		assert.ErrorIs(t, err, domainUser.ErrEmailAlreadyExists)
		assert.Nil(t, user)
	})
}

func TestUserService_Operations(t *testing.T) {
	ctx := context.Background()
	repo := mockUser.NewMockRepository(t)
	svc := usecase.NewUserService(repo)
	expectedUser := &domainUser.User{ID: "user-123", Name: "Test User", Email: "test@example.com"}

	t.Run("GetUserByID_Success", func(t *testing.T) {
		repo.EXPECT().GetByID(ctx, "user-123").Return(expectedUser, nil).Once()
		user, err := svc.GetUserByID(ctx, "user-123")
		require.NoError(t, err)
		assert.Equal(t, "user-123", user.ID)
	})

	t.Run("ListUsers_Success", func(t *testing.T) {
		repo.EXPECT().List(ctx).Return([]*domainUser.User{expectedUser}, nil).Once()
		users, err := svc.ListUsers(ctx)
		require.NoError(t, err)
		assert.Len(t, users, 1)
	})

	t.Run("CountUsers_Success", func(t *testing.T) {
		repo.EXPECT().Count(ctx).Return(int64(10), nil).Once()
		count, err := svc.CountUsers(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(10), count)
	})

	t.Run("DeleteUser_Success", func(t *testing.T) {
		repo.EXPECT().Delete(ctx, "user-123").Return(nil).Once()
		err := svc.DeleteUser(ctx, "user-123")
		require.NoError(t, err)
	})
}
