package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
			appointmentToSave:    NewAppointment("first", "last", time.Now().Add(time.Hour)),
			publicHolidayChecker: publicHolidayError{},
			wantErr:              errors.New("isPublicHoliday: some error"),
		},
		{
			name:                 "return error when appt is on public holiday",
			appointmentToSave:    NewAppointment("first", "last", time.Now().Add(time.Hour)),
			publicHolidayChecker: publicHolidayCheckerIsPublicHoliday{},
			wantErr:              ErrAppointmentOnPublicHoliday,
		},
		{
			name:                 "return error from repository on save",
			appointmentToSave:    NewAppointment("first", "last", time.Now().Add(time.Hour)),
			publicHolidayChecker: publicHolidayCheckerSuccess{},
			appointmentPersistor: appointmentPesistorError{},
			wantErr:              errors.New("save appointment: some error"),
		},
		{
			name:                 "success",
			appointmentToSave:    NewAppointment("first", "last", time.Now().Add(time.Hour).UTC()),
			publicHolidayChecker: publicHolidayCheckerSuccess{},
			appointmentPersistor: appointmentPersistorSuccess{},
			want:                 NewAppointment("first", "last", time.Now().Add(time.Hour).UTC()),
		},
		{
			name:              "do not allow appointment in the past",
			appointmentToSave: NewAppointment("first", "last", time.Now().Add(-time.Nanosecond)),
			wantErr:           ErrAppointmentInPast,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unitUnderTest := NewAppointmentCreatorService(tt.appointmentPersistor, tt.publicHolidayChecker)
			got, err := unitUnderTest.Create(t.Context(), tt.appointmentToSave)
			if tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
			}
			require.Equal(t, tt.want, got)
		})
	}
}

type publicHolidayError struct{}

func (p publicHolidayError) IsPublicHoliday(_ context.Context, _ time.Time) (bool, error) {
	return false, fmt.Errorf("some error")
}

type publicHolidayCheckerIsPublicHoliday struct{}

func (p publicHolidayCheckerIsPublicHoliday) IsPublicHoliday(_ context.Context, _ time.Time) (bool, error) {
	return true, nil
}

type publicHolidayCheckerSuccess struct{}

func (p publicHolidayCheckerSuccess) IsPublicHoliday(_ context.Context, _ time.Time) (bool, error) {
	return false, nil
}

type appointmentPesistorError struct{}

func (a appointmentPesistorError) Save(_ context.Context, _ *Appointment) (*Appointment, error) {
	return nil, fmt.Errorf("some error")
}

type appointmentPersistorSuccess struct{}

func (a appointmentPersistorSuccess) Save(_ context.Context, appt *Appointment) (*Appointment, error) {
	return appt, nil
}
