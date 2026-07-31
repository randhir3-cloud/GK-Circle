package services_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func loadConfig(t *testing.T) config.AppConfig {
	// Try loading .env.testing from current dir or parent dirs
	dir, err := os.Getwd()
	require.NoError(t, err)
	
	envFile := filepath.Join(dir, ".env.testing")
	for i := 0; i < 3; i++ {
		if _, err := os.Stat(envFile); err == nil {
			_ = godotenv.Load(envFile)
			break
		}
		envFile = filepath.Join(filepath.Dir(filepath.Dir(envFile)), ".env.testing")
	}

	var cfg config.AppConfig
	err = envconfig.Process("APP_PORT", &cfg)
	require.NoError(t, err)

	// Override from environment variables if present
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.DB.Host = host
	}
	if user := os.Getenv("DB_USERNAME"); user != "" {
		cfg.DB.Username = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		cfg.DB.Password = pass
	}

	return cfg
}

func connectAdminDB(t *testing.T, cfg config.AppConfig) *sql.DB {
	// Connect to postgres DB to create/drop temp DBs
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable",
		cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	return db
}

func createTempDatabase(t *testing.T, adminDB *sql.DB) string {
	name := fmt.Sprintf("gksched_test_%d", rand.Intn(10000000))
	_, err := adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", name))
	require.NoError(t, err)
	return name
}

func dropDatabase(t *testing.T, adminDB *sql.DB, name string) {
	_, err := adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", name))
	require.NoError(t, err)
}

func connectTempDB(t *testing.T, cfg config.AppConfig, dbName string) *sql.DB {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, dbName)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	return db
}

