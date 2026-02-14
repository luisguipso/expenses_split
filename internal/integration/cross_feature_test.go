package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
)

func TestCrossFeature_FullWorkflow(t *testing.T) {
	cleanDB(t)

	// 1. Register two users
	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	bob := registerUser(t, "Bob", "bob@test.com", "secret456")

	// 2. Alice creates household
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	// 3. Bob joins
	joinHousehold(t, bob.AccessToken, hhID, alice.AccessToken)

	// 4. Set salaries: Alice R$5000, Bob R$3000
	doJSON(t, http.MethodPut,
		fmt.Sprintf("/households/%s/members/%s/salary", hhID, alice.ID),
		domain.UpdateSalaryInput{SalaryCents: 500000},
		alice.AccessToken, http.StatusNoContent)
	doJSON(t, http.MethodPut,
		fmt.Sprintf("/households/%s/members/%s/salary", hhID, bob.ID),
		domain.UpdateSalaryInput{SalaryCents: 300000},
		alice.AccessToken, http.StatusNoContent)

	// 5. Create categories
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/categories", hhID),
		domain.CreateCategoryInput{Name: "Moradia", Icon: "🏠"},
		alice.AccessToken, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := cat["id"].(string)

	// 6. Add fixed bills
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Aluguel",
			AmountCents: 200000,
			DueDay:      5,
			IsShared:    true,
			CategoryID:  catID,
			PaidBy:      alice.ID,
		}, alice.AccessToken, http.StatusCreated)

	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Internet",
			AmountCents: 12000,
			DueDay:      10,
			IsShared:    true,
			PaidBy:      bob.ID,
		}, alice.AccessToken, http.StatusCreated)

	// 7. Add expenses
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Supermercado",
			AmountCents: 45000,
			ExpenseDate: "2024-03-08",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Farmácia Bob",
			AmountCents: 3500,
			ExpenseDate: "2024-03-12",
			IsShared:    false,
			AssignedTo:  bob.ID,
		}, bob.AccessToken, http.StatusCreated)

	// 8. Generate summary
	resp = doGet(t, fmt.Sprintf("/households/%s/summary?year=2024&month=3", hhID),
		alice.AccessToken, http.StatusOK)

	var summary map[string]interface{}
	decodeJSON(t, resp, &summary)

	// Verify totals
	// Shared: aluguel(200000) + internet(12000) + mercado(45000) = 257000
	// Personal: farmácia(3500) for Bob
	totalShared := int64(summary["total_shared_cents"].(float64))
	if totalShared != 257000 {
		t.Errorf("total shared: expected 257000, got %d", totalShared)
	}

	items := summary["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Verify each member
	for _, raw := range items {
		item := raw.(map[string]interface{})
		name := item["user_name"].(string)
		due := int64(item["amount_due_cents"].(float64))
		paid := int64(item["total_paid_cents"].(float64))
		balance := int64(item["balance_cents"].(float64))

		switch name {
		case "Alice":
			// Alice shared: 62.5% of 257000 = 160625
			// Alice personal: 0
			// Alice paid: aluguel(200000) + mercado(45000) = 245000
			// Balance: 245000 - 160625 = +84375
			if due != 160625 {
				t.Errorf("Alice due: expected 160625, got %d", due)
			}
			if paid != 245000 {
				t.Errorf("Alice paid: expected 245000, got %d", paid)
			}
			if balance != 84375 {
				t.Errorf("Alice balance: expected +84375, got %d", balance)
			}
		case "Bob":
			// Bob shared: 37.5% of 257000 = 96375
			// Bob personal: 3500
			// Bob due: 96375 + 3500 = 99875
			// Bob paid: internet(12000) + farmácia(3500) = 15500
			// Balance: 15500 - 99875 = -84375
			if due != 99875 {
				t.Errorf("Bob due: expected 99875, got %d", due)
			}
			if paid != 15500 {
				t.Errorf("Bob paid: expected 15500, got %d", paid)
			}
			if balance != -84375 {
				t.Errorf("Bob balance: expected -84375, got %d", balance)
			}
		}
	}

	// Verify settlements: Bob → Alice R$843.75
	settlements := summary["settlements"].([]interface{})
	if len(settlements) != 1 {
		t.Fatalf("expected 1 settlement, got %d", len(settlements))
	}

	s := settlements[0].(map[string]interface{})
	if s["from_user_name"] != "Bob" || s["to_user_name"] != "Alice" {
		t.Errorf("expected Bob→Alice, got %s→%s", s["from_user_name"], s["to_user_name"])
	}
	if int64(s["amount_cents"].(float64)) != 84375 {
		t.Errorf("settlement: expected 84375, got %v", s["amount_cents"])
	}
}

