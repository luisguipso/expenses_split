package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func importUploadContext(e *echo.Echo, filename string, content []byte) (echo.Context, *httptest.ResponseRecorder) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	part.Write(content)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/households/hh-1/import/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-1")
	c.SetParamNames("householdId")
	c.SetParamValues("hh-1")
	return c, rec
}

func importConfirmContext(e *echo.Echo, jsonBody string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/households/hh-1/import/confirm", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-1")
	c.SetParamNames("householdId")
	c.SetParamValues("hh-1")
	return c, rec
}

func TestImportHandler_Upload_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.ImportService{
		ParseBillFn: func(ctx context.Context, filename string, content []byte, householdID, userID string) (*domain.ImportPreviewResponse, error) {
			return &domain.ImportPreviewResponse{
				Provider: "nubank",
				Items: []domain.ImportPreviewItem{
					{Description: "Mercado", AmountCents: 15000, Date: "2024-01-15"},
				},
			}, nil
		},
	}
	h := NewImportHandler(svc)

	c, rec := importUploadContext(e, "fatura.pdf", []byte("pdf-content"))
	if err := h.Upload(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp domain.ImportPreviewResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Provider != "nubank" {
		t.Errorf("expected nubank, got %s", resp.Provider)
	}
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
}

func TestImportHandler_Upload_NoFile(t *testing.T) {
	e := echo.New()
	h := NewImportHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/households/hh-1/import/upload", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "user-1")
	c.SetParamNames("householdId")
	c.SetParamValues("hh-1")

	err := h.Upload(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", he.Code)
	}
}

func TestImportHandler_Upload_NonPDF(t *testing.T) {
	e := echo.New()
	h := NewImportHandler(nil)

	c, _ := importUploadContext(e, "data.csv", []byte("csv-content"))
	err := h.Upload(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", he.Code)
	}
}

func TestImportHandler_Upload_UnsupportedFormat(t *testing.T) {
	e := echo.New()
	svc := &mock.ImportService{
		ParseBillFn: func(ctx context.Context, filename string, content []byte, householdID, userID string) (*domain.ImportPreviewResponse, error) {
			return nil, fmt.Errorf("%w: unknown", domain.ErrUnsupportedBillFormat)
		},
	}
	h := NewImportHandler(svc)

	c, _ := importUploadContext(e, "fatura.pdf", []byte("pdf-content"))
	err := h.Upload(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", he.Code)
	}
}

func TestImportHandler_Upload_Forbidden(t *testing.T) {
	e := echo.New()
	svc := &mock.ImportService{
		ParseBillFn: func(ctx context.Context, filename string, content []byte, householdID, userID string) (*domain.ImportPreviewResponse, error) {
			return nil, domain.ErrForbidden
		},
	}
	h := NewImportHandler(svc)

	c, _ := importUploadContext(e, "fatura.pdf", []byte("pdf-content"))
	err := h.Upload(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", he.Code)
	}
}

func TestImportHandler_Confirm_Success(t *testing.T) {
	e := echo.New()
	svc := &mock.ImportService{
		ConfirmImportFn: func(ctx context.Context, input domain.ImportConfirmInput, householdID, userID string) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "exp-1", Description: input.Items[0].Description, AmountCents: input.Items[0].AmountCents, PaidBy: userID},
			}, nil
		},
	}
	h := NewImportHandler(svc)

	c, rec := importConfirmContext(e, `{"items":[{"description":"Mercado","amount_cents":15000,"expense_date":"2024-01-15","category_id":"cat-1","is_shared":true}]}`)
	if err := h.Confirm(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var resp []domain.ExpenseResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(resp))
	}
	if resp[0].Description != "Mercado" {
		t.Errorf("expected Mercado, got %s", resp[0].Description)
	}
}

func TestImportHandler_Confirm_EmptyItems(t *testing.T) {
	e := echo.New()
	h := NewImportHandler(nil)

	c, _ := importConfirmContext(e, `{"items":[]}`)
	err := h.Confirm(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", he.Code)
	}
}

func TestImportHandler_Confirm_Forbidden(t *testing.T) {
	e := echo.New()
	svc := &mock.ImportService{
		ConfirmImportFn: func(ctx context.Context, input domain.ImportConfirmInput, householdID, userID string) ([]domain.Expense, error) {
			return nil, domain.ErrForbidden
		},
	}
	h := NewImportHandler(svc)

	c, _ := importConfirmContext(e, `{"items":[{"description":"X","amount_cents":100,"expense_date":"2024-01-01"}]}`)
	err := h.Confirm(c)
	he := err.(*echo.HTTPError)
	if he.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", he.Code)
	}
}
