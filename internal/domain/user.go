package domain

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (u *User) SetPassword(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) VerifyPassword(plaintext string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plaintext)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (u *User) RequireVerifiedEmail() error {
	if !u.EmailVerified {
		return ErrEmailNotVerified
	}
	return nil
}

func (u *User) IsSamePassword(plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plaintext)) == nil
}
