package handler

import (
"context"
"encoding/json"
"errors"
"net/http"
"net/http/httptest"
"strings"
"testing"

"github.com/labstack/echo/v4"
"github.com/lguilherme/contas/internal/domain"
"github.com/lguilherme/contas/internal/mock"
)

func TestAuthHandler_Register_Validation(t *testing.T) {
e := echo.New()
h := NewAuthHandler(nil)

tests := []struct {
name       string
body       string
wantStatus int
}{
{"empty body", `{}`, http.StatusBadRequest},
{"missing email", `{"name":"Test","password":"123456"}`, http.StatusBadRequest},
{"missing name", `{"email":"test@example.com","password":"123456"}`, http.StatusBadRequest},
{"missing password", `{"name":"Test","email":"test@example.com"}`, http.StatusBadRequest},
{"short password", `{"name":"Test","email":"test@example.com","password":"12345"}`, http.StatusBadRequest},
{"invalid json", `not json`, http.StatusBadRequest},
{"invalid email format", `{"name":"Test","email":"notanemail","password":"123456"}`, http.StatusBadRequest},
{"name too long", `{"name":"` + strings.Repeat("a", 256) + `","email":"test@example.com","password":"123456"}`, http.StatusBadRequest},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.Register(c)
if err == nil {
t.Fatal("expected error")
}
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != tt.wantStatus {
t.Errorf("expected status %d, got %d", tt.wantStatus, he.Code)
}
})
}
}

