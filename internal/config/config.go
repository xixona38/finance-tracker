package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL        string
	HTTPAddr           string
	SessionIdleTTL     time.Duration
	SessionAbsoluteTTL time.Duration
	CookieSecure       bool
}

func Load() (Config, error) {
	config := Config{}
	pgURL, ok := os.LookupEnv("PG_URL")
	if !ok || pgURL == "" {
		return Config{}, fmt.Errorf("PG_URL is required")
	}
	config.DatabaseURL = pgURL

	httpAddr, ok := os.LookupEnv("HTTP_ADDR")
	if !ok || httpAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	config.HTTPAddr = httpAddr

	sessionIdleTtlStr, ok := os.LookupEnv("SESSION_IDLE_TTL")
	if !ok || sessionIdleTtlStr == "" {
		return Config{}, fmt.Errorf("SESSION_IDLE_TTL is required")
	}
	sessionIdleTtl, err := time.ParseDuration(sessionIdleTtlStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid SESSION_IDLE_TTL: %w", err)
	}
	config.SessionIdleTTL = sessionIdleTtl

	sessionAbsTtlStr, ok := os.LookupEnv("SESSION_ABSOLUTE_TTL")
	if !ok || sessionAbsTtlStr == "" {
		return Config{}, fmt.Errorf("SESSION_ABSOLUTE_TTL is required")
	}
	sessionAbsTtl, err := time.ParseDuration(sessionAbsTtlStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid SESSION_ABSOLUTE_TTL: %w", err)
	}
	config.SessionAbsoluteTTL = sessionAbsTtl

	cookieSecureStr, ok := os.LookupEnv("COOKIE_SECURE")
	if !ok || cookieSecureStr == "" {
		return Config{}, fmt.Errorf("COOKIE_SECURE is required")
	}
	cookieSecure, err := strconv.ParseBool(cookieSecureStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid COOKIE_SECURE: %w", err)
	}
	config.CookieSecure = cookieSecure

	return config, nil

}
