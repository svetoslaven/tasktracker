package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type config struct {
	port        int
	environment string
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
