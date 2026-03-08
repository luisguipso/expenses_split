package domain

import (
	"testing"
	"time"
)

func TestNewEmailVerification(t *testing.T) {
	v := NewEmailVerification("user-1", "test@example.com", 15*time.Minute)

	if v.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", v.UserID, "user-1")
	}
	if v.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", v.Email, "test@example.com")
	}
	if len(v.Code) != 6 {
		t.Errorf("Code length = %d, want 6", len(v.Code))
	}
	if v.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestEmailVerification_IsExpired(t *testing.T) {
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
			v := &EmailVerification{ExpiresAt: tt.expires}
			if got := v.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailVerification_MatchesCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		stored  string
		expires time.Time
		wantErr error
	}{
		{"valid code", "123456", "123456", time.Now().Add(10 * time.Minute), nil},
		{"wrong code", "000000", "123456", time.Now().Add(10 * time.Minute), ErrInvalidVerificationCode},
		{"expired code", "123456", "123456", time.Now().Add(-1 * time.Minute), ErrVerificationExpired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &EmailVerification{Code: tt.stored, ExpiresAt: tt.expires}
			if err := v.MatchesCode(tt.code); err != tt.wantErr {
				t.Errorf("MatchesCode(%q) = %v, want %v", tt.code, err, tt.wantErr)
			}
		})
	}
}
