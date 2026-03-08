package domain

import (
	"testing"
)

func TestExpense_Validate(t *testing.T) {
	tests := []struct {
		name    string
		amount  int64
		wantErr bool
	}{
		{"valid", 5000, false},
		{"zero", 0, true},
		{"negative", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Expense{AmountCents: tt.amount}
			err := e.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpense_SetDefaults(t *testing.T) {
	e := &Expense{}
	e.SetDefaults("user-1")
	if e.PaidBy != "user-1" {
		t.Errorf("PaidBy = %q, want %q", e.PaidBy, "user-1")
	}
	if e.ExpenseDate == "" {
		t.Error("ExpenseDate should be set to today")
	}
}

func TestExpense_SetDefaults_PreservesExisting(t *testing.T) {
	e := &Expense{ExpenseDate: "2026-01-01", PaidBy: "existing"}
	e.SetDefaults("user-1")
	if e.PaidBy != "existing" {
		t.Errorf("PaidBy = %q, want %q", e.PaidBy, "existing")
	}
	if e.ExpenseDate != "2026-01-01" {
		t.Errorf("ExpenseDate = %q, want %q", e.ExpenseDate, "2026-01-01")
	}
}

func TestExpense_EffectiveAssignee(t *testing.T) {
	tests := []struct {
		name       string
		assignedTo string
		paidBy     string
		want       string
	}{
		{"assigned returns assigned", "user-2", "user-1", "user-2"},
		{"empty falls back to paid by", "", "user-1", "user-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Expense{AssignedTo: tt.assignedTo, PaidBy: tt.paidBy}
			if got := e.EffectiveAssignee(); got != tt.want {
				t.Errorf("EffectiveAssignee() = %q, want %q", got, tt.want)
			}
		})
	}
}
