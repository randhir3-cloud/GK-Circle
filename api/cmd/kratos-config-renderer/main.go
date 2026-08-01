package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering Kratos config: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	templatePath := os.Getenv("KRATOS_TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "/etc/config/kratos/kratos.template.yml"
	}

	outputPath := os.Getenv("KRATOS_OUTPUT_PATH")
	if outputPath == "" {
		outputPath = "/tmp/kratos.yml"
	}

	appEnv := strings.ToLower(os.Getenv("APP_ENV"))
	kratosEnv := strings.ToLower(os.Getenv("KRATOS_ENV"))
	isProduction := appEnv == "production" || kratosEnv == "production"

	// Read template
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file at %s: %w", templatePath, err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(templateData, &config); err != nil {
		return fmt.Errorf("failed to unmarshal YAML template: %w", err)
	}

	if isProduction {
		if err := injectProductionConfig(config); err != nil {
			return err
		}
	} else {
		// Set basic defaults if present in environment or leave template as is
		_ = injectOptionalConfig(config)
	}

	// Marshal back to YAML
	outputData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config back to YAML: %w", err)
	}

	// Write atomically
	tmpFile := outputPath + ".tmp"
	f, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(outputData); err != nil {
		return fmt.Errorf("failed to write to temp config file: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp config file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close temp config file: %w", err)
	}

	if err := os.Rename(tmpFile, outputPath); err != nil {
		return fmt.Errorf("failed to atomically rename config file: %w", err)
	}

	return nil
}

func injectProductionConfig(config map[string]interface{}) error {
	requiredVars := []string{
		"SELFSERVICE_FLOWS_ERROR_UI_URL",
		"SELFSERVICE_FLOWS_LOGIN_UI_URL",
		"SELFSERVICE_FLOWS_REGISTRATION_UI_URL",
		"SELFSERVICE_FLOWS_RECOVERY_UI_URL",
		"SELFSERVICE_FLOWS_VERIFICATION_UI_URL",
		"SELFSERVICE_FLOWS_SETTINGS_UI_URL",
		"SELFSERVICE_FLOWS_LOGIN_AFTER_DEFAULT_BROWSER_RETURN_URL",
		"SELFSERVICE_FLOWS_REGISTRATION_AFTER_DEFAULT_BROWSER_RETURN_URL",
		"SELFSERVICE_FLOWS_VERIFICATION_AFTER_DEFAULT_BROWSER_RETURN_URL",
		"SELFSERVICE_FLOWS_LOGOUT_AFTER_DEFAULT_BROWSER_RETURN_URL",
		"SELFSERVICE_DEFAULT_BROWSER_RETURN_URL",
		"SELFSERVICE_ALLOWED_RETURN_URLS",
		"CORS_ALLOWED_ORIGINS",
		"COOKIES_DOMAIN",
	}

	// Verify all environment variables are present and not empty
	for _, envVar := range requiredVars {
		val := os.Getenv(envVar)
		if val == "" {
			return fmt.Errorf("mandatory environment variable %s is missing or empty in production", envVar)
		}
	}

	// Validate browser-facing hostnames structurally
	for _, envVar := range requiredVars {
		val := os.Getenv(envVar)
		if envVar == "COOKIES_DOMAIN" {
			if err := validateHostnameOnly(val); err != nil {
				return fmt.Errorf("invalid COOKIES_DOMAIN: %w", err)
			}
			continue
		}

		if envVar == "SELFSERVICE_ALLOWED_RETURN_URLS" || envVar == "CORS_ALLOWED_ORIGINS" {
			origins := strings.Split(val, ",")
			for _, origin := range origins {
				if err := validatePublicURL(origin, envVar); err != nil {
					return err
				}
			}
			continue
		}

		if err := validatePublicURL(val, envVar); err != nil {
			return err
		}
	}

	// Programmatically set YAML fields
	setMapPath(config, []string{"selfservice", "flows", "error", "ui_url"}, os.Getenv("SELFSERVICE_FLOWS_ERROR_UI_URL"))
	setMapPath(config, []string{"selfservice", "flows", "login", "ui_url"}, os.Getenv("SELFSERVICE_FLOWS_LOGIN_UI_URL"))
	setMapPath(config, []string{"selfservice", "flows", "registration", "ui_url"}, os.Getenv("SELFSERVICE_FLOWS_REGISTRATION_UI_URL"))
	setMapPath(config, []string{"selfservice", "flows", "recovery", "ui_url"}, os.Getenv("SELFSERVICE_FLOWS_RECOVERY_UI_URL"))
	setMapPath(config, []string{"selfservice", "flows", "verification", "ui_url"}, os.Getenv("SELFSERVICE_FLOWS_VERIFICATION_UI_URL"))
	setMapPath(config, []string{"selfservice", "flows", "settings", "ui_url"}, os.Getenv("SELFSERVICE_FLOWS_SETTINGS_UI_URL"))

	setMapPath(config, []string{"selfservice", "flows", "login", "after", "default_browser_return_url"}, os.Getenv("SELFSERVICE_FLOWS_LOGIN_AFTER_DEFAULT_BROWSER_RETURN_URL"))
	setMapPath(config, []string{"selfservice", "flows", "registration", "after", "default_browser_return_url"}, os.Getenv("SELFSERVICE_FLOWS_REGISTRATION_AFTER_DEFAULT_BROWSER_RETURN_URL"))
	setMapPath(config, []string{"selfservice", "flows", "verification", "after", "default_browser_return_url"}, os.Getenv("SELFSERVICE_FLOWS_VERIFICATION_AFTER_DEFAULT_BROWSER_RETURN_URL"))
	setMapPath(config, []string{"selfservice", "flows", "logout", "after", "default_browser_return_url"}, os.Getenv("SELFSERVICE_FLOWS_LOGOUT_AFTER_DEFAULT_BROWSER_RETURN_URL"))

	setMapPath(config, []string{"selfservice", "default_browser_return_url"}, os.Getenv("SELFSERVICE_DEFAULT_BROWSER_RETURN_URL"))

	setMapPath(config, []string{"selfservice", "allowed_return_urls"}, strings.Split(os.Getenv("SELFSERVICE_ALLOWED_RETURN_URLS"), ","))
	setMapPath(config, []string{"serve", "public", "cors", "allowed_origins"}, strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ","))
	setMapPath(config, []string{"cookies", "domain"}, os.Getenv("COOKIES_DOMAIN"))

	return nil
}

