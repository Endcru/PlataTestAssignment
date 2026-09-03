package mock

import "fmt"

type mockApiQuotationService struct {
	name  string
	rates map[string]map[string]float64 // base -> quote -> rate
	err   error
}

func NewMockApiQuotationService() *mockApiQuotationService {
	return &mockApiQuotationService{
		name:  "mock-api",
		rates: make(map[string]map[string]float64),
	}
}

func (m *mockApiQuotationService) Name() string {
	return m.name
}

func (m *mockApiQuotationService) SetName(name string) {
	m.name = name
}

func (m *mockApiQuotationService) SetRate(base, quote string, rate float64) {
	if _, ok := m.rates[base]; !ok {
		m.rates[base] = make(map[string]float64)
	}
	m.rates[base][quote] = rate
}

func (m *mockApiQuotationService) SetError(err error) {
	m.err = err
}

func (m *mockApiQuotationService) GetQuotation(base string, quotes []string) (map[string]float64, error) {
	if m.err != nil {
		return nil, m.err
	}

	baseRates, ok := m.rates[base]
	if !ok {
		return nil, fmt.Errorf("rates for base %q not found", base)
	}

	result := make(map[string]float64, len(quotes))
	for _, quote := range quotes {
		rate, ok := baseRates[quote]
		if !ok {
			return nil, fmt.Errorf("rate for %s_%s not found", base, quote)
		}
		result[quote] = rate
	}
	return result, nil
}
