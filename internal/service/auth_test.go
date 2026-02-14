package service

import (
"context"
"errors"
"testing"
"time"

"github.com/golang-jwt/jwt/v5"
"github.com/lguilherme/contas/internal/domain"
"github.com/lguilherme/contas/internal/mock"
"golang.org/x/crypto/bcrypt"
)

const testSecret = "test-secret-key-for-testing"

// --- TokenService tests ---

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

// --- AuthService.RefreshToken tests ---

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

// --- AuthService.Register tests ---

func TestAuthService_Register_Success(t *testing.T) {
repo := &mock.UserRepository{
CreateFn: func(ctx context.Context, user *domain.User) error {
user.ID = "new-user-id"
return nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc)

user, tokens, err := authSvc.Register(context.Background(), domain.RegisterInput{
Name:     "Test User",
Email:    "test@example.com",
Password: "password123",
})
if err != nil {
t.Fatalf("Register failed: %v", err)
}
if user.ID != "new-user-id" {
t.Errorf("expected user ID new-user-id, got %s", user.ID)
}
if user.Email != "test@example.com" {
t.Errorf("expected email test@example.com, got %s", user.Email)
}
if tokens.AccessToken == "" {
t.Error("access token is empty")
}
if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
t.Error("password hash doesn't match")
}
}

func TestAuthService_Register_EmailExists(t *testing.T) {
repo := &mock.UserRepository{
CreateFn: func(ctx context.Context, user *domain.User) error {
return domain.ErrEmailExists
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc)

_, _, err := authSvc.Register(context.Background(), domain.RegisterInput{
Name:     "Test",
Email:    "taken@example.com",
Password: "password123",
})
if !errors.Is(err, domain.ErrEmailExists) {
t.Errorf("expected ErrEmailExists, got %v", err)
}
}

func TestAuthService_Register_RepoError(t *testing.T) {
repoErr := errors.New("database connection lost")
repo := &mock.UserRepository{
CreateFn: func(ctx context.Context, user *domain.User) error {
return repoErr
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc)

_, _, err := authSvc.Register(context.Background(), domain.RegisterInput{
Name:     "Test",
Email:    "test@example.com",
Password: "password123",
})
if err == nil {
t.Fatal("expected error, got nil")
}
if errors.Is(err, domain.ErrEmailExists) {
t.Error("should not be ErrEmailExists")
}
}

// --- AuthService.Login tests ---

func TestAuthService_Login_Success(t *testing.T) {
hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
repo := &mock.UserRepository{
FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
return &domain.User{
ID:           "user-789",
Name:         "Test User",
Email:        email,
PasswordHash: string(hash),
}, nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc)

user, tokens, err := authSvc.Login(context.Background(), domain.LoginInput{
Email:    "test@example.com",
Password: "correctpassword",
})
if err != nil {
t.Fatalf("Login failed: %v", err)
}
if user.ID != "user-789" {
t.Errorf("expected user ID user-789, got %s", user.ID)
}
if tokens.AccessToken == "" {
t.Error("access token is empty")
}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
repo := &mock.UserRepository{
FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
return nil, domain.ErrUserNotFound
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc)

_, _, err := authSvc.Login(context.Background(), domain.LoginInput{
Email:    "noone@example.com",
Password: "password123",
})
if !errors.Is(err, domain.ErrInvalidCredentials) {
t.Errorf("expected ErrInvalidCredentials, got %v", err)
}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
repo := &mock.UserRepository{
FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
return &domain.User{
ID:           "user-789",
Email:        email,
PasswordHash: string(hash),
}, nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc)

_, _, err := authSvc.Login(context.Background(), domain.LoginInput{
Email:    "test@example.com",
Password: "wrongpassword",
})
if !errors.Is(err, domain.ErrInvalidCredentials) {
t.Errorf("expected ErrInvalidCredentials, got %v", err)
}
}

func TestAuthService_Login_RepoError(t *testing.T) {
repo := &mock.UserRepository{
FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
return nil, errors.New("database timeout")
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc)

_, _, err := authSvc.Login(context.Background(), domain.LoginInput{
Email:    "test@example.com",
Password: "password123",
})
if err == nil {
t.Fatal("expected error, got nil")
}
if errors.Is(err, domain.ErrInvalidCredentials) {
t.Error("should not be ErrInvalidCredentials for repo error")
}
}
