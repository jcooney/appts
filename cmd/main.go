package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jcooney/appts/api"
	"github.com/jcooney/appts/domain"
	"github.com/jcooney/appts/publichols"
)

func main() {
	// dependencies initialised here
	getter, err := publichols.NewPublicHolidayGetter("https://date.nager.at")
	if err != nil {
		log.Fatalf("publichols.NewPublicHolidayGetter: %v", err)
	}
	service := domain.NewAppointmentCreatorService(nil, getter)

	server := &http.Server{Addr: "0.0.0.0:3333", Handler: api.ChiHandler(service)}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
