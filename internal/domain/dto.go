package domain

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type TokenClaims struct {
	UserID string
	Email  string
}

type UserResponse struct {
	ID    interface{} `json:"id"`
	Name  string      `json:"name"`
	Email string      `json:"email"`
}

type AuthResponse struct {
	User   UserResponse `json:"user"`
	Tokens *TokenPair   `json:"tokens"`
}

type TokenResponse struct {
	Tokens *TokenPair `json:"tokens"`
}

type MeResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// Household DTOs

type CreateHouseholdInput struct {
	Name string `json:"name"`
}

type UpdateHouseholdInput struct {
	Name string `json:"name"`
}

type JoinHouseholdInput struct {
	InviteCode string `json:"invite_code"`
}

type UpdateSalaryInput struct {
	SalaryCents int64 `json:"salary_cents"`
}

type HouseholdResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type MemberResponse struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	UserEmail   string `json:"user_email"`
	SalaryCents int64  `json:"salary_cents"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joined_at"`
}

// Category DTOs

type CreateCategoryInput struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type UpdateCategoryInput struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type CategoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// Fixed Bill DTOs

type CreateFixedBillInput struct {
	CategoryID  string `json:"category_id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	DueDay      int    `json:"due_day"`
	IsShared    bool   `json:"is_shared"`
	AssignedTo  string `json:"assigned_to"`
}

type UpdateFixedBillInput struct {
	CategoryID  string `json:"category_id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	DueDay      int    `json:"due_day"`
	IsShared    bool   `json:"is_shared"`
	AssignedTo  string `json:"assigned_to"`
	IsActive    bool   `json:"is_active"`
}

type FixedBillResponse struct {
	ID           string `json:"id"`
	CategoryID   string `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
	Description  string `json:"description"`
	AmountCents  int64  `json:"amount_cents"`
	DueDay       int    `json:"due_day"`
	IsShared     bool   `json:"is_shared"`
	AssignedTo   string `json:"assigned_to,omitempty"`
	IsActive     bool   `json:"is_active"`
}

// Expense DTOs

type CreateExpenseInput struct {
	CategoryID  string `json:"category_id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	ExpenseDate string `json:"expense_date"`
	IsShared    bool   `json:"is_shared"`
	AssignedTo  string `json:"assigned_to"`
}

type UpdateExpenseInput struct {
	CategoryID  string `json:"category_id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	ExpenseDate string `json:"expense_date"`
	IsShared    bool   `json:"is_shared"`
	AssignedTo  string `json:"assigned_to"`
}

type ExpenseFilter struct {
	Month      int    `query:"month"`
	Year       int    `query:"year"`
	CategoryID string `query:"category_id"`
	UserID     string `query:"user_id"`
}

type ExpenseResponse struct {
	ID           string `json:"id"`
	CategoryID   string `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
	Description  string `json:"description"`
	AmountCents  int64  `json:"amount_cents"`
	ExpenseDate  string `json:"expense_date"`
	IsShared     bool   `json:"is_shared"`
	PaidBy       string `json:"paid_by"`
	PaidByName   string `json:"paid_by_name,omitempty"`
	AssignedTo   string `json:"assigned_to,omitempty"`
}

// Summary DTOs

type SummaryItemResponse struct {
	UserID             string  `json:"user_id"`
	UserName           string  `json:"user_name"`
	SalaryCents        int64   `json:"salary_cents"`
	Proportion         float64 `json:"proportion"`
	TotalSharedCents   int64   `json:"total_shared_cents"`
	TotalPersonalCents int64   `json:"total_personal_cents"`
	AmountDueCents     int64   `json:"amount_due_cents"`
}

type SummaryResponse struct {
	ID               string                `json:"id"`
	HouseholdID      string                `json:"household_id"`
	Year             int                   `json:"year"`
	Month            int                   `json:"month"`
	TotalSharedCents int64                 `json:"total_shared_cents"`
	TotalAllCents    int64                 `json:"total_all_cents"`
	GeneratedAt      string                `json:"generated_at"`
	Items            []SummaryItemResponse `json:"items"`
}

type DashboardResponse struct {
	HouseholdName      string                `json:"household_name"`
	Year               int                   `json:"year"`
	Month              int                   `json:"month"`
	TotalExpenses      int64                 `json:"total_expenses"`
	TotalFixedBills    int64                 `json:"total_fixed_bills"`
	TotalShared        int64                 `json:"total_shared"`
	TotalPersonal      int64                 `json:"total_personal"`
	ExpenseCount       int                   `json:"expense_count"`
	FixedBillCount     int                   `json:"fixed_bill_count"`
	MemberBreakdown    []SummaryItemResponse `json:"member_breakdown"`
}
