package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func summaryContext(e *echo.Echo, method, path string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-1")
	c.SetParamNames("householdId")
	c.SetParamValues("hh-1")
	return c, rec
}

func TestSummaryHandler_GetSummary_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.SummaryService{
		GenerateFn: func(ctx context.Context, householdID string, year, month int, userID string) (*domain.SummaryResponse, error) {
			return &domain.SummaryResponse{
				ID: "sum-1", HouseholdID: householdID,
				Year: year, Month: month,
				TotalSharedCents: 10000,
				TotalAllCents:    12000,
				Items: []domain.SummaryItemResponse{
					{UserID: "u1", UserName: "Alice", AmountDueCents: 7000},
					{UserID: "u2", UserName: "Bob", AmountDueCents: 5000},
				},
			}, nil
		},
	}
	h := NewSummaryHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/households/hh-1/summary?year=2024&month=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-1")
	c.SetParamNames("householdId")
	c.SetParamValues("hh-1")

	if err := h.GetSummary(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp domain.SummaryResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.TotalSharedCents != 10000 {
		t.Errorf("expected 10000, got %d", resp.TotalSharedCents)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
}

func TestSummaryHandler_GetSummary_MissingParams(t *testing.T) {
	e := echo.New()
	h := NewSummaryHandler(nil)

	tests := []struct {
		name string
		url  string
	}{
		{"missing both", "/households/hh-1/summary"},
		{"missing month", "/households/hh-1/summary?year=2024"},
		{"missing year", "/households/hh-1/summary?month=1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "user-1")
			c.SetParamNames("householdId")
			c.SetParamValues("hh-1")

			err := h.GetSummary(c)
			he := err.(*echo.HTTPError)
			if he.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", he.Code)
			}
		})
	}
}

func TestSummaryHandler_GetSummary_InvalidParams(t *testing.T) {
	e := echo.New()
	h := NewSummaryHandler(nil)

	tests := []struct {
		name string
		url  string
	}{
		{"bad year", "/households/hh-1/summary?year=abc&month=1"},
		{"bad month", "/households/hh-1/summary?year=2024&month=13"},
		{"month zero", "/households/hh-1/summary?year=2024&month=0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "user-1")
			c.SetParamNames("householdId")
			c.SetParamValues("hh-1")

			err := h.GetSummary(c)
			he := err.(*echo.HTTPError)
			if he.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", he.Code)
			}
		})
	}
}

func TestSummaryHandler_GetSummary_Forbidden(t *testing.T) {
	e := echo.New()
	svc := &mock.SummaryService{
		GenerateFn: func(ctx context.Context, householdID string, year, month int, userID string) (*domain.SummaryResponse, error) {
			return nil, domain.ErrForbidden
		},
	}
	h := NewSummaryHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/households/hh-1/summary?year=2024&month=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-1")
	c.SetParamNames("householdId")
	c.SetParamValues("hh-1")

	err := h.GetSummary(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", he.Code)
	}
}

func TestSummaryHandler_GetSummary_NoSalary(t *testing.T) {
	e := echo.New()
	svc := &mock.SummaryService{
		GenerateFn: func(ctx context.Context, householdID string, year, month int, userID string) (*domain.SummaryResponse, error) {
			return nil, domain.ErrNoMembersWithSalary
		},
	}
	h := NewSummaryHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/households/hh-1/summary?year=2024&month=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-1")
	c.SetParamNames("householdId")
	c.SetParamValues("hh-1")

	err := h.GetSummary(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", he.Code)
	}
}

func TestSummaryHandler_GetDashboard_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.SummaryService{
		GetDashboardFn: func(ctx context.Context, householdID, userID string) (*domain.DashboardResponse, error) {
			return &domain.DashboardResponse{
				HouseholdName:   "Casa",
				TotalExpenses:   7000,
				TotalFixedBills: 4000,
				TotalShared:     8000,
				TotalPersonal:   3000,
				ExpenseCount:    2,
				FixedBillCount:  2,
			}, nil
		},
	}
	h := NewSummaryHandler(svc)

	c, rec := summaryContext(e, http.MethodGet, "/households/hh-1/dashboard")
	if err := h.GetDashboard(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp domain.DashboardResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.HouseholdName != "Casa" {
		t.Errorf("expected Casa, got %s", resp.HouseholdName)
	}
	if resp.TotalExpenses != 7000 {
		t.Errorf("expected 7000, got %d", resp.TotalExpenses)
	}
}

func TestSummaryHandler_GetDashboard_Forbidden(t *testing.T) {
	e := echo.New()
	svc := &mock.SummaryService{
		GetDashboardFn: func(ctx context.Context, householdID, userID string) (*domain.DashboardResponse, error) {
			return nil, domain.ErrForbidden
		},
	}
	h := NewSummaryHandler(svc)

	c, _ := summaryContext(e, http.MethodGet, "/households/hh-1/dashboard")
	err := h.GetDashboard(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", he.Code)
	}
}
