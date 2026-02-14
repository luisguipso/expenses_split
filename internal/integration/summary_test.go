package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
)

// setupTwoMemberHousehold creates a household with Alice (salary 5000) and Bob (salary 3000),
// returning their auth info and the household ID.
func setupTwoMemberHousehold(t *testing.T) (alice, bob authUser, hhID string) {
	t.Helper()
	cleanDB(t)

	alice = registerUser(t, "Alice", "alice@test.com", "secret123")
	bob = registerUser(t, "Bob", "bob@test.com", "secret456")
	hhID = createHousehold(t, alice.AccessToken, "Casa")
	joinHousehold(t, bob.AccessToken, hhID, alice.AccessToken)

	// Set salaries: Alice 5000, Bob 3000 → proportions 62.5% / 37.5%
	doJSON(t, http.MethodPut,
		fmt.Sprintf("/households/%s/members/%s/salary", hhID, alice.ID),
		domain.UpdateSalaryInput{SalaryCents: 500000},
		alice.AccessToken, http.StatusNoContent)

	doJSON(t, http.MethodPut,
		fmt.Sprintf("/households/%s/members/%s/salary", hhID, bob.ID),
		domain.UpdateSalaryInput{SalaryCents: 300000},
		alice.AccessToken, http.StatusNoContent)

	return
}

func TestSummary_ProportionalSplit(t *testing.T) {
	alice, _, hhID := setupTwoMemberHousehold(t)

	// Add a shared expense of R$100
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Mercado",
			AmountCents: 10000,
			ExpenseDate: "2024-02-10",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	// Generate summary
	resp := doGet(t, fmt.Sprintf("/households/%s/summary?year=2024&month=2", hhID),
		alice.AccessToken, http.StatusOK)

	var summary map[string]interface{}
	decodeJSON(t, resp, &summary)

	if int64(summary["total_shared_cents"].(float64)) != 10000 {
		t.Errorf("expected total_shared 10000, got %v", summary["total_shared_cents"])
	}

	items := summary["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Alice: 62.5% of 10000 = 6250, Bob: 37.5% of 10000 = 3750
	for _, raw := range items {
		item := raw.(map[string]interface{})
		name := item["user_name"].(string)
		due := int64(item["amount_due_cents"].(float64))
		switch name {
		case "Alice":
			if due != 6250 {
				t.Errorf("Alice due: expected 6250, got %d", due)
			}
		case "Bob":
			if due != 3750 {
				t.Errorf("Bob due: expected 3750, got %d", due)
			}
		}
	}
}

func TestSummary_BalanceFields(t *testing.T) {
	alice, bob, hhID := setupTwoMemberHousehold(t)

	// Alice pays R$100 shared expense
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Mercado",
			AmountCents: 10000,
			ExpenseDate: "2024-02-10",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	// Bob pays R$50 shared fixed bill
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Internet",
			AmountCents: 5000,
			DueDay:      10,
			IsShared:    true,
			PaidBy:      bob.ID,
		}, alice.AccessToken, http.StatusCreated)

	resp := doGet(t, fmt.Sprintf("/households/%s/summary?year=2024&month=2", hhID),
		alice.AccessToken, http.StatusOK)

	var summary map[string]interface{}
	decodeJSON(t, resp, &summary)

	// Total shared: 10000 + 5000 = 15000
	// Alice owes: 62.5% of 15000 = 9375, paid 10000 → balance +625
	// Bob owes: 37.5% of 15000 = 5625, paid 5000 → balance -625
	for _, raw := range summary["items"].([]interface{}) {
		item := raw.(map[string]interface{})
		name := item["user_name"].(string)
		paid := int64(item["total_paid_cents"].(float64))
		balance := int64(item["balance_cents"].(float64))
		switch name {
		case "Alice":
			if paid != 10000 {
				t.Errorf("Alice paid: expected 10000, got %d", paid)
			}
			if balance != 625 {
				t.Errorf("Alice balance: expected +625, got %d", balance)
			}
		case "Bob":
			if paid != 5000 {
				t.Errorf("Bob paid: expected 5000, got %d", paid)
			}
			if balance != -625 {
				t.Errorf("Bob balance: expected -625, got %d", balance)
			}
		}
	}
}

