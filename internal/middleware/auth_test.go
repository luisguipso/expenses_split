package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/service"
)

const testSecret = "test-secret-for-middleware"

func makeTestAuthService() *service.AuthService {
	return service.NewAuthService(nil, testSecret)
}

func generateTestToken(userID, email string, expired bool) string {
	expiry := time.Now().Add(15 * time.Minute)
	if expired {
		expiry = time.Now().Add(-1 * time.Hour)
	}

	claims := &service.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))
	return tokenStr
}

func TestJWTAuth_ValidToken(t *testing.T) {
	e := echo.New()
	authSvc := makeTestAuthService()
	mw := JWTAuth(authSvc)

	token := generateTestToken("user-123", "test@example.com", false)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var capturedUserID, capturedEmail string
	handler := mw(func(c echo.Context) error {
		capturedUserID = c.Get("user_id").(string)
		capturedEmail = c.Get("user_email").(string)
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedUserID != "user-123" {
		t.Errorf("expected user_id user-123, got %s", capturedUserID)
	}
	if capturedEmail != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", capturedEmail)
	}
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	e := echo.New()
	authSvc := makeTestAuthService()
	mw := JWTAuth(authSvc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %v", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	e := echo.New()
	authSvc := makeTestAuthService()
	mw := JWTAuth(authSvc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %v", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	e := echo.New()
	authSvc := makeTestAuthService()
	mw := JWTAuth(authSvc)

	token := generateTestToken("user-123", "test@example.com", true)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %v", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	e := echo.New()
	authSvc := makeTestAuthService()
	mw := JWTAuth(authSvc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %v", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}
