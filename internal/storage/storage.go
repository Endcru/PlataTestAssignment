package storage

import "errors"

var (
	ErrQuotationNotFound = errors.New("quotation not found")
	ErrQuotationAlreadyExists = errors.New("quotation already exists")
	ErrQuotationRequestNotFound = errors.New("quotation request not found")
	ErrQuotationRequestNotDone = errors.New("quotation request not done")
)