package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/database"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/kratos"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSeedAdminCommand_Execution(t *testing.T) {
	logger := zap.NewNop()

	// 1. Mock Kratos Admin endpoints
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		identities := []kratos.KratosIdentity{
			{
				ID:       "kratos-id-123",
				SchemaID: "default",
				State:    "active",
			},
		}
		identities[0].Traits.Email = services.TargetSuperAdminEmail
		identities[0].Traits.Name.First = "Randhir"
		identities[0].Traits.Name.Last = "Sandhu"
		identities[0].VerifiableAddresses = []kratos.KratosVerifiableAddress{
			{
				Value:    services.TargetSuperAdminEmail,
				Verified: true,
				Status:   "completed",
			},
		}

		data, _ := json.Marshal(identities)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer server.Close()

	// 2. Setup mock DB
	db, mockSql, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockSql.ExpectBegin()
	mockSql.ExpectQuery("users").
		WillReturnRows(sqlmock.NewRows([]string{"id", "kratos_id", "email", "roles"}).AddRow("db-id-123", "kratos-id-123", services.TargetSuperAdminEmail, "user"))

	mockSql.ExpectExec("users").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockSql.ExpectCommit()

	// Inject the mock DB into database package
	database.SetTestingDB(db)
	defer database.SetTestingDB(nil)

	// 3. Configure app config
	cfg := config.AppConfig{}
	cfg.Kratos.AdminUrl = server.URL
	cfg.DB.Dialect = "postgres"

	// 4. Initialize and run command
	cmd := GetSeedAdminCommand(cfg, logger)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--email", services.TargetSuperAdminEmail})

	err = cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
}

func TestSeedAdminCommand_RestrictedEmail(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.AppConfig{}
	cfg.Kratos.AdminUrl = "http://localhost:4434"
	cfg.DB.Dialect = "postgres"

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	database.SetTestingDB(db)
	defer database.SetTestingDB(nil)

	cmd := GetSeedAdminCommand(cfg, logger)
	cmd.SetArgs([]string{"--email", "attacker@example.com"})

	err = cmd.ExecuteContext(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "restricted")
}
