package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func newMockHouseholdService() *mock.HouseholdService {
	return &mock.HouseholdService{}
}

func householdContext(e *echo.Echo, method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
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
	return c, rec
}

func TestHouseholdHandler_Create_Success(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.CreateFn = func(ctx context.Context, input domain.CreateHouseholdInput, userID string) (*domain.Household, error) {
		return &domain.Household{ID: "hh-1", Name: input.Name, InviteCode: "abc123", CreatedAt: time.Now()}, nil
	}
	h := NewHouseholdHandler(svc)

	c, rec := householdContext(e, http.MethodPost, "/households", `{"name":"Casa"}`)
	if err := h.Create(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	var resp domain.HouseholdResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Name != "Casa" {
		t.Errorf("expected name Casa, got %s", resp.Name)
	}
}

func TestHouseholdHandler_Create_Validation(t *testing.T) {
	e := echo.New()
	h := NewHouseholdHandler(nil)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", `{}`},
		{"empty name", `{"name":""}`},
		{"invalid json", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := householdContext(e, http.MethodPost, "/households", tt.body)
			err := h.Create(c)
			if err == nil {
				t.Fatal("expected error")
			}
			he, ok := err.(*echo.HTTPError)
			if !ok {
				t.Fatalf("expected HTTPError, got %T", err)
			}
			if he.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", he.Code)
			}
		})
	}
}

