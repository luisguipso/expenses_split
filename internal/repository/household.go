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

type householdRepository struct {
	db *pgxpool.Pool
}

func NewHouseholdRepository(db *pgxpool.Pool) domain.HouseholdRepository {
	return &householdRepository{db: db}
}

func (r *householdRepository) Create(ctx context.Context, h *domain.Household, adminUserID string) error {
	slog.Info("repo: creating household",
		"admin_user_id", adminUserID,
	)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		slog.Error("repo: failed to begin tx for household creation",
			"error", err,
		)
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO households (name, invite_code) VALUES ($1, $2)
		 RETURNING id, created_at, updated_at`,
		h.Name, h.InviteCode,
	).Scan(&h.ID, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		slog.Error("repo: failed to insert household",
			"error", err,
		)
		return fmt.Errorf("insert household: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO household_members (household_id, user_id, role) VALUES ($1, $2, 'admin')`,
		h.ID, adminUserID,
	)
	if err != nil {
		slog.Error("repo: failed to insert admin member",
			"error", err,
			"household_id", h.ID,
			"admin_user_id", adminUserID,
		)
		return fmt.Errorf("insert admin member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("repo: failed to commit household creation",
			"error", err,
			"household_id", h.ID,
		)
		return err
	}
	slog.Info("repo: household created",
		"household_id", h.ID,
		"admin_user_id", adminUserID,
	)
	return nil
}

func (r *householdRepository) FindByID(ctx context.Context, id string) (*domain.Household, error) {
	slog.Debug("repo: fetching household by ID",
		"household_id", id,
	)
	h := &domain.Household{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, invite_code, created_at, updated_at FROM households WHERE id = $1`, id,
	).Scan(&h.ID, &h.Name, &h.InviteCode, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Debug("repo: household not found",
				"household_id", id,
			)
			return nil, domain.ErrHouseholdNotFound
		}
		slog.Error("repo: failed to fetch household",
			"error", err,
			"household_id", id,
		)
		return nil, fmt.Errorf("find household: %w", err)
	}
	return h, nil
}

func (r *householdRepository) FindByInviteCode(ctx context.Context, code string) (*domain.Household, error) {
	slog.Debug("repo: fetching household by invite code")
	h := &domain.Household{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, invite_code, created_at, updated_at FROM households WHERE invite_code = $1`, code,
	).Scan(&h.ID, &h.Name, &h.InviteCode, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Debug("repo: household not found by invite code")
			return nil, domain.ErrInvalidInviteCode
		}
		slog.Error("repo: failed to fetch household by invite code",
			"error", err,
		)
		return nil, fmt.Errorf("find household by invite: %w", err)
	}
	return h, nil
}

func (r *householdRepository) ListByUser(ctx context.Context, userID string) ([]domain.Household, error) {
	slog.Debug("repo: listing households by user",
		"user_id", userID,
	)
	rows, err := r.db.Query(ctx,
		`SELECT h.id, h.name, h.invite_code, h.created_at, h.updated_at
		 FROM households h
		 JOIN household_members hm ON h.id = hm.household_id
		 WHERE hm.user_id = $1
		 ORDER BY h.created_at DESC`, userID,
	)
	if err != nil {
		slog.Error("repo: failed to list households",
			"error", err,
			"user_id", userID,
		)
		return nil, fmt.Errorf("list households: %w", err)
	}
	defer rows.Close()

	var households []domain.Household
	for rows.Next() {
		var h domain.Household
		if err := rows.Scan(&h.ID, &h.Name, &h.InviteCode, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan household: %w", err)
		}
		households = append(households, h)
	}
	return households, nil
}

