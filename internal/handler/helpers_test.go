package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestGetUserID_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-123")

	id, err := getUserID(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "user-123" {
		t.Fatalf("expected user-123, got %s", id)
	}
}

func TestGetUserID_Missing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := getUserID(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", he.Code)
	}
}

func TestGetUserID_WrongType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", 12345) // wrong type

	_, err := getUserID(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", he.Code)
	}
}

func TestGetUserID_EmptyString(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "")

	_, err := getUserID(c)
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestValidateMaxLen_OK(t *testing.T) {
	err := validateMaxLen("name", "short", 255)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateMaxLen_TooLong(t *testing.T) {
	long := strings.Repeat("a", 256)
	err := validateMaxLen("name", long, 255)
	if err == nil {
		t.Fatal("expected error for long string")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", he.Code)
	}
}

func TestValidateMaxLen_ExactLimit(t *testing.T) {
	exact := strings.Repeat("a", 255)
	err := validateMaxLen("name", exact, 255)
	if err != nil {
		t.Fatalf("expected no error for exact limit, got %v", err)
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"a@b.c", true},
		{"user@domain.co.uk", true},
		{"", false},
		{"noatsign", false},
		{"@domain.com", false},
		{"user@", false},
		{"user@domain", false},
		{"user@.com", false},
	}

	for _, tt := range tests {
		got := isValidEmail(tt.email)
		if got != tt.valid {
			t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, got, tt.valid)
		}
	}
}
