package domain

import (
	"crypto/rand"
	"fmt"
	"time"
)

type PasswordReset struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

func NewPasswordReset(userID, email string, expiration time.Duration) (*PasswordReset, error) {
	token, err := generateResetToken()
	if err != nil {
		return nil, err
	}
	return &PasswordReset{
		UserID:    userID,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(expiration),
	}, nil
}

func (r *PasswordReset) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}
