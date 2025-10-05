package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestAppointmentCreatorService_Create(t *testing.T) {
	tests := []struct {
		name                 string
		appointmentToSave    *Appointment
		want                 *Appointment
		wantErr              error
		appointmentPersistor AppointmentPersistorRepository
		publicHolidayChecker PublicHolidayChecker
	}{
		{
			name:              "return error when appt is nil",
			appointmentToSave: nil,
			wantErr:           errors.New("appointment is nil"),
		},
		{
			name:                 "return error from public holiday checker",
			appointmentToSave:    NewAppointment("first", "last", ptr.To(fixedTimeFunc().Add(time.Hour).UTC())),
			publicHolidayChecker: publicHolidayError{},
			wantErr:              errors.New("isPublicHoliday: some error"),
		},
		{
			name:                 "return error when appt is on public holiday",
			appointmentToSave:    NewAppointment("first", "last", ptr.To(fixedTimeFunc().Add(time.Hour).UTC())),
			publicHolidayChecker: publicHolidayCheckerIsPublicHoliday{},
			wantErr:              ErrAppointmentOnPublicHoliday,
		},
		{
			name:                 "return error from repository on save",
			appointmentToSave:    NewAppointment("first", "last", ptr.To(fixedTimeFunc().Add(time.Hour).UTC())),
			publicHolidayChecker: publicHolidayCheckerSuccess{},
			appointmentPersistor: appointmentPesistorError{},
			wantErr:              errors.New("save appointment: some error"),
		},
		{
			name:                 "bubble up conflict error",
			appointmentToSave:    NewAppointment("first", "last", ptr.To(fixedTimeFunc().Add(time.Hour).UTC())),
			publicHolidayChecker: publicHolidayCheckerSuccess{},
			appointmentPersistor: conflict{},
			wantErr:              ErrAppointmentDateTaken,
		},
		{
			name:                 "success",
			appointmentToSave:    NewAppointment("first", "last", ptr.To(fixedTimeFunc().Add(time.Hour).UTC())),
			publicHolidayChecker: publicHolidayCheckerSuccess{},
			appointmentPersistor: appointmentPersistorSuccess{},
			want:                 NewAppointment("first", "last", ptr.To(fixedTimeFunc().Add(time.Hour).UTC())),
		},
		{
			name:              "do not allow appointment in the past",
			appointmentToSave: NewAppointment("first", "last", ptr.To(fixedTimeFunc().Add(-time.Nanosecond))),
			wantErr:           ErrAppointmentInPast,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unitUnderTest := NewAppointmentCreatorService(tt.appointmentPersistor, tt.publicHolidayChecker, fixedTimeFunc)
			got, err := unitUnderTest.Create(t.Context(), tt.appointmentToSave)
			if tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func fixedTimeFunc() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}

type publicHolidayError struct{}

func (p publicHolidayError) IsPublicHoliday(_ context.Context, _ *time.Time) (bool, error) {
	return false, fmt.Errorf("some error")
}

type publicHolidayCheckerIsPublicHoliday struct{}

func (p publicHolidayCheckerIsPublicHoliday) IsPublicHoliday(_ context.Context, _ *time.Time) (bool, error) {
	return true, nil
}

type publicHolidayCheckerSuccess struct{}

func (p publicHolidayCheckerSuccess) IsPublicHoliday(_ context.Context, _ *time.Time) (bool, error) {
	return false, nil
}

type appointmentPesistorError struct{}

func (a appointmentPesistorError) CreateAppointment(_ context.Context, _ *Appointment) (*Appointment, error) {
	return nil, fmt.Errorf("some error")
}

type appointmentPersistorSuccess struct{}

func (a appointmentPersistorSuccess) CreateAppointment(_ context.Context, appt *Appointment) (*Appointment, error) {
	return appt, nil
}

type conflict struct{}

func (c conflict) CreateAppointment(_ context.Context, _ *Appointment) (*Appointment, error) {
	return nil, ErrAppointmentDateTaken
}
