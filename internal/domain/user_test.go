package domain

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestUser_SetPassword(t *testing.T) {
	u := &User{}
	if err := u.SetPassword("secret123"); err != nil {
		t.Fatalf("SetPassword failed: %v", err)
	}
	if u.PasswordHash == "" {
		t.Fatal("expected PasswordHash to be set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("secret123")); err != nil {
		t.Error("hash does not match plaintext")
	}
}

func TestUser_VerifyPassword(t *testing.T) {
	u := &User{}
	_ = u.SetPassword("correct")

	tests := []struct {
		name      string
		password  string
		wantErr   error
	}{
		{"correct password", "correct", nil},
		{"wrong password", "wrong", ErrInvalidCredentials},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.VerifyPassword(tt.password)
			if err != tt.wantErr {
				t.Errorf("VerifyPassword(%q) = %v, want %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestUser_RequireVerifiedEmail(t *testing.T) {
	tests := []struct {
		name     string
		verified bool
		wantErr  error
	}{
		{"verified", true, nil},
		{"not verified", false, ErrEmailNotVerified},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{EmailVerified: tt.verified}
			if err := u.RequireVerifiedEmail(); err != tt.wantErr {
				t.Errorf("RequireVerifiedEmail() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_IsSamePassword(t *testing.T) {
	u := &User{}
	_ = u.SetPassword("mypassword")

	if !u.IsSamePassword("mypassword") {
		t.Error("expected IsSamePassword to return true for same password")
	}
	if u.IsSamePassword("different") {
		t.Error("expected IsSamePassword to return false for different password")
	}
}
