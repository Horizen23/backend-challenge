package auth_test

import (
	"testing"
	"time"

	"backend-challenge/apps/gateway-api/internal/infrastructure/auth"
)

func TestJWTProvider(t *testing.T) {
	provider := auth.NewJWTProvider("my-test-secret-32-bytes-long-key!", time.Hour)

	token, err := provider.GenerateToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := provider.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != "user-123" || claims.Email != "test@example.com" {
		t.Errorf("unexpected claims: %+v", claims)
	}

	_, err = provider.ValidateToken("invalid.token.payload")
	if err == nil {
		t.Error("expected error for malformed token")
	}
}

func TestPasswordHashing(t *testing.T) {
	pass := "secretPassword"
	hash, err := auth.HashPassword(pass)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !auth.CheckPassword(pass, hash) {
		t.Error("expected password to match hash")
	}

	if auth.CheckPassword("wrongPassword", hash) {
		t.Error("expected wrong password to fail")
	}
}
