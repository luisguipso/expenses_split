package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func summaryUpsertOK() func(ctx context.Context, s *domain.MonthlySummary) error {
	return func(ctx context.Context, s *domain.MonthlySummary) error {
		s.ID = "sum-1"
		s.GeneratedAt = time.Now()
		return nil
	}
}

func noExpenses() func(ctx context.Context, householdID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
	return func(ctx context.Context, householdID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
		return nil, nil
	}
}

func noBills() func(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
	return func(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
		return nil, nil
	}
}

func twoMembers() func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
	return func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
		return []domain.HouseholdMember{
			{UserID: "u1", UserName: "Alice", SalaryCents: 500000},
			{UserID: "u2", UserName: "Bob", SalaryCents: 300000},
		}, nil
	}
}

func threeMembersEqualSalary() func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
	return func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
		return []domain.HouseholdMember{
			{UserID: "u1", UserName: "Alice", SalaryCents: 300000},
			{UserID: "u2", UserName: "Bob", SalaryCents: 300000},
			{UserID: "u3", UserName: "Carol", SalaryCents: 300000},
		}, nil
	}
}

func makeSummaryService(
	hhRepo *mock.HouseholdRepository,
	expRepo *mock.ExpenseRepository,
	billRepo *mock.FixedBillRepository,
	sumRepo *mock.SummaryRepository,
) domain.SummaryService {
	return NewSummaryService(sumRepo, hhRepo, expRepo, billRepo)
}

// --- Generate tests ---

func TestSummaryService_Generate_SharedOnlyProportional(t *testing.T) {
	// Alice: 5000, Bob: 3000 => Alice 62.5%, Bob 37.5%
	// Total shared: 10000 cents (1 expense)
	// Alice should pay ~6250, Bob ~3750
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 10000, IsShared: true, PaidBy: "u1"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.TotalSharedCents != 10000 {
		t.Errorf("expected total shared 10000, got %d", resp.TotalSharedCents)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}

	// Alice: 5000/8000 = 62.5% of 10000 = 6250
	alice := resp.Items[0]
	if alice.UserID != "u1" {
		t.Errorf("expected u1, got %s", alice.UserID)
	}
	if alice.TotalSharedCents != 6250 {
		t.Errorf("Alice shared: expected 6250, got %d", alice.TotalSharedCents)
	}
	if alice.TotalPersonalCents != 0 {
		t.Errorf("Alice personal: expected 0, got %d", alice.TotalPersonalCents)
	}
	if alice.AmountDueCents != 6250 {
		t.Errorf("Alice total: expected 6250, got %d", alice.AmountDueCents)
	}

	// Bob: 3000/8000 = 37.5% of 10000 = 3750
	bob := resp.Items[1]
	if bob.TotalSharedCents != 3750 {
		t.Errorf("Bob shared: expected 3750, got %d", bob.TotalSharedCents)
	}
	if bob.AmountDueCents != 3750 {
		t.Errorf("Bob total: expected 3750, got %d", bob.AmountDueCents)
	}

	// Verify totals add up
	if alice.TotalSharedCents+bob.TotalSharedCents != 10000 {
		t.Errorf("shared should sum to 10000, got %d", alice.TotalSharedCents+bob.TotalSharedCents)
	}
}

func TestSummaryService_Generate_PersonalExpenses(t *testing.T) {
	// Shared: 0, Personal: u1 has 5000, u2 has 2000
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 5000, IsShared: false, PaidBy: "u1", AssignedTo: "u1"},
				{ID: "e2", AmountCents: 2000, IsShared: false, PaidBy: "u2", AssignedTo: "u2"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.TotalSharedCents != 0 {
		t.Errorf("expected 0 shared, got %d", resp.TotalSharedCents)
	}

	alice := resp.Items[0]
	if alice.TotalSharedCents != 0 {
		t.Errorf("Alice shared: expected 0, got %d", alice.TotalSharedCents)
	}
	if alice.TotalPersonalCents != 5000 {
		t.Errorf("Alice personal: expected 5000, got %d", alice.TotalPersonalCents)
	}
	if alice.AmountDueCents != 5000 {
		t.Errorf("Alice total: expected 5000, got %d", alice.AmountDueCents)
	}

	bob := resp.Items[1]
	if bob.TotalPersonalCents != 2000 {
		t.Errorf("Bob personal: expected 2000, got %d", bob.TotalPersonalCents)
	}
}

