package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/theun1c/effective-mobile-test-task/internal/app"
	"github.com/theun1c/effective-mobile-test-task/internal/config"
	applogger "github.com/theun1c/effective-mobile-test-task/internal/logger"
)

func main() {
	bootstrapLogger := applogger.New("info", os.Stdout)

	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := applogger.New(cfg.LogLevel, os.Stdout)
	logger.Info("configuration loaded", "app_env", cfg.AppEnv, "log_level", cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("failed to bootstrap application", "error", err)
		os.Exit(1)
	}

	if err := application.Run(ctx); err != nil {
		logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("application stopped")
}
