package infrastructure

import (
	"context"
	"fmt"
	"net/smtp"
	"time"

	app "github.com/opencorex-org/openrevenue/internal/administration/application"
)

type SMTPNotifier struct {
	Address string
	From    string
}

func (n SMTPNotifier) Send(ctx context.Context, message app.Notification) error {
	done := make(chan error, 1)
	go func() {
		body := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", n.From, message.To, message.Subject, message.Body)
		done <- smtp.SendMail(n.Address, nil, n.From, []string{message.To}, []byte(body))
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("SMTP delivery timed out")
	}
}
