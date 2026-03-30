package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

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
	slog.Info("repo: creating fixed bill snapshot",
		"fixed_bill_id", s.FixedBillID,
		"household_id", s.HouseholdID,
		"year", s.Year,
		"month", s.Month,
	)
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
		slog.Error("repo: failed to create fixed bill snapshot",
			"error", err,
			"fixed_bill_id", s.FixedBillID,
			"household_id", s.HouseholdID,
		)
		return fmt.Errorf("create fixed bill snapshot: %w", err)
	}
	slog.Info("repo: fixed bill snapshot created",
		"snapshot_id", s.ID,
		"household_id", s.HouseholdID,
	)
	return nil
}

func (r *fixedBillSnapshotRepository) FindByMonth(ctx context.Context, householdID string, year, month int) ([]domain.FixedBillSnapshot, error) {
	slog.Debug("repo: fetching fixed bill snapshots by month",
		"household_id", householdID,
		"year", year,
		"month", month,
	)
	rows, err := r.db.Query(ctx,
		`SELECT id, fixed_bill_id, household_id, year, month, category_id, description,
		        amount_cents, due_day, is_shared, paid_by, assigned_to,
		        frozen_at, created_at, updated_at
		 FROM fixed_bill_snapshots
		 WHERE household_id = $1 AND year = $2 AND month = $3
		 ORDER BY due_day, description`, householdID, year, month,
	)
	if err != nil {
		slog.Error("repo: failed to list fixed bill snapshots",
			"error", err,
			"household_id", householdID,
		)
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
	slog.Debug("repo: fetching fixed bill snapshot by ID",
		"snapshot_id", id,
	)
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
			slog.Debug("repo: fixed bill snapshot not found",
				"snapshot_id", id,
			)
			return nil, domain.ErrFixedBillSnapshotNotFound
		}
		slog.Error("repo: failed to fetch fixed bill snapshot",
			"error", err,
			"snapshot_id", id,
		)
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
	slog.Info("repo: updating fixed bill snapshot",
		"snapshot_id", s.ID,
		"household_id", s.HouseholdID,
	)
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
		slog.Error("repo: failed to update fixed bill snapshot",
			"error", err,
			"snapshot_id", s.ID,
		)
		return fmt.Errorf("update fixed bill snapshot: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrFixedBillSnapshotNotFound
	}
	slog.Info("repo: fixed bill snapshot updated",
		"snapshot_id", s.ID,
	)
	return nil
}

func (r *fixedBillSnapshotRepository) Delete(ctx context.Context, id string) error {
	slog.Info("repo: deleting fixed bill snapshot",
		"snapshot_id", id,
	)
	result, err := r.db.Exec(ctx, `DELETE FROM fixed_bill_snapshots WHERE id = $1`, id)
	if err != nil {
		slog.Error("repo: failed to delete fixed bill snapshot",
			"error", err,
			"snapshot_id", id,
		)
		return fmt.Errorf("delete fixed bill snapshot: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrFixedBillSnapshotNotFound
	}
	slog.Info("repo: fixed bill snapshot deleted",
		"snapshot_id", id,
	)
	return nil
}
