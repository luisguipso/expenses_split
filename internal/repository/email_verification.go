package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lguilherme/contas/internal/domain"
)

type emailVerificationRepository struct {
	db *pgxpool.Pool
}

func NewEmailVerificationRepository(db *pgxpool.Pool) domain.EmailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(ctx context.Context, v *domain.EmailVerification) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO email_verifications (user_id, email, code, expires_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		v.UserID, v.Email, v.Code, v.ExpiresAt,
	).Scan(&v.ID, &v.CreatedAt)
	if err != nil {
		return fmt.Errorf("create email verification: %w", err)
	}
	return nil
}

func (r *emailVerificationRepository) FindLatestByEmail(ctx context.Context, email string) (*domain.EmailVerification, error) {
	v := &domain.EmailVerification{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, email, code, expires_at, used, created_at
		 FROM email_verifications
		 WHERE email = $1 AND used = false
		 ORDER BY created_at DESC
		 LIMIT 1`,
		email,
	).Scan(&v.ID, &v.UserID, &v.Email, &v.Code, &v.ExpiresAt, &v.Used, &v.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidVerificationCode
		}
		return nil, fmt.Errorf("find latest verification by email: %w", err)
	}
	return v, nil
}

func (r *emailVerificationRepository) MarkUsed(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE email_verifications SET used = true WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("mark verification used: %w", err)
	}
	return nil
}
