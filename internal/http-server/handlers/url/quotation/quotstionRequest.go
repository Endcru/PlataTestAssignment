package quotation

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	models "github.com/Endcru/PlataTestAssignment/internal/models"
	response "github.com/Endcru/PlataTestAssignment/internal/lib/api/response"
	quotationService "github.com/Endcru/PlataTestAssignment/internal/service/quotation"
	storage "github.com/Endcru/PlataTestAssignment/internal/storage"
)

// ResponseQuotationFromUpdateIDSuccess wraps quotation data returned by update request id.
type ResponseQuotationFromUpdateIDSuccess struct {
	Status string           `json:"status" example:"OK"`
	Data   models.Quotation `json:"data"`
	Error  string           `json:"error" example:""`
}

// GetQuotationFromRequestUpdateID returns quotation by update request id.
//
// @Summary      Get quotation by update id
// @Description  Returns the current rate and update time for a quotation by update request id.
// @Tags         quotation
// @Produce      json
// @Param        update_id path int true "Update request id" example(1)
// @Success      200 {object} ResponseQuotationFromUpdateIDSuccess
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /quotation/request/{update_id} [get]
func GetQuotationFromRequestUpdateID(log *slog.Logger, quotationService quotationService.QuotationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.quotation.GetQuotationFromRequestUpdateID"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		updateID := chi.URLParam(r, "update_id")

		if updateID == "" {
			log.Info("update id is empty")
			response.ResponseError(w, r, http.StatusBadRequest, "invalid update id")
			return
		}

		updateIDInt, err := strconv.Atoi(updateID)
		if err != nil {
			log.Error("failed to convert update id to int", slog.String("error", err.Error()))
			response.ResponseError(w, r, http.StatusBadRequest, "failed to convert update id to int")
			return
		}

		quotation, err := quotationService.GetQuotationByRequestUpdateID(updateIDInt)
		if errors.Is(err, storage.ErrQuotationRequestNotDone) {
			log.Info("quotation request not done", slog.String("update_id", updateID))
			response.ResponseError(w, r, http.StatusConflict, "quotation request not done")
			return
		}
		if errors.Is(err, storage.ErrQuotationNotFound) || errors.Is(err, storage.ErrQuotationRequestNotFound) {
			log.Info("quotation not found", slog.String("update_id", updateID))
			response.ResponseError(w, r, http.StatusNotFound, "quotation not found")
			return
		}
		if err != nil {
			log.Error("failed to get quotation", slog.String("error", err.Error()))
			response.ResponseError(w, r, http.StatusInternalServerError, "failed to get quotation")
			return
		}

		log.Info("quotation found from Request", slog.String("Request Update ID", updateID), slog.Float64("rate", quotation.Rate), slog.Time("updated_at", quotation.UpdatedAt))

		response.ResponseOK(w, r, quotation)
	}
}
