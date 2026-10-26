package ctxutils

import "context"

type contextKey string

const (
	RequestIDKey contextKey = "X-Request-ID"
	UserIDKey    contextKey = "UserID"
	EmailKey     contextKey = "Email"
)

func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(RequestIDKey).(string); ok {
		return val
	}
	return ""
}

func SetAuthUser(ctx context.Context, userID, email string) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	return context.WithValue(ctx, EmailKey, email)
}

func GetAuthUser(ctx context.Context) (string, string) {
	userID, _ := ctx.Value(UserIDKey).(string)
	email, _ := ctx.Value(EmailKey).(string)
	return userID, email
}
