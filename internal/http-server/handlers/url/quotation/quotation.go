package quotation

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	models "github.com/Endcru/PlataTestAssignment/internal/models"
	response "github.com/Endcru/PlataTestAssignment/internal/lib/api/response"
	quotationService "github.com/Endcru/PlataTestAssignment/internal/service/quotation"
	storage "github.com/Endcru/PlataTestAssignment/internal/storage"
)

// ResponseQuotationSuccess wraps quotation data in a standard API response.
type ResponseQuotationSuccess struct {
	Status string           `json:"status" example:"OK"`
	Data   models.Quotation `json:"data"`
	Error  string           `json:"error" example:""`
}

// GetQuotation returns the latest quotation value by currency pair name.
//
// @Summary      Get latest quotation
// @Description  Returns the current rate and update time for a currency pair (e.g. EUR_MXN).
// @Tags         quotation
// @Produce      json
// @Param        name path string true "Currency pair code" example(EUR_MXN)
// @Success      200 {object} ResponseQuotationSuccess
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /quotation/{name} [get]
func GetQuotation(log *slog.Logger, quotationService quotationService.QuotationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.quotation.GetQuotation"

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

		log.Info("quotation found", slog.String("quotation_name", quotationName), slog.Float64("rate", quotation.Rate), slog.Time("updated_at", quotation.UpdatedAt))

		response.ResponseOK(w, r, quotation)
	}
}