func TestSummaryService_Generate_MixedSharedAndPersonal(t *testing.T) {
	// Shared expense: 8000, Shared bill: 2000 => total shared = 10000
	// Personal expense u1: 1500
	// Alice: 62.5% of 10000 = 6250 + 1500 personal = 7750
	// Bob: 37.5% of 10000 = 3750 + 0 personal = 3750
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 8000, IsShared: true, PaidBy: "u1"},
				{ID: "e2", AmountCents: 1500, IsShared: false, PaidBy: "u1", AssignedTo: "u1"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", AmountCents: 2000, IsShared: true, IsActive: true},
			}, nil
		},
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.TotalSharedCents != 10000 {
		t.Errorf("expected total shared 10000, got %d", resp.TotalSharedCents)
	}

	alice := resp.Items[0]
	if alice.TotalSharedCents != 6250 {
		t.Errorf("Alice shared: expected 6250, got %d", alice.TotalSharedCents)
	}
	if alice.TotalPersonalCents != 1500 {
		t.Errorf("Alice personal: expected 1500, got %d", alice.TotalPersonalCents)
	}
	if alice.AmountDueCents != 7750 {
		t.Errorf("Alice total: expected 7750, got %d", alice.AmountDueCents)
	}

	bob := resp.Items[1]
	if bob.TotalSharedCents != 3750 {
		t.Errorf("Bob shared: expected 3750, got %d", bob.TotalSharedCents)
	}
	if bob.AmountDueCents != 3750 {
		t.Errorf("Bob total: expected 3750, got %d", bob.AmountDueCents)
	}
}

func TestSummaryService_Generate_FixedBillsSharedAndPersonal(t *testing.T) {
	// Shared bill: 6000, Personal bill assigned to u2: 3000
	// No expenses
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", AmountCents: 6000, IsShared: true, IsActive: true},
				{ID: "b2", AmountCents: 3000, IsShared: false, AssignedTo: "u2", IsActive: true},
			}, nil
		},
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Alice: 62.5% of 6000 = 3750
	alice := resp.Items[0]
	if alice.TotalSharedCents != 3750 {
		t.Errorf("Alice shared: expected 3750, got %d", alice.TotalSharedCents)
	}
	if alice.TotalPersonalCents != 0 {
		t.Errorf("Alice personal: expected 0, got %d", alice.TotalPersonalCents)
	}

	// Bob: 37.5% of 6000 = 2250 + 3000 personal = 5250
	bob := resp.Items[1]
	if bob.TotalSharedCents != 2250 {
		t.Errorf("Bob shared: expected 2250, got %d", bob.TotalSharedCents)
	}
	if bob.TotalPersonalCents != 3000 {
		t.Errorf("Bob personal: expected 3000, got %d", bob.TotalPersonalCents)
	}
	if bob.AmountDueCents != 5250 {
		t.Errorf("Bob total: expected 5250, got %d", bob.AmountDueCents)
	}
}

func TestSummaryService_Generate_EqualSalaryThreeWay(t *testing.T) {
	// 3 members equal salary, shared 10000
	// Each should pay 3333 or 3334 (with rounding, last gets remainder)
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: threeMembersEqualSalary(),
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 10000, IsShared: true, PaidBy: "u1"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// All items should sum to exactly 10000
	var total int64
	for _, item := range resp.Items {
		total += item.TotalSharedCents
	}
	if total != 10000 {
		t.Errorf("total shared should be exactly 10000, got %d", total)
	}

	// First two get 3333 (rounded), last gets remainder 3334
	if resp.Items[0].TotalSharedCents != 3333 {
		t.Errorf("u1 expected 3333, got %d", resp.Items[0].TotalSharedCents)
	}
	if resp.Items[1].TotalSharedCents != 3333 {
		t.Errorf("u2 expected 3333, got %d", resp.Items[1].TotalSharedCents)
	}
	if resp.Items[2].TotalSharedCents != 3334 {
		t.Errorf("u3 expected 3334 (remainder), got %d", resp.Items[2].TotalSharedCents)
	}
}

func TestSummaryService_Generate_InactiveBillsExcluded(t *testing.T) {
	// Active bill: 5000 shared, Inactive bill: 9999 shared (should be excluded)
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", AmountCents: 5000, IsShared: true, IsActive: true},
				{ID: "b2", AmountCents: 9999, IsShared: true, IsActive: false},
			}, nil
		},
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.TotalSharedCents != 5000 {
		t.Errorf("expected 5000 (inactive excluded), got %d", resp.TotalSharedCents)
	}
}

