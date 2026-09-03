package quotationService

import (
	"time"
	"regexp"
	"fmt"
	"log/slog"
	model "github.com/Endcru/PlataTestAssignment/internal/models"
	storage "github.com/Endcru/PlataTestAssignment/internal/storage/postgres"
	storageBase "github.com/Endcru/PlataTestAssignment/internal/storage"
	quotationAPIService "github.com/Endcru/PlataTestAssignment/internal/service/quotationAPI"
)

const (
	QuotationNameRegex = `^[A-Z]{3}_[A-Z]{3}$`
)

var CurrencyMap = map[string][]string{
	"USD": {"EUR", "GBP", "JPY", "CAD", "AUD", "CHF"},
	"EUR": {"USD", "GBP", "JPY", "CAD", "AUD", "CHF"},
	"GBP": {"USD", "EUR", "JPY", "CAD", "AUD", "CHF"},
	"JPY": {"USD", "EUR", "GBP", "CAD", "AUD", "CHF"},
	"CAD": {"USD", "EUR", "GBP", "JPY", "AUD", "CHF"},
	"AUD": {"USD", "EUR", "GBP", "JPY", "CAD", "CHF"},
	"CHF": {"USD", "EUR", "GBP", "JPY", "CAD", "AUD"},
}


type QuotationService interface {
	ValidateQuotationName(name string) bool
	CreateStartQuotations(numberOfQuotations int) error
	GetQuotationByName(name string) (model.Quotation, error)
	UpdateQuotation(name string, newCurrencyRate float64) error
	CreateQuotationRequest(name string, requestedAt time.Time) (int, error)
	GetQuotationRequest(id int) (model.QuotationRequest, error)
	GetQuotationByRequestUpdateID(updateID int) (model.Quotation, error)
	DoneQuotationRequest(id int) error
	DeleteQuotationRequest(id int) error
	CreateQuotationUpdate(name string, previousRate float64, newRate float64, source string) error
	GetQuotationUpdate(name string) ([]model.QuotationUpdate, error)

}

type QuotationServiceImpl struct {
	storage *storage.Storage
	log *slog.Logger
	quotationNameRegex *regexp.Regexp
	quotationAPIService quotationAPIService.QuotationAPIService
}

func NewQuotationService(storage *storage.Storage, log *slog.Logger, quotationAPIService quotationAPIService.QuotationAPIService) QuotationService {
	return &QuotationServiceImpl{storage: storage, log: log, quotationNameRegex: regexp.MustCompile(QuotationNameRegex), quotationAPIService: quotationAPIService}
}

func (s *QuotationServiceImpl) ValidateQuotationName(name string) bool {
	const op = "quotationService.ValidateQuotationName"

	return s.quotationNameRegex.MatchString(name)
}

func (s *QuotationServiceImpl) CreateStartQuotations(numberOfQuotations int) error {
	const op = "quotationService.CreateStartQuotations"

	for base, quotes := range CurrencyMap {
		rates, err := s.quotationAPIService.GetQuotation(base, quotes)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		for quote, rate := range rates {
			err := s.storage.CreateQuotation(model.Quotation{Name: fmt.Sprintf("%s_%s", base, quote), Rate: rate, UpdatedAt: time.Now()})
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			err = s.storage.CreateQuotationUpdate(model.QuotationUpdate{Name: fmt.Sprintf("%s_%s", base, quote), PreviousRate: 0, NewRate: rate, Source: s.quotationAPIService.Name(), UpdatedAt: time.Now()})
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}
	return nil
}

func (s *QuotationServiceImpl) GetQuotationByName(name string) (model.Quotation, error) {
	const op = "quotationService.GetQuotationByName"
	quotation, err := s.storage.GetQuotation(name)
	if err != nil {
		return model.Quotation{}, fmt.Errorf("%s: %w", op, err)
	}
	return quotation, nil
}

func (s *QuotationServiceImpl) GetQuotationByRequestUpdateID(updateID int) (model.Quotation, error) {
	const op = "quotationService.GetQuotationByUpdateID"
	quotationRequest, err := s.storage.GetQuotationRequest(updateID)
	if err != nil {
		return model.Quotation{}, fmt.Errorf("%s: %w", op, err)
	}
	if !quotationRequest.Done {
		return model.Quotation{}, storageBase.ErrQuotationRequestNotDone
	}
	quotation, err := s.storage.GetQuotation(quotationRequest.Name)
	if err != nil {
		return model.Quotation{}, fmt.Errorf("%s: %w", op, err)
	}
	return quotation, nil
}
func (s *QuotationServiceImpl) UpdateQuotation(name string, newCurrencyRate float64) error {
	const op = "quotationService.UpdateQuotation"
	err := s.storage.UpdateQuotation(model.Quotation{Name: name, UpdatedAt: time.Now(), Rate: newCurrencyRate})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *QuotationServiceImpl) CreateQuotationRequest(name string, requestedAt time.Time) (int, error) {
	const op = "quotationService.CreateQuotationRequest"
	id, err := s.storage.CreateQuotationRequest(model.QuotationRequest{Name: name, RequestedAt: requestedAt})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *QuotationServiceImpl) GetQuotationRequest(id int) (model.QuotationRequest, error) {
	const op = "quotationService.GetQuotationRequest"
	quotationRequest, err := s.storage.GetQuotationRequest(id)
	if err != nil {
		return model.QuotationRequest{}, fmt.Errorf("%s: %w", op, err)
	}
	return quotationRequest, nil
}

func (s *QuotationServiceImpl) DoneQuotationRequest(id int) error {
	const op = "quotationService.DoneQuotationRequest"
	err := s.storage.DoneQuotationRequest(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *QuotationServiceImpl) DeleteQuotationRequest(id int) error {
	const op = "quotationService.DeleteQuotationRequest"
	err := s.storage.DeleteQuotationRequest(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *QuotationServiceImpl) CreateQuotationUpdate(name string, previousRate float64, newRate float64, source string) error {
	const op = "quotationService.CreateQuotationUpdate"
	err := s.storage.CreateQuotationUpdate(model.QuotationUpdate{Name: name, PreviousRate: previousRate, NewRate: newRate, Source: source})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *QuotationServiceImpl) GetQuotationUpdate(name string) ([]model.QuotationUpdate, error) {
	const op = "quotationService.GetQuotationUpdate"
	quotationUpdates, err := s.storage.GetQuotationUpdate(name)
	if err != nil {
		return []model.QuotationUpdate{}, fmt.Errorf("%s: %w", op, err)
	}
	return quotationUpdates, nil
}