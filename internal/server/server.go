package server

import (
	"appointments/internal/config"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	cfg    config.Config
	logger *slog.Logger
	db     *sql.DB
}

func New(cfg config.Config, logger *slog.Logger, db *sql.DB) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
		db:     db,
	}
}

func (s *Server) Serve() error {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.Port),
		Handler:      s.route(),
		ErrorLog:     slog.NewLogLogger(s.logger.Handler(), slog.LevelError),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	s.logger.Info("starting server", "port", s.cfg.Port, "env", s.cfg.Env)

	return srv.ListenAndServe()
}
