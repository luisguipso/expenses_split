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
	snapshotRepo  domain.FixedBillSnapshotRepository
}

func NewSummaryService(
	summaryRepo domain.SummaryRepository,
	householdRepo domain.HouseholdRepository,
	expenseRepo domain.ExpenseRepository,
	fixedBillRepo domain.FixedBillRepository,
	snapshotRepo domain.FixedBillSnapshotRepository,
) domain.SummaryService {
	return &summaryService{
		summaryRepo:   summaryRepo,
		householdRepo: householdRepo,
		expenseRepo:   expenseRepo,
		fixedBillRepo: fixedBillRepo,
		snapshotRepo:  snapshotRepo,
	}
}

func (s *summaryService) Generate(ctx context.Context, householdID string, year, month int, userID string) (*domain.SummaryResponse, error) {
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	breakdown, totalShared, bills, warnings, err := s.calculate(ctx, householdID, year, month)
	if err != nil {
		return nil, err
	}

	fixedBillResponses := make([]domain.FixedBillSnapshotResponse, len(bills))
	for i, b := range bills {
		fixedBillResponses[i] = domain.FixedBillSnapshotResponse{
			ID:           b.SnapshotID,
			FixedBillID:  b.FixedBillID,
			CategoryID:   b.CategoryID,
			CategoryName: b.CategoryName,
			Description:  b.Description,
			AmountCents:  b.AmountCents,
			DueDay:       b.DueDay,
			IsShared:     b.IsShared,
			PaidBy:       b.PaidBy,
			PaidByName:   b.PaidByName,
			AssignedTo:   b.AssignedTo,
			IsFrozen:     b.IsFrozen,
		}
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
		FixedBills:       fixedBillResponses,
		Warnings:         warnings,
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

	// Get resolved fixed bills (snapshot or live)
	resolvedBills, err := s.resolveFixedBills(ctx, householdID, year, month)
	if err != nil {
		return nil, err
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
	for _, b := range resolvedBills {
		totalFixedBills += b.AmountCents
		if b.IsShared {
			totalShared += b.AmountCents
		} else {
			totalPersonal += b.AmountCents
		}
	}

	breakdown, _, _, _, err := s.calculate(ctx, householdID, year, month)
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
		FixedBillCount:  len(resolvedBills),
		MemberBreakdown: breakdown,
	}, nil
}

// resolvedBill represents a fixed bill's effective values for a given month,
// either from the live bill or from a frozen snapshot.
type resolvedBill struct {
	FixedBillID  string
	CategoryID   string
	CategoryName string
	Description  string
	AmountCents  int64
	DueDay       int
	IsShared     bool
	PaidBy       string
	PaidByName   string
	AssignedTo   string
	IsFrozen     bool
	SnapshotID   string
}

// dueDayPassed checks if a bill's due_day has passed for the given year/month.
// If due_day > days in month, it's treated as the last day of the month.
func dueDayPassed(year, month, dueDay int, now time.Time) bool {
	lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, now.Location()).Day()
	effectiveDueDay := dueDay
	if effectiveDueDay > lastDay {
		effectiveDueDay = lastDay
	}
	dueDate := time.Date(year, time.Month(month), effectiveDueDay, 23, 59, 59, 0, now.Location())
	return now.After(dueDate)
}

// resolveFixedBills returns the effective fixed bill values for a given month.
// Bills whose due_day has passed are frozen (snapshotted); others use live values.
func (s *summaryService) resolveFixedBills(ctx context.Context, householdID string, year, month int) ([]resolvedBill, error) {
	now := time.Now()

	allBills, err := s.fixedBillRepo.ListByHousehold(ctx, householdID)
	if err != nil {
		return nil, fmt.Errorf("list fixed bills: %w", err)
	}

	existingSnapshots, err := s.snapshotRepo.FindByMonth(ctx, householdID, year, month)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}

	snapshotByBillID := make(map[string]domain.FixedBillSnapshot, len(existingSnapshots))
	for _, snap := range existingSnapshots {
		snapshotByBillID[snap.FixedBillID] = snap
	}

	// Build category/user name lookup from the original bills
	categoryNameByID := make(map[string]string)
	paidByNameByID := make(map[string]string)
	for _, b := range allBills {
		if b.CategoryID != "" && b.CategoryName != "" {
			categoryNameByID[b.CategoryID] = b.CategoryName
		}
		if b.PaidBy != "" && b.PaidByName != "" {
			paidByNameByID[b.PaidBy] = b.PaidByName
		}
	}

	var resolved []resolvedBill
	for _, b := range allBills {
		if !b.IsActive {
			continue
		}

		if snap, ok := snapshotByBillID[b.ID]; ok {
			// Snapshot already exists — use it
			resolved = append(resolved, resolvedBill{
				FixedBillID:  b.ID,
				CategoryID:   snap.CategoryID,
				CategoryName: categoryNameByID[snap.CategoryID],
				Description:  snap.Description,
				AmountCents:  snap.AmountCents,
				DueDay:       snap.DueDay,
				IsShared:     snap.IsShared,
				PaidBy:       snap.PaidBy,
				PaidByName:   paidByNameByID[snap.PaidBy],
				AssignedTo:   snap.AssignedTo,
				IsFrozen:     true,
				SnapshotID:   snap.ID,
			})
			continue
		}

		if dueDayPassed(year, month, b.DueDay, now) {
			// Due day passed, create snapshot
			snap := &domain.FixedBillSnapshot{
				FixedBillID: b.ID,
				HouseholdID: householdID,
				Year:        year,
				Month:       month,
				CategoryID:  b.CategoryID,
				Description: b.Description,
				AmountCents: b.AmountCents,
				DueDay:      b.DueDay,
				IsShared:    b.IsShared,
				PaidBy:      b.PaidBy,
				AssignedTo:  b.AssignedTo,
			}
			if err := s.snapshotRepo.Create(ctx, snap); err != nil {
				return nil, fmt.Errorf("create snapshot for bill %s: %w", b.ID, err)
			}
			resolved = append(resolved, resolvedBill{
				FixedBillID:  b.ID,
				CategoryID:   snap.CategoryID,
				CategoryName: categoryNameByID[snap.CategoryID],
				Description:  snap.Description,
				AmountCents:  snap.AmountCents,
				DueDay:       snap.DueDay,
				IsShared:     snap.IsShared,
				PaidBy:       snap.PaidBy,
				PaidByName:   paidByNameByID[snap.PaidBy],
				AssignedTo:   snap.AssignedTo,
				IsFrozen:     true,
				SnapshotID:   snap.ID,
			})
		} else {
			// Due day not yet passed — use live values
			resolved = append(resolved, resolvedBill{
				FixedBillID:  b.ID,
				CategoryID:   b.CategoryID,
				CategoryName: b.CategoryName,
				Description:  b.Description,
				AmountCents:  b.AmountCents,
				DueDay:       b.DueDay,
				IsShared:     b.IsShared,
				PaidBy:       b.PaidBy,
				PaidByName:   b.PaidByName,
				AssignedTo:   b.AssignedTo,
				IsFrozen:     false,
			})
		}
	}

	return resolved, nil
}

