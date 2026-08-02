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
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUserController_GetUserMeta_Compatibility(t *testing.T) {
	logger := zap.NewNop()

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	goquDB := goqu.New("postgres", db)

	cfg := config.AppConfig{}
	ctrl, err := NewUserController(goquDB, logger, cfg)
	assert.NoError(t, err)

	app := fiber.New()
	app.Get("/who", func(c *fiber.Ctx) error {
		// Mock ContextUser and UID Context
		c.Locals(constants.ContextUser, models.User{
			ID:        "user-123",
			Email:     "randhirsandhu81@gmail.com",
			Username:  "randhirsandhu",
			FirstName: "Randhir",
			Roles:     "super_admin",
		})
		c.Locals(constants.ContextUid, "user-123")
		c.Locals(constants.KratosID, "kratos-id-123")
		return ctrl.GetUserMeta(c)
	})

	req := httptest.NewRequest("GET", "/who", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	data := result["data"].(map[string]any)
	assert.Equal(t, "admin-user", data["role"])
	assert.Equal(t, true, data["can_create_public_quiz"])

	roles := data["roles"].([]any)
	assert.Len(t, roles, 1)
	assert.Equal(t, "super_admin", roles[0])
}
