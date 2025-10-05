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
	"k8s.io/utils/ptr"
)

func TestCreateAppointment(t *testing.T) {
	now := time.Now()
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
				VisitDate: (*api.VisitDate)(&now),
			},
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{ErrorText: "Key: 'AppointmentRequest.FirstName' Error:Field validation for 'FirstName' failed on the 'required' tag", StatusText: "Bad Request", HTTPStatusCode: 400},
		},
		{
			name: "400 when missing last name",
			request: api.AppointmentRequest{
				FirstName: "John",
				VisitDate: (*api.VisitDate)(&now),
			},
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{ErrorText: "Key: 'AppointmentRequest.LastName' Error:Field validation for 'LastName' failed on the 'required' tag", StatusText: "Bad Request", HTTPStatusCode: 400},
		},
		{
			name: "400 when date is missing",
			request: api.AppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
			},
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{ErrorText: "Key: 'AppointmentRequest.VisitDate' Error:Field validation for 'VisitDate' failed on the 'required' tag", StatusText: "Bad Request", HTTPStatusCode: 400},
		},
		{
			name: "201 when all fields are present",
			request: api.AppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: (*api.VisitDate)(ptr.To(time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC))),
			},
			wantStatus:  http.StatusCreated,
			mockService: success{},
			wantResponse: &api.AppointmentResponse{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: (*api.VisitDate)(ptr.To(time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC))),
			},
		},
		{
			name: "500 when mapping from unsupported service error",
			request: api.AppointmentRequest{
				FirstName: "John",
				LastName:  "Doe",
				VisitDate: (*api.VisitDate)(ptr.To(time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC))), // Christmas, assuming it's a public holiday
			},
			wantStatus:  http.StatusInternalServerError,
			wantErrBody: &api.ErrResponse{ErrorText: "internal server error", StatusText: "Internal Server Error", HTTPStatusCode: 500},
			mockService: unhandlerError{},
		},
		{
			name: "409 when appoint is already booked for the day",
			request: api.AppointmentRequest{
				FirstName: "Jane",
				LastName:  "Doe",
				VisitDate: (*api.VisitDate)(ptr.To(time.Date(2024, 7, 4, 0, 0, 0, 0, time.UTC))), // Assuming July 4th is already booked
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
				VisitDate: (*api.VisitDate)(ptr.To(time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC))), // Christmas
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
				VisitDate: (*api.VisitDate)(ptr.To(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))), // A past date
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
				require.Equal(t, *tt.wantResponse.VisitDate, *got.VisitDate)
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

func TestCreateAppointmentJson(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantStatus  int
		wantErrBody *api.ErrResponse
	}{
		{
			name:        "400 when invalid JSON (malformed)",
			requestBody: `{"firstName": "John", "lastName": "Doe", "visitDate": "2024-07-15T00:00:00Z`,
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{ErrorText: "unexpected EOF", StatusText: "Bad Request", HTTPStatusCode: 400},
		},
		{
			name:        "400 when invalid date format",
			requestBody: `{"firstName": "John", "lastName": "Doe", "visitDate": "15-07-2024"}`, // Incorrect date format
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{ErrorText: "parsing time \"15-07-2024\" as \"2006-01-02\": cannot parse \"15-07-2024\" as \"2006\"", StatusText: "Bad Request", HTTPStatusCode: 400},
		},
		{
			name:        "400 when invalid date format 2",
			requestBody: `{"firstName": "John", "lastName": "Doe", "visitDate": "jnoefinefnioefwinoefwino"}`,
			wantStatus:  http.StatusBadRequest,
			wantErrBody: &api.ErrResponse{HTTPStatusCode: 400, StatusText: "Bad Request", ErrorText: "parsing time \"jnoefinefnioefwinoefwino\" as \"2006-01-02\": cannot parse \"jnoefinefnioefwinoefwino\" as \"2006\""},
		},
		{
			name: "200 when valid JSON",
			requestBody: `{
				"firstName": "John",
				"lastName": "Doe",
				"visitDate": "2024-07-15"
			}`,
			wantStatus: http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(api.CreateAppointmentFunc(success{}))
			defer ts.Close()

			req, err := http.NewRequest("POST", ts.URL+"/appts/", bytes.NewBufferString(tt.requestBody))
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

			if tt.wantErrBody == nil && tt.wantStatus != resp.StatusCode {
				// log response body
				all, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				t.Logf("response body: %s", string(all))
			}
			require.Equal(t, tt.wantStatus, resp.StatusCode)
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
