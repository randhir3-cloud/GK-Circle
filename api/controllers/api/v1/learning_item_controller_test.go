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

type learningItemHTTPEnv struct {
	app  *fiber.App
	mock sqlmock.Sqlmock
}

func newLearningItemHTTPEnv(t *testing.T, auth fiber.Handler) *learningItemHTTPEnv {
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
	itemCtrl, err := InitLearningItemController(db, logger, &cfg)
	if err != nil {
		t.Fatalf("init learning item controller: %v", err)
	}

	app := fiber.New()
	admin := app.Group("/api/v1/admin")
	admin.Use(auth)
	courses := admin.Group("/courses")
	nodes := courses.Group("/:" + constants.CourseId + "/nodes")
	items := nodes.Group("/:" + constants.NodeId + "/learning-items")
	items.Post("/", itemCtrl.Create)
	items.Get("/", itemCtrl.List)
	items.Post("/reorder", itemCtrl.Reorder)
	items.Post("/move", itemCtrl.Move)
	items.Get("/:"+constants.ItemId, itemCtrl.GetByID)
	items.Patch("/:"+constants.ItemId, itemCtrl.Update)
	items.Delete("/:"+constants.ItemId, itemCtrl.Delete)

	return &learningItemHTTPEnv{app: app, mock: mock}
}

func learningItemsBase(courseID, nodeID uuid.UUID) string {
	return "/api/v1/admin/courses/" + courseID.String() + "/nodes/" + nodeID.String() + "/learning-items"
}

func learningItemRow(
	itemID, courseID, nodeID uuid.UUID,
	title string,
	itemType models.LearningItemType,
	description interface{},
	metadata []byte,
	position int,
	now time.Time,
) *sqlmock.Rows {
	return learningItemRowWithPublishState(
		itemID, courseID, nodeID, title, itemType, description, metadata, position, now,
		models.LearningItemPublishStateDraft,
	)
}

func learningItemRowWithPublishState(
	itemID, courseID, nodeID uuid.UUID,
	title string,
	itemType models.LearningItemType,
	description interface{},
	metadata []byte,
	position int,
	now time.Time,
	publishState models.LearningItemPublishState,
) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "course_id", "course_node_id", "title", "item_type",
		"description", "metadata", "position", "publish_state", "created_at", "updated_at",
	}).AddRow(
		itemID, courseID, nodeID, title, itemType, description, metadata,
		position, publishState, now, now,
	)
}

func expectLearningItemNodeLockHTTP(mock sqlmock.Sqlmock, courseID, nodeID uuid.UUID, exists bool) {
	rows := sqlmock.NewRows([]string{"id", "course_id"})
	if exists {
		rows.AddRow(nodeID, courseID)
	}
	mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes".*FOR UPDATE`).
		WithArgs(courseID, nodeID, uint(1)).
		WillReturnRows(rows)
}

func expectLearningItemNodeLookupHTTP(mock sqlmock.Sqlmock, courseID, nodeID uuid.UUID, exists bool) {
	rows := sqlmock.NewRows([]string{"id", "course_id"})
	if exists {
		rows.AddRow(nodeID, courseID)
	}
	mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
		WithArgs(courseID, nodeID, uint(1)).
		WillReturnRows(rows)
}

func TestAdminLearningItemRoutesUnauthenticated(t *testing.T) {
	env := newLearningItemHTTPEnv(t, kratosAuthMiddleware(t))
	courseID := uuid.New()
	nodeID := uuid.New()
	resp, payload := doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, nodeID), nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
}

func TestAdminLearningItemRoutesNonAdminForbidden(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.New()
	nodeID := uuid.New()
	resp, payload := doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, nodeID), nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrCourseAdminForbidden {
		t.Fatalf("forbidden message = %v", payload["data"])
	}
}

