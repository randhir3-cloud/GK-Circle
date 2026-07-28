package v1

import (
	"errors"
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

type publishContractHTTPEnv struct {
	app  *fiber.App
	mock sqlmock.Sqlmock
}

func newPublishContractHTTPEnv(t *testing.T, auth fiber.Handler) *publishContractHTTPEnv {
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

	return &publishContractHTTPEnv{app: app, mock: mock}
}

func TestLearnerLearningItemPublishContract_Unauthenticated(t *testing.T) {
	env := newPublishContractHTTPEnv(t, kratosAuthMiddleware(t))
	courseID := uuid.New()
	nodeID := uuid.New()
	itemID := uuid.New()

	// 1. Test List Endpoint
	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID), nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for List, got status = %d payload = %v", resp.StatusCode, payload)
	}
	assertPublishContractUnauthorized(t, payload)

	// 2. Test GetByID Endpoint
	detailURL := learnerLearningItemsBase(courseID, nodeID) + "/" + itemID.String()
	resp, payload = doJSON(t, env.app, http.MethodGet, detailURL, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for GetByID, got status = %d payload = %v", resp.StatusCode, payload)
	}
	assertPublishContractUnauthorized(t, payload)

	// Ensure no SQL expectations were configured or consumed on this env
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet SQL expectations: %v", err)
	}
}

func assertPublishContractUnauthorized(t *testing.T, payload map[string]any) {
	t.Helper()
	if status, _ := payload["status"].(string); status != "error" {
		t.Errorf("expected stable status 'error', got %v", status)
	}
	if message, _ := payload["message"].(string); message != constants.ErrKratosIDEmpty {
		t.Errorf("expected stable message %q, got %v", constants.ErrKratosIDEmpty, payload["message"])
	}
	if code, _ := payload["code"].(float64); code != http.StatusUnauthorized {
		t.Errorf("expected stable code %d, got %v", http.StatusUnauthorized, payload["code"])
	}
	if data, exists := payload["data"]; exists && data != nil {
		t.Errorf("unauthorized response must not expose repository payload, got %v", data)
	}
}

