package repository

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lguilherme/contas/internal/domain"
)

type expenseRepository struct {
	db *pgxpool.Pool
}

func NewExpenseRepository(db *pgxpool.Pool) domain.ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) Create(ctx context.Context, e *domain.Expense) error {
	slog.Info("repo: creating expense",
		"household_id", e.HouseholdID,
		"paid_by", e.PaidBy,
	)
	query := `
		INSERT INTO expenses (household_id, category_id, description, amount_cents, expense_date, is_shared, paid_by, assigned_to)
		VALUES ($1, nullif($2,'')::uuid, $3, $4, $5, $6, $7, nullif($8,'')::uuid)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		e.HouseholdID, e.CategoryID, e.Description, e.AmountCents,
		e.ExpenseDate, e.IsShared, e.PaidBy, e.AssignedTo,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		slog.Error("repo: failed to create expense",
			"error", err,
			"household_id", e.HouseholdID,
		)
		return fmt.Errorf("expense create: %w", err)
	}
	slog.Info("repo: expense created",
		"expense_id", e.ID,
		"household_id", e.HouseholdID,
	)
	return nil
}

func (r *expenseRepository) CreateBatch(ctx context.Context, expenses []*domain.Expense) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("expense batch create begin: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO expenses (household_id, category_id, description, amount_cents, expense_date, is_shared, paid_by, assigned_to)
		VALUES ($1, nullif($2,'')::uuid, $3, $4, $5, $6, $7, nullif($8,'')::uuid)
		RETURNING id, created_at, updated_at`

	for _, e := range expenses {
		err := tx.QueryRow(ctx, query,
			e.HouseholdID, e.CategoryID, e.Description, e.AmountCents,
			e.ExpenseDate, e.IsShared, e.PaidBy, e.AssignedTo,
		).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return fmt.Errorf("expense batch create: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *expenseRepository) FindByID(ctx context.Context, id string) (*domain.Expense, error) {
	slog.Debug("repo: fetching expense by ID",
		"expense_id", id,
	)
	query := `
		SELECT e.id, e.household_id, COALESCE(e.category_id::text,''), COALESCE(c.name,''),
		       e.description, e.amount_cents, e.expense_date::text, e.is_shared,
		       e.paid_by::text, COALESCE(u.name,''), COALESCE(e.assigned_to::text,''),
		       e.created_at, e.updated_at
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		LEFT JOIN users u ON e.paid_by = u.id
		WHERE e.id = $1`

	var e domain.Expense
	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.HouseholdID, &e.CategoryID, &e.CategoryName,
		&e.Description, &e.AmountCents, &e.ExpenseDate, &e.IsShared,
		&e.PaidBy, &e.PaidByName, &e.AssignedTo,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		slog.Debug("repo: expense not found",
			"expense_id", id,
		)
		return nil, domain.ErrExpenseNotFound
	}
	if err != nil {
		slog.Error("repo: failed to fetch expense",
			"error", err,
			"expense_id", id,
		)
		return nil, fmt.Errorf("expense find by id: %w", err)
	}
	return &e, nil
}

func (r *expenseRepository) ListByHousehold(ctx context.Context, householdID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
	slog.Debug("repo: listing expenses by household",
		"household_id", householdID,
		"month", filter.Month,
		"year", filter.Year,
	)
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("e.household_id = $%d", argIdx))
	args = append(args, householdID)
	argIdx++

	if filter.Month > 0 && filter.Year > 0 {
		conditions = append(conditions, fmt.Sprintf("EXTRACT(MONTH FROM e.expense_date) = $%d", argIdx))
		args = append(args, filter.Month)
		argIdx++
		conditions = append(conditions, fmt.Sprintf("EXTRACT(YEAR FROM e.expense_date) = $%d", argIdx))
		args = append(args, filter.Year)
		argIdx++
	}

	if filter.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("e.category_id = $%d", argIdx))
		args = append(args, filter.CategoryID)
		argIdx++
	}

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("(e.paid_by = $%d OR e.assigned_to = $%d)", argIdx, argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.household_id, COALESCE(e.category_id::text,''), COALESCE(c.name,''),
		       e.description, e.amount_cents, e.expense_date::text, e.is_shared,
		       e.paid_by::text, COALESCE(u.name,''), COALESCE(e.assigned_to::text,''),
		       e.created_at, e.updated_at
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		LEFT JOIN users u ON e.paid_by = u.id
		WHERE %s
		ORDER BY e.expense_date DESC, e.created_at DESC`,
		strings.Join(conditions, " AND "))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		slog.Error("repo: failed to list expenses",
			"error", err,
			"household_id", householdID,
		)
		return nil, fmt.Errorf("expense list: %w", err)
	}
	defer rows.Close()

	var expenses []domain.Expense
	for rows.Next() {
		var e domain.Expense
		if err := rows.Scan(
			&e.ID, &e.HouseholdID, &e.CategoryID, &e.CategoryName,
			&e.Description, &e.AmountCents, &e.ExpenseDate, &e.IsShared,
			&e.PaidBy, &e.PaidByName, &e.AssignedTo,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("expense list scan: %w", err)
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

func (r *expenseRepository) Update(ctx context.Context, e *domain.Expense) error {
	slog.Info("repo: updating expense",
		"expense_id", e.ID,
		"household_id", e.HouseholdID,
	)
	query := `
		UPDATE expenses
		SET category_id = nullif($1,'')::uuid, description = $2, amount_cents = $3,
		    expense_date = $4, is_shared = $5, assigned_to = nullif($6,'')::uuid,
		    updated_at = now()
		WHERE id = $7`

	result, err := r.db.Exec(ctx, query,
		e.CategoryID, e.Description, e.AmountCents,
		e.ExpenseDate, e.IsShared, e.AssignedTo, e.ID,
	)
	if err != nil {
		slog.Error("repo: failed to update expense",
			"error", err,
			"expense_id", e.ID,
		)
		return fmt.Errorf("expense update: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrExpenseNotFound
	}
	slog.Info("repo: expense updated",
		"expense_id", e.ID,
	)
	return nil
}

func (r *expenseRepository) Delete(ctx context.Context, id string) error {
	slog.Info("repo: deleting expense",
		"expense_id", id,
	)
	result, err := r.db.Exec(ctx, "DELETE FROM expenses WHERE id = $1", id)
	if err != nil {
		slog.Error("repo: failed to delete expense",
			"error", err,
			"expense_id", id,
		)
		return fmt.Errorf("expense delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrExpenseNotFound
	}
	slog.Info("repo: expense deleted",
		"expense_id", id,
	)
	return nil
}
