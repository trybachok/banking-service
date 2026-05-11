package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv                 string
	AppPort                string
	LogLevel               string
	DatabaseURL            string
	JWTSecret              string
	JWTTTLHours            int
	PGPSymKey              string
	CardHMACKey            string
	SMTPHost               string
	SMTPPort               string
	SMTPUser               string
	SMTPPassword           string
	SMTPFrom               string
	CBRSoapURL             string
	SchedulerIntervalHours int
}

func Load() (*Config, error) {
	_ = loadDotEnv(".env")

	cfg := &Config{
		AppEnv:                 getEnv("APP_ENV", "local"),
		AppPort:                getEnv("APP_PORT", "8085"),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		JWTSecret:              getEnv("JWT_SECRET", ""),
		JWTTTLHours:            getEnvInt("JWT_TTL_HOURS", 24),
		PGPSymKey:              getEnv("PGP_SYM_KEY", ""),
		CardHMACKey:            getEnv("CARD_HMAC_KEY", ""),
		SMTPHost:               getEnv("SMTP_HOST", "localhost"),
		SMTPPort:               getEnv("SMTP_PORT", "1025"),
		SMTPUser:               getEnv("SMTP_USER", ""),
		SMTPPassword:           getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:               getEnv("SMTP_FROM", "no-reply@banking.local"),
		CBRSoapURL:             getEnv("CBR_SOAP_URL", "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"),
		SchedulerIntervalHours: getEnvInt("SCHEDULER_INTERVAL_HOURS", 12),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var problems []string

	if c.AppPort == "" {
		problems = append(problems, "APP_PORT is required")
	}

	if _, err := strconv.Atoi(c.AppPort); err != nil {
		problems = append(problems, "APP_PORT must be a number")
	}

	if c.LogLevel == "" {
		problems = append(problems, "LOG_LEVEL is required")
	}

	if c.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL is required")
	}

	if c.JWTSecret == "" {
		problems = append(problems, "JWT_SECRET is required")
	}

	if c.PGPSymKey == "" {
		problems = append(problems, "PGP_SYM_KEY is required")
	}

	if c.CardHMACKey == "" {
		problems = append(problems, "CARD_HMAC_KEY is required")
	}

	if c.JWTTTLHours <= 0 {
		problems = append(problems, "JWT_TTL_HOURS must be greater than 0")
	}

	if c.SchedulerIntervalHours <= 0 {
		problems = append(problems, "SCHEDULER_INTERVAL_HOURS must be greater than 0")
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}

	return nil
}

func (c *Config) SafeForLog() map[string]any {
	return map[string]any{
		"app_env":                  c.AppEnv,
		"app_port":                 c.AppPort,
		"log_level":                c.LogLevel,
		"jwt_ttl_hours":            c.JWTTTLHours,
		"smtp_host":                c.SMTPHost,
		"smtp_port":                c.SMTPPort,
		"smtp_from":                c.SMTPFrom,
		"cbr_soap_url":             c.CBRSoapURL,
		"scheduler_interval_hours": c.SchedulerIntervalHours,
	}
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set env %s: %w", key, err)
			}
		}
	}

	return scanner.Err()
}

func getEnv(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}
