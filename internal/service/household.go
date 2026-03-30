package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lguilherme/contas/internal/domain"
)

type householdService struct {
	repo domain.HouseholdRepository
}

func NewHouseholdService(repo domain.HouseholdRepository) domain.HouseholdService {
	return &householdService{repo: repo}
}

func (s *householdService) Create(ctx context.Context, input domain.CreateHouseholdInput, userID string) (*domain.Household, error) {
	slog.Info("service: creating household",
		"user_id", userID,
		"name", input.Name,
	)

	code, err := generateInviteCode()
	if err != nil {
		slog.Error("service: failed to generate invite code",
			"error", err,
		)
		return nil, fmt.Errorf("generate invite code: %w", err)
	}

	h := &domain.Household{
		Name:       input.Name,
		InviteCode: code,
	}

	if err := s.repo.Create(ctx, h, userID); err != nil {
		slog.Error("service: failed to create household",
			"error", err,
			"user_id", userID,
		)
		return nil, err
	}

	slog.Info("service: household created",
		"household_id", h.ID,
		"user_id", userID,
	)
	return h, nil
}

func (s *householdService) GetByID(ctx context.Context, id, userID string) (*domain.Household, error) {
	slog.Debug("service: getting household by id",
		"household_id", id,
		"user_id", userID,
	)

	if _, err := s.repo.GetMember(ctx, id, userID); err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *householdService) ListByUser(ctx context.Context, userID string) ([]domain.Household, error) {
	slog.Debug("service: listing households by user",
		"user_id", userID,
	)

	return s.repo.ListByUser(ctx, userID)
}

func (s *householdService) Update(ctx context.Context, id string, input domain.UpdateHouseholdInput, userID string) (*domain.Household, error) {
	slog.Info("service: updating household",
		"household_id", id,
		"user_id", userID,
	)

	member, err := s.repo.GetMember(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}
	if member.Role != "admin" {
		slog.Warn("service: non-admin attempted household update",
			"household_id", id,
			"user_id", userID,
			"role", member.Role,
		)
		return nil, domain.ErrForbidden
	}

	h := &domain.Household{ID: id, Name: input.Name}
	if err := s.repo.Update(ctx, h); err != nil {
		slog.Error("service: failed to update household",
			"error", err,
			"household_id", id,
		)
		return nil, err
	}

	slog.Info("service: household updated",
		"household_id", id,
	)
	return s.repo.FindByID(ctx, id)
}

func (s *householdService) Delete(ctx context.Context, id, userID string) error {
	slog.Info("service: deleting household",
		"household_id", id,
		"user_id", userID,
	)

	member, err := s.repo.GetMember(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	if member.Role != "admin" {
		slog.Warn("service: non-admin attempted household deletion",
			"household_id", id,
			"user_id", userID,
			"role", member.Role,
		)
		return domain.ErrForbidden
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		slog.Error("service: failed to delete household",
			"error", err,
			"household_id", id,
		)
		return err
	}

	slog.Info("service: household deleted",
		"household_id", id,
	)
	return nil
}

func (s *householdService) Join(ctx context.Context, inviteCode, userID string) (*domain.Household, error) {
	slog.Info("service: user joining household",
		"user_id", userID,
	)

	h, err := s.repo.FindByInviteCode(ctx, inviteCode)
	if err != nil {
		slog.Error("service: failed to find household by invite code",
			"error", err,
		)
		return nil, err
	}

	if err := s.repo.AddMember(ctx, h.ID, userID, "member"); err != nil {
		slog.Error("service: failed to add member to household",
			"error", err,
			"household_id", h.ID,
			"user_id", userID,
		)
		return nil, err
	}

	slog.Info("service: user joined household",
		"household_id", h.ID,
		"user_id", userID,
	)
	return h, nil
}

func (s *householdService) ListMembers(ctx context.Context, householdID, userID string) ([]domain.HouseholdMember, error) {
	slog.Debug("service: listing household members",
		"household_id", householdID,
		"user_id", userID,
	)

	if _, err := s.repo.GetMember(ctx, householdID, userID); err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}

	return s.repo.ListMembers(ctx, householdID)
}

func (s *householdService) UpdateMemberSalary(ctx context.Context, householdID, memberID string, salaryCents int64, userID string) error {
	slog.Info("service: updating member salary",
		"household_id", householdID,
		"member_id", memberID,
		"user_id", userID,
	)

	// Only admin or the member themselves can update salary
	member, err := s.repo.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	if member.Role != "admin" && userID != memberID {
		slog.Warn("service: unauthorized salary update attempt",
			"household_id", householdID,
			"member_id", memberID,
			"user_id", userID,
		)
		return domain.ErrForbidden
	}

	if err := s.repo.UpdateMemberSalary(ctx, householdID, memberID, salaryCents); err != nil {
		slog.Error("service: failed to update member salary",
			"error", err,
			"household_id", householdID,
			"member_id", memberID,
		)
		return err
	}

	slog.Info("service: member salary updated",
		"household_id", householdID,
		"member_id", memberID,
	)
	return nil
}

func (s *householdService) RemoveMember(ctx context.Context, householdID, memberID, userID string) error {
	slog.Info("service: removing household member",
		"household_id", householdID,
		"member_id", memberID,
		"user_id", userID,
	)

	member, err := s.repo.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	if member.Role != "admin" && userID != memberID {
		slog.Warn("service: unauthorized member removal attempt",
			"household_id", householdID,
			"member_id", memberID,
			"user_id", userID,
		)
		return domain.ErrForbidden
	}

	if err := s.repo.RemoveMember(ctx, householdID, memberID); err != nil {
		slog.Error("service: failed to remove household member",
			"error", err,
			"household_id", householdID,
			"member_id", memberID,
		)
		return err
	}

	slog.Info("service: household member removed",
		"household_id", householdID,
		"member_id", memberID,
	)
	return nil
}

func generateInviteCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
