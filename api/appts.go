package api

import (
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

// CreateAppointment persists the posted Article and returns it
// back to the client as an acknowledgement.
func CreateAppointment(w http.ResponseWriter, r *http.Request) {
	req := &AppointmentRequest{}
	if err := render.Bind(r, req); err != nil {
		err := render.Render(w, r, ErrInvalidRequest(err))
		if err != nil {
			return
		}
		return
	}

	// make a service call to persist the appointment here
	stub := domain.NewAppointment("stub", "stub", time.Now())

	render.Status(r, http.StatusCreated)
	_ = render.Render(w, r, NewAppointmentResponse(stub))
}

type AppointmentResponse struct {
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	VisitDate time.Time `json:"visitDate"`
}

func (a AppointmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// TODO unsure what to do here
	return nil
}

func NewAppointmentResponse(stub *domain.Appointment) AppointmentResponse {
	return AppointmentResponse{
		FirstName: stub.FirstName,
		LastName:  stub.LastName,
		VisitDate: stub.VisitDate,
	}
}

func (a *AppointmentRequest) Bind(r *http.Request) error {
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
