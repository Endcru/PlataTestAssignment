package storage

import (
	"errors"

	model "github.com/Endcru/PlataTestAssignment/internal/models"
)

var (
	ErrQuotationNotFound = errors.New("quotation not found")
	ErrQuotationAlreadyExists = errors.New("quotation already exists")
	ErrQuotationRequestNotFound = errors.New("quotation request not found")
	ErrQuotationRequestNotDone = errors.New("quotation request not done")
	ErrQuotationRequestAlreadyExists = errors.New("quotation request already exists")
	ErrQuotationUpdateAlreadyExists = errors.New("quotation update already exists")
	ErrQuotationRequestUncompletedNotFound = errors.New("quotation request uncompleted not found")
)

type Storage interface {
	CreateQuotation(quotation model.Quotation) error
	CreateQuotationUpdate(quotationUpdate model.QuotationUpdate) error
	CreateQuotationRequest(quotationRequest model.QuotationRequest) (int, error)
	GetQuotation(name string) (model.Quotation, error)
	GetQuotationUpdates(name string) ([]model.QuotationUpdate, error)
	GetQuotationRequest(id int) (model.QuotationRequest, error)
	GetQuotationRequestUncompletedByName(name string) (int, error)
	GetQuotationRequestsUncompleted() ([]model.QuotationRequest, error)
	UpdateQuotation(quotation model.Quotation) error
	DoneQuotationRequest(id int) error
	DeleteQuotationRequest(id int) error
}