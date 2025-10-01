package domain

import (
	"context"
	"fmt"
	"time"
)

var ErrAppointmentOnPublicHoliday = fmt.Errorf("cannot book appointment on public holiday")

type Appointment struct {
	FirstName string
	LastName  string
	VisitDate time.Time
}

type AppointmentPersistorRepository interface {
	Save(context.Context, *Appointment) error
}

type PublicHolidayChecker interface {
	IsPublicHoliday(context.Context, time.Time) (bool, error)
}

func NewAppointment(name string, name2 string, date time.Time) *Appointment {
	return &Appointment{
		FirstName: name,
		LastName:  name2,
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

func (s *AppointmentCreatorService) Create(ctx context.Context, appt *Appointment) error {
	ok, err := s.checker.IsPublicHoliday(ctx, appt.VisitDate)
	if err != nil {
		return fmt.Errorf("isPublicHoliday: %w", err)
	}

	if ok {
		return ErrAppointmentOnPublicHoliday
	}

	return s.repo.Save(ctx, appt)
}