func TestHouseholdHandler_Get_Success(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.GetByIDFn = func(ctx context.Context, id, userID string) (*domain.Household, error) {
		return &domain.Household{ID: id, Name: "Casa", CreatedAt: time.Now()}, nil
	}
	h := NewHouseholdHandler(svc)

	c, rec := householdContext(e, http.MethodGet, "/households/hh-1", "")
	c.SetParamNames("id")
	c.SetParamValues("hh-1")
	if err := h.Get(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHouseholdHandler_Get_Forbidden(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.GetByIDFn = func(ctx context.Context, id, userID string) (*domain.Household, error) {
		return nil, domain.ErrForbidden
	}
	h := NewHouseholdHandler(svc)

	c, _ := householdContext(e, http.MethodGet, "/households/hh-1", "")
	c.SetParamNames("id")
	c.SetParamValues("hh-1")
	err := h.Get(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", he.Code)
	}
}

func TestHouseholdHandler_List_Success(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.ListByUserFn = func(ctx context.Context, userID string) ([]domain.Household, error) {
		return []domain.Household{
			{ID: "hh-1", Name: "Casa", CreatedAt: time.Now()},
		}, nil
	}
	h := NewHouseholdHandler(svc)

	c, rec := householdContext(e, http.MethodGet, "/households", "")
	if err := h.List(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp []domain.HouseholdResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Errorf("expected 1 household, got %d", len(resp))
	}
}

func TestHouseholdHandler_Update_Success(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.UpdateFn = func(ctx context.Context, id string, input domain.UpdateHouseholdInput, userID string) (*domain.Household, error) {
		return &domain.Household{ID: id, Name: input.Name, CreatedAt: time.Now()}, nil
	}
	h := NewHouseholdHandler(svc)

	c, rec := householdContext(e, http.MethodPut, "/households/hh-1", `{"name":"Casa Nova"}`)
	c.SetParamNames("id")
	c.SetParamValues("hh-1")
	if err := h.Update(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHouseholdHandler_Update_Validation(t *testing.T) {
	e := echo.New()
	h := NewHouseholdHandler(nil)

	c, _ := householdContext(e, http.MethodPut, "/households/hh-1", `{"name":""}`)
	c.SetParamNames("id")
	c.SetParamValues("hh-1")
	err := h.Update(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", he.Code)
	}
}

func TestHouseholdHandler_Delete_Success(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.DeleteFn = func(ctx context.Context, id, userID string) error {
		return nil
	}
	h := NewHouseholdHandler(svc)

	c, rec := householdContext(e, http.MethodDelete, "/households/hh-1", "")
	c.SetParamNames("id")
	c.SetParamValues("hh-1")
	if err := h.Delete(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHouseholdHandler_Delete_Forbidden(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.DeleteFn = func(ctx context.Context, id, userID string) error {
		return domain.ErrForbidden
	}
	h := NewHouseholdHandler(svc)

	c, _ := householdContext(e, http.MethodDelete, "/households/hh-1", "")
	c.SetParamNames("id")
	c.SetParamValues("hh-1")
	err := h.Delete(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", he.Code)
	}
}

func TestHouseholdHandler_Join_Success(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.JoinFn = func(ctx context.Context, inviteCode, userID string) (*domain.Household, error) {
		return &domain.Household{ID: "hh-1", Name: "Casa", CreatedAt: time.Now()}, nil
	}
	h := NewHouseholdHandler(svc)

	c, rec := householdContext(e, http.MethodPost, "/households/join", `{"invite_code":"abc123"}`)
	if err := h.Join(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHouseholdHandler_Join_InvalidCode(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.JoinFn = func(ctx context.Context, inviteCode, userID string) (*domain.Household, error) {
		return nil, domain.ErrInvalidInviteCode
	}
	h := NewHouseholdHandler(svc)

	c, _ := householdContext(e, http.MethodPost, "/households/join", `{"invite_code":"bad"}`)
	err := h.Join(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", he.Code)
	}
}

func TestHouseholdHandler_Join_AlreadyMember(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.JoinFn = func(ctx context.Context, inviteCode, userID string) (*domain.Household, error) {
		return nil, domain.ErrAlreadyMember
	}
	h := NewHouseholdHandler(svc)

	c, _ := householdContext(e, http.MethodPost, "/households/join", `{"invite_code":"abc123"}`)
	err := h.Join(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", he.Code)
	}
}

func TestHouseholdHandler_Join_Validation(t *testing.T) {
	e := echo.New()
	h := NewHouseholdHandler(nil)

	c, _ := householdContext(e, http.MethodPost, "/households/join", `{"invite_code":""}`)
	err := h.Join(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", he.Code)
	}
}

func TestHouseholdHandler_ListMembers_Success(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.ListMembersFn = func(ctx context.Context, householdID, userID string) ([]domain.HouseholdMember, error) {
		return []domain.HouseholdMember{
			{UserID: "u1", UserName: "Alice", JoinedAt: time.Now()},
		}, nil
	}
	h := NewHouseholdHandler(svc)

	c, rec := householdContext(e, http.MethodGet, "/households/hh-1/members", "")
	c.SetParamNames("id")
	c.SetParamValues("hh-1")
	if err := h.ListMembers(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHouseholdHandler_UpdateMemberSalary_Success(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.UpdateMemberSalaryFn = func(ctx context.Context, householdID, memberID string, salaryCents int64, userID string) error {
		return nil
	}
	h := NewHouseholdHandler(svc)

	c, rec := householdContext(e, http.MethodPut, "/households/hh-1/members/u1/salary", `{"salary_cents":500000}`)
	c.SetParamNames("id", "memberId")
	c.SetParamValues("hh-1", "u1")
	if err := h.UpdateMemberSalary(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHouseholdHandler_UpdateMemberSalary_NegativeValue(t *testing.T) {
	e := echo.New()
	h := NewHouseholdHandler(nil)

	c, _ := householdContext(e, http.MethodPut, "/households/hh-1/members/u1/salary", `{"salary_cents":-100}`)
	c.SetParamNames("id", "memberId")
	c.SetParamValues("hh-1", "u1")
	err := h.UpdateMemberSalary(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", he.Code)
	}
}

func TestHouseholdHandler_RemoveMember_Success(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.RemoveMemberFn = func(ctx context.Context, householdID, memberID, userID string) error {
		return nil
	}
	h := NewHouseholdHandler(svc)

	c, rec := householdContext(e, http.MethodDelete, "/households/hh-1/members/u1", "")
	c.SetParamNames("id", "memberId")
	c.SetParamValues("hh-1", "u1")
	if err := h.RemoveMember(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHouseholdHandler_RemoveMember_NotMember(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.RemoveMemberFn = func(ctx context.Context, householdID, memberID, userID string) error {
		return domain.ErrNotMember
	}
	h := NewHouseholdHandler(svc)

	c, _ := householdContext(e, http.MethodDelete, "/households/hh-1/members/u1", "")
	c.SetParamNames("id", "memberId")
	c.SetParamValues("hh-1", "u1")
	err := h.RemoveMember(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", he.Code)
	}
}

func TestHouseholdHandler_InternalError(t *testing.T) {
	e := echo.New()
	svc := newMockHouseholdService()
	svc.CreateFn = func(ctx context.Context, input domain.CreateHouseholdInput, userID string) (*domain.Household, error) {
		return nil, errors.New("unexpected")
	}
	h := NewHouseholdHandler(svc)

	c, _ := householdContext(e, http.MethodPost, "/households", `{"name":"Casa"}`)
	err := h.Create(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", he.Code)
	}
}
