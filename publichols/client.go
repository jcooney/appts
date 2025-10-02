package publichols

import (
	"context"
	"fmt"
	"time"
)

type PublicHolidayGetter struct {
	client *ClientWithResponses
}

func NewPublicHolidayGetter(host string) (*PublicHolidayGetter, error) {
	client, err := NewClientWithResponses(host)
	if err != nil {
		return nil, fmt.Errorf("NewClient: %w", err)
	}
	return &PublicHolidayGetter{
		client: client,
	}, nil
}

func (g *PublicHolidayGetter) IsPublicHoliday(ctx context.Context, date time.Time) (bool, error) {
	publicHolidaysV3, err := g.client.PublicHolidayPublicHolidaysV3WithResponse(ctx, int32(date.Year()), "GB")
	if err != nil {
		return false, fmt.Errorf("PublicHolidayPublicHolidaysV3WithResponse: %w", err)
	}

	resp := publicHolidaysV3.JSON200
	if resp == nil {
		return false, fmt.Errorf("no response from PublicHolidayPublicHolidaysV3WithResponse: status code: %d", publicHolidaysV3.StatusCode())
	}

	for i := range *resp {
		if (*resp)[i].Date.Format(time.DateOnly) == date.Format(time.DateOnly) {
			return true, nil
		}
	}
	return false, nil
}
