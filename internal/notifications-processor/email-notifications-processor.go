package notifications_processor

import (
	"context"
	"evrone_course_final/config"
	"evrone_course_final/internal/entity"
	"fmt"
	"log/slog"

	mail "github.com/xhit/go-simple-mail/v2"
)

type EmailNotificationsProcessor struct {
	cfg        *config.Config
	smtpClient *mail.SMTPClient
}

func NewEmailNotificationsProcessor(cfg *config.Config, smtpClient *mail.SMTPClient) *EmailNotificationsProcessor {
	return &EmailNotificationsProcessor{
		cfg:        cfg,
		smtpClient: smtpClient,
	}
}

func (e *EmailNotificationsProcessor) Process(ctx context.Context, notification *entity.Notification) error {
	email := mail.NewMSG()
	email.SetFrom(fmt.Sprintf("From Example <%s>", e.cfg.FromEmail)).
		AddTo(notification.UserEmail).
		SetSubject(notification.Subject)

	email.SetBody(mail.TextHTML, notification.Body)

	if email.Error != nil {
		return reportAndWrapErrorEmail(email.Error, notification.CurrentRetry)
	}

	err := email.Send(e.smtpClient)
	if err != nil {
		return reportAndWrapErrorEmail(err, notification.CurrentRetry)
	}

	return nil
}

func (e *EmailNotificationsProcessor) Terminate() {
	slog.Info("Terminating EmailNotificationsProcessor")
}

func reportAndWrapErrorEmail(err error, currentRetry int) error {
	slog.Error("EmailNotificationsProcessor error send notification", slog.String("error", err.Error()), slog.Int("current retry", currentRetry))
	return fmt.Errorf("EmailNotificationsProcessor error send notification: %w", err)
}
