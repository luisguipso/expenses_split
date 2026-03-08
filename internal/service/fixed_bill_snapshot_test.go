package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

// --- HasDueDatePassed tests (now a domain method on FixedBill) ---

func TestFixedBill_HasDueDatePassed(t *testing.T) {
	tests := []struct {
		name    string
		year    int
		month   int
		dueDay  int
		now     time.Time
		want    bool
	}{
		{
			name:   "due day in the past",
			year:   2026, month: 1, dueDay: 5,
			now:    time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "due day today before end of day",
			year:   2026, month: 1, dueDay: 10,
			now:    time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "due day in the future",
			year:   2026, month: 1, dueDay: 20,
			now:    time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "due day 31 in february normalizes to 28",
			year:   2026, month: 2, dueDay: 31,
			now:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "due day 31 in february not yet passed",
			year:   2026, month: 2, dueDay: 31,
			now:    time.Date(2026, 2, 27, 12, 0, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "due day 30 in april (30-day month)",
			year:   2026, month: 4, dueDay: 30,
			now:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			want:   true,
		},
		{
			name:   "future month not yet passed",
			year:   2026, month: 6, dueDay: 5,
			now:    time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC),
			want:   false,
		},
		{
			name:   "past month always passed",
			year:   2025, month: 12, dueDay: 25,
			now:    time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bill := &domain.FixedBill{DueDay: tt.dueDay}
			got := bill.HasDueDatePassed(tt.year, tt.month, tt.now)
			if got != tt.want {
				t.Errorf("HasDueDatePassed(%d, %d, %v) = %v, want %v",
					tt.year, tt.month, tt.now, got, tt.want)
			}
		})
	}
}

// --- Snapshot resolution tests ---

func TestSummaryService_Generate_UsesSnapshotWhenFrozen(t *testing.T) {
	// Bill with due_day=5, and an existing snapshot with different amount
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", HouseholdID: "hh-1", Description: "Aluguel", AmountCents: 200000, DueDay: 5, IsShared: true, PaidBy: "u1", IsActive: true},
			}, nil
		},
	}
	snapRepo := &mock.FixedBillSnapshotRepository{
		FindByMonthFn: func(ctx context.Context, householdID string, year, month int) ([]domain.FixedBillSnapshot, error) {
			return []domain.FixedBillSnapshot{
				{ID: "snap-1", FixedBillID: "b1", AmountCents: 150000, DueDay: 5, IsShared: true, PaidBy: "u1", Description: "Aluguel Antigo"},
			}, nil
		},
		CreateFn: snapshotCreateOK(),
	}
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo, snapRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2026, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should use snapshot amount (150000), not live amount (200000)
	if resp.TotalSharedCents != 150000 {
		t.Errorf("expected TotalSharedCents 150000 (from snapshot), got %d", resp.TotalSharedCents)
	}
	if len(resp.FixedBills) != 1 {
		t.Fatalf("expected 1 fixed bill, got %d", len(resp.FixedBills))
	}
	if !resp.FixedBills[0].IsFrozen {
		t.Error("expected bill to be frozen")
	}
	if resp.FixedBills[0].AmountCents != 150000 {
		t.Errorf("expected snapshot amount 150000, got %d", resp.FixedBills[0].AmountCents)
	}
	if resp.FixedBills[0].Description != "Aluguel Antigo" {
		t.Errorf("expected snapshot description, got %s", resp.FixedBills[0].Description)
	}
}

