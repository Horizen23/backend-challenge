package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	pb "backend-challenge/apps/common/proto/pb"
	domainUser "backend-challenge/apps/user-service/internal/domain/user"
	grpcMapperGenerated "backend-challenge/apps/user-service/internal/presentation/grpc/mapper/generated"
	mockUseCase "backend-challenge/apps/user-service/tests/mocks/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupGRPCTest(t *testing.T, uc *mockUseCase.MockUserUseCase) (pb.UserServiceClient, func()) {
	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	mapper := &grpcMapperGenerated.GrpcServerMapperImpl{}
	userServer := NewUserGrpcServer(uc, mapper)
	RegisterService(server, userServer)

	go func() {
		_ = server.Serve(lis)
	}()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := pb.NewUserServiceClient(conn)
	cleanup := func() {
		_ = conn.Close()
		server.Stop()
		_ = lis.Close()
	}

	return client, cleanup
}

func TestUserGrpcServer_Operations(t *testing.T) {
	mockUC := mockUseCase.NewMockUserUseCase(t)
	client, cleanup := setupGRPCTest(t, mockUC)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()
	expectedUser := &domainUser.User{
		ID:        "user-123",
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: now,
	}

	t.Run("CreateUser_Success", func(t *testing.T) {
		mockUC.EXPECT().
			CreateUser(mock.Anything, "John Doe", "john@example.com", "password123").
			Return(expectedUser, nil).
			Once()

		resp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "password123",
		})

		require.NoError(t, err)
		assert.Equal(t, "user-123", resp.Id)
		assert.Equal(t, "John Doe", resp.Name)
	})

	t.Run("GetUser_Success", func(t *testing.T) {
		mockUC.EXPECT().
			GetUserByID(mock.Anything, "user-123").
			Return(expectedUser, nil).
			Once()

		resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: "user-123"})
		require.NoError(t, err)
		assert.Equal(t, "user-123", resp.Id)
	})

	t.Run("ListUsers_Success", func(t *testing.T) {
		mockUC.EXPECT().
			ListUsers(mock.Anything).
			Return([]*domainUser.User{expectedUser}, nil).
			Once()

		resp, err := client.ListUsers(ctx, &pb.ListUsersRequest{})
		require.NoError(t, err)
		assert.Len(t, resp.Users, 1)
	})

	t.Run("CountUsers_Success", func(t *testing.T) {
		mockUC.EXPECT().
			CountUsers(mock.Anything).
			Return(int64(42), nil).
			Once()

		resp, err := client.CountUsers(ctx, &pb.CountUsersRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(42), resp.Count)
	})

	t.Run("VerifyPassword_Success", func(t *testing.T) {
		mockUC.EXPECT().
			VerifyPassword(mock.Anything, "john@example.com", "password123").
			Return(expectedUser, true, nil).
			Once()

		resp, err := client.VerifyPassword(ctx, &pb.VerifyPasswordRequest{
			Email:    "john@example.com",
			Password: "password123",
		})
		require.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, "user-123", resp.User.Id)
	})

	t.Run("UpdateUser_Success", func(t *testing.T) {
		name := "Updated Name"
		mockUC.EXPECT().
			UpdateUser(mock.Anything, "user-123", &name, (*string)(nil)).
			Return(expectedUser, nil).
			Once()

		resp, err := client.UpdateUser(ctx, &pb.UpdateUserRequest{
			Id:   "user-123",
			Name: &name,
		})
		require.NoError(t, err)
		assert.Equal(t, "user-123", resp.Id)
	})

	t.Run("UpdateUser_EmptyID_ReturnsInvalidArgument", func(t *testing.T) {
		_, err := client.UpdateUser(ctx, &pb.UpdateUserRequest{
			Id: "",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user id is required")
	})

	t.Run("DeleteUser_Success", func(t *testing.T) {
		mockUC.EXPECT().
			DeleteUser(mock.Anything, "user-123").
			Return(nil).
			Once()

		resp, err := client.DeleteUser(ctx, &pb.DeleteUserRequest{
			Id: "user-123",
		})
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("DeleteUser_EmptyID_ReturnsInvalidArgument", func(t *testing.T) {
		_, err := client.DeleteUser(ctx, &pb.DeleteUserRequest{
			Id: "   ",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user id is required")
	})
}