func TestAuthHandler_Register_Success(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
RegisterFn: func(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
return &domain.User{ID: "new-id", Name: input.Name, Email: input.Email}, nil
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/register",
strings.NewReader(`{"name":"Test","email":"test@example.com","password":"password123"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

if err := h.Register(c); err != nil {
t.Fatalf("expected no error, got %v", err)
}
if rec.Code != http.StatusCreated {
t.Errorf("expected 201, got %d", rec.Code)
}

var resp domain.RegisterResponse
if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if resp.User.Email != "test@example.com" {
t.Errorf("expected email test@example.com, got %s", resp.User.Email)
}
if resp.Message == "" {
t.Error("expected non-empty message")
}
}

func TestAuthHandler_Register_EmailConflict(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
RegisterFn: func(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
return nil, domain.ErrEmailExists
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/register",
strings.NewReader(`{"name":"Test","email":"taken@example.com","password":"password123"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.Register(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusConflict {
t.Errorf("expected 409, got %d", he.Code)
}
}

func TestAuthHandler_Register_InternalError(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
RegisterFn: func(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
return nil, errors.New("unexpected")
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/register",
strings.NewReader(`{"name":"Test","email":"test@example.com","password":"password123"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.Register(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusInternalServerError {
t.Errorf("expected 500, got %d", he.Code)
}
}

func TestAuthHandler_Login_Validation(t *testing.T) {
e := echo.New()
h := NewAuthHandler(nil)

tests := []struct {
name       string
body       string
wantStatus int
}{
{"empty body", `{}`, http.StatusBadRequest},
{"missing email", `{"password":"123456"}`, http.StatusBadRequest},
{"missing password", `{"email":"test@example.com"}`, http.StatusBadRequest},
{"invalid json", `not json`, http.StatusBadRequest},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.Login(c)
if err == nil {
t.Fatal("expected error")
}
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != tt.wantStatus {
t.Errorf("expected status %d, got %d", tt.wantStatus, he.Code)
}
})
}
}

func TestAuthHandler_Login_Success(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
LoginFn: func(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error) {
return &domain.User{ID: "user-1", Name: "Test", Email: input.Email},
&domain.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresAt: 123},
nil
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/login",
strings.NewReader(`{"email":"test@example.com","password":"password123"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

if err := h.Login(c); err != nil {
t.Fatalf("expected no error, got %v", err)
}
if rec.Code != http.StatusOK {
t.Errorf("expected 200, got %d", rec.Code)
}

var resp domain.AuthResponse
if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if resp.User.Email != "test@example.com" {
t.Errorf("expected email test@example.com, got %s", resp.User.Email)
}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
LoginFn: func(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error) {
return nil, nil, domain.ErrInvalidCredentials
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/login",
strings.NewReader(`{"email":"test@example.com","password":"wrong"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.Login(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusUnauthorized {
t.Errorf("expected 401, got %d", he.Code)
}
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
LoginFn: func(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error) {
return nil, nil, errors.New("db error")
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/login",
strings.NewReader(`{"email":"test@example.com","password":"password123"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.Login(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusInternalServerError {
t.Errorf("expected 500, got %d", he.Code)
}
}

func TestAuthHandler_Refresh_Validation(t *testing.T) {
e := echo.New()
h := NewAuthHandler(nil)

tests := []struct {
name       string
body       string
wantStatus int
}{
{"empty body", `{}`, http.StatusBadRequest},
{"missing refresh_token", `{"refresh_token":""}`, http.StatusBadRequest},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(tt.body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.Refresh(c)
if err == nil {
t.Fatal("expected error")
}
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != tt.wantStatus {
t.Errorf("expected status %d, got %d", tt.wantStatus, he.Code)
}
})
}
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
RefreshTokenFn: func(refreshToken string) (*domain.TokenPair, error) {
return &domain.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh", ExpiresAt: 456}, nil
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/refresh",
strings.NewReader(`{"refresh_token":"old-refresh"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

if err := h.Refresh(c); err != nil {
t.Fatalf("expected no error, got %v", err)
}
if rec.Code != http.StatusOK {
t.Errorf("expected 200, got %d", rec.Code)
}

var resp domain.TokenResponse
if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if resp.Tokens.AccessToken != "new-access" {
t.Errorf("expected access token 'new-access', got %s", resp.Tokens.AccessToken)
}
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
RefreshTokenFn: func(refreshToken string) (*domain.TokenPair, error) {
return nil, domain.ErrInvalidToken
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/refresh",
strings.NewReader(`{"refresh_token":"bad-token"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.Refresh(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusUnauthorized {
t.Errorf("expected 401, got %d", he.Code)
}
}

func TestAuthHandler_Me(t *testing.T) {
e := echo.New()
h := NewAuthHandler(nil)

req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)
c.Set("user_id", "test-user-id")
c.Set("user_email", "test@example.com")

if err := h.Me(c); err != nil {
t.Fatalf("Me failed: %v", err)
}

if rec.Code != http.StatusOK {
t.Errorf("expected 200, got %d", rec.Code)
}

var result domain.MeResponse
if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if result.UserID != "test-user-id" {
t.Errorf("expected user_id test-user-id, got %v", result.UserID)
}
if result.Email != "test@example.com" {
t.Errorf("expected email test@example.com, got %v", result.Email)
}
}

// --- Email Verification handler tests ---

func TestAuthHandler_Login_EmailNotVerified(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
LoginFn: func(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error) {
return nil, nil, domain.ErrEmailNotVerified
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/login",
strings.NewReader(`{"email":"test@example.com","password":"password123"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.Login(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusForbidden {
t.Errorf("expected 403, got %d", he.Code)
}
}

func TestAuthHandler_VerifyEmail_Success(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
VerifyEmailFn: func(ctx context.Context, input domain.VerifyEmailInput) (*domain.User, *domain.TokenPair, error) {
return &domain.User{ID: "user-1", Name: "Test", Email: input.Email, EmailVerified: true},
&domain.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresAt: 123},
nil
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/verify-email",
strings.NewReader(`{"email":"test@example.com","code":"123456"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

if err := h.VerifyEmail(c); err != nil {
t.Fatalf("expected no error, got %v", err)
}
if rec.Code != http.StatusOK {
t.Errorf("expected 200, got %d", rec.Code)
}

var resp domain.AuthResponse
if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if resp.User.Email != "test@example.com" {
t.Errorf("expected email test@example.com, got %s", resp.User.Email)
}
if resp.Tokens.AccessToken != "access" {
t.Errorf("expected access token 'access', got %s", resp.Tokens.AccessToken)
}
}

func TestAuthHandler_VerifyEmail_InvalidCode(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
VerifyEmailFn: func(ctx context.Context, input domain.VerifyEmailInput) (*domain.User, *domain.TokenPair, error) {
return nil, nil, domain.ErrInvalidVerificationCode
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/verify-email",
strings.NewReader(`{"email":"test@example.com","code":"000000"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.VerifyEmail(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusBadRequest {
t.Errorf("expected 400, got %d", he.Code)
}
}

func TestAuthHandler_VerifyEmail_ExpiredCode(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
VerifyEmailFn: func(ctx context.Context, input domain.VerifyEmailInput) (*domain.User, *domain.TokenPair, error) {
return nil, nil, domain.ErrVerificationExpired
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/verify-email",
strings.NewReader(`{"email":"test@example.com","code":"123456"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.VerifyEmail(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusBadRequest {
t.Errorf("expected 400, got %d", he.Code)
}
}

func TestAuthHandler_VerifyEmail_MissingFields(t *testing.T) {
e := echo.New()
h := NewAuthHandler(nil)

tests := []struct {
name string
body string
}{
{"missing email", `{"code":"123456"}`},
{"missing code", `{"email":"test@example.com"}`},
{"both empty", `{}`},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
req := httptest.NewRequest(http.MethodPost, "/auth/verify-email", strings.NewReader(tt.body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.VerifyEmail(c)
if err == nil {
t.Fatal("expected error")
}
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusBadRequest {
t.Errorf("expected 400, got %d", he.Code)
}
})
}
}

func TestAuthHandler_ResendCode_Success(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
ResendCodeFn: func(ctx context.Context, input domain.ResendCodeInput) error {
return nil
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/resend-code",
strings.NewReader(`{"email":"test@example.com"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

if err := h.ResendCode(c); err != nil {
t.Fatalf("expected no error, got %v", err)
}
if rec.Code != http.StatusOK {
t.Errorf("expected 200, got %d", rec.Code)
}
}

func TestAuthHandler_ResendCode_MissingEmail(t *testing.T) {
e := echo.New()
h := NewAuthHandler(nil)

req := httptest.NewRequest(http.MethodPost, "/auth/resend-code",
strings.NewReader(`{}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.ResendCode(c)
if err == nil {
t.Fatal("expected error")
}
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusBadRequest {
t.Errorf("expected 400, got %d", he.Code)
}
}

func TestAuthHandler_ResendCode_UserNotFound(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
ResendCodeFn: func(ctx context.Context, input domain.ResendCodeInput) error {
return domain.ErrUserNotFound
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/resend-code",
strings.NewReader(`{"email":"unknown@example.com"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.ResendCode(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusNotFound {
t.Errorf("expected 404, got %d", he.Code)
}
}

// --- ForgotPassword handler tests ---

func TestAuthHandler_ForgotPassword_Success(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
ForgotPasswordFn: func(ctx context.Context, input domain.ForgotPasswordInput) error {
return nil
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password",
strings.NewReader(`{"email":"user@example.com"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

if err := h.ForgotPassword(c); err != nil {
t.Fatalf("expected no error, got %v", err)
}
if rec.Code != http.StatusOK {
t.Errorf("expected 200, got %d", rec.Code)
}
}

func TestAuthHandler_ForgotPassword_Validation(t *testing.T) {
e := echo.New()
h := NewAuthHandler(nil)

tests := []struct {
name       string
body       string
wantStatus int
}{
{"empty body", `{}`, http.StatusBadRequest},
{"missing email", `{"email":""}`, http.StatusBadRequest},
{"invalid email format", `{"email":"notanemail"}`, http.StatusBadRequest},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(tt.body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.ForgotPassword(c)
if err == nil {
t.Fatal("expected error")
}
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != tt.wantStatus {
t.Errorf("expected status %d, got %d", tt.wantStatus, he.Code)
}
})
}
}

func TestAuthHandler_ForgotPassword_InternalError(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
ForgotPasswordFn: func(ctx context.Context, input domain.ForgotPasswordInput) error {
return errors.New("smtp failure")
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password",
strings.NewReader(`{"email":"user@example.com"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.ForgotPassword(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusInternalServerError {
t.Errorf("expected 500, got %d", he.Code)
}
}

// --- ResetPassword handler tests ---

func TestAuthHandler_ResetPassword_Success(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
ResetPasswordFn: func(ctx context.Context, input domain.ResetPasswordInput) error {
return nil
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/reset-password",
strings.NewReader(`{"token":"abc123","new_password":"newpass123"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

if err := h.ResetPassword(c); err != nil {
t.Fatalf("expected no error, got %v", err)
}
if rec.Code != http.StatusOK {
t.Errorf("expected 200, got %d", rec.Code)
}
}

func TestAuthHandler_ResetPassword_Validation(t *testing.T) {
e := echo.New()
h := NewAuthHandler(nil)

tests := []struct {
name       string
body       string
wantStatus int
}{
{"empty body", `{}`, http.StatusBadRequest},
{"missing token", `{"new_password":"newpass123"}`, http.StatusBadRequest},
{"short password", `{"token":"abc123","new_password":"12345"}`, http.StatusBadRequest},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(tt.body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.ResetPassword(c)
if err == nil {
t.Fatal("expected error")
}
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != tt.wantStatus {
t.Errorf("expected status %d, got %d", tt.wantStatus, he.Code)
}
})
}
}

func TestAuthHandler_ResetPassword_InvalidToken(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
ResetPasswordFn: func(ctx context.Context, input domain.ResetPasswordInput) error {
return domain.ErrResetTokenInvalid
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/reset-password",
strings.NewReader(`{"token":"invalid","new_password":"newpass123"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.ResetPassword(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusBadRequest {
t.Errorf("expected 400, got %d", he.Code)
}
}

func TestAuthHandler_ResetPassword_ExpiredToken(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
ResetPasswordFn: func(ctx context.Context, input domain.ResetPasswordInput) error {
return domain.ErrResetTokenExpired
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/reset-password",
strings.NewReader(`{"token":"expired","new_password":"newpass123"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.ResetPassword(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusBadRequest {
t.Errorf("expected 400, got %d", he.Code)
}
}

func TestAuthHandler_ResetPassword_SamePassword(t *testing.T) {
e := echo.New()
authSvc := &mock.AuthService{
ResetPasswordFn: func(ctx context.Context, input domain.ResetPasswordInput) error {
return domain.ErrPasswordSameAsOld
},
}
h := NewAuthHandler(authSvc)

req := httptest.NewRequest(http.MethodPost, "/auth/reset-password",
strings.NewReader(`{"token":"valid","new_password":"samepass"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
c := e.NewContext(req, rec)

err := h.ResetPassword(c)
he, ok := err.(*echo.HTTPError)
if !ok {
t.Fatalf("expected HTTPError, got %T: %v", err, err)
}
if he.Code != http.StatusBadRequest {
t.Errorf("expected 400, got %d", he.Code)
}
}
