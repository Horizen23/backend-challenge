package client

import (
	"context"

	pb "backend-challenge/apps/common/proto/pb"
	"backend-challenge/apps/common/pkg/ctxutils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ==============================================================================
// Architectural & Security Design: API Gateway to Internal gRPC Propagation
// ==============================================================================
// `gateway-api` sits at the edge and terminates public HTTP/HTTPS traffic.
// After successfully validating external JWT tokens via `JWTAuthMiddleware`,
// this interceptor propagates the verified user identity (user_id, email, request_id)
// as gRPC metadata to internal microservices without requiring internal services
// to duplicate full JWT signature parsing or token expiration validation.
// ==============================================================================

// MetadataClientInterceptor propagates RequestID, UserID, and Email from context to outgoing gRPC metadata
func MetadataClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		md := metadata.MD{}

		if reqID := ctxutils.GetRequestID(ctx); reqID != "" {
			md.Set(string(ctxutils.RequestIDKey), reqID)
		}

		if userID, email := ctxutils.GetAuthUser(ctx); userID != "" {
			md.Set(string(ctxutils.UserIDKey), userID)
			md.Set(string(ctxutils.EmailKey), email)
		}

		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func NewUserGrpcClient(serverAddr string) (pb.UserServiceClient, func(), error) {
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(MetadataClientInterceptor()),
	)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = conn.Close()
	}

	return pb.NewUserServiceClient(conn), cleanup, nil
}
