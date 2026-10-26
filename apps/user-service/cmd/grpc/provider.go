package main

import (
	"context"
	"time"

	"backend-challenge/apps/user-service/internal/config"
	domainUser "backend-challenge/apps/user-service/internal/domain/user"
	"backend-challenge/apps/user-service/internal/infrastructure/datastore/mongodb"
	mongoMapper "backend-challenge/apps/user-service/internal/infrastructure/datastore/mongodb/mapper"
	"backend-challenge/apps/common/pkg/logger"
	grpcPresentation "backend-challenge/apps/user-service/internal/presentation/grpc"
	grpcMapper "backend-challenge/apps/user-service/internal/presentation/grpc/mapper"
	grpcMapperGenerated "backend-challenge/apps/user-service/internal/presentation/grpc/mapper/generated"
	"backend-challenge/apps/user-service/internal/usecase"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func ProvideLogger() logger.Logger {
	return logger.NewLogger()
}

func ProvideMongoClient(cfg *config.Config) (*mongo.Client, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.Database.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
	}

	return client, cleanup, nil
}

func ProvideMongoDatabase(cfg *config.Config, client *mongo.Client) *mongo.Database {
	return client.Database(cfg.Database.DBName)
}

func ProvideMongoMapper() mongoMapper.MongoMapper {
	return &mongoMapper.MongoMapperImpl{}
}

func ProvideGrpcServerMapper() grpcMapper.GrpcServerMapper {
	return &grpcMapperGenerated.GrpcServerMapperImpl{}
}

func ProvideUserRepository(db *mongo.Database) domainUser.Repository {
	return mongodb.NewUserRepository(db)
}

func ProvideGRPCServer(uc usecase.UserUseCase, mapper grpcMapper.GrpcServerMapper, log logger.Logger) *grpc.Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcPresentation.MetadataServerInterceptor(log)),
	)
	userGrpcServer := grpcPresentation.NewUserGrpcServer(uc, mapper)
	grpcPresentation.RegisterService(server, userGrpcServer)
	return server
}

var CommonSet = wire.NewSet(
	ProvideLogger,
	ProvideMongoClient,
	ProvideMongoDatabase,
	ProvideMongoMapper,
	ProvideGrpcServerMapper,
	ProvideUserRepository,
	usecase.NewUserService,
	ProvideGRPCServer,
)
