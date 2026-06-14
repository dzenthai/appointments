package main

import (
	"flag"
	"log/slog"
	"os"
)

type application struct {
	cfg    config
	logger *slog.Logger
}

type config struct {
	port int
	env  string
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", getIntEnv("PORT", 4000), "server port")
	flag.StringVar(&cfg.env, "env", getStringEnv("ENV", "dev"), "server environment: dev|stage|prod")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := application{
		cfg:    cfg,
		logger: log,
	}

	err := app.serve()
	if err != nil {
		app.logger.Error("server stopped unexpectedly", "err", err)
		os.Exit(1)
	}
}
