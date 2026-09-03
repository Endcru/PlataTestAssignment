package mock

import (
	model "github.com/Endcru/PlataTestAssignment/internal/models"
	storageBase "github.com/Endcru/PlataTestAssignment/internal/storage"
)

type mockStorage struct {
	quotations map[string]model.Quotation
	quotationRequests map[int]model.QuotationRequest
	quotationRequestsUpdates map[int]model.QuotationUpdate
	nextRequestID int
	nextUpdateID int
}

func NewMockStorage() *mockStorage {
	return &mockStorage{
		quotations: make(map[string]model.Quotation),
		quotationRequests: make(map[int]model.QuotationRequest),
		quotationRequestsUpdates: make(map[int]model.QuotationUpdate),
		nextRequestID: 1,
		nextUpdateID: 1,
	}
}

func (m *mockStorage) CreateQuotation(quotation model.Quotation) error {
	if _, exists := m.quotations[quotation.Name]; exists {
		return storageBase.ErrQuotationAlreadyExists
	}
	m.quotations[quotation.Name] = quotation
	return nil
}

func (m *mockStorage) CreateQuotationUpdate(quotationUpdate model.QuotationUpdate) error {
	if quotationUpdate.ID == 0 {
		quotationUpdate.ID = m.nextUpdateID
		m.nextUpdateID++
	}
	if _, exists := m.quotationRequestsUpdates[quotationUpdate.ID]; exists {
		return storageBase.ErrQuotationUpdateAlreadyExists
	}
	m.quotationRequestsUpdates[quotationUpdate.ID] = quotationUpdate
	return nil
}

func (m *mockStorage) CreateQuotationRequest(quotationRequest model.QuotationRequest) (int, error) {
	if quotationRequest.ID == 0 {
		quotationRequest.ID = m.nextRequestID
		m.nextRequestID++
	}
	if _, exists := m.quotationRequests[quotationRequest.ID]; exists {
		return 0, storageBase.ErrQuotationRequestAlreadyExists
	}
	m.quotationRequests[quotationRequest.ID] = quotationRequest
	return quotationRequest.ID, nil
}

func (m *mockStorage) GetQuotation(name string) (model.Quotation, error) {
	quotation, exists := m.quotations[name]
	if !exists {
		return model.Quotation{}, storageBase.ErrQuotationNotFound
	}
	return quotation, nil
}

func (m *mockStorage) GetQuotationUpdates(name string) ([]model.QuotationUpdate, error) {
	quotationUpdates := []model.QuotationUpdate{}
	for _, v := range m.quotationRequestsUpdates {
		if v.Name == name {
			quotationUpdates = append(quotationUpdates, v)
		}
	}
	return quotationUpdates, nil
}

func (m *mockStorage) GetQuotationRequest(id int) (model.QuotationRequest, error) {
	quotationRequest, exists := m.quotationRequests[id]
	if !exists {
		return model.QuotationRequest{}, storageBase.ErrQuotationRequestNotFound
	}
	return quotationRequest, nil
}

func (m *mockStorage) GetQuotationRequestUncompletedByName(name string) (int, error) {
	for k, v := range m.quotationRequests {
		if v.Name == name && !v.Done {
			return k, nil
		}
	}
	return 0, storageBase.ErrQuotationRequestUncompletedNotFound
}

func (m *mockStorage) GetQuotationRequestsUncompleted() ([]model.QuotationRequest, error) {
	quotationRequests := []model.QuotationRequest{}
	for _, v := range m.quotationRequests {
		if !v.Done {
			quotationRequests = append(quotationRequests, v)
		}
	}
	return quotationRequests, nil
}

func (m *mockStorage) UpdateQuotation(quotation model.Quotation) error {
	if _, exists := m.quotations[quotation.Name]; !exists {
		return storageBase.ErrQuotationNotFound
	}
	m.quotations[quotation.Name] = quotation
	return nil
}

func (m *mockStorage) DoneQuotationRequest(id int) error {
	quotationRequest, exists := m.quotationRequests[id]
	if !exists {
		return storageBase.ErrQuotationRequestNotFound
	}
	quotationRequest.Done = true
	m.quotationRequests[id] = quotationRequest
	return nil
}

func (m *mockStorage) DeleteQuotationRequest(id int) error {
	if _, exists := m.quotationRequests[id]; !exists {
		return storageBase.ErrQuotationRequestNotFound
	}
	delete(m.quotationRequests, id)
	return nil
}