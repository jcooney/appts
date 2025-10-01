package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jcooney/appts/api"
	"github.com/stretchr/testify/require"
)

func TestCreateAppointment(t *testing.T) {

	tests := []struct {
		name        string
		request     api.AppointmentRequest
		wantStatus  int
		wantErrBody *api.ErrResponse
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
				VisitDate: time.Now(),
			},
			wantStatus:  http.StatusCreated,
			wantErrBody: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(api.ChiHandler())
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
