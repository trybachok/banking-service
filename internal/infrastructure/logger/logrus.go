package logger

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewLogrus(level string) (*logrus.Logger, error) {
	log := logrus.New()

	parsedLevel, err := logrus.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", level, err)
	}

	log.SetLevel(parsedLevel)
	log.SetFormatter(&logrus.JSONFormatter{})

	return log, nil
}
