package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"gopkg.in/gomail.v2"
)

// SMTPDialer defines the interface to interact with gomail.Dialer, allowing mock tests.
type SMTPDialer interface {
	DialAndSend(m ...*gomail.Message) error
}

// SMTPProvider implements the EmailProvider interface for SMTP relay mail delivery.
type SMTPProvider struct {
	dialer SMTPDialer
	config config.EmailConfig
	clock  Clock
}

// NewSMTPProvider constructs a new SMTPProvider with given config and Clock.
func NewSMTPProvider(cfg config.EmailConfig, clock Clock) (*SMTPProvider, error) {
	if cfg.SMTP.Host == "" {
		return nil, errors.New("SMTP host cannot be empty")
	}
	if cfg.SMTP.Port <= 0 {
		return nil, errors.New("SMTP port must be a positive integer")
	}

	d := gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password)
	d.SSL = !cfg.SMTP.DisableSTARTTLS && (cfg.SMTP.Port == 465)

	if cfg.SMTP.InsecureSkipVerify {
		d.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return &SMTPProvider{
		dialer: d,
		config: cfg,
		clock:  clock,
	}, nil
}

// NewSMTPProviderWithDialer creates an SMTPProvider injecting a custom dialer (used in tests).
func NewSMTPProviderWithDialer(cfg config.EmailConfig, clock Clock, dialer SMTPDialer) *SMTPProvider {
	return &SMTPProvider{
		dialer: dialer,
		config: cfg,
		clock:  clock,
	}
}

// Send delivers the compiled EmailMessage via SMTP relay.
func (p *SMTPProvider) Send(ctx context.Context, message EmailMessage) (SendResult, error) {
	if err := ctx.Err(); err != nil {
		return SendResult{}, &ProviderError{
			Provider: "smtp",
			Kind:     classifyContextError(err),
			Err:      err,
		}
	}

	m := gomail.NewMessage()

	// SMTP_FROM in config overrides EMAIL_FROM
	fromAddr := p.config.From
	fromName := p.config.FromName
	if p.config.SMTP.From != "" {
		fromAddr = p.config.SMTP.From
	}

	m.SetHeader("From", m.FormatAddress(fromAddr, fromName))

	toHeaders := make([]string, len(message.To))
	for i, r := range message.To {
		toHeaders[i] = m.FormatAddress(r.Address, r.Name)
	}
	m.SetHeader("To", toHeaders...)

	if message.ReplyTo != nil {
		m.SetHeader("Reply-To", m.FormatAddress(message.ReplyTo.Address, message.ReplyTo.Name))
	}

	m.SetHeader("Subject", message.Subject)

	if message.MessageID != "" {
		m.SetHeader("Message-ID", fmt.Sprintf("<%s>", message.MessageID))
	}
	if message.IdempotencyKey != "" {
		m.SetHeader("X-Idempotency-Key", message.IdempotencyKey)
	}

	if message.PlainText != "" && message.HTMLBody != "" {
		m.SetBody("text/plain", message.PlainText)
		m.AddAlternative("text/html", message.HTMLBody)
	} else if message.HTMLBody != "" {
		m.SetBody("text/html", message.HTMLBody)
	} else {
		m.SetBody("text/plain", message.PlainText)
	}

	type sendResult struct {
		err error
	}
	ch := make(chan sendResult, 1)

	go func() {
		ch <- sendResult{err: p.dialer.DialAndSend(m)}
	}()

	select {
	case <-ctx.Done():
		return SendResult{}, &ProviderError{
			Provider: "smtp",
			Kind:     classifyContextError(ctx.Err()),
			Err:      ctx.Err(),
		}
	case res := <-ch:
		if res.err != nil {
			return SendResult{}, &ProviderError{
				Provider: "smtp",
				Kind:     classifySMTPError(res.err),
				Err:      res.err,
			}
		}
	}

	return SendResult{
		ProviderMessageID: "", // SMTP does not return a provider message ID
		AcceptedAt:        p.clock.Now(),
	}, nil
}

func classifySMTPError(err error) ProviderErrorKind {
	if err == nil {
		return ""
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ProviderErrorTimeout
		}
		return ProviderErrorTransient
	}

	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "auth") || strings.Contains(errStr, "credentials") || strings.Contains(errStr, "login") {
		return ProviderErrorAuthentication
	}
	if strings.Contains(errStr, "timeout") {
		return ProviderErrorTimeout
	}
	if strings.Contains(errStr, "dns") || strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "dial") {
		return ProviderErrorTransient
	}

	return ProviderErrorPermanent
}

func classifyContextError(err error) ProviderErrorKind {
	if errors.Is(err, context.Canceled) {
		return ProviderErrorCancelled
	}
	if errors.Is(err, context.Context.Err(context.Background())) || errors.Is(err, context.DeadlineExceeded) {
		return ProviderErrorTimeout
	}
	return ProviderErrorTransient
}
