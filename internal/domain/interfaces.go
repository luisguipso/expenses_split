package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type TokenService interface {
	Generate(userID, email string) (*TokenPair, error)
	Validate(tokenStr string) (*TokenClaims, error)
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*User, *TokenPair, error)
	Login(ctx context.Context, input LoginInput) (*User, *TokenPair, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
}

type HouseholdRepository interface {
	Create(ctx context.Context, household *Household, adminUserID string) error
	FindByID(ctx context.Context, id string) (*Household, error)
	FindByInviteCode(ctx context.Context, code string) (*Household, error)
	ListByUser(ctx context.Context, userID string) ([]Household, error)
	Update(ctx context.Context, household *Household) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, householdID, userID, role string) error
	RemoveMember(ctx context.Context, householdID, userID string) error
	UpdateMemberSalary(ctx context.Context, householdID, userID string, salaryCents int64) error
	ListMembers(ctx context.Context, householdID string) ([]HouseholdMember, error)
	GetMember(ctx context.Context, householdID, userID string) (*HouseholdMember, error)
}

type HouseholdService interface {
	Create(ctx context.Context, input CreateHouseholdInput, userID string) (*Household, error)
	GetByID(ctx context.Context, id, userID string) (*Household, error)
	ListByUser(ctx context.Context, userID string) ([]Household, error)
	Update(ctx context.Context, id string, input UpdateHouseholdInput, userID string) (*Household, error)
	Delete(ctx context.Context, id, userID string) error
	Join(ctx context.Context, inviteCode, userID string) (*Household, error)
	ListMembers(ctx context.Context, householdID, userID string) ([]HouseholdMember, error)
	UpdateMemberSalary(ctx context.Context, householdID, memberID string, salaryCents int64, userID string) error
	RemoveMember(ctx context.Context, householdID, memberID, userID string) error
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}
