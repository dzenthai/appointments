package main

import (
	"appointments/internal/config"
	"appointments/internal/server"
	"appointments/internal/storage"
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := config.Load()

	db, err := storage.OpenDB(cfg.DBCfg)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	srv := server.Server{
		Cfg:    cfg,
		Logger: logger,
		DB:     db,
	}

	err = srv.Serve()
	if err != nil {
		return err
	}

	return nil
}
