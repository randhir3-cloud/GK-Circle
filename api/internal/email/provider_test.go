package email

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"gopkg.in/gomail.v2"
)

// mockClock implements Clock returning a static time.
type mockClock struct {
	now time.Time
}

func (m mockClock) Now() time.Time {
	return m.now
}

// mockSleeper implements Sleeper recording sleep durations without real-time delay.
type mockSleeper struct {
	mu     sync.Mutex
	sleeps []time.Duration
}

func (m *mockSleeper) Sleep(ctx context.Context, delay time.Duration) error {
	m.mu.Lock()
	m.sleeps = append(m.sleeps, delay)
	m.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (m *mockSleeper) getSleeps() []time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]time.Duration, len(m.sleeps))
	copy(copied, m.sleeps)
	return copied
}

type mockSMTPDialer struct {
	lastMessages []*gomail.Message
	err          error
}

func (m *mockSMTPDialer) DialAndSend(msgs ...*gomail.Message) error {
	m.lastMessages = msgs
	return m.err
}

// rewriteTransport intercepts all requests to api.resend.com and routes them to a local test server.
type rewriteTransport struct {
	targetURL string
	original  http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "api.resend.com" {
		target, err := url.Parse(t.targetURL)
		if err != nil {
			return nil, err
		}
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
	}
	return t.original.RoundTrip(req)
}

func TestSMTPProvider_Send(t *testing.T) {
	cfg := config.EmailConfig{
		Provider: config.ProviderSMTP,
		FromName: "GK Circle",
		From:     "notifications@gkcircle.com",
		SMTP: config.TransactionalSMTPConfig{
			Host: "127.0.0.1",
			Port: 1025,
		},
	}

	mClock := mockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	dialer := &mockSMTPDialer{}
	provider := NewSMTPProviderWithDialer(cfg, mClock, dialer)

	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "key-123",
		To:             []EmailRecipient{{Name: "Candidate", Address: "candidate@gkcircle.com"}},
		Subject:        "Welcome",
		HTMLBody:       "<p>Welcome</p>",
		PlainText:      "Welcome",
	}

	result, err := provider.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.ProviderMessageID != "" {
		t.Errorf("Expected empty ProviderMessageID for SMTP, got: %s", result.ProviderMessageID)
	}

	if !result.AcceptedAt.Equal(mClock.now) {
		t.Errorf("Expected AcceptedAt %v, got %v", mClock.now, result.AcceptedAt)
	}

	if len(dialer.lastMessages) != 1 {
		t.Fatalf("Expected 1 email sent, got %d", len(dialer.lastMessages))
	}
}

func TestResendAPIProvider_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer re_secretkey" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Idempotency-Key") != "inv-key-1" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var body resendEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if body.From != "GK Circle <notifications@gkcircle.com>" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "resend-msg-uuid-99"}`))
	}))
	defer server.Close()

	cfg := config.EmailConfig{
		Provider:    config.ProviderResend,
		FromName:    "GK Circle",
		From:        "notifications@gkcircle.com",
		MaxAttempts: 3,
		Resend: config.ResendConfig{
			APIKey: "re_secretkey",
		},
	}

	mClock := mockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	mSleeper := &mockSleeper{}

	client := server.Client()
	client.Transport = &rewriteTransport{
		targetURL: server.URL,
		original:  http.DefaultTransport,
	}

	provider, err := NewResendAPIProvider(cfg, client, mClock, mSleeper)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	msg := EmailMessage{
		MessageID:      "msg-1",
		IdempotencyKey: "inv-key-1",
		To:             []EmailRecipient{{Name: "User", Address: "user@gkcircle.com"}},
		Subject:        "Hello",
		HTMLBody:       "<p>HTML</p>",
		PlainText:      "Plain",
	}

	result, err := provider.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.ProviderMessageID != "resend-msg-uuid-99" {
		t.Errorf("Expected ProviderMessageID 'resend-msg-uuid-99', got: %s", result.ProviderMessageID)
	}
}

func TestResendAPIProvider_RetryOnRateLimit(t *testing.T) {
	var mu sync.Mutex
	calls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		currCalls := calls
		mu.Unlock()

		if currCalls == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limited"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "success-id"}`))
	}))
	defer server.Close()

	cfg := config.EmailConfig{
		Provider:       config.ProviderResend,
		FromName:       "GK",
		From:           "notifications@gkcircle.com",
		MaxAttempts:    3,
		RetryBaseDelay: 100 * time.Millisecond,
		RetryMaxDelay:  5 * time.Second,
		Resend: config.ResendConfig{
			APIKey: "re_secret",
		},
	}

	mClock := mockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	mSleeper := &mockSleeper{}
	client := server.Client()
	client.Transport = &rewriteTransport{
		targetURL: server.URL,
		original:  http.DefaultTransport,
	}

	provider, err := NewResendAPIProvider(cfg, client, mClock, mSleeper)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	msg := EmailMessage{
		MessageID:      "msg-1",
		IdempotencyKey: "key-1",
		To:             []EmailRecipient{{Name: "User", Address: "u@gkcircle.com"}},
		Subject:        "Hello",
		HTMLBody:       "HTML",
	}

	result, err := provider.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Expected successful retry, got err: %v", err)
	}

	if result.ProviderMessageID != "success-id" {
		t.Errorf("Expected 'success-id', got: %s", result.ProviderMessageID)
	}

	sleeps := mSleeper.getSleeps()
	if len(sleeps) != 1 {
		t.Fatalf("Expected exactly 1 sleep, got %d", len(sleeps))
	}

	// Should respect Retry-After header of 2 seconds
	if sleeps[0] != 2*time.Second {
		t.Errorf("Expected sleep duration of 2s from Retry-After header, got %v", sleeps[0])
	}
}

