package domain

import "time"

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
