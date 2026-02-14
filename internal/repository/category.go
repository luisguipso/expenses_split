package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lguilherme/contas/internal/domain"
)

type categoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, c *domain.Category) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO categories (household_id, name, icon) VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		c.HouseholdID, c.Name, c.Icon,
	).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return domain.ErrCategoryExists
		}
		return fmt.Errorf("create category: %w", err)
	}
	return nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id string) (*domain.Category, error) {
	c := &domain.Category{}
	err := r.db.QueryRow(ctx,
		`SELECT id, household_id, name, icon, created_at FROM categories WHERE id = $1`, id,
	).Scan(&c.ID, &c.HouseholdID, &c.Name, &c.Icon, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("find category: %w", err)
	}
	return c, nil
}

func (r *categoryRepository) ListByHousehold(ctx context.Context, householdID string) ([]domain.Category, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, name, icon, created_at FROM categories
		 WHERE household_id = $1 ORDER BY name`, householdID,
	)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.HouseholdID, &c.Name, &c.Icon, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, c *domain.Category) error {
	result, err := r.db.Exec(ctx,
		`UPDATE categories SET name = $1, icon = $2 WHERE id = $3`,
		c.Name, c.Icon, c.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return domain.ErrCategoryExists
		}
		return fmt.Errorf("update category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}
