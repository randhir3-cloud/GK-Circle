package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	goqu "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql" // import mysql if it is used
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

var db *sql.DB
var dbURL string
var err error

const (
	POSTGRES = "postgres"
	MYSQL    = "mysql"
	SQLITE3  = "sqlite3"
)

// Connect with database
func Connect(cfg config.DBConfig) (*goqu.Database, error) {
	start := time.Now()
	zap.L().Info("database connect start",
		zap.String("dialect", cfg.Dialect),
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Db),
	)

	var goquDB *goqu.Database
	var err error
	switch cfg.Dialect {
	case POSTGRES:
		goquDB, err = postgresDBConnection(cfg)
	case MYSQL:
		goquDB, err = mysqlDBConnection(cfg)
	case SQLITE3:
		goquDB, err = sqlite3DBConnection(cfg)
	default:
		err = errors.New("no suitable dialect found")
	}

	elapsed := time.Since(start)

	if err != nil {
		zap.L().Error("database connect failure",
			zap.String("dialect", cfg.Dialect),
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.Db),
			zap.Duration("elapsed", elapsed),
			zap.Error(err),
		)
		return nil, err
	}

	// Bounded startup check
	var rawDB *sql.DB
	if goquDB != nil {
		if underlyingDB, ok := goquDB.Db.(*sql.DB); ok {
			rawDB = underlyingDB
		}
	}

	if rawDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if pingErr := rawDB.PingContext(ctx); pingErr != nil {
			_ = rawDB.Close()
			zap.L().Error("database ping failure",
				zap.String("dialect", cfg.Dialect),
				zap.String("host", cfg.Host),
				zap.Int("port", cfg.Port),
				zap.String("database", cfg.Db),
				zap.Duration("elapsed", time.Since(start)),
				zap.Error(pingErr),
			)
			return nil, fmt.Errorf("database startup check failed: %w", pingErr)
		}
	}

	zap.L().Info("database connect success",
		zap.String("dialect", cfg.Dialect),
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Db),
		zap.Duration("elapsed", time.Since(start)),
	)

	return goquDB, nil
}

func sqlite3DBConnection(cfg config.DBConfig) (*goqu.Database, error) {

	if _, err = os.Stat(cfg.SQLiteFilePath); err != nil {
		file, err := os.Create(cfg.SQLiteFilePath)
		if err != nil {
			panic(err)
		}
		err = file.Close()
		if err != nil {
			return nil, err
		}
	}
	db, err = sql.Open(SQLITE3, "./"+cfg.SQLiteFilePath)
	if err != nil {
		return nil, err
	}
	return goqu.New(SQLITE3, db), err
}

func mysqlDBConnection(cfg config.DBConfig) (*goqu.Database, error) {
	dbURL = cfg.Username + ":" + cfg.Password + "@tcp(" + cfg.Host + ":" + strconv.Itoa(cfg.Port) + ")/" + cfg.Db
	if db == nil {
		db, err = sql.Open(MYSQL, dbURL)
		if err != nil {
			return nil, err
		}
		return goqu.New(MYSQL, db), err
	}
	return goqu.New(MYSQL, db), err
}

func postgresDBConnection(cfg config.DBConfig) (*goqu.Database, error) {
	dbURL = "postgres://" + cfg.Username + ":" + cfg.Password + "@" + cfg.Host + ":" + strconv.Itoa(cfg.Port) + "/" + cfg.Db + "?" + cfg.QueryString
	if db == nil {
		db, err = sql.Open(POSTGRES, dbURL)
		if err != nil {
			return nil, err
		}
		
		// Configure pool limits
		maxOpen := cfg.MaxOpenConns
		if maxOpen <= 0 {
			maxOpen = 20 // default open
		}
		db.SetMaxOpenConns(maxOpen)

		maxIdle := cfg.MaxIdleConns
		if maxIdle <= 0 {
			maxIdle = 10 // default idle
		}
		db.SetMaxIdleConns(maxIdle)

		if cfg.ConnMaxLifetime != "" {
			if d, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
				db.SetConnMaxLifetime(d)
			}
		} else {
			db.SetConnMaxLifetime(30 * time.Minute)
		}

		if cfg.ConnMaxIdleTime != "" {
			if d, err := time.ParseDuration(cfg.ConnMaxIdleTime); err == nil {
				db.SetConnMaxIdleTime(d)
			}
		} else {
			db.SetConnMaxIdleTime(15 * time.Minute)
		}

		return goqu.New(POSTGRES, db), err
	}
	return goqu.New(POSTGRES, db), err
}
