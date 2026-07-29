package email

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"go.uber.org/zap"
)

// NewProvider initializes the designated EmailProvider based on configuration settings.
func NewProvider(
	cfg config.EmailConfig,
	logger *zap.Logger,
	client *http.Client,
	clock Clock,
	sleeper Sleeper,
) (EmailProvider, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	switch cfg.Provider {
	case config.ProviderResend:
		if cfg.Resend.APIKey == "" {
			return nil, errors.New("cannot initialize Resend provider: API key is empty")
		}
		return NewResendAPIProvider(cfg, client, clock, sleeper)

	case config.ProviderSMTP:
		if cfg.SMTP.Host == "" {
			return nil, errors.New("cannot initialize SMTP provider: host is empty")
		}
		return NewSMTPProvider(cfg, clock)

	default:
		return nil, fmt.Errorf("unsupported email provider name: %s", cfg.Provider)
	}
}
