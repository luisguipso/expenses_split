package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/service"
)

const testSecret = "test-secret-for-middleware"

func makeTestTokenService() *service.JWTTokenService {
	return service.NewJWTTokenService(testSecret)
}

func generateTestToken(userID, email string) string {
	svc := makeTestTokenService()
	tokens, _ := svc.Generate(userID, email)
	return tokens.AccessToken
}

func TestJWTAuth_ValidToken(t *testing.T) {
	e := echo.New()
	tokenSvc := makeTestTokenService()
	mw := JWTAuth(tokenSvc)

	token := generateTestToken("user-123", "test@example.com")

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
	tokenSvc := makeTestTokenService()
	mw := JWTAuth(tokenSvc)

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
	tokenSvc := makeTestTokenService()
	mw := JWTAuth(tokenSvc)

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
	tokenSvc := makeTestTokenService()
	mw := JWTAuth(tokenSvc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer expired-token-string")
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
	tokenSvc := makeTestTokenService()
	mw := JWTAuth(tokenSvc)

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
