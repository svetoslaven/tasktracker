package main

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.pg.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.pg.maxOpenConns)
	db.SetMaxIdleConns(cfg.pg.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.pg.connMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
