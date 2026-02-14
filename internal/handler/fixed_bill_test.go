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

func billContext(e *echo.Echo, method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
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

func TestFixedBillHandler_Create_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.FixedBillService{
		CreateFn: func(ctx context.Context, input domain.CreateFixedBillInput, householdID, userID string) (*domain.FixedBill, error) {
			return &domain.FixedBill{
				ID: "bill-1", Description: input.Description,
				AmountCents: input.AmountCents, DueDay: input.DueDay,
				IsShared: input.IsShared, IsActive: true,
			}, nil
		},
	}
	h := NewFixedBillHandler(svc)

	c, rec := billContext(e, http.MethodPost, "/households/hh-1/bills",
		`{"description":"Aluguel","amount_cents":250000,"due_day":5,"is_shared":true}`)
	if err := h.Create(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	var resp domain.FixedBillResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Description != "Aluguel" {
		t.Errorf("expected Aluguel, got %s", resp.Description)
	}
}

func TestFixedBillHandler_Create_Validation(t *testing.T) {
	e := echo.New()
	h := NewFixedBillHandler(nil)

	tests := []struct {
		name string
		body string
	}{
		{"missing description", `{"amount_cents":100,"due_day":1}`},
		{"zero amount", `{"description":"X","amount_cents":0,"due_day":1}`},
		{"negative amount", `{"description":"X","amount_cents":-1,"due_day":1}`},
		{"due_day too low", `{"description":"X","amount_cents":100,"due_day":0}`},
		{"due_day too high", `{"description":"X","amount_cents":100,"due_day":32}`},
		{"invalid json", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := billContext(e, http.MethodPost, "/households/hh-1/bills", tt.body)
			err := h.Create(c)
			he := err.(*echo.HTTPError)
			if he.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", he.Code)
			}
		})
	}
}

func TestFixedBillHandler_Create_Forbidden(t *testing.T) {
	e := echo.New()
	svc := &mock.FixedBillService{
		CreateFn: func(ctx context.Context, input domain.CreateFixedBillInput, householdID, userID string) (*domain.FixedBill, error) {
			return nil, domain.ErrForbidden
		},
	}
	h := NewFixedBillHandler(svc)

	c, _ := billContext(e, http.MethodPost, "/households/hh-1/bills",
		`{"description":"X","amount_cents":100,"due_day":1}`)
	err := h.Create(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", he.Code)
	}
}

func TestFixedBillHandler_List_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.FixedBillService{
		ListFn: func(ctx context.Context, householdID, userID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", Description: "Aluguel"},
				{ID: "b2", Description: "Internet"},
			}, nil
		},
	}
	h := NewFixedBillHandler(svc)

	c, rec := billContext(e, http.MethodGet, "/households/hh-1/bills", "")
	if err := h.List(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp []domain.FixedBillResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 bills, got %d", len(resp))
	}
}

func TestFixedBillHandler_Update_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.FixedBillService{
		UpdateFn: func(ctx context.Context, id string, input domain.UpdateFixedBillInput, userID string) (*domain.FixedBill, error) {
			return &domain.FixedBill{ID: id, Description: input.Description, AmountCents: input.AmountCents, DueDay: input.DueDay, IsActive: true}, nil
		},
	}
	h := NewFixedBillHandler(svc)

	c, rec := billContext(e, http.MethodPut, "/households/hh-1/bills/bill-1",
		`{"description":"Updated","amount_cents":300000,"due_day":10,"is_shared":true,"is_active":true}`)
	c.SetParamNames("householdId", "id")
	c.SetParamValues("hh-1", "bill-1")
	if err := h.Update(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFixedBillHandler_Update_Validation(t *testing.T) {
	e := echo.New()
	h := NewFixedBillHandler(nil)

	tests := []struct {
		name string
		body string
	}{
		{"missing description", `{"amount_cents":100,"due_day":1}`},
		{"zero amount", `{"description":"X","amount_cents":0,"due_day":1}`},
		{"due_day too high", `{"description":"X","amount_cents":100,"due_day":32}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := billContext(e, http.MethodPut, "/households/hh-1/bills/b1", tt.body)
			c.SetParamNames("householdId", "id")
			c.SetParamValues("hh-1", "b1")
			err := h.Update(c)
			he := err.(*echo.HTTPError)
			if he.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", he.Code)
			}
		})
	}
}

func TestFixedBillHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.FixedBillService{
		DeleteFn: func(ctx context.Context, id, userID string) error { return nil },
	}
	h := NewFixedBillHandler(svc)

	c, rec := billContext(e, http.MethodDelete, "/households/hh-1/bills/bill-1", "")
	c.SetParamNames("householdId", "id")
	c.SetParamValues("hh-1", "bill-1")
	if err := h.Delete(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestFixedBillHandler_Delete_NotFound(t *testing.T) {
	e := echo.New()
	svc := &mock.FixedBillService{
		DeleteFn: func(ctx context.Context, id, userID string) error { return domain.ErrFixedBillNotFound },
	}
	h := NewFixedBillHandler(svc)

	c, _ := billContext(e, http.MethodDelete, "/households/hh-1/bills/bad", "")
	c.SetParamNames("householdId", "id")
	c.SetParamValues("hh-1", "bad")
	err := h.Delete(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", he.Code)
	}
}
