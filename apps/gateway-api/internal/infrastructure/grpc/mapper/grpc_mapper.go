package mapper

import (
	"strings"
	"time"

	pb "backend-challenge/apps/common/proto/pb"
	domainUser "backend-challenge/apps/gateway-api/internal/domain/user"
	"backend-challenge/apps/gateway-api/internal/presentation/http/dto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func StringToTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func TimeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

func CleanEmail(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func CleanString(s string) string {
	return strings.TrimSpace(s)
}

func CleanStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	return &trimmed
}

func CleanEmailPtr(s *string) *string {
	if s == nil {
		return nil
	}
	cleaned := strings.TrimSpace(strings.ToLower(*s))
	return &cleaned
}

func MapGrpcError(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return domainUser.ErrUserNotFound
	case codes.AlreadyExists:
		return domainUser.ErrEmailAlreadyExists
	case codes.InvalidArgument:
		return domainUser.ErrInvalidInput
	case codes.Unauthenticated:
		return domainUser.ErrUnauthorized
	default:
		return err
	}
}

// goverter:converter
// goverter:extend StringToTime
// goverter:extend TimeToString
// goverter:extend CleanEmail
// goverter:extend CleanString
// goverter:extend CleanStringPtr
// goverter:extend CleanEmailPtr
// goverter:ignoreUnexported
// goverter:useZeroValueOnPointerInconsistency
type GrpcMapper interface {
	// goverter:map Id ID
	// goverter:ignore Password
	ToDomain(pbUser *pb.UserResponse) *domainUser.User
	ToDomainList(pbUsers []*pb.UserResponse) []*domainUser.User

	// goverter:map Name Name | CleanString
	// goverter:map Email Email | CleanEmail
	ToPbCreateUserRequest(req dto.CreateUserRequest) *pb.CreateUserRequest

	// goverter:map Name Name | CleanString
	// goverter:map Email Email | CleanEmail
	ToPbRegisterRequest(req dto.RegisterRequest) *pb.CreateUserRequest

	// goverter:map Email Email | CleanEmail
	ToPbVerifyPasswordRequest(req dto.LoginRequest) *pb.VerifyPasswordRequest

	// goverter:map Name Name | CleanStringPtr
	// goverter:map Email Email | CleanEmailPtr
	// goverter:ignore Id
	ToPbUpdateUserRequest(req dto.UpdateUserRequest) *pb.UpdateUserRequest
}
