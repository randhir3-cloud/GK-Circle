package cli

import (
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Init app initialization
func Init(cfg config.AppConfig, logger *zap.Logger) error {
	migrationCmd := GetMigrationCommandDef(cfg)
	apiCmd := GetAPICommandDef(cfg, logger)
	deleteOrphanedKratosUserCmd := GetDeleteOrphanedCommand(cfg)

	rootCmd := &cobra.Command{Use: "gk-circle"}
	rootCmd.AddCommand(&migrationCmd, &apiCmd, &deleteOrphanedKratosUserCmd)
	return rootCmd.Execute()
}
