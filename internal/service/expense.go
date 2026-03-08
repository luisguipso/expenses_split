package service

import (
	"context"
	"errors"

	"github.com/lguilherme/contas/internal/domain"
)

type expenseService struct {
	repo      domain.ExpenseRepository
	household domain.HouseholdRepository
}

func NewExpenseService(repo domain.ExpenseRepository, household domain.HouseholdRepository) domain.ExpenseService {
	return &expenseService{repo: repo, household: household}
}

func (s *expenseService) Create(ctx context.Context, input domain.CreateExpenseInput, householdID, userID string) (*domain.Expense, error) {
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	e := &domain.Expense{
		HouseholdID: householdID,
		CategoryID:  input.CategoryID,
		Description: input.Description,
		AmountCents: input.AmountCents,
		ExpenseDate: input.ExpenseDate,
		IsShared:    input.IsShared,
		AssignedTo:  input.AssignedTo,
	}
	e.SetDefaults(userID)
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *expenseService) List(ctx context.Context, householdID, userID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByHousehold(ctx, householdID, filter)
}

func (s *expenseService) Update(ctx context.Context, id string, input domain.UpdateExpenseInput, userID string) (*domain.Expense, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkMembership(ctx, e.HouseholdID, userID); err != nil {
		return nil, err
	}

	e.CategoryID = input.CategoryID
	e.Description = input.Description
	e.AmountCents = input.AmountCents
	e.ExpenseDate = input.ExpenseDate
	e.IsShared = input.IsShared
	e.AssignedTo = input.AssignedTo

	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *expenseService) Delete(ctx context.Context, id, userID string) error {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.checkMembership(ctx, e.HouseholdID, userID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *expenseService) checkMembership(ctx context.Context, householdID, userID string) error {
	_, err := s.household.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}