func TestSummaryService_Generate_PersonalExpenseNoAssignedTo(t *testing.T) {
	// Personal expense with no assigned_to should fall back to paid_by
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 4000, IsShared: false, PaidBy: "u2", AssignedTo: ""},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	bob := resp.Items[1]
	if bob.TotalPersonalCents != 4000 {
		t.Errorf("Bob personal: expected 4000, got %d", bob.TotalPersonalCents)
	}
}

func TestSummaryService_Generate_ZeroTotalNoData(t *testing.T) {
	// No expenses and no bills => everyone pays 0
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	for _, item := range resp.Items {
		if item.AmountDueCents != 0 {
			t.Errorf("%s should owe 0, got %d", item.UserName, item.AmountDueCents)
		}
	}
}

func TestSummaryService_Generate_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewSummaryService(nil, hhRepo, nil, nil)

	_, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestSummaryService_Generate_NoSalary(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn: memberOK(),
		ListMembersFn: func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
			return []domain.HouseholdMember{
				{UserID: "u1", SalaryCents: 0},
				{UserID: "u2", SalaryCents: 0},
			}, nil
		},
	}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}

	svc := NewSummaryService(nil, hhRepo, expRepo, billRepo)
	_, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if !errors.Is(err, domain.ErrNoMembersWithSalary) {
		t.Errorf("expected ErrNoMembersWithSalary, got %v", err)
	}
}

func TestSummaryService_Generate_SingleMember(t *testing.T) {
	// Single member pays 100% of shared expenses
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn: memberOK(),
		ListMembersFn: func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
			return []domain.HouseholdMember{
				{UserID: "u1", UserName: "Alice", SalaryCents: 500000},
			}, nil
		},
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 15000, IsShared: true, PaidBy: "u1"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	if resp.Items[0].TotalSharedCents != 15000 {
		t.Errorf("expected 15000, got %d", resp.Items[0].TotalSharedCents)
	}
	if resp.Items[0].Proportion != 1.0 {
		t.Errorf("expected proportion 1.0, got %f", resp.Items[0].Proportion)
	}
	if resp.TotalAllCents != 15000 {
		t.Errorf("expected TotalAllCents 15000, got %d", resp.TotalAllCents)
	}
}

func TestSummaryService_Generate_UnevenRoundingTwoMembers(t *testing.T) {
	// Alice: 7000, Bob: 3000 => Alice 70%, Bob 30%
	// Total shared: 9999 cents
	// Alice: 70% of 9999 = 6999.3 => rounds to 6999
	// Bob: remainder = 9999 - 6999 = 3000
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn: memberOK(),
		ListMembersFn: func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
			return []domain.HouseholdMember{
				{UserID: "u1", UserName: "Alice", SalaryCents: 700000},
				{UserID: "u2", UserName: "Bob", SalaryCents: 300000},
			}, nil
		},
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 9999, IsShared: true, PaidBy: "u1"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Must always sum to exact total
	total := resp.Items[0].TotalSharedCents + resp.Items[1].TotalSharedCents
	if total != 9999 {
		t.Errorf("shared should sum to exactly 9999, got %d", total)
	}
	if resp.TotalSharedCents != 9999 {
		t.Errorf("expected TotalSharedCents 9999, got %d", resp.TotalSharedCents)
	}
}

func TestSummaryService_Generate_LargeRealisticAmounts(t *testing.T) {
	// Realistic BRL scenario:
	// Alice: R$8000 salary, Bob: R$5000 salary, Carol: R$3000 salary
	// Shared: R$3000 rent + R$500 internet + R$1200 groceries = R$4700 (470000 cents)
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn: memberOK(),
		ListMembersFn: func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
			return []domain.HouseholdMember{
				{UserID: "u1", UserName: "Alice", SalaryCents: 800000},
				{UserID: "u2", UserName: "Bob", SalaryCents: 500000},
				{UserID: "u3", UserName: "Carol", SalaryCents: 300000},
			}, nil
		},
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 120000, IsShared: true, PaidBy: "u1"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", AmountCents: 300000, IsShared: true, IsActive: true},
				{ID: "b2", AmountCents: 50000, IsShared: true, IsActive: true},
			}, nil
		},
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Total shared: 470000
	if resp.TotalSharedCents != 470000 {
		t.Errorf("expected TotalSharedCents 470000, got %d", resp.TotalSharedCents)
	}

	// Verify sum adds up exactly
	var totalShared int64
	for _, item := range resp.Items {
		totalShared += item.TotalSharedCents
	}
	if totalShared != 470000 {
		t.Errorf("shared should sum to exactly 470000, got %d", totalShared)
	}

	// Alice: 800000/1600000 = 50%
	if resp.Items[0].Proportion != 0.5 {
		t.Errorf("Alice proportion: expected 0.5, got %f", resp.Items[0].Proportion)
	}
	// Bob: 500000/1600000 = 31.25%
	if resp.Items[1].Proportion != 0.3125 {
		t.Errorf("Bob proportion: expected 0.3125, got %f", resp.Items[1].Proportion)
	}
	// Carol: 300000/1600000 = 18.75%
	if resp.Items[2].Proportion != 0.1875 {
		t.Errorf("Carol proportion: expected 0.1875, got %f", resp.Items[2].Proportion)
	}
}

