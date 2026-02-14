package domain

import "time"

type Category struct {
	ID          string    `json:"id"`
	HouseholdID string    `json:"household_id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	CreatedAt   time.Time `json:"created_at"`
}

type FixedBill struct {
	ID           string    `json:"id"`
	HouseholdID  string    `json:"household_id"`
	CategoryID   string    `json:"category_id,omitempty"`
	CategoryName string    `json:"category_name,omitempty"`
	Description  string    `json:"description"`
	AmountCents  int64     `json:"amount_cents"`
	DueDay       int       `json:"due_day"`
	IsShared     bool      `json:"is_shared"`
	PaidBy       string    `json:"paid_by"`
	PaidByName   string    `json:"paid_by_name,omitempty"`
	AssignedTo   string    `json:"assigned_to,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
