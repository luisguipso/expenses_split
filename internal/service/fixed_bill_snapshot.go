package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/lguilherme/contas/internal/domain"
)

type fixedBillSnapshotService struct {
	snapshotRepo  domain.FixedBillSnapshotRepository
	householdRepo domain.HouseholdRepository
}

func NewFixedBillSnapshotService(
	snapshotRepo domain.FixedBillSnapshotRepository,
	householdRepo domain.HouseholdRepository,
) domain.FixedBillSnapshotService {
	return &fixedBillSnapshotService{
		snapshotRepo:  snapshotRepo,
		householdRepo: householdRepo,
	}
}

func (s *fixedBillSnapshotService) Update(ctx context.Context, id string, input domain.UpdateFixedBillSnapshotInput, userID string) (*domain.FixedBillSnapshot, error) {
	snap, err := s.snapshotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkMembership(ctx, snap.HouseholdID, userID); err != nil {
		return nil, err
	}

	snap.CategoryID = input.CategoryID
	snap.Description = input.Description
	snap.AmountCents = input.AmountCents
	snap.DueDay = input.DueDay
	snap.IsShared = input.IsShared
	snap.PaidBy = input.PaidBy
	snap.AssignedTo = input.AssignedTo

	if snap.PaidBy == "" {
		snap.PaidBy = userID
	}

	if err := s.snapshotRepo.Update(ctx, snap); err != nil {
		return nil, fmt.Errorf("update snapshot: %w", err)
	}

	return s.snapshotRepo.FindByID(ctx, id)
}

func (s *fixedBillSnapshotService) checkMembership(ctx context.Context, householdID, userID string) error {
	_, err := s.householdRepo.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}
