package grpc

import (
	"context"
	"errors"
	"strings"

	domainUser "backend-challenge/apps/user-service/internal/domain/user"
	grpcMapper "backend-challenge/apps/user-service/internal/presentation/grpc/mapper"
	"backend-challenge/apps/user-service/internal/usecase"
	pb "backend-challenge/apps/common/proto/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGrpcServer struct {
	pb.UnimplementedUserServiceServer
	useCase usecase.UserUseCase
	mapper  grpcMapper.GrpcServerMapper
}

func NewUserGrpcServer(useCase usecase.UserUseCase, mapper grpcMapper.GrpcServerMapper) *UserGrpcServer {
	return &UserGrpcServer{
		useCase: useCase,
		mapper:  mapper,
	}
}

func (s *UserGrpcServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	user, err := s.useCase.CreateUser(ctx, req.GetName(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, domainUser.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, domainUser.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.mapper.ToPbUser(user), nil
}

func (s *UserGrpcServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := s.useCase.GetUserByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domainUser.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, domainUser.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.mapper.ToPbUser(user), nil
}

func (s *UserGrpcServer) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.UserResponse, error) {
	user, err := s.useCase.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, domainUser.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, domainUser.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.mapper.ToPbUser(user), nil
}

func (s *UserGrpcServer) ListUsers(ctx context.Context, _ *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := s.useCase.ListUsers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListUsersResponse{Users: s.mapper.ToPbUserList(users)}, nil
}

func (s *UserGrpcServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	var namePtr, emailPtr *string
	if req.Name != nil {
		namePtr = req.Name
	}
	if req.Email != nil {
		emailPtr = req.Email
	}

	user, err := s.useCase.UpdateUser(ctx, req.GetId(), namePtr, emailPtr)
	if err != nil {
		if errors.Is(err, domainUser.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, domainUser.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		if errors.Is(err, domainUser.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.mapper.ToPbUser(user), nil
}

func (s *UserGrpcServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	err := s.useCase.DeleteUser(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domainUser.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, domainUser.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

func (s *UserGrpcServer) CountUsers(ctx context.Context, _ *pb.CountUsersRequest) (*pb.CountUsersResponse, error) {
	count, err := s.useCase.CountUsers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CountUsersResponse{Count: count}, nil
}

func (s *UserGrpcServer) VerifyPassword(ctx context.Context, req *pb.VerifyPasswordRequest) (*pb.VerifyPasswordResponse, error) {
	user, valid, err := s.useCase.VerifyPassword(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, domainUser.ErrInvalidCredentials) {
			return &pb.VerifyPasswordResponse{Valid: false}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !valid || user == nil {
		return &pb.VerifyPasswordResponse{Valid: false}, nil
	}

	return &pb.VerifyPasswordResponse{
		Valid: true,
		User:  s.mapper.ToPbUser(user),
	}, nil
}
