package quotationService

import (
	"testing"
	"time"

	model "github.com/Endcru/PlataTestAssignment/internal/models"
	storageBase "github.com/Endcru/PlataTestAssignment/internal/storage"
	"github.com/Endcru/PlataTestAssignment/internal/test/mock"
)

func TestQuotationService_GetQuotationByName(t *testing.T) {
	st := mock.NewMockStorage()
	want := model.Quotation{
		Name:      "EUR_MXN",
		Rate:      21.5,
		UpdatedAt: time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC),
	}
	if err := st.CreateQuotation(want); err != nil {
		t.Fatalf("failed to seed mock storage: %v", err)
	}

	svc := NewQuotationService(st, mock.NewMockLogger(), nil)

	got, err := svc.GetQuotationByName("EUR_MXN")
	if err != nil {
		t.Fatalf("Failed to get quotation: %v", err)
	}

	if got.Name != want.Name || got.Rate != want.Rate {
		t.Fatalf("unexpected quotation: got %+v, want %+v", got, want)
	}
}

func TestQuotationService_GetQuotationByName_NotFound(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	_, err := svc.GetQuotationByName("EUR_MXN")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQuotationService_CreateQuotationRequest(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	id, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}
	if id != 1 {
		t.Fatalf("unexpected request id: got %d, want 1", id)
	}
}

func TestQuotationService_CreateQuotationRequest_AlreadyExists(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	id1, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}
	id2, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 13, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}
	if id2 != id1 {
		t.Fatalf("unexpected request id: got %d, want %d", id2, id1)
	}
}

func TestQuotationService_GetQuotationRequest(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	id, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}
	got, err := svc.GetQuotationRequest(id)
	if err != nil {
		t.Fatalf("failed to get quotation request: %v", err)
	}
	if got.Name != "EUR_MXN" {
		t.Fatalf("unexpected quotation request: got %+v, want EUR_MXN", got)
	}
}

func TestQuotationService_GetQuotationRequest_NotFound(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	_, err := svc.GetQuotationRequest(1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQuotationService_DoneQuotationRequest(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	id, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}
	err = svc.DoneQuotationRequest(id)
	if err != nil {
		t.Fatalf("failed to done quotation request: %v", err)
	}
	got, err := svc.GetQuotationRequest(id)
	if err != nil {
		t.Fatalf("failed to get quotation request: %v", err)
	}
	if !got.Done {
		t.Fatalf("unexpected quotation request: got %+v, want done", got)
	}
}

func TestQuotationService_DeleteQuotationRequest(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	id, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}
	err = svc.DeleteQuotationRequest(id)
	if err != nil {
		t.Fatalf("failed to delete quotation request: %v", err)
	}
}

func TestQuotationService_DeleteQuotationRequest_NotFound(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	err := svc.DeleteQuotationRequest(1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQuotationService_UpdateQuotation(t *testing.T) {
	st := mock.NewMockStorage()
	baseQuotation := model.Quotation{
		Name:      "EUR_MXN",
		Rate:      21.5,
		UpdatedAt: time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC),
	}
	if err := st.CreateQuotation(baseQuotation); err != nil {
		t.Fatalf("failed to seed mock storage: %v", err)
	}

	svc := NewQuotationService(st, mock.NewMockLogger(), nil)
	newRate := 21.6

	err := svc.UpdateQuotation("EUR_MXN", newRate)
	if err != nil {
		t.Fatalf("failed to update quotation: %v", err)
	}
	got, err := svc.GetQuotationByName("EUR_MXN")
	if err != nil {
		t.Fatalf("failed to get quotation: %v", err)
	}
	if got.Rate != newRate {
		t.Fatalf("unexpected quotation: got %+v, want %+v", got, newRate)
	}
}

func TestQuotationService_UpdateQuotation_NotFound(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	err := svc.UpdateQuotation("EUR_MXN", 21.6)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQuotationService_GetQuotationByRequestUpdateID(t *testing.T) {
	st := mock.NewMockStorage()
	baseQuotation := model.Quotation{
		Name:      "EUR_MXN",
		Rate:      21.5,
		UpdatedAt: time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC),
	}
	if err := st.CreateQuotation(baseQuotation); err != nil {
		t.Fatalf("failed to seed mock storage: %v", err)
	}

	svc := NewQuotationService(st, mock.NewMockLogger(), nil)

	requestID, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}

	newRate := 21.6
	err = svc.UpdateQuotation("EUR_MXN", newRate)
	if err != nil {
		t.Fatalf("failed to update quotation: %v", err)
	}

	err = svc.DoneQuotationRequest(requestID)
	if err != nil {
		t.Fatalf("failed to done quotation request: %v", err)
	}

	quotationRequest, err := st.GetQuotationRequest(requestID)
	if err != nil {
		t.Fatalf("failed to get quotation request: %v", err)
	}
	if !quotationRequest.Done {
		t.Fatalf("unexpected quotation request: got %+v, want done", quotationRequest)
	}

	got, err := svc.GetQuotationByRequestUpdateID(requestID)
	if err != nil {
		t.Fatalf("failed to get quotation by request update id: %v", err)
	}
	if got.Rate != newRate {
		t.Fatalf("unexpected quotation: got %+v, want %+v", got, newRate)
	}
}