func TestCrossFeature_TwoUsersPayingDifferentBills(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	bob := registerUser(t, "Bob", "bob@test.com", "secret456")
	hhID := createHousehold(t, alice.AccessToken, "Casa")
	joinHousehold(t, bob.AccessToken, hhID, alice.AccessToken)

	// Equal salaries → 50/50 split
	doJSON(t, http.MethodPut,
		fmt.Sprintf("/households/%s/members/%s/salary", hhID, alice.ID),
		domain.UpdateSalaryInput{SalaryCents: 400000},
		alice.AccessToken, http.StatusNoContent)
	doJSON(t, http.MethodPut,
		fmt.Sprintf("/households/%s/members/%s/salary", hhID, bob.ID),
		domain.UpdateSalaryInput{SalaryCents: 400000},
		alice.AccessToken, http.StatusNoContent)

	// Alice pays R$1000 shared
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Aluguel",
			AmountCents: 100000,
			ExpenseDate: "2024-05-01",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	// Bob pays R$600 shared
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Mercado",
			AmountCents: 60000,
			ExpenseDate: "2024-05-10",
			IsShared:    true,
		}, bob.AccessToken, http.StatusCreated)

	resp := doGet(t, fmt.Sprintf("/households/%s/summary?year=2024&month=5", hhID),
		alice.AccessToken, http.StatusOK)

	var summary map[string]interface{}
	decodeJSON(t, resp, &summary)

	// Total shared: 160000. Each owes 80000.
	// Alice paid 100000, balance = +20000
	// Bob paid 60000, balance = -20000
	// Settlement: Bob → Alice R$200
	items := summary["items"].([]interface{})
	var totalBalance int64
	for _, raw := range items {
		item := raw.(map[string]interface{})
		name := item["user_name"].(string)
		paid := int64(item["total_paid_cents"].(float64))
		balance := int64(item["balance_cents"].(float64))
		totalBalance += balance

		switch name {
		case "Alice":
			if paid != 100000 {
				t.Errorf("Alice paid: expected 100000, got %d", paid)
			}
			if balance != 20000 {
				t.Errorf("Alice balance: expected +20000, got %d", balance)
			}
		case "Bob":
			if paid != 60000 {
				t.Errorf("Bob paid: expected 60000, got %d", paid)
			}
			if balance != -20000 {
				t.Errorf("Bob balance: expected -20000, got %d", balance)
			}
		}
	}

	if totalBalance != 0 {
		t.Errorf("balances must sum to 0, got %d", totalBalance)
	}

	settlements := summary["settlements"].([]interface{})
	if len(settlements) != 1 {
		t.Fatalf("expected 1 settlement, got %d", len(settlements))
	}

	s := settlements[0].(map[string]interface{})
	if s["from_user_name"] != "Bob" || s["to_user_name"] != "Alice" {
		t.Errorf("expected Bob→Alice, got %s→%s", s["from_user_name"], s["to_user_name"])
	}
	if int64(s["amount_cents"].(float64)) != 20000 {
		t.Errorf("settlement: expected 20000, got %v", s["amount_cents"])
	}
}
