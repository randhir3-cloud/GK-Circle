package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/internal/retry"
)

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
	ReplyTo string   `json:"reply_to,omitempty"`
}

type resendEmailResponse struct {
	ID string `json:"id"`
}

// ResendAPIProvider implements EmailProvider using the official Resend HTTP API.
type ResendAPIProvider struct {
	client  *http.Client
	config  config.EmailConfig
	clock   Clock
	sleeper Sleeper
}

// NewResendAPIProvider constructs a ResendAPIProvider instance.
func NewResendAPIProvider(cfg config.EmailConfig, client *http.Client, clock Clock, sleeper Sleeper) (*ResendAPIProvider, error) {
	if cfg.Resend.APIKey == "" {
		return nil, errors.New("Resend API key cannot be empty")
	}
	if client == nil {
		client = &http.Client{
			Timeout: cfg.HTTPTimeout,
		}
	}
	return &ResendAPIProvider{
		client:  client,
		config:  cfg,
		clock:   clock,
		sleeper: sleeper,
	}, nil
}

// Send dispatches the EmailMessage to Resend HTTP API with bounded retries.
func (p *ResendAPIProvider) Send(ctx context.Context, message EmailMessage) (SendResult, error) {
	policy := retry.Policy{
		MaxAttempts:    p.config.MaxAttempts,
		RetryBaseDelay: p.config.RetryBaseDelay,
		RetryMaxDelay:  p.config.RetryMaxDelay,
	}

	var lastErr error
	var attempt int

	for {
		if err := ctx.Err(); err != nil {
			return SendResult{}, &ProviderError{
				Provider: "resend",
				Kind:     classifyContextError(err),
				Err:      err,
			}
		}

		result, err, isRetryable, retryAfter := p.trySend(ctx, message)
		if err == nil {
			return result, nil
		}

		lastErr = err
		attempt++

		if attempt >= p.config.MaxAttempts || !isRetryable {
			break
		}

		delay := policy.CalculateDelay(attempt-1, retryAfter)
		if delay <= 0 {
			break
		}

		// Sleep utilizing the injected sleeper
		if err := p.sleeper.Sleep(ctx, delay); err != nil {
			return SendResult{}, &ProviderError{
				Provider: "resend",
				Kind:     classifyContextError(err),
				Err:      err,
			}
		}
	}

	return SendResult{}, lastErr
}

func (p *ResendAPIProvider) trySend(ctx context.Context, message EmailMessage) (SendResult, error, bool, time.Duration) {
	from := fmt.Sprintf("%s <%s>", p.config.FromName, p.config.From)
	reqBody := resendEmailRequest{
		From:    from,
		Subject: message.Subject,
		HTML:    message.HTMLBody,
		Text:    message.PlainText,
	}

	for _, rec := range message.To {
		reqBody.To = append(reqBody.To, fmt.Sprintf("%s <%s>", rec.Name, rec.Address))
	}

	if message.ReplyTo != nil {
		reqBody.ReplyTo = fmt.Sprintf("%s <%s>", message.ReplyTo.Name, message.ReplyTo.Address)
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return SendResult{}, &ProviderError{
			Provider: "resend",
			Kind:     ProviderErrorPermanent,
			Err:      err,
		}, false, 0
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(jsonBytes))
	if err != nil {
		return SendResult{}, &ProviderError{
			Provider: "resend",
			Kind:     ProviderErrorPermanent,
			Err:      err,
		}, false, 0
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.Resend.APIKey)
	if message.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", message.IdempotencyKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return SendResult{}, &ProviderError{
			Provider: "resend",
			Kind:     classifyHTTPClientError(err),
			Err:      err,
		}, true, 0
	}
	defer resp.Body.Close()

	// Limit response read to 1MB to prevent memory exhaustion
	bodyReader := io.LimitReader(resp.Body, 1<<20)
	respBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return SendResult{}, &ProviderError{
			Provider:   "resend",
			Kind:       ProviderErrorTransient,
			StatusCode: resp.StatusCode,
			Err:        err,
		}, true, 0
	}

	var retryAfter time.Duration
	if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
		if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
			retryAfter = time.Duration(seconds) * time.Second
		}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var res resendEmailResponse
		if err := json.Unmarshal(respBytes, &res); err != nil {
			return SendResult{}, &ProviderError{
				Provider:   "resend",
				Kind:       ProviderErrorTransient,
				StatusCode: resp.StatusCode,
				Err:        err,
			}, true, 0
		}
		if res.ID == "" {
			return SendResult{}, &ProviderError{
				Provider:   "resend",
				Kind:       ProviderErrorTransient,
				StatusCode: resp.StatusCode,
				Err:        errors.New("missing provider message ID"),
			}, true, 0
		}

		return SendResult{
			ProviderMessageID: res.ID,
			AcceptedAt:        p.clock.Now(),
		}, nil, false, 0
	}

	kind := ProviderErrorPermanent
	isRetryable := false
	switch resp.StatusCode {
	case 401, 403:
		kind = ProviderErrorAuthentication
	case 429:
		kind = ProviderErrorRateLimited
		isRetryable = true
	case 408:
		kind = ProviderErrorTimeout
		isRetryable = true
	case 500, 502, 503, 504:
		kind = ProviderErrorTransient
		isRetryable = true
	}

	// Do not include full HTTP response body directly in public error text to avoid leaking secrets
	return SendResult{}, &ProviderError{
		Provider:   "resend",
		Kind:       kind,
		StatusCode: resp.StatusCode,
		Err:        fmt.Errorf("received non-2xx status code %d", resp.StatusCode),
	}, isRetryable, retryAfter
}

func classifyHTTPClientError(err error) ProviderErrorKind {
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ProviderErrorTimeout
		}
	}
	if errors.Is(err, context.Canceled) {
		return ProviderErrorCancelled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ProviderErrorTimeout
	}
	return ProviderErrorTransient
}
