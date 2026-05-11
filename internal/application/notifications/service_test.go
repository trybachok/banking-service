package notifications

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/banking-service/internal/domain"
)

func TestServiceProcessPendingSendsEmail(t *testing.T) {
	outbox := &fakeOutbox{
		messages: []domain.EmailOutboxMessage{
			{
				ID:             "email-1",
				RecipientEmail: "user@example.com",
				Subject:        "MFA code",
				TemplateCode:   "mfa_code",
				Payload: map[string]any{
					"code":    "123456",
					"purpose": "transfer",
				},
				Attempts:  0,
				CreatedAt: time.Now().UTC(),
			},
		},
	}

	sender := &fakeSender{}

	service := NewService(outbox, sender)

	result, err := service.ProcessPending(context.Background(), time.Now().UTC(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Processed != 1 {
		t.Fatalf("expected processed 1, got %d", result.Processed)
	}

	if result.Sent != 1 {
		t.Fatalf("expected sent 1, got %d", result.Sent)
	}

	if len(sender.sent) != 1 {
		t.Fatalf("expected one sent email, got %d", len(sender.sent))
	}

	if outbox.sentID != "email-1" {
		t.Fatalf("expected email-1 to be marked sent, got %s", outbox.sentID)
	}
}

func TestServiceProcessPendingMarksFailedOnSendError(t *testing.T) {
	outbox := &fakeOutbox{
		messages: []domain.EmailOutboxMessage{
			{
				ID:             "email-1",
				RecipientEmail: "user@example.com",
				Subject:        "MFA code",
				TemplateCode:   "mfa_code",
				Payload: map[string]any{
					"code": "123456",
				},
				Attempts: 0,
			},
		},
	}

	sender := &fakeSender{err: errors.New("smtp failed")}

	service := NewService(outbox, sender)

	result, err := service.ProcessPending(context.Background(), time.Now().UTC(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Failed != 1 {
		t.Fatalf("expected failed 1, got %d", result.Failed)
	}

	if outbox.failedID != "email-1" {
		t.Fatalf("expected email-1 to be marked failed, got %s", outbox.failedID)
	}
}

type fakeSender struct {
	sent []sentEmail
	err  error
}

type sentEmail struct {
	to      string
	subject string
	body    string
}

func (s *fakeSender) Send(_ context.Context, to string, subject string, body string) error {
	if s.err != nil {
		return s.err
	}

	s.sent = append(s.sent, sentEmail{
		to:      to,
		subject: subject,
		body:    body,
	})

	return nil
}

type fakeOutbox struct {
	messages []domain.EmailOutboxMessage
	sentID   string
	failedID string
}

func (o *fakeOutbox) Create(_ context.Context, _ domain.UserID, _ string, _ string, _ string, _ map[string]any) error {
	return nil
}

func (o *fakeOutbox) FindPending(_ context.Context, _ time.Time, _ int) ([]domain.EmailOutboxMessage, error) {
	return o.messages, nil
}

func (o *fakeOutbox) MarkSent(_ context.Context, id string, _ time.Time) error {
	o.sentID = id
	return nil
}

func (o *fakeOutbox) MarkFailed(_ context.Context, id string, _ time.Time, _ string) error {
	o.failedID = id
	return nil
}
