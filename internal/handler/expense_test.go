package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func expenseContext(e *echo.Echo, method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-1")
	c.SetParamNames("householdId")
	c.SetParamValues("hh-1")
	return c, rec
}

func TestExpenseHandler_Create_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.ExpenseService{
		CreateFn: func(ctx context.Context, input domain.CreateExpenseInput, householdID, userID string) (*domain.Expense, error) {
			return &domain.Expense{
				ID: "exp-1", Description: input.Description,
				AmountCents: input.AmountCents, PaidBy: userID,
			}, nil
		},
	}
	h := NewExpenseHandler(svc)

	c, rec := expenseContext(e, http.MethodPost, "/households/hh-1/expenses",
		`{"description":"Mercado","amount_cents":15000,"expense_date":"2024-01-15","is_shared":true}`)
	if err := h.Create(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	var resp domain.ExpenseResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Description != "Mercado" {
		t.Errorf("expected Mercado, got %s", resp.Description)
	}
}

func TestExpenseHandler_Create_Validation(t *testing.T) {
	e := echo.New()
	h := NewExpenseHandler(nil)

	tests := []struct {
		name string
		body string
	}{
		{"missing description", `{"amount_cents":100}`},
		{"zero amount", `{"description":"X","amount_cents":0}`},
		{"negative amount", `{"description":"X","amount_cents":-1}`},
		{"invalid json", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := expenseContext(e, http.MethodPost, "/households/hh-1/expenses", tt.body)
			err := h.Create(c)
			he := err.(*echo.HTTPError)
			if he.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", he.Code)
			}
		})
	}
}

func TestExpenseHandler_Create_Forbidden(t *testing.T) {
	e := echo.New()
	svc := &mock.ExpenseService{
		CreateFn: func(ctx context.Context, input domain.CreateExpenseInput, householdID, userID string) (*domain.Expense, error) {
			return nil, domain.ErrForbidden
		},
	}
	h := NewExpenseHandler(svc)

	c, _ := expenseContext(e, http.MethodPost, "/households/hh-1/expenses",
		`{"description":"X","amount_cents":100}`)
	err := h.Create(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", he.Code)
	}
}

func TestExpenseHandler_List_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.ExpenseService{
		ListFn: func(ctx context.Context, householdID, userID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", Description: "Mercado"},
				{ID: "e2", Description: "Luz"},
			}, nil
		},
	}
	h := NewExpenseHandler(svc)

	c, rec := expenseContext(e, http.MethodGet, "/households/hh-1/expenses", "")
	if err := h.List(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp []domain.ExpenseResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2, got %d", len(resp))
	}
}

func TestExpenseHandler_List_WithFilters(t *testing.T) {
	e := echo.New()
	svc := &mock.ExpenseService{
		ListFn: func(ctx context.Context, householdID, userID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
			if filter.Month != 1 || filter.Year != 2024 {
				t.Errorf("expected month=1 year=2024, got %d %d", filter.Month, filter.Year)
			}
			if filter.CategoryID != "cat-1" {
				t.Errorf("expected category_id cat-1, got %s", filter.CategoryID)
			}
			return []domain.Expense{}, nil
		},
	}
	h := NewExpenseHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/households/hh-1/expenses?month=1&year=2024&category_id=cat-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-1")
	c.SetParamNames("householdId")
	c.SetParamValues("hh-1")

	if err := h.List(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExpenseHandler_Update_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.ExpenseService{
		UpdateFn: func(ctx context.Context, id string, input domain.UpdateExpenseInput, userID string) (*domain.Expense, error) {
			return &domain.Expense{ID: id, Description: input.Description, AmountCents: input.AmountCents}, nil
		},
	}
	h := NewExpenseHandler(svc)

	c, rec := expenseContext(e, http.MethodPut, "/households/hh-1/expenses/exp-1",
		`{"description":"Updated","amount_cents":20000,"expense_date":"2024-01-20"}`)
	c.SetParamNames("householdId", "id")
	c.SetParamValues("hh-1", "exp-1")
	if err := h.Update(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestExpenseHandler_Update_Validation(t *testing.T) {
	e := echo.New()
	h := NewExpenseHandler(nil)

	tests := []struct {
		name string
		body string
	}{
		{"missing description", `{"amount_cents":100,"expense_date":"2024-01-01"}`},
		{"zero amount", `{"description":"X","amount_cents":0,"expense_date":"2024-01-01"}`},
		{"missing date", `{"description":"X","amount_cents":100}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := expenseContext(e, http.MethodPut, "/households/hh-1/expenses/e1", tt.body)
			c.SetParamNames("householdId", "id")
			c.SetParamValues("hh-1", "e1")
			err := h.Update(c)
			he := err.(*echo.HTTPError)
			if he.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", he.Code)
			}
		})
	}
}

func TestExpenseHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.ExpenseService{
		DeleteFn: func(ctx context.Context, id, userID string) error { return nil },
	}
	h := NewExpenseHandler(svc)

	c, rec := expenseContext(e, http.MethodDelete, "/households/hh-1/expenses/exp-1", "")
	c.SetParamNames("householdId", "id")
	c.SetParamValues("hh-1", "exp-1")
	if err := h.Delete(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestExpenseHandler_Delete_NotFound(t *testing.T) {
	e := echo.New()
	svc := &mock.ExpenseService{
		DeleteFn: func(ctx context.Context, id, userID string) error { return domain.ErrExpenseNotFound },
	}
	h := NewExpenseHandler(svc)

	c, _ := expenseContext(e, http.MethodDelete, "/households/hh-1/expenses/bad", "")
	c.SetParamNames("householdId", "id")
	c.SetParamValues("hh-1", "bad")
	err := h.Delete(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", he.Code)
	}
}
