package usecase

import (
	"context"
	"strings"

	domainUser "backend-challenge/apps/gateway-api/internal/domain/user"
	"backend-challenge/apps/gateway-api/internal/infrastructure/auth"
	grpcMapper "backend-challenge/apps/gateway-api/internal/infrastructure/grpc/mapper"
	"backend-challenge/apps/gateway-api/internal/presentation/http/dto"
	pb "backend-challenge/apps/common/proto/pb"
)

type UserUseCase interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*domainUser.User, string, error)
	Login(ctx context.Context, req dto.LoginRequest) (*domainUser.User, string, error)
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*domainUser.User, error)
	GetUserByID(ctx context.Context, id string) (*domainUser.User, error)
	ListUsers(ctx context.Context) ([]*domainUser.User, error)
	UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (*domainUser.User, error)
	DeleteUser(ctx context.Context, id string) error
	CountUsers(ctx context.Context) (int64, error)
}

type userService struct {
	grpcClient    pb.UserServiceClient
	tokenProvider auth.TokenProvider
	mapper        grpcMapper.GrpcMapper
}

func NewUserService(
	grpcClient pb.UserServiceClient,
	tokenProvider auth.TokenProvider,
	mapper grpcMapper.GrpcMapper,
) UserUseCase {
	return &userService{
		grpcClient:    grpcClient,
		tokenProvider: tokenProvider,
		mapper:        mapper,
	}
}

func (s *userService) Register(ctx context.Context, req dto.RegisterRequest) (*domainUser.User, string, error) {
	if len(req.Password) < 6 {
		return nil, "", domainUser.ErrInvalidInput
	}

	// Validate business invariants via domain factory
	if _, err := domainUser.NewUser(req.Name, req.Email, "dummy_hash_for_validation"); err != nil {
		return nil, "", err
	}

	grpcReq := s.mapper.ToPbRegisterRequest(req)
	resp, err := s.grpcClient.CreateUser(ctx, grpcReq)
	if err != nil {
		return nil, "", grpcMapper.MapGrpcError(err)
	}

	u := s.mapper.ToDomain(resp)
	token, err := s.tokenProvider.GenerateToken(u.ID, u.Email)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (*domainUser.User, string, error) {
	if !domainUser.IsValidEmail(req.Email) || req.Password == "" {
		return nil, "", domainUser.ErrInvalidCredentials
	}

	grpcReq := s.mapper.ToPbVerifyPasswordRequest(req)
	resp, err := s.grpcClient.VerifyPassword(ctx, grpcReq)
	if err != nil {
		return nil, "", grpcMapper.MapGrpcError(err)
	}

	if !resp.GetValid() || resp.GetUser() == nil {
		return nil, "", domainUser.ErrInvalidCredentials
	}

	u := s.mapper.ToDomain(resp.GetUser())
	token, err := s.tokenProvider.GenerateToken(u.ID, u.Email)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

func (s *userService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*domainUser.User, error) {
	if len(req.Password) < 6 {
		return nil, domainUser.ErrInvalidInput
	}

	// Validate business invariants via domain factory
	if _, err := domainUser.NewUser(req.Name, req.Email, "dummy_hash_for_validation"); err != nil {
		return nil, err
	}

	grpcReq := s.mapper.ToPbCreateUserRequest(req)
	resp, err := s.grpcClient.CreateUser(ctx, grpcReq)
	if err != nil {
		return nil, grpcMapper.MapGrpcError(err)
	}
	return s.mapper.ToDomain(resp), nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*domainUser.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, domainUser.ErrInvalidInput
	}

	resp, err := s.grpcClient.GetUser(ctx, &pb.GetUserRequest{Id: id})
	if err != nil {
		return nil, grpcMapper.MapGrpcError(err)
	}
	return s.mapper.ToDomain(resp), nil
}

func (s *userService) ListUsers(ctx context.Context) ([]*domainUser.User, error) {
	resp, err := s.grpcClient.ListUsers(ctx, &pb.ListUsersRequest{})
	if err != nil {
		return nil, grpcMapper.MapGrpcError(err)
	}

	return s.mapper.ToDomainList(resp.GetUsers()), nil
}

func (s *userService) UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (*domainUser.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domainUser.ErrInvalidInput
	}

	// Validate field invariants using domain method
	dummyUser := &domainUser.User{}
	if err := dummyUser.UpdateProfile(req.Name, req.Email); err != nil {
		return nil, err
	}

	grpcReq := s.mapper.ToPbUpdateUserRequest(req)
	grpcReq.Id = id

	resp, err := s.grpcClient.UpdateUser(ctx, grpcReq)
	if err != nil {
		return nil, grpcMapper.MapGrpcError(err)
	}
	return s.mapper.ToDomain(resp), nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return domainUser.ErrInvalidInput
	}

	_, err := s.grpcClient.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
	if err != nil {
		return grpcMapper.MapGrpcError(err)
	}
	return nil
}

func (s *userService) CountUsers(ctx context.Context) (int64, error) {
	resp, err := s.grpcClient.CountUsers(ctx, &pb.CountUsersRequest{})
	if err != nil {
		return 0, grpcMapper.MapGrpcError(err)
	}
	return resp.GetCount(), nil
}
