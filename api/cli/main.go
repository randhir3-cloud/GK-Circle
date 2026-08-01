package cli

import (
	"fmt"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Init app initialization
func Init(cfg config.AppConfig, logger *zap.Logger) error {
	migrationCmd := GetMigrationCommandDef(cfg)
	apiCmd := GetAPICommandDef(cfg, logger)
	deleteOrphanedKratosUserCmd := GetDeleteOrphanedCommand(cfg)
	qaCmd := GetQACommand(cfg, logger)

	rootCmd := &cobra.Command{
		Use:           "gk-circle",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("no command specified: expected one of api, migrate, or another supported command")
		},
	}

	rootCmd.AddCommand(&migrationCmd, &apiCmd, &deleteOrphanedKratosUserCmd, &qaCmd)
	return rootCmd.Execute()
}
