package domain

import (
	"context"
	"io"
)

// ParsedExpense represents a single transaction extracted from a credit card bill.
type ParsedExpense struct {
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	Date        string `json:"date"` // YYYY-MM-DD
}

// ParsedBill represents the result of parsing a credit card bill PDF.
type ParsedBill struct {
	Provider string          `json:"provider"`
	Items    []ParsedExpense `json:"items"`
}

// BillParser extracts transactions from a credit card bill.
// Each bank/provider has its own implementation (Open/Closed Principle).
type BillParser interface {
	Parse(ctx context.Context, reader io.Reader) (*ParsedBill, error)
	Supports(content []byte) bool
}

// ImportPreviewItem represents a single expense in the import preview,
// enriched with a suggested category.
type ImportPreviewItem struct {
	Description       string `json:"description"`
	AmountCents       int64  `json:"amount_cents"`
	Date              string `json:"date"`
	SuggestedCategory string `json:"suggested_category_id,omitempty"`
}

// ImportPreviewResponse is returned after parsing a bill, before user confirmation.
type ImportPreviewResponse struct {
	Provider string              `json:"provider"`
	Items    []ImportPreviewItem `json:"items"`
}

// ImportConfirmItem represents a single expense the user confirmed for import.
type ImportConfirmItem struct {
	CategoryID  string `json:"category_id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	ExpenseDate string `json:"expense_date"`
	IsShared    bool   `json:"is_shared"`
	PaidBy      string `json:"paid_by"`
	AssignedTo  string `json:"assigned_to"`
}

// ImportConfirmInput is the request body for confirming an import.
type ImportConfirmInput struct {
	Items []ImportConfirmItem `json:"items"`
}
