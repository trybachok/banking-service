package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/banking-service/internal/infrastructure/config"

	_ "github.com/lib/pq"
)

type Postgres struct {
	DB *sql.DB
}

func Connect(ctx context.Context, cfg *config.Config) (*Postgres, error) {
	database, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	database.SetMaxOpenConns(20)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(30 * time.Minute)
	database.SetConnMaxIdleTime(10 * time.Minute)

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Postgres{DB: database}, nil
}

func (p *Postgres) Close() error {
	if p == nil || p.DB == nil {
		return nil
	}

	return p.DB.Close()
}
