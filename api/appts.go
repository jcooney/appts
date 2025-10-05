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

type AppointmentRequest struct {
	FirstName string     `json:"firstName" validate:"required,max=50"`
	LastName  string     `json:"lastName" validate:"required,max=50"`
	VisitDate *VisitDate `json:"visitDate" validate:"required"`
}

type VisitDate time.Time

func (v *VisitDate) UnmarshalJSON(b []byte) error {
	t, err := time.Parse(time.DateOnly, string(b))
	if err != nil {
		return err
	}
	*v = VisitDate(t)
	return nil
}

func (v *VisitDate) Time() *time.Time {
	t := time.Time(*v)
	return &t
}

type AppointmentResponse struct {
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	VisitDate *time.Time `json:"visitDate"`
}

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

		appointment, err := service.Create(r.Context(), domain.NewAppointment(req.FirstName, req.LastName, req.VisitDate.Time()))
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
