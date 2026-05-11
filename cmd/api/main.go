package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	accountapp "github.com/example/banking-service/internal/application/accounts"
	adminapp "github.com/example/banking-service/internal/application/admin"
	analyticsapp "github.com/example/banking-service/internal/application/analytics"
	authapp "github.com/example/banking-service/internal/application/auth"
	cardapp "github.com/example/banking-service/internal/application/cards"
	creditapp "github.com/example/banking-service/internal/application/credits"
	mfaapp "github.com/example/banking-service/internal/application/mfa"
	transferapp "github.com/example/banking-service/internal/application/transfers"
	"github.com/example/banking-service/internal/infrastructure/cbr"
	"github.com/example/banking-service/internal/infrastructure/config"
	appcrypto "github.com/example/banking-service/internal/infrastructure/crypto"
	appdb "github.com/example/banking-service/internal/infrastructure/db"
	appjwt "github.com/example/banking-service/internal/infrastructure/jwt"
	"github.com/example/banking-service/internal/infrastructure/logger"
	pgrepo "github.com/example/banking-service/internal/infrastructure/repository/postgres"
	"github.com/example/banking-service/internal/transport/http/handlers"
	"github.com/example/banking-service/internal/transport/http/router"
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

	log.WithFields(cfg.SafeForLog()).Info("api bootstrap started")

	ctx := context.Background()

	postgres, err := appdb.Connect(ctx, cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to postgres")
	}
	defer postgres.Close()

	txManager := appdb.NewTxManager(postgres.DB)

	userRepository := pgrepo.NewUserRepository(txManager)
	accountRepository := pgrepo.NewAccountRepository(txManager)
	transactionRepository := pgrepo.NewTransactionRepository(txManager)
	cardRepository := pgrepo.NewCardRepository(txManager)
	creditRepository := pgrepo.NewCreditRepository(txManager)
	scheduleRepository := pgrepo.NewPaymentScheduleRepository(txManager)
	mfaRepository := pgrepo.NewMFARepository(txManager)
	emailOutboxRepository := pgrepo.NewEmailOutboxRepository(txManager)

	passwordHasher := appcrypto.NewPasswordHasher()
	jwtManager := appjwt.NewManager(cfg.JWTSecret, cfg.JWTTTLHours)
	cardProtector := appcrypto.NewCardCrypto(postgres.DB, cfg.PGPSymKey, cfg.CardHMACKey)
	cbrClient := cbr.NewClient(cfg.CBRSoapURL)

	authService := authapp.NewService(userRepository, passwordHasher, jwtManager)
	accountService := accountapp.NewService(txManager, accountRepository, transactionRepository)
	transferService := transferapp.NewService(txManager, accountRepository, transactionRepository)
	cardService := cardapp.NewService(txManager, accountRepository, cardRepository, transactionRepository, cardProtector)
	creditService := creditapp.NewService(txManager, accountRepository, creditRepository, scheduleRepository, transactionRepository, cbrClient)
	analyticsService := analyticsapp.NewService(accountRepository, transactionRepository, creditRepository, scheduleRepository)
	mfaService := mfaapp.NewService(userRepository, mfaRepository, emailOutboxRepository, passwordHasher)
	adminService := adminapp.NewService(userRepository, accountRepository)

	authHandler := handlers.NewAuthHandler(authService, log)
	accountHandler := handlers.NewAccountHandler(accountService, log)
	transferHandler := handlers.NewTransferHandler(transferService, log)
	cardHandler := handlers.NewCardHandler(cardService, log)
	creditHandler := handlers.NewCreditHandler(creditService, log)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, log)
	mfaHandler := handlers.NewMFAHandler(mfaService, log)
	adminHandler := handlers.NewAdminHandler(adminService, log)

	httpHandler := router.New(
		authHandler,
		accountHandler,
		transferHandler,
		cardHandler,
		creditHandler,
		analyticsHandler,
		mfaHandler,
		adminHandler,
		mfaService,
		jwtManager,
		log,
	)

	addr := ":" + cfg.AppPort

	log.WithField("addr", addr).Info("api server listening")

	if err := http.ListenAndServe(addr, httpHandler); err != nil {
		log.WithError(err).Fatal("api server stopped")
	}
}
