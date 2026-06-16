package config

import (
	"appointments/internal/env"
	"appointments/internal/storage"
	"flag"
)

type Config struct {
	Port  int
	Env   string
	DBCfg storage.DBCfg
}

func Load() (cfg Config) {
	flag.IntVar(&cfg.Port, "port", env.GetInt("PORT", 4000), "server port")
	flag.StringVar(&cfg.Env, "env", env.GetString("ENV", "dev"), "server environment: dev|stage|prod")
	flag.StringVar(&cfg.DBCfg.DSN, "dsn", env.GetString("DSN", "postgres://dzenthai:pa55word@localhost:5432/appointments"), "data source name")
	flag.IntVar(&cfg.DBCfg.MaxOpenConns, "max-open-conns", env.GetInt("MAX_OPEN_CONNS", 25), "database max open connections")
	flag.IntVar(&cfg.DBCfg.MaxIdleConns, "max-idle-conns", env.GetInt("MAX_IDLE_CONNS", 25), "database max idle connections")
	flag.StringVar(&cfg.DBCfg.MaxIdleTime, "max-idle-time", env.GetString("MAX_IDLE_TIME", "15m"), "database max idle time")
	flag.Parse()
	
	return cfg
}
