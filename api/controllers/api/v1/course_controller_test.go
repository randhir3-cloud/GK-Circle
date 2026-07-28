package v1

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/middlewares"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

const (
	adminEmail    = "course-admin@example.com"
	nonAdminEmail = "learner@example.com"
	adminOwnerID  = "owner000000000000001"
)

type courseHTTPEnv struct {
	app    *fiber.App
	mock   sqlmock.Sqlmock
	cfg    config.AppConfig
	logger *zap.Logger
}

func newCourseHTTPEnv(t *testing.T, auth fiber.Handler) *courseHTTPEnv {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	logger := zap.NewNop()
	cfg := config.AppConfig{
		Quiz: config.QuizConfig{PublicQuizAdminEmails: []string{adminEmail}},
	}
	db := goqu.New("postgres", sqlDB)
	ctrl, err := InitCourseController(db, logger, &cfg)
	if err != nil {
		t.Fatalf("init controller: %v", err)
	}

	app := fiber.New()
	admin := app.Group("/api/v1/admin")
	admin.Use(auth)
	courses := admin.Group("/courses")
	courses.Post("/", ctrl.CreateCourse)
	courses.Get("/", ctrl.ListCourses)
	courses.Get("/:"+constants.CourseId, ctrl.GetCourse)
	courses.Patch("/:"+constants.CourseId, ctrl.UpdateCourse)

	return &courseHTTPEnv{app: app, mock: mock, cfg: cfg, logger: logger}
}

func injectAuthenticatedUser(user models.User) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, user)
		c.Locals(constants.ContextUid, user.ID)
		return c.Next()
	}
}

func kratosAuthMiddleware(t *testing.T) fiber.Handler {
	t.Helper()
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	mw := middlewares.NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))
	return mw.KratosAuthenticated
}

func adminUser() models.User {
	return models.User{ID: adminOwnerID, Email: adminEmail, FirstName: "Admin", LastName: "User", Username: "admin"}
}

func nonAdminUser() models.User {
	return models.User{ID: "owner000000000000099", Email: nonAdminEmail, FirstName: "Learner", LastName: "User", Username: "learner"}
}

func doJSON(t *testing.T, app *fiber.App, method, path string, body any) (*http.Response, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil && err != io.EOF {
		t.Fatalf("decode: %v", err)
	}
	return resp, payload
}

