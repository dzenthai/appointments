package config

import (
	"appointments/internal/env"
	"flag"
)

type Config struct {
	Port int
	Env  string
	DB   DB
}

type DB struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func Load() Config {
	var cfg Config
	flag.IntVar(&cfg.Port, "port", env.GetInt("PORT", 4000), "server port")
	flag.StringVar(&cfg.Env, "env", env.GetString("ENV", "dev"), "server environment: dev|stage|prod")
	flag.StringVar(&cfg.DB.DSN, "dsn", env.GetString("DSN", "postgres://dzenthai:pa55word@localhost:5432/appointments"), "data source name")
	flag.IntVar(&cfg.DB.MaxOpenConns, "max-open-conns", env.GetInt("MAX_OPEN_CONNS", 25), "postgres max open connections")
	flag.IntVar(&cfg.DB.MaxIdleConns, "max-idle-conns", env.GetInt("MAX_IDLE_CONNS", 25), "postgres max idle connections")
	flag.StringVar(&cfg.DB.MaxIdleTime, "max-idle-time", env.GetString("MAX_IDLE_TIME", "15m"), "postgres max idle time")
	flag.Parse()

	return cfg
}
