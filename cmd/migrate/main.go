package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/sirupsen/logrus"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

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

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	m, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize migrator")
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.WithError(err).Fatal("failed to read migration version")
	}

	log.WithFields(logrus.Fields{
		"command":        command,
		"before_version": version,
		"before_dirty":   dirty,
	}).Info("migration started")

	switch command {
	case "up":
		err = m.Up()

	case "down":
		err = m.Steps(-1)

	case "version":
		printVersionAndExit(m, log)

	default:
		log.WithField("command", command).Fatal("unsupported migration command")
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.WithField("command", command).Info("migration no change")
		return
	}

	if err != nil {
		version, dirty, versionErr := m.Version()

		fields := logrus.Fields{
			"command": command,
		}

		if versionErr == nil {
			fields["failed_version"] = version
			fields["dirty"] = dirty
		}

		log.WithFields(fields).WithError(err).Fatal("migration failed")
	}

	version, dirty, err = m.Version()
	if err != nil {
		log.WithError(err).Fatal("failed to read migration version after migration")
	}

	log.WithFields(logrus.Fields{
		"command":       command,
		"after_version": version,
		"after_dirty":   dirty,
	}).Info("migration finished")
}

func printVersionAndExit(m *migrate.Migrate, log *logrus.Logger) {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		log.WithFields(logrus.Fields{
			"version": nil,
			"dirty":   false,
		}).Info("migration version")
		return
	}

	if err != nil {
		log.WithError(err).Fatal("failed to read migration version")
	}

	log.WithFields(logrus.Fields{
		"version": version,
		"dirty":   dirty,
	}).Info("migration version")
}
