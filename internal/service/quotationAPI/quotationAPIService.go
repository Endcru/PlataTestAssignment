package quotationAPIService

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type QuotationAPIService interface {
	GetQuotation(base string, quote []string) (map[string]float64, error)
	Name() string
}

type QuotationAPIServiceImpl struct {
	name string
	apiKey string
	log *slog.Logger
}

func NewQuotationAPIService(name string, apiKey string, log *slog.Logger) QuotationAPIService {
	return &QuotationAPIServiceImpl{name: name, apiKey: apiKey, log: log}
}

type QuotationResponse struct {
	Meta struct {
		Code int `json:"code"`
		Disclaimer string `json:"disclaimer"`
	} `json:"meta"`
	Response struct {
		Date string `json:"date"`
		Base string `json:"base"`
		Rates map[string]float64 `json:"rates"`
	} `json:"response"`
}

func (s *QuotationAPIServiceImpl) Name() string {
	return s.name
}

func (s *QuotationAPIServiceImpl) GetQuotation(base string, quote []string) (map[string]float64, error) {
	const op = "quotationAPIService.GetQuotation"
	url := fmt.Sprintf("https://api.currencybeacon.com/v1/latest?api_key=%s&base=%s&symbols=%s", s.apiKey, base, strings.Join(quote, ","))
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var quotationResponse QuotationResponse
	err = json.Unmarshal(body, &quotationResponse)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if quotationResponse.Meta.Code != 200 {
		return nil, fmt.Errorf("%s: %w", op, errors.New("failed to get quotation"))
	}
	return quotationResponse.Response.Rates, nil
}