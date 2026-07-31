package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// AllConfig variable of type AppConfig
var AllConfig AppConfig

type RedisClientConfig struct {
	RedisAddr string `envconfig:"REDIS_HOST"`
	RedisPort string `envconfig:"REDIS_PORT"`
	RedisPass string `envconfig:"REDIS_PASSWORD"`
	RedisDb   int    `envconfig:"REDIS_DATABASES"`
}

// ReportConfig holds configuration for the export and scheduled reporting subsystem.
type ReportConfig struct {
	StorageProvider          string `envconfig:"REPORT_STORAGE_PROVIDER"` // local | s3
	LocalStoragePath         string `envconfig:"REPORT_LOCAL_STORAGE_PATH"`
	S3Bucket                 string `envconfig:"REPORT_S3_BUCKET"`
	S3Region                 string `envconfig:"REPORT_S3_REGION"`
	S3Endpoint               string `envconfig:"REPORT_S3_ENDPOINT"`                // MinIO-compatible
	RetentionDays            int    `envconfig:"REPORT_RETENTION_DAYS"`             // default 30
	MaxAttachmentBytes       int64  `envconfig:"REPORT_EMAIL_MAX_ATTACHMENT_BYTES"` // default 10485760
	WorkerPoolSize           int    `envconfig:"REPORT_WORKER_POOL_SIZE"`           // default 3
	SchedulerIntervalSeconds int    `envconfig:"REPORT_SCHEDULER_INTERVAL_SECONDS"` // default 60
	SchedulerTimeoutSeconds  int    `envconfig:"REPORT_SCHEDULER_TIMEOUT_SECONDS"`  // default 10
	ReclaimIntervalSeconds   int    `envconfig:"REPORT_RECLAIM_INTERVAL_SECONDS"`   // default 30
	EmailRetryAttempts       int    `envconfig:"REPORT_EMAIL_RETRY_ATTEMPTS"`       // default 3
}

// AppConfig type AppConfig
type AppConfig struct {
	IsDevelopment bool   `envconfig:"IS_DEVELOPMENT"`
	Debug         bool   `envconfig:"DEBUG"`
	Env           string `envconfig:"APP_ENV"`
	Port          string `envconfig:"APP_PORT"`
	Secret        string `envconfig:"JWT_SECRET"`
	WebUrl        string `envconfig:"WEB_URL"`
	JWTIssuer     string `envconfig:"ISSUER"`
	BodyLimitMB   int    `envconfig:"BODY_LIMIT_MB"`
	RedisClient   RedisClientConfig
	DB            DBConfig
	Kratos        KratosConfig
	MQ            MQConfig
	Quiz          QuizConfig
	SMTP          SMTPConfig
	Report        ReportConfig
	Email         EmailConfig
}

// LoadConfig loads and normalizes the configuration from environment variables.
func LoadConfig() (AppConfig, error) {
	// godotenv.Load ignores missing file
	_ = godotenv.Load()

	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		return cfg, err
	}

	// Fallback to PORT if APP_PORT is not set
	if cfg.Port == "" {
		portVal := os.Getenv("PORT")
		if portVal != "" {
			if _, err := strconv.Atoi(portVal); err == nil {
				cfg.Port = ":" + portVal
			} else {
				cfg.Port = portVal
			}
		}
	}

	// Normalize APP_ENV
	if cfg.Env == "" {
		cfg.Env = "development"
	}
	cfg.Env = strings.ToLower(cfg.Env)

	// Validate APP_ENV value
	switch cfg.Env {
	case "production", "development", "testing", "local":
		// valid
	default:
		return cfg, fmt.Errorf("invalid APP_ENV: %q", cfg.Env)
	}

	// Reject contradictory modes
	if cfg.Env == "production" && cfg.IsDevelopment {
		return cfg, fmt.Errorf("contradictory environment: IS_DEVELOPMENT cannot be true in production mode")
	}

	return cfg, nil
}

// GetConfig is a compatibility wrapper that loads the configuration and panics on error.
func GetConfig() AppConfig {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	AllConfig = cfg
	return AllConfig
}

// GetConfigByName collects a config variable directly from OS environment.
func GetConfigByName(key string) string {
	_ = godotenv.Load()
	return os.Getenv(key)
}

// LoadTestEnv loads environment variables from .env.testing file
func LoadTestEnv() AppConfig {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = godotenv.Load(fmt.Sprintf("%s/.env.testing", cwd))
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	AllConfig = cfg
	return AllConfig
}

