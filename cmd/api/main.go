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
	"sync"
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

	wg := &sync.WaitGroup{}

	db, err := postgres.Open(cfg.DB)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	m := mailer.New(cfg.Resend.APIKey, cfg.Resend.Sender, logger)

	s := store.New(db)
	vryDur, err := time.ParseDuration(cfg.VryTokenTTL)
	if err != nil {
		return err
	}
	authDur, err := time.ParseDuration(cfg.AuthTokenTTL)
	if err != nil {
		return err
	}
	userHandler := user.NewHandler(s.User, s.Token, logger, wg, m, vryDur, authDur)

	srv := server.New(cfg, logger, wg, userHandler, s.User)

	return srv.Serve()
}
