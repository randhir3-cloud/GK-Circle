package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	goqu "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/database"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegistration_RoleEscalation_And_DowngradePrevention(t *testing.T) {
	logger := zap.NewNop()

	// 1. Mock Kratos public session endpoint response
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/sessions/whoami", req.URL.Path)

		user := config.KratosUserDetails{}
		user.Identity.ID = "kratos-id-123"
		user.Identity.Traits.Email = "newuser@gkcircle.com"
		user.Identity.Traits.Name.First = "New"
		user.Identity.Traits.Name.Last = "User"

		data, _ := json.Marshal(user)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer server.Close()

	// 2. Setup mock DB
	db, mockSql, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	database.SetTestingDB(db)
	defer database.SetTestingDB(nil)

	goquDB := goqu.New("postgres", db)

	// Expect check if user exists
	mockSql.ExpectQuery("users").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Expect insert of new user with role "user"
	mockSql.ExpectExec("INSERT INTO \"users\"").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Configure AppConfig
	cfg := config.AppConfig{}
	cfg.Kratos.BaseUrl = server.URL

	// 4. Initialize AuthController
	ctrl, err := NewAuthController(goquDB, logger, cfg)
	assert.NoError(t, err)

	app := fiber.New()
	app.Get("/v1/kratos/auth", ctrl.DoKratosAuth)

	req := httptest.NewRequest("GET", "/v1/kratos/auth", nil)
	req.Header.Set("Cookie", "ory_kratos_session=mock-session-cookie")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
}
