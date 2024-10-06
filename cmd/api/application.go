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

	"github.com/svetoslaven/tasktracker/internal/jsonlog"
)

type application struct {
	cfg    config
	logger *jsonlog.Logger
}

func (app *application) run() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.cfg.port),
		Handler:      app.registerRoutes(),
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownErrorCh := make(chan error)

	go func() {
		quitSignalCh := make(chan os.Signal, 1)
		signal.Notify(quitSignalCh, syscall.SIGINT, syscall.SIGTERM)

		quitSignal := <-quitSignalCh

		app.logger.LogInfo("caught signal", map[string]string{
			"signal": quitSignal.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			shutdownErrorCh <- err
			return
		}

		close(shutdownErrorCh)
	}()

	app.logger.LogInfo("starting server", map[string]string{
		"addr":        srv.Addr,
		"environment": app.cfg.environment,
	})

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownErrorCh; err != nil {
		return err
	}

	app.logger.LogInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
