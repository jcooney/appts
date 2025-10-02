package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jcooney/appts/api"
	"github.com/jcooney/appts/domain"
	"github.com/stretchr/testify/require"
)

func TestCreateAppointment(t *testing.T) {

	tests := []struct {
		name         string
		request      api.AppointmentRequest
		wantStatus   int
		wantErrBody  *api.ErrResponse
		mockService  api.AppointmentCreator
		wantResponse *api.AppointmentResponse
	}{
		{
			name: "400 when missing first name",
			request: api.AppointmentRequest{
				LastName:  "Doe",
				VisitDate: time.Now(),
			},
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{ErrorText: "missing first name", StatusText: "Bad Request", HTTPStatusCode: 400},
		},
		{
			name: "400 when missing last name",
			request: api.AppointmentRequest{
				FirstName: "John",
				VisitDate: time.Now(),
			},
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{ErrorText: "missing last name", StatusText: "Bad Request", HTTPStatusCode: 400},
		},
		{
			name: "400 when date is missing",
			request: api.AppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
			},
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{ErrorText: "missing visit date", StatusText: "Bad Request", HTTPStatusCode: 400},
		},
		{
			name: "201 when all fields are present",
			request: api.AppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC),
			},
			wantStatus:  http.StatusCreated,
			mockService: success{},
			wantResponse: &api.AppointmentResponse{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "500 when mapping from unsupported service error",
			request: api.AppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC), // Christmas, assuming it's a public holiday
			},
			wantStatus:  http.StatusInternalServerError,
			wantErrBody: &api.ErrResponse{ErrorText: "unhandled error", StatusText: "Internal Server Error", HTTPStatusCode: 500},
			mockService: unhandlerError{},
		},
		{
			name: "409 when appoint is already booked for the day",
			request: api.AppointmentRequest{
				FirstName: "Jane",
				LastName:  "Doe",
				VisitDate: time.Date(2024, 7, 4, 0, 0, 0, 0, time.UTC), // Assuming July 4th is already booked
			},
			wantStatus:  http.StatusConflict,
			wantErrBody: &api.ErrResponse{ErrorText: "appointment date already taken", StatusText: "Conflict", HTTPStatusCode: 409},
			mockService: dateTaken{},
		},
		{
			name: "400 when appoint is on a public holiday",
			request: api.AppointmentRequest{
				FirstName: "Jane",
				LastName:  "Doe",
				VisitDate: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC), // Christmas
			},
			wantStatus:  http.StatusBadRequest,
			mockService: publicHoliday{},
			wantErrBody: &api.ErrResponse{ErrorText: "cannot book appointment on public holiday", StatusText: "Bad Request", HTTPStatusCode: 400},
		},
		{
			name: "400 when appoint is in the past",
			request: api.AppointmentRequest{
				FirstName: "Jane",
				LastName:  "Doe",
				VisitDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), // A past date
			},
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{ErrorText: "cannot book appointment in the past", StatusText: "Bad Request", HTTPStatusCode: 400},
			mockService: dateInPast{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(api.CreateAppointmentFunc(tt.mockService))
			defer ts.Close()

			marshal, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", ts.URL+"/appts/", bytes.NewBuffer(marshal))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Logf("error closing response body: %v", err)
				}
			}(resp.Body)

			require.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus == http.StatusCreated {
				all, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				var got api.AppointmentResponse
				err = json.Unmarshal(all, &got)
				require.NoError(t, err)
				require.Equal(t, tt.wantResponse.FirstName, got.FirstName)
				require.Equal(t, tt.wantResponse.LastName, got.LastName)
				require.True(t, tt.wantResponse.VisitDate.Equal(got.VisitDate))
			}
			if tt.wantErrBody != nil {
				all, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				var gotErr api.ErrResponse
				err = json.Unmarshal(all, &gotErr)
				require.NoError(t, err)
				require.Equal(t, *tt.wantErrBody, gotErr)
			}
		})
	}
}

type success struct{}

func (s success) Create(_ context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	return appt, nil
}

type unhandlerError struct{}

func (u unhandlerError) Create(_ context.Context, _ *domain.Appointment) (*domain.Appointment, error) {
	return nil, errors.New("unhandled error")
}

type dateTaken struct{}

func (d dateTaken) Create(_ context.Context, _ *domain.Appointment) (*domain.Appointment, error) {
	return nil, domain.ErrAppointmentDateTaken
}

type publicHoliday struct{}

func (p publicHoliday) Create(_ context.Context, _ *domain.Appointment) (*domain.Appointment, error) {
	return nil, domain.ErrAppointmentOnPublicHoliday
}

type dateInPast struct{}

func (d dateInPast) Create(_ context.Context, _ *domain.Appointment) (*domain.Appointment, error) {
	return nil, domain.ErrAppointmentInPast
}
