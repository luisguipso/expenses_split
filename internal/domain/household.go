package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Household struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func NewHousehold(name string) (*Household, error) {
	h := &Household{Name: name}
	if err := h.GenerateInviteCode(); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Household) GenerateInviteCode() error {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("generate invite code: %w", err)
	}
	h.InviteCode = hex.EncodeToString(b)
	return nil
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

func (m *HouseholdMember) IsAdmin() bool {
	return m.Role == "admin"
}

// CanManage returns true if this member can manage the target member
// (update salary, remove). Admins can manage anyone; non-admins can only
// manage themselves.
func (m *HouseholdMember) CanManage(targetUserID string) bool {
	return m.IsAdmin() || m.UserID == targetUserID
}

func (m *HouseholdMember) ValidateSalary() error {
	if m.SalaryCents < 0 {
		return fmt.Errorf("salary must be non-negative, got %d", m.SalaryCents)
	}
	return nil
}
