package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/parser"
)

type importService struct {
	registry     *parser.Registry
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
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	p, err := s.registry.FindParser(content)
	if err != nil {
		return nil, err
	}

	bill, err := p.Parse(ctx, bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse bill: %w", err)
	}

	categories, err := s.categoryRepo.ListByHousehold(ctx, householdID)
	if err != nil {
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
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	if len(input.Items) == 0 {
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
		if item.AmountCents <= 0 {
			return nil, fmt.Errorf("confirm import: item %d: amount_cents must be positive", i)
		}
		if item.ExpenseDate == "" {
			return nil, fmt.Errorf("confirm import: item %d: expense_date is required", i)
		}

		expenses[i] = &domain.Expense{
			HouseholdID: householdID,
			CategoryID:  item.CategoryID,
			Description: item.Description,
			AmountCents: item.AmountCents,
			ExpenseDate: item.ExpenseDate,
			IsShared:    item.IsShared,
			PaidBy:      userID,
			AssignedTo:  item.AssignedTo,
		}
	}

	if err := s.expenseRepo.CreateBatch(ctx, expenses); err != nil {
		return nil, fmt.Errorf("create batch: %w", err)
	}

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
			return domain.ErrForbidden
		}
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
