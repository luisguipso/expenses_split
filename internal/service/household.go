package service

import (
	"context"
	"errors"

	"github.com/lguilherme/contas/internal/domain"
)

type householdService struct {
	repo domain.HouseholdRepository
}

func NewHouseholdService(repo domain.HouseholdRepository) domain.HouseholdService {
	return &householdService{repo: repo}
}

func (s *householdService) Create(ctx context.Context, input domain.CreateHouseholdInput, userID string) (*domain.Household, error) {
	h, err := domain.NewHousehold(input.Name)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, h, userID); err != nil {
		return nil, err
	}

	return h, nil
}

func (s *householdService) GetByID(ctx context.Context, id, userID string) (*domain.Household, error) {
	if _, err := s.repo.GetMember(ctx, id, userID); err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *householdService) ListByUser(ctx context.Context, userID string) ([]domain.Household, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *householdService) Update(ctx context.Context, id string, input domain.UpdateHouseholdInput, userID string) (*domain.Household, error) {
	member, err := s.repo.GetMember(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}
	if !member.IsAdmin() {
		return nil, domain.ErrForbidden
	}

	h := &domain.Household{ID: id, Name: input.Name}
	if err := s.repo.Update(ctx, h); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *householdService) Delete(ctx context.Context, id, userID string) error {
	member, err := s.repo.GetMember(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	if !member.IsAdmin() {
		return domain.ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}

func (s *householdService) Join(ctx context.Context, inviteCode, userID string) (*domain.Household, error) {
	h, err := s.repo.FindByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, err
	}

	if err := s.repo.AddMember(ctx, h.ID, userID, "member"); err != nil {
		return nil, err
	}

	return h, nil
}

func (s *householdService) ListMembers(ctx context.Context, householdID, userID string) ([]domain.HouseholdMember, error) {
	if _, err := s.repo.GetMember(ctx, householdID, userID); err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}

	return s.repo.ListMembers(ctx, householdID)
}

func (s *householdService) UpdateMemberSalary(ctx context.Context, householdID, memberID string, salaryCents int64, userID string) error {
	// Only admin or the member themselves can update salary
	member, err := s.repo.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	if !member.CanManage(memberID) {
		return domain.ErrForbidden
	}

	return s.repo.UpdateMemberSalary(ctx, householdID, memberID, salaryCents)
}

func (s *householdService) RemoveMember(ctx context.Context, householdID, memberID, userID string) error {
	member, err := s.repo.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	if !member.CanManage(memberID) {
		return domain.ErrForbidden
	}

	return s.repo.RemoveMember(ctx, householdID, memberID)
}