func TestSummaryService_Generate_LiveBillWhenNotFrozen(t *testing.T) {
	// Bill with due_day=25, current month, no snapshot exists
	// The due_day hasn't passed so it should use live values
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", HouseholdID: "hh-1", Description: "Internet", AmountCents: 10000, DueDay: 25, IsShared: true, PaidBy: "u1", IsActive: true},
			}, nil
		},
	}
	snapRepo := &mock.FixedBillSnapshotRepository{
		FindByMonthFn: noSnapshots(),
		CreateFn:      snapshotCreateOK(),
	}
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}

	// Use a future month so due_day=25 hasn't passed
	now := time.Now()
	futureYear := now.Year() + 1
	futureMonth := int(now.Month())

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo, snapRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", futureYear, futureMonth, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(resp.FixedBills) != 1 {
		t.Fatalf("expected 1 fixed bill, got %d", len(resp.FixedBills))
	}
	if resp.FixedBills[0].IsFrozen {
		t.Error("expected bill to NOT be frozen (due_day in future)")
	}
	if resp.FixedBills[0].AmountCents != 10000 {
		t.Errorf("expected live amount 10000, got %d", resp.FixedBills[0].AmountCents)
	}
}

func TestSummaryService_Generate_CreatesSnapshotWhenDueDayPassed(t *testing.T) {
	// Bill with due_day=1, past month, no snapshot exists — should create one
	var createdSnapshot *domain.FixedBillSnapshot
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", HouseholdID: "hh-1", Description: "Aluguel", AmountCents: 250000, DueDay: 1, IsShared: true, PaidBy: "u1", IsActive: true},
			}, nil
		},
	}
	snapRepo := &mock.FixedBillSnapshotRepository{
		FindByMonthFn: noSnapshots(),
		CreateFn: func(ctx context.Context, s *domain.FixedBillSnapshot) error {
			createdSnapshot = s
			s.ID = "snap-new"
			s.FrozenAt = time.Now()
			return nil
		},
	}
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}

	// Past month
	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo, snapRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2025, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if createdSnapshot == nil {
		t.Fatal("expected snapshot to be created")
	}
	if createdSnapshot.FixedBillID != "b1" {
		t.Errorf("expected snapshot for bill b1, got %s", createdSnapshot.FixedBillID)
	}
	if createdSnapshot.AmountCents != 250000 {
		t.Errorf("expected snapshot amount 250000, got %d", createdSnapshot.AmountCents)
	}
	if createdSnapshot.Year != 2025 || createdSnapshot.Month != 1 {
		t.Errorf("expected snapshot for 2025/1, got %d/%d", createdSnapshot.Year, createdSnapshot.Month)
	}

	if len(resp.FixedBills) != 1 {
		t.Fatalf("expected 1 fixed bill, got %d", len(resp.FixedBills))
	}
	if !resp.FixedBills[0].IsFrozen {
		t.Error("expected bill to be frozen after snapshot creation")
	}
}

func TestSummaryService_Generate_InactiveBillsExcludedFromSnapshots(t *testing.T) {
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", Description: "Active", AmountCents: 10000, DueDay: 5, IsShared: true, PaidBy: "u1", IsActive: true},
				{ID: "b2", Description: "Inactive", AmountCents: 20000, DueDay: 5, IsShared: true, PaidBy: "u1", IsActive: false},
			}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo, nil)
	resp, err := svc.Generate(context.Background(), "hh-1", 2025, 6, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(resp.FixedBills) != 1 {
		t.Fatalf("expected 1 fixed bill (only active), got %d", len(resp.FixedBills))
	}
	if resp.FixedBills[0].Description != "Active" {
		t.Errorf("expected active bill, got %s", resp.FixedBills[0].Description)
	}
	// Only active bill's amount should be counted
	if resp.TotalSharedCents != 10000 {
		t.Errorf("expected TotalSharedCents 10000, got %d", resp.TotalSharedCents)
	}
}

