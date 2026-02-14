package service

import (
	"context"
	"errors"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func TestFixedBillService_Create_Success(t *testing.T) {
	billRepo := &mock.FixedBillRepository{
		CreateFn: func(ctx context.Context, b *domain.FixedBill) error {
			b.ID = "bill-1"
			b.IsActive = true
			return nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewFixedBillService(billRepo, hhRepo)

	bill, err := svc.Create(context.Background(), domain.CreateFixedBillInput{
		Description: "Aluguel",
		AmountCents: 250000,
		DueDay:      5,
		IsShared:    true,
	}, "hh-1", "user-1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if bill.ID != "bill-1" {
		t.Errorf("expected ID bill-1, got %s", bill.ID)
	}
	if bill.AmountCents != 250000 {
		t.Errorf("expected 250000, got %d", bill.AmountCents)
	}
}

func TestFixedBillService_Create_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewFixedBillService(nil, hhRepo)

	_, err := svc.Create(context.Background(), domain.CreateFixedBillInput{
		Description: "X", AmountCents: 100, DueDay: 1,
	}, "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestFixedBillService_List_Success(t *testing.T) {
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", Description: "Aluguel"},
				{ID: "b2", Description: "Internet"},
			}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewFixedBillService(billRepo, hhRepo)

	bills, err := svc.List(context.Background(), "hh-1", "user-1")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(bills) != 2 {
		t.Errorf("expected 2 bills, got %d", len(bills))
	}
}

func TestFixedBillService_List_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewFixedBillService(nil, hhRepo)

	_, err := svc.List(context.Background(), "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestFixedBillService_Update_Success(t *testing.T) {
	calls := 0
	billRepo := &mock.FixedBillRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.FixedBill, error) {
			calls++
			if calls == 1 {
				return &domain.FixedBill{ID: id, HouseholdID: "hh-1", Description: "Old"}, nil
			}
			return &domain.FixedBill{ID: id, HouseholdID: "hh-1", Description: "Updated", AmountCents: 300000, DueDay: 10, IsActive: true}, nil
		},
		UpdateFn: func(ctx context.Context, b *domain.FixedBill) error { return nil },
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewFixedBillService(billRepo, hhRepo)

	bill, err := svc.Update(context.Background(), "bill-1", domain.UpdateFixedBillInput{
		Description: "Updated",
		AmountCents: 300000,
		DueDay:      10,
		IsShared:    true,
		IsActive:    true,
	}, "user-1")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if bill.Description != "Updated" {
		t.Errorf("expected description Updated, got %s", bill.Description)
	}
}

func TestFixedBillService_Update_NotFound(t *testing.T) {
	billRepo := &mock.FixedBillRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.FixedBill, error) {
			return nil, domain.ErrFixedBillNotFound
		},
	}
	svc := NewFixedBillService(billRepo, nil)

	_, err := svc.Update(context.Background(), "bad", domain.UpdateFixedBillInput{}, "user-1")
	if !errors.Is(err, domain.ErrFixedBillNotFound) {
		t.Errorf("expected ErrFixedBillNotFound, got %v", err)
	}
}

func TestFixedBillService_Update_NotMember(t *testing.T) {
	billRepo := &mock.FixedBillRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.FixedBill, error) {
			return &domain.FixedBill{ID: id, HouseholdID: "hh-1"}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewFixedBillService(billRepo, hhRepo)

	_, err := svc.Update(context.Background(), "bill-1", domain.UpdateFixedBillInput{}, "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestFixedBillService_Delete_Success(t *testing.T) {
	billRepo := &mock.FixedBillRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.FixedBill, error) {
			return &domain.FixedBill{ID: id, HouseholdID: "hh-1"}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error { return nil },
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewFixedBillService(billRepo, hhRepo)

	err := svc.Delete(context.Background(), "bill-1", "user-1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestFixedBillService_Delete_NotFound(t *testing.T) {
	billRepo := &mock.FixedBillRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.FixedBill, error) {
			return nil, domain.ErrFixedBillNotFound
		},
	}
	svc := NewFixedBillService(billRepo, nil)

	err := svc.Delete(context.Background(), "bad", "user-1")
	if !errors.Is(err, domain.ErrFixedBillNotFound) {
		t.Errorf("expected ErrFixedBillNotFound, got %v", err)
	}
}

func TestFixedBillService_Delete_NotMember(t *testing.T) {
	billRepo := &mock.FixedBillRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.FixedBill, error) {
			return &domain.FixedBill{ID: id, HouseholdID: "hh-1"}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewFixedBillService(billRepo, hhRepo)

	err := svc.Delete(context.Background(), "bill-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
