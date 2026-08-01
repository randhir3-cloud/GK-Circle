package cli

import (
	"fmt"
	"strings"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/database"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/kratos"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func GetSeedAdminCommand(cfg config.AppConfig, logger *zap.Logger) cobra.Command {
	var email string

	cmd := cobra.Command{
		Use:   "seed-admin",
		Short: "Seed and promote the super admin user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(cfg.Kratos.AdminUrl) == "" {
				return fmt.Errorf("SERVE_ADMIN_BASE_URL must be set")
			}

			db, err := database.Connect(cfg.DB)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}

			userModel, err := models.InitUserModel(db, logger)
			if err != nil {
				return fmt.Errorf("failed to initialize user model: %w", err)
			}

			kratosClient := kratos.NewKratosAdminClient(cfg.Kratos.AdminUrl)
			bootstrapService := services.NewAdminBootstrapService(db, &userModel, kratosClient, logger)

			logger.Info("Executing seed-admin command", zap.String("email", email))
			if err := bootstrapService.BootstrapAdmin(cmd.Context(), email); err != nil {
				return err
			}

			logger.Info("Successfully completed seed-admin command", zap.String("email", email))
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", services.TargetSuperAdminEmail, "email address of the super admin to seed/promote")
	return cmd
}