func (r *householdRepository) Update(ctx context.Context, h *domain.Household) error {
	slog.Info("repo: updating household",
		"household_id", h.ID,
	)
	result, err := r.db.Exec(ctx,
		`UPDATE households SET name = $1, updated_at = now() WHERE id = $2`,
		h.Name, h.ID,
	)
	if err != nil {
		slog.Error("repo: failed to update household",
			"error", err,
			"household_id", h.ID,
		)
		return fmt.Errorf("update household: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrHouseholdNotFound
	}
	slog.Info("repo: household updated",
		"household_id", h.ID,
	)
	return nil
}

func (r *householdRepository) Delete(ctx context.Context, id string) error {
	slog.Info("repo: deleting household",
		"household_id", id,
	)
	result, err := r.db.Exec(ctx, `DELETE FROM households WHERE id = $1`, id)
	if err != nil {
		slog.Error("repo: failed to delete household",
			"error", err,
			"household_id", id,
		)
		return fmt.Errorf("delete household: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrHouseholdNotFound
	}
	slog.Info("repo: household deleted",
		"household_id", id,
	)
	return nil
}

func (r *householdRepository) AddMember(ctx context.Context, householdID, userID, role string) error {
	slog.Info("repo: adding member to household",
		"household_id", householdID,
		"user_id", userID,
		"role", role,
	)
	_, err := r.db.Exec(ctx,
		`INSERT INTO household_members (household_id, user_id, role) VALUES ($1, $2, $3)`,
		householdID, userID, role,
	)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			slog.Debug("repo: member already exists",
				"household_id", householdID,
				"user_id", userID,
			)
			return domain.ErrAlreadyMember
		}
		slog.Error("repo: failed to add member",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return fmt.Errorf("add member: %w", err)
	}
	slog.Info("repo: member added to household",
		"household_id", householdID,
		"user_id", userID,
	)
	return nil
}

func (r *householdRepository) RemoveMember(ctx context.Context, householdID, userID string) error {
	slog.Info("repo: removing member from household",
		"household_id", householdID,
		"user_id", userID,
	)
	result, err := r.db.Exec(ctx,
		`DELETE FROM household_members WHERE household_id = $1 AND user_id = $2`,
		householdID, userID,
	)
	if err != nil {
		slog.Error("repo: failed to remove member",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return fmt.Errorf("remove member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotMember
	}
	slog.Info("repo: member removed from household",
		"household_id", householdID,
		"user_id", userID,
	)
	return nil
}

func (r *householdRepository) UpdateMemberSalary(ctx context.Context, householdID, userID string, salaryCents int64) error {
	slog.Info("repo: updating member salary",
		"household_id", householdID,
		"user_id", userID,
	)
	result, err := r.db.Exec(ctx,
		`UPDATE household_members SET salary_cents = $1 WHERE household_id = $2 AND user_id = $3`,
		salaryCents, householdID, userID,
	)
	if err != nil {
		slog.Error("repo: failed to update member salary",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return fmt.Errorf("update salary: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotMember
	}
	slog.Info("repo: member salary updated",
		"household_id", householdID,
		"user_id", userID,
	)
	return nil
}

func (r *householdRepository) ListMembers(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
	slog.Debug("repo: listing household members",
		"household_id", householdID,
	)
	rows, err := r.db.Query(ctx,
		`SELECT hm.household_id, hm.user_id, u.name, u.email, hm.salary_cents, hm.role, hm.joined_at
		 FROM household_members hm
		 JOIN users u ON u.id = hm.user_id
		 WHERE hm.household_id = $1
		 ORDER BY hm.joined_at`, householdID,
	)
	if err != nil {
		slog.Error("repo: failed to list members",
			"error", err,
			"household_id", householdID,
		)
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var members []domain.HouseholdMember
	for rows.Next() {
		var m domain.HouseholdMember
		if err := rows.Scan(&m.HouseholdID, &m.UserID, &m.UserName, &m.UserEmail, &m.SalaryCents, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *householdRepository) GetMember(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
	slog.Debug("repo: fetching household member",
		"household_id", householdID,
		"user_id", userID,
	)
	m := &domain.HouseholdMember{}
	err := r.db.QueryRow(ctx,
		`SELECT hm.household_id, hm.user_id, u.name, u.email, hm.salary_cents, hm.role, hm.joined_at
		 FROM household_members hm
		 JOIN users u ON u.id = hm.user_id
		 WHERE hm.household_id = $1 AND hm.user_id = $2`,
		householdID, userID,
	).Scan(&m.HouseholdID, &m.UserID, &m.UserName, &m.UserEmail, &m.SalaryCents, &m.Role, &m.JoinedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Debug("repo: household member not found",
				"household_id", householdID,
				"user_id", userID,
			)
			return nil, domain.ErrNotMember
		}
		slog.Error("repo: failed to get member",
			"error", err,
			"household_id", householdID,
			"user_id", userID,
		)
		return nil, fmt.Errorf("get member: %w", err)
	}
	return m, nil
}
