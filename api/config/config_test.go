package config_test

import (
	"os"
	"testing"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Binding(t *testing.T) {
	// Set representative environment variables
	t.Setenv("APP_PORT", "127.0.0.1:3000")
	t.Setenv("APP_ENV", "production")
	t.Setenv("IS_DEVELOPMENT", "false")
	t.Setenv("DB_DIALECT", "postgres")
	t.Setenv("DB_HOST", "db-server")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USERNAME", "admin")
	t.Setenv("DB_PASSWORD", "secret_pass")
	t.Setenv("DB_NAME", "gk_circle_prod")
	t.Setenv("MIGRATION_DIR", "database/migrations")
	t.Setenv("REDIS_HOST", "redis-server")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REPORT_SCHEDULER_TIMEOUT_SECONDS", "30")
	t.Setenv("DB_MAX_OPEN_CONNS", "25")
	t.Setenv("DB_MAX_IDLE_CONNS", "15")

	cfg, err := config.LoadConfig()
	assert.NoError(t, err)

	// Assert root-level bindings
	assert.Equal(t, "127.0.0.1:3000", cfg.Port)
	assert.Equal(t, "production", cfg.Env)
	assert.False(t, cfg.IsDevelopment)

	// Assert nested DB bindings
	assert.Equal(t, "postgres", cfg.DB.Dialect)
	assert.Equal(t, "db-server", cfg.DB.Host)
	assert.Equal(t, 5432, cfg.DB.Port)
	assert.Equal(t, "admin", cfg.DB.Username)
	assert.Equal(t, "secret_pass", cfg.DB.Password)
	assert.Equal(t, "gk_circle_prod", cfg.DB.Db)
	assert.Equal(t, 25, cfg.DB.MaxOpenConns)
	assert.Equal(t, 15, cfg.DB.MaxIdleConns)

	// Assert nested Redis bindings
	assert.Equal(t, "redis-server", cfg.RedisClient.RedisAddr)
	assert.Equal(t, "6379", cfg.RedisClient.RedisPort)

	// Assert scheduler settings
	assert.Equal(t, 30, cfg.Report.SchedulerTimeoutSeconds)
}

func TestLoadConfig_FallbackPort(t *testing.T) {
	t.Setenv("MIGRATION_DIR", "db/migs")
	t.Setenv("DB_DIALECT", "sqlite3")
	t.Setenv("APP_PORT", "")
	t.Setenv("PORT", "8080")

	cfg, err := config.LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, ":8080", cfg.Port)
}

func TestLoadConfig_ContradictoryEnv(t *testing.T) {
	t.Setenv("MIGRATION_DIR", "db/migs")
	t.Setenv("DB_DIALECT", "sqlite3")
	t.Setenv("APP_ENV", "production")
	t.Setenv("IS_DEVELOPMENT", "true")

	_, err := config.LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contradictory environment")
}

func TestLoadConfig_InvalidEnv(t *testing.T) {
	t.Setenv("MIGRATION_DIR", "db/migs")
	t.Setenv("DB_DIALECT", "sqlite3")
	t.Setenv("APP_ENV", "invalid_mode")

	_, err := config.LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid APP_ENV")
}

