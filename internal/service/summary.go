package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/lguilherme/contas/internal/domain"
)

type summaryService struct {
	summaryRepo   domain.SummaryRepository
	householdRepo domain.HouseholdRepository
	expenseRepo   domain.ExpenseRepository
	fixedBillRepo domain.FixedBillRepository
}

func NewSummaryService(
	summaryRepo domain.SummaryRepository,
	householdRepo domain.HouseholdRepository,
	expenseRepo domain.ExpenseRepository,
	fixedBillRepo domain.FixedBillRepository,
) domain.SummaryService {
	return &summaryService{
		summaryRepo:   summaryRepo,
		householdRepo: householdRepo,
		expenseRepo:   expenseRepo,
		fixedBillRepo: fixedBillRepo,
	}
}

func (s *summaryService) Generate(ctx context.Context, householdID string, year, month int, userID string) (*domain.SummaryResponse, error) {
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	breakdown, totalShared, err := s.calculate(ctx, householdID, year, month)
	if err != nil {
		return nil, err
	}

	// Persist the summary
	summary := &domain.MonthlySummary{
		HouseholdID: householdID,
		Year:        year,
		Month:       month,
	}
	for _, b := range breakdown {
		summary.Items = append(summary.Items, domain.MonthlySummaryItem{
			UserID:             b.UserID,
			TotalSharedCents:   b.TotalSharedCents,
			TotalPersonalCents: b.TotalPersonalCents,
			AmountDueCents:     b.AmountDueCents,
			TotalPaidCents:     b.TotalPaidCents,
			BalanceCents:       b.BalanceCents,
		})
	}
	if err := s.summaryRepo.Upsert(ctx, summary); err != nil {
		return nil, fmt.Errorf("persist summary: %w", err)
	}

	var totalAll int64
	for _, b := range breakdown {
		totalAll += b.AmountDueCents
	}

	return &domain.SummaryResponse{
		ID:               summary.ID,
		HouseholdID:      householdID,
		Year:             year,
		Month:            month,
		TotalSharedCents: totalShared,
		TotalAllCents:    totalAll,
		GeneratedAt:      summary.GeneratedAt.Format(time.RFC3339),
		Items:            breakdown,
	}, nil
}

func (s *summaryService) GetDashboard(ctx context.Context, householdID, userID string) (*domain.DashboardResponse, error) {
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	now := time.Now()
	year, month := now.Year(), int(now.Month())

	household, err := s.householdRepo.FindByID(ctx, householdID)
	if err != nil {
		return nil, err
	}

	// Get expenses for this month
	expenses, err := s.expenseRepo.ListByHousehold(ctx, householdID, domain.ExpenseFilter{
		Year: year, Month: month,
	})
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}

	// Get active fixed bills
	allBills, err := s.fixedBillRepo.ListByHousehold(ctx, householdID)
	if err != nil {
		return nil, fmt.Errorf("list fixed bills: %w", err)
	}
	var activeBills []domain.FixedBill
	for _, b := range allBills {
		if b.IsActive {
			activeBills = append(activeBills, b)
		}
	}

	var totalExpenses, totalFixedBills, totalShared, totalPersonal int64

	for _, e := range expenses {
		totalExpenses += e.AmountCents
		if e.IsShared {
			totalShared += e.AmountCents
		} else {
			totalPersonal += e.AmountCents
		}
	}
	for _, b := range activeBills {
		totalFixedBills += b.AmountCents
		if b.IsShared {
			totalShared += b.AmountCents
		} else {
			totalPersonal += b.AmountCents
		}
	}

	breakdown, _, err := s.calculate(ctx, householdID, year, month)
	if err != nil && !errors.Is(err, domain.ErrNoMembersWithSalary) {
		return nil, err
	}

	return &domain.DashboardResponse{
		HouseholdName:   household.Name,
		Year:            year,
		Month:           month,
		TotalExpenses:   totalExpenses,
		TotalFixedBills: totalFixedBills,
		TotalShared:     totalShared,
		TotalPersonal:   totalPersonal,
		ExpenseCount:    len(expenses),
		FixedBillCount:  len(activeBills),
		MemberBreakdown: breakdown,
	}, nil
}

// calculate computes the proportional split for a given month.
// Returns per-member breakdown and total shared amount.
func (s *summaryService) calculate(ctx context.Context, householdID string, year, month int) ([]domain.SummaryItemResponse, int64, error) {
	// 1. Get members with salaries
	members, err := s.householdRepo.ListMembers(ctx, householdID)
	if err != nil {
		return nil, 0, fmt.Errorf("list members: %w", err)
	}

	var totalSalary int64
	for _, m := range members {
		totalSalary += m.SalaryCents
	}
	if totalSalary == 0 {
		return nil, 0, domain.ErrNoMembersWithSalary
	}

	// 2. Get expenses for this month
	expenses, err := s.expenseRepo.ListByHousehold(ctx, householdID, domain.ExpenseFilter{
		Year: year, Month: month,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list expenses: %w", err)
	}

	// 3. Get active fixed bills
	allBills, err := s.fixedBillRepo.ListByHousehold(ctx, householdID)
	if err != nil {
		return nil, 0, fmt.Errorf("list fixed bills: %w", err)
	}

	// 4. Calculate totals
	var totalShared int64
	personalByUser := make(map[string]int64)

	for _, e := range expenses {
		if e.IsShared {
			totalShared += e.AmountCents
		} else {
			assignee := e.AssignedTo
			if assignee == "" {
				assignee = e.PaidBy
			}
			personalByUser[assignee] += e.AmountCents
		}
	}

	for _, b := range allBills {
		if !b.IsActive {
			continue
		}
		if b.IsShared {
			totalShared += b.AmountCents
		} else {
			if b.AssignedTo != "" {
				personalByUser[b.AssignedTo] += b.AmountCents
			}
		}
	}

	// 5. Calculate proportional split
	breakdown := make([]domain.SummaryItemResponse, len(members))
	var allocatedShared int64

	for i, m := range members {
		proportion := float64(m.SalaryCents) / float64(totalSalary)
		sharedDue := int64(math.Round(float64(totalShared) * proportion))

		// For the last member, assign remainder to avoid rounding drift
		if i == len(members)-1 {
			sharedDue = totalShared - allocatedShared
		}
		allocatedShared += sharedDue

		personalDue := personalByUser[m.UserID]

		breakdown[i] = domain.SummaryItemResponse{
			UserID:             m.UserID,
			UserName:           m.UserName,
			SalaryCents:        m.SalaryCents,
			Proportion:         proportion,
			TotalSharedCents:   sharedDue,
			TotalPersonalCents: personalDue,
			AmountDueCents:     sharedDue + personalDue,
		}
	}

	return breakdown, totalShared, nil
}

func (s *summaryService) checkMembership(ctx context.Context, householdID, userID string) error {
	_, err := s.householdRepo.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}
