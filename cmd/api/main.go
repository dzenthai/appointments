package main

import (
	"appointments/internal/config"
	"appointments/internal/database"
	"appointments/internal/server"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {

	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := database.Open(cfg.DB)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	srv := server.New(cfg, logger, db)

	return srv.Serve()
}
