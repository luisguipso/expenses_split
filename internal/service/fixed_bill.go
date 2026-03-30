package service

import (
	"context"
	"errors"
	"log/slog"

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
	slog.Info("service: creating fixed bill",
		"household_id", householdID,
		"user_id", userID,
		"amount_cents", input.AmountCents,
		"due_day", input.DueDay,
	)

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
	if b.PaidBy == "" {
		b.PaidBy = userID
	}
	if err := s.repo.Create(ctx, b); err != nil {
		slog.Error("service: failed to create fixed bill",
			"error", err,
			"household_id", householdID,
		)
		return nil, err
	}

	slog.Info("service: fixed bill created",
		"fixed_bill_id", b.ID,
		"household_id", householdID,
	)
	return b, nil
}

func (s *fixedBillService) List(ctx context.Context, householdID, userID string) ([]domain.FixedBill, error) {
	slog.Debug("service: listing fixed bills",
		"household_id", householdID,
		"user_id", userID,
	)

	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByHousehold(ctx, householdID)
}

func (s *fixedBillService) Update(ctx context.Context, id string, input domain.UpdateFixedBillInput, userID string) (*domain.FixedBill, error) {
	slog.Info("service: updating fixed bill",
		"fixed_bill_id", id,
		"user_id", userID,
	)

	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.Error("service: failed to find fixed bill for update",
			"error", err,
			"fixed_bill_id", id,
		)
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

	if b.PaidBy == "" {
		b.PaidBy = userID
	}

	if err := s.repo.Update(ctx, b); err != nil {
		slog.Error("service: failed to update fixed bill",
			"error", err,
			"fixed_bill_id", id,
		)
		return nil, err
	}

	slog.Info("service: fixed bill updated",
		"fixed_bill_id", id,
		"household_id", b.HouseholdID,
	)
	return s.repo.FindByID(ctx, id)
}

func (s *fixedBillService) Delete(ctx context.Context, id, userID string) error {
	slog.Info("service: deleting fixed bill",
		"fixed_bill_id", id,
		"user_id", userID,
	)

	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.Error("service: failed to find fixed bill for deletion",
			"error", err,
			"fixed_bill_id", id,
		)
		return err
	}

	if err := s.checkMembership(ctx, b.HouseholdID, userID); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		slog.Error("service: failed to delete fixed bill",
			"error", err,
			"fixed_bill_id", id,
		)
		return err
	}

	slog.Info("service: fixed bill deleted",
		"fixed_bill_id", id,
		"household_id", b.HouseholdID,
	)
	return nil
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
