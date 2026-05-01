package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string
	LogLevel string
	HTTP     HTTPConfig
	Postgres PostgresConfig
}

type HTTPConfig struct {
	Host            string
	Port            string
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	Host        string
	Port        string
	Name        string
	User        string
	Password    string
	SSLMode     string
	PingTimeout time.Duration
}

func (c HTTPConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c PostgresConfig) DSN() string {
	query := url.Values{}
	query.Set("sslmode", c.SSLMode)

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.Host, c.Port),
		Path:     c.Name,
		RawQuery: query.Encode(),
	}).String()
}

func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTP: HTTPConfig{
			Host: getEnv("HTTP_HOST", "0.0.0.0"),
			Port: getEnv("HTTP_PORT", "8080"),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "subscriptions"),
			User:     getEnv("DB_USER", "subscriptions"),
			Password: getEnv("DB_PASSWORD", "subscriptions"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	var err error

	cfg.HTTP.ShutdownTimeout, err = getDurationEnv("HTTP_SHUTDOWN_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("parse HTTP_SHUTDOWN_TIMEOUT: %w", err)
	}

	cfg.Postgres.PingTimeout, err = getDurationEnv("DB_PING_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("parse DB_PING_TIMEOUT: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadDotEnv(path string) error {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("stat %s: %w", path, err)
	}

	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}

	return nil
}

func (c Config) validate() error {
	var result error

	if c.HTTP.Host == "" {
		result = errors.Join(result, errors.New("HTTP_HOST must not be empty"))
	}

	if c.HTTP.Port == "" {
		result = errors.Join(result, errors.New("HTTP_PORT must not be empty"))
	}

	if _, err := strconv.Atoi(c.HTTP.Port); err != nil {
		result = errors.Join(result, fmt.Errorf("HTTP_PORT must be numeric: %w", err))
	}

	if c.Postgres.Host == "" {
		result = errors.Join(result, errors.New("DB_HOST must not be empty"))
	}

	if c.Postgres.Port == "" {
		result = errors.Join(result, errors.New("DB_PORT must not be empty"))
	}

	if _, err := strconv.Atoi(c.Postgres.Port); err != nil {
		result = errors.Join(result, fmt.Errorf("DB_PORT must be numeric: %w", err))
	}

	if c.Postgres.Name == "" {
		result = errors.Join(result, errors.New("DB_NAME must not be empty"))
	}

	if c.Postgres.User == "" {
		result = errors.Join(result, errors.New("DB_USER must not be empty"))
	}

	if c.Postgres.SSLMode == "" {
		result = errors.Join(result, errors.New("DB_SSLMODE must not be empty"))
	}

	if c.HTTP.ShutdownTimeout <= 0 {
		result = errors.Join(result, errors.New("HTTP_SHUTDOWN_TIMEOUT must be greater than zero"))
	}

	if c.Postgres.PingTimeout <= 0 {
		result = errors.Join(result, errors.New("DB_PING_TIMEOUT must be greater than zero"))
	}

	return result
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}

	return duration, nil
}
