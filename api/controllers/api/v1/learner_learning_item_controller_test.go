package v1

import (
	"net/http"
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
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

type learnerLearningItemHTTPEnv struct {
	app  *fiber.App
	mock sqlmock.Sqlmock
}

func newLearnerLearningItemHTTPEnv(t *testing.T, auth fiber.Handler) *learnerLearningItemHTTPEnv {
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
	ctrl, err := InitLearnerLearningItemController(db, logger, &cfg)
	if err != nil {
		t.Fatalf("init learner learning item controller: %v", err)
	}

	app := fiber.New()
	learner := app.Group("/api/v1/learner")
	learner.Use(auth)
	courses := learner.Group("/courses")
	nodes := courses.Group("/:" + constants.CourseId + "/nodes")
	items := nodes.Group("/:" + constants.NodeId + "/learning-items")
	items.Get("/", ctrl.List)
	items.Get("/:"+constants.ItemId, ctrl.GetByID)

	return &learnerLearningItemHTTPEnv{app: app, mock: mock}
}

func learnerLearningItemsBase(courseID, nodeID uuid.UUID) string {
	return "/api/v1/learner/courses/" + courseID.String() + "/nodes/" + nodeID.String() + "/learning-items"
}

func expectLearnerCourseEnrollment(mock sqlmock.Sqlmock, userID string, courseID uuid.UUID, enrolled bool) {
	query := mock.ExpectQuery(`SELECT "id" FROM "course_enrollments"`).
		WithArgs(courseID, userID, uint(1))
	if enrolled {
		query.WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		return
	}
	query.WillReturnRows(sqlmock.NewRows([]string{"id"}))
}

func TestLearnerLearningItemRoutesUnauthenticated(t *testing.T) {
	env := newLearnerLearningItemHTTPEnv(t, kratosAuthMiddleware(t))
	courseID := uuid.New()
	nodeID := uuid.New()
	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID), nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
}

func TestLearnerLearningItemRoutesAllowNonAdmin(t *testing.T) {
	env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	metadata := []byte(`{"version":1,"blocks":[]}`)

	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("non-admin learner status = %d payload=%v", resp.StatusCode, payload)
	}
	list, ok := payload["data"].([]any)
	if !ok || len(list) != 0 {
		t.Fatalf("list = %v", payload["data"])
	}
	_ = metadata
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL: %v", err)
	}
}

func TestLearnerLearningItemRoutesDenyUnenrolled(t *testing.T) {
	env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000031")

	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, false)
	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unenrolled list status = %d payload=%v", resp.StatusCode, payload)
	}
	if msg, _ := payload["data"].(string); msg != constants.ErrCourseEnrollmentRequired {
		t.Fatalf("unenrolled list message = %v", payload["data"])
	}

	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, false)
	resp, payload = doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID)+"/"+itemID.String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unenrolled get status = %d payload=%v", resp.StatusCode, payload)
	}
	if msg, _ := payload["data"].(string); msg != constants.ErrCourseEnrollmentRequired {
		t.Fatalf("unenrolled get message = %v", payload["data"])
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL: %v", err)
	}
}

func TestLearnerLearningItemListPublishedOnly(t *testing.T) {
	env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemA := uuid.MustParse("019c02a0-1111-7000-8000-000000000021")
	itemB := uuid.MustParse("019c02a0-1111-7000-8000-000000000022")
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)

	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
	rows := learningItemRowWithPublishState(
		itemA, courseID, nodeID, "First", models.LearningItemTypeArticle,
		nil, metadata, 0, now, models.LearningItemPublishStatePublished,
	)
	rows.AddRow(
		itemB, courseID, nodeID, "Second", models.LearningItemTypeVideo,
		nil, metadata, 1, models.LearningItemPublishStatePublished, now, now,
	)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished).
		WillReturnRows(rows)

	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d payload=%v", resp.StatusCode, payload)
	}
	assertLearningItemBackendSuiteJSend(t, payload, "success")
	assertLearningItemBackendSuiteKeys(t, payload, "status", "data")
	list, ok := payload["data"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("list = %v", payload["data"])
	}
	first, _ := list[0].(map[string]any)
	second, _ := list[1].(map[string]any)
	learnerKeys := []string{"id", "title", "item_type", "description", "metadata", "publish_state"}
	assertLearningItemBackendSuiteKeys(t, first, learnerKeys...)
	assertLearningItemBackendSuiteKeys(t, second, learnerKeys...)
	if first["title"] != "First" || second["title"] != "Second" {
		t.Fatalf("order = %v", payload["data"])
	}
	if first["publish_state"] != string(models.LearningItemPublishStatePublished) {
		t.Fatalf("publish_state = %v", first["publish_state"])
	}
	if _, hasPosition := first["position"]; hasPosition {
		t.Fatalf("learner response should omit position: %v", first)
	}
	if _, hasCourseID := first["course_id"]; hasCourseID {
		t.Fatalf("learner response should omit course_id: %v", first)
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL: %v", err)
	}
}

func TestLearnerLearningItemGetPublishedAndDraft(t *testing.T) {
	env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000031")
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)
	base := learnerLearningItemsBase(courseID, nodeID) + "/" + itemID.String()

	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	// 1. GetPublishedLearningItemByID lookup
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			itemID, courseID, nodeID, "Live", models.LearningItemTypeArticle,
			"desc", metadata, 0, now, models.LearningItemPublishStatePublished,
		))
	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
	// 2. GetAdjacentPublishedLearningItems current lookup
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			itemID, courseID, nodeID, "Live", models.LearningItemTypeArticle,
			"desc", metadata, 0, now, models.LearningItemPublishStatePublished,
		))
	// 3. Previous sibling lookup (empty)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, itemID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))
	// 4. Next sibling lookup (empty)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, itemID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	resp, payload := doJSON(t, env.app, http.MethodGet, base, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get published status = %d payload=%v", resp.StatusCode, payload)
	}
	assertLearningItemBackendSuiteJSend(t, payload, "success")
	assertLearningItemBackendSuiteKeys(t, payload, "status", "data")
	data, _ := payload["data"].(map[string]any)
	assertLearningItemBackendSuiteKeys(t, data, "learning_item", "previous", "next")
	item, _ := data["learning_item"].(map[string]any)
	assertLearningItemBackendSuiteKeys(
		t,
		item,
		"id",
		"title",
		"item_type",
		"description",
		"metadata",
		"publish_state",
	)
	if item["id"] != itemID.String() || item["title"] != "Live" {
		t.Fatalf("data = %v", data)
	}
	if item["publish_state"] != string(models.LearningItemPublishStatePublished) {
		t.Fatalf("publish_state = %v", item["publish_state"])
	}

	// For draft, enrollment then GetPublishedLearningItemByID lookup fails
	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	resp, payload = doJSON(t, env.app, http.MethodGet, base, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("draft/missing status = %d payload=%v", resp.StatusCode, payload)
	}
	if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemNotFound {
		t.Fatalf("message = %v", payload["data"])
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL: %v", err)
	}
}

func TestLearnerLearningItemWrongNode(t *testing.T) {
	env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	missingNode := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600281")

	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	expectLearningItemNodeLookupHTTP(env.mock, courseID, missingNode, false)
	env.mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
		WithArgs(missingNode, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))

	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, missingNode), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrCourseNodeNotFound {
		t.Fatalf("message = %v", payload["data"])
	}
}
