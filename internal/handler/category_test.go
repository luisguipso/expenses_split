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

func categoryContext(e *echo.Echo, method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
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

func TestCategoryHandler_Create_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.CategoryService{
		CreateFn: func(ctx context.Context, input domain.CreateCategoryInput, householdID, userID string) (*domain.Category, error) {
			return &domain.Category{ID: "cat-1", Name: input.Name, Icon: input.Icon}, nil
		},
	}
	h := NewCategoryHandler(svc)

	c, rec := categoryContext(e, http.MethodPost, "/households/hh-1/categories", `{"name":"Aluguel","icon":"🏠"}`)
	if err := h.Create(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	var resp domain.CategoryResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Name != "Aluguel" {
		t.Errorf("expected name Aluguel, got %s", resp.Name)
	}
}

func TestCategoryHandler_Create_Validation(t *testing.T) {
	e := echo.New()
	h := NewCategoryHandler(nil)

	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":""}`},
		{"missing name", `{}`},
		{"invalid json", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := categoryContext(e, http.MethodPost, "/households/hh-1/categories", tt.body)
			err := h.Create(c)
			he := err.(*echo.HTTPError)
			if he.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", he.Code)
			}
		})
	}
}

func TestCategoryHandler_Create_Forbidden(t *testing.T) {
	e := echo.New()
	svc := &mock.CategoryService{
		CreateFn: func(ctx context.Context, input domain.CreateCategoryInput, householdID, userID string) (*domain.Category, error) {
			return nil, domain.ErrForbidden
		},
	}
	h := NewCategoryHandler(svc)

	c, _ := categoryContext(e, http.MethodPost, "/households/hh-1/categories", `{"name":"X"}`)
	err := h.Create(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", he.Code)
	}
}

func TestCategoryHandler_Create_Duplicate(t *testing.T) {
	e := echo.New()
	svc := &mock.CategoryService{
		CreateFn: func(ctx context.Context, input domain.CreateCategoryInput, householdID, userID string) (*domain.Category, error) {
			return nil, domain.ErrCategoryExists
		},
	}
	h := NewCategoryHandler(svc)

	c, _ := categoryContext(e, http.MethodPost, "/households/hh-1/categories", `{"name":"Dup"}`)
	err := h.Create(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", he.Code)
	}
}

func TestCategoryHandler_List_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.CategoryService{
		ListFn: func(ctx context.Context, householdID, userID string) ([]domain.Category, error) {
			return []domain.Category{{ID: "c1", Name: "A"}}, nil
		},
	}
	h := NewCategoryHandler(svc)

	c, rec := categoryContext(e, http.MethodGet, "/households/hh-1/categories", "")
	if err := h.List(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp []domain.CategoryResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Errorf("expected 1, got %d", len(resp))
	}
}

func TestCategoryHandler_Update_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.CategoryService{
		UpdateFn: func(ctx context.Context, id string, input domain.UpdateCategoryInput, userID string) (*domain.Category, error) {
			return &domain.Category{ID: id, Name: input.Name, Icon: "📦"}, nil
		},
	}
	h := NewCategoryHandler(svc)

	c, rec := categoryContext(e, http.MethodPut, "/households/hh-1/categories/cat-1", `{"name":"Updated"}`)
	c.SetParamNames("householdId", "id")
	c.SetParamValues("hh-1", "cat-1")
	if err := h.Update(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCategoryHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.CategoryService{
		DeleteFn: func(ctx context.Context, id, userID string) error { return nil },
	}
	h := NewCategoryHandler(svc)

	c, rec := categoryContext(e, http.MethodDelete, "/households/hh-1/categories/cat-1", "")
	c.SetParamNames("householdId", "id")
	c.SetParamValues("hh-1", "cat-1")
	if err := h.Delete(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestCategoryHandler_Delete_NotFound(t *testing.T) {
	e := echo.New()
	svc := &mock.CategoryService{
		DeleteFn: func(ctx context.Context, id, userID string) error { return domain.ErrCategoryNotFound },
	}
	h := NewCategoryHandler(svc)

	c, _ := categoryContext(e, http.MethodDelete, "/households/hh-1/categories/bad", "")
	c.SetParamNames("householdId", "id")
	c.SetParamValues("hh-1", "bad")
	err := h.Delete(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", he.Code)
	}
}
