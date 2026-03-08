package domain

import (
	"testing"
	"time"
)

func TestNewPasswordReset(t *testing.T) {
	r, err := NewPasswordReset("user-1", "test@example.com", 30*time.Minute)
	if err != nil {
		t.Fatalf("NewPasswordReset failed: %v", err)
	}
	if r.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", r.UserID, "user-1")
	}
	if r.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", r.Email, "test@example.com")
	}
	if r.Token == "" {
		t.Error("Token should not be empty")
	}
	if len(r.Token) != 64 {
		t.Errorf("Token length = %d, want 64 hex chars", len(r.Token))
	}
	if r.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestPasswordReset_IsExpired(t *testing.T) {
	tests := []struct {
		name    string
		expires time.Time
		want    bool
	}{
		{"not expired", time.Now().Add(10 * time.Minute), false},
		{"expired", time.Now().Add(-1 * time.Minute), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &PasswordReset{ExpiresAt: tt.expires}
			if got := r.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPasswordReset_UniqueTokens(t *testing.T) {
	r1, _ := NewPasswordReset("u1", "a@b.com", time.Minute)
	r2, _ := NewPasswordReset("u1", "a@b.com", time.Minute)
	if r1.Token == r2.Token {
		t.Error("expected unique tokens for separate calls")
	}
}
