package email

import (
	"context"
	"fmt"
	"strconv"

	"gopkg.in/gomail.v2"
)

type SMTPSender struct {
	dialer *gomail.Dialer
	from   string
}

func NewSMTPSender(host string, portRaw string, username string, password string, from string) (*SMTPSender, error) {
	port, err := strconv.Atoi(portRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP port %q: %w", portRaw, err)
	}

	if host == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}

	if from == "" {
		return nil, fmt.Errorf("SMTP from address is required")
	}

	return &SMTPSender{
		dialer: gomail.NewDialer(host, port, username, password),
		from:   from,
	}, nil
}

func (s *SMTPSender) Send(ctx context.Context, to string, subject string, body string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	message := gomail.NewMessage()
	message.SetHeader("From", s.from)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)

	if err := s.dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
