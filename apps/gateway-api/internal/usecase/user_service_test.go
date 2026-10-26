package usecase_test

import (
	"context"
	"testing"
	"time"

	pb "backend-challenge/apps/common/proto/pb"
	domainUser "backend-challenge/apps/gateway-api/internal/domain/user"
	grpcMapperGenerated "backend-challenge/apps/gateway-api/internal/infrastructure/grpc/mapper/generated"
	"backend-challenge/apps/gateway-api/internal/presentation/http/dto"
	"backend-challenge/apps/gateway-api/internal/usecase"
	mockAuth "backend-challenge/apps/gateway-api/tests/mocks/infrastructure/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockUserServiceClient struct {
	mock.Mock
}

func (m *mockUserServiceClient) CreateUser(ctx context.Context, in *pb.CreateUserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *mockUserServiceClient) GetUser(ctx context.Context, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *mockUserServiceClient) GetUserByEmail(ctx context.Context, in *pb.GetUserByEmailRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *mockUserServiceClient) ListUsers(ctx context.Context, in *pb.ListUsersRequest, opts ...grpc.CallOption) (*pb.ListUsersResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ListUsersResponse), args.Error(1)
}

func (m *mockUserServiceClient) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *mockUserServiceClient) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest, opts ...grpc.CallOption) (*pb.DeleteUserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteUserResponse), args.Error(1)
}

func (m *mockUserServiceClient) CountUsers(ctx context.Context, in *pb.CountUsersRequest, opts ...grpc.CallOption) (*pb.CountUsersResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CountUsersResponse), args.Error(1)
}

func (m *mockUserServiceClient) VerifyPassword(ctx context.Context, in *pb.VerifyPasswordRequest, opts ...grpc.CallOption) (*pb.VerifyPasswordResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.VerifyPasswordResponse), args.Error(1)
}

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()
	mapper := &grpcMapperGenerated.GrpcMapperImpl{}

	t.Run("Success_Register", func(t *testing.T) {
		grpcClient := new(mockUserServiceClient)
		tokenProvider := mockAuth.NewMockTokenProvider(t)
		svc := usecase.NewUserService(grpcClient, tokenProvider, mapper)

		grpcClient.On("CreateUser", ctx, &pb.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
		}).Return(&pb.UserResponse{
			Id:        "user-123",
			Name:      "Test User",
			Email:     "test@example.com",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}, nil).Once()

		tokenProvider.EXPECT().
			GenerateToken("user-123", "test@example.com").
			Return("mocked-jwt-token", nil).
			Once()

		user, token, err := svc.Register(ctx, dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
		})

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "user-123", user.ID)
		assert.Equal(t, "mocked-jwt-token", token)
		grpcClient.AssertExpectations(t)
	})

	t.Run("Duplicate_Email_Returns_ErrEmailAlreadyExists", func(t *testing.T) {
		grpcClient := new(mockUserServiceClient)
		tokenProvider := mockAuth.NewMockTokenProvider(t)
		svc := usecase.NewUserService(grpcClient, tokenProvider, mapper)

		grpcClient.On("CreateUser", ctx, mock.Anything).
			Return(nil, status.Error(codes.AlreadyExists, "email already exists")).
			Once()

		user, token, err := svc.Register(ctx, dto.RegisterRequest{
			Name:     "Test User",
			Email:    "dup@example.com",
			Password: "password123",
		})

		assert.ErrorIs(t, err, domainUser.ErrEmailAlreadyExists)
		assert.Nil(t, user)
		assert.Empty(t, token)
	})
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()
	mapper := &grpcMapperGenerated.GrpcMapperImpl{}

	t.Run("Success_Login", func(t *testing.T) {
		grpcClient := new(mockUserServiceClient)
		tokenProvider := mockAuth.NewMockTokenProvider(t)
		svc := usecase.NewUserService(grpcClient, tokenProvider, mapper)

		grpcClient.On("VerifyPassword", ctx, &pb.VerifyPasswordRequest{
			Email:    "test@example.com",
			Password: "password123",
		}).Return(&pb.VerifyPasswordResponse{
			Valid: true,
			User: &pb.UserResponse{
				Id:        "user-123",
				Name:      "Test User",
				Email:     "test@example.com",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
		}, nil).Once()

		tokenProvider.EXPECT().
			GenerateToken("user-123", "test@example.com").
			Return("mocked-jwt-token", nil).
			Once()

		user, token, err := svc.Login(ctx, dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "user-123", user.ID)
		assert.Equal(t, "mocked-jwt-token", token)
	})
}

func TestUserService_CRUD(t *testing.T) {
	ctx := context.Background()
	mapper := &grpcMapperGenerated.GrpcMapperImpl{}
	grpcClient := new(mockUserServiceClient)
	tokenProvider := mockAuth.NewMockTokenProvider(t)
	svc := usecase.NewUserService(grpcClient, tokenProvider, mapper)

	t.Run("GetUserByID", func(t *testing.T) {
		grpcClient.On("GetUser", ctx, &pb.GetUserRequest{Id: "user-123"}).
			Return(&pb.UserResponse{Id: "user-123", Name: "User"}, nil).
			Once()

		user, err := svc.GetUserByID(ctx, "user-123")
		require.NoError(t, err)
		assert.Equal(t, "user-123", user.ID)
	})

	t.Run("CountUsers", func(t *testing.T) {
		grpcClient.On("CountUsers", ctx, &pb.CountUsersRequest{}).
			Return(&pb.CountUsersResponse{Count: 50}, nil).
			Once()

		count, err := svc.CountUsers(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(50), count)
	})
}
