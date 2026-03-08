package domain

import "time"

type FixedBillSnapshot struct {
	ID          string    `json:"id"`
	FixedBillID string    `json:"fixed_bill_id"`
	HouseholdID string    `json:"household_id"`
	Year        int       `json:"year"`
	Month       int       `json:"month"`
	CategoryID  string    `json:"category_id,omitempty"`
	Description string    `json:"description"`
	AmountCents int64     `json:"amount_cents"`
	DueDay      int       `json:"due_day"`
	IsShared    bool      `json:"is_shared"`
	PaidBy      string    `json:"paid_by"`
	AssignedTo  string    `json:"assigned_to,omitempty"`
	FrozenAt    time.Time `json:"frozen_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *FixedBillSnapshot) SetDefaultPaidBy(userID string) {
	if s.PaidBy == "" {
		s.PaidBy = userID
	}
}
