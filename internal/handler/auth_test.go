package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/service"
)

type mockAuthService struct {
	registerFunc func() (*service.TokenPair, error)
	loginFunc    func() (*service.TokenPair, error)
}

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

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["user_id"] != "test-user-id" {
		t.Errorf("expected user_id test-user-id, got %v", result["user_id"])
	}
	if result["email"] != "test@example.com" {
		t.Errorf("expected email test@example.com, got %v", result["email"])
	}
}
