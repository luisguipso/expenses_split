package nubank

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dslipak/pdf"
	"github.com/lguilherme/contas/internal/domain"
)

var (
	monthMap = map[string]int{
		"JAN": 1, "FEV": 2, "MAR": 3, "ABR": 4,
		"MAI": 5, "JUN": 6, "JUL": 7, "AGO": 8,
		"SET": 9, "OUT": 10, "NOV": 11, "DEZ": 12,
	}

	// Matches: DD MMM    DESCRIPTION    VALUE (Brazilian format)
	// Day must be 01-31 range.
	transactionRe = regexp.MustCompile(
		`^(0[1-9]|[12]\d|3[01])\s+(JAN|FEV|MAR|ABR|MAI|JUN|JUL|AGO|SET|OUT|NOV|DEZ)\s+(.+?)\s+(-?\d{1,3}(?:\.\d{3})*,\d{2})$`,
	)

	// Matches a 4-digit year in the document to infer bill year.
	yearRe = regexp.MustCompile(`\b(20\d{2})\b`)
)

// Parser extracts transactions from Nubank credit card bill PDFs.
type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Supports(content []byte) bool {
	lower := bytes.ToLower(content)
	return bytes.Contains(lower, []byte("nu pagamentos")) ||
		bytes.Contains(lower, []byte("nubank"))
}

func (p *Parser) Parse(ctx context.Context, reader io.Reader) (*domain.ParsedBill, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("nubank parse: read input: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "nubank-bill-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("nubank parse: create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("nubank parse: write temp file: %w", err)
	}
	tmpFile.Close()

	pdfReader, err := pdf.Open(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("nubank parse: open pdf: %w", err)
	}

	var buf bytes.Buffer
	plainText, err := pdfReader.GetPlainText()
	if err != nil {
		return nil, fmt.Errorf("nubank parse: extract text: %w", err)
	}
	if _, err := io.Copy(&buf, plainText); err != nil {
		return nil, fmt.Errorf("nubank parse: read text: %w", err)
	}

	text := buf.String()
	year := inferYear(text)
	items := parseTransactionLines(text, year)

	return &domain.ParsedBill{
		Provider: "Nubank",
		Items:    items,
	}, nil
}

// inferYear extracts the most likely bill year from the document text.
// It returns the last 4-digit year found (typically the closing date year),
// or the current year as fallback.
func inferYear(text string) int {
	matches := yearRe.FindAllString(text, -1)
	if len(matches) == 0 {
		return time.Now().Year()
	}
	// Use the last year found — usually the closing/due date near the top.
	y, _ := strconv.Atoi(matches[len(matches)-1])
	return y
}

// parseTransactionLines extracts transactions from the plain text of a Nubank bill.
// It infers the correct year for each transaction, handling year boundaries
// (e.g., a Jan bill with Dec transactions from the previous year).
func parseTransactionLines(text string, billYear int) []domain.ParsedExpense {
	var items []domain.ParsedExpense
	lines := strings.Split(text, "\n")

	// First pass: collect all transaction months to detect year boundary.
	var parsedRows []struct {
		day         int
		month       int
		description string
		cents       int64
	}
	minMonth := 13

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := transactionRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		day, _ := strconv.Atoi(matches[1])
		monthStr := matches[2]
		description := strings.TrimSpace(matches[3])
		amountStr := matches[4]

		if description == "" {
			continue
		}

		month, ok := monthMap[monthStr]
		if !ok {
			continue
		}

		cents, err := parseBRLAmount(amountStr)
		if err != nil {
			continue
		}

		// Validate the date is real (e.g., reject Feb 31).
		candidate := time.Date(billYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if candidate.Day() != day || candidate.Month() != time.Month(month) {
			continue
		}

		if month < minMonth {
			minMonth = month
		}

		parsedRows = append(parsedRows, struct {
			day         int
			month       int
			description string
			cents       int64
		}{day, month, description, cents})
	}

	// Second pass: assign year, adjusting for year boundary.
	// If the bill has early-year months (Jan/Feb) AND late-year months (Nov/Dec),
	// the late-year transactions belong to the previous year.
	for _, r := range parsedRows {
		y := billYear
		if minMonth <= 3 && r.month >= 11 {
			y = billYear - 1
		}
		date := fmt.Sprintf("%04d-%02d-%02d", y, r.month, r.day)
		items = append(items, domain.ParsedExpense{
			Description: r.description,
			AmountCents: r.cents,
			Date:        date,
		})
	}

	return items
}

// parseBRLAmount converts a Brazilian-format amount string to int64 cents.
// Examples: "23,45" → 2345, "1.234,56" → 123456, "-42,00" → -4200.
func parseBRLAmount(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}

	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	// Remove thousand separators (dots) and replace decimal comma with dot.
	s = strings.ReplaceAll(s, ".", "")
	s = strings.Replace(s, ",", ".", 1)

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse amount %q: %w", s, err)
	}

	cents := int64(f*100 + 0.5)
	if negative {
		cents = -cents
	}
	return cents, nil
}