// ValidateMigrationConfig validates configuration settings required specifically for running migrations.
func ValidateMigrationConfig(cfg AppConfig) error {
	if cfg.DB.Dialect == "" || isWhitespaceOnly(cfg.DB.Dialect) {
		return fmt.Errorf("DB_DIALECT is a required field")
	}

	switch cfg.DB.Dialect {
	case "sqlite3":
		if isWhitespaceOnly(cfg.DB.SQLiteFilePath) {
			return fmt.Errorf("SQLITE_FILEPATH is required when DB_DIALECT is sqlite3")
		}
	case "postgres", "mysql":
		var missing []string
		if isWhitespaceOnly(cfg.DB.Host) {
			missing = append(missing, "DB_HOST")
		}
		if cfg.DB.Port <= 0 || cfg.DB.Port > 65535 {
			missing = append(missing, "DB_PORT")
		}
		if isWhitespaceOnly(cfg.DB.Username) {
			missing = append(missing, "DB_USERNAME")
		}
		if isWhitespaceOnly(cfg.DB.Password) {
			missing = append(missing, "DB_PASSWORD")
		}
		if isWhitespaceOnly(cfg.DB.Db) {
			missing = append(missing, "DB_NAME")
		}
		if len(missing) > 0 {
			return fmt.Errorf("missing required database configuration variables: %v", missing)
		}
	default:
		return fmt.Errorf("unsupported DB_DIALECT: %q", cfg.DB.Dialect)
	}

	if isWhitespaceOnly(cfg.DB.MigrationDir) {
		return fmt.Errorf("MIGRATION_DIR is a required field")
	}

	return nil
}

// ValidateAPIConfig validates configuration settings required for running the API server.
func ValidateAPIConfig(cfg AppConfig) error {
	var missing []string

	// Validate DB settings
	if cfg.DB.Dialect == "" || isWhitespaceOnly(cfg.DB.Dialect) {
		missing = append(missing, "DB_DIALECT")
	} else if cfg.DB.Dialect == "postgres" || cfg.DB.Dialect == "mysql" {
		if isWhitespaceOnly(cfg.DB.Host) {
			missing = append(missing, "DB_HOST")
		}
		if cfg.DB.Port <= 0 || cfg.DB.Port > 65535 {
			missing = append(missing, "DB_PORT")
		}
		if isWhitespaceOnly(cfg.DB.Username) {
			missing = append(missing, "DB_USERNAME")
		}
		if isWhitespaceOnly(cfg.DB.Password) {
			missing = append(missing, "DB_PASSWORD")
		}
		if isWhitespaceOnly(cfg.DB.Db) {
			missing = append(missing, "DB_NAME")
		}
	} else if cfg.DB.Dialect == "sqlite3" {
		if isWhitespaceOnly(cfg.DB.SQLiteFilePath) {
			missing = append(missing, "SQLITE_FILEPATH")
		}
	}

	if isWhitespaceOnly(cfg.DB.MigrationDir) {
		missing = append(missing, "MIGRATION_DIR")
	}

	// Validate app port
	if isWhitespaceOnly(cfg.Port) {
		missing = append(missing, "APP_PORT")
	} else {
		if err := validatePortStr(cfg.Port); err != nil {
			return fmt.Errorf("invalid APP_PORT/PORT: %v", err)
		}
	}

	// Validate Redis port
	if isWhitespaceOnly(cfg.RedisClient.RedisPort) {
		missing = append(missing, "REDIS_PORT")
	} else {
		rp, err := strconv.Atoi(cfg.RedisClient.RedisPort)
		if err != nil || rp <= 0 || rp > 65535 {
			return fmt.Errorf("invalid REDIS_PORT value: %q", cfg.RedisClient.RedisPort)
		}
	}

	// In production (non-development, non-testing), check missing production settings
	if cfg.Env == "production" || (!cfg.IsDevelopment && cfg.Env != "testing") {
		if isWhitespaceOnly(cfg.RedisClient.RedisAddr) {
			missing = append(missing, "REDIS_HOST")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing mandatory production environment variables: %v", missing)
	}

	// Validate database pool limits
	if cfg.DB.MaxOpenConns < 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS cannot be negative: %d", cfg.DB.MaxOpenConns)
	}
	if cfg.DB.MaxIdleConns < 0 {
		return fmt.Errorf("DB_MAX_IDLE_CONNS cannot be negative: %d", cfg.DB.MaxIdleConns)
	}

	// Validate scheduler timeout range
	if cfg.Report.SchedulerTimeoutSeconds <= 0 || cfg.Report.SchedulerTimeoutSeconds > 300 {
		return fmt.Errorf("REPORT_SCHEDULER_TIMEOUT_SECONDS must be between 1 and 300: got %d", cfg.Report.SchedulerTimeoutSeconds)
	}

	return nil
}

func isWhitespaceOnly(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func validatePortStr(portStr string) error {
	port := portStr
	if strings.Contains(portStr, ":") {
		_, p, err := net.SplitHostPort(portStr)
		if err == nil {
			port = p
		} else {
			port = strings.TrimPrefix(portStr, ":")
		}
	}
	p, err := strconv.Atoi(port)
	if err != nil || p <= 0 || p > 65535 {
		return fmt.Errorf("invalid port number %q", portStr)
	}
	return nil
}