func TestAdminLearningItemCreateValid(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000001")
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)

	env.mock.ExpectBegin()
	expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
	env.mock.ExpectQuery(`SELECT MAX\("position"\) FROM "learning_items"`).
		WithArgs(nodeID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))
	env.mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
		WithArgs(
			courseID,
			nodeID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			models.LearningItemTypeArticle,
			metadata,
			0,
			models.LearningItemPublishStateDraft,
			"Lesson One",
		).
		WillReturnRows(learningItemRow(
			itemID, courseID, nodeID, "Lesson One", models.LearningItemTypeArticle,
			nil, metadata, 0, now,
		))
	env.mock.ExpectCommit()

	resp, payload := doJSON(t, env.app, http.MethodPost, learningItemsBase(courseID, nodeID), map[string]any{
		"title":     "Lesson One",
		"item_type": string(models.LearningItemTypeArticle),
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	if data["id"] != itemID.String() || data["course_node_id"] != nodeID.String() {
		t.Fatalf("create data = %v", data)
	}
	if data["title"] != "Lesson One" {
		t.Fatalf("title = %v", data["title"])
	}
	if data["publish_state"] != string(models.LearningItemPublishStateDraft) {
		t.Fatalf("publish_state = %v", data["publish_state"])
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL: %v", err)
	}
}

func TestAdminLearningItemCreateInvalidDTO(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.New()
	nodeID := uuid.New()

	resp, payload := doJSON(t, env.app, http.MethodPost, learningItemsBase(courseID, nodeID), map[string]any{
		"item_type": string(models.LearningItemTypeArticle),
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrLearningItemTitleInvalid {
		t.Fatalf("message = %v", payload["data"])
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("DTO validation should not query DB: %v", err)
	}
}

func TestAdminLearningItemCreateInvalidMetadata(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.New()
	nodeID := uuid.New()

	resp, payload := doJSON(t, env.app, http.MethodPost, learningItemsBase(courseID, nodeID), map[string]any{
		"title":     "Lesson",
		"item_type": string(models.LearningItemTypeArticle),
		"metadata":  []any{},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrLearningItemMetadataInvalid {
		t.Fatalf("message = %v", payload["data"])
	}
}

func TestAdminLearningItemGetSuccessAndMisses(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	wrongCourse := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e482")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000001")
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)
	base := learningItemsBase(courseID, nodeID) + "/" + itemID.String()

	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*LIMIT`).
		WithArgs(courseID, nodeID, itemID, uint(1)).
		WillReturnRows(learningItemRow(
			itemID, courseID, nodeID, "Lesson", models.LearningItemTypeArticle,
			"desc", metadata, 0, now,
		))

	resp, payload := doJSON(t, env.app, http.MethodGet, base, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	if data["description"] != "desc" {
		t.Fatalf("description = %v", data["description"])
	}

	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*LIMIT`).
		WithArgs(wrongCourse, nodeID, itemID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	resp, payload = doJSON(t, env.app, http.MethodGet, learningItemsBase(wrongCourse, nodeID)+"/"+itemID.String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("wrong course status = %d payload=%v", resp.StatusCode, payload)
	}

	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*LIMIT`).
		WithArgs(courseID, nodeID, itemID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	resp, payload = doJSON(t, env.app, http.MethodGet, base, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing status = %d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrLearningItemNotFound {
		t.Fatalf("message = %v", payload["data"])
	}
}

func TestAdminLearningItemListEmptyOrderedAndWrongNode(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	missingNode := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600281")
	itemA := uuid.MustParse("019c02a0-1111-7000-8000-000000000001")
	itemB := uuid.MustParse("019c02a0-1111-7000-8000-000000000002")
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)
	base := learningItemsBase(courseID, nodeID)

	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`).
		WithArgs(courseID, nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	resp, payload := doJSON(t, env.app, http.MethodGet, base, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("empty list status = %d payload=%v", resp.StatusCode, payload)
	}
	list, ok := payload["data"].([]any)
	if !ok || len(list) != 0 {
		t.Fatalf("empty list = %v", payload["data"])
	}

	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
	ordered := learningItemRow(itemA, courseID, nodeID, "A", models.LearningItemTypeArticle, nil, metadata, 0, now)
	ordered.AddRow(itemB, courseID, nodeID, "B", models.LearningItemTypeVideo, nil, metadata, 1, models.LearningItemPublishStateDraft, now, now)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`).
		WithArgs(courseID, nodeID).
		WillReturnRows(ordered)

	resp, payload = doJSON(t, env.app, http.MethodGet, base, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ordered list status = %d payload=%v", resp.StatusCode, payload)
	}
	list, ok = payload["data"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("ordered list = %v", payload["data"])
	}
	first, _ := list[0].(map[string]any)
	second, _ := list[1].(map[string]any)
	if first["title"] != "A" || second["title"] != "B" {
		t.Fatalf("order = %v", payload["data"])
	}

	expectLearningItemNodeLookupHTTP(env.mock, courseID, missingNode, false)
	env.mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
		WithArgs(missingNode, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))

	resp, payload = doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, missingNode), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("wrong node status = %d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrCourseNodeNotFound {
		t.Fatalf("message = %v", payload["data"])
	}
}

func TestAdminLearningItemUpdateCases(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000001")
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)
	path := learningItemsBase(courseID, nodeID) + "/" + itemID.String()

	env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
		WithArgs("Renamed", courseID, nodeID, itemID).
		WillReturnRows(learningItemRow(
			itemID, courseID, nodeID, "Renamed", models.LearningItemTypeArticle,
			nil, metadata, 0, now,
		))

	resp, payload := doJSON(t, env.app, http.MethodPatch, path, map[string]any{
		"title": "Renamed",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("partial update status = %d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	if data["title"] != "Renamed" {
		t.Fatalf("title = %v", data["title"])
	}

	resp, payload = doJSON(t, env.app, http.MethodPatch, path, map[string]any{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty patch status = %d payload=%v", resp.StatusCode, payload)
	}
	if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemEmptyPatch {
		t.Fatalf("empty patch message = %v", payload["data"])
	}

	resp, payload = doJSON(t, env.app, http.MethodPatch, path, map[string]any{
		"metadata": []any{},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid metadata status = %d payload=%v", resp.StatusCode, payload)
	}
	if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemMetadataInvalid {
		t.Fatalf("metadata message = %v", payload["data"])
	}

	resp, payload = doJSON(t, env.app, http.MethodPatch, path, map[string]any{
		"metadata": map[string]any{
			"version": 1,
			"blocks": []any{
				map[string]any{
					"id":   "b1",
					"type": "TEXT",
					"data": map[string]any{"text": "{{student-name}}"},
				},
			},
		},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("placeholder status = %d payload=%v", resp.StatusCode, payload)
	}
	if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemPlaceholder {
		t.Fatalf("placeholder message = %v", payload["data"])
	}

	resp, payload = doJSON(t, env.app, http.MethodPatch, path, map[string]any{
		"metadata": map[string]any{
			"version": 1,
			"blocks": []any{
				map[string]any{
					"id":         "b1",
					"type":       "TEXT",
					"data":       map[string]any{},
					"visibility": map[string]any{"mode": "PUBLIC"},
				},
			},
		},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("visibility status = %d payload=%v", resp.StatusCode, payload)
	}
	if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemVisibility {
		t.Fatalf("visibility message = %v", payload["data"])
	}
}

func TestAdminLearningItemDeleteSuccessAndNotFound(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000001")
	path := learningItemsBase(courseID, nodeID) + "/" + itemID.String()

	env.mock.ExpectExec(`DELETE FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, payload := doJSON(t, env.app, http.MethodDelete, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete status = %d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != "success" {
		t.Fatalf("delete data = %v", payload["data"])
	}

	env.mock.ExpectExec(`DELETE FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, payload = doJSON(t, env.app, http.MethodDelete, path, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("delete missing status = %d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrLearningItemNotFound {
		t.Fatalf("message = %v", payload["data"])
	}
}

func TestAdminLearningItemCreateRejectsCourseNodeIDInBody(t *testing.T) {
	// Body course_node_id is ignored; path node is authoritative. Valid create still uses path node.
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000001")
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)
	otherNode := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600299")

	env.mock.ExpectBegin()
	expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
	env.mock.ExpectQuery(`SELECT MAX\("position"\) FROM "learning_items"`).
		WithArgs(nodeID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))
	env.mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
		WithArgs(
			courseID,
			nodeID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			models.LearningItemTypeArticle,
			metadata,
			0,
			models.LearningItemPublishStateDraft,
			"Lesson",
		).
		WillReturnRows(learningItemRow(
			itemID, courseID, nodeID, "Lesson", models.LearningItemTypeArticle,
			nil, metadata, 0, now,
		))
	env.mock.ExpectCommit()

	body := map[string]any{
		"title":          "Lesson",
		"item_type":      string(models.LearningItemTypeArticle),
		"course_node_id": otherNode.String(),
	}

	resp, payload := doJSON(t, env.app, http.MethodPost, learningItemsBase(courseID, nodeID), body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	if data["course_node_id"] != nodeID.String() {
		t.Fatalf("path node must win; got %v", data["course_node_id"])
	}
}

func TestAdminLearningItemPublishState(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000001")
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)
	base := learningItemsBase(courseID, nodeID)
	path := base + "/" + itemID.String()

	t.Run("create omitted defaults draft", func(t *testing.T) {
		env.mock.ExpectBegin()
		expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT MAX\("position"\) FROM "learning_items"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))
		env.mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
			WithArgs(
				courseID, nodeID, sqlmock.AnyArg(), sqlmock.AnyArg(),
				models.LearningItemTypeArticle, metadata, 0,
				models.LearningItemPublishStateDraft, "Lesson",
			).
			WillReturnRows(learningItemRow(
				itemID, courseID, nodeID, "Lesson", models.LearningItemTypeArticle, nil, metadata, 0, now,
			))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title": "Lesson", "item_type": string(models.LearningItemTypeArticle),
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["publish_state"] != "DRAFT" {
			t.Fatalf("publish_state = %v", data["publish_state"])
		}
	})

	t.Run("create explicit published", func(t *testing.T) {
		env.mock.ExpectBegin()
		expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT MAX\("position"\) FROM "learning_items"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))
		env.mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
			WithArgs(
				courseID, nodeID, sqlmock.AnyArg(), sqlmock.AnyArg(),
				models.LearningItemTypeArticle, metadata, 0,
				models.LearningItemPublishStatePublished, "Lesson",
			).
			WillReturnRows(learningItemRowWithPublishState(
				itemID, courseID, nodeID, "Lesson", models.LearningItemTypeArticle, nil, metadata, 0, now,
				models.LearningItemPublishStatePublished,
			))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title": "Lesson", "item_type": string(models.LearningItemTypeArticle),
			"publish_state": "PUBLISHED",
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["publish_state"] != "PUBLISHED" {
			t.Fatalf("publish_state = %v", data["publish_state"])
		}
	})

	t.Run("create null rejected", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title": "Lesson", "item_type": string(models.LearningItemTypeArticle),
			"publish_state": nil,
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemPublishState {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("patch published", func(t *testing.T) {
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(models.LearningItemPublishStatePublished, courseID, nodeID, itemID).
			WillReturnRows(learningItemRowWithPublishState(
				itemID, courseID, nodeID, "Lesson", models.LearningItemTypeArticle, nil, metadata, 0, now,
				models.LearningItemPublishStatePublished,
			))
		resp, payload := doJSON(t, env.app, http.MethodPatch, path, map[string]any{
			"publish_state": "PUBLISHED",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["publish_state"] != "PUBLISHED" {
			t.Fatalf("publish_state = %v", data["publish_state"])
		}
	})

	t.Run("patch null rejected", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPatch, path, map[string]any{
			"publish_state": nil,
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemPublishState {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("patch lowercase rejected", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPatch, path, map[string]any{
			"publish_state": "published",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemPublishState {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("patch empty rejected", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPatch, path, map[string]any{
			"publish_state": "",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
	})
}

func TestAdminLearningItemReorderAuthAndValidation(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	aID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	bID := uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a")
	foreignID := uuid.MustParse("019c01cc-1111-7222-8333-944455556666")
	base := learningItemsBase(courseID, nodeID) + "/reorder"
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)

	t.Run("unauthenticated", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, kratosAuthMiddleware(t))
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{},
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("non-admin forbidden", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{},
		})
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
	})

	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))

	t.Run("malformed uuid", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{"not-a-uuid"},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemReorderInvalid {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("duplicate ids", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{aID.String(), aID.String()},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemReorderMismatch {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("missing id mismatch", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items".*FOR UPDATE`).
			WithArgs(courseID, nodeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}).AddRow(aID, courseID, nodeID, "A", models.LearningItemTypeArticle, nil, []byte(`null`), 0, models.LearningItemPublishStateDraft, now, now).
				AddRow(bID, courseID, nodeID, "B", models.LearningItemTypeArticle, nil, []byte(`null`), 1, models.LearningItemPublishStateDraft, now, now))
		env.mock.ExpectRollback()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{aID.String()},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemReorderMismatch {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("foreign node id mismatch", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items".*FOR UPDATE`).
			WithArgs(courseID, nodeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}).AddRow(aID, courseID, nodeID, "A", models.LearningItemTypeArticle, nil, []byte(`null`), 0, models.LearningItemPublishStateDraft, now, now))
		env.mock.ExpectRollback()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{foreignID.String()},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
	})

	t.Run("empty node success", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items".*FOR UPDATE`).
			WithArgs(courseID, nodeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["noop"] != true || data["learning_item_count"] != float64(0) || data["positions_updated"] != float64(0) {
			t.Fatalf("data = %v", data)
		}
		if data["course_node_id"] != nodeID.String() {
			t.Fatalf("course_node_id = %v", data["course_node_id"])
		}
	})

	t.Run("single item noop", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items".*FOR UPDATE`).
			WithArgs(courseID, nodeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}).AddRow(aID, courseID, nodeID, "A", models.LearningItemTypeArticle, nil, []byte(`null`), 0, models.LearningItemPublishStateDraft, now, now))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{aID.String()},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["noop"] != true || data["positions_updated"] != float64(0) || data["learning_item_count"] != float64(1) {
			t.Fatalf("data = %v", data)
		}
	})

	t.Run("success reorder payload", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items".*FOR UPDATE`).
			WithArgs(courseID, nodeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}).AddRow(aID, courseID, nodeID, "A", models.LearningItemTypeArticle, nil, []byte(`null`), 0, models.LearningItemPublishStateDraft, now, now).
				AddRow(bID, courseID, nodeID, "B", models.LearningItemTypeArticle, nil, []byte(`null`), 1, models.LearningItemPublishStateDraft, now, now))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(4, courseID, nodeID, aID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(aID))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(5, courseID, nodeID, bID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(bID))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(1, courseID, nodeID, aID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(aID))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(0, courseID, nodeID, bID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(bID))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{bID.String(), aID.String()},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["noop"] != false || data["positions_updated"] != float64(2) || data["learning_item_count"] != float64(2) {
			t.Fatalf("data = %v", data)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items".*FOR UPDATE`).
			WithArgs(courseID, nodeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}).AddRow(aID, courseID, nodeID, "A", models.LearningItemTypeArticle, nil, []byte(`null`), 0, models.LearningItemPublishStateDraft, now, now).
				AddRow(bID, courseID, nodeID, "B", models.LearningItemTypeArticle, nil, []byte(`null`), 1, models.LearningItemPublishStateDraft, now, now))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(4, courseID, nodeID, aID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		env.mock.ExpectRollback()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_item_ids": []string{bID.String(), aID.String()},
		})
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemReorderConflict {
			t.Fatalf("message = %v", payload["data"])
		}
	})
}

