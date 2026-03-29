package parser

import (
	"fmt"

	"github.com/lguilherme/contas/internal/domain"
)

// Registry holds all registered bill parsers and selects the appropriate one.
type Registry struct {
	parsers []domain.BillParser
}

func NewRegistry(parsers ...domain.BillParser) *Registry {
	return &Registry{parsers: parsers}
}

// FindParser returns the first parser that supports the given content.
func (r *Registry) FindParser(content []byte) (domain.BillParser, error) {
	for _, p := range r.parsers {
		if p.Supports(content) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("%w: no parser matched the uploaded file", domain.ErrUnsupportedBillFormat)
}
