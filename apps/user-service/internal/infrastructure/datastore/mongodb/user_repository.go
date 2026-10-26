package mongodb

import (
	"context"
	"errors"
	"time"

	domainUser "backend-challenge/apps/user-service/internal/domain/user"
	mongoMapper "backend-challenge/apps/user-service/internal/infrastructure/datastore/mongodb/mapper"
	"backend-challenge/apps/user-service/internal/infrastructure/datastore/mongodb/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	coll   *mongo.Collection
	mapper mongoMapper.MongoMapper
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	repo := &UserRepository{
		coll:   db.Collection("users"),
		mapper: &mongoMapper.MongoMapperImpl{},
	}
	repo.ensureIndexes(context.Background())
	return repo
}

func (r *UserRepository) ensureIndexes(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	idx := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, _ = r.coll.Indexes().CreateOne(ctx, idx)
}

func (r *UserRepository) Create(ctx context.Context, u *domainUser.User) error {
	m := r.mapper.ToModel(u)

	res, err := r.coll.InsertOne(ctx, m)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domainUser.ErrEmailAlreadyExists
		}
		return err
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		u.ID = oid.Hex()
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domainUser.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domainUser.ErrUserNotFound
	}

	var m model.UserModel
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&m)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domainUser.ErrUserNotFound
		}
		return nil, err
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	var m model.UserModel
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&m)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domainUser.ErrUserNotFound
		}
		return nil, err
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *UserRepository) List(ctx context.Context) ([]*domainUser.User, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []*model.UserModel
	for cursor.Next(ctx) {
		var m model.UserModel
		if err := cursor.Decode(&m); err != nil {
			return nil, err
		}
		models = append(models, &m)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return r.mapper.ToDomainList(models), nil
}

func (r *UserRepository) Update(ctx context.Context, u *domainUser.User) error {
	oid, err := primitive.ObjectIDFromHex(u.ID)
	if err != nil {
		return domainUser.ErrUserNotFound
	}

	update := bson.M{
		"$set": bson.M{
			"name":  u.Name,
			"email": u.Email,
		},
	}

	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domainUser.ErrEmailAlreadyExists
		}
		return err
	}

	if res.MatchedCount == 0 {
		return domainUser.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domainUser.ErrUserNotFound
	}

	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return domainUser.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{})
}
