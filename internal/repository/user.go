package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lguilherme/contas/internal/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already exists")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		user.Name, user.Email, user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrEmailExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && (fmt.Sprintf("%v", err) == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" ||
		contains(err.Error(), "23505"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