func TestQuotationService_GetQuotationByRequestUpdateID_NotFound(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	_, err := svc.GetQuotationByRequestUpdateID(1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQuotationService_GetQuotationByRequestUpdateID_NotDone(t *testing.T) {
	st := mock.NewMockStorage()
	baseQuotation := model.Quotation{
		Name:      "EUR_MXN",
		Rate:      21.5,
		UpdatedAt: time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC),
	}
	if err := st.CreateQuotation(baseQuotation); err != nil {
		t.Fatalf("failed to seed mock storage: %v", err)
	}

	svc := NewQuotationService(st, mock.NewMockLogger(), nil)

	requestID, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}

	_, err = svc.GetQuotationByRequestUpdateID(requestID)
	if err != storageBase.ErrQuotationRequestNotDone {
		t.Fatalf("expected ErrQuotationRequestNotDone, got %v", err)
	}
}

func TestQuotationService_GetQuotationByRequestUpdateID_QuotationNotFound(t *testing.T) {
	st := mock.NewMockStorage()
	svc := NewQuotationService(st, mock.NewMockLogger(), nil)

	requestID, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}
	if err := svc.DoneQuotationRequest(requestID); err != nil {
		t.Fatalf("failed to done quotation request: %v", err)
	}

	_, err = svc.GetQuotationByRequestUpdateID(requestID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQuotationService_ValidateQuotationName(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	if !svc.ValidateQuotationName("EUR_MXN") {
		t.Fatal("expected EUR_MXN to be valid")
	}
	if svc.ValidateQuotationName("eur_mxn") {
		t.Fatal("expected eur_mxn to be invalid")
	}
	if svc.ValidateQuotationName("EUR/MXN") {
		t.Fatal("expected EUR/MXN to be invalid")
	}
	if svc.ValidateQuotationName("EU_MXN") {
		t.Fatal("expected EU_MXN to be invalid")
	}
}

func TestQuotationService_CreateQuotationUpdate(t *testing.T) {
	st := mock.NewMockStorage()
	svc := NewQuotationService(st, mock.NewMockLogger(), nil)

	err := svc.CreateQuotationUpdate("EUR_MXN", 21.5, 21.6, "mock-api")
	if err != nil {
		t.Fatalf("failed to create quotation update: %v", err)
	}

	got, err := svc.GetQuotationUpdates("EUR_MXN")
	if err != nil {
		t.Fatalf("failed to get quotation updates: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("unexpected updates count: got %d, want 1", len(got))
	}
	if got[0].PreviousRate != 21.5 || got[0].NewRate != 21.6 || got[0].Source != "mock-api" {
		t.Fatalf("unexpected update: %+v", got[0])
	}
}

func TestQuotationService_GetQuotationUpdates(t *testing.T) {
	st := mock.NewMockStorage()
	svc := NewQuotationService(st, mock.NewMockLogger(), nil)

	if err := svc.CreateQuotationUpdate("EUR_MXN", 0, 21.5, "mock-api"); err != nil {
		t.Fatalf("failed to create quotation update: %v", err)
	}
	if err := svc.CreateQuotationUpdate("EUR_MXN", 21.5, 21.6, "mock-api"); err != nil {
		t.Fatalf("failed to create quotation update: %v", err)
	}

	got, err := svc.GetQuotationUpdates("EUR_MXN")
	if err != nil {
		t.Fatalf("failed to get quotation updates: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("unexpected updates count: got %d, want 2", len(got))
	}
}

func TestQuotationService_GetQuotationUpdates_Empty(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	got, err := svc.GetQuotationUpdates("EUR_MXN")
	if err != nil {
		t.Fatalf("failed to get quotation updates: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("unexpected updates count: got %d, want 0", len(got))
	}
}

func TestQuotationService_DoneQuotationRequest_NotFound(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), nil)

	err := svc.DoneQuotationRequest(1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQuotationService_CreateStartQuotations(t *testing.T) {
	st := mock.NewMockStorage()
	api := mock.NewMockApiQuotationService()
	for base, quotes := range CurrencyMap {
		for _, quote := range quotes {
			api.SetRate(base, quote, 1.23)
		}
	}

	svc := NewQuotationService(st, mock.NewMockLogger(), api)
	err := svc.CreateStartQuotations()
	if err != nil {
		t.Fatalf("failed to create start quotations: %v", err)
	}

	got, err := svc.GetQuotationByName("EUR_USD")
	if err != nil {
		t.Fatalf("failed to get quotation: %v", err)
	}
	if got.Rate != 1.23 {
		t.Fatalf("unexpected rate: got %v, want 1.23", got.Rate)
	}

	updates, err := svc.GetQuotationUpdates("EUR_USD")
	if err != nil {
		t.Fatalf("failed to get quotation updates: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("unexpected updates count: got %d, want 1", len(updates))
	}
}

func TestQuotationService_CreateStartQuotations_AlreadyExists(t *testing.T) {
	st := mock.NewMockStorage()
	api := mock.NewMockApiQuotationService()
	for base, quotes := range CurrencyMap {
		for _, quote := range quotes {
			name := base + "_" + quote
			err := st.CreateQuotation(model.Quotation{Name: name, Rate: 1.23, UpdatedAt: time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)})
			if err != nil {
				t.Fatalf("failed to seed quotation %s: %v", name, err)
			}
		}
	}

	svc := NewQuotationService(st, mock.NewMockLogger(), api)
	err := svc.CreateStartQuotations()
	if err != nil {
		t.Fatalf("failed to create start quotations: %v", err)
	}
}

func TestQuotationService_UpdateQuotations(t *testing.T) {
	st := mock.NewMockStorage()
	api := mock.NewMockApiQuotationService()
	api.SetRate("EUR", "MXN", 22.0)

	err := st.CreateQuotation(model.Quotation{Name: "EUR_MXN", Rate: 21.5, UpdatedAt: time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("failed to seed quotation: %v", err)
	}

	svc := NewQuotationService(st, mock.NewMockLogger(), api)
	requestID, err := svc.CreateQuotationRequest("EUR_MXN", time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create quotation request: %v", err)
	}

	err = svc.UpdateQuotations()
	if err != nil {
		t.Fatalf("failed to update quotations: %v", err)
	}

	got, err := svc.GetQuotationByName("EUR_MXN")
	if err != nil {
		t.Fatalf("failed to get quotation: %v", err)
	}
	if got.Rate != 22.0 {
		t.Fatalf("unexpected rate: got %v, want 22.0", got.Rate)
	}

	request, err := svc.GetQuotationRequest(requestID)
	if err != nil {
		t.Fatalf("failed to get quotation request: %v", err)
	}
	if !request.Done {
		t.Fatalf("expected request to be done, got %+v", request)
	}

	updates, err := svc.GetQuotationUpdates("EUR_MXN")
	if err != nil {
		t.Fatalf("failed to get quotation updates: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("unexpected updates count: got %d, want 1", len(updates))
	}
	if updates[0].PreviousRate != 21.5 || updates[0].NewRate != 22.0 {
		t.Fatalf("unexpected update: %+v", updates[0])
	}
}

func TestQuotationService_UpdateQuotations_Empty(t *testing.T) {
	svc := NewQuotationService(mock.NewMockStorage(), mock.NewMockLogger(), mock.NewMockApiQuotationService())

	err := svc.UpdateQuotations()
	if err != nil {
		t.Fatalf("failed to update quotations: %v", err)
	}
}

func TestQuotationService_UpdateQuotations_InvalidName(t *testing.T) {
	st := mock.NewMockStorage()
	_, err := st.CreateQuotationRequest(model.QuotationRequest{Name: "INVALID", RequestedAt: time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("failed to seed request: %v", err)
	}

	svc := NewQuotationService(st, mock.NewMockLogger(), mock.NewMockApiQuotationService())
	err = svc.UpdateQuotations()
	if err == nil {
		t.Fatal("expected error for invalid quotation name, got nil")
	}
}