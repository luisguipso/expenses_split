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
			name: "all months parse correctly (mid-year bill)",
			text: `01 JAN    A    10,00
01 FEV    B    10,00
01 MAR    C    10,00
01 ABR    D    10,00
01 MAI    E    10,00
01 JUN    F    10,00
01 JUL    G    10,00
01 AGO    H    10,00
01 SET    I    10,00
01 OUT    J    10,00`,
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
			},
		},
		{
			name: "year boundary: Jan bill with Dec transactions",
			text: `28 NOV    COMPRA NOV               50,00
15 DEZ    COMPRA DEZ              100,00
02 JAN    COMPRA JAN               25,00`,
			year: 2025,
			want: []domain.ParsedExpense{
				{Description: "COMPRA NOV", AmountCents: 5000, Date: "2024-11-28"},
				{Description: "COMPRA DEZ", AmountCents: 10000, Date: "2024-12-15"},
				{Description: "COMPRA JAN", AmountCents: 2500, Date: "2025-01-02"},
			},
		},
		{
			name: "no year boundary when all months are mid-year",
			text: `10 MAI    COMPRA A    50,00
15 JUN    COMPRA B    75,00`,
			year: 2024,
			want: []domain.ParsedExpense{
				{Description: "COMPRA A", AmountCents: 5000, Date: "2024-05-10"},
				{Description: "COMPRA B", AmountCents: 7500, Date: "2024-06-15"},
			},
		},
		{
			name: "rejects invalid date Feb 31",
			text: "31 FEV    IMPOSSIBLE DATE    10,00",
			year: 2024,
			want: nil,
		},
		{
			name: "rejects invalid day 00",
			text: "00 JAN    BAD DAY    10,00",
			year: 2024,
			want: nil,
		},
		{
			name: "filters empty description",
			text: "15 MAR        23,45",
			year: 2024,
			want: nil,
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
			name: "extracts year from FATURA header",
			text: "FATURA 09 MAR 2026EMISSÃO E ENVIO 01 MAR 2026\nsome text\nanos de 2025",
			want: 2026,
		},
		{
			name: "falls back to first year when no FATURA header",
			text: "Vencimento: 15 ABR 2024\nsome transactions\nano de 2023",
			want: 2024,
		},
		{
			name: "uses first year found",
			text: "Período: 2024\nVencimento: 10 JAN 2023",
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

func TestFormatTransactionLine(t *testing.T) {
	tests := []struct {
		name   string
		date   string
		detail string
		want   string
		wantOK bool
	}{
		{
			name:   "standard transaction with card mask",
			date:   "01 FEV",
			detail: "•••• 9581Dm Auto Mecanica - Parcela 2/3R$ 406,00",
			want:   "01 FEV    Dm Auto Mecanica - Parcela 2/3    406,00",
			wantOK: true,
		},
		{
			name:   "refund with unicode minus",
			date:   "03 FEV",
			detail: "Estorno de \"Shein *Mega Kids Moda\"\u2212R$ 41,90",
			want:   "03 FEV    Estorno de \"Shein *Mega Kids Moda\"    -41,90",
			wantOK: true,
		},
		{
			name:   "refund with ascii minus",
			date:   "08 FEV",
			detail: "Estorno de \"Mercadolivre*Grupovoke\"-R$ 80,33",
			want:   "08 FEV    Estorno de \"Mercadolivre*Grupovoke\"    -80,33",
			wantOK: true,
		},
		{
			name:   "NuPay without card mask",
			date:   "18 FEV",
			detail: "Cobasi - NuPayR$ 35,64",
			want:   "18 FEV    Cobasi - NuPay    35,64",
			wantOK: true,
		},
		{
			name:   "thousands separator in amount",
			date:   "10 FEV",
			detail: "•••• 9581Jim.Com* Souza Gas eR$ 1.250,00",
			want:   "10 FEV    Jim.Com* Souza Gas e    1.250,00",
			wantOK: true,
		},
		{
			name:   "skips payment entries",
			date:   "09 FEV",
			detail: "Pagamento em 09 FEV\u2212R$ 6.236,67",
			wantOK: false,
		},
		{
			name:   "rejects non-detail row",
			date:   "01 FEV",
			detail: "TRANSAÇÕES",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := formatTransactionLine(tc.date, tc.detail)
			if ok != tc.wantOK {
				t.Fatalf("formatTransactionLine() ok = %v, want %v (got %q)", ok, tc.wantOK, got)
			}
			if ok && got != tc.want {
				t.Errorf("formatTransactionLine() = %q, want %q", got, tc.want)
			}
		})
	}
}