// calculate computes the proportional split for a given month.
// Returns per-member breakdown, total shared amount, resolved bills, and warnings.
func (s *summaryService) calculate(ctx context.Context, householdID string, year, month int) ([]domain.SummaryItemResponse, int64, []resolvedBill, []string, error) {
	// 1. Get household to check split mode
	household, err := s.householdRepo.FindByID(ctx, householdID)
	if err != nil {
		return nil, 0, nil, nil, fmt.Errorf("find household: %w", err)
	}

	// 2. Get members with salaries/percentages
	members, err := s.householdRepo.ListMembers(ctx, householdID)
	if err != nil {
		return nil, 0, nil, nil, fmt.Errorf("list members: %w", err)
	}

	var warnings []string

	// 3. Determine proportions based on split mode
	proportions := make([]float64, len(members))
	if household.SplitMode == "percentage" {
		var totalPercentage int
		for _, m := range members {
			totalPercentage += m.SplitPercentage
		}
		if totalPercentage == 0 {
			return nil, 0, nil, nil, domain.ErrNoMembersWithSalary
		}
		if totalPercentage != 10000 {
			warnings = append(warnings, fmt.Sprintf("A soma dos percentuais dos moradores é %.2f%%. O total deve ser 100%%.", float64(totalPercentage)/100.0))
		}
		for i, m := range members {
			proportions[i] = float64(m.SplitPercentage) / 10000.0
		}
	} else {
		var totalSalary int64
		for _, m := range members {
			totalSalary += m.SalaryCents
		}
		if totalSalary == 0 {
			return nil, 0, nil, nil, domain.ErrNoMembersWithSalary
		}
		for i, m := range members {
			proportions[i] = float64(m.SalaryCents) / float64(totalSalary)
		}
	}

	// 4. Get expenses for this month
	expenses, err := s.expenseRepo.ListByHousehold(ctx, householdID, domain.ExpenseFilter{
		Year: year, Month: month,
	})
	if err != nil {
		return nil, 0, nil, nil, fmt.Errorf("list expenses: %w", err)
	}

	// 5. Resolve fixed bills (snapshot or live)
	bills, err := s.resolveFixedBills(ctx, householdID, year, month)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	// 6. Calculate totals
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

	for _, b := range bills {
		paidByUser[b.PaidBy] += b.AmountCents
		if b.IsShared {
			totalShared += b.AmountCents
		} else {
			if b.AssignedTo != "" {
				personalByUser[b.AssignedTo] += b.AmountCents
			}
		}
	}

	// 7. Calculate proportional split
	breakdown := make([]domain.SummaryItemResponse, len(members))
	var allocatedShared int64

	for i, m := range members {
		proportion := proportions[i]
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

	return breakdown, totalShared, bills, warnings, nil
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

	household, err := s.householdRepo.FindByID(ctx, householdID)
	if err != nil {
		return nil, fmt.Errorf("find household: %w", err)
	}

	var targetMember *domain.HouseholdMember
	for i, m := range members {
		if m.UserID == targetUserID {
			targetMember = &members[i]
		}
	}
	if targetMember == nil {
		return nil, domain.ErrNotMember
	}

	var proportion float64
	if household.SplitMode == "percentage" {
		var totalPercentage int
		for _, m := range members {
			totalPercentage += m.SplitPercentage
		}
		if totalPercentage == 0 {
			return nil, domain.ErrNoMembersWithSalary
		}
		proportion = float64(targetMember.SplitPercentage) / 10000.0
	} else {
		var totalSalary int64
		for _, m := range members {
			totalSalary += m.SalaryCents
		}
		if totalSalary == 0 {
			return nil, domain.ErrNoMembersWithSalary
		}
		proportion = float64(targetMember.SalaryCents) / float64(totalSalary)
	}

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

	allBills, err := s.resolveFixedBills(ctx, householdID, year, month)
	if err != nil {
		return nil, err
	}

	var items []domain.SummaryDetailItem
	var totalShared, totalPersonal, totalPaid int64

	// Fixed bills — all shared bills, plus personal bills assigned to/paid by target user
	for _, b := range allBills {
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