func TestSummaryService_Generate_TotalAllCentsIncludesPersonal(t *testing.T) {
	// Shared: 10000, Personal u1: 3000, Personal u2: 2000
	// TotalAllCents should include shared splits + personal
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 10000, IsShared: true, PaidBy: "u1"},
				{ID: "e2", AmountCents: 3000, IsShared: false, PaidBy: "u1", AssignedTo: "u1"},
				{ID: "e3", AmountCents: 2000, IsShared: false, PaidBy: "u2", AssignedTo: "u2"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// TotalAllCents = sum of all AmountDueCents
	// Alice: 6250 shared + 3000 personal = 9250
	// Bob: 3750 shared + 2000 personal = 5750
	// Total = 15000
	expectedTotal := int64(15000)
	if resp.TotalAllCents != expectedTotal {
		t.Errorf("expected TotalAllCents %d, got %d", expectedTotal, resp.TotalAllCents)
	}

	var computed int64
	for _, item := range resp.Items {
		computed += item.AmountDueCents
	}
	if computed != resp.TotalAllCents {
		t.Errorf("sum of AmountDueCents (%d) != TotalAllCents (%d)", computed, resp.TotalAllCents)
	}
}

func TestSummaryService_Generate_MultipleSharedExpensesAndBills(t *testing.T) {
	// Multiple shared expenses + multiple shared bills all sum together
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 3000, IsShared: true, PaidBy: "u1"},
				{ID: "e2", AmountCents: 4000, IsShared: true, PaidBy: "u2"},
				{ID: "e3", AmountCents: 1500, IsShared: true, PaidBy: "u1"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", AmountCents: 2000, IsShared: true, IsActive: true},
				{ID: "b2", AmountCents: 1500, IsShared: true, IsActive: true},
			}, nil
		},
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	resp, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Total shared: 3000 + 4000 + 1500 + 2000 + 1500 = 12000
	if resp.TotalSharedCents != 12000 {
		t.Errorf("expected TotalSharedCents 12000, got %d", resp.TotalSharedCents)
	}

	// Alice: 62.5% of 12000 = 7500
	if resp.Items[0].TotalSharedCents != 7500 {
		t.Errorf("Alice shared: expected 7500, got %d", resp.Items[0].TotalSharedCents)
	}
	// Bob: 37.5% of 12000 = 4500
	if resp.Items[1].TotalSharedCents != 4500 {
		t.Errorf("Bob shared: expected 4500, got %d", resp.Items[1].TotalSharedCents)
	}
}

func TestSummaryService_Generate_ExpenseRepoError(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn: memberOK(),
		ListMembersFn: func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
			return []domain.HouseholdMember{{UserID: "u1", SalaryCents: 100000}}, nil
		},
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return nil, errors.New("db connection error")
		},
	}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	_, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err == nil {
		t.Fatal("expected error from expense repo, got nil")
	}
}

func TestSummaryService_Generate_BillRepoError(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn: memberOK(),
		ListMembersFn: func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
			return []domain.HouseholdMember{{UserID: "u1", SalaryCents: 100000}}, nil
		},
	}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string) ([]domain.FixedBill, error) {
			return nil, errors.New("db error")
		},
	}
	sumRepo := &mock.SummaryRepository{UpsertFn: summaryUpsertOK()}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	_, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err == nil {
		t.Fatal("expected error from bill repo, got nil")
	}
}

func TestSummaryService_Generate_MemberListError(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn: memberOK(),
		ListMembersFn: func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
			return nil, errors.New("db error")
		},
	}

	svc := NewSummaryService(nil, hhRepo, nil, nil)
	_, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err == nil {
		t.Fatal("expected error from member list, got nil")
	}
}

