package domain

import "time"

type MonthlySummary struct {
	ID          string               `json:"id"`
	HouseholdID string               `json:"household_id"`
	Year        int                  `json:"year"`
	Month       int                  `json:"month"`
	GeneratedAt time.Time            `json:"generated_at"`
	Items       []MonthlySummaryItem `json:"items"`
}

type MonthlySummaryItem struct {
	ID                 string `json:"id"`
	SummaryID          string `json:"summary_id"`
	UserID             string `json:"user_id"`
	UserName           string `json:"user_name"`
	TotalSharedCents   int64  `json:"total_shared_cents"`
	TotalPersonalCents int64  `json:"total_personal_cents"`
	AmountDueCents     int64  `json:"amount_due_cents"`
	TotalPaidCents     int64  `json:"total_paid_cents"`
	BalanceCents       int64  `json:"balance_cents"`
}

func (item *MonthlySummaryItem) AddSharedExpense(amountCents int64) {
	item.TotalSharedCents += amountCents
}

func (item *MonthlySummaryItem) AddPersonalExpense(amountCents int64) {
	item.TotalPersonalCents += amountCents
}

func (item *MonthlySummaryItem) AddPaidAmount(amountCents int64) {
	item.TotalPaidCents += amountCents
}

// CalculateBalance derives BalanceCents and AmountDueCents from the
// accumulated shared/personal/paid values.
func (item *MonthlySummaryItem) CalculateBalance() {
	item.AmountDueCents = item.TotalSharedCents + item.TotalPersonalCents
	item.BalanceCents = item.TotalPaidCents - item.AmountDueCents
}
