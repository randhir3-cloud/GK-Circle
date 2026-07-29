package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kelseyhightower/envconfig"
)

func TestLoadEmailConfig_Resend(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "resend")
	os.Setenv("EMAIL_FROM_NAME", "GK Test")
	os.Setenv("EMAIL_FROM", "test@gkcircle.com")
	os.Setenv("EMAIL_REPLY_TO", "reply@gkcircle.com")
	os.Setenv("RESEND_API_KEY", "re_testkey")
	os.Setenv("EMAIL_HTTP_TIMEOUT", "5s")
	os.Setenv("EMAIL_MAX_ATTEMPTS", "4")
	os.Setenv("EMAIL_RETRY_BASE_DELAY", "100ms")
	os.Setenv("EMAIL_RETRY_MAX_DELAY", "1s")
	defer func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("EMAIL_FROM_NAME")
		os.Unsetenv("EMAIL_FROM")
		os.Unsetenv("EMAIL_REPLY_TO")
		os.Unsetenv("RESEND_API_KEY")
		os.Unsetenv("EMAIL_HTTP_TIMEOUT")
		os.Unsetenv("EMAIL_MAX_ATTEMPTS")
		os.Unsetenv("EMAIL_RETRY_BASE_DELAY")
		os.Unsetenv("EMAIL_RETRY_MAX_DELAY")
	}()

	var cfg struct {
		Email EmailConfig
	}
	err := envconfig.Process("", &cfg)
	if err != nil {
		t.Fatalf("Failed to process config: %v", err)
	}

	emailCfg := cfg.Email
	if err := emailCfg.Validate("development"); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if emailCfg.Provider != ProviderResend {
		t.Errorf("Expected resend provider, got %s", emailCfg.Provider)
	}
	if emailCfg.FromName != "GK Test" {
		t.Errorf("Expected FromName 'GK Test', got %s", emailCfg.FromName)
	}
	if emailCfg.From != "test@gkcircle.com" {
		t.Errorf("Expected From 'test@gkcircle.com', got %s", emailCfg.From)
	}
	if emailCfg.Resend.APIKey != "re_testkey" {
		t.Errorf("Expected APIKey 're_testkey', got %s", emailCfg.Resend.APIKey)
	}
	if emailCfg.HTTPTimeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", emailCfg.HTTPTimeout)
	}
	if emailCfg.MaxAttempts != 4 {
		t.Errorf("Expected max attempts 4, got %d", emailCfg.MaxAttempts)
	}
}

func TestLoadEmailConfig_SMTP(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "smtp")
	os.Setenv("EMAIL_FROM", "test@gkcircle.com")
	os.Setenv("SMTP_HOST", "127.0.0.1")
	os.Setenv("SMTP_PORT", "1025")
	os.Setenv("SMTP_USER", "smtp_user")
	os.Setenv("SMTP_PASSWORD", "smtp_password")
	os.Setenv("SMTP_FROM", "override@gkcircle.com")
	os.Setenv("SMTP_DISABLE_STARTTLS", "true")
	os.Setenv("SMTP_INSECURE_SKIP_VERIFY", "true")
	defer func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("EMAIL_FROM")
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("SMTP_USER")
		os.Unsetenv("SMTP_PASSWORD")
		os.Unsetenv("SMTP_FROM")
		os.Unsetenv("SMTP_DISABLE_STARTTLS")
		os.Unsetenv("SMTP_INSECURE_SKIP_VERIFY")
	}()

	var cfg struct {
		Email EmailConfig
	}
	err := envconfig.Process("", &cfg)
	if err != nil {
		t.Fatalf("Failed to process config: %v", err)
	}

	emailCfg := cfg.Email
	if err := emailCfg.Validate("development"); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if emailCfg.Provider != ProviderSMTP {
		t.Errorf("Expected SMTP provider, got %s", emailCfg.Provider)
	}
	if emailCfg.SMTP.Host != "127.0.0.1" {
		t.Errorf("Expected SMTP host '127.0.0.1', got %s", emailCfg.SMTP.Host)
	}
	if emailCfg.SMTP.Port != 1025 {
		t.Errorf("Expected SMTP port 1025, got %d", emailCfg.SMTP.Port)
	}
	if emailCfg.SMTP.Username != "smtp_user" {
		t.Errorf("Expected SMTP username 'smtp_user', got %s", emailCfg.SMTP.Username)
	}
	if emailCfg.SMTP.Password != "smtp_password" {
		t.Errorf("Expected SMTP password 'smtp_password', got %s", emailCfg.SMTP.Password)
	}
	if emailCfg.SMTP.From != "override@gkcircle.com" {
		t.Errorf("Expected SMTP from override, got %s", emailCfg.SMTP.From)
	}
	if !emailCfg.SMTP.DisableSTARTTLS {
		t.Error("Expected DisableSTARTTLS to be true")
	}
	if !emailCfg.SMTP.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be true")
	}
}

func TestLoadEmailConfig_RejectsUnknownProvider(t *testing.T) {
	cfg := EmailConfig{
		Provider: "unknown",
		From:     "test@gkcircle.com",
	}
	if err := cfg.Validate("development"); err == nil {
		t.Error("Expected error for unknown provider, got nil")
	}
}

func TestLoadEmailConfig_RejectsInsecureProductionSMTP(t *testing.T) {
	cfg := EmailConfig{
		Provider:       ProviderSMTP,
		From:           "test@gkcircle.com",
		MaxAttempts:    3,
		HTTPTimeout:    time.Second,
		RetryBaseDelay: time.Second,
		RetryMaxDelay:  time.Second,
		SMTP: TransactionalSMTPConfig{
			Host:               "smtp.production.com",
			Port:               587,
			InsecureSkipVerify: true, // insecure
		},
	}
	if err := cfg.Validate("production"); err == nil {
		t.Error("Expected validation error for InsecureSkipVerify in production, got nil")
	}
}

func TestLoadEmailConfig_DoesNotExposeSecrets(t *testing.T) {
	cfg := EmailConfig{
		Provider: ProviderResend,
		From:     "test@gkcircle.com",
		Resend: ResendConfig{
			APIKey: "", // missing API key
		},
	}

	err := cfg.Validate("production")
	if err == nil {
		t.Fatal("Expected validation error for missing API key, got nil")
	}

	// Make sure the error message does not contain credentials
	errMsg := err.Error()
	if strings.Contains(errMsg, "re_") || strings.Contains(errMsg, "api") && strings.Contains(errMsg, "key:") {
		t.Errorf("Error message might contain API key info: %s", errMsg)
	}
}
