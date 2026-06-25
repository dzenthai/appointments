package server

import (
	"appointments/internal/appointment"
	"appointments/internal/config"
	"appointments/internal/user"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	cfg        config.Config
	logger     *slog.Logger
	wg         *sync.WaitGroup
	users      *user.Handler
	userStore  *user.Store
	appHandler *appointment.Handler
}

func New(
	cfg config.Config,
	logger *slog.Logger,
	wg *sync.WaitGroup,
	users *user.Handler,
	userStore *user.Store,
	appHandler *appointment.Handler,
) *Server {
	return &Server{
		cfg:        cfg,
		logger:     logger,
		wg:         wg,
		users:      users,
		userStore:  userStore,
		appHandler: appHandler,
	}
}

func (s *Server) Serve() error {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.Port),
		Handler:      s.routes(),
		ErrorLog:     slog.NewLogLogger(s.logger.Handler(), slog.LevelError),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	s.logger.Info("starting server", "port", s.cfg.Port, "env", s.cfg.Env)

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		s.logger.Info("shutting down the server", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
			return
		}

		s.logger.Info("completing background tasks", "addr", srv.Addr)

		s.wg.Wait()

		shutdownError <- nil
	}()

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return <-shutdownError
}
