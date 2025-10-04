package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jcooney/appts/domain"
	"github.com/jcooney/appts/repository/gen"
)

type Repository struct {
	queries *sqlcappts.Queries
}

func NewRepository(db sqlcappts.DBTX) *Repository {
	return &Repository{queries: sqlcappts.New(db)}
}

func (r *Repository) CreateAppointment(ctx context.Context, appt *domain.Appointment) (*domain.Appointment, error) {
	appointmentRow, err := r.queries.CreateDailyAppointment(ctx, sqlcappts.CreateDailyAppointmentParams{
		FirstName:       appt.FirstName,
		LastName:        appt.LastName,
		AppointmentDate: pgtype.Timestamptz{Time: *appt.VisitDate, Valid: true},
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, domain.ErrAppointmentDateTaken
			}
		}
		return nil, fmt.Errorf("create daily appointment: %w", err)
	}

	return domain.NewAppointment(appointmentRow.FirstName, appointmentRow.LastName, &appointmentRow.AppointmentDate.Time), nil
}
