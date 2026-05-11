package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/example/banking-service/internal/domain"
)

type EmailSender interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

type ProcessResult struct {
	Processed int `json:"processed"`
	Sent      int `json:"sent"`
	Failed    int `json:"failed"`
}

type Service struct {
	outbox domain.EmailOutboxRepository
	sender EmailSender
}

func NewService(outbox domain.EmailOutboxRepository, sender EmailSender) *Service {
	return &Service{
		outbox: outbox,
		sender: sender,
	}
}

func (s *Service) ProcessPending(ctx context.Context, now time.Time, limit int) (*ProcessResult, error) {
	if limit <= 0 {
		limit = 100
	}

	messages, err := s.outbox.FindPending(ctx, now, limit)
	if err != nil {
		return nil, err
	}

	result := &ProcessResult{}

	for _, message := range messages {
		result.Processed++

		body := renderBody(message)

		if err := s.sender.Send(ctx, message.RecipientEmail, message.Subject, body); err != nil {
			nextAttemptAt := now.Add(retryDelay(message.Attempts))

			if markErr := s.outbox.MarkFailed(ctx, message.ID, nextAttemptAt, err.Error()); markErr != nil {
				return result, markErr
			}

			result.Failed++
			continue
		}

		if err := s.outbox.MarkSent(ctx, message.ID, now); err != nil {
			return result, err
		}

		result.Sent++
	}

	return result, nil
}

func renderBody(message domain.EmailOutboxMessage) string {
	switch message.TemplateCode {
	case "mfa_code":
		code, _ := message.Payload["code"].(string)
		purpose, _ := message.Payload["purpose"].(string)

		return fmt.Sprintf(
			"Your banking confirmation code is: %s\n\nPurpose: %s\n\nIf you did not request this operation, ignore this email.",
			code,
			purpose,
		)

	default:
		return fmt.Sprintf("Notification: %s\n\nPayload: %v", message.TemplateCode, message.Payload)
	}
}

func retryDelay(attempts int) time.Duration {
	switch {
	case attempts <= 0:
		return time.Minute
	case attempts == 1:
		return 5 * time.Minute
	default:
		return 15 * time.Minute
	}
}
