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

type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	FindByID(ctx context.Context, id string) (*Category, error)
	ListByHousehold(ctx context.Context, householdID string) ([]Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id string) error
}

type CategoryService interface {
	Create(ctx context.Context, input CreateCategoryInput, householdID, userID string) (*Category, error)
	List(ctx context.Context, householdID, userID string) ([]Category, error)
	Update(ctx context.Context, id string, input UpdateCategoryInput, userID string) (*Category, error)
	Delete(ctx context.Context, id, userID string) error
}

type FixedBillRepository interface {
	Create(ctx context.Context, bill *FixedBill) error
	FindByID(ctx context.Context, id string) (*FixedBill, error)
	ListByHousehold(ctx context.Context, householdID string) ([]FixedBill, error)
	Update(ctx context.Context, bill *FixedBill) error
	Delete(ctx context.Context, id string) error
}

type FixedBillService interface {
	Create(ctx context.Context, input CreateFixedBillInput, householdID, userID string) (*FixedBill, error)
	List(ctx context.Context, householdID, userID string) ([]FixedBill, error)
	Update(ctx context.Context, id string, input UpdateFixedBillInput, userID string) (*FixedBill, error)
	Delete(ctx context.Context, id, userID string) error
}

type ExpenseRepository interface {
	Create(ctx context.Context, expense *Expense) error
	FindByID(ctx context.Context, id string) (*Expense, error)
	ListByHousehold(ctx context.Context, householdID string, filter ExpenseFilter) ([]Expense, error)
	Update(ctx context.Context, expense *Expense) error
	Delete(ctx context.Context, id string) error
}

type ExpenseService interface {
	Create(ctx context.Context, input CreateExpenseInput, householdID, userID string) (*Expense, error)
	List(ctx context.Context, householdID, userID string, filter ExpenseFilter) ([]Expense, error)
	Update(ctx context.Context, id string, input UpdateExpenseInput, userID string) (*Expense, error)
	Delete(ctx context.Context, id, userID string) error
}

type SummaryRepository interface {
	Upsert(ctx context.Context, summary *MonthlySummary) error
	FindByMonth(ctx context.Context, householdID string, year, month int) (*MonthlySummary, error)
}

type SummaryService interface {
	Generate(ctx context.Context, householdID string, year, month int, userID string) (*SummaryResponse, error)
	GetDashboard(ctx context.Context, householdID, userID string) (*DashboardResponse, error)
}