func TestLearnerLearningItemPublishContract_AuthenticatedList(t *testing.T) {
	env := newPublishContractHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")

	// Set lookup expectations on course_nodes lookup
	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)

	// Configure sqlmock with out-of-order and adversarial draft-labeled results.
	lowerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	higherID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	now := time.Now()
	emptyMetadata := []byte(`{"version":1,"blocks":[]}`)

	rows := sqlmock.NewRows([]string{
		"id", "course_id", "course_node_id", "title", "item_type",
		"description", "metadata", "position", "publish_state", "created_at", "updated_at",
	}).AddRow(
		lowerID, courseID, nodeID, "Adversarial Draft Item", models.LearningItemTypeArticle,
		nil, emptyMetadata, 10, models.LearningItemPublishStateDraft, now, now,
	).AddRow(
		higherID, courseID, nodeID, "Adversarial Sort Item", models.LearningItemTypeVideo,
		nil, emptyMetadata, 2, models.LearningItemPublishStatePublished, now, now,
	)

	env.mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished).
		WillReturnRows(rows)

	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got status = %d payload = %v", resp.StatusCode, payload)
		if err := env.mock.ExpectationsWereMet(); err != nil {
			t.Errorf("SQL MOCK UNMET IN LIST: %v", err)
		}
		return
	}

	dataList, ok := payload["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data key to contain an array, got %v", payload["data"])
	}
	if len(dataList) != 2 {
		t.Fatalf("expected 2 items, got %d", len(dataList))
	}

	// Verify the items are forwarded in the exact order returned by the repository,
	// proving that the controller does not sort by position.
	item1, _ := dataList[0].(map[string]any)
	item2, _ := dataList[1].(map[string]any)

	if item1["id"] != lowerID.String() {
		t.Fatalf("item1 mismatch: %v", item1)
	}
	if _, exists := item1["position"]; exists {
		t.Fatalf("learner list contract must omit position: %v", item1)
	}
	// The mock deliberately returns a DRAFT-labelled row despite the repository’s
	// published-only query to prove the controller does not perform a second,
	// Go-level publication filter.
	if item1["publish_state"] != string(models.LearningItemPublishStateDraft) {
		t.Fatalf("expected draft item to not be filtered out by controller, got publish_state = %v", item1["publish_state"])
	}

	if item2["id"] != higherID.String() {
		t.Fatalf("item2 mismatch: %v", item2)
	}
	if item1["title"] != "Adversarial Draft Item" ||
		item1["item_type"] != string(models.LearningItemTypeArticle) ||
		item1["description"] != nil ||
		item1["metadata"] == nil {
		t.Fatalf("learner list response contract mismatch: %v", item1)
	}

	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestLearnerLearningItemPublishContract_AuthenticatedGetByID(t *testing.T) {
	env := newPublishContractHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	prevID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	nextID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	now := time.Now()

	// Metadata represents a visibility-projected payload already formatted by repository.
	projectedMetadata := []byte(`{"version":2,"blocks":[{"id":"b1","type":"TEXT","data":{},"visibility":{"mode":"AUTHENTICATED"}}]}`)

	// 1. GetPublishedLearningItemByID lookup (adversarially marked DRAFT to check filtering bypass)
	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			itemID, courseID, nodeID, "Adversarial Detail Item", models.LearningItemTypeArticle,
			nil, projectedMetadata, 2, now, models.LearningItemPublishStateDraft,
		))

	// 2. ensureLearningItemNodeInCourse lookup
	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)

	// 3. GetAdjacentPublishedLearningItems current lookup (to fetch its position and ID)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			itemID, courseID, nodeID, "Adversarial Detail Item", models.LearningItemTypeArticle,
			nil, projectedMetadata, 2, now, models.LearningItemPublishStateDraft,
		))

	// 4. Fetch previous
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 2, 2, itemID, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			prevID, courseID, nodeID, "Prev A", models.LearningItemTypeArticle,
			nil, nil, 1, now, models.LearningItemPublishStatePublished,
		))

	// 5. Fetch next
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 2, 2, itemID, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			nextID, courseID, nodeID, "Next C", models.LearningItemTypeVideo,
			nil, nil, 3, now, models.LearningItemPublishStatePublished,
		))

	detailURL := learnerLearningItemsBase(courseID, nodeID) + "/" + itemID.String()
	resp, payload := doJSON(t, env.app, http.MethodGet, detailURL, nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got status = %d payload = %v", resp.StatusCode, payload)
		if err := env.mock.ExpectationsWereMet(); err != nil {
			t.Errorf("SQL MOCK UNMET IN GET_BY_ID: %v", err)
		}
		return
	}

	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data key to contain an object, got %v", payload["data"])
	}

	// Verify wrapped response structure
	learningItem, _ := data["learning_item"].(map[string]interface{})
	previousLink, _ := data["previous"].(map[string]interface{})
	nextLink, _ := data["next"].(map[string]interface{})

	if learningItem["id"] != itemID.String() {
		t.Fatalf("learning_item ID mismatch: %v", learningItem["id"])
	}
	// Verify detail bypasses draft filtering
	if learningItem["publish_state"] != string(models.LearningItemPublishStateDraft) {
		t.Fatalf("expected detail draft bypass, got publish_state = %v", learningItem["publish_state"])
	}

	// Projected metadata is forwarded byte-for-byte without controller-side parsing
	metadataMap, _ := learningItem["metadata"].(map[string]interface{})
	if metadataMap == nil {
		t.Fatalf("expected metadata map, got %v", learningItem["metadata"])
	}
	if metadataMap["version"] != float64(2) {
		t.Fatalf("expected version 2, got %v", metadataMap["version"])
	}

	// Verify navigation values are serialized without controller filtering
	if previousLink["id"] != prevID.String() || previousLink["title"] != "Prev A" {
		t.Fatalf("previous navigation mismatch: %v", previousLink)
	}
	if nextLink["id"] != nextID.String() || nextLink["title"] != "Next C" {
		t.Fatalf("next navigation mismatch: %v", nextLink)
	}

	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestLearnerLearningItemPublishContract_ErrorMapping(t *testing.T) {
	env := newPublishContractHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.New()

	// 1. Test ErrLearningItemNotFound maps to 404
	// To simulate found = false without database errors, we return successfully with no rows.
	emptyRows := sqlmock.NewRows([]string{
		"id", "course_id", "course_node_id", "title", "item_type",
		"description", "metadata", "position", "publish_state", "created_at", "updated_at",
	})
	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(emptyRows)

	detailURL := learnerLearningItemsBase(courseID, nodeID) + "/" + itemID.String()
	resp, payload := doJSON(t, env.app, http.MethodGet, detailURL, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got status = %d payload = %v", resp.StatusCode, payload)
		if err := env.mock.ExpectationsWereMet(); err != nil {
			t.Errorf("SQL MOCK UNMET IN NOT FOUND: %v", err)
		}
		return
	}
	if status, _ := payload["status"].(string); status != "fail" {
		t.Errorf("expected status 'fail', got %v", status)
	}
	if data, _ := payload["data"].(string); data != constants.ErrLearningItemNotFound {
		t.Errorf("expected error message %q, got %q", constants.ErrLearningItemNotFound, data)
	}

	// 2. Test internal error maps to 500
	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnError(errors.New("unexpected database error"))

	resp, payload = doJSON(t, env.app, http.MethodGet, detailURL, nil)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got status = %d payload = %v", resp.StatusCode, payload)
		if err := env.mock.ExpectationsWereMet(); err != nil {
			t.Errorf("SQL MOCK UNMET IN INTERNAL ERROR: %v", err)
		}
		return
	}
	if status, _ := payload["status"].(string); status != "error" {
		t.Errorf("expected status 'error', got %v", status)
	}
	if message, _ := payload["message"].(string); message != constants.ErrGetLearningItem {
		t.Errorf("expected internal error message %q, got %q", constants.ErrGetLearningItem, message)
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
