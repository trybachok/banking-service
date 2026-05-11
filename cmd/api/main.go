package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

	log.WithFields(cfg.SafeForLog()).Info("api bootstrap started")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "api",
		})
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ready",
			"service": "api",
		})
	})

	addr := ":" + cfg.AppPort

	log.WithField("addr", addr).Info("api server listening")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.WithError(err).Fatal("api server stopped")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, `{"error":{"code":"internal_error","message":"failed to encode response"}}`, http.StatusInternalServerError)
	}
}