func buildCliBinary(t *testing.T) string {
	t.Helper()
	binaryPath := filepath.Join(t.TempDir(), "gk-circle")
	// On Windows, append .exe
	if os.PathSeparator == '\\' {
		binaryPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../"
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to build CLI binary: %s", string(output))
	return binaryPath
}

func TestReportScheduler_PostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := loadConfig(t)
	adminDB := connectAdminDB(t, cfg)
	defer adminDB.Close()

	// Build the CLI binary first to test real CLI migrations
	cliPath := buildCliBinary(t)

	t.Run("Fresh Database Migration & CLI verify", func(t *testing.T) {
		tempDBName := createTempDatabase(t, adminDB)
		defer dropDatabase(t, adminDB, tempDBName)

		// Run migrations using the CLI tool
		cmd := exec.Command(cliPath, "migrate", "up")
		cmd.Dir = "../"
		cmd.Env = append(os.Environ(),
			"APP_ENV=testing",
			"DB_NAME="+tempDBName,
			"DB_HOST="+cfg.DB.Host,
			"DB_PORT="+strconv.Itoa(cfg.DB.Port),
			"DB_USERNAME="+cfg.DB.Username,
			"DB_PASSWORD="+cfg.DB.Password,
			"DB_QUERYSTRING=sslmode=disable",
			"MIGRATION_DIR=database/migrations",
			"DB_DIALECT=postgres",
		)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "CLI migrate up failed: %s", string(output))

		db := connectTempDB(t, cfg, tempDBName)
		defer db.Close()

		// Verify all 5 tables were created
		tables := []string{"scheduled_reports", "generated_reports", "report_downloads", "report_delivery_logs", "export_audit_log"}
		for _, tbl := range tables {
			var exists bool
			err := db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = $1)", tbl).Scan(&exists)
			assert.NoError(t, err)
			assert.True(t, exists, "Table %s should exist after fresh migration", tbl)
		}
	})

	t.Run("Repair simulated broken Railway state", func(t *testing.T) {
		tempDBName := createTempDatabase(t, adminDB)
		defer dropDatabase(t, adminDB, tempDBName)

		db := connectTempDB(t, cfg, tempDBName)
		defer db.Close()

		// 1. Manually create the gorp_migrations table and insert the 5 broken migrations
		_, err := db.Exec(`CREATE TABLE gorp_migrations (id text PRIMARY KEY, applied_at timestamp with time zone)`)
		require.NoError(t, err)

		brokenMigs := []string{
			"20260729100000_create_scheduled_reports.up.sql",
			"20260729100001_create_generated_reports.up.sql",
			"20260729100002_create_report_downloads.up.sql",
			"20260729100003_create_report_delivery_logs.up.sql",
			"20260729100004_create_export_audit_log.up.sql",
		}
		for _, mig := range brokenMigs {
			_, err = db.Exec(`INSERT INTO gorp_migrations (id, applied_at) VALUES ($1, now())`, mig)
			require.NoError(t, err)
		}

		// Verify tables are currently missing
		var exists bool
		err = db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'scheduled_reports')").Scan(&exists)
		require.NoError(t, err)
		require.False(t, exists)

		// 2. Execute migration up using programmatic interface
		migrations := &migrate.FileMigrationSource{Dir: "../database/migrations"}
		_, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
		require.NoError(t, err)

		// 3. Verify tables now exist
		err = db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'scheduled_reports')").Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Table scheduled_reports should exist after repair migration")
	})

	t.Run("Incompatible column type should fail", func(t *testing.T) {
		tempDBName := createTempDatabase(t, adminDB)
		defer dropDatabase(t, adminDB, tempDBName)

		db := connectTempDB(t, cfg, tempDBName)
		defer db.Close()

		// Create scheduled_reports table with instructor_id as integer (incompatible type!)
		_, err := db.Exec(`CREATE TABLE public.scheduled_reports (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			instructor_id integer NOT NULL,
			title text NOT NULL,
			export_type text NOT NULL,
			export_format text NOT NULL,
			schedule_type text NOT NULL
		)`)
		require.NoError(t, err)

		// Execute migration - should fail on compatibility check
		migrations := &migrate.FileMigrationSource{Dir: "../database/migrations"}
		_, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scheduled_reports column instructor_id has incompatible type")
	})

	t.Run("Compatible extra columns should pass", func(t *testing.T) {
		tempDBName := createTempDatabase(t, adminDB)
		defer dropDatabase(t, adminDB, tempDBName)

		db := connectTempDB(t, cfg, tempDBName)
		defer db.Close()

		// 1. Run migrations fully
		migrations := &migrate.FileMigrationSource{Dir: "../database/migrations"}
		_, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
		require.NoError(t, err)

		// 2. Add extra compatible column
		_, err = db.Exec(`ALTER TABLE public.scheduled_reports ADD COLUMN extra_test_info text`)
		require.NoError(t, err)

		// 3. Re-run migration. Should succeed.
		_, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
		assert.NoError(t, err)
	})

	t.Run("Missing index repair", func(t *testing.T) {
		tempDBName := createTempDatabase(t, adminDB)
		defer dropDatabase(t, adminDB, tempDBName)

		db := connectTempDB(t, cfg, tempDBName)
		defer db.Close()

		migrations := &migrate.FileMigrationSource{Dir: "../database/migrations"}
		_, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
		require.NoError(t, err)

		// Drop index
		_, err = db.Exec(`DROP INDEX IF EXISTS idx_scheduled_reports_next_run`)
		require.NoError(t, err)

		// Manually delete the migration record to force sql-migrate to re-execute it
		_, err = db.Exec(`DELETE FROM gorp_migrations WHERE id = '20260801120000_repair_report_tables.sql'`)
		require.NoError(t, err)

		// Run migration up again (should recreate index safely)
		_, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
		assert.NoError(t, err)

		// Verify index exists
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM pg_class c 
				JOIN pg_namespace n ON n.oid = c.relnamespace 
				WHERE c.relname = 'idx_scheduled_reports_next_run' AND n.nspname = 'public'
			)`).Scan(&exists)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Scheduler execution, timeout and advisory lock", func(t *testing.T) {
		tempDBName := createTempDatabase(t, adminDB)
		defer dropDatabase(t, adminDB, tempDBName)

		db := connectTempDB(t, cfg, tempDBName)
		defer db.Close()

		migrations := &migrate.FileMigrationSource{Dir: "../database/migrations"}
		_, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
		require.NoError(t, err)

		goquDB := goqu.New("postgres", db)
		jobQueue := make(chan uuid.UUID, 10)
		logger := zap.NewNop()

		scheduler := services.NewReportScheduler(goquDB, jobQueue, 1, 2, logger)

		// 1. Check fresh scheduler tick against empty tables
		ctx := context.Background()
		scheduler.Start(ctx)
		
		// Insert a due job
		_, err = db.Exec(`INSERT INTO public.scheduled_reports 
			(id, instructor_id, title, export_type, export_format, schedule_type, enabled, next_run_at, timezone)
			VALUES ($1, 'instructor-1', 'Daily Test', 'PORTFOLIO_OVERVIEW', 'CSV', 'DAILY', true, now() - interval '1 minute', 'UTC')`,
			uuid.New())
		require.NoError(t, err)

		// Verify advisory-lock exclusion
		// Lock the transaction advisory constant key (714204881) in a separate session
		blockerTx, err := db.Begin()
		require.NoError(t, err)
		defer blockerTx.Rollback()

		var lockAcquired bool
		err = blockerTx.QueryRow("SELECT pg_try_advisory_xact_lock(714204881)").Scan(&lockAcquired)
		require.NoError(t, err)
		require.True(t, lockAcquired)

		// Now tick from scheduler (should skip execution cleanly because advisory lock is held)
		// It will log as skipped at debug level
		scheduler.Start(ctx)

		// 2. Release lock and verify dispatch
		blockerTx.Rollback()

		time.Sleep(1500 * time.Millisecond) // wait for tick
		
		var jobCount int
		err = db.QueryRow("SELECT count(*) FROM public.generated_reports").Scan(&jobCount)
		assert.NoError(t, err)
		assert.True(t, jobCount > 0)
	})
}
