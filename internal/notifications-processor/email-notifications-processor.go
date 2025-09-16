package notifications_processor

import (
	"context"
	"evrone_course_final/config"
	"evrone_course_final/internal/entity"
	"fmt"
	"log/slog"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type EmailNotificationsProcessor struct {
	cfg *config.Config
}

func NewEmailNotificationsProcessor(cfg *config.Config) *EmailNotificationsProcessor {
	return &EmailNotificationsProcessor{cfg: cfg}
}

func (e *EmailNotificationsProcessor) Process(ctx context.Context, notification *entity.Notification) error {
	server := mail.NewSMTPClient()

	server.Host = e.cfg.SmtpServerHost
	server.Port = e.cfg.SmtpServerPort
	server.Username = e.cfg.SmtpUsername
	server.Password = e.cfg.SmtpPassword
	server.Encryption = mail.EncryptionSTARTTLS

	// раньше думал переделать, чтобы держать коннект открытым, но в итоге решил, что пока не нужно
	server.KeepAlive = false

	server.ConnectTimeout = time.Duration(e.cfg.SmtpTimeoutSeconds) * time.Second

	server.SendTimeout = time.Duration(e.cfg.SmtpTimeoutSeconds) * time.Second

	smtpClient, err := server.Connect()

	if err != nil {
		return reportAndWrapErrorEmail(err, notification.CurrentRetry)
	}

	email := mail.NewMSG()
	email.SetFrom(fmt.Sprintf("From Example <%s>", e.cfg.FromEmail)).
		AddTo(notification.UserEmail).
		SetSubject(notification.Subject)

	email.SetBody(mail.TextHTML, notification.Body)

	if email.Error != nil {
		return reportAndWrapErrorEmail(email.Error, notification.CurrentRetry)
	}

	err = email.Send(smtpClient)
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
