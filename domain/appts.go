package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/utils/ptr"
)

var ErrAppointmentOnPublicHoliday = fmt.Errorf("cannot book appointment on public holiday")
var ErrAppointmentDateTaken = fmt.Errorf("appointment date already taken")
var ErrAppointmentInPast = fmt.Errorf("cannot book appointment in the past")

type Appointment struct {
	FirstName string
	LastName  string
	VisitDate *time.Time
}

type AppointmentPersistorRepository interface {
	CreateAppointment(ctx context.Context, appt *Appointment) (*Appointment, error)
}

type PublicHolidayChecker interface {
	IsPublicHoliday(context.Context, *time.Time) (bool, error)
}

func NewAppointment(firstName string, lastName string, date *time.Time) *Appointment {
	return &Appointment{
		FirstName: firstName,
		LastName:  lastName,
		VisitDate: ptr.To(time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)),
	}
}

type AppointmentCreatorService struct {
	repo    AppointmentPersistorRepository
	checker PublicHolidayChecker
	nowFunc func() time.Time
}

func NewAppointmentCreatorService(repo AppointmentPersistorRepository, checker PublicHolidayChecker, nowFunc func() time.Time) *AppointmentCreatorService {
	return &AppointmentCreatorService{
		repo:    repo,
		checker: checker,
		nowFunc: nowFunc,
	}
}

func (s *AppointmentCreatorService) Create(ctx context.Context, appt *Appointment) (*Appointment, error) {
	if appt == nil {
		return nil, fmt.Errorf("appointment is nil")
	}
	if appt.VisitDate.Before(s.nowFunc()) {
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
		if errors.Is(err, ErrAppointmentDateTaken) {
			return nil, ErrAppointmentDateTaken // bubble up to allow http error handling
		}
		return nil, fmt.Errorf("save appointment: %w", err)
	}
	return save, nil
}
