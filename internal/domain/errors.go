package domain

import "errors"

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrEmailExists               = errors.New("email already exists")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrInvalidToken              = errors.New("invalid token")
	ErrHouseholdNotFound         = errors.New("household not found")
	ErrAlreadyMember             = errors.New("user is already a member")
	ErrNotMember                 = errors.New("user is not a member")
	ErrInvalidInviteCode         = errors.New("invalid invite code")
	ErrForbidden                 = errors.New("forbidden")
	ErrCategoryNotFound          = errors.New("category not found")
	ErrCategoryExists            = errors.New("category already exists")
	ErrFixedBillNotFound         = errors.New("fixed bill not found")
	ErrExpenseNotFound           = errors.New("expense not found")
	ErrFixedBillSnapshotNotFound = errors.New("fixed bill snapshot not found")
	ErrSummaryNotFound           = errors.New("summary not found")
	ErrNoMembersWithSalary       = errors.New("no members with salary configured")
	ErrEmailNotVerified          = errors.New("email not verified")
	ErrInvalidVerificationCode   = errors.New("invalid verification code")
	ErrVerificationExpired       = errors.New("verification code expired")
	ErrAlreadyVerified           = errors.New("email already verified")
)
