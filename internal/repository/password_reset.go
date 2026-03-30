package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lguilherme/contas/internal/domain"
)

type passwordResetRepository struct {
	db *pgxpool.Pool
}

func NewPasswordResetRepository(db *pgxpool.Pool) domain.PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(ctx context.Context, pr *domain.PasswordReset) error {
	slog.Info("repo: creating password reset",
		"user_id", pr.UserID,
	)
	err := r.db.QueryRow(ctx,
		`INSERT INTO password_resets (user_id, email, token, expires_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		pr.UserID, pr.Email, pr.Token, pr.ExpiresAt,
	).Scan(&pr.ID, &pr.CreatedAt)
	if err != nil {
		slog.Error("repo: failed to create password reset",
			"error", err,
			"user_id", pr.UserID,
		)
		return fmt.Errorf("create password reset: %w", err)
	}
	slog.Info("repo: password reset created",
		"reset_id", pr.ID,
		"user_id", pr.UserID,
	)
	return nil
}

func (r *passwordResetRepository) FindByToken(ctx context.Context, token string) (*domain.PasswordReset, error) {
	slog.Debug("repo: fetching password reset by token")
	pr := &domain.PasswordReset{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, email, token, expires_at, used, created_at
		 FROM password_resets
		 WHERE token = $1 AND used = false
		 ORDER BY created_at DESC
		 LIMIT 1`,
		token,
	).Scan(&pr.ID, &pr.UserID, &pr.Email, &pr.Token, &pr.ExpiresAt, &pr.Used, &pr.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Debug("repo: password reset not found by token")
			return nil, domain.ErrResetTokenInvalid
		}
		slog.Error("repo: failed to fetch password reset by token",
			"error", err,
		)
		return nil, fmt.Errorf("find password reset by token: %w", err)
	}
	return pr, nil
}

func (r *passwordResetRepository) MarkUsed(ctx context.Context, id string) error {
	slog.Info("repo: marking password reset as used",
		"reset_id", id,
	)
	_, err := r.db.Exec(ctx,
		`UPDATE password_resets SET used = true WHERE id = $1`,
		id,
	)
	if err != nil {
		slog.Error("repo: failed to mark password reset as used",
			"error", err,
			"reset_id", id,
		)
		return fmt.Errorf("mark password reset used: %w", err)
	}
	slog.Info("repo: password reset marked as used",
		"reset_id", id,
	)
	return nil
}
