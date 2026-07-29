package config

import (
	"errors"
	"time"
)

type EmailProviderName string

const (
	ProviderResend EmailProviderName = "resend"
	ProviderSMTP   EmailProviderName = "smtp"
)

type ResendConfig struct {
	APIKey string `envconfig:"RESEND_API_KEY"`
}

type TransactionalSMTPConfig struct {
	Host               string `envconfig:"SMTP_HOST"`
	Port               int    `envconfig:"SMTP_PORT"`
	Username           string `envconfig:"SMTP_USER"`
	Password           string `envconfig:"SMTP_PASSWORD"`
	From               string `envconfig:"SMTP_FROM"`
	DisableSTARTTLS    bool   `envconfig:"SMTP_DISABLE_STARTTLS" default:"false"`
	InsecureSkipVerify bool   `envconfig:"SMTP_INSECURE_SKIP_VERIFY" default:"false"`
}

type EmailConfig struct {
	Provider       EmailProviderName `envconfig:"EMAIL_PROVIDER" default:"smtp" required:"true"`
	FromName       string            `envconfig:"EMAIL_FROM_NAME" default:"GK Circle"`
	From           string            `envconfig:"EMAIL_FROM" default:"notifications@gkcircle.com" required:"true"`
	ReplyTo        string            `envconfig:"EMAIL_REPLY_TO"`
	HTTPTimeout    time.Duration     `envconfig:"EMAIL_HTTP_TIMEOUT" default:"10s"`
	MaxAttempts    int               `envconfig:"EMAIL_MAX_ATTEMPTS" default:"3"`
	RetryBaseDelay time.Duration     `envconfig:"EMAIL_RETRY_BASE_DELAY" default:"250ms"`
	RetryMaxDelay  time.Duration     `envconfig:"EMAIL_RETRY_MAX_DELAY" default:"2s"`
	Resend         ResendConfig
	SMTP           TransactionalSMTPConfig
}

// Validate validates the EmailConfig struct settings and enforces production constraints
func (c *EmailConfig) Validate(appEnv string) error {
	if c.Provider != ProviderResend && c.Provider != ProviderSMTP {
		return errors.New("EMAIL_PROVIDER must be either 'resend' or 'smtp'")
	}
	if c.From == "" {
		return errors.New("EMAIL_FROM is required")
	}
	if c.MaxAttempts < 1 || c.MaxAttempts > 5 {
		return errors.New("EMAIL_MAX_ATTEMPTS must be between 1 and 5")
	}
	if c.HTTPTimeout <= 0 {
		return errors.New("EMAIL_HTTP_TIMEOUT must be positive")
	}
	if c.RetryBaseDelay <= 0 {
		return errors.New("EMAIL_RETRY_BASE_DELAY must be positive")
	}
	if c.RetryMaxDelay < c.RetryBaseDelay {
		return errors.New("EMAIL_RETRY_MAX_DELAY must be greater than or equal to EMAIL_RETRY_BASE_DELAY")
	}

	switch c.Provider {
	case ProviderResend:
		if c.Resend.APIKey == "" {
			return errors.New("RESEND_API_KEY is required when EMAIL_PROVIDER is 'resend'")
		}
	case ProviderSMTP:
		if c.SMTP.Host == "" {
			return errors.New("SMTP_HOST is required when EMAIL_PROVIDER is 'smtp'")
		}
		if c.SMTP.Port <= 0 {
			return errors.New("SMTP_PORT must be positive when EMAIL_PROVIDER is 'smtp'")
		}
		if appEnv == "production" && c.SMTP.InsecureSkipVerify {
			return errors.New("SMTP_INSECURE_SKIP_VERIFY=true is rejected in production")
		}
	}

	return nil
}
