package postgres

import (
	"appointments/internal/config"
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(dbCfg config.DB) (*sql.DB, error) {
	db, err := sql.Open("pgx", dbCfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(dbCfg.MaxOpenConns)
	db.SetMaxIdleConns(dbCfg.MaxIdleConns)

	dur, err := time.ParseDuration(dbCfg.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(dur)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
