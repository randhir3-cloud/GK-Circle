package v1

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

func TestLearningItemVisibility_LearnerOmitsHiddenInstructorPremium(t *testing.T) {
	env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000301")
	now := time.Date(2026, time.July, 26, 13, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[{"id":"all","type":"TEXT","data":{"text":"a"},"visibility":{"mode":"ALL"}},{"id":"auth","type":"TEXT","data":{"text":"b"},"visibility":{"mode":"AUTHENTICATED"}},{"id":"hidden","type":"TEXT","data":{"text":"c"},"visibility":{"mode":"HIDDEN"}},{"id":"instructor","type":"TEXT","data":{"text":"d"},"visibility":{"mode":"INSTRUCTOR"}},{"id":"premium","type":"TEXT","data":{"text":"e"},"visibility":{"mode":"PREMIUM"}},{"id":"tail","type":"TEXT","data":{"text":"f"},"visibility":{"mode":"ALL"}}]}`)

	// 1. GetPublishedLearningItemByID lookup
	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			itemID, courseID, nodeID, "Learner Visible", models.LearningItemTypeArticle,
			nil, metadata, 0, now, models.LearningItemPublishStatePublished,
		))
	// 2. ensureLearningItemNodeInCourse lookup
	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
	// 3. GetAdjacentPublishedLearningItems current lookup
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			itemID, courseID, nodeID, "Learner Visible", models.LearningItemTypeArticle,
			nil, metadata, 0, now, models.LearningItemPublishStatePublished,
		))
	// 4. Previous sibling lookup (empty)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, itemID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))
	// 5. Next sibling lookup (empty)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, itemID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID)+"/"+itemID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	item, _ := data["learning_item"].(map[string]any)
	meta, _ := item["metadata"].(map[string]any)
	blocks, _ := meta["blocks"].([]any)
	ids := learnerBlockIDs(t, blocks)
	want := []string{"all", "auth", "tail"}
	if len(ids) != len(want) {
		t.Fatalf("ids=%v want=%v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("order ids=%v want=%v", ids, want)
		}
	}
	for _, blocked := range []string{"hidden", "instructor", "premium"} {
		for _, id := range ids {
			if id == blocked {
				t.Fatalf("blocked mode %s leaked into learner response", blocked)
			}
		}
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLearningItemVisibility_LearnerListPreservesOrderAndPublishFilter(t *testing.T) {
	env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemA := uuid.MustParse("019c02a0-1111-7000-8000-000000000302")
	itemB := uuid.MustParse("019c02a0-1111-7000-8000-000000000303")
	now := time.Date(2026, time.July, 26, 13, 0, 0, 0, time.UTC)
	metaA := []byte(`{"version":1,"blocks":[{"id":"a1","type":"TEXT","data":{},"visibility":{"mode":"ALL"}},{"id":"a2","type":"TEXT","data":{},"visibility":{"mode":"HIDDEN"}}]}`)
	metaB := []byte(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{},"visibility":{"mode":"AUTHENTICATED"}},{"id":"b2","type":"TEXT","data":{},"visibility":{"mode":"PREMIUM"}}]}`)

	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
	rows := learningItemRowWithPublishState(
		itemA, courseID, nodeID, "First", models.LearningItemTypeArticle,
		nil, metaA, 0, now, models.LearningItemPublishStatePublished,
	)
	rows.AddRow(
		itemB, courseID, nodeID, "Second", models.LearningItemTypeVideo,
		nil, metaB, 1, models.LearningItemPublishStatePublished, now, now,
	)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
		WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished).
		WillReturnRows(rows)

	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	list, _ := payload["data"].([]any)
	if len(list) != 2 {
		t.Fatalf("list len=%d payload=%v", len(list), payload["data"])
	}
	first, _ := list[0].(map[string]any)
	second, _ := list[1].(map[string]any)
	if first["title"] != "First" || second["title"] != "Second" {
		t.Fatalf("order broken: %v", payload["data"])
	}
	firstMeta, _ := first["metadata"].(map[string]any)
	secondMeta, _ := second["metadata"].(map[string]any)
	assertLearnerBlockIDs(t, firstMeta["blocks"], []string{"a1"})
	assertLearnerBlockIDs(t, secondMeta["blocks"], []string{"b1"})
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLearningItemVisibility_LearnerDraftStill404(t *testing.T) {
	env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000304")

	expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
		WithArgs(courseID, nodeID, itemID, models.LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID)+"/"+itemID.String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLearningItemVisibility_AdminContrastUnfiltered(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000305")
	now := time.Date(2026, time.July, 26, 13, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[{"id":"all","type":"TEXT","data":{},"visibility":{"mode":"ALL"}},{"id":"hidden","type":"TEXT","data":{},"visibility":{"mode":"HIDDEN"}},{"id":"instructor","type":"TEXT","data":{},"visibility":{"mode":"INSTRUCTOR"}},{"id":"premium","type":"TEXT","data":{},"visibility":{"mode":"PREMIUM"}}]}`)

	env.mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			itemID, courseID, nodeID, "Admin Full", models.LearningItemTypeArticle,
			nil, metadata, 0, now, models.LearningItemPublishStatePublished,
		))

	resp, payload := doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, nodeID)+"/"+itemID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin status = %d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	meta, _ := data["metadata"].(map[string]any)
	assertLearnerBlockIDs(t, meta["blocks"], []string{"all", "hidden", "instructor", "premium"})
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func learnerBlockIDs(t *testing.T, blocks []any) []string {
	t.Helper()
	ids := make([]string, 0, len(blocks))
	for _, raw := range blocks {
		block, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("block type = %T", raw)
		}
		id, _ := block["id"].(string)
		ids = append(ids, id)
	}
	return ids
}

func assertLearnerBlockIDs(t *testing.T, raw any, want []string) {
	t.Helper()
	blocks, ok := raw.([]any)
	if !ok {
		// json decoder may leave numbers etc; also accept json.RawMessage path via re-marshal
		encoded, err := json.Marshal(raw)
		if err != nil {
			t.Fatalf("blocks type = %T", raw)
		}
		if err := json.Unmarshal(encoded, &blocks); err != nil {
			t.Fatalf("blocks type = %T value=%v", raw, raw)
		}
	}
	if blocks == nil {
		t.Fatal("blocks is nil")
	}
	ids := learnerBlockIDs(t, blocks)
	if len(ids) != len(want) {
		t.Fatalf("ids=%v want=%v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids=%v want=%v", ids, want)
		}
	}
}
