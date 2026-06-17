package main

import (
	"appointments/internal/config"
	"appointments/internal/postgres"
	"appointments/internal/server"
	"appointments/internal/store"
	"appointments/internal/user"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
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

	s := store.New(db)
	userHandler := user.NewHandler(s.User, logger)

	srv := server.New(cfg, logger, db, userHandler)

	return srv.Serve()
}
