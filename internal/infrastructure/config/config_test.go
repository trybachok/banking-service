package config

import (
	"os"
	"testing"
)

func TestLoadConfigUsesDefaults(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/banking_service?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	t.Setenv("PGP_SYM_KEY", "test-pgp-key")
	t.Setenv("CARD_HMAC_KEY", "test-hmac-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	if cfg.AppPort != "8085" {
		t.Fatalf("expected default APP_PORT=8085, got %q", cfg.AppPort)
	}

	if cfg.LogLevel != "info" {
		t.Fatalf("expected default LOG_LEVEL=info, got %q", cfg.LogLevel)
	}

	if cfg.JWTTTLHours != 24 {
		t.Fatalf("expected default JWT_TTL_HOURS=24, got %d", cfg.JWTTTLHours)
	}

	if cfg.SchedulerIntervalHours != 12 {
		t.Fatalf("expected default SCHEDULER_INTERVAL_HOURS=12, got %d", cfg.SchedulerIntervalHours)
	}
}

func TestLoadConfigReadsEnvValues(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/banking_service?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	t.Setenv("JWT_TTL_HOURS", "48")
	t.Setenv("PGP_SYM_KEY", "test-pgp-key")
	t.Setenv("CARD_HMAC_KEY", "test-hmac-key")
	t.Setenv("SCHEDULER_INTERVAL_HOURS", "6")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	if cfg.AppEnv != "test" {
		t.Fatalf("expected APP_ENV=test, got %q", cfg.AppEnv)
	}

	if cfg.AppPort != "9090" {
		t.Fatalf("expected APP_PORT=9090, got %q", cfg.AppPort)
	}

	if cfg.LogLevel != "debug" {
		t.Fatalf("expected LOG_LEVEL=debug, got %q", cfg.LogLevel)
	}

	if cfg.JWTTTLHours != 48 {
		t.Fatalf("expected JWT_TTL_HOURS=48, got %d", cfg.JWTTTLHours)
	}

	if cfg.SchedulerIntervalHours != 6 {
		t.Fatalf("expected SCHEDULER_INTERVAL_HOURS=6, got %d", cfg.SchedulerIntervalHours)
	}
}

func TestLoadConfigFailsWhenRequiredSecretsAreMissing(t *testing.T) {
	clearConfigEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when required config values are missing")
	}
}

func TestSafeForLogDoesNotExposeSecrets(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/banking_service?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	t.Setenv("PGP_SYM_KEY", "test-pgp-key")
	t.Setenv("CARD_HMAC_KEY", "test-hmac-key")
	t.Setenv("SMTP_PASSWORD", "test-smtp-password")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	safe := cfg.SafeForLog()

	for _, forbiddenKey := range []string{
		"database_url",
		"jwt_secret",
		"pgp_sym_key",
		"card_hmac_key",
		"smtp_password",
	} {
		if _, exists := safe[forbiddenKey]; exists {
			t.Fatalf("secret key %q must not be present in SafeForLog", forbiddenKey)
		}
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"APP_ENV",
		"APP_PORT",
		"LOG_LEVEL",
		"DATABASE_URL",
		"JWT_SECRET",
		"JWT_TTL_HOURS",
		"PGP_SYM_KEY",
		"CARD_HMAC_KEY",
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_USER",
		"SMTP_PASSWORD",
		"SMTP_FROM",
		"CBR_SOAP_URL",
		"SCHEDULER_INTERVAL_HOURS",
	}

	for _, key := range keys {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("failed to unset env %s: %v", key, err)
		}
	}
}
