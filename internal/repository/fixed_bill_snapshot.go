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

type fixedBillSnapshotRepository struct {
	db *pgxpool.Pool
}

func NewFixedBillSnapshotRepository(db *pgxpool.Pool) domain.FixedBillSnapshotRepository {
	return &fixedBillSnapshotRepository{db: db}
}

func (r *fixedBillSnapshotRepository) Create(ctx context.Context, s *domain.FixedBillSnapshot) error {
	var categoryID, assignedTo *string
	if s.CategoryID != "" {
		categoryID = &s.CategoryID
	}
	if s.AssignedTo != "" {
		assignedTo = &s.AssignedTo
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO fixed_bill_snapshots (fixed_bill_id, household_id, year, month, category_id, description, amount_cents, due_day, is_shared, paid_by, assigned_to)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, frozen_at, created_at, updated_at`,
		s.FixedBillID, s.HouseholdID, s.Year, s.Month, categoryID, s.Description, s.AmountCents, s.DueDay, s.IsShared, s.PaidBy, assignedTo,
	).Scan(&s.ID, &s.FrozenAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create fixed bill snapshot: %w", err)
	}
	return nil
}

func (r *fixedBillSnapshotRepository) FindByMonth(ctx context.Context, householdID string, year, month int) ([]domain.FixedBillSnapshot, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, fixed_bill_id, household_id, year, month, category_id, description,
		        amount_cents, due_day, is_shared, paid_by, assigned_to,
		        frozen_at, created_at, updated_at
		 FROM fixed_bill_snapshots
		 WHERE household_id = $1 AND year = $2 AND month = $3
		 ORDER BY due_day, description`, householdID, year, month,
	)
	if err != nil {
		return nil, fmt.Errorf("list fixed bill snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []domain.FixedBillSnapshot
	for rows.Next() {
		var s domain.FixedBillSnapshot
		var categoryID, assignedTo sql.NullString
		if err := rows.Scan(&s.ID, &s.FixedBillID, &s.HouseholdID, &s.Year, &s.Month,
			&categoryID, &s.Description,
			&s.AmountCents, &s.DueDay, &s.IsShared, &s.PaidBy, &assignedTo,
			&s.FrozenAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan fixed bill snapshot: %w", err)
		}
		if categoryID.Valid {
			s.CategoryID = categoryID.String
		}
		if assignedTo.Valid {
			s.AssignedTo = assignedTo.String
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, nil
}

func (r *fixedBillSnapshotRepository) FindByID(ctx context.Context, id string) (*domain.FixedBillSnapshot, error) {
	s := &domain.FixedBillSnapshot{}
	var categoryID, assignedTo sql.NullString
	err := r.db.QueryRow(ctx,
		`SELECT id, fixed_bill_id, household_id, year, month, category_id, description,
		        amount_cents, due_day, is_shared, paid_by, assigned_to,
		        frozen_at, created_at, updated_at
		 FROM fixed_bill_snapshots
		 WHERE id = $1`, id,
	).Scan(&s.ID, &s.FixedBillID, &s.HouseholdID, &s.Year, &s.Month,
		&categoryID, &s.Description,
		&s.AmountCents, &s.DueDay, &s.IsShared, &s.PaidBy, &assignedTo,
		&s.FrozenAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFixedBillSnapshotNotFound
		}
		return nil, fmt.Errorf("find fixed bill snapshot: %w", err)
	}
	if categoryID.Valid {
		s.CategoryID = categoryID.String
	}
	if assignedTo.Valid {
		s.AssignedTo = assignedTo.String
	}
	return s, nil
}

func (r *fixedBillSnapshotRepository) Update(ctx context.Context, s *domain.FixedBillSnapshot) error {
	var categoryID, assignedTo *string
	if s.CategoryID != "" {
		categoryID = &s.CategoryID
	}
	if s.AssignedTo != "" {
		assignedTo = &s.AssignedTo
	}

	result, err := r.db.Exec(ctx,
		`UPDATE fixed_bill_snapshots SET category_id = $1, description = $2, amount_cents = $3,
		 due_day = $4, is_shared = $5, paid_by = $6, assigned_to = $7, updated_at = now()
		 WHERE id = $8`,
		categoryID, s.Description, s.AmountCents, s.DueDay, s.IsShared, s.PaidBy, assignedTo, s.ID,
	)
	if err != nil {
		return fmt.Errorf("update fixed bill snapshot: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrFixedBillSnapshotNotFound
	}
	return nil
}

func (r *fixedBillSnapshotRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM fixed_bill_snapshots WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete fixed bill snapshot: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrFixedBillSnapshotNotFound
	}
	return nil
}
