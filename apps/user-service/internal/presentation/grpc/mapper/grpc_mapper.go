package mapper

import (
	"time"

	pb "backend-challenge/apps/common/proto/pb"
	domainUser "backend-challenge/apps/user-service/internal/domain/user"
)

func TimeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

func StringToTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// goverter:converter
// goverter:extend TimeToString
// goverter:extend StringToTime
// goverter:ignoreUnexported
// goverter:useZeroValueOnPointerInconsistency
type GrpcServerMapper interface {
	// goverter:map ID Id
	ToPbUser(u *domainUser.User) *pb.UserResponse
	ToPbUserList(users []*domainUser.User) []*pb.UserResponse
}