func injectOptionalConfig(config map[string]interface{}) error {
	// Local development overrides (only apply if environment variables are explicitly set)
	if val := os.Getenv("SELFSERVICE_FLOWS_ERROR_UI_URL"); val != "" {
		setMapPath(config, []string{"selfservice", "flows", "error", "ui_url"}, val)
	}
	if val := os.Getenv("SELFSERVICE_FLOWS_LOGIN_UI_URL"); val != "" {
		setMapPath(config, []string{"selfservice", "flows", "login", "ui_url"}, val)
	}
	if val := os.Getenv("SELFSERVICE_FLOWS_REGISTRATION_UI_URL"); val != "" {
		setMapPath(config, []string{"selfservice", "flows", "registration", "ui_url"}, val)
	}
	if val := os.Getenv("SELFSERVICE_FLOWS_RECOVERY_UI_URL"); val != "" {
		setMapPath(config, []string{"selfservice", "flows", "recovery", "ui_url"}, val)
	}
	if val := os.Getenv("SELFSERVICE_FLOWS_VERIFICATION_UI_URL"); val != "" {
		setMapPath(config, []string{"selfservice", "flows", "verification", "ui_url"}, val)
	}
	if val := os.Getenv("SELFSERVICE_FLOWS_SETTINGS_UI_URL"); val != "" {
		setMapPath(config, []string{"selfservice", "flows", "settings", "ui_url"}, val)
	}
	if val := os.Getenv("COOKIES_DOMAIN"); val != "" {
		setMapPath(config, []string{"cookies", "domain"}, val)
	}
	return nil
}

func validatePublicURL(rawURL, envVar string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL in %s: %w", envVar, err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("URL scheme for %s must be 'https', got '%s'", envVar, u.Scheme)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("URL hostname for %s cannot be empty", envVar)
	}

	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return fmt.Errorf("URL hostname for %s cannot resolve to localhost/loopback in production: %s", envVar, hostname)
	}

	if strings.HasSuffix(hostname, ".up.railway.app") || strings.HasSuffix(hostname, ".railway.internal") {
		return fmt.Errorf("URL hostname for %s cannot be a Railway auto-generated domain in production: %s", envVar, hostname)
	}

	return nil
}

func validateHostnameOnly(val string) error {
	if val == "" {
		return errors.New("hostname cannot be empty")
	}
	if strings.Contains(val, "://") || strings.Contains(val, "/") || strings.Contains(val, ":") {
		return fmt.Errorf("hostname contains invalid characters: %s", val)
	}
	if val == "localhost" || val == "127.0.0.1" || val == "::1" {
		return fmt.Errorf("hostname cannot resolve to localhost/loopback in production: %s", val)
	}
	if strings.HasSuffix(val, ".up.railway.app") || strings.HasSuffix(val, ".railway.internal") {
		return fmt.Errorf("hostname cannot be a Railway auto-generated domain in production: %s", val)
	}
	return nil
}

func setMapPath(m map[string]interface{}, path []string, value interface{}) {
	if len(path) == 0 {
		return
	}
	if len(path) == 1 {
		m[path[0]] = value
		return
	}

	key := path[0]
	subMapRaw, exists := m[key]
	var subMap map[string]interface{}

	if !exists {
		subMap = make(map[string]interface{})
		m[key] = subMap
	} else {
		switch typed := subMapRaw.(type) {
		case map[string]interface{}:
			subMap = typed
		case map[interface{}]interface{}:
			subMap = make(map[string]interface{})
			for k, v := range typed {
				subMap[fmt.Sprintf("%v", k)] = v
			}
			m[key] = subMap
		default:
			subMap = make(map[string]interface{})
			m[key] = subMap
		}
	}

	setMapPath(subMap, path[1:], value)
}
