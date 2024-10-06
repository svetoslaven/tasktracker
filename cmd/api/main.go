package main

import (
	"os"

	"github.com/svetoslaven/tasktracker/internal/jsonlog"
)

const version = "1.0.0"

func main() {
	cfg := loadConfig()

	logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	app := &application{
		cfg:    cfg,
		logger: logger,
	}

	if err := app.run(); err != nil {
		logger.LogFatal(err, nil)
	}
}
