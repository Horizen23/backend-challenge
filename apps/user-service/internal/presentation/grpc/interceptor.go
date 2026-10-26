package grpc

import (
	"context"
	"strings"
	"time"

	"backend-challenge/apps/common/pkg/ctxutils"
	"backend-challenge/apps/common/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// protectedMethods lists the gRPC methods that strictly require an authenticated caller identity (user_id)
var protectedMethods = map[string]struct{}{
	"/pb.UserService/UpdateUser": {},
	"/pb.UserService/DeleteUser": {},
}

func isProtectedMethod(fullMethod string) bool {
	_, ok := protectedMethods[fullMethod]
	return ok
}

// ==============================================================================
// Architectural & Security Design: Internal gRPC Service Trust Boundary
// ==============================================================================
// In this microservice architecture, `user-service` operates strictly as an INTERNAL
// service within a private network/VPC and is NOT exposed directly to public clients.
//
// 1. Separation of Concerns & Public Boundary:
//    - Public clients interact only with `gateway-api` (API Gateway / BFF).
//    - `gateway-api` validates JWT Bearer tokens and authenticates external callers.
//
// 2. Performance & Simplicity (Avoiding Redundant JWT Verification):
//    - Re-verifying cryptographic JWT signatures inside internal gRPC calls
//      adds unnecessary CPU overhead and cryptographic key distribution complexity.
//    - Instead, `gateway-api` securely propagates authenticated caller metadata
//      (user_id, email, x-request-id) over gRPC metadata headers.
//
// 3. Auditing & Context Tracing:
//    - MetadataServerInterceptor extracts these identity attributes, injects them
//      into the Go standard context.Context, and logs structured audit records
//      correlated with the originating Request ID and User ID.
// ==============================================================================

// MetadataServerInterceptor extracts RequestID, UserID, and Email from incoming metadata and injects into context
func MetadataServerInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		var reqID, userID, email string

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get(string(ctxutils.RequestIDKey)); len(vals) > 0 {
				reqID = vals[0]
			} else if vals := md.Get(strings.ToLower(string(ctxutils.RequestIDKey))); len(vals) > 0 {
				reqID = vals[0]
			}

			if vals := md.Get(string(ctxutils.UserIDKey)); len(vals) > 0 {
				userID = vals[0]
			} else if vals := md.Get(strings.ToLower(string(ctxutils.UserIDKey))); len(vals) > 0 {
				userID = vals[0]
			}

			if vals := md.Get(string(ctxutils.EmailKey)); len(vals) > 0 {
				email = vals[0]
			} else if vals := md.Get(strings.ToLower(string(ctxutils.EmailKey))); len(vals) > 0 {
				email = vals[0]
			}
		}

		if reqID != "" {
			ctx = ctxutils.SetRequestID(ctx, reqID)
		}
		if userID != "" || email != "" {
			ctx = ctxutils.SetAuthUser(ctx, userID, email)
		}

		// Security / Enforcement: Ensure caller identity (user_id) is present for mutative methods
		if isProtectedMethod(info.FullMethod) && strings.TrimSpace(userID) == "" {
			log.Warn("Unauthorized gRPC invocation: missing authenticated user_id in metadata",
				logger.String("method", info.FullMethod),
				logger.String("request_id", reqID),
			)
			return nil, status.Error(codes.Unauthenticated, "unauthenticated: user_id is required in metadata")
		}

		resp, err := handler(ctx, req)
		latency := time.Since(start)

		if err != nil {
			log.Error("gRPC request failed",
				logger.String("method", info.FullMethod),
				logger.String("request_id", reqID),
				logger.String("user_id", userID),
				logger.Duration("latency", latency),
				logger.Err(err),
			)
		} else {
			log.Info("gRPC request completed",
				logger.String("method", info.FullMethod),
				logger.String("request_id", reqID),
				logger.String("user_id", userID),
				logger.Duration("latency", latency),
			)
		}

		return resp, err
	}
}
