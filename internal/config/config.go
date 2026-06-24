package config

import (
	"appointments/internal/env"
	"flag"
)

type Config struct {
	Port         int
	Env          string
	VryTokenTTL  string
	AuthTokenTTL string
	DB
	Cache
	Resend
}

type DB struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type Cache struct {
	URL string
}

type Resend struct {
	APIKey string
	Sender string
}

func Load() Config {
	var cfg Config
	flag.IntVar(&cfg.Port, "port", env.GetInt("PORT", 4000), "server port")
	flag.StringVar(&cfg.Env, "env", env.GetString("ENV", "dev"), "server environment: dev|stage|prod")
	flag.StringVar(&cfg.DB.DSN, "dsn", env.GetString("DB_DSN", "postgres://dzenthai:pa55word@localhost:5432/appointments"),
		"postgres data source name")
	flag.IntVar(&cfg.DB.MaxOpenConns, "max-open-conns", env.GetInt("DB_MAX_OPEN_CONNS", 25), "postgres max open connections")
	flag.IntVar(&cfg.DB.MaxIdleConns, "max-idle-conns", env.GetInt("DB_MAX_IDLE_CONNS", 25), "postgres max idle connections")
	flag.StringVar(&cfg.DB.MaxIdleTime, "max-idle-time", env.GetString("DB_MAX_IDLE_TIME", "15m"), "postgres max idle time")
	flag.StringVar(&cfg.VryTokenTTL, "vry-token-ttl", env.GetString("VERIF_TOKEN_TTL", "15m"), "time to live of verification token")
	flag.StringVar(&cfg.AuthTokenTTL, "auth-token-ttl", env.GetString("AUTH_TOKEN_TTL", "24h"), "time to live of authentication token")
	flag.StringVar(&cfg.Resend.APIKey, "resend-api-key", env.GetString("RESEND_API_KEY", "-"), "resend api key")
	flag.StringVar(&cfg.Resend.Sender, "resend-sender", env.GetString("RESEND_SENDER", "-"), "resend sender")
	flag.Parse()

	return cfg
}
