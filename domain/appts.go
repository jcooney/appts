package domain

import (
	"context"
	"fmt"
	"time"
)

var ErrAppointmentOnPublicHoliday = fmt.Errorf("cannot book appointment on public holiday")
var ErrAppointmentDateTaken = fmt.Errorf("appointment date already taken")
var ErrAppointmentInPast = fmt.Errorf("cannot book appointment in the past")

type Appointment struct {
	FirstName string
	LastName  string
	VisitDate time.Time
}

type AppointmentPersistorRepository interface {
	CreateAppointment(ctx context.Context, appt *Appointment) (*Appointment, error)
}

type PublicHolidayChecker interface {
	IsPublicHoliday(context.Context, time.Time) (bool, error)
}

func NewAppointment(firstName string, lastName string, date time.Time) *Appointment {
	return &Appointment{
		FirstName: firstName,
		LastName:  lastName,
		VisitDate: date,
	}
}

type AppointmentCreatorService struct {
	repo    AppointmentPersistorRepository
	checker PublicHolidayChecker
}

func NewAppointmentCreatorService(repo AppointmentPersistorRepository, checker PublicHolidayChecker) *AppointmentCreatorService {
	return &AppointmentCreatorService{
		repo:    repo,
		checker: checker,
	}
}

func (s *AppointmentCreatorService) Create(ctx context.Context, appt *Appointment) (*Appointment, error) {
	if appt == nil {
		return nil, fmt.Errorf("appointment is nil")
	}
	if appt.VisitDate.Before(time.Now()) {
		return nil, ErrAppointmentInPast
	}
	ok, err := s.checker.IsPublicHoliday(ctx, appt.VisitDate)
	if err != nil {
		return nil, fmt.Errorf("isPublicHoliday: %w", err)
	}

	if ok {
		return nil, ErrAppointmentOnPublicHoliday
	}

	save, err := s.repo.CreateAppointment(ctx, appt)
	if err != nil {
		return nil, fmt.Errorf("save appointment: %w", err)
	}
	return save, nil
}
