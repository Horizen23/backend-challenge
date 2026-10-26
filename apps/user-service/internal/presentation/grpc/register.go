package grpc

import (
	pb "backend-challenge/apps/common/proto/pb"
	"google.golang.org/grpc"
)

func RegisterService(server *grpc.Server, userServer *UserGrpcServer) {
	pb.RegisterUserServiceServer(server, userServer)
}
