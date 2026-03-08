package domain

import (
	"fmt"
	"time"
)

type Expense struct {
	ID           string    `json:"id"`
	HouseholdID  string    `json:"household_id"`
	CategoryID   string    `json:"category_id,omitempty"`
	CategoryName string    `json:"category_name,omitempty"`
	Description  string    `json:"description"`
	AmountCents  int64     `json:"amount_cents"`
	ExpenseDate  string    `json:"expense_date"`
	IsShared     bool      `json:"is_shared"`
	PaidBy       string    `json:"paid_by"`
	PaidByName   string    `json:"paid_by_name,omitempty"`
	AssignedTo   string    `json:"assigned_to,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (e *Expense) Validate() error {
	if e.AmountCents <= 0 {
		return fmt.Errorf("amount must be positive, got %d", e.AmountCents)
	}
	return nil
}

func (e *Expense) SetDefaults(userID string) {
	if e.ExpenseDate == "" {
		e.ExpenseDate = time.Now().Format("2006-01-02")
	}
	if e.PaidBy == "" {
		e.PaidBy = userID
	}
}

// EffectiveAssignee returns AssignedTo if set, otherwise falls back to PaidBy.
func (e *Expense) EffectiveAssignee() string {
	if e.AssignedTo != "" {
		return e.AssignedTo
	}
	return e.PaidBy
}