func TestAdminLearningItemMoveAuthAndValidation(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	sourceNodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	destNodeID := uuid.MustParse("019c01c7-1111-7222-8333-944455556666")
	aID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	bID := uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a")
	foreignID := uuid.MustParse("019c01cc-1111-7222-8333-944455556666")
	base := learningItemsBase(courseID, sourceNodeID) + "/move"
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)

	siblingLock := func(mock sqlmock.Sqlmock, nodeID uuid.UUID, rows *sqlmock.Rows) {
		mock.ExpectQuery(`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items".*FOR UPDATE`).
			WithArgs(courseID, nodeID).
			WillReturnRows(rows)
	}
	emptyItemRows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		})
	}
	itemRows := func(nodeID uuid.UUID, items ...struct {
		id       uuid.UUID
		position int
		title    string
	}) *sqlmock.Rows {
		rows := emptyItemRows()
		for _, item := range items {
			rows.AddRow(
				item.id, courseID, nodeID, item.title, models.LearningItemTypeArticle, nil, []byte(`null`),
				item.position, models.LearningItemPublishStateDraft, now, now,
			)
		}
		return rows
	}

	t.Run("unauthenticated", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, kratosAuthMiddleware(t))
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"destination_node_id": destNodeID.String(),
			"ordered_item_ids":    []string{},
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("non-admin forbidden", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"destination_node_id": destNodeID.String(),
			"ordered_item_ids":    []string{},
		})
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
	})

	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))

	t.Run("invalid json", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, "{not-json")
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("invalid destination uuid", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"destination_node_id": "not-a-uuid",
			"ordered_item_ids":    []string{},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemMoveInvalid {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("invalid ordered uuid", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"destination_node_id": destNodeID.String(),
			"ordered_item_ids":    []string{"not-a-uuid"},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemMoveInvalid {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("same node", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"destination_node_id": sourceNodeID.String(),
			"ordered_item_ids":    []string{aID.String()},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemMoveSameNode {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("mismatch foreign id", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, destNodeID, true)
		expectLearningItemNodeLockHTTP(env.mock, courseID, sourceNodeID, true)
		siblingLock(env.mock, destNodeID, emptyItemRows())
		siblingLock(env.mock, sourceNodeID, itemRows(sourceNodeID, struct {
			id       uuid.UUID
			position int
			title    string
		}{aID, 0, "A"}))
		env.mock.ExpectRollback()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"destination_node_id": destNodeID.String(),
			"ordered_item_ids":    []string{foreignID.String()},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemMoveMismatch {
			t.Fatalf("message = %v", payload["data"])
		}
	})

	t.Run("empty noop payload", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, destNodeID, true)
		expectLearningItemNodeLockHTTP(env.mock, courseID, sourceNodeID, true)
		siblingLock(env.mock, destNodeID, itemRows(destNodeID, struct {
			id       uuid.UUID
			position int
			title    string
		}{bID, 0, "B"}))
		siblingLock(env.mock, sourceNodeID, itemRows(sourceNodeID, struct {
			id       uuid.UUID
			position int
			title    string
		}{aID, 0, "A"}))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"destination_node_id": destNodeID.String(),
			"ordered_item_ids":    []string{},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["noop"] != true || data["items_moved"] != float64(0) {
			t.Fatalf("data = %v", data)
		}
		if data["source_item_count"] != float64(1) || data["destination_item_count"] != float64(1) {
			t.Fatalf("counts = %v", data)
		}
		if data["source_node_id"] != sourceNodeID.String() || data["destination_node_id"] != destNodeID.String() {
			t.Fatalf("node ids = %v", data)
		}
	})

	t.Run("success move payload", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, destNodeID, true)
		expectLearningItemNodeLockHTTP(env.mock, courseID, sourceNodeID, true)
		siblingLock(env.mock, destNodeID, emptyItemRows())
		siblingLock(env.mock, sourceNodeID, itemRows(sourceNodeID,
			struct {
				id       uuid.UUID
				position int
				title    string
			}{aID, 0, "A"},
			struct {
				id       uuid.UUID
				position int
				title    string
			}{bID, 1, "B"},
		))
		// sourceTempBase=4
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(4, courseID, sourceNodeID, aID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(aID))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(5, courseID, sourceNodeID, bID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(bID))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(destNodeID, courseID, sourceNodeID, bID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(bID))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(0, courseID, sourceNodeID, aID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(aID))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(0, courseID, destNodeID, bID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(bID))
		siblingLock(env.mock, sourceNodeID, itemRows(sourceNodeID, struct {
			id       uuid.UUID
			position int
			title    string
		}{aID, 0, "A"}))
		siblingLock(env.mock, destNodeID, itemRows(destNodeID, struct {
			id       uuid.UUID
			position int
			title    string
		}{bID, 0, "B"}))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"destination_node_id": destNodeID.String(),
			"ordered_item_ids":    []string{bID.String()},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["noop"] != false || data["items_moved"] != float64(1) {
			t.Fatalf("data = %v", data)
		}
		if data["source_item_count"] != float64(1) || data["destination_item_count"] != float64(1) {
			t.Fatalf("counts = %v", data)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		env.mock.ExpectBegin()
		env.mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
		expectLearningItemNodeLockHTTP(env.mock, courseID, destNodeID, true)
		expectLearningItemNodeLockHTTP(env.mock, courseID, sourceNodeID, true)
		siblingLock(env.mock, destNodeID, emptyItemRows())
		siblingLock(env.mock, sourceNodeID, itemRows(sourceNodeID,
			struct {
				id       uuid.UUID
				position int
				title    string
			}{aID, 0, "A"},
			struct {
				id       uuid.UUID
				position int
				title    string
			}{bID, 1, "B"},
		))
		env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(4, courseID, sourceNodeID, aID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		env.mock.ExpectRollback()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"destination_node_id": destNodeID.String(),
			"ordered_item_ids":    []string{bID.String()},
		})
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		if msg, _ := payload["data"].(string); msg != constants.ErrLearningItemMoveConflict {
			t.Fatalf("message = %v", payload["data"])
		}
	})
}