func TestSummaryService_Generate_MixedFrozenAndLiveBills(t *testing.T) {
	// Two bills: one with existing snapshot, one without (future due_day)
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", Description: "Rent", AmountCents: 200000, DueDay: 5, IsShared: true, PaidBy: "u1", IsActive: true},
				{ID: "b2", Description: "Internet", AmountCents: 10000, DueDay: 25, IsShared: true, PaidBy: "u2", IsActive: true},
			}, nil
		},
	}
	snapRepo := &mock.FixedBillSnapshotRepository{
		FindByMonthFn: func(ctx context.Context, householdID string, year, month int) ([]domain.FixedBillSnapshot, error) {
			return []domain.FixedBillSnapshot{
				{ID: "snap-1", FixedBillID: "b1", AmountCents: 180000, DueDay: 5, IsShared: true, PaidBy: "u1", Description: "Rent Old"},
			}, nil
		},
		CreateFn: snapshotCreateOK(),
	}
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}

	// Use future month so b2's due_day=25 hasn't passed
	now := time.Now()
	futureYear := now.Year() + 1

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo, snapRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", futureYear, 3, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(resp.FixedBills) != 2 {
		t.Fatalf("expected 2 fixed bills, got %d", len(resp.FixedBills))
	}

	// b1 should be frozen with snapshot amount
	if !resp.FixedBills[0].IsFrozen {
		t.Error("expected b1 to be frozen")
	}
	if resp.FixedBills[0].AmountCents != 180000 {
		t.Errorf("expected b1 snapshot amount 180000, got %d", resp.FixedBills[0].AmountCents)
	}

	// b2 should be live
	if resp.FixedBills[1].IsFrozen {
		t.Error("expected b2 to NOT be frozen")
	}
	if resp.FixedBills[1].AmountCents != 10000 {
		t.Errorf("expected b2 live amount 10000, got %d", resp.FixedBills[1].AmountCents)
	}

	// Total should use snapshot for b1 + live for b2
	expectedTotal := int64(180000 + 10000)
	if resp.TotalSharedCents != expectedTotal {
		t.Errorf("expected TotalSharedCents %d, got %d", expectedTotal, resp.TotalSharedCents)
	}
}

// --- FixedBillSnapshotService tests ---

func TestFixedBillSnapshotService_Update_Success(t *testing.T) {
	calls := 0
	snapRepo := &mock.FixedBillSnapshotRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.FixedBillSnapshot, error) {
			calls++
			if calls == 1 {
				return &domain.FixedBillSnapshot{ID: id, HouseholdID: "hh-1", FixedBillID: "b1", Description: "Old", AmountCents: 100000}, nil
			}
			return &domain.FixedBillSnapshot{ID: id, HouseholdID: "hh-1", FixedBillID: "b1", Description: "Updated", AmountCents: 120000}, nil
		},
		UpdateFn: func(ctx context.Context, s *domain.FixedBillSnapshot) error { return nil },
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewFixedBillSnapshotService(snapRepo, hhRepo)

	snap, err := svc.Update(context.Background(), "snap-1", domain.UpdateFixedBillSnapshotInput{
		Description: "Updated",
		AmountCents: 120000,
		DueDay:      5,
		IsShared:    true,
		PaidBy:      "u1",
	}, "user-1")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if snap.Description != "Updated" {
		t.Errorf("expected description Updated, got %s", snap.Description)
	}
	if snap.AmountCents != 120000 {
		t.Errorf("expected amount 120000, got %d", snap.AmountCents)
	}
}

func TestFixedBillSnapshotService_Update_NotFound(t *testing.T) {
	snapRepo := &mock.FixedBillSnapshotRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.FixedBillSnapshot, error) {
			return nil, domain.ErrFixedBillSnapshotNotFound
		},
	}
	svc := NewFixedBillSnapshotService(snapRepo, nil)

	_, err := svc.Update(context.Background(), "bad", domain.UpdateFixedBillSnapshotInput{}, "user-1")
	if !errors.Is(err, domain.ErrFixedBillSnapshotNotFound) {
		t.Errorf("expected ErrFixedBillSnapshotNotFound, got %v", err)
	}
}

func TestFixedBillSnapshotService_Update_NotMember(t *testing.T) {
	snapRepo := &mock.FixedBillSnapshotRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.FixedBillSnapshot, error) {
			return &domain.FixedBillSnapshot{ID: id, HouseholdID: "hh-1"}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewFixedBillSnapshotService(snapRepo, hhRepo)

	_, err := svc.Update(context.Background(), "snap-1", domain.UpdateFixedBillSnapshotInput{}, "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
