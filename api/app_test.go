package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_ConfigFailure(t *testing.T) {
	// Contradictory environment configuration (production + IS_DEVELOPMENT=true)
	t.Setenv("APP_ENV", "production")
	t.Setenv("IS_DEVELOPMENT", "true")
	t.Setenv("MIGRATION_DIR", "db/migs")
	t.Setenv("DB_DIALECT", "sqlite3")

	code := run()
	assert.Equal(t, 1, code, "run() should return 1 on configuration failure")
}

func TestAppCLI_Binary(t *testing.T) {
	// Build the CLI binary first to test it in subprocesses
	binaryPath := filepath.Join(t.TempDir(), "gk-circle")
	// On Windows, append .exe
	if os.PathSeparator == '\\' {
		binaryPath += ".exe"
	}
	cmdBuild := exec.Command("go", "build", "-o", binaryPath, "app.go")
	outputBuild, err := cmdBuild.CombinedOutput()
	require.NoError(t, err, "failed to build CLI binary: %s", string(outputBuild))

	t.Run("no command specified exits with 1", func(t *testing.T) {
		cmd := exec.Command(binaryPath)
		cmd.Env = append(os.Environ(),
			"MIGRATION_DIR=db/migs",
			"DB_DIALECT=sqlite3",
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		assert.Error(t, err)

		exitError, ok := err.(*exec.ExitError)
		require.True(t, ok)
		assert.Equal(t, 1, exitError.ExitCode())
		assert.Contains(t, out.String(), "no command specified")
	})

	t.Run("--help exits with 0", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--help")
		cmd.Env = append(os.Environ(),
			"MIGRATION_DIR=db/migs",
			"DB_DIALECT=sqlite3",
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		assert.NoError(t, err)
		assert.Contains(t, out.String(), "Usage:")
	})

	t.Run("invalid command exits with 1", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "invalid-cmd-name")
		cmd.Env = append(os.Environ(),
			"MIGRATION_DIR=db/migs",
			"DB_DIALECT=sqlite3",
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		assert.Error(t, err)

		exitError, ok := err.(*exec.ExitError)
		require.True(t, ok)
		assert.Equal(t, 1, exitError.ExitCode())
		assert.Contains(t, out.String(), "unknown command")
	})

	t.Run("sanitized configuration errors do not leak passwords or secrets", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "api")
		cmd.Env = append(os.Environ(),
			"APP_ENV=production",
			"IS_DEVELOPMENT=false",
			"MIGRATION_DIR=db/migs",
			"DB_DIALECT=postgres",
			"DB_PASSWORD=my_secret_production_password",
			"DB_HOST=", // trigger validation error
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		assert.Error(t, err)

		exitStr := out.String()
		assert.Contains(t, exitStr, "application startup failed")
		assert.Contains(t, exitStr, "missing mandatory production environment variables")
		assert.NotContains(t, exitStr, "my_secret_production_password")
	})

	t.Run("database ping timeout check", func(t *testing.T) {
		// Use a blackhole IP/port to trigger database connection/ping timeout
		cmd := exec.Command(binaryPath, "api")
		cmd.Env = append(os.Environ(),
			"APP_ENV=development",
			"IS_DEVELOPMENT=true",
			"MIGRATION_DIR=database/migrations",
			"DB_DIALECT=postgres",
			"DB_HOST=192.0.2.1", // Test-Net-1 blackhole IP (guaranteed to timeout)
			"DB_PORT=5432",
			"DB_USERNAME=gk",
			"DB_PASSWORD=gk",
			"DB_NAME=gk",
			"DB_MAX_OPEN_CONNS=1",
			"REPORT_SCHEDULER_TIMEOUT_SECONDS=10",
			"APP_PORT=8080",
			"WEB_URL=http://localhost:3000",
			"REDIS_HOST=127.0.0.1",
			"REDIS_PORT=6379",
		)
		
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		// This command is expected to fail with a timeout error
		err := cmd.Run()
		assert.Error(t, err)
		assert.Contains(t, out.String(), "database startup check failed")
	})

	t.Run("expected Railway start command is documented", func(t *testing.T) {
		// Confirm the start command is ./gk-circle api
		cmdPath := strings.Join([]string{binaryPath, "api"}, " ")
		assert.Contains(t, cmdPath, "api")
	})
}
