package mapper

import (
	"time"

	domainUser "backend-challenge/apps/user-service/internal/domain/user"
	"backend-challenge/apps/user-service/internal/infrastructure/datastore/mongodb/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// goverter:converter
// goverter:output:file ./generated_mongo_mapper.go
// goverter:useZeroValueOnPointerInconsistency yes
// goverter:extend ObjectIDToString
// goverter:extend StringToObjectID
// goverter:extend TimeToTime
type MongoMapper interface {
	ToDomain(m *model.UserModel) *domainUser.User
	ToDomainList(models []*model.UserModel) []*domainUser.User
	ToModel(d *domainUser.User) *model.UserModel
}

func ObjectIDToString(oid primitive.ObjectID) string {
	if oid.IsZero() {
		return ""
	}
	return oid.Hex()
}

func StringToObjectID(s string) primitive.ObjectID {
	oid, _ := primitive.ObjectIDFromHex(s)
	return oid
}

func TimeToTime(t time.Time) time.Time {
	return t
}
