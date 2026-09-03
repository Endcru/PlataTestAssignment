package quotation

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	response "github.com/Endcru/PlataTestAssignment/internal/lib/api/response"
	models "github.com/Endcru/PlataTestAssignment/internal/models"
	quotationService "github.com/Endcru/PlataTestAssignment/internal/service/quotation"
	storage "github.com/Endcru/PlataTestAssignment/internal/storage"
)

type RequestQuotationUpdate struct {
	QuotationName string `json:"quotation_name" validate:"required"`
}

// ResponseQuotationUpdateSuccess wraps update request id in a standard API response.
type ResponseQuotationUpdateSuccess struct {
	Status string                      `json:"status" example:"OK"`
	Data   ResponseQuotationUpdateData `json:"data"`
	Error  string                      `json:"error" example:""`
}

// ResponseQuotationUpdateData contains the id of the created update request.
type ResponseQuotationUpdateData struct {
	QuotationRequestID int `json:"quotation_request_id" example:"1"`
}

// ResponseQuotationUpdatesSuccess wraps quotation update history in a standard API response.
type ResponseQuotationUpdatesSuccess struct {
	Status string                  `json:"status" example:"OK"`
	Data   []models.QuotationUpdate `json:"data"`
	Error  string                  `json:"error" example:""`
}

// QuotationUpdate creates an asynchronous quotation update request.
//
// @Summary      Request quotation update
// @Description  Creates a background update request and returns its id. The currency pair is taken from the JSON body (`quotation_name`); the `{name}` path segment is not used.
// @Tags         quotation
// @Accept       json
// @Produce      json
// @Param        request body RequestQuotationUpdate true "Update request payload"
// @Success      200 {object} ResponseQuotationUpdateSuccess
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /quotation/{name}/update [post]
func QuotationUpdate(log *slog.Logger, quotationService quotationService.QuotationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.quotation.QuotationUpdate"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		var request RequestQuotationUpdate

		err := render.DecodeJSON(r.Body, &request)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			response.ResponseError(w, r, http.StatusBadRequest, "empty request")
			return
		}
		if err != nil {
			log.Error("failed to decode request body", slog.String("error", err.Error()))
			response.ResponseError(w, r, http.StatusBadRequest, "failed to decode request")
			return
		}

		if err := validator.New().Struct(request); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", slog.String("error", err.Error()))
			response.ResponseError(w, r, http.StatusBadRequest, validateErr.Error())
			return
		}

		quotationName := request.QuotationName

		if !quotationService.ValidateQuotationName(quotationName) {
			log.Error("invalid quotation name")
			response.ResponseError(w, r, http.StatusBadRequest, "invalid quotation name")
			return
		}

		quotation, err := quotationService.GetQuotationByName(quotationName)
		if errors.Is(err, storage.ErrQuotationNotFound) {
			log.Info("quotation not found", slog.String("quotation_name", quotationName))
			response.ResponseError(w, r, http.StatusNotFound, "quotation not found")
			return
		}

		if err != nil {
			log.Error("failed to get quotation", slog.String("error", err.Error()))
			response.ResponseError(w, r, http.StatusInternalServerError, "failed to get quotation")
			return
		}

		quotationRequestID, err := quotationService.CreateQuotationRequest(quotationName, time.Now())
		if err != nil {
			log.Error("failed to create quotation request", slog.String("error", err.Error()))
			response.ResponseError(w, r, http.StatusInternalServerError, "failed to create quotation request")
			return
		}

		log.Info("quotation found", slog.String("quotation_name", quotationName), slog.Float64("rate", quotation.Rate), slog.Time("updated_at", quotation.UpdatedAt))

		response.ResponseOK(w, r, ResponseQuotationUpdateData{
			QuotationRequestID: quotationRequestID,
		})
	}
}

// GetQuotation returns the latest quotation value by currency pair name.
//
// @Summary      Get all quotation updates
// @Description  Returns all quotation updates for a currency pair (e.g. EUR_MXN).
// @Tags         quotation
// @Produce      json
// @Param        name path string true "Currency pair code" example(EUR_MXN)
// @Success      200 {object} ResponseQuotationUpdatesSuccess
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /quotation/{name}/updates [get]
func GetQuotationUpdates(log *slog.Logger, quotationService quotationService.QuotationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.quotation.GetQuotationUpdates"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		quotationName := chi.URLParam(r, "name")

		if quotationName == "" {
			log.Info("quotation name is empty")
			response.ResponseError(w, r, http.StatusBadRequest, "invalid quotation name")
			return
		}

		if !quotationService.ValidateQuotationName(quotationName) {
			log.Info("invalid quotation name", slog.String("quotation_name", quotationName))
			response.ResponseError(w, r, http.StatusBadRequest, "invalid quotation name")
			return
		}

		quotationUpdates, err := quotationService.GetQuotationUpdates(quotationName)
		if errors.Is(err, storage.ErrQuotationNotFound) {
			log.Info("quotation not found", slog.String("quotation_name", quotationName))
			response.ResponseError(w, r, http.StatusNotFound, "quotation not found")
			return
		}
		if err != nil {
			log.Error("failed to get quotation", slog.String("error", err.Error()))
			response.ResponseError(w, r, http.StatusInternalServerError, "failed to get quotation")
			return
		}

		log.Info("quotation updates found", slog.String("quotation_name", quotationName), slog.Int("number_of_updates", len(quotationUpdates)))

		response.ResponseOK(w, r, quotationUpdates)
	}
}