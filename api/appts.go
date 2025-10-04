package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/jcooney/appts/domain"
)

type AppointmentCreator interface {
	Create(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error)
}

func createAppointment(service AppointmentCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &AppointmentRequest{}
		if err := render.Bind(r, req); err != nil {
			err := render.Render(w, r, errInvalidRequest(err))
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
				renderErr := render.Render(w, r, &ErrResponse{
					HTTPStatusCode: http.StatusInternalServerError,
					StatusText:     http.StatusText(http.StatusInternalServerError),
					ErrorText:      "internal server error",
				})
				slog.Error("unknown error creating appointment: %v", "error", err)
				if renderErr != nil {
					slog.Warn("error rendering response: %v", "render error", renderErr)
					return
				}
				return
			}
		}
		render.Status(r, http.StatusCreated)
		_ = render.Render(w, r, NewAppointmentResponse(appointment))
	}
}

type AppointmentRequest struct {
	FirstName string     `json:"firstName" validate:"required,max=50"`
	LastName  string     `json:"lastName" validate:"required,max=50"`
	VisitDate *time.Time `json:"visitDate" validate:"required"` //TODO: is there a better way to handle date-only values in JSON?
}

func (a *AppointmentRequest) Bind(_ *http.Request) error {
	v := validator.New()
	if err := v.Struct(a); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return ve
		}
		return err
	}
	return nil
}

type AppointmentResponse struct {
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	VisitDate *time.Time `json:"visitDate"`
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
