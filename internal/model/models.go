package model

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Household struct {
	ID         pgtype.UUID `json:"id"`
	Name       string      `json:"name"`
	InviteCode string      `json:"invite_code,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type HouseholdMember struct {
	HouseholdID pgtype.UUID `json:"household_id"`
	UserID      pgtype.UUID `json:"user_id"`
	SalaryCents int64       `json:"salary_cents"`
	Role        string      `json:"role"`
	JoinedAt    time.Time   `json:"joined_at"`
}

type Category struct {
	ID          pgtype.UUID `json:"id"`
	HouseholdID pgtype.UUID `json:"household_id"`
	Name        string      `json:"name"`
	Icon        string      `json:"icon"`
	CreatedAt   time.Time   `json:"created_at"`
}

type FixedBill struct {
	ID          pgtype.UUID `json:"id"`
	HouseholdID pgtype.UUID `json:"household_id"`
	CategoryID  pgtype.UUID `json:"category_id,omitempty"`
	Description string      `json:"description"`
	AmountCents int64       `json:"amount_cents"`
	DueDay      int16       `json:"due_day"`
	IsShared    bool        `json:"is_shared"`
	AssignedTo  pgtype.UUID `json:"assigned_to,omitempty"`
	IsActive    bool        `json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type Expense struct {
	ID          pgtype.UUID `json:"id"`
	HouseholdID pgtype.UUID `json:"household_id"`
	CategoryID  pgtype.UUID `json:"category_id,omitempty"`
	Description string      `json:"description"`
	AmountCents int64       `json:"amount_cents"`
	ExpenseDate time.Time   `json:"expense_date"`
	IsShared    bool        `json:"is_shared"`
	PaidBy      pgtype.UUID `json:"paid_by"`
	AssignedTo  pgtype.UUID `json:"assigned_to,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type MonthlySummary struct {
	ID          pgtype.UUID `json:"id"`
	HouseholdID pgtype.UUID `json:"household_id"`
	Year        int16       `json:"year"`
	Month       int16       `json:"month"`
	GeneratedAt time.Time   `json:"generated_at"`
}

type MonthlySummaryItem struct {
	ID                pgtype.UUID `json:"id"`
	SummaryID         pgtype.UUID `json:"summary_id"`
	UserID            pgtype.UUID `json:"user_id"`
	TotalSharedCents  int64       `json:"total_shared_cents"`
	TotalPersonalCents int64      `json:"total_personal_cents"`
	AmountDueCents    int64       `json:"amount_due_cents"`
}
