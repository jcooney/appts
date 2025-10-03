package schema_test

import (
	"log"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestAppliesMigration(t *testing.T) {
	ctx := t.Context()

	dbName := "tabeo"
	dbUser := "postgres"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	})
	require.NoError(t, err)

	connectionString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	connectionString = connectionString + " dbname=" + dbName + " user=" + dbUser + " password=" + dbPassword

	m, err := migrate.New("file://ddl", connectionString+"&sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, m.Up())

	v, _, _ := m.Version()
	require.Equal(t, v, uint(1))
}
