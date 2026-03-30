package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	slog.Info("repo: creating category",
		"household_id", c.HouseholdID,
		"name", c.Name,
	)
	err := r.db.QueryRow(ctx,
		`INSERT INTO categories (household_id, name, icon) VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		c.HouseholdID, c.Name, c.Icon,
	).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			slog.Debug("repo: category already exists",
				"household_id", c.HouseholdID,
				"name", c.Name,
			)
			return domain.ErrCategoryExists
		}
		slog.Error("repo: failed to create category",
			"error", err,
			"household_id", c.HouseholdID,
		)
		return fmt.Errorf("create category: %w", err)
	}
	slog.Info("repo: category created",
		"category_id", c.ID,
		"household_id", c.HouseholdID,
	)
	return nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id string) (*domain.Category, error) {
	slog.Debug("repo: fetching category by ID",
		"category_id", id,
	)
	c := &domain.Category{}
	err := r.db.QueryRow(ctx,
		`SELECT id, household_id, name, icon, created_at FROM categories WHERE id = $1`, id,
	).Scan(&c.ID, &c.HouseholdID, &c.Name, &c.Icon, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Debug("repo: category not found",
				"category_id", id,
			)
			return nil, domain.ErrCategoryNotFound
		}
		slog.Error("repo: failed to fetch category",
			"error", err,
			"category_id", id,
		)
		return nil, fmt.Errorf("find category: %w", err)
	}
	return c, nil
}

func (r *categoryRepository) ListByHousehold(ctx context.Context, householdID string) ([]domain.Category, error) {
	slog.Debug("repo: listing categories by household",
		"household_id", householdID,
	)
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, name, icon, created_at FROM categories
		 WHERE household_id = $1 ORDER BY name`, householdID,
	)
	if err != nil {
		slog.Error("repo: failed to list categories",
			"error", err,
			"household_id", householdID,
		)
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
	slog.Info("repo: updating category",
		"category_id", c.ID,
		"household_id", c.HouseholdID,
	)
	result, err := r.db.Exec(ctx,
		`UPDATE categories SET name = $1, icon = $2 WHERE id = $3`,
		c.Name, c.Icon, c.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			slog.Debug("repo: category name conflict on update",
				"category_id", c.ID,
				"name", c.Name,
			)
			return domain.ErrCategoryExists
		}
		slog.Error("repo: failed to update category",
			"error", err,
			"category_id", c.ID,
		)
		return fmt.Errorf("update category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}
	slog.Info("repo: category updated",
		"category_id", c.ID,
	)
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	slog.Info("repo: deleting category",
		"category_id", id,
	)
	result, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		slog.Error("repo: failed to delete category",
			"error", err,
			"category_id", id,
		)
		return fmt.Errorf("delete category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}
	slog.Info("repo: category deleted",
		"category_id", id,
	)
	return nil
}
