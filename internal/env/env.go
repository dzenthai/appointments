package env

import (
	"os"
	"strconv"
)

func GetString(key string, fallback string) string {
	env := os.Getenv(key)
	if env == "" {
		return fallback
	}
	return env
}

func GetInt(key string, fallback int) int {
	env := os.Getenv(key)
	if env == "" {
		return fallback
	}
	i, err := strconv.Atoi(env)
	if err != nil {
		return fallback
	}
	return i
}

func GetBool(key string, fallback bool) bool {
	env := os.Getenv(key)
	if env == "" {
		return fallback
	}
	b, err := strconv.ParseBool(env)
	if err != nil {
		return fallback
	}
	return b
}
