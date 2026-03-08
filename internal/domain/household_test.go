package domain

import (
	"testing"
)

func TestNewHousehold(t *testing.T) {
	h, err := NewHousehold("My House")
	if err != nil {
		t.Fatalf("NewHousehold failed: %v", err)
	}
	if h.Name != "My House" {
		t.Errorf("Name = %q, want %q", h.Name, "My House")
	}
	if h.InviteCode == "" {
		t.Error("InviteCode should not be empty")
	}
	if len(h.InviteCode) != 16 {
		t.Errorf("InviteCode length = %d, want 16 hex chars", len(h.InviteCode))
	}
}

func TestHousehold_GenerateInviteCode_Unique(t *testing.T) {
	h, _ := NewHousehold("Test")
	first := h.InviteCode
	_ = h.GenerateInviteCode()
	if h.InviteCode == first {
		t.Error("GenerateInviteCode should produce a new code")
	}
}

func TestHouseholdMember_IsAdmin(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{"admin", true},
		{"member", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			m := &HouseholdMember{Role: tt.role}
			if got := m.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHouseholdMember_CanManage(t *testing.T) {
	tests := []struct {
		name         string
		role         string
		userID       string
		targetUserID string
		want         bool
	}{
		{"admin manages anyone", "admin", "user-1", "user-2", true},
		{"admin manages self", "admin", "user-1", "user-1", true},
		{"member manages self", "member", "user-1", "user-1", true},
		{"member cannot manage other", "member", "user-1", "user-2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HouseholdMember{UserID: tt.userID, Role: tt.role}
			if got := m.CanManage(tt.targetUserID); got != tt.want {
				t.Errorf("CanManage(%q) = %v, want %v", tt.targetUserID, got, tt.want)
			}
		})
	}
}

func TestHouseholdMember_ValidateSalary(t *testing.T) {
	tests := []struct {
		name    string
		salary  int64
		wantErr bool
	}{
		{"positive salary", 500000, false},
		{"zero salary", 0, false},
		{"negative salary", -100, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HouseholdMember{SalaryCents: tt.salary}
			err := m.ValidateSalary()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSalary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
