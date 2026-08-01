package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRendererRun(t *testing.T) {
	// Create temporary template file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "kratos.template.yml")
	outputPath := filepath.Join(tmpDir, "kratos.yml")

	templateContent := `
selfservice:
  flows:
    error:
      ui_url: http://localhost:3000/error
  default_browser_return_url: http://localhost:3000
`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write test template: %v", err)
	}

	t.Setenv("KRATOS_TEMPLATE_PATH", templatePath)
	t.Setenv("KRATOS_OUTPUT_PATH", outputPath)

	t.Run("development bypasses production validation", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("KRATOS_ENV", "development")

		err := run()
		if err != nil {
			t.Fatalf("expected no error in development mode, got: %v", err)
		}

		// Verify output exists
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Fatalf("output file was not generated")
		}
	})

	t.Run("production rejects missing environment variables", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")

		err := run()
		if err == nil {
			t.Fatal("expected error in production due to missing vars, got nil")
		}
		if !strings.Contains(err.Error(), "mandatory environment variable") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("production validates and sets fields successfully", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("SELFSERVICE_FLOWS_ERROR_UI_URL", "https://gkcircle.com/error")
		t.Setenv("SELFSERVICE_FLOWS_LOGIN_UI_URL", "https://gkcircle.com/account/login")
		t.Setenv("SELFSERVICE_FLOWS_REGISTRATION_UI_URL", "https://gkcircle.com/account/register")
		t.Setenv("SELFSERVICE_FLOWS_RECOVERY_UI_URL", "https://gkcircle.com/recovery")
		t.Setenv("SELFSERVICE_FLOWS_VERIFICATION_UI_URL", "https://gkcircle.com/verification")
		t.Setenv("SELFSERVICE_FLOWS_SETTINGS_UI_URL", "https://gkcircle.com/account/change-password")
		t.Setenv("SELFSERVICE_FLOWS_LOGIN_AFTER_DEFAULT_BROWSER_RETURN_URL", "https://gkcircle.com/api/v1/kratos/auth")
		t.Setenv("SELFSERVICE_FLOWS_REGISTRATION_AFTER_DEFAULT_BROWSER_RETURN_URL", "https://gkcircle.com/api/v1/kratos/auth")
		t.Setenv("SELFSERVICE_FLOWS_VERIFICATION_AFTER_DEFAULT_BROWSER_RETURN_URL", "https://gkcircle.com/account/login")
		t.Setenv("SELFSERVICE_FLOWS_LOGOUT_AFTER_DEFAULT_BROWSER_RETURN_URL", "https://gkcircle.com/account/login")
		t.Setenv("SELFSERVICE_DEFAULT_BROWSER_RETURN_URL", "https://gkcircle.com")
		t.Setenv("SELFSERVICE_ALLOWED_RETURN_URLS", "https://gkcircle.com")
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://gkcircle.com")
		t.Setenv("COOKIES_DOMAIN", "gkcircle.com")

		err := run()
		if err != nil {
			t.Fatalf("expected successful rendering, got: %v", err)
		}

		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("failed to read rendered config: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "https://gkcircle.com/error") {
			t.Fatalf("rendered config is missing injected URL")
		}
	})

	t.Run("production rejects invalid URL scheme", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("SELFSERVICE_FLOWS_ERROR_UI_URL", "http://gkcircle.com/error") // invalid scheme

		err := run()
		if err == nil {
			t.Fatal("expected error due to invalid http scheme, got nil")
		}
	})

	t.Run("production rejects localhost URLs", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("SELFSERVICE_FLOWS_ERROR_UI_URL", "https://localhost:3000/error") // localhost

		err := run()
		if err == nil {
			t.Fatal("expected error due to localhost, got nil")
		}
	})

	t.Run("production rejects Railway generated domains", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("SELFSERVICE_FLOWS_ERROR_UI_URL", "https://test.up.railway.app/error") // generated domain

		err := run()
		if err == nil {
			t.Fatal("expected error due to Railway domain, got nil")
		}
	})
}
