package service

import (
	"context"
	"errors"

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
		return nil, err
	}
	return c, nil
}

func (s *categoryService) List(ctx context.Context, householdID, userID string) ([]domain.Category, error) {
	if err := s.checkMembership(ctx, householdID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByHousehold(ctx, householdID)
}

func (s *categoryService) Update(ctx context.Context, id string, input domain.UpdateCategoryInput, userID string) (*domain.Category, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
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
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *categoryService) Delete(ctx context.Context, id, userID string) error {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.checkMembership(ctx, c.HouseholdID, userID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
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
