package main

import (
	"os"

	"github.com/svetoslaven/tasktracker/internal/jsonlog"
	"github.com/svetoslaven/tasktracker/internal/mailer"
	"github.com/svetoslaven/tasktracker/internal/repositories/postgres"
	"github.com/svetoslaven/tasktracker/internal/services/domain"
)

const version = "1.0.0"

func main() {
	cfg := loadConfig()

	logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.LogFatal(err, nil)
	}

	defer db.Close()

	logger.LogInfo("database connection pool established", nil)

	app := &application{
		cfg:      cfg,
		logger:   logger,
		services: domain.NewServiceRegistry(postgres.NewRepositoryRegistry(db)),
		mailer:   mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	if err := app.run(); err != nil {
		logger.LogFatal(err, nil)
	}
}
