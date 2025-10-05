package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jcooney/appts/api"
	"github.com/jcooney/appts/domain"
	"github.com/jcooney/appts/publichols"
	"github.com/jcooney/appts/repository"
)

func main() {
	slog.Info("Starting appointment service")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	migrationDBURL, ok := os.LookupEnv("MIGRATE_DB_URL")
	if !ok {
		log.Fatal("DB_URL environment variable not set")
	}
	m, err := migrate.New("file://schema/ddl", migrationDBURL)
	if err != nil {
		log.Fatalf("error initialising migrations: %v", err)
	}
	err = m.Up()
	if err != nil {
		log.Fatalf("error running migrations: %v", err)
	}

	publicHolidayGetter, err := publichols.NewPublicHolidayGetter("https://date.nager.at")
	if err != nil {
		log.Fatalf("error initialising public holiday checker client: %v", err)
	}

	dbURL, ok := os.LookupEnv("DB_URL")
	if !ok {
		log.Fatal("DB_URL environment variable not set")
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	repo := repository.NewRepository(pool)
	service := domain.NewAppointmentCreatorService(repo, publicHolidayGetter, time.Now)
	server := &http.Server{Addr: "0.0.0.0:3333", Handler: api.ChiHandler(service)}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	slog.Info("Shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
	slog.Info("Server gracefully stopped")
}
