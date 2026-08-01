package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/database"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var errKratosIdentityNotFound = errors.New("kratos identity not found")
var verificationRunIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{5,63}$`)

type qaKratosIdentity struct {
	ID     string `json:"id"`
	Traits struct {
		Email string `json:"email"`
	} `json:"traits"`
}

type qaKratosAdmin interface {
	GetIdentity(context.Context, string) (qaKratosIdentity, error)
	DeleteIdentity(context.Context, string) error
}

type qaUserStore interface {
	GetUserByKratosID(string) (models.User, error)
}

type qaUserDeleter interface {
	DeleteUserDataById(string, string) error
}

type verificationCleanupRunner struct {
	admin   qaKratosAdmin
	users   qaUserStore
	deleter qaUserDeleter
}

func (r verificationCleanupRunner) Run(ctx context.Context, identityID, runID string, confirm bool) error {
	if _, err := uuid.Parse(identityID); err != nil {
		return fmt.Errorf("identity ID must be a valid UUID")
	}
	if !verificationRunIDPattern.MatchString(runID) {
		return fmt.Errorf("run ID must be 6-64 safe characters")
	}

	user, userErr := r.users.GetUserByKratosID(identityID)
	if userErr != nil && !errors.Is(userErr, sql.ErrNoRows) {
		return fmt.Errorf("application identity lookup failed")
	}

	identity, identityErr := r.admin.GetIdentity(ctx, identityID)
	if identityErr != nil && !errors.Is(identityErr, errKratosIdentityNotFound) {
		return fmt.Errorf("Kratos identity lookup failed")
	}
	if errors.Is(userErr, sql.ErrNoRows) && errors.Is(identityErr, errKratosIdentityNotFound) {
		return nil
	}

	email := identity.Traits.Email
	if errors.Is(identityErr, errKratosIdentityNotFound) {
		email = user.Email
	}
	if !emailContainsExactRunTag(email, runID) {
		return fmt.Errorf("identity run tag does not match")
	}
	if userErr == nil && !emailContainsExactRunTag(user.Email, runID) {
		return fmt.Errorf("application identity run tag does not match")
	}

	if !confirm {
		return nil
	}
	if userErr == nil {
		return r.deleter.DeleteUserDataById(user.ID, identityID)
	}
	return r.admin.DeleteIdentity(ctx, identityID)
}

func emailContainsExactRunTag(email, runID string) bool {
	separator := strings.LastIndex(email, "@")
	return separator > 0 && strings.HasSuffix(email[:separator], "+"+runID)
}

func validateQAAdminURL(rawURL, environment string) (*url.URL, error) {
	adminURL, err := url.ParseRequestURI(rawURL)
	if err != nil || adminURL.Host == "" || (adminURL.Scheme != "http" && adminURL.Scheme != "https") {
		return nil, fmt.Errorf("SERVE_ADMIN_BASE_URL is invalid")
	}

	hostname := strings.ToLower(adminURL.Hostname())
	if strings.EqualFold(environment, "production") &&
		(hostname == "gkcircle.com" ||
			strings.HasSuffix(hostname, ".gkcircle.com") ||
			strings.HasSuffix(hostname, ".up.railway.app")) {
		return nil, fmt.Errorf("SERVE_ADMIN_BASE_URL must use the private Kratos Admin API")
	}
	return adminURL, nil
}

type httpKratosAdmin struct {
	baseURL string
	client  *http.Client
}

func (a httpKratosAdmin) GetIdentity(ctx context.Context, identityID string) (qaKratosIdentity, error) {
	var identity qaKratosIdentity
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/identities/"+url.PathEscape(identityID), nil)
	if err != nil {
		return identity, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return identity, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return identity, errKratosIdentityNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return identity, fmt.Errorf("unexpected Kratos status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return identity, err
	}
	return identity, nil
}

func (a httpKratosAdmin) DeleteIdentity(ctx context.Context, identityID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, a.baseURL+"/identities/"+url.PathEscape(identityID), nil)
	if err != nil {
		return err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected Kratos status %d", resp.StatusCode)
	}
	return nil
}

func GetQACommand(cfg config.AppConfig, logger *zap.Logger) cobra.Command {
	qaCommand := cobra.Command{Use: "qa", Short: "Internal QA maintenance commands"}
	var identityID, runID string
	var dryRun, confirm bool

	cleanupCommand := cobra.Command{
		Use:   "cleanup-verification",
		Short: "Safely clean a run-tagged verification identity",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if dryRun == confirm {
				return fmt.Errorf("select exactly one of --dry-run or --confirm")
			}
			if strings.TrimSpace(cfg.Kratos.AdminUrl) == "" {
				return fmt.Errorf("SERVE_ADMIN_BASE_URL must be set")
			}
			adminURL, err := validateQAAdminURL(cfg.Kratos.AdminUrl, cfg.Env)
			if err != nil {
				return err
			}

			db, err := database.Connect(cfg.DB)
			if err != nil {
				return err
			}
			userModel, err := models.InitUserModel(db, logger)
			if err != nil {
				return err
			}
			userService, err := services.NewUserService(db, logger, cfg)
			if err != nil {
				return err
			}
			runner := verificationCleanupRunner{
				admin: httpKratosAdmin{
					baseURL: strings.TrimRight(adminURL.String(), "/"),
					client:  &http.Client{Timeout: 15 * time.Second},
				},
				users:   &userModel,
				deleter: userService,
			}
			if err := runner.Run(cmd.Context(), identityID, runID, confirm); err != nil {
				return err
			}
			logger.Info("verification QA cleanup completed",
				zap.String("run_id", runID),
				zap.String("identity_prefix", identityID[:8]),
				zap.Bool("dry_run", dryRun),
			)
			return nil
		},
	}
	cleanupCommand.Flags().StringVar(&identityID, "identity-id", "", "Kratos identity UUID")
	cleanupCommand.Flags().StringVar(&runID, "run-id", "", "unique QA run ID")
	cleanupCommand.Flags().BoolVar(&dryRun, "dry-run", false, "validate without deletion")
	cleanupCommand.Flags().BoolVar(&confirm, "confirm", false, "perform deletion")
	_ = cleanupCommand.MarkFlagRequired("identity-id")
	_ = cleanupCommand.MarkFlagRequired("run-id")
	qaCommand.AddCommand(&cleanupCommand)
	return qaCommand
}
