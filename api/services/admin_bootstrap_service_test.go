package services

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/kratos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockKratosAdminClient struct {
	mock.Mock
}

func (m *mockKratosAdminClient) FindIdentitiesByEmail(ctx context.Context, email string) ([]kratos.KratosIdentity, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]kratos.KratosIdentity), args.Error(1)
}

func (m *mockKratosAdminClient) GetIdentity(ctx context.Context, identityID string) (*kratos.KratosIdentity, error) {
	args := m.Called(ctx, identityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kratos.KratosIdentity), args.Error(1)
}

func (m *mockKratosAdminClient) VerifyEmailAddress(ctx context.Context, identityID string, email string) (*kratos.KratosIdentity, error) {
	args := m.Called(ctx, identityID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kratos.KratosIdentity), args.Error(1)
}

func TestAdminBootstrapService_BootstrapAdmin_RestrictedEmail(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAdminBootstrapService(nil, nil, nil, logger)

	err := svc.BootstrapAdmin(context.Background(), "attacker@example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "restricted")
}

func TestAdminBootstrapService_BootstrapAdmin_NoIdentity(t *testing.T) {
	logger := zap.NewNop()
	mockKratos := new(mockKratosAdminClient)
	mockKratos.On("FindIdentitiesByEmail", mock.Anything, TargetSuperAdminEmail).Return([]kratos.KratosIdentity{}, nil)

	svc := NewAdminBootstrapService(nil, nil, mockKratos, logger)
	err := svc.BootstrapAdmin(context.Background(), TargetSuperAdminEmail)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestAdminBootstrapService_BootstrapAdmin_MultipleIdentities(t *testing.T) {
	logger := zap.NewNop()
	mockKratos := new(mockKratosAdminClient)
	mockKratos.On("FindIdentitiesByEmail", mock.Anything, TargetSuperAdminEmail).Return([]kratos.KratosIdentity{
		{ID: "id-1"},
		{ID: "id-2"},
	}, nil)

	svc := NewAdminBootstrapService(nil, nil, mockKratos, logger)
	err := svc.BootstrapAdmin(context.Background(), TargetSuperAdminEmail)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple identities")
}

func TestAdminBootstrapService_BootstrapAdmin_SuccessfulCreate(t *testing.T) {
	logger := zap.NewNop()
	mockKratos := new(mockKratosAdminClient)

	identity := kratos.KratosIdentity{
		ID:       "kratos-id-123",
		SchemaID: "default",
		State:    "active",
	}
	identity.Traits.Email = TargetSuperAdminEmail
	identity.Traits.Name.First = "Randhir"
	identity.Traits.Name.Last = "Sandhu"

	mockKratos.On("FindIdentitiesByEmail", mock.Anything, TargetSuperAdminEmail).Return([]kratos.KratosIdentity{identity}, nil)

	verifiedIdentity := identity
	verifiedIdentity.VerifiableAddresses = []kratos.KratosVerifiableAddress{
		{
			Value:    TargetSuperAdminEmail,
			Verified: true,
			Status:   "completed",
		},
	}

	mockKratos.On("VerifyEmailAddress", mock.Anything, "kratos-id-123", TargetSuperAdminEmail).Return(&verifiedIdentity, nil)

	db, mockSql, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockSql.ExpectBegin()
	mockSql.ExpectQuery("users").
		WillReturnRows(sqlmock.NewRows([]string{"id", "kratos_id", "email", "roles"}))

	mockSql.ExpectQuery("users").
		WillReturnRows(sqlmock.NewRows([]string{"id", "kratos_id", "email", "roles"}))

	mockSql.ExpectExec("users").WillReturnResult(sqlmock.NewResult(1, 1))
	mockSql.ExpectCommit()

	goquDB := goqu.New("postgres", db)
	userModel, err := models.InitUserModel(goquDB, logger)
	assert.NoError(t, err)

	svc := NewAdminBootstrapService(goquDB, &userModel, mockKratos, logger)
	err = svc.BootstrapAdmin(context.Background(), TargetSuperAdminEmail)
	assert.NoError(t, err)
}

func TestAdminBootstrapService_BootstrapAdmin_SuccessfulUpdate(t *testing.T) {
	logger := zap.NewNop()
	mockKratos := new(mockKratosAdminClient)

	identity := kratos.KratosIdentity{
		ID:       "kratos-id-123",
		SchemaID: "default",
		State:    "active",
	}
	identity.Traits.Email = TargetSuperAdminEmail
	identity.Traits.Name.First = "Randhir"
	identity.Traits.Name.Last = "Sandhu"
	identity.VerifiableAddresses = []kratos.KratosVerifiableAddress{
		{
			Value:    TargetSuperAdminEmail,
			Verified: true,
			Status:   "completed",
		},
	}

	mockKratos.On("FindIdentitiesByEmail", mock.Anything, TargetSuperAdminEmail).Return([]kratos.KratosIdentity{identity}, nil)

	db, mockSql, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockSql.ExpectBegin()
	mockSql.ExpectQuery("users").
		WillReturnRows(sqlmock.NewRows([]string{"id", "kratos_id", "email", "roles"}).AddRow("db-id-123", "kratos-id-123", TargetSuperAdminEmail, "user"))

	mockSql.ExpectExec("users").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockSql.ExpectCommit()

	goquDB := goqu.New("postgres", db)
	userModel, err := models.InitUserModel(goquDB, logger)
	assert.NoError(t, err)

	svc := NewAdminBootstrapService(goquDB, &userModel, mockKratos, logger)
	err = svc.BootstrapAdmin(context.Background(), TargetSuperAdminEmail)
	assert.NoError(t, err)
}