func TestSummaryService_Generate_UpsertError(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
	}
	expRepo := &mock.ExpenseRepository{ListByHouseholdFn: noExpenses()}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}
	sumRepo := &mock.SummaryRepository{
		UpsertFn: func(ctx context.Context, s *domain.MonthlySummary) error {
			return errors.New("upsert failed")
		},
	}

	svc := makeSummaryService(hhRepo, expRepo, billRepo, sumRepo)
	_, err := svc.Generate(context.Background(), "hh-1", 2024, 1, "u1")
	if err == nil {
		t.Fatal("expected error from upsert, got nil")
	}
}

func TestSummaryService_Dashboard_NoSalaryGraceful(t *testing.T) {
	// Dashboard should succeed even when no members have salary
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn: memberOK(),
		ListMembersFn: func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
			return []domain.HouseholdMember{
				{UserID: "u1", SalaryCents: 0},
			}, nil
		},
		FindByIDFn: func(ctx context.Context, id string) (*domain.Household, error) {
			return &domain.Household{ID: id, Name: "Casa"}, nil
		},
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 5000, IsShared: true, PaidBy: "u1"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{ListByHouseholdFn: noBills()}

	svc := NewSummaryService(nil, hhRepo, expRepo, billRepo)
	dash, err := svc.GetDashboard(context.Background(), "hh-1", "u1")
	if err != nil {
		t.Fatalf("Dashboard should succeed with no salary, got: %v", err)
	}
	if dash.TotalExpenses != 5000 {
		t.Errorf("expected total expenses 5000, got %d", dash.TotalExpenses)
	}
	// MemberBreakdown should be nil/empty since calculate returns ErrNoMembersWithSalary
	if dash.MemberBreakdown != nil && len(dash.MemberBreakdown) > 0 {
		t.Errorf("expected no member breakdown with zero salaries, got %d items", len(dash.MemberBreakdown))
	}
}

// --- Dashboard tests ---

func TestSummaryService_Dashboard_Success(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{
		GetMemberFn:   memberOK(),
		ListMembersFn: twoMembers(),
		FindByIDFn: func(ctx context.Context, id string) (*domain.Household, error) {
			return &domain.Household{ID: id, Name: "Casa"}, nil
		},
	}
	expRepo := &mock.ExpenseRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string, f domain.ExpenseFilter) ([]domain.Expense, error) {
			return []domain.Expense{
				{ID: "e1", AmountCents: 5000, IsShared: true, PaidBy: "u1"},
				{ID: "e2", AmountCents: 2000, IsShared: false, PaidBy: "u2", AssignedTo: "u2"},
			}, nil
		},
	}
	billRepo := &mock.FixedBillRepository{
		ListByHouseholdFn: func(ctx context.Context, hID string) ([]domain.FixedBill, error) {
			return []domain.FixedBill{
				{ID: "b1", AmountCents: 3000, IsShared: true, IsActive: true},
				{ID: "b2", AmountCents: 1000, IsShared: false, AssignedTo: "u1", IsActive: true},
				{ID: "b3", AmountCents: 9000, IsShared: true, IsActive: false}, // inactive, excluded from counts
			}, nil
		},
	}

	svc := NewSummaryService(nil, hhRepo, expRepo, billRepo)
	dash, err := svc.GetDashboard(context.Background(), "hh-1", "u1")
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}

	if dash.HouseholdName != "Casa" {
		t.Errorf("expected Casa, got %s", dash.HouseholdName)
	}
	if dash.TotalExpenses != 7000 {
		t.Errorf("expected total expenses 7000, got %d", dash.TotalExpenses)
	}
	if dash.TotalFixedBills != 4000 {
		t.Errorf("expected total fixed bills 4000 (active only), got %d", dash.TotalFixedBills)
	}
	// Shared: 5000 (expense) + 3000 (bill) = 8000
	if dash.TotalShared != 8000 {
		t.Errorf("expected total shared 8000, got %d", dash.TotalShared)
	}
	// Personal: 2000 (expense) + 1000 (bill) = 3000
	if dash.TotalPersonal != 3000 {
		t.Errorf("expected total personal 3000, got %d", dash.TotalPersonal)
	}
	if dash.ExpenseCount != 2 {
		t.Errorf("expected 2 expenses, got %d", dash.ExpenseCount)
	}
	if dash.FixedBillCount != 2 {
		t.Errorf("expected 2 active bills, got %d", dash.FixedBillCount)
	}
	if len(dash.MemberBreakdown) != 2 {
		t.Errorf("expected 2 members, got %d", len(dash.MemberBreakdown))
	}
}

func TestSummaryService_Dashboard_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewSummaryService(nil, hhRepo, nil, nil)

	_, err := svc.GetDashboard(context.Background(), "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
