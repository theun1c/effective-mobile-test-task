package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/theun1c/effective-mobile-test-task/internal/app"
	"github.com/theun1c/effective-mobile-test-task/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("bootstrap application: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("run application: %v", err)
	}
}