func TestSummary_Settlements(t *testing.T) {
	alice, _, hhID := setupTwoMemberHousehold(t)

	// Alice pays R$200 shared expense → Bob owes Alice
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Aluguel",
			AmountCents: 20000,
			ExpenseDate: "2024-03-01",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	resp := doGet(t, fmt.Sprintf("/households/%s/summary?year=2024&month=3", hhID),
		alice.AccessToken, http.StatusOK)

	var summary map[string]interface{}
	decodeJSON(t, resp, &summary)

	settlements := summary["settlements"].([]interface{})
	if len(settlements) != 1 {
		t.Fatalf("expected 1 settlement, got %d", len(settlements))
	}

	s := settlements[0].(map[string]interface{})
	if s["from_user_name"] != "Bob" || s["to_user_name"] != "Alice" {
		t.Errorf("expected Bob→Alice, got %s→%s", s["from_user_name"], s["to_user_name"])
	}

	// Bob owes 37.5% of 20000 = 7500
	amount := int64(s["amount_cents"].(float64))
	if amount != 7500 {
		t.Errorf("expected settlement 7500, got %d", amount)
	}
}

func TestSummary_DashboardBalance(t *testing.T) {
	alice, bob, hhID := setupTwoMemberHousehold(t)

	// Both pay expenses in current month
	now := fmt.Sprintf("%d-%02d-10", currentYear(), currentMonth())

	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Mercado",
			AmountCents: 8000,
			ExpenseDate: now,
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Padaria",
			AmountCents: 2000,
			ExpenseDate: now,
			IsShared:    true,
		}, bob.AccessToken, http.StatusCreated)

	resp := doGet(t, fmt.Sprintf("/households/%s/dashboard", hhID),
		alice.AccessToken, http.StatusOK)

	var dash map[string]interface{}
	decodeJSON(t, resp, &dash)

	members := dash["member_breakdown"].([]interface{})
	if len(members) != 2 {
		t.Fatalf("expected 2 members in breakdown, got %d", len(members))
	}

	// Total shared: 10000. Alice 62.5%=6250, Bob 37.5%=3750
	// Alice paid 8000, balance = 8000-6250 = +1750
	// Bob paid 2000, balance = 2000-3750 = -1750
	var totalBalance int64
	for _, raw := range members {
		item := raw.(map[string]interface{})
		balance := int64(item["balance_cents"].(float64))
		totalBalance += balance
	}
	if totalBalance != 0 {
		t.Errorf("balances must sum to 0, got %d", totalBalance)
	}
}

func TestSummary_BalancesSumToZero(t *testing.T) {
	alice, bob, hhID := setupTwoMemberHousehold(t)

	// Complex scenario: multiple expenses paid by different users
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Aluguel",
			AmountCents: 200000,
			ExpenseDate: "2024-04-01",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Internet",
			AmountCents: 12000,
			ExpenseDate: "2024-04-05",
			IsShared:    true,
		}, bob.AccessToken, http.StatusCreated)

	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Condomínio",
			AmountCents: 50000,
			DueDay:      10,
			IsShared:    true,
			PaidBy:      alice.ID,
		}, alice.AccessToken, http.StatusCreated)

	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Remédio Alice",
			AmountCents: 8000,
			ExpenseDate: "2024-04-15",
			IsShared:    false,
			AssignedTo:  alice.ID,
		}, alice.AccessToken, http.StatusCreated)

	resp := doGet(t, fmt.Sprintf("/households/%s/summary?year=2024&month=4", hhID),
		alice.AccessToken, http.StatusOK)

	var summary map[string]interface{}
	decodeJSON(t, resp, &summary)

	items := summary["items"].([]interface{})
	var totalBalance int64
	for _, raw := range items {
		item := raw.(map[string]interface{})
		totalBalance += int64(item["balance_cents"].(float64))
	}
	if totalBalance != 0 {
		t.Errorf("balances must sum to 0 (invariant), got %d", totalBalance)
	}

	// Verify settlements net effect
	settlements := summary["settlements"].([]interface{})
	netByName := make(map[string]int64)
	for _, raw := range settlements {
		s := raw.(map[string]interface{})
		amount := int64(s["amount_cents"].(float64))
		netByName[s["from_user_name"].(string)] -= amount
		netByName[s["to_user_name"].(string)] += amount
	}

	for _, raw := range items {
		item := raw.(map[string]interface{})
		name := item["user_name"].(string)
		balance := int64(item["balance_cents"].(float64))
		if balance != 0 && netByName[name] != balance {
			t.Errorf("%s: settlement net %d != balance %d", name, netByName[name], balance)
		}
	}
}

