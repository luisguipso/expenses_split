package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lguilherme/contas/internal/domain"
)

type summaryRepository struct {
	db *pgxpool.Pool
}

func NewSummaryRepository(db *pgxpool.Pool) domain.SummaryRepository {
	return &summaryRepository{db: db}
}

func (r *summaryRepository) Upsert(ctx context.Context, s *domain.MonthlySummary) error {
	slog.Info("repo: upserting monthly summary",
		"household_id", s.HouseholdID,
		"year", s.Year,
		"month", s.Month,
	)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		slog.Error("repo: failed to begin tx for summary upsert",
			"error", err,
		)
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Upsert the summary header
	err = tx.QueryRow(ctx,
		`INSERT INTO monthly_summaries (household_id, year, month)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (household_id, year, month)
		 DO UPDATE SET generated_at = now()
		 RETURNING id, generated_at`,
		s.HouseholdID, s.Year, s.Month,
	).Scan(&s.ID, &s.GeneratedAt)
	if err != nil {
		slog.Error("repo: failed to upsert summary",
			"error", err,
			"household_id", s.HouseholdID,
		)
		return fmt.Errorf("upsert summary: %w", err)
	}

	// Delete old items for this summary
	_, err = tx.Exec(ctx, `DELETE FROM monthly_summary_items WHERE summary_id = $1`, s.ID)
	if err != nil {
		slog.Error("repo: failed to delete old summary items",
			"error", err,
			"summary_id", s.ID,
		)
		return fmt.Errorf("delete old items: %w", err)
	}

	// Insert new items
	for i := range s.Items {
		item := &s.Items[i]
		item.SummaryID = s.ID
		err = tx.QueryRow(ctx,
			`INSERT INTO monthly_summary_items (summary_id, user_id, total_shared_cents, total_personal_cents, amount_due_cents, total_paid_cents, balance_cents)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING id`,
			item.SummaryID, item.UserID, item.TotalSharedCents, item.TotalPersonalCents, item.AmountDueCents, item.TotalPaidCents, item.BalanceCents,
		).Scan(&item.ID)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *summaryRepository) FindByMonth(ctx context.Context, householdID string, year, month int) (*domain.MonthlySummary, error) {
	slog.Debug("repo: fetching monthly summary",
		"household_id", householdID,
		"year", year,
		"month", month,
	)
	s := &domain.MonthlySummary{}
	err := r.db.QueryRow(ctx,
		`SELECT id, household_id, year, month, generated_at
		 FROM monthly_summaries
		 WHERE household_id = $1 AND year = $2 AND month = $3`,
		householdID, year, month,
	).Scan(&s.ID, &s.HouseholdID, &s.Year, &s.Month, &s.GeneratedAt)
	if err == pgx.ErrNoRows {
		slog.Debug("repo: monthly summary not found",
			"household_id", householdID,
			"year", year,
			"month", month,
		)
		return nil, domain.ErrSummaryNotFound
	}
	if err != nil {
		slog.Error("repo: failed to fetch monthly summary",
			"error", err,
			"household_id", householdID,
		)
		return nil, fmt.Errorf("find summary: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT si.id, si.summary_id, si.user_id, u.name,
		        si.total_shared_cents, si.total_personal_cents, si.amount_due_cents,
		        si.total_paid_cents, si.balance_cents
		 FROM monthly_summary_items si
		 JOIN users u ON u.id = si.user_id
		 WHERE si.summary_id = $1
		 ORDER BY si.amount_due_cents DESC`, s.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.MonthlySummaryItem
		if err := rows.Scan(
			&item.ID, &item.SummaryID, &item.UserID, &item.UserName,
			&item.TotalSharedCents, &item.TotalPersonalCents, &item.AmountDueCents,
			&item.TotalPaidCents, &item.BalanceCents,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		s.Items = append(s.Items, item)
	}

	return s, nil
}
