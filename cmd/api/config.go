package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type config struct {
	port        int
	environment string

	pg struct {
		dsn             string
		maxOpenConns    int
		maxIdleConns    int
		connMaxIdleTime time.Duration
	}
}

func loadConfig() config {
	const (
		environmentDevelopment = "development"
		environmentStaging     = "staging"
		environmentProduction  = "production"
	)

	var cfg config

	flag.IntVar(&cfg.port, "port", parseIntEnv("PORT", 8080), "Set API server port")
	flag.StringVar(
		&cfg.environment,
		"environment",
		parseStringEnv("ENVIRONMENT", environmentDevelopment),
		fmt.Sprintf("Set environment (%s|%s|%s)", environmentDevelopment, environmentStaging, environmentProduction),
	)

	flag.StringVar(&cfg.pg.dsn, "pg-dsn", os.Getenv("PG_DSN"), "Set PostgreSQL DSN")
	flag.IntVar(
		&cfg.pg.maxOpenConns,
		"pg-max-open-conns",
		parseIntEnv("PG_MAX_OPEN_CONNS", 25),
		"Set PostgreSQL max open connections",
	)
	flag.IntVar(
		&cfg.pg.maxIdleConns,
		"pg-max-idle-conns",
		parseIntEnv("PG_MAX_IDLE_CONNS", 25),
		"Set PostgreSQL max idle connections",
	)
	flag.DurationVar(
		&cfg.pg.connMaxIdleTime,
		"pg-conn-max-idle-time",
		parseDurationEnv("PG_CONN_MAX_IDLE_TIME", 15*time.Minute),
		"Set PostgreSQL connections max idle time",
	)

	flag.Parse()

	cfg.environment = strings.ToLower(cfg.environment)

	if !slices.Contains([]string{environmentDevelopment, environmentStaging, environmentProduction}, cfg.environment) {
		fmt.Printf(
			"Invalid environment: %s, Must be one of: %s, %s, %s.\n",
			cfg.environment,
			environmentDevelopment, environmentStaging, environmentDevelopment,
		)
		os.Exit(1)
	}

	return cfg
}

func parseIntEnv(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return intValue
}

func parseStringEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}
