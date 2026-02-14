package service

import (
	"context"
	"errors"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func memberOK() func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
	return func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member"}, nil
	}
}

func memberForbidden() func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
	return func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return nil, domain.ErrNotMember
	}
}

func TestCategoryService_Create_Success(t *testing.T) {
	catRepo := &mock.CategoryRepository{
		CreateFn: func(ctx context.Context, c *domain.Category) error {
			c.ID = "cat-1"
			return nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewCategoryService(catRepo, hhRepo)

	cat, err := svc.Create(context.Background(), domain.CreateCategoryInput{Name: "Aluguel", Icon: "🏠"}, "hh-1", "user-1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if cat.ID != "cat-1" {
		t.Errorf("expected ID cat-1, got %s", cat.ID)
	}
	if cat.Icon != "🏠" {
		t.Errorf("expected icon 🏠, got %s", cat.Icon)
	}
}

func TestCategoryService_Create_DefaultIcon(t *testing.T) {
	catRepo := &mock.CategoryRepository{
		CreateFn: func(ctx context.Context, c *domain.Category) error {
			if c.Icon != "📦" {
				t.Errorf("expected default icon 📦, got %s", c.Icon)
			}
			c.ID = "cat-2"
			return nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewCategoryService(catRepo, hhRepo)

	_, err := svc.Create(context.Background(), domain.CreateCategoryInput{Name: "Outros"}, "hh-1", "user-1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestCategoryService_Create_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewCategoryService(nil, hhRepo)

	_, err := svc.Create(context.Background(), domain.CreateCategoryInput{Name: "X"}, "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestCategoryService_Create_Duplicate(t *testing.T) {
	catRepo := &mock.CategoryRepository{
		CreateFn: func(ctx context.Context, c *domain.Category) error {
			return domain.ErrCategoryExists
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewCategoryService(catRepo, hhRepo)

	_, err := svc.Create(context.Background(), domain.CreateCategoryInput{Name: "Dup"}, "hh-1", "user-1")
	if !errors.Is(err, domain.ErrCategoryExists) {
		t.Errorf("expected ErrCategoryExists, got %v", err)
	}
}

func TestCategoryService_List_Success(t *testing.T) {
	catRepo := &mock.CategoryRepository{
		ListByHouseholdFn: func(ctx context.Context, householdID string) ([]domain.Category, error) {
			return []domain.Category{{ID: "c1", Name: "A"}, {ID: "c2", Name: "B"}}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewCategoryService(catRepo, hhRepo)

	cats, err := svc.List(context.Background(), "hh-1", "user-1")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(cats) != 2 {
		t.Errorf("expected 2 categories, got %d", len(cats))
	}
}

func TestCategoryService_List_NotMember(t *testing.T) {
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewCategoryService(nil, hhRepo)

	_, err := svc.List(context.Background(), "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestCategoryService_Update_Success(t *testing.T) {
	calls := 0
	catRepo := &mock.CategoryRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Category, error) {
			calls++
			if calls == 1 {
				return &domain.Category{ID: id, HouseholdID: "hh-1", Name: "Old", Icon: "📦"}, nil
			}
			return &domain.Category{ID: id, HouseholdID: "hh-1", Name: "New", Icon: "🏠"}, nil
		},
		UpdateFn: func(ctx context.Context, c *domain.Category) error { return nil },
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewCategoryService(catRepo, hhRepo)

	cat, err := svc.Update(context.Background(), "cat-1", domain.UpdateCategoryInput{Name: "New", Icon: "🏠"}, "user-1")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if cat.Name != "New" {
		t.Errorf("expected name New, got %s", cat.Name)
	}
}

func TestCategoryService_Update_NotFound(t *testing.T) {
	catRepo := &mock.CategoryRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Category, error) {
			return nil, domain.ErrCategoryNotFound
		},
	}
	svc := NewCategoryService(catRepo, nil)

	_, err := svc.Update(context.Background(), "bad", domain.UpdateCategoryInput{Name: "X"}, "user-1")
	if !errors.Is(err, domain.ErrCategoryNotFound) {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCategoryService_Delete_Success(t *testing.T) {
	catRepo := &mock.CategoryRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Category, error) {
			return &domain.Category{ID: id, HouseholdID: "hh-1"}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error { return nil },
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberOK()}
	svc := NewCategoryService(catRepo, hhRepo)

	err := svc.Delete(context.Background(), "cat-1", "user-1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestCategoryService_Delete_NotMember(t *testing.T) {
	catRepo := &mock.CategoryRepository{
		FindByIDFn: func(ctx context.Context, id string) (*domain.Category, error) {
			return &domain.Category{ID: id, HouseholdID: "hh-1"}, nil
		},
	}
	hhRepo := &mock.HouseholdRepository{GetMemberFn: memberForbidden()}
	svc := NewCategoryService(catRepo, hhRepo)

	err := svc.Delete(context.Background(), "cat-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