func TestResendAPIProvider_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "ok"}`))
	}))
	defer server.Close()

	cfg := config.EmailConfig{
		Provider:    config.ProviderResend,
		FromName:    "GK",
		From:        "notifications@gkcircle.com",
		MaxAttempts: 3,
		Resend: config.ResendConfig{
			APIKey: "re_secret",
		},
	}

	client := server.Client()
	client.Transport = &rewriteTransport{
		targetURL: server.URL,
		original:  http.DefaultTransport,
	}

	provider, _ := NewResendAPIProvider(cfg, client, mockClock{}, &mockSleeper{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	msg := EmailMessage{
		MessageID:      "msg-1",
		IdempotencyKey: "key-1",
		To:             []EmailRecipient{{Name: "User", Address: "u@gkcircle.com"}},
		Subject:        "Hello",
		HTMLBody:       "HTML",
	}

	_, err := provider.Send(ctx, msg)
	if err == nil {
		t.Fatal("Expected context cancelled error, got nil")
	}

	var pErr *ProviderError
	if !errors.As(err, &pErr) {
		t.Fatalf("Expected ProviderError, got %T", err)
	}
	if pErr.Kind != ProviderErrorCancelled {
		t.Errorf("Expected ProviderErrorCancelled, got %s", pErr.Kind)
	}
}

func TestResendAPIProvider_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid-json`))
	}))
	defer server.Close()

	cfg := config.EmailConfig{
		Provider:    config.ProviderResend,
		FromName:    "GK",
		From:        "notifications@gkcircle.com",
		MaxAttempts: 1,
		Resend: config.ResendConfig{
			APIKey: "re_secret",
		},
	}

	client := server.Client()
	client.Transport = &rewriteTransport{
		targetURL: server.URL,
		original:  http.DefaultTransport,
	}

	provider, _ := NewResendAPIProvider(cfg, client, mockClock{}, &mockSleeper{})

	msg := EmailMessage{
		MessageID:      "msg-1",
		IdempotencyKey: "key-1",
		To:             []EmailRecipient{{Name: "User", Address: "u@gkcircle.com"}},
		Subject:        "Hello",
		HTMLBody:       "HTML",
	}

	_, err := provider.Send(context.Background(), msg)
	if err == nil {
		t.Fatal("Expected JSON decode error, got nil")
	}
}

func TestResendAPIProvider_MissingMessageID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": ""}`))
	}))
	defer server.Close()

	cfg := config.EmailConfig{
		Provider:    config.ProviderResend,
		FromName:    "GK",
		From:        "notifications@gkcircle.com",
		MaxAttempts: 1,
		Resend: config.ResendConfig{
			APIKey: "re_secret",
		},
	}

	client := server.Client()
	client.Transport = &rewriteTransport{
		targetURL: server.URL,
		original:  http.DefaultTransport,
	}

	provider, _ := NewResendAPIProvider(cfg, client, mockClock{}, &mockSleeper{})

	msg := EmailMessage{
		MessageID:      "msg-1",
		IdempotencyKey: "key-1",
		To:             []EmailRecipient{{Name: "User", Address: "u@gkcircle.com"}},
		Subject:        "Hello",
		HTMLBody:       "HTML",
	}

	_, err := provider.Send(context.Background(), msg)
	if err == nil {
		t.Fatal("Expected missing message ID error, got nil")
	}
	var pErr *ProviderError
	if !errors.As(err, &pErr) {
		t.Fatalf("Expected ProviderError, got %T", err)
	}
	if pErr.Err == nil || !strings.Contains(pErr.Err.Error(), "missing provider message ID") {
		t.Errorf("Unexpected underlying error: %v", pErr.Err)
	}
}
