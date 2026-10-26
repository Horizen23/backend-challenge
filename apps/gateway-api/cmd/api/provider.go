package main

import (
	"backend-challenge/apps/common/proto/pb"
	"backend-challenge/apps/gateway-api/internal/config"
	"backend-challenge/apps/gateway-api/internal/infrastructure/auth"
	grpcClient "backend-challenge/apps/gateway-api/internal/infrastructure/grpc/client"
	grpcMapper "backend-challenge/apps/gateway-api/internal/infrastructure/grpc/mapper"
	grpcMapperGenerated "backend-challenge/apps/gateway-api/internal/infrastructure/grpc/mapper/generated"
	"backend-challenge/apps/common/pkg/logger"
	"backend-challenge/apps/gateway-api/internal/presentation/http"
	"backend-challenge/apps/gateway-api/internal/presentation/http/handlers"
	presentationMapper "backend-challenge/apps/gateway-api/internal/presentation/http/mapper"
	"backend-challenge/apps/gateway-api/internal/presentation/http/router"
	"backend-challenge/apps/gateway-api/internal/presentation/worker"
	"backend-challenge/apps/gateway-api/internal/usecase"

	"github.com/google/wire"
	"github.com/labstack/echo/v4"
)

func ProvideLogger() logger.Logger {
	return logger.NewLogger()
}

func ProvideTokenProvider(cfg *config.Config) auth.TokenProvider {
	return auth.NewJWTProvider(cfg.Auth.JWTSecret, cfg.Auth.Duration)
}

func ProvideUserGrpcClient(cfg *config.Config) (pb.UserServiceClient, func(), error) {
	return grpcClient.NewUserGrpcClient(cfg.GRPC.ServerAddr)
}

func ProvideGrpcMapper() grpcMapper.GrpcMapper {
	return &grpcMapperGenerated.GrpcMapperImpl{}
}

func ProvidePresentationMapper() presentationMapper.PresentationMapper {
	return &presentationMapper.PresentationMapperImpl{}
}

func ProvideUserCountWorker(uc usecase.UserUseCase, log logger.Logger) *worker.UserCountWorker {
	return worker.NewUserCountWorker(uc, log, 0)
}

func ProvideEchoServer(
	cfg *config.Config,
	r *router.Router,
	log logger.Logger,
	tokenProvider auth.TokenProvider,
) *echo.Echo {
	return http.NewServer(cfg, r, log, tokenProvider)
}

var CommonSet = wire.NewSet(
	ProvideLogger,
	ProvideTokenProvider,
	ProvideUserGrpcClient,
	ProvideGrpcMapper,
	ProvidePresentationMapper,
	usecase.NewUserService,
	handlers.NewUserHandler,
	router.NewRouter,
	ProvideUserCountWorker,
	ProvideEchoServer,
)
