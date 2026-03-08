package domain

import "testing"

func TestMonthlySummaryItem_Accumulators(t *testing.T) {
	item := &MonthlySummaryItem{}
	item.AddSharedExpense(10000)
	item.AddSharedExpense(5000)
	item.AddPersonalExpense(3000)
	item.AddPaidAmount(20000)

	if item.TotalSharedCents != 15000 {
		t.Errorf("TotalSharedCents = %d, want 15000", item.TotalSharedCents)
	}
	if item.TotalPersonalCents != 3000 {
		t.Errorf("TotalPersonalCents = %d, want 3000", item.TotalPersonalCents)
	}
	if item.TotalPaidCents != 20000 {
		t.Errorf("TotalPaidCents = %d, want 20000", item.TotalPaidCents)
	}
}

func TestMonthlySummaryItem_CalculateBalance(t *testing.T) {
	tests := []struct {
		name           string
		shared         int64
		personal       int64
		paid           int64
		wantDue        int64
		wantBalance    int64
	}{
		{"overpaid", 10000, 5000, 20000, 15000, 5000},
		{"underpaid", 10000, 5000, 10000, 15000, -5000},
		{"exact", 10000, 5000, 15000, 15000, 0},
		{"zero", 0, 0, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &MonthlySummaryItem{
				TotalSharedCents:   tt.shared,
				TotalPersonalCents: tt.personal,
				TotalPaidCents:     tt.paid,
			}
			item.CalculateBalance()
			if item.AmountDueCents != tt.wantDue {
				t.Errorf("AmountDueCents = %d, want %d", item.AmountDueCents, tt.wantDue)
			}
			if item.BalanceCents != tt.wantBalance {
				t.Errorf("BalanceCents = %d, want %d", item.BalanceCents, tt.wantBalance)
			}
		})
	}
}
