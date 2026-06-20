package main

import (
	"appointments/internal/config"
	"appointments/internal/mailer"
	"appointments/internal/postgres"
	"appointments/internal/server"
	"appointments/internal/store"
	"appointments/internal/user"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {

	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := postgres.Open(cfg.DB)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	m := mailer.New(cfg.Resend.APIKey, cfg.Resend.Sender)

	s := store.New(db)
	dur, err := time.ParseDuration(cfg.VryTokenTTL)
	if err != nil {
		return err
	}
	userHandler := user.NewHandler(s.User, s.Verification, logger, m, dur)

	srv := server.New(cfg, logger, userHandler)

	return srv.Serve()
}
