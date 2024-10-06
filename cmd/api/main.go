package main

import "log"

const version = "1.0.0"

func main() {
	cfg := loadConfig()

	app := &application{cfg: cfg}

	if err := app.run(); err != nil {
		log.Fatal(err)
	}
}
