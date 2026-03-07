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
authSvc := NewAuthService(nil, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

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
authSvc := NewAuthService(nil, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

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
verifyRepo := &mock.EmailVerificationRepository{
CreateFn: func(ctx context.Context, v *domain.EmailVerification) error {
return nil
},
}
emailSvc := &mock.EmailServiceMock{
SendVerificationCodeFn: func(to, code string) error {
return nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc, verifyRepo, emailSvc, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

user, err := authSvc.Register(context.Background(), domain.RegisterInput{
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
authSvc := NewAuthService(repo, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

_, err := authSvc.Register(context.Background(), domain.RegisterInput{
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
authSvc := NewAuthService(repo, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

_, err := authSvc.Register(context.Background(), domain.RegisterInput{
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
ID:            "user-789",
Name:          "Test User",
Email:         email,
PasswordHash:  string(hash),
EmailVerified: true,
}, nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

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
authSvc := NewAuthService(repo, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

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
authSvc := NewAuthService(repo, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

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
authSvc := NewAuthService(repo, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

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

// --- Email Verification tests ---

func TestAuthService_Register_SendsVerificationCode(t *testing.T) {
var emailSent bool
var sentTo, sentCode string
repo := &mock.UserRepository{
CreateFn: func(ctx context.Context, user *domain.User) error {
user.ID = "new-user-id"
return nil
},
}
verifyRepo := &mock.EmailVerificationRepository{
CreateFn: func(ctx context.Context, v *domain.EmailVerification) error {
return nil
},
}
emailSvc := &mock.EmailServiceMock{
SendVerificationCodeFn: func(to, code string) error {
emailSent = true
sentTo = to
sentCode = code
return nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc, verifyRepo, emailSvc, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

user, err := authSvc.Register(context.Background(), domain.RegisterInput{
Name:     "Test User",
Email:    "verify@example.com",
Password: "password123",
})
if err != nil {
t.Fatalf("Register failed: %v", err)
}
if user == nil {
t.Fatal("expected user, got nil")
}
if !emailSent {
t.Error("expected SendVerificationCode to be called")
}
if sentTo != "verify@example.com" {
t.Errorf("expected email sent to verify@example.com, got %s", sentTo)
}
if sentCode == "" {
t.Error("expected non-empty verification code")
}
}

func TestAuthService_Login_RejectsUnverified(t *testing.T) {
hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
repo := &mock.UserRepository{
FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
return &domain.User{
ID:            "user-unverified",
Email:         email,
PasswordHash:  string(hash),
EmailVerified: false,
}, nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

_, _, err := authSvc.Login(context.Background(), domain.LoginInput{
Email:    "unverified@example.com",
Password: "password123",
})
if !errors.Is(err, domain.ErrEmailNotVerified) {
t.Errorf("expected ErrEmailNotVerified, got %v", err)
}
}

func TestAuthService_VerifyEmail_Success(t *testing.T) {
var markUsedCalled bool
var verifyEmailCalled bool
verifyRepo := &mock.EmailVerificationRepository{
FindLatestByEmailFn: func(ctx context.Context, email string) (*domain.EmailVerification, error) {
return &domain.EmailVerification{
ID:        "v-1",
UserID:    "user-1",
Email:     email,
Code:      "123456",
ExpiresAt: time.Now().Add(10 * time.Minute),
Used:      false,
}, nil
},
MarkUsedFn: func(ctx context.Context, id string) error {
markUsedCalled = true
return nil
},
}
repo := &mock.UserRepository{
VerifyEmailFn: func(ctx context.Context, userID string) error {
verifyEmailCalled = true
return nil
},
FindByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
return &domain.User{
ID:            id,
Email:         "verify@example.com",
EmailVerified: true,
}, nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc, verifyRepo, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

user, tokens, err := authSvc.VerifyEmail(context.Background(), domain.VerifyEmailInput{
Email: "verify@example.com",
Code:  "123456",
})
if err != nil {
t.Fatalf("VerifyEmail failed: %v", err)
}
if user == nil {
t.Fatal("expected user, got nil")
}
if tokens == nil || tokens.AccessToken == "" {
t.Error("expected tokens with access token")
}
if !markUsedCalled {
t.Error("expected MarkUsed to be called")
}
if !verifyEmailCalled {
t.Error("expected VerifyEmail to be called on user repo")
}
}

func TestAuthService_VerifyEmail_InvalidCode(t *testing.T) {
verifyRepo := &mock.EmailVerificationRepository{
FindLatestByEmailFn: func(ctx context.Context, email string) (*domain.EmailVerification, error) {
return &domain.EmailVerification{
ID:        "v-1",
UserID:    "user-1",
Email:     email,
Code:      "999999",
ExpiresAt: time.Now().Add(10 * time.Minute),
Used:      false,
}, nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(nil, tokenSvc, verifyRepo, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

_, _, err := authSvc.VerifyEmail(context.Background(), domain.VerifyEmailInput{
Email: "verify@example.com",
Code:  "000000",
})
if !errors.Is(err, domain.ErrInvalidVerificationCode) {
t.Errorf("expected ErrInvalidVerificationCode, got %v", err)
}
}

func TestAuthService_VerifyEmail_ExpiredCode(t *testing.T) {
verifyRepo := &mock.EmailVerificationRepository{
FindLatestByEmailFn: func(ctx context.Context, email string) (*domain.EmailVerification, error) {
return &domain.EmailVerification{
ID:        "v-1",
UserID:    "user-1",
Email:     email,
Code:      "123456",
ExpiresAt: time.Now().Add(-1 * time.Hour),
Used:      false,
}, nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(nil, tokenSvc, verifyRepo, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

_, _, err := authSvc.VerifyEmail(context.Background(), domain.VerifyEmailInput{
Email: "verify@example.com",
Code:  "123456",
})
if !errors.Is(err, domain.ErrVerificationExpired) {
t.Errorf("expected ErrVerificationExpired, got %v", err)
}
}

func TestAuthService_ResendCode_Success(t *testing.T) {
var codeCreated bool
var emailSent bool
repo := &mock.UserRepository{
FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
return &domain.User{
ID:            "user-1",
Email:         email,
EmailVerified: false,
}, nil
},
}
verifyRepo := &mock.EmailVerificationRepository{
CreateFn: func(ctx context.Context, v *domain.EmailVerification) error {
codeCreated = true
return nil
},
}
emailSvc := &mock.EmailServiceMock{
SendVerificationCodeFn: func(to, code string) error {
emailSent = true
return nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc, verifyRepo, emailSvc, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

err := authSvc.ResendCode(context.Background(), domain.ResendCodeInput{
Email: "resend@example.com",
})
if err != nil {
t.Fatalf("ResendCode failed: %v", err)
}
if !codeCreated {
t.Error("expected verification code to be created")
}
if !emailSent {
t.Error("expected verification email to be sent")
}
}

func TestAuthService_ResendCode_UserNotFound(t *testing.T) {
repo := &mock.UserRepository{
FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
return nil, domain.ErrUserNotFound
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

err := authSvc.ResendCode(context.Background(), domain.ResendCodeInput{
Email: "unknown@example.com",
})
if !errors.Is(err, domain.ErrUserNotFound) {
t.Errorf("expected ErrUserNotFound, got %v", err)
}
}

func TestAuthService_ResendCode_AlreadyVerified(t *testing.T) {
repo := &mock.UserRepository{
FindByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
return &domain.User{
ID:            "user-1",
Email:         email,
EmailVerified: true,
}, nil
},
}
tokenSvc := NewJWTTokenService(testSecret)
authSvc := NewAuthService(repo, tokenSvc, nil, nil, 15*time.Minute, nil, 30*time.Minute, "http://localhost:5173")

err := authSvc.ResendCode(context.Background(), domain.ResendCodeInput{
Email: "verified@example.com",
})
if !errors.Is(err, domain.ErrAlreadyVerified) {
t.Errorf("expected ErrAlreadyVerified, got %v", err)
}
}
