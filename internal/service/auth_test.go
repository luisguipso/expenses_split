package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lguilherme/contas/internal/domain"
)

const testSecret = "test-secret-key-for-testing"

func TestTokenService_GenerateAndValidate(t *testing.T) {
	svc := NewJWTTokenService(testSecret)

	tokens, err := svc.Generate("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("access token is empty")
	}
	if tokens.RefreshToken == "" {
		t.Error("refresh token is empty")
	}
	if tokens.ExpiresAt == 0 {
		t.Error("expires_at is zero")
	}

	// Validate access token
	claims, err := svc.Validate(tokens.AccessToken)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", claims.Email)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected user_id user-123, got %s", claims.UserID)
	}

	// Validate refresh token
	refreshClaims, err := svc.Validate(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Validate (refresh) failed: %v", err)
	}
	if refreshClaims.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", refreshClaims.Email)
	}
}

func TestTokenService_Validate_InvalidToken(t *testing.T) {
	svc := NewJWTTokenService(testSecret)

	_, err := svc.Validate("invalid-token")
	if err != domain.ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestTokenService_Validate_WrongSecret(t *testing.T) {
	svc := NewJWTTokenService(testSecret)

	claims := &jwtClaims{
		UserID: "some-id",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("wrong-secret"))

	_, err := svc.Validate(tokenStr)
	if err != domain.ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestTokenService_Validate_ExpiredToken(t *testing.T) {
	svc := NewJWTTokenService(testSecret)

	claims := &jwtClaims{
		UserID: "some-id",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))

	_, err := svc.Validate(tokenStr)
	if err != domain.ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_RefreshToken_Valid(t *testing.T) {
	tokenSvc := NewJWTTokenService(testSecret)
	authSvc := NewAuthService(nil, tokenSvc)

	tokens, err := tokenSvc.Generate("user-456", "refresh@example.com")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	newTokens, err := authSvc.RefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if newTokens.AccessToken == "" {
		t.Error("new access token is empty")
	}

	newClaims, err := tokenSvc.Validate(newTokens.AccessToken)
	if err != nil {
		t.Fatalf("new access token is invalid: %v", err)
	}
	if newClaims.Email != "refresh@example.com" {
		t.Errorf("expected email refresh@example.com, got %s", newClaims.Email)
	}
}

func TestAuthService_RefreshToken_Invalid(t *testing.T) {
	tokenSvc := NewJWTTokenService(testSecret)
	authSvc := NewAuthService(nil, tokenSvc)

	_, err := authSvc.RefreshToken("invalid-refresh-token")
	if err != domain.ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
