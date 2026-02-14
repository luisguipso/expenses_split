package domain

import "time"

type Household struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type HouseholdMember struct {
	HouseholdID string    `json:"household_id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	SalaryCents int64     `json:"salary_cents"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}
