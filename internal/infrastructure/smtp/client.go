package smtp

import (
	"fmt"
	"time"

	"github.com/mwsbkru/evrone-go-final/config"

	mail "github.com/xhit/go-simple-mail/v2"
)

// Client wraps SMTP client functionality
type Client struct {
	client *mail.SMTPClient
}

// NewClient creates and configures an SMTP client based on the provided configuration
func NewClient(cfg *config.Config) (*Client, error) {
	server := mail.NewSMTPClient()
	server.Host = cfg.Email.SmtpServerHost
	server.Port = cfg.Email.SmtpServerPort
	server.Username = cfg.Email.SmtpUsername
	server.Password = cfg.Email.SmtpPassword
	server.Encryption = mail.EncryptionSTARTTLS

	server.KeepAlive = true

	server.ConnectTimeout = time.Duration(cfg.Email.SmtpTimeoutSeconds) * time.Second
	server.SendTimeout = time.Duration(cfg.Email.SmtpTimeoutSeconds) * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return nil, fmt.Errorf("can't create SMTP client: %w", err)
	}

	return &Client{client: smtpClient}, nil
}

// Close closes the SMTP client
func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying mail.SMTPClient
func (c *Client) GetClient() *mail.SMTPClient {
	return c.client
}
