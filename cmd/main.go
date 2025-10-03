package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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

	publicHolidayGetter, err := publichols.NewPublicHolidayGetter("https://date.nager.at")
	if err != nil {
		log.Fatalf("error initialising public holiday checker client: %v", err)
	}
	pool, err := pgxpool.New(ctx, "postgres://appt_user:appt_password@localhost:5432/tabeo?sslmode=disable")
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	repo := repository.NewRepository(pool)
	service := domain.NewAppointmentCreatorService(repo, publicHolidayGetter)
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
