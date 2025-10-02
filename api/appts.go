package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/jcooney/appts/domain"
)

type AppointmentRequest struct {
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	VisitDate time.Time `json:"visitDate"` //TODO: is there a better way to handle date-only values in JSON?
}

type AppointmentCreator interface {
	Create(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error)
}

var errorMap = map[error]int{
	domain.ErrAppointmentOnPublicHoliday: http.StatusBadRequest,
	domain.ErrAppointmentDateTaken:       http.StatusConflict,
	domain.ErrAppointmentInPast:          http.StatusBadRequest,
}

// CreateAppointment persists the posted Article and returns it
// back to the client as an acknowledgement.
func CreateAppointment(service AppointmentCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &AppointmentRequest{}
		if err := render.Bind(r, req); err != nil {
			err := render.Render(w, r, ErrInvalidRequest(err))
			if err != nil {
				return
			}
			return
		}

		appointment, err := service.Create(r.Context(), domain.NewAppointment(req.FirstName, req.LastName, req.VisitDate))
		if err != nil {
			if code, ok := errorMap[err]; ok {
				renderErr := render.Render(w, r, &ErrResponse{
					HTTPStatusCode: code,
					StatusText:     http.StatusText(code),
					ErrorText:      err.Error(),
				})
				if renderErr != nil {
					return
				}
				return
			} else {
				// unhandled error
				renderErr := render.Render(w, r, &ErrResponse{
					HTTPStatusCode: http.StatusInternalServerError,
					StatusText:     http.StatusText(http.StatusInternalServerError),
					ErrorText:      err.Error(),
				})
				if renderErr != nil {
					return
				}
				return
			}
		}
		render.Status(r, http.StatusCreated)
		_ = render.Render(w, r, NewAppointmentResponse(appointment))
	}
}

type AppointmentResponse struct {
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	VisitDate time.Time `json:"visitDate"`
}

func (a AppointmentResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	// TODO unsure what to do here
	return nil
}

func NewAppointmentResponse(appointment *domain.Appointment) AppointmentResponse {
	return AppointmentResponse{
		FirstName: appointment.FirstName,
		LastName:  appointment.LastName,
		VisitDate: appointment.VisitDate,
	}
}

func (a *AppointmentRequest) Bind(_ *http.Request) error {
	if a.FirstName == "" {
		return errors.New("missing first name")
	}

	if a.LastName == "" {
		return errors.New("missing last name")
	}

	if a.VisitDate.IsZero() {
		return errors.New("missing visit date")
	}

	return nil
}
