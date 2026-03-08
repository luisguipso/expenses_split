package service

import (
	"context"
	"errors"

	"github.com/lguilherme/contas/internal/domain"
)

type fixedBillService struct {
	repo      domain.FixedBillRepository
	household domain.HouseholdRepository
}

func NewFixedBillService(repo domain.FixedBillRepository, household domain.HouseholdRepository) domain.FixedBillService {
	return &fixedBillService{repo: repo, household: household}
}

func (s *fixedBillService) Create(ctx context.Context, input domain.CreateFixedBillInput, householdID, userID string) (*domain.FixedBill, error) {
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	b := &domain.FixedBill{
		HouseholdID: householdID,
		CategoryID:  input.CategoryID,
		Description: input.Description,
		AmountCents: input.AmountCents,
		DueDay:      input.DueDay,
		IsShared:    input.IsShared,
		PaidBy:      input.PaidBy,
		AssignedTo:  input.AssignedTo,
	}
	b.SetDefaultPaidBy(userID)
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *fixedBillService) List(ctx context.Context, householdID, userID string) ([]domain.FixedBill, error) {
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByHousehold(ctx, householdID)
}

func (s *fixedBillService) Update(ctx context.Context, id string, input domain.UpdateFixedBillInput, userID string) (*domain.FixedBill, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkMembership(ctx, b.HouseholdID, userID); err != nil {
		return nil, err
	}

	b.CategoryID = input.CategoryID
	b.Description = input.Description
	b.AmountCents = input.AmountCents
	b.DueDay = input.DueDay
	b.IsShared = input.IsShared
	b.PaidBy = input.PaidBy
	b.AssignedTo = input.AssignedTo
	b.IsActive = input.IsActive

	b.SetDefaultPaidBy(userID)

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *fixedBillService) Delete(ctx context.Context, id, userID string) error {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.checkMembership(ctx, b.HouseholdID, userID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *fixedBillService) checkMembership(ctx context.Context, householdID, userID string) error {
	_, err := s.household.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}
