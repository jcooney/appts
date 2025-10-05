package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jcooney/appts/domain"
)

var errorMap = map[error]int{
	domain.ErrAppointmentOnPublicHoliday: http.StatusBadRequest,
	domain.ErrAppointmentDateTaken:       http.StatusConflict,
	domain.ErrAppointmentInPast:          http.StatusBadRequest,
}

func errInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     http.StatusText(http.StatusBadRequest),
		ErrorText:      err.Error(),
	}
}

func errInternalServerError() render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     http.StatusText(http.StatusInternalServerError),
		ErrorText:      "internal server error",
	}
}

type ErrResponse struct {
	HTTPStatusCode int    `json:"code"`             // http response status code
	StatusText     string `json:"status,omitempty"` // user-level status message
	ErrorText      string `json:"error,omitempty"`  // application-level error message, for debugging
}

func (e *ErrResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}
