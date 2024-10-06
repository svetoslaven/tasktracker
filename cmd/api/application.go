package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type application struct {
	cfg config
}

func (app *application) run() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.cfg.port),
		Handler:      app.registerRoutes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownErrorCh := make(chan error)

	go func() {
		quitSignalCh := make(chan os.Signal, 1)
		signal.Notify(quitSignalCh, syscall.SIGINT, syscall.SIGTERM)

		quitSignal := <-quitSignalCh

		log.Printf("caught %s signal", quitSignal.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			shutdownErrorCh <- err
			return
		}

		close(shutdownErrorCh)
	}()

	log.Printf("starting %s server on %s", app.cfg.environment, srv.Addr)

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownErrorCh; err != nil {
		return err
	}

	log.Printf("stopped server on %s", srv.Addr)

	return nil
}
