package storage

import "errors"

var (
	ErrQuotationNotFound = errors.New("quotation not found")
	ErrQuotationRequestNotFound = errors.New("quotation request not found")
	ErrQuotationRequestNotDone = errors.New("quotation request not done")
)