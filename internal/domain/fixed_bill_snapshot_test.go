package domain

import "testing"

func TestFixedBillSnapshot_SetDefaultPaidBy(t *testing.T) {
	s := &FixedBillSnapshot{}
	s.SetDefaultPaidBy("user-1")
	if s.PaidBy != "user-1" {
		t.Errorf("PaidBy = %q, want %q", s.PaidBy, "user-1")
	}

	s.SetDefaultPaidBy("user-2")
	if s.PaidBy != "user-1" {
		t.Error("SetDefaultPaidBy should not overwrite existing PaidBy")
	}
}
