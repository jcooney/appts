package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func ChiHandler(service AppointmentCreator) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Post("/appts/", CreateAppointmentFunc(service))

	return r
}

func CreateAppointmentFunc(service AppointmentCreator) http.HandlerFunc {
	return createAppointment(service)
}
