package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	creditapp "github.com/example/banking-service/internal/application/credits"
	"github.com/example/banking-service/internal/infrastructure/cbr"
	"github.com/example/banking-service/internal/infrastructure/config"
	appdb "github.com/example/banking-service/internal/infrastructure/db"
	"github.com/example/banking-service/internal/infrastructure/logger"
	pgrepo "github.com/example/banking-service/internal/infrastructure/repository/postgres"

	"github.com/sirupsen/logrus"
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

	postgres, err := appdb.Connect(ctx, cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to postgres")
	}
	defer postgres.Close()

	txManager := appdb.NewTxManager(postgres.DB)

	accountRepository := pgrepo.NewAccountRepository(txManager)
	creditRepository := pgrepo.NewCreditRepository(txManager)
	scheduleRepository := pgrepo.NewPaymentScheduleRepository(txManager)
	transactionRepository := pgrepo.NewTransactionRepository(txManager)
	cbrClient := cbr.NewClient(cfg.CBRSoapURL)

	creditService := creditapp.NewService(txManager, accountRepository, creditRepository, scheduleRepository, transactionRepository, cbrClient)

	interval := time.Duration(cfg.SchedulerIntervalHours) * time.Hour
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.WithField("interval", interval.String()).Info("credit payment scheduler started")

	runCreditPaymentJob(ctx, log, creditService)

	for {
		select {
		case <-ctx.Done():
			log.Info("scheduler stopped")
			return
		case <-ticker.C:
			runCreditPaymentJob(ctx, log, creditService)
		}
	}
}

func runCreditPaymentJob(ctx context.Context, log *logrus.Logger, service *creditapp.Service) {
	result, err := service.ProcessDuePayments(ctx, time.Now().UTC(), 100)
	if err != nil {
		log.WithError(err).Error("credit payment scheduler job failed")
		return
	}

	log.WithFields(logrus.Fields{
		"processed": result.Processed,
		"paid":      result.Paid,
		"penalized": result.Penalized,
		"failed":    result.Failed,
	}).Info("credit payment scheduler job completed")
}
