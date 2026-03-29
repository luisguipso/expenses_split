package nubank

import (
	"testing"

	"github.com/lguilherme/contas/internal/domain"
)

func TestSupports(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{
			name:    "matches Nu Pagamentos",
			content: []byte("some header\nNu Pagamentos S.A.\nmore content"),
			want:    true,
		},
		{
			name:    "matches NU PAGAMENTOS uppercase",
			content: []byte("NU PAGAMENTOS S.A. CNPJ"),
			want:    true,
		},
		{
			name:    "matches nubank.com.br",
			content: []byte("visit nubank.com.br for details"),
			want:    true,
		},
		{
			name:    "matches Nubank mixed case",
			content: []byte("Nubank credit card bill"),
			want:    true,
		},
		{
			name:    "does not match unrelated content",
			content: []byte("Itaú Unibanco S.A. credit card bill"),
			want:    false,
		},
		{
			name:    "empty content",
			content: []byte{},
			want:    false,
		},
	}

	p := NewParser()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := p.Supports(tc.content)
			if got != tc.want {
				t.Errorf("Supports() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseBRLAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{name: "simple cents", input: "23,45", want: 2345},
		{name: "whole reais", input: "100,00", want: 10000},
		{name: "thousands separator", input: "1.234,56", want: 123456},
		{name: "large amount", input: "12.345,67", want: 1234567},
		{name: "multiple thousands", input: "1.234.567,89", want: 123456789},
		{name: "single digit reais", input: "5,00", want: 500},
		{name: "negative amount", input: "-42,00", want: -4200},
		{name: "negative with thousands", input: "-1.234,56", want: -123456},
		{name: "empty string", input: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseBRLAmount(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("parseBRLAmount(%q) expected error, got %d", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseBRLAmount(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseBRLAmount(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseTransactionLines(t *testing.T) {
	tests := []struct {
		name string
		text string
		year int
		want []domain.ParsedExpense
	}{
		{
			name: "single transaction",
			text: "15 MAR    UBER *TRIP               23,45",
			year: 2024,
			want: []domain.ParsedExpense{
				{Description: "UBER *TRIP", AmountCents: 2345, Date: "2024-03-15"},
			},
		},
		{
			name: "multiple transactions",
			text: `15 MAR    UBER *TRIP               23,45
16 MAR    IFOOD *IFOOD             45,90
02 ABR    AMAZON MKTPLACE          123,00`,
			year: 2024,
			want: []domain.ParsedExpense{
				{Description: "UBER *TRIP", AmountCents: 2345, Date: "2024-03-15"},
				{Description: "IFOOD *IFOOD", AmountCents: 4590, Date: "2024-03-16"},
				{Description: "AMAZON MKTPLACE", AmountCents: 12300, Date: "2024-04-02"},
			},
		},
		{
			name: "with installment info",
			text: "10 FEV    CASAS BAHIA Parcela 3/12   199,90",
			year: 2025,
			want: []domain.ParsedExpense{
				{Description: "CASAS BAHIA Parcela 3/12", AmountCents: 19990, Date: "2025-02-10"},
			},
		},
		{
			name: "large amount with thousands separator",
			text: "01 JAN    APPLE.COM/BILL           1.234,56",
			year: 2025,
			want: []domain.ParsedExpense{
				{Description: "APPLE.COM/BILL", AmountCents: 123456, Date: "2025-01-01"},
			},
		},
		{
			name: "negative amount (credit/refund)",
			text: "20 JUL    ESTORNO COMPRA            -50,00",
			year: 2024,
			want: []domain.ParsedExpense{
				{Description: "ESTORNO COMPRA", AmountCents: -5000, Date: "2024-07-20"},
			},
		},
		{
			name: "all months parse correctly",
			text: `01 JAN    A    10,00
01 FEV    B    10,00
01 MAR    C    10,00
01 ABR    D    10,00
01 MAI    E    10,00
01 JUN    F    10,00
01 JUL    G    10,00
01 AGO    H    10,00
01 SET    I    10,00
01 OUT    J    10,00
01 NOV    K    10,00
01 DEZ    L    10,00`,
			year: 2024,
			want: []domain.ParsedExpense{
				{Description: "A", AmountCents: 1000, Date: "2024-01-01"},
				{Description: "B", AmountCents: 1000, Date: "2024-02-01"},
				{Description: "C", AmountCents: 1000, Date: "2024-03-01"},
				{Description: "D", AmountCents: 1000, Date: "2024-04-01"},
				{Description: "E", AmountCents: 1000, Date: "2024-05-01"},
				{Description: "F", AmountCents: 1000, Date: "2024-06-01"},
				{Description: "G", AmountCents: 1000, Date: "2024-07-01"},
				{Description: "H", AmountCents: 1000, Date: "2024-08-01"},
				{Description: "I", AmountCents: 1000, Date: "2024-09-01"},
				{Description: "J", AmountCents: 1000, Date: "2024-10-01"},
				{Description: "K", AmountCents: 1000, Date: "2024-11-01"},
				{Description: "L", AmountCents: 1000, Date: "2024-12-01"},
			},
		},
		{
			name: "filters non-transaction lines",
			text: `FATURA DO CARTÃO
Nu Pagamentos S.A.
Vencimento: 15 ABR 2024
TRANSAÇÕES
15 MAR    UBER *TRIP               23,45
TOTAL                              23,45
Obrigado por usar Nubank`,
			year: 2024,
			want: []domain.ParsedExpense{
				{Description: "UBER *TRIP", AmountCents: 2345, Date: "2024-03-15"},
			},
		},
		{
			name: "empty text",
			text: "",
			year: 2024,
			want: nil,
		},
		{
			name: "no matching lines",
			text: `This is just random text
without any transaction lines
at all`,
			year: 2024,
			want: nil,
		},
		{
			name: "whitespace around lines",
			text: "  15 MAR    UBER *TRIP               23,45  ",
			year: 2024,
			want: []domain.ParsedExpense{
				{Description: "UBER *TRIP", AmountCents: 2345, Date: "2024-03-15"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseTransactionLines(tc.text, tc.year)
			if len(got) != len(tc.want) {
				t.Fatalf("parseTransactionLines() returned %d items, want %d\ngot: %+v", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("item[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestInferYear(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "extracts year from closing date",
			text: "Vencimento: 15 ABR 2024\nsome transactions",
			want: 2024,
		},
		{
			name: "uses last year found",
			text: "Período: 2023\nVencimento: 10 JAN 2024",
			want: 2024,
		},
		{
			name: "single year in document",
			text: "Fatura de 2025\n01 JAN COMPRA 10,00",
			want: 2025,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := inferYear(tc.text)
			if got != tc.want {
				t.Errorf("inferYear() = %d, want %d", got, tc.want)
			}
		})
	}
}
