package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/banking-service/internal/infrastructure/config"
	"github.com/example/banking-service/internal/infrastructure/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.NewLogrus(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.WithFields(cfg.SafeForLog()).Info("scheduler bootstrap started")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	interval := time.Duration(cfg.SchedulerIntervalHours) * time.Hour
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.WithField("interval", interval.String()).Info("scheduler started")

	runStubJob(log)

	for {
		select {
		case <-ctx.Done():
			log.Info("scheduler stopped")
			return
		case <-ticker.C:
			runStubJob(log)
		}
	}
}

type loggerWithInfo interface {
	Info(args ...any)
}

func runStubJob(log loggerWithInfo) {
	log.Info("scheduler stub tick: no jobs implemented yet")
}
