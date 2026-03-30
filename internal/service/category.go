package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/lguilherme/contas/internal/domain"
)

type categoryService struct {
	repo      domain.CategoryRepository
	household domain.HouseholdRepository
}

func NewCategoryService(repo domain.CategoryRepository, household domain.HouseholdRepository) domain.CategoryService {
	return &categoryService{repo: repo, household: household}
}

func (s *categoryService) Create(ctx context.Context, input domain.CreateCategoryInput, householdID, userID string) (*domain.Category, error) {
	slog.Info("service: creating category",
		"household_id", householdID,
		"user_id", userID,
		"name", input.Name,
	)

	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}

	icon := input.Icon
	if icon == "" {
		icon = "📦"
	}

	c := &domain.Category{
		HouseholdID: householdID,
		Name:        input.Name,
		Icon:        icon,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		slog.Error("service: failed to create category",
			"error", err,
			"household_id", householdID,
		)
		return nil, err
	}

	slog.Info("service: category created",
		"category_id", c.ID,
		"household_id", householdID,
	)
	return c, nil
}

func (s *categoryService) List(ctx context.Context, householdID, userID string) ([]domain.Category, error) {
	slog.Debug("service: listing categories",
		"household_id", householdID,
		"user_id", userID,
	)

	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByHousehold(ctx, householdID)
}

func (s *categoryService) Update(ctx context.Context, id string, input domain.UpdateCategoryInput, userID string) (*domain.Category, error) {
	slog.Info("service: updating category",
		"category_id", id,
		"user_id", userID,
	)

	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.Error("service: failed to find category for update",
			"error", err,
			"category_id", id,
		)
		return nil, err
	}

	if err := s.checkMembership(ctx, c.HouseholdID, userID); err != nil {
		return nil, err
	}

	if input.Name != "" {
		c.Name = input.Name
	}
	if input.Icon != "" {
		c.Icon = input.Icon
	}

	if err := s.repo.Update(ctx, c); err != nil {
		slog.Error("service: failed to update category",
			"error", err,
			"category_id", id,
		)
		return nil, err
	}

	slog.Info("service: category updated",
		"category_id", id,
		"household_id", c.HouseholdID,
	)
	return s.repo.FindByID(ctx, id)
}

func (s *categoryService) Delete(ctx context.Context, id, userID string) error {
	slog.Info("service: deleting category",
		"category_id", id,
		"user_id", userID,
	)

	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.Error("service: failed to find category for deletion",
			"error", err,
			"category_id", id,
		)
		return err
	}

	if err := s.checkMembership(ctx, c.HouseholdID, userID); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		slog.Error("service: failed to delete category",
			"error", err,
			"category_id", id,
		)
		return err
	}

	slog.Info("service: category deleted",
		"category_id", id,
		"household_id", c.HouseholdID,
	)
	return nil
}

func (s *categoryService) checkMembership(ctx context.Context, householdID, userID string) error {
	_, err := s.household.GetMember(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}
