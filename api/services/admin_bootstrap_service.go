package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/kratos"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

const TargetSuperAdminEmail = "randhirsandhu81@gmail.com"

type AdminBootstrapService struct {
	db           *goqu.Database
	userModel    *models.UserModel
	kratosClient kratos.KratosAdminClient
	logger       *zap.Logger
}

func NewAdminBootstrapService(db *goqu.Database, userModel *models.UserModel, kratosClient kratos.KratosAdminClient, logger *zap.Logger) *AdminBootstrapService {
	return &AdminBootstrapService{
		db:           db,
		userModel:    userModel,
		kratosClient: kratosClient,
		logger:       logger,
	}
}

func (s *AdminBootstrapService) BootstrapAdmin(ctx context.Context, email string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail != TargetSuperAdminEmail {
		return fmt.Errorf("seed-admin is restricted to the configured super-admin identity %q", TargetSuperAdminEmail)
	}

	identities, err := s.kratosClient.FindIdentitiesByEmail(ctx, normalizedEmail)
	if err != nil {
		return fmt.Errorf("find Kratos identity: %w", err)
	}

	if len(identities) == 0 {
		return fmt.Errorf("Kratos identity %q does not exist", normalizedEmail)
	}

	if len(identities) > 1 {
		return fmt.Errorf("multiple identities match email %q", normalizedEmail)
	}

	identity := identities[0]

	isAlreadyVerified := false
	for _, addr := range identity.VerifiableAddresses {
		if strings.EqualFold(strings.TrimSpace(addr.Value), normalizedEmail) && addr.Verified {
			isAlreadyVerified = true
			break
		}
	}

	var verifiedIdentity *kratos.KratosIdentity
	if !isAlreadyVerified {
		s.logger.Info("Marking Kratos identity email as verified", zap.String("email", normalizedEmail))
		verifiedIdentity, err = s.kratosClient.VerifyEmailAddress(ctx, identity.ID, normalizedEmail)
		if err != nil {
			return fmt.Errorf("verify Kratos email address: %w", err)
		}
	} else {
		s.logger.Info("Kratos identity email is already verified", zap.String("email", normalizedEmail))
		verifiedIdentity = &identity
	}

	hasVerified := false
	for _, addr := range verifiedIdentity.VerifiableAddresses {
		if strings.EqualFold(strings.TrimSpace(addr.Value), normalizedEmail) && addr.Verified {
			hasVerified = true
			break
		}
	}
	if !hasVerified {
		return fmt.Errorf("Kratos did not persist verified status for %q", normalizedEmail)
	}

	err = s.db.WithTx(func(tx *goqu.TxDatabase) error {
		var dbUser models.User
		found, err := tx.From(models.UserTable).Where(goqu.Ex{"kratos_id": verifiedIdentity.ID}).ScanStruct(&dbUser)
		if err != nil {
			return fmt.Errorf("failed to query database for user: %w", err)
		}

		if !found {
			foundByEmail, err := tx.From(models.UserTable).Where(goqu.Ex{"email": normalizedEmail}).ScanStruct(&dbUser)
			if err != nil {
				return fmt.Errorf("failed to query database for user by email: %w", err)
			}

			if foundByEmail {
				newRoles := models.AddRole(dbUser.Roles, models.SystemRoleSuperAdmin)
				firstName := dbUser.FirstName
				lastName := dbUser.LastName
				if strings.TrimSpace(firstName) == "" {
					firstName = verifiedIdentity.Traits.Name.First
				}
				if strings.TrimSpace(lastName) == "" {
					lastName = verifiedIdentity.Traits.Name.Last
				}

				_, err = tx.Update(models.UserTable).
					Set(goqu.Record{
						"kratos_id":  verifiedIdentity.ID,
						"roles":      newRoles,
						"first_name": firstName,
						"last_name":  lastName,
						"updated_at": goqu.L("NOW()"),
					}).
					Where(goqu.Ex{"id": dbUser.ID}).
					Executor().Exec()
				if err != nil {
					return fmt.Errorf("failed to promote existing user by email: %w", err)
				}
				s.logger.Info("Promoted existing application user by email to super_admin", zap.String("email", normalizedEmail))
			} else {
				firstName := verifiedIdentity.Traits.Name.First
				lastName := verifiedIdentity.Traits.Name.Last
				if strings.TrimSpace(firstName) == "" {
					firstName = "Randhir"
				}
				if strings.TrimSpace(lastName) == "" {
					lastName = "Sandhu"
				}
				username := firstName
				if strings.TrimSpace(username) == "" {
					username = "superadmin"
				}

				newUserID := xid.New().String()
				_, err = tx.Insert(models.UserTable).Rows(
					goqu.Record{
						"id":         newUserID,
						"kratos_id":  verifiedIdentity.ID,
						"first_name": firstName,
						"last_name":  lastName,
						"email":      normalizedEmail,
						"username":   username,
						"roles":      models.SystemRoleSuperAdmin,
					},
				).Executor().Exec()
				if err != nil {
					return fmt.Errorf("failed to insert new super_admin user: %w", err)
				}
				s.logger.Info("Created and promoted new application user to super_admin", zap.String("email", normalizedEmail))
			}
		} else {
			newRoles := models.AddRole(dbUser.Roles, models.SystemRoleSuperAdmin)
			firstName := dbUser.FirstName
			lastName := dbUser.LastName
			if strings.TrimSpace(firstName) == "" {
				firstName = verifiedIdentity.Traits.Name.First
			}
			if strings.TrimSpace(lastName) == "" {
				lastName = verifiedIdentity.Traits.Name.Last
			}

			_, err = tx.Update(models.UserTable).
				Set(goqu.Record{
					"roles":      newRoles,
					"first_name": firstName,
					"last_name":  lastName,
					"updated_at": goqu.L("NOW()"),
				}).
				Where(goqu.Ex{"id": dbUser.ID}).
				Executor().Exec()
			if err != nil {
				return fmt.Errorf("failed to promote existing user: %w", err)
			}
			s.logger.Info("Updated existing application user roles to super_admin", zap.String("email", normalizedEmail))
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("database transaction failed: %w", err)
	}

	return nil
}
