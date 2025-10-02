package publichols

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPublicHolidayGetter_IsPublicHoliday(t *testing.T) {
	tests := []struct {
		name           string
		date           time.Time
		wantPublic     bool
		jsonResp       string
		responseStatus int
		wantErr        error
	}{
		{
			name:           "can get public holiday",
			date:           time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			wantPublic:     true,
			jsonResp:       `[{"date":"2024-12-25"}]`,
			responseStatus: http.StatusOK,
		},
		{
			name:           "not a public holiday",
			date:           time.Date(2024, 12, 26, 0, 0, 0, 0, time.UTC),
			wantPublic:     false,
			jsonResp:       `[{"date":"2024-12-25"}]`,
			responseStatus: http.StatusOK,
		},
		{
			name:           "client error from server",
			date:           time.Date(2024, 12, 26, 0, 0, 0, 0, time.UTC),
			responseStatus: http.StatusBadRequest,
			wantErr:        fmt.Errorf("no response from PublicHolidayPublicHolidaysV3WithResponse: status code: 400"),
		},
		{
			name:           "server error from server",
			date:           time.Date(2024, 12, 26, 0, 0, 0, 0, time.UTC),
			responseStatus: http.StatusInternalServerError,
			wantErr:        fmt.Errorf("no response from PublicHolidayPublicHolidaysV3WithResponse: status code: 500"),
		},
		{
			name:           "invalid json from server",
			date:           time.Date(2024, 12, 26, 0, 0, 0, 0, time.UTC),
			responseStatus: http.StatusOK,
			jsonResp:       `invalid json`,
			wantErr:        fmt.Errorf("PublicHolidayPublicHolidaysV3WithResponse: invalid character 'i' looking for beginning of value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, err := w.Write([]byte(tt.jsonResp))
				require.NoError(t, err)
				require.Equal(t, r.URL.Path, fmt.Sprintf("/api/v3/PublicHolidays/%d/GB", tt.date.Year()))
			}))
			defer server.Close()

			getter, err := NewPublicHolidayGetter(server.URL)
			require.NoError(t, err)
			isPublicHoliday, err := getter.IsPublicHoliday(t.Context(), tt.date)
			if tt.wantErr != nil {
				require.ErrorContains(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantPublic, isPublicHoliday)
			}
		})
	}
}
