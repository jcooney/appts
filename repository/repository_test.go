package repository_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jcooney/appts/domain"
	"github.com/jcooney/appts/repository"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"k8s.io/utils/ptr"
)

func TestInsertAppointment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	postgresContainer, err := postgres.Run(t.Context(),
		"postgres:16-alpine",
		postgres.WithDatabase("tabeo"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("dbPassword"),
		postgres.BasicWaitStrategies(),
	)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	})
	require.NoError(t, err)

	connectionString, err := postgresContainer.ConnectionString(t.Context())
	require.NoError(t, err)

	m, err := migrate.New("file://../schema/ddl", connectionString+"&sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, m.Up())

	dbpool, err := pgxpool.New(t.Context(), connectionString)
	defer dbpool.Close() //nolint:staticcheck // we require no error below
	require.NoError(t, err)
	tx, err := dbpool.Begin(t.Context())
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Printf("failed to rollback transaction: %s", err)
		}
	}(tx, t.Context())
	require.NoError(t, err)

	underTest := repository.NewRepository(tx)
	appointment, err := underTest.CreateAppointment(t.Context(),
		&domain.Appointment{
			FirstName: "first",
			LastName:  "last",
			VisitDate: ptr.To(time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)),
		})
	require.NoError(t, err)
	require.Equal(t, "first", appointment.FirstName)
	require.Equal(t, "last", appointment.LastName)
	require.Equal(t, time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC), appointment.VisitDate.UTC())
}

func TestInsertDuplicateAppointmentError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	postgresContainer, err := postgres.Run(t.Context(),
		"postgres:16-alpine",
		postgres.WithDatabase("tabeo"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("dbPassword"),
		postgres.BasicWaitStrategies(),
	)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	})
	require.NoError(t, err)

	connectionString, err := postgresContainer.ConnectionString(t.Context())
	require.NoError(t, err)

	m, err := migrate.New("file://../schema/ddl", connectionString+"&sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, m.Up())

	dbpool, err := pgxpool.New(t.Context(), connectionString)
	defer dbpool.Close() //nolint:staticcheck // we require no error below
	require.NoError(t, err)
	tx, err := dbpool.Begin(t.Context())
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Printf("failed to rollback transaction: %s", err)
		}
	}(tx, t.Context())
	require.NoError(t, err)

	underTest := repository.NewRepository(tx)
	_, err = underTest.CreateAppointment(t.Context(),
		&domain.Appointment{
			FirstName: "first",
			LastName:  "last",
			VisitDate: ptr.To(time.Date(2024, 12, 25, 9, 0, 0, 0, time.UTC)),
		})
	require.NoError(t, err)

	_, dupeErr := underTest.CreateAppointment(t.Context(),
		&domain.Appointment{
			FirstName: "first",
			LastName:  "last",
			VisitDate: ptr.To(time.Date(2024, 12, 25, 9, 0, 0, 0, time.UTC)),
		})
	require.Error(t, dupeErr)
	require.ErrorIs(t, dupeErr, domain.ErrAppointmentDateTaken)
}
