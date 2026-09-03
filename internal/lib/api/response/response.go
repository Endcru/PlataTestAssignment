package response

import (
	"net/http"

	"github.com/go-chi/render"
)

type Response struct {
	Status string `json:"status" example:"OK" enums:"OK,ERROR"`
	Data   any    `json:"data"`
	Error  string `json:"error" example:""`
}

// ErrorResponse documents error payloads returned with non-2xx HTTP status codes.
type ErrorResponse struct {
	Status string `json:"status" example:"ERROR" enums:"ERROR"`
	Data   any    `json:"data" swaggertype:"object"`
	Error  string `json:"error" example:"quotation not found"`
}

const (
	StatusOK    = "OK"
	StatusError = "ERROR"
)

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func ResponseOK(w http.ResponseWriter, r *http.Request, data any) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, Response{
		Status: StatusOK,
		Data:   data,
	})
}

func ResponseError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	render.Status(r, status)
	render.JSON(w, r, Error(msg))
}
