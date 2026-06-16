package server

import (
	"appointments/internal/config"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

type Server struct {
	Cfg    config.Config
	Logger *slog.Logger
	DB     *sql.DB
}

func (s *Server) Serve() error {
	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", s.Cfg.Port),
		Handler:           s.route(),
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		ErrorLog:          slog.NewLogLogger(slog.NewTextHandler(os.Stdout, nil), slog.LevelError),
	}

	s.Logger.Info("starting server", "port", s.Cfg.Port, "env", s.Cfg.Env)

	return srv.ListenAndServe()
}
