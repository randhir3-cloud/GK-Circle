// GK Circle
//
//	Schemes: http
//	Host: 127.0.0.1
//	BasePath: /api
//	Version: 0.0.1-alpha
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package main

import (
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/randhir3-cloud/GK-Circle-v2/api/cli"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/logger"
	"github.com/randhir3-cloud/GK-Circle-v2/api/routinewrapper"
	sentry "github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

func run() (exitCode int) {
	exitCode = 1

	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("application panic: %v\n%s", recovered, debug.Stack())

			// Check if global logger is set and use it if possible
			rootLog := zap.L()
			if rootLog != nil {
				rootLog.Error(
					"application panicked",
					zap.Any("panic", recovered),
					zap.ByteString("stack", debug.Stack()),
				)
				_ = rootLog.Sync()
			}

			sentry.CurrentHub().Recover(recovered)
			sentry.Flush(2 * time.Second)

			exitCode = 1
		}
	}()

	// 1. Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("configuration error: %v", err)
		return 1
	}
	log.Printf("configuration loaded successfully")

	// 2. Init root logger
	rootLogger, err := logger.NewRootLogger(cfg.Debug, cfg.IsDevelopment)
	if err != nil {
		log.Printf("logger initialization failed: %v", err)
		return 1
	}

	zap.ReplaceGlobals(rootLogger)
	defer func() {
		_ = rootLogger.Sync()
	}()
	rootLogger.Info("logger initialized successfully")

	// 3. Setup Sentry and Panic/Routinewrapper
	sentryLoggedFunc := func() {
		recovered := recover()
		if recovered != nil {
			rootLogger.Error("worker routine panicked", zap.Any("panic", recovered), zap.ByteString("stack", debug.Stack()))
			sentry.CurrentHub().Recover(recovered)
			sentry.Flush(2 * time.Second)
		}
	}
	routinewrapper.Init(sentryLoggedFunc)

	// 4. Run CLI
	if err := cli.Init(cfg, rootLogger); err != nil {
		rootLogger.Error("application startup failed", zap.Error(err))
		sentry.CaptureException(err)
		sentry.Flush(2 * time.Second)
		return 1
	}

	exitCode = 0
	return
}

func main() {
	os.Exit(run())
}
