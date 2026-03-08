package domain

import (
	"testing"
	"time"
)

func TestCategory_SetDefaultIcon(t *testing.T) {
	tests := []struct {
		name     string
		icon     string
		wantIcon string
	}{
		{"empty icon gets default", "", "📦"},
		{"existing icon preserved", "🏠", "🏠"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Category{Icon: tt.icon}
			c.SetDefaultIcon()
			if c.Icon != tt.wantIcon {
				t.Errorf("Icon = %q, want %q", c.Icon, tt.wantIcon)
			}
		})
	}
}

func TestFixedBill_Validate(t *testing.T) {
	tests := []struct {
		name    string
		amount  int64
		dueDay  int
		wantErr bool
	}{
		{"valid", 10000, 15, false},
		{"zero amount", 0, 15, true},
		{"negative amount", -100, 15, true},
		{"due day 0", 10000, 0, true},
		{"due day 32", 10000, 32, true},
		{"due day 1", 10000, 1, false},
		{"due day 31", 10000, 31, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &FixedBill{AmountCents: tt.amount, DueDay: tt.dueDay}
			err := b.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFixedBill_SetDefaultPaidBy(t *testing.T) {
	b := &FixedBill{}
	b.SetDefaultPaidBy("user-1")
	if b.PaidBy != "user-1" {
		t.Errorf("PaidBy = %q, want %q", b.PaidBy, "user-1")
	}

	b.SetDefaultPaidBy("user-2")
	if b.PaidBy != "user-1" {
		t.Error("SetDefaultPaidBy should not overwrite existing PaidBy")
	}
}

func TestFixedBill_HasDueDatePassed(t *testing.T) {
	b := &FixedBill{DueDay: 15}
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	if !b.HasDueDatePassed(2026, 3, now) {
		t.Error("expected due date to have passed")
	}
	if b.HasDueDatePassed(2026, 3, time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)) {
		t.Error("expected due date to not have passed")
	}
}
