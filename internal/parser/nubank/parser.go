package nubank

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

	// Matches the FATURA header: "FATURA DD MMM YYYY"
	faturaYearRe = regexp.MustCompile(`FATURA\s+\d{2}\s+(?:JAN|FEV|MAR|ABR|MAI|JUN|JUL|AGO|SET|OUT|NOV|DEZ)\s+(20\d{2})`)

	// Matches a standalone date row from GetTextByRow: "DD MMM"
	dateRowRe = regexp.MustCompile(
		`^(0[1-9]|[12]\d|3[01])\s+(JAN|FEV|MAR|ABR|MAI|JUN|JUL|AGO|SET|OUT|NOV|DEZ)$`,
	)

	// Matches a detail row: [•••• DDDD]Description[−/-]R$ D.DDD,DD
	detailRowRe = regexp.MustCompile(
		`^(?:\x{2022}+\s*\d{4})?(.+?)([−-]?)R\$\s*(\d{1,3}(?:\.\d{3})*,\d{2})$`,
	)
)

// Parser extracts transactions from Nubank credit card bill PDFs.
type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Supports(content []byte) bool {
	// First try raw bytes (works for plain text input and tests).
	lower := bytes.ToLower(content)
	if bytes.Contains(lower, []byte("nu pagamentos")) ||
		bytes.Contains(lower, []byte("nubank")) {
		return true
	}

	// PDF files store text in compressed streams, so keywords won't appear
	// in raw bytes. Extract via GetPlainText which works for keyword matching.
	text, err := extractPlainText(content)
	if err != nil {
		return false
	}
	textLower := strings.ToLower(text)
	return strings.Contains(textLower, "nu pagamentos") ||
		strings.Contains(textLower, "nubank")
}

func (p *Parser) Parse(ctx context.Context, reader io.Reader) (*domain.ParsedBill, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("nubank parse: read input: %w", err)
	}

	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("nubank parse: open pdf: %w", err)
	}

	text := buildTextFromRows(r)
	year := inferYear(text)
	items := parseTransactionLines(text, year)

	return &domain.ParsedBill{
		Provider: "Nubank",
		Items:    items,
	}, nil
}

// extractPlainText uses GetPlainText for simple text extraction (keyword search, year inference).
func extractPlainText(content []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", err
	}
	pt, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, pt); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// buildTextFromRows uses GetTextByRow to reconstruct properly formatted text.
// Nubank PDFs render each transaction as two rows:
//
//	Row 1: "DD MMM" (date)
//	Row 2: "[•••• DDDD]Description[−]R$ D.DDD,DD" (card mask + description + amount)
//
// This function combines them into: "DD MMM    Description    [-]D.DDD,DD"
// which matches the format expected by parseTransactionLines.
// Non-transaction rows are included as-is for year inference.
func buildTextFromRows(r *pdf.Reader) string {
	var lines []string

	for pg := 1; pg <= r.NumPage(); pg++ {
		page := r.Page(pg)
		rows, err := page.GetTextByRow()
		if err != nil {
			continue
		}

		var textRows []string
		for _, row := range rows {
			var parts []string
			for _, t := range row.Content {
				if t.S != "" {
					parts = append(parts, t.S)
				}
			}
			text := strings.Join(parts, "")
			if strings.TrimSpace(text) != "" {
				textRows = append(textRows, text)
			}
		}

		for i := 0; i < len(textRows); i++ {
			row := textRows[i]

			dateMatch := dateRowRe.FindStringSubmatch(row)
			if dateMatch == nil {
				lines = append(lines, row)
				continue
			}

			// Found a date row — check if the next row is a transaction detail.
			if i+1 < len(textRows) {
				if line, ok := formatTransactionLine(dateMatch[0], textRows[i+1]); ok {
					lines = append(lines, line)
					i++ // skip the detail row
					continue
				}
			}

			lines = append(lines, row)
		}
	}

	return strings.Join(lines, "\n")
}

// formatTransactionLine combines a date row and a detail row into a single line
// in the format "DD MMM    Description    [-]Amount" that parseTransactionLines expects.
func formatTransactionLine(date, detail string) (string, bool) {
	matches := detailRowRe.FindStringSubmatch(detail)
	if matches == nil {
		return "", false
	}

	description := strings.TrimSpace(matches[1])
	sign := matches[2]
	amount := matches[3]

	if description == "" {
		return "", false
	}

	// Skip payment entries — these are not expenses.
	if strings.HasPrefix(strings.ToLower(description), "pagamento") {
		return "", false
	}

	prefix := ""
	if sign != "" {
		prefix = "-"
	}

	return fmt.Sprintf("%s    %s    %s%s", date, description, prefix, amount), true
}

// inferYear extracts the bill year from the document text.
// It first looks for the FATURA header pattern (e.g., "FATURA 09 MAR 2026"),
// then falls back to the first 4-digit year found, then to the current year.
func inferYear(text string) int {
	// Prefer the year from the FATURA header — most reliable.
	if m := faturaYearRe.FindStringSubmatch(text); m != nil {
		y, _ := strconv.Atoi(m[1])
		return y
	}

	matches := yearRe.FindAllString(text, -1)
	if len(matches) == 0 {
		return time.Now().Year()
	}
	// Use the first year found — typically from the bill header/closing date.
	y, _ := strconv.Atoi(matches[0])
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
