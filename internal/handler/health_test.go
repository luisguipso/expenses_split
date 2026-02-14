package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func TestHealthHandler_Healthy(t *testing.T) {
	e := echo.New()
	health := &mock.HealthChecker{
		PingFn: func(ctx context.Context) error {
			return nil
		},
	}
	RegisterHealthRoutes(e, health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp domain.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
	if resp.Database != "ok" {
		t.Errorf("expected database ok, got %s", resp.Database)
	}
}

func TestHealthHandler_DBUnhealthy(t *testing.T) {
	e := echo.New()
	health := &mock.HealthChecker{
		PingFn: func(ctx context.Context) error {
			return errors.New("connection refused")
		},
	}
	RegisterHealthRoutes(e, health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp domain.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
	if resp.Database != "error" {
		t.Errorf("expected database error, got %s", resp.Database)
	}
}
