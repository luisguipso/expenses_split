package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
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
		Settlements:      minimizeTransfers(breakdown),
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
	paidByUser := make(map[string]int64)

	for _, e := range expenses {
		paidByUser[e.PaidBy] += e.AmountCents
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
		paidByUser[b.PaidBy] += b.AmountCents
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
		totalPaid := paidByUser[m.UserID]
		amountDue := sharedDue + personalDue

		breakdown[i] = domain.SummaryItemResponse{
			UserID:             m.UserID,
			UserName:           m.UserName,
			SalaryCents:        m.SalaryCents,
			Proportion:         proportion,
			TotalSharedCents:   sharedDue,
			TotalPersonalCents: personalDue,
			AmountDueCents:     amountDue,
			TotalPaidCents:     totalPaid,
			BalanceCents:       totalPaid - amountDue,
		}
	}

	return breakdown, totalShared, nil
}

// minimizeTransfers computes the minimum number of transfers to settle all balances.
// Uses a greedy algorithm: repeatedly match the largest creditor with the largest debtor.
func minimizeTransfers(items []domain.SummaryItemResponse) []domain.SettlementTransfer {
	type entry struct {
		userID   string
		userName string
		balance  int64
	}

	var creditors, debtors []entry
	for _, item := range items {
		if item.BalanceCents > 0 {
			creditors = append(creditors, entry{item.UserID, item.UserName, item.BalanceCents})
		} else if item.BalanceCents < 0 {
			debtors = append(debtors, entry{item.UserID, item.UserName, -item.BalanceCents})
		}
	}

	// Sort descending by amount
	sort.Slice(creditors, func(i, j int) bool { return creditors[i].balance > creditors[j].balance })
	sort.Slice(debtors, func(i, j int) bool { return debtors[i].balance > debtors[j].balance })

	var transfers []domain.SettlementTransfer
	ci, di := 0, 0

	for ci < len(creditors) && di < len(debtors) {
		amount := creditors[ci].balance
		if debtors[di].balance < amount {
			amount = debtors[di].balance
		}

		transfers = append(transfers, domain.SettlementTransfer{
			FromUserID:   debtors[di].userID,
			FromUserName: debtors[di].userName,
			ToUserID:     creditors[ci].userID,
			ToUserName:   creditors[ci].userName,
			AmountCents:  amount,
		})

		creditors[ci].balance -= amount
		debtors[di].balance -= amount

		if creditors[ci].balance == 0 {
			ci++
		}
		if debtors[di].balance == 0 {
			di++
		}
	}

	return transfers
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

func (s *summaryService) GetUserDetail(ctx context.Context, householdID string, year, month int, targetUserID, requestingUserID string) (*domain.SummaryDetailResponse, error) {
	if err := s.checkMembership(ctx, householdID, requestingUserID); err != nil {
		return nil, err
	}

	members, err := s.householdRepo.ListMembers(ctx, householdID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	var totalSalary int64
	var targetMember *domain.HouseholdMember
	for i, m := range members {
		totalSalary += m.SalaryCents
		if m.UserID == targetUserID {
			targetMember = &members[i]
		}
	}
	if totalSalary == 0 {
		return nil, domain.ErrNoMembersWithSalary
	}
	if targetMember == nil {
		return nil, domain.ErrNotMember
	}

	proportion := float64(targetMember.SalaryCents) / float64(totalSalary)

	// Build name lookup for paid_by
	nameByID := make(map[string]string, len(members))
	for _, m := range members {
		nameByID[m.UserID] = m.UserName
	}

	expenses, err := s.expenseRepo.ListByHousehold(ctx, householdID, domain.ExpenseFilter{
		Year: year, Month: month,
	})
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}

	allBills, err := s.fixedBillRepo.ListByHousehold(ctx, householdID)
	if err != nil {
		return nil, fmt.Errorf("list fixed bills: %w", err)
	}

	var items []domain.SummaryDetailItem
	var totalShared, totalPersonal, totalPaid int64

	// Fixed bills — all shared bills, plus personal bills assigned to/paid by target user
	for _, b := range allBills {
		if !b.IsActive {
			continue
		}
		
		// Track what this user actually paid
		if b.PaidBy == targetUserID {
			totalPaid += b.AmountCents
		}
		
		if b.IsShared {
			shareCents := int64(math.Round(float64(b.AmountCents) * proportion))
			items = append(items, domain.SummaryDetailItem{
				Description:    b.Description,
				Type:           "fixed_bill",
				CategoryName:   b.CategoryName,
				TotalCents:     b.AmountCents,
				UserShareCents: shareCents,
				Proportion:     proportion,
				IsShared:       true,
				PaidByName:     nameByID[b.PaidBy],
			})
			totalShared += shareCents
		} else if b.AssignedTo == targetUserID || (b.AssignedTo == "" && b.PaidBy == targetUserID) {
			// Only include personal bills if assigned to or paid by this user
			items = append(items, domain.SummaryDetailItem{
				Description:    b.Description,
				Type:           "fixed_bill",
				CategoryName:   b.CategoryName,
				TotalCents:     b.AmountCents,
				UserShareCents: b.AmountCents,
				Proportion:     1.0,
				IsShared:       false,
				PaidByName:     nameByID[b.PaidBy],
			})
			totalPersonal += b.AmountCents
		}
	}

	// Expenses — all shared expenses, plus personal expenses assigned to/paid by target user
	for _, e := range expenses {
		// Track what this user actually paid
		if e.PaidBy == targetUserID {
			totalPaid += e.AmountCents
		}
		
		if e.IsShared {
			shareCents := int64(math.Round(float64(e.AmountCents) * proportion))
			items = append(items, domain.SummaryDetailItem{
				Description:    e.Description,
				Type:           "expense",
				CategoryName:   e.CategoryName,
				TotalCents:     e.AmountCents,
				UserShareCents: shareCents,
				Proportion:     proportion,
				IsShared:       true,
				PaidByName:     nameByID[e.PaidBy],
			})
			totalShared += shareCents
		} else {
			// Only include personal expenses if assigned to this user (or paid by if no assignment)
			assignee := e.AssignedTo
			if assignee == "" {
				assignee = e.PaidBy
			}
			if assignee == targetUserID {
				items = append(items, domain.SummaryDetailItem{
					Description:    e.Description,
					Type:           "expense",
					CategoryName:   e.CategoryName,
					TotalCents:     e.AmountCents,
					UserShareCents: e.AmountCents,
					Proportion:     1.0,
					IsShared:       false,
					PaidByName:     nameByID[e.PaidBy],
				})
				totalPersonal += e.AmountCents
			}
		}
	}

	amountDue := totalShared + totalPersonal

	return &domain.SummaryDetailResponse{
		UserID:             targetUserID,
		UserName:           targetMember.UserName,
		Items:              items,
		TotalSharedCents:   totalShared,
		TotalPersonalCents: totalPersonal,
		AmountDueCents:     amountDue,
		TotalPaidCents:     totalPaid,
		BalanceCents:       totalPaid - amountDue,
	}, nil
}
