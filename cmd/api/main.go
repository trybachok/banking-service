package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	authapp "github.com/example/banking-service/internal/application/auth"
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
	passwordHasher := appcrypto.NewPasswordHasher()
	jwtManager := appjwt.NewManager(cfg.JWTSecret, cfg.JWTTTLHours)

	authService := authapp.NewService(userRepository, passwordHasher, jwtManager)
	authHandler := handlers.NewAuthHandler(authService, log)

	httpHandler := router.New(authHandler, jwtManager, log)

	addr := ":" + cfg.AppPort

	log.WithField("addr", addr).Info("api server listening")

	if err := http.ListenAndServe(addr, httpHandler); err != nil {
		log.WithError(err).Fatal("api server stopped")
	}
}