func TestSummaryDetail_Breakdown(t *testing.T) {
	alice, bob, hhID := setupTwoMemberHousehold(t)

	// Create a category
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/categories", hhID),
		domain.CreateCategoryInput{Name: "Moradia", Icon: "🏠"},
		alice.AccessToken, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := cat["id"].(string)

	// Alice pays shared fixed bill R$2000
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Aluguel",
			AmountCents: 200000,
			DueDay:      5,
			IsShared:    true,
			CategoryID:  catID,
			PaidBy:      alice.ID,
		}, alice.AccessToken, http.StatusCreated)

	// Bob pays shared expense R$500
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Mercado",
			AmountCents: 50000,
			ExpenseDate: "2024-06-10",
			IsShared:    true,
		}, bob.AccessToken, http.StatusCreated)

	// Bob personal expense R$35
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Farmácia Bob",
			AmountCents: 3500,
			ExpenseDate: "2024-06-12",
			IsShared:    false,
			AssignedTo:  bob.ID,
		}, bob.AccessToken, http.StatusCreated)

	// Get Alice detail
	resp = doGet(t,
		fmt.Sprintf("/households/%s/summary/detail?year=2024&month=6&user_id=%s", hhID, alice.ID),
		alice.AccessToken, http.StatusOK)

	var detail map[string]interface{}
	decodeJSON(t, resp, &detail)

	if detail["user_name"] != "Alice" {
		t.Errorf("expected Alice, got %v", detail["user_name"])
	}

	items := detail["items"].([]interface{})
	// Alice should see: Aluguel (shared) + Mercado (shared) = 2 items, no personal
	if len(items) != 2 {
		t.Fatalf("Alice: expected 2 items, got %d", len(items))
	}

	for _, raw := range items {
		item := raw.(map[string]interface{})
		desc := item["description"].(string)
		share := int64(item["user_share_cents"].(float64))
		total := int64(item["total_cents"].(float64))

		switch desc {
		case "Aluguel":
			// 62.5% of 200000 = 125000
			if share != 125000 {
				t.Errorf("Aluguel share: expected 125000, got %d", share)
			}
			if total != 200000 {
				t.Errorf("Aluguel total: expected 200000, got %d", total)
			}
			if item["type"] != "fixed_bill" {
				t.Errorf("Aluguel type: expected fixed_bill, got %v", item["type"])
			}
			if item["category_name"] != "Moradia" {
				t.Errorf("Aluguel category: expected Moradia, got %v", item["category_name"])
			}
		case "Mercado":
			// 62.5% of 50000 = 31250
			if share != 31250 {
				t.Errorf("Mercado share: expected 31250, got %d", share)
			}
		}
	}

	// Totals must match summary row
	amountDue := int64(detail["amount_due_cents"].(float64))
	totalShared := int64(detail["total_shared_cents"].(float64))
	totalPersonal := int64(detail["total_personal_cents"].(float64))

	if totalShared != 156250 { // 125000 + 31250
		t.Errorf("total shared: expected 156250, got %d", totalShared)
	}
	if totalPersonal != 0 {
		t.Errorf("total personal: expected 0, got %d", totalPersonal)
	}
	if amountDue != 156250 {
		t.Errorf("amount due: expected 156250, got %d", amountDue)
	}

	// Now get Bob detail — should have 3 items (2 shared + 1 personal)
	resp = doGet(t,
		fmt.Sprintf("/households/%s/summary/detail?year=2024&month=6&user_id=%s", hhID, bob.ID),
		alice.AccessToken, http.StatusOK)

	var bobDetail map[string]interface{}
	decodeJSON(t, resp, &bobDetail)

	bobItems := bobDetail["items"].([]interface{})
	if len(bobItems) != 3 {
		t.Fatalf("Bob: expected 3 items, got %d", len(bobItems))
	}

	bobPersonal := int64(bobDetail["total_personal_cents"].(float64))
	if bobPersonal != 3500 {
		t.Errorf("Bob personal: expected 3500, got %d", bobPersonal)
	}
}