func TestAdminCourseRoutesUnauthenticated(t *testing.T) {
	env := newCourseHTTPEnv(t, kratosAuthMiddleware(t))
	resp, payload := doJSON(t, env.app, http.MethodPost, "/api/v1/admin/courses", map[string]any{"title": "Course"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; payload=%v", resp.StatusCode, payload)
	}
}

func TestAdminCourseRoutesNonAdminForbidden(t *testing.T) {
	env := newCourseHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	resp, payload := doJSON(t, env.app, http.MethodGet, "/api/v1/admin/courses", nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrCourseAdminForbidden {
		t.Fatalf("forbidden message = %v, want course administration wording", payload["data"])
	}
}

func TestAdminCourseCreateGetListPatch(t *testing.T) {
	env := newCourseHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	now := time.Date(2026, time.July, 26, 7, 22, 0, 0, time.UTC)

	env.mock.ExpectQuery(`INSERT INTO "courses".*RETURNING`).
		WithArgs(
			"Intermediate",
			sqlmock.AnyArg(),
			"English",
			adminOwnerID,
			"Short",
			"DRAFT",
			"PCS Foundation",
			"PRIVATE",
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "owner_id", "title", "short_description", "language",
			"difficulty", "visibility", "status", "created_at", "updated_at",
		}).AddRow(courseID, adminOwnerID, "PCS Foundation", "Short", "English", "Intermediate", "PRIVATE", "DRAFT", now, now))

	resp, payload := doJSON(t, env.app, http.MethodPost, "/api/v1/admin/courses", map[string]any{
		"title":             " PCS Foundation ",
		"short_description": "Short",
		"language":          "English",
		"difficulty":        "Intermediate",
		"visibility":        "PRIVATE",
		"owner_id":          "attacker-controlled",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d payload=%v", resp.StatusCode, payload)
	}
	data := payload["data"].(map[string]any)
	if data["owner_id"] != adminOwnerID {
		t.Fatalf("owner_id = %v, want authenticated owner (attacker body ignored)", data["owner_id"])
	}
	if data["status"] != "DRAFT" {
		t.Fatalf("status = %v", data["status"])
	}

	env.mock.ExpectQuery(`SELECT .* FROM "courses".*ORDER BY`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "owner_id", "title", "short_description", "language",
			"difficulty", "visibility", "status", "created_at", "updated_at",
		}).AddRow(courseID, adminOwnerID, "PCS Foundation", "Short", "English", "Intermediate", "PRIVATE", "DRAFT", now, now))
	resp, payload = doJSON(t, env.app, http.MethodGet, "/api/v1/admin/courses", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d", resp.StatusCode)
	}
	list := payload["data"].([]any)
	if len(list) != 1 {
		t.Fatalf("list len = %d", len(list))
	}

	env.mock.ExpectQuery(`SELECT .* FROM "courses"`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "owner_id", "title", "short_description", "language",
			"difficulty", "visibility", "status", "created_at", "updated_at",
		}).AddRow(courseID, adminOwnerID, "PCS Foundation", "Short", "English", "Intermediate", "PRIVATE", "DRAFT", now, now))
	resp, _ = doJSON(t, env.app, http.MethodGet, "/api/v1/admin/courses/"+courseID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d", resp.StatusCode)
	}

	env.mock.ExpectQuery(`UPDATE "courses".*RETURNING`).
		WithArgs("Renamed", courseID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "owner_id", "title", "short_description", "language",
			"difficulty", "visibility", "status", "created_at", "updated_at",
		}).AddRow(courseID, adminOwnerID, "Renamed", "Short", "English", "Intermediate", "PRIVATE", "DRAFT", now, now))
	resp, payload = doJSON(t, env.app, http.MethodPatch, "/api/v1/admin/courses/"+courseID.String(), map[string]any{
		"title": "Renamed",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("patch status = %d payload=%v", resp.StatusCode, payload)
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAdminCourseEmptyListNonNil(t *testing.T) {
	env := newCourseHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	env.mock.ExpectQuery(`SELECT .* FROM "courses".*ORDER BY`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "owner_id", "title", "short_description", "language",
			"difficulty", "visibility", "status", "created_at", "updated_at",
		}))
	resp, payload := doJSON(t, env.app, http.MethodGet, "/api/v1/admin/courses", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	list, ok := payload["data"].([]any)
	if !ok || list == nil {
		t.Fatalf("data = %#v, want non-nil array", payload["data"])
	}
}

func TestAdminCourseValidationAndNotFound(t *testing.T) {
	env := newCourseHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")

	resp, _ := doJSON(t, env.app, http.MethodPost, "/api/v1/admin/courses", map[string]any{"title": ""})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty title status = %d", resp.StatusCode)
	}

	resp, _ = doJSON(t, env.app, http.MethodPatch, "/api/v1/admin/courses/"+courseID.String(), map[string]any{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty patch status = %d", resp.StatusCode)
	}

	resp, _ = doJSON(t, env.app, http.MethodGet, "/api/v1/admin/courses/not-a-uuid", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad uuid status = %d", resp.StatusCode)
	}

	env.mock.ExpectQuery(`SELECT .* FROM "courses"`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "owner_id", "title", "short_description", "language",
			"difficulty", "visibility", "status", "created_at", "updated_at",
		}))
	resp, payload := doJSON(t, env.app, http.MethodGet, "/api/v1/admin/courses/"+courseID.String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing course status = %d payload=%v", resp.StatusCode, payload)
	}
}

func TestAdminCoursePatchNullClearsOptional(t *testing.T) {
	env := newCourseHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	now := time.Date(2026, time.July, 26, 7, 22, 0, 0, time.UTC)

	env.mock.ExpectQuery(`UPDATE "courses".*RETURNING`).
		WithArgs(nil, courseID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "owner_id", "title", "short_description", "language",
			"difficulty", "visibility", "status", "created_at", "updated_at",
		}).AddRow(courseID, adminOwnerID, "Title", nil, nil, nil, nil, "DRAFT", now, now))

	raw := []byte(`{"short_description":null}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/courses/"+courseID.String(), bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, body)
	}
}

func TestAdminCourseMissingIdentity(t *testing.T) {
	env := newCourseHTTPEnv(t, func(c *fiber.Ctx) error { return c.Next() })
	resp, payload := doJSON(t, env.app, http.MethodGet, "/api/v1/admin/courses", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
}

func TestAdminCourseRoutesDoNotExposeCourseNodePaths(t *testing.T) {
	env := newCourseHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/courses/"+uuid.New().String()+"/nodes", nil)
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		t.Fatal("CourseNode route must not be registered")
	}
}

func TestAdminCoursePublishStatus(t *testing.T) {
	env := newCourseHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	now := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)

	env.mock.ExpectQuery(`UPDATE "courses".*RETURNING`).
		WithArgs("PUBLISHED", courseID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "owner_id", "title", "short_description", "language",
			"difficulty", "visibility", "status", "created_at", "updated_at",
		}).AddRow(courseID, adminOwnerID, "PCS", nil, nil, nil, nil, "PUBLISHED", now, now))

	resp, payload := doJSON(t, env.app, http.MethodPatch, "/api/v1/admin/courses/"+courseID.String(), map[string]any{
		"status": "PUBLISHED",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	if data["status"] != "PUBLISHED" {
		t.Fatalf("data=%v", data)
	}

	resp, payload = doJSON(t, env.app, http.MethodPatch, "/api/v1/admin/courses/"+courseID.String(), map[string]any{
		"status": "LIVE",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid status=%d payload=%v", resp.StatusCode, payload)
	}
}