func TestValidateMigrationConfig(t *testing.T) {
	t.Run("Valid Postgres Migration config", func(t *testing.T) {
		t.Setenv("DB_DIALECT", "postgres")
		t.Setenv("DB_HOST", "db-host")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_USERNAME", "usr")
		t.Setenv("DB_PASSWORD", "pwd")
		t.Setenv("DB_NAME", "dbname")
		t.Setenv("MIGRATION_DIR", "db/migs")

		cfg, err := config.LoadConfig()
		assert.NoError(t, err)

		err = config.ValidateMigrationConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("Missing Host in Postgres", func(t *testing.T) {
		t.Setenv("DB_DIALECT", "postgres")
		t.Setenv("DB_HOST", "")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_USERNAME", "usr")
		t.Setenv("DB_PASSWORD", "pwd")
		t.Setenv("DB_NAME", "dbname")
		t.Setenv("MIGRATION_DIR", "db/migs")

		cfg, err := config.LoadConfig()
		assert.NoError(t, err)

		err = config.ValidateMigrationConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DB_HOST")
	})

	t.Run("Valid Sqlite3 Migration config", func(t *testing.T) {
		t.Setenv("DB_DIALECT", "sqlite3")
		t.Setenv("SQLITE_FILEPATH", "test.db")
		t.Setenv("MIGRATION_DIR", "db/migs")

		cfg, err := config.LoadConfig()
		assert.NoError(t, err)

		err = config.ValidateMigrationConfig(cfg)
		assert.NoError(t, err)
	})
}

func TestValidateAPIConfig(t *testing.T) {
	t.Run("Valid API Config", func(t *testing.T) {
		t.Setenv("APP_PORT", "3000")
		t.Setenv("APP_ENV", "production")
		t.Setenv("IS_DEVELOPMENT", "false")
		t.Setenv("DB_DIALECT", "postgres")
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_USERNAME", "gk")
		t.Setenv("DB_PASSWORD", "gk")
		t.Setenv("DB_NAME", "gk")
		t.Setenv("MIGRATION_DIR", "db/migs")
		t.Setenv("REDIS_HOST", "redis")
		t.Setenv("REDIS_PORT", "6379")
		t.Setenv("REPORT_SCHEDULER_TIMEOUT_SECONDS", "10")
		t.Setenv("WEB_URL", "https://gkcircle.com")
		t.Setenv("SELF_SERVICE_DEFAULT_BROWSER_RETURN_URL", "https://gkcircle.com")

		cfg, err := config.LoadConfig()
		assert.NoError(t, err)

		err = config.ValidateAPIConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("Missing Production Env Vars error lists all missing and avoids secrets", func(t *testing.T) {
		// Clean env vars
		os.Clearenv()
		t.Setenv("APP_PORT", "")
		t.Setenv("APP_ENV", "production")
		t.Setenv("IS_DEVELOPMENT", "false")
		t.Setenv("DB_DIALECT", "postgres")
		t.Setenv("DB_HOST", "")
		t.Setenv("DB_PORT", "0")
		t.Setenv("DB_USERNAME", "")
		t.Setenv("DB_PASSWORD", "")
		t.Setenv("DB_NAME", "")
		t.Setenv("MIGRATION_DIR", "db/migs")
		t.Setenv("REDIS_HOST", "")
		t.Setenv("REDIS_PORT", "")

		cfg, err := config.LoadConfig()
		assert.NoError(t, err)

		err = config.ValidateAPIConfig(cfg)
		assert.Error(t, err)
		
		errStr := err.Error()
		assert.Contains(t, errStr, "APP_PORT")
		assert.Contains(t, errStr, "DB_HOST")
		assert.Contains(t, errStr, "DB_PORT")
		assert.Contains(t, errStr, "DB_USERNAME")
		assert.Contains(t, errStr, "DB_PASSWORD")
		assert.Contains(t, errStr, "DB_NAME")
		assert.Contains(t, errStr, "REDIS_PORT")
		assert.Contains(t, errStr, "REDIS_HOST")
		
		// Secrets must NOT be leaked
		assert.NotContains(t, errStr, "secret_pass")
	})

	t.Run("Malformed ports", func(t *testing.T) {
		t.Setenv("APP_PORT", "invalid-port")
		t.Setenv("APP_ENV", "development")
		t.Setenv("IS_DEVELOPMENT", "true")
		t.Setenv("DB_DIALECT", "sqlite3")
		t.Setenv("SQLITE_FILEPATH", "test.db")
		t.Setenv("MIGRATION_DIR", "db/migs")
		t.Setenv("REDIS_PORT", "6379")

		cfg, err := config.LoadConfig()
		assert.NoError(t, err)

		err = config.ValidateAPIConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid APP_PORT")
	})

	t.Run("Pool limits and scheduler timeout validations", func(t *testing.T) {
		t.Setenv("APP_PORT", "3000")
		t.Setenv("APP_ENV", "development")
		t.Setenv("IS_DEVELOPMENT", "true")
		t.Setenv("DB_DIALECT", "sqlite3")
		t.Setenv("SQLITE_FILEPATH", "test.db")
		t.Setenv("MIGRATION_DIR", "db/migs")
		t.Setenv("REDIS_PORT", "6379")
		t.Setenv("DB_MAX_OPEN_CONNS", "-5")

		cfg, err := config.LoadConfig()
		assert.NoError(t, err)

		err = config.ValidateAPIConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DB_MAX_OPEN_CONNS cannot be negative")
	})
}
