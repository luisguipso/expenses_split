package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
	"github.com/lguilherme/contas/internal/parser"
)

func TestImportService_ParseBill_Success(t *testing.T) {
	billParser := &mock.BillParser{
		SupportsFn: func(content []byte) bool { return true },
		ParseFn: func(ctx context.Context, reader io.Reader) (*domain.ParsedBill, error) {
			return &domain.ParsedBill{
				Provider: "nubank",
				Items: []domain.ParsedExpense{
					{Description: "SUPERMERCADO EXTRA", AmountCents: 15000, Date: "2024-01-15"},
					{Description: "FARMACIA RAIA", AmountCents: 5000, Date: "2024-01-16"},
				},
			}, nil
		},
	}
	registry := parser.NewRegistry(billParser)

	catRepo := &mock.CategoryRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string) ([]domain.Category, error) {
			return []domain.Category{
				{ID: "cat-1", Name: "Mercado"},
				{ID: "cat-2", Name: "Farmacia"},
			}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}

	svc := NewImportService(registry, hhRepo, catRepo, nil)

	resp, err := svc.ParseBill(context.Background(), "fatura.pdf", []byte("pdf-content"), "hh-1", "user-1")
	if err != nil {
		t.Fatalf("ParseBill failed: %v", err)
	}
	if resp.Provider != "nubank" {
		t.Errorf("expected provider nubank, got %s", resp.Provider)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].Description != "SUPERMERCADO EXTRA" {
		t.Errorf("expected SUPERMERCADO EXTRA, got %s", resp.Items[0].Description)
	}
	if resp.Items[0].AmountCents != 15000 {
		t.Errorf("expected 15000, got %d", resp.Items[0].AmountCents)
	}
}

func TestImportService_ParseBill_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewImportService(nil, hhRepo, nil, nil)

	_, err := svc.ParseBill(context.Background(), "f.pdf", []byte("x"), "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestImportService_ParseBill_NoParser(t *testing.T) {
	billParser := &mock.BillParser{
		SupportsFn: func(content []byte) bool { return false },
	}
	registry := parser.NewRegistry(billParser)
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}

	svc := NewImportService(registry, hhRepo, nil, nil)

	_, err := svc.ParseBill(context.Background(), "f.pdf", []byte("x"), "hh-1", "user-1")
	if !errors.Is(err, domain.ErrUnsupportedBillFormat) {
		t.Errorf("expected ErrUnsupportedBillFormat, got %v", err)
	}
}

func TestImportService_ParseBill_CategorySuggestion(t *testing.T) {
	billParser := &mock.BillParser{
		SupportsFn: func(content []byte) bool { return true },
		ParseFn: func(ctx context.Context, reader io.Reader) (*domain.ParsedBill, error) {
			return &domain.ParsedBill{
				Provider: "test",
				Items: []domain.ParsedExpense{
					{Description: "SUPERMERCADO EXTRA", AmountCents: 100, Date: "2024-01-01"},
					{Description: "UBER TRIP", AmountCents: 200, Date: "2024-01-02"},
					{Description: "farmacia popular", AmountCents: 300, Date: "2024-01-03"},
				},
			}, nil
		},
	}
	registry := parser.NewRegistry(billParser)

	catRepo := &mock.CategoryRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string) ([]domain.Category, error) {
			return []domain.Category{
				{ID: "cat-mercado", Name: "Mercado"},
				{ID: "cat-farmacia", Name: "Farmacia"},
			}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}

	svc := NewImportService(registry, hhRepo, catRepo, nil)

	resp, err := svc.ParseBill(context.Background(), "f.pdf", []byte("x"), "hh-1", "user-1")
	if err != nil {
		t.Fatalf("ParseBill failed: %v", err)
	}

	// "SUPERMERCADO EXTRA" contains "mercado" (case-insensitive)
	if resp.Items[0].SuggestedCategory != "cat-mercado" {
		t.Errorf("expected cat-mercado, got %s", resp.Items[0].SuggestedCategory)
	}
	// "UBER TRIP" does not match any category
	if resp.Items[1].SuggestedCategory != "" {
		t.Errorf("expected empty suggestion, got %s", resp.Items[1].SuggestedCategory)
	}
	// "farmacia popular" contains "farmacia" (case-insensitive)
	if resp.Items[2].SuggestedCategory != "cat-farmacia" {
		t.Errorf("expected cat-farmacia, got %s", resp.Items[2].SuggestedCategory)
	}
}

func TestImportService_ConfirmImport_Success(t *testing.T) {
	var created []*domain.Expense
	expRepo := &mock.ExpenseRepository{
		CreateBatchFn: func(ctx context.Context, expenses []*domain.Expense) error {
			for i, e := range expenses {
				e.ID = fmt.Sprintf("exp-%d", i+1)
			}
			created = expenses
			return nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}

	svc := NewImportService(nil, hhRepo, nil, expRepo)

	input := domain.ImportConfirmInput{
		Items: []domain.ImportConfirmItem{
			{Description: "Mercado", AmountCents: 15000, ExpenseDate: "2024-01-15", CategoryID: "cat-1", IsShared: true},
			{Description: "Farmacia", AmountCents: 5000, ExpenseDate: "2024-01-16", CategoryID: "cat-2"},
		},
	}

	result, err := svc.ConfirmImport(context.Background(), input, "hh-1", "user-1")
	if err != nil {
		t.Fatalf("ConfirmImport failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(result))
	}
	if result[0].PaidBy != "user-1" {
		t.Errorf("expected PaidBy user-1, got %s", result[0].PaidBy)
	}
	if result[0].HouseholdID != "hh-1" {
		t.Errorf("expected HouseholdID hh-1, got %s", result[0].HouseholdID)
	}
	if result[0].ID != "exp-1" {
		t.Errorf("expected ID exp-1, got %s", result[0].ID)
	}
	if len(created) != 2 {
		t.Errorf("expected CreateBatch called with 2 expenses, got %d", len(created))
	}
}

func TestImportService_ConfirmImport_EmptyItems(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewImportService(nil, hhRepo, nil, nil)

	_, err := svc.ConfirmImport(context.Background(), domain.ImportConfirmInput{Items: nil}, "hh-1", "user-1")
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestImportService_ConfirmImport_InvalidItems(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewImportService(nil, hhRepo, nil, nil)

	tests := []struct {
		name string
		item domain.ImportConfirmItem
	}{
		{"empty description", domain.ImportConfirmItem{Description: "", AmountCents: 100, ExpenseDate: "2024-01-01"}},
		{"zero amount", domain.ImportConfirmItem{Description: "X", AmountCents: 0, ExpenseDate: "2024-01-01"}},
		{"negative amount", domain.ImportConfirmItem{Description: "X", AmountCents: -1, ExpenseDate: "2024-01-01"}},
		{"empty date", domain.ImportConfirmItem{Description: "X", AmountCents: 100, ExpenseDate: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := domain.ImportConfirmInput{Items: []domain.ImportConfirmItem{tt.item}}
			_, err := svc.ConfirmImport(context.Background(), input, "hh-1", "user-1")
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestImportService_ConfirmImport_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewImportService(nil, hhRepo, nil, nil)

	input := domain.ImportConfirmInput{
		Items: []domain.ImportConfirmItem{
			{Description: "X", AmountCents: 100, ExpenseDate: "2024-01-01"},
		},
	}
	_, err := svc.ConfirmImport(context.Background(), input, "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
