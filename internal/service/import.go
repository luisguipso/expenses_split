package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/parser"
)

type importService struct {
	registry      *parser.Registry
	householdRepo domain.HouseholdRepository
	categoryRepo  domain.CategoryRepository
	expenseRepo   domain.ExpenseRepository
}

func NewImportService(
	registry *parser.Registry,
	householdRepo domain.HouseholdRepository,
	categoryRepo domain.CategoryRepository,
	expenseRepo domain.ExpenseRepository,
) domain.ImportService {
	return &importService{
		registry:      registry,
		householdRepo: householdRepo,
		categoryRepo:  categoryRepo,
		expenseRepo:   expenseRepo,
	}
}

func (s *importService) ParseBill(ctx context.Context, filename string, content []byte, householdID, userID string) (*domain.ImportPreviewResponse, error) {
	slog.Info("service: parsing bill",
		"household_id", householdID,
		"user_id", userID,
		"filename", filename,
		"content_size", len(content),
	)

	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	p, err := s.registry.FindParser(content)
	if err != nil {
		slog.Warn("service: no parser matched uploaded file",
			"error", err,
			"household_id", householdID,
			"filename", filename,
		)
		return nil, err
	}

	bill, err := p.Parse(ctx, bytes.NewReader(content))
	if err != nil {
		slog.Error("service: failed to parse bill",
			"error", err,
			"household_id", householdID,
			"filename", filename,
		)
		return nil, fmt.Errorf("parse bill: %w", err)
	}

	slog.Info("service: bill parsed successfully",
		"household_id", householdID,
		"provider", bill.Provider,
		"items_count", len(bill.Items),
	)

	categories, err := s.categoryRepo.ListByHousehold(ctx, householdID)
	if err != nil {
		slog.Error("service: failed to list categories for import",
			"error", err,
			"household_id", householdID,
		)
		return nil, fmt.Errorf("list categories: %w", err)
	}

	items := make([]domain.ImportPreviewItem, len(bill.Items))
	for i, pe := range bill.Items {
		items[i] = domain.ImportPreviewItem{
			Description:       pe.Description,
			AmountCents:       pe.AmountCents,
			Date:              pe.Date,
			SuggestedCategory: suggestCategory(pe.Description, categories),
		}
	}

	return &domain.ImportPreviewResponse{
		Provider: bill.Provider,
		Items:    items,
	}, nil
}

func (s *importService) ConfirmImport(ctx context.Context, input domain.ImportConfirmInput, householdID, userID string) ([]domain.Expense, error) {
	slog.Info("service: confirming import",
		"household_id", householdID,
		"user_id", userID,
		"items_count", len(input.Items),
	)

	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	if len(input.Items) == 0 {
		slog.Warn("service: confirm import called with empty items",
			"household_id", householdID,
			"user_id", userID,
		)
		return nil, fmt.Errorf("confirm import: %w", errors.New("items cannot be empty"))
	}

	expenses := make([]*domain.Expense, len(input.Items))
	for i, item := range input.Items {
		if item.Description == "" {
			return nil, fmt.Errorf("confirm import: item %d: description is required", i)
		}
		if len(item.Description) > 255 {
			return nil, fmt.Errorf("confirm import: item %d: description too long", i)
		}
		//if item.AmountCents <= 0 {
		//	return nil, fmt.Errorf("confirm import: item %d: amount_cents must be positive", i)
		//}
		if item.ExpenseDate == "" {
			return nil, fmt.Errorf("confirm import: item %d: expense_date is required", i)
		}

		paidBy := item.PaidBy
		if paidBy == "" {
			paidBy = userID
		}

		expenses[i] = &domain.Expense{
			HouseholdID: householdID,
			CategoryID:  item.CategoryID,
			Description: item.Description,
			AmountCents: item.AmountCents,
			ExpenseDate: item.ExpenseDate,
			IsShared:    item.IsShared,
			PaidBy:      paidBy,
			AssignedTo:  item.AssignedTo,
		}
	}

	if err := s.expenseRepo.CreateBatch(ctx, expenses); err != nil {
		slog.Error("service: failed to create expense batch",
			"error", err,
			"household_id", householdID,
			"items_count", len(expenses),
		)
		return nil, fmt.Errorf("create batch: %w", err)
	}

	slog.Info("service: import confirmed successfully",
		"household_id", householdID,
		"user_id", userID,
		"expenses_created", len(expenses),
	)

	result := make([]domain.Expense, len(expenses))
	for i, e := range expenses {
		result[i] = *e
	}
	return result, nil
}

func (s *importService) checkMembership(ctx context.Context, householdID, userID string) error {
	_, err := s.householdRepo.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			slog.Warn("service: import denied, user is not a member",
				"household_id", householdID,
				"user_id", userID,
			)
			return domain.ErrForbidden
		}
		slog.Error("service: failed to check membership for import",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return err
	}
	return nil
}

func suggestCategory(description string, categories []domain.Category) string {
	descLower := strings.ToLower(description)
	for _, cat := range categories {
		if strings.Contains(descLower, strings.ToLower(cat.Name)) {
			return cat.ID
		}
	}
	return ""
}
