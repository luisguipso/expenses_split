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

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	slog.Info("repo: creating user",
		"email", user.Email,
	)
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash, email_verified)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		user.Name, user.Email, user.PasswordHash, user.EmailVerified,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			slog.Debug("repo: user email already exists",
				"email", user.Email,
			)
			return domain.ErrEmailExists
		}
		slog.Error("repo: failed to create user",
			"error", err,
		)
		return fmt.Errorf("create user: %w", err)
	}
	slog.Info("repo: user created",
		"user_id", user.ID,
	)
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	slog.Debug("repo: fetching user by email",
		"email", email,
	)
	user := &domain.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, email_verified, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Debug("repo: user not found by email",
				"email", email,
			)
			return nil, domain.ErrUserNotFound
		}
		slog.Error("repo: failed to fetch user by email",
			"error", err,
		)
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	slog.Debug("repo: fetching user by ID",
		"user_id", id,
	)
	user := &domain.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, email_verified, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Debug("repo: user not found by ID",
				"user_id", id,
			)
			return nil, domain.ErrUserNotFound
		}
		slog.Error("repo: failed to fetch user by ID",
			"error", err,
			"user_id", id,
		)
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, userID string) error {
	slog.Info("repo: verifying user email",
		"user_id", userID,
	)
	_, err := r.db.Exec(ctx,
		`UPDATE users SET email_verified = true, updated_at = now() WHERE id = $1`,
		userID,
	)
	if err != nil {
		slog.Error("repo: failed to verify user email",
			"error", err,
			"user_id", userID,
		)
		return fmt.Errorf("verify email: %w", err)
	}
	slog.Info("repo: user email verified",
		"user_id", userID,
	)
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	slog.Info("repo: updating user password",
		"user_id", userID,
	)
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`,
		passwordHash, userID,
	)
	if err != nil {
		slog.Error("repo: failed to update user password",
			"error", err,
			"user_id", userID,
		)
		return fmt.Errorf("update password: %w", err)
	}
	slog.Info("repo: user password updated",
		"user_id", userID,
	)
	return nil
}
