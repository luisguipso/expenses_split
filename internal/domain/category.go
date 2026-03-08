package domain

import (
	"fmt"
	"time"
)

const defaultCategoryIcon = "📦"

type Category struct {
	ID          string    `json:"id"`
	HouseholdID string    `json:"household_id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	CreatedAt   time.Time `json:"created_at"`
}

func (c *Category) SetDefaultIcon() {
	if c.Icon == "" {
		c.Icon = defaultCategoryIcon
	}
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

func (b *FixedBill) Validate() error {
	if b.AmountCents <= 0 {
		return fmt.Errorf("amount must be positive, got %d", b.AmountCents)
	}
	if b.DueDay < 1 || b.DueDay > 31 {
		return fmt.Errorf("due day must be between 1 and 31, got %d", b.DueDay)
	}
	return nil
}

func (b *FixedBill) SetDefaultPaidBy(userID string) {
	if b.PaidBy == "" {
		b.PaidBy = userID
	}
}

// HasDueDatePassed checks if the bill's due day has passed for the given
// year/month. If DueDay exceeds the number of days in the month, it's
// treated as the last day of that month.
func (b *FixedBill) HasDueDatePassed(year, month int, now time.Time) bool {
	lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, now.Location()).Day()
	effectiveDueDay := b.DueDay
	if effectiveDueDay > lastDay {
		effectiveDueDay = lastDay
	}
	dueDate := time.Date(year, time.Month(month), effectiveDueDay, 23, 59, 59, 0, now.Location())
	return now.After(dueDate)
}
