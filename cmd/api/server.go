package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func (app *application) serve() error {
	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", app.cfg.port),
		Handler:           app.handle(),
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		ErrorLog:          slog.NewLogLogger(slog.NewTextHandler(os.Stdout, nil), slog.LevelError),
	}

	app.logger.Info("starting server", "port", app.cfg.port, "env", app.cfg.env)

	return srv.ListenAndServe()
}
