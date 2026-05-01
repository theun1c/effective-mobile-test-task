package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/theun1c/effective-mobile-test-task/internal/config"
	"github.com/theun1c/effective-mobile-test-task/internal/http/router"
	postgresrepo "github.com/theun1c/effective-mobile-test-task/internal/repository/postgres"
	subscriptionservice "github.com/theun1c/effective-mobile-test-task/internal/service/subscription"
	subscriptiontotalservice "github.com/theun1c/effective-mobile-test-task/internal/service/subscription_total"
)

type App struct {
	cfg    config.Config
	db     *sql.DB
	logger *slog.Logger
	server *http.Server
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	logger.Info(
		"initializing application",
		"app_env", cfg.AppEnv,
		"log_level", cfg.LogLevel,
	)

	db, err := newPostgresDB(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	subscriptionRepository := postgresrepo.NewSubscriptionRepository(db)
	subscriptionTotalRepository := postgresrepo.NewSubscriptionTotalRepository(db)
	subscriptionService := subscriptionservice.NewWithLogger(subscriptionRepository, logger)
	subscriptionTotalService := subscriptiontotalservice.New(subscriptionTotalRepository)
	httpRouter := router.NewWithTotal(logger, subscriptionService, subscriptionTotalService)

	server := &http.Server{
		Addr:              cfg.HTTP.Address(),
		Handler:           httpRouter,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		cfg:    cfg,
		db:     db,
		logger: logger,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	shutdownDone := make(chan struct{})

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
		defer cancel()

		if err := a.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("graceful shutdown failed", "error", err)
		}

		close(shutdownDone)
	}()

	a.logger.Info("postgres connection is ready")
	a.logger.Info("http server is listening", "address", a.server.Addr)

	err := a.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	if ctx.Err() != nil {
		<-shutdownDone
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	var result error

	if err := a.server.Shutdown(ctx); err != nil {
		result = errors.Join(result, fmt.Errorf("shutdown http server: %w", err))
	}

	if err := a.db.Close(); err != nil {
		result = errors.Join(result, fmt.Errorf("close postgres connection: %w", err))
	}

	return result
}

func newPostgresDB(cfg config.Config, logger *slog.Logger) (*sql.DB, error) {
	logger.Info(
		"connecting to postgres",
		"host", cfg.Postgres.Host,
		"port", cfg.Postgres.Port,
		"database", cfg.Postgres.Name,
		"user", cfg.Postgres.User,
	)

	db, err := sql.Open("pgx", cfg.Postgres.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Postgres.PingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	logger.Info("postgres connection established")

	return db, nil
}
