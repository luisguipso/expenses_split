package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lguilherme/contas/internal/model"
)

const testSecret = "test-secret-key-for-testing"

func TestGenerateAndValidateToken(t *testing.T) {
	s := &AuthService{jwtSecret: []byte(testSecret)}

	user := &model.User{
		Email: "test@example.com",
	}
	// Set a valid UUID
	user.ID.Bytes = [16]byte{0xa0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}
	user.ID.Valid = true

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		t.Fatalf("generateTokenPair failed: %v", err)
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
	claims, err := s.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", claims.Email)
	}
	if claims.UserID == "" {
		t.Error("user_id is empty")
	}

	// Validate refresh token
	refreshClaims, err := s.ValidateToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("ValidateToken (refresh) failed: %v", err)
	}
	if refreshClaims.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", refreshClaims.Email)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	s := &AuthService{jwtSecret: []byte(testSecret)}

	_, err := s.ValidateToken("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	s := &AuthService{jwtSecret: []byte(testSecret)}

	// Generate token with different secret
	claims := &Claims{
		UserID: "some-id",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("wrong-secret"))

	_, err := s.ValidateToken(tokenStr)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	s := &AuthService{jwtSecret: []byte(testSecret)}

	claims := &Claims{
		UserID: "some-id",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(s.jwtSecret)

	_, err := s.ValidateToken(tokenStr)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestRefreshToken_Valid(t *testing.T) {
	s := &AuthService{jwtSecret: []byte(testSecret)}

	user := &model.User{
		Email: "refresh@example.com",
	}
	user.ID = pgtype.UUID{
		Bytes: [16]byte{0xa0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02},
		Valid: true,
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		t.Fatalf("generateTokenPair failed: %v", err)
	}

	newTokens, err := s.RefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if newTokens.AccessToken == "" {
		t.Error("new access token is empty")
	}
	// Validate the new access token works
	newClaims, err := s.ValidateToken(newTokens.AccessToken)
	if err != nil {
		t.Fatalf("new access token is invalid: %v", err)
	}
	if newClaims.Email != "refresh@example.com" {
		t.Errorf("expected email refresh@example.com, got %s", newClaims.Email)
	}
}

func TestRefreshToken_Invalid(t *testing.T) {
	s := &AuthService{jwtSecret: []byte(testSecret)}

	_, err := s.RefreshToken("invalid-refresh-token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
