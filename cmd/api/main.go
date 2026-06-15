package main

import (
	"appointments/internal/env"
	"appointments/internal/storage"
	"database/sql"
	"flag"
	"log/slog"
	"os"
)

type application struct {
	cfg    config
	logger *slog.Logger
	db     *sql.DB
}

type config struct {
	port  int
	env   string
	dbCfg storage.DBCfg
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", env.GetInt("PORT", 4000), "server port")
	flag.StringVar(&cfg.env, "env", env.GetString("ENV", "dev"), "server environment: dev|stage|prod")
	flag.StringVar(&cfg.dbCfg.DSN, "dsn", env.GetString("DSN", "postgres://dzenthai:pa55word@localhost:5432/appointments"), "data source name")
	flag.IntVar(&cfg.dbCfg.MaxOpenConns, "max-open-conns", env.GetInt("MAX_OPEN_CONNS", 25), "database max open connections")
	flag.IntVar(&cfg.dbCfg.MaxIdleConns, "max-idle-conns", env.GetInt("MAX_IDLE_CONNS", 25), "database max idle connections")
	flag.StringVar(&cfg.dbCfg.MaxIdleTime, "max-idle-time", env.GetString("MAX_IDLE_TIME", "15m"), "database max idle time")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := storage.OpenDB(cfg.dbCfg)
	if err != nil {
		logger.Error("unable to establish a connection with the database", "err", err)
		return
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	app := application{
		cfg:    cfg,
		logger: logger,
		db:     db,
	}

	logger.Info("database connection pool established")

	err = app.serve()
	if err != nil {
		app.logger.Error("server stopped unexpectedly", "err", err)
		os.Exit(1)
	}
}
