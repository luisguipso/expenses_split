package service

import (
	"context"
	"errors"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func TestExpenseService_Create_Success(t *testing.T) {
	expRepo := &mock.ExpenseRepository{
		CreateFn: func(ctx context.Context, e *domain.Expense) error {
			e.ID = "exp-1"
			return nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewExpenseService(expRepo, hhRepo)

	exp, err := svc.Create(context.Background(), domain.CreateExpenseInput{
		Description: "Mercado",
		AmountCents: 15000,
		ExpenseDate: "2024-01-15",
		IsShared:    true,
	}, "hh-1", "user-1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if exp.ID != "exp-1" {
		t.Errorf("expected ID exp-1, got %s", exp.ID)
	}
	if exp.PaidBy != "user-1" {
		t.Errorf("expected PaidBy user-1, got %s", exp.PaidBy)
	}
	if exp.ExpenseDate != "2024-01-15" {
		t.Errorf("expected date 2024-01-15, got %s", exp.ExpenseDate)
	}
}

func TestExpenseService_Create_DefaultDate(t *testing.T) {
	expRepo := &mock.ExpenseRepository{
		CreateFn: func(ctx context.Context, e *domain.Expense) error {
			if e.ExpenseDate == "" {
				t.Error("expected non-empty date")
			}
			e.ID = "exp-2"
			return nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewExpenseService(expRepo, hhRepo)

	_, err := svc.Create(context.Background(), domain.CreateExpenseInput{
		Description: "Lanche",
		AmountCents: 2500,
	}, "hh-1", "user-1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestExpenseService_Create_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewExpenseService(nil, hhRepo)

	_, err := svc.Create(context.Background(), domain.CreateExpenseInput{
		Description: "X", AmountCents: 100,
	}, "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestExpenseService_List_Success(t *testing.T) {
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", Description: "Mercado"},
				{ID: "e2", Description: "Luz"},
			}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewExpenseService(expRepo, hhRepo)

	exps, err := svc.List(context.Background(), "hh-1", "user-1", domain.ExpenseFilter{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(exps) != 2 {
		t.Errorf("expected 2 expenses, got %d", len(exps))
	}
}

func TestExpenseService_List_WithFilter(t *testing.T) {
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
			if filter.Month != 1 || filter.Year != 2024 {
				t.Errorf("expected month=1 year=2024, got month=%d year=%d", filter.Month, filter.Year)
			}
			return []domain.Expense{{ID: "e1"}}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewExpenseService(expRepo, hhRepo)

	_, err := svc.List(context.Background(), "hh-1", "user-1", domain.ExpenseFilter{Month: 1, Year: 2024})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestExpenseService_List_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewExpenseService(nil, hhRepo)

	_, err := svc.List(context.Background(), "hh-1", "user-1", domain.ExpenseFilter{})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestExpenseService_Update_Success(t *testing.T) {
	calls := 0
	expRepo := &mock.ExpenseRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Expense, error) {
			calls++
			if calls == 1 {
				return &domain.Expense{ID: id, HouseholdID: "hh-1", Description: "Old"}, nil
			}
			return &domain.Expense{ID: id, HouseholdID: "hh-1", Description: "Updated", AmountCents: 20000}, nil
		},
		UpdateFn: func(ctx context.Context, e *domain.Expense) error { return nil },
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewExpenseService(expRepo, hhRepo)

	exp, err := svc.Update(context.Background(), "exp-1", domain.UpdateExpenseInput{
		Description: "Updated",
		AmountCents: 20000,
		ExpenseDate: "2024-01-20",
	}, "user-1")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if exp.Description != "Updated" {
		t.Errorf("expected Updated, got %s", exp.Description)
	}
}

func TestExpenseService_Update_NotFound(t *testing.T) {
	expRepo := &mock.ExpenseRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Expense, error) {
			return nil, domain.ErrExpenseNotFound
		},
	}
	svc := NewExpenseService(expRepo, nil)

	_, err := svc.Update(context.Background(), "bad", domain.UpdateExpenseInput{}, "user-1")
	if !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("expected ErrExpenseNotFound, got %v", err)
	}
}

func TestExpenseService_Update_NotMember(t *testing.T) {
	expRepo := &mock.ExpenseRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Expense, error) {
			return &domain.Expense{ID: id, HouseholdID: "hh-1"}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewExpenseService(expRepo, hhRepo)

	_, err := svc.Update(context.Background(), "exp-1", domain.UpdateExpenseInput{}, "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestExpenseService_Delete_Success(t *testing.T) {
	expRepo := &mock.ExpenseRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Expense, error) {
			return &domain.Expense{ID: id, HouseholdID: "hh-1"}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error { return nil },
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewExpenseService(expRepo, hhRepo)

	err := svc.Delete(context.Background(), "exp-1", "user-1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestExpenseService_Delete_NotFound(t *testing.T) {
	expRepo := &mock.ExpenseRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Expense, error) {
			return nil, domain.ErrExpenseNotFound
		},
	}
	svc := NewExpenseService(expRepo, nil)

	err := svc.Delete(context.Background(), "bad", "user-1")
	if !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("expected ErrExpenseNotFound, got %v", err)
	}
}

func TestExpenseService_Delete_NotMember(t *testing.T) {
	expRepo := &mock.ExpenseRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Expense, error) {
			return &domain.Expense{ID: id, HouseholdID: "hh-1"}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewExpenseService(expRepo, hhRepo)

	err := svc.Delete(context.Background(), "exp-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
