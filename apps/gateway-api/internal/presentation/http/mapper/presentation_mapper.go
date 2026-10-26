package mapper

import (
	domainUser "backend-challenge/apps/gateway-api/internal/domain/user"
	"backend-challenge/apps/gateway-api/internal/presentation/http/dto"
	"time"
)

func TimeToTime(t time.Time) time.Time {
	return t
}

// goverter:converter
// goverter:extend TimeToTime
// goverter:useZeroValueOnPointerInconsistency
type PresentationMapper interface {
	ToUserResponse(u *domainUser.User) dto.UserResponse
	ToUserResponseList(users []*domainUser.User) []dto.UserResponse
}
