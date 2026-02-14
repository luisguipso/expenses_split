package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lguilherme/contas/internal/domain"
)

type fixedBillRepository struct {
	db *pgxpool.Pool
}

func NewFixedBillRepository(db *pgxpool.Pool) domain.FixedBillRepository {
	return &fixedBillRepository{db: db}
}

func (r *fixedBillRepository) Create(ctx context.Context, b *domain.FixedBill) error {
	var categoryID, assignedTo *string
	if b.CategoryID != "" {
		categoryID = &b.CategoryID
	}
	if b.AssignedTo != "" {
		assignedTo = &b.AssignedTo
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO fixed_bills (household_id, category_id, description, amount_cents, due_day, is_shared, paid_by, assigned_to)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, is_active, created_at, updated_at`,
		b.HouseholdID, categoryID, b.Description, b.AmountCents, b.DueDay, b.IsShared, b.PaidBy, assignedTo,
	).Scan(&b.ID, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create fixed bill: %w", err)
	}
	return nil
}

func (r *fixedBillRepository) FindByID(ctx context.Context, id string) (*domain.FixedBill, error) {
	b := &domain.FixedBill{}
	var categoryID, assignedTo, categoryName, paidByName sql.NullString
	err := r.db.QueryRow(ctx,
		`SELECT fb.id, fb.household_id, fb.category_id, c.name, fb.description,
		        fb.amount_cents, fb.due_day, fb.is_shared, fb.paid_by::text, u.name,
		        fb.assigned_to, fb.is_active,
		        fb.created_at, fb.updated_at
		 FROM fixed_bills fb
		 LEFT JOIN categories c ON c.id = fb.category_id
		 LEFT JOIN users u ON u.id = fb.paid_by
		 WHERE fb.id = $1`, id,
	).Scan(&b.ID, &b.HouseholdID, &categoryID, &categoryName, &b.Description,
		&b.AmountCents, &b.DueDay, &b.IsShared, &b.PaidBy, &paidByName,
		&assignedTo, &b.IsActive,
		&b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFixedBillNotFound
		}
		return nil, fmt.Errorf("find fixed bill: %w", err)
	}
	if categoryID.Valid {
		b.CategoryID = categoryID.String
	}
	if categoryName.Valid {
		b.CategoryName = categoryName.String
	}
	if paidByName.Valid {
		b.PaidByName = paidByName.String
	}
	if assignedTo.Valid {
		b.AssignedTo = assignedTo.String
	}
	return b, nil
}

func (r *fixedBillRepository) ListByHousehold(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
	rows, err := r.db.Query(ctx,
		`SELECT fb.id, fb.household_id, fb.category_id, c.name, fb.description,
		        fb.amount_cents, fb.due_day, fb.is_shared, fb.paid_by::text, u.name,
		        fb.assigned_to, fb.is_active,
		        fb.created_at, fb.updated_at
		 FROM fixed_bills fb
		 LEFT JOIN categories c ON c.id = fb.category_id
		 LEFT JOIN users u ON u.id = fb.paid_by
		 WHERE fb.household_id = $1
		 ORDER BY fb.due_day, fb.description`, householdID,
	)
	if err != nil {
		return nil, fmt.Errorf("list fixed bills: %w", err)
	}
	defer rows.Close()

	var bills []domain.FixedBill
	for rows.Next() {
		var b domain.FixedBill
		var categoryID, assignedTo, categoryName, paidByName sql.NullString
		if err := rows.Scan(&b.ID, &b.HouseholdID, &categoryID, &categoryName, &b.Description,
			&b.AmountCents, &b.DueDay, &b.IsShared, &b.PaidBy, &paidByName,
			&assignedTo, &b.IsActive,
			&b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan fixed bill: %w", err)
		}
		if categoryID.Valid {
			b.CategoryID = categoryID.String
		}
		if categoryName.Valid {
			b.CategoryName = categoryName.String
		}
		if paidByName.Valid {
			b.PaidByName = paidByName.String
		}
		if assignedTo.Valid {
			b.AssignedTo = assignedTo.String
		}
		bills = append(bills, b)
	}
	return bills, nil
}

func (r *fixedBillRepository) Update(ctx context.Context, b *domain.FixedBill) error {
	var categoryID, assignedTo *string
	if b.CategoryID != "" {
		categoryID = &b.CategoryID
	}
	if b.AssignedTo != "" {
		assignedTo = &b.AssignedTo
	}

	result, err := r.db.Exec(ctx,
		`UPDATE fixed_bills SET category_id = $1, description = $2, amount_cents = $3,
		 due_day = $4, is_shared = $5, paid_by = $6, assigned_to = $7, is_active = $8, updated_at = now()
		 WHERE id = $9`,
		categoryID, b.Description, b.AmountCents, b.DueDay, b.IsShared, b.PaidBy, assignedTo, b.IsActive, b.ID,
	)
	if err != nil {
		return fmt.Errorf("update fixed bill: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrFixedBillNotFound
	}
	return nil
}

func (r *fixedBillRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM fixed_bills WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete fixed bill: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrFixedBillNotFound
	}
	return nil
}
