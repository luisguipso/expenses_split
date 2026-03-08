package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type EmailVerification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Code      string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

func NewEmailVerification(userID, email string, expiration time.Duration) *EmailVerification {
	return &EmailVerification{
		UserID:    userID,
		Email:     email,
		Code:      generateVerificationCode(),
		ExpiresAt: time.Now().Add(expiration),
	}
}

func (v *EmailVerification) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

func (v *EmailVerification) MatchesCode(code string) error {
	if v.IsExpired() {
		return ErrVerificationExpired
	}
	if v.Code != code {
		return ErrInvalidVerificationCode
	}
	return nil
}

func generateVerificationCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}
