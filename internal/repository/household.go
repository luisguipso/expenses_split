package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lguilherme/contas/internal/domain"
)

type householdRepository struct {
	db *pgxpool.Pool
}

func NewHouseholdRepository(db *pgxpool.Pool) domain.HouseholdRepository {
	return &householdRepository{db: db}
}

func (r *householdRepository) Create(ctx context.Context, h *domain.Household, adminUserID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO households (name, invite_code) VALUES ($1, $2)
		 RETURNING id, created_at, updated_at`,
		h.Name, h.InviteCode,
	).Scan(&h.ID, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert household: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO household_members (household_id, user_id, role) VALUES ($1, $2, 'admin')`,
		h.ID, adminUserID,
	)
	if err != nil {
		return fmt.Errorf("insert admin member: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *householdRepository) FindByID(ctx context.Context, id string) (*domain.Household, error) {
	h := &domain.Household{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, invite_code, split_mode, created_at, updated_at FROM households WHERE id = $1`, id,
	).Scan(&h.ID, &h.Name, &h.InviteCode, &h.SplitMode, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHouseholdNotFound
		}
		return nil, fmt.Errorf("find household: %w", err)
	}
	return h, nil
}

func (r *householdRepository) FindByInviteCode(ctx context.Context, code string) (*domain.Household, error) {
	h := &domain.Household{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, invite_code, split_mode, created_at, updated_at FROM households WHERE invite_code = $1`, code,
	).Scan(&h.ID, &h.Name, &h.InviteCode, &h.SplitMode, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidInviteCode
		}
		return nil, fmt.Errorf("find household by invite: %w", err)
	}
	return h, nil
}

func (r *householdRepository) ListByUser(ctx context.Context, userID string) ([]domain.Household, error) {
	rows, err := r.db.Query(ctx,
		`SELECT h.id, h.name, h.invite_code, h.split_mode, h.created_at, h.updated_at
		 FROM households h
		 JOIN household_members hm ON h.id = hm.household_id
		 WHERE hm.user_id = $1
		 ORDER BY h.created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list households: %w", err)
	}
	defer rows.Close()

	var households []domain.Household
	for rows.Next() {
		var h domain.Household
		if err := rows.Scan(&h.ID, &h.Name, &h.InviteCode, &h.SplitMode, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan household: %w", err)
		}
		households = append(households, h)
	}
	return households, nil
}

func (r *householdRepository) Update(ctx context.Context, h *domain.Household) error {
	result, err := r.db.Exec(ctx,
		`UPDATE households SET name = $1, updated_at = now() WHERE id = $2`,
		h.Name, h.ID,
	)
	if err != nil {
		return fmt.Errorf("update household: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrHouseholdNotFound
	}
	return nil
}

func (r *householdRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM households WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete household: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrHouseholdNotFound
	}
	return nil
}

func (r *householdRepository) AddMember(ctx context.Context, householdID, userID, role string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO household_members (household_id, user_id, role) VALUES ($1, $2, $3)`,
		householdID, userID, role,
	)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return domain.ErrAlreadyMember
		}
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *householdRepository) RemoveMember(ctx context.Context, householdID, userID string) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM household_members WHERE household_id = $1 AND user_id = $2`,
		householdID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotMember
	}
	return nil
}

func (r *householdRepository) UpdateMemberSalary(ctx context.Context, householdID, userID string, salaryCents int64) error {
	result, err := r.db.Exec(ctx,
		`UPDATE household_members SET salary_cents = $1 WHERE household_id = $2 AND user_id = $3`,
		salaryCents, householdID, userID,
	)
	if err != nil {
		return fmt.Errorf("update salary: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotMember
	}
	return nil
}

func (r *householdRepository) UpdateSplitMode(ctx context.Context, householdID, splitMode string) error {
	result, err := r.db.Exec(ctx,
		`UPDATE households SET split_mode = $1, updated_at = now() WHERE id = $2`,
		splitMode, householdID,
	)
	if err != nil {
		return fmt.Errorf("update split mode: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrHouseholdNotFound
	}
	return nil
}

func (r *householdRepository) UpdateMemberSplitPercentage(ctx context.Context, householdID, userID string, percentage int) error {
	result, err := r.db.Exec(ctx,
		`UPDATE household_members SET split_percentage = $1 WHERE household_id = $2 AND user_id = $3`,
		percentage, householdID, userID,
	)
	if err != nil {
		return fmt.Errorf("update split percentage: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotMember
	}
	return nil
}

func (r *householdRepository) ListMembers(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT hm.household_id, hm.user_id, u.name, u.email, hm.salary_cents, hm.split_percentage, hm.role, hm.joined_at
		 FROM household_members hm
		 JOIN users u ON u.id = hm.user_id
		 WHERE hm.household_id = $1
		 ORDER BY hm.joined_at`, householdID,
	)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var members []domain.HouseholdMember
	for rows.Next() {
		var m domain.HouseholdMember
		if err := rows.Scan(&m.HouseholdID, &m.UserID, &m.UserName, &m.UserEmail, &m.SalaryCents, &m.SplitPercentage, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *householdRepository) GetMember(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
	m := &domain.HouseholdMember{}
	err := r.db.QueryRow(ctx,
		`SELECT hm.household_id, hm.user_id, u.name, u.email, hm.salary_cents, hm.split_percentage, hm.role, hm.joined_at
		 FROM household_members hm
		 JOIN users u ON u.id = hm.user_id
		 WHERE hm.household_id = $1 AND hm.user_id = $2`,
		householdID, userID,
	).Scan(&m.HouseholdID, &m.UserID, &m.UserName, &m.UserEmail, &m.SalaryCents, &m.SplitPercentage, &m.Role, &m.JoinedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotMember
		}
		return nil, fmt.Errorf("get member: %w", err)
	}
	return m, nil
}
