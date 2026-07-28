package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

func learningItemDraftFilteringHTTPIds() (courseID, nodeID, publishedID, draftID uuid.UUID) {
	return uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481"),
		uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"),
		uuid.MustParse("019c02a0-1111-7000-8000-000000000201"),
		uuid.MustParse("019c02a0-1111-7000-8000-000000000202")
}

func assertNoDraftDiscoveryInLearnerPayload(t *testing.T, payload map[string]any, draftID uuid.UUID, draftTitle string) {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	body := string(raw)
	if strings.Contains(body, draftID.String()) {
		t.Fatalf("learner payload leaked draft ID: %s", body)
	}
	if draftTitle != "" && strings.Contains(body, draftTitle) {
		t.Fatalf("learner payload leaked draft title: %s", body)
	}
	if strings.Contains(body, `"publish_state":"DRAFT"`) || strings.Contains(body, `"publish_state": "DRAFT"`) {
		t.Fatalf("learner payload leaked DRAFT publish_state: %s", body)
	}
	if strings.Contains(strings.ToLower(body), "unpublished") {
		t.Fatalf("learner payload leaked unpublished wording: %s", body)
	}
}

func publicFailFields(payload map[string]any) (status any, data any) {
	return payload["status"], payload["data"]
}

func TestLearningItemDraftFilteringLearnerList(t *testing.T) {
	courseID, nodeID, publishedID, draftID := learningItemDraftFilteringHTTPIds()
	now := time.Date(2026, time.July, 26, 10, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)
	draftTitle := "Secret Draft Title Never Leak"

	t.Run("mixed_excludes_draft_ids_and_fields", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
		// SQL returns only published rows; draft never appears in response.
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished).
			WillReturnRows(learningItemRowWithPublishState(
				publishedID, courseID, nodeID, "Published Only", models.LearningItemTypeArticle,
				nil, metadata, 0, now, models.LearningItemPublishStatePublished,
			))

		resp, payload := doJSON(t, env.app, http.MethodGet, learnerLearningItemsBase(courseID, nodeID), nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		list, ok := payload["data"].([]any)
		if !ok || len(list) != 1 {
			t.Fatalf("list = %v", payload["data"])
		}
		item, _ := list[0].(map[string]any)
		if item["id"] != publishedID.String() {
			t.Fatalf("id = %v", item["id"])
		}
		if item["publish_state"] != string(models.LearningItemPublishStatePublished) {
			t.Fatalf("publish_state = %v", item["publish_state"])
		}
		assertNoDraftDiscoveryInLearnerPayload(t, payload, draftID, draftTitle)
		if err := env.mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("all_draft_empty_200", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
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
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		list, ok := payload["data"].([]any)
		if !ok || len(list) != 0 {
			t.Fatalf("list = %v", payload["data"])
		}
		assertNoDraftDiscoveryInLearnerPayload(t, payload, draftID, draftTitle)
	})
}

func TestLearningItemDraftFilteringLearnerGetEquivalence(t *testing.T) {
	courseID, nodeID, publishedID, draftID := learningItemDraftFilteringHTTPIds()
	now := time.Date(2026, time.July, 26, 10, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)
	missingID := uuid.MustParse("019c02a0-1111-7000-8000-000000000299")
	base := learnerLearningItemsBase(courseID, nodeID)

	t.Run("published_200", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		// 1. GetPublishedLearningItemByID lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, publishedID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				publishedID, courseID, nodeID, "Live", models.LearningItemTypeArticle,
				"desc", metadata, 0, now, models.LearningItemPublishStatePublished,
			))
		// 2. ensureLearningItemNodeInCourse lookup
		expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
		// 3. GetAdjacentPublishedLearningItems current lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, publishedID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				publishedID, courseID, nodeID, "Live", models.LearningItemTypeArticle,
				"desc", metadata, 0, now, models.LearningItemPublishStatePublished,
			))
		// 4. Previous adjacent sibling lookup (empty)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, publishedID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		// 5. Next adjacent sibling lookup (empty)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, publishedID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))

		resp, payload := doJSON(t, env.app, http.MethodGet, base+"/"+publishedID.String(), nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		item, _ := data["learning_item"].(map[string]any)
		if item["id"] != publishedID.String() || item["publish_state"] != string(models.LearningItemPublishStatePublished) {
			t.Fatalf("item = %v", item)
		}
	})

	t.Run("draft_and_missing_same_public_404", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		empty := sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		})
		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, draftID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(empty)
		draftResp, draftPayload := doJSON(t, env.app, http.MethodGet, base+"/"+draftID.String(), nil)

		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, missingID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		missingResp, missingPayload := doJSON(t, env.app, http.MethodGet, base+"/"+missingID.String(), nil)

		if draftResp.StatusCode != http.StatusNotFound || missingResp.StatusCode != http.StatusNotFound {
			t.Fatalf("status draft=%d missing=%d", draftResp.StatusCode, missingResp.StatusCode)
		}
		dStatus, dData := publicFailFields(draftPayload)
		mStatus, mData := publicFailFields(missingPayload)
		if dStatus != mStatus || dData != mData {
			t.Fatalf("public fields differ draft=%v/%v missing=%v/%v", dStatus, dData, mStatus, mData)
		}
		if dData != constants.ErrLearningItemNotFound {
			t.Fatalf("message = %v", dData)
		}
		assertNoDraftDiscoveryInLearnerPayload(t, draftPayload, draftID, "Secret Draft Title Never Leak")
		assertNoDraftDiscoveryInLearnerPayload(t, missingPayload, draftID, "Secret Draft Title Never Leak")
	})
}

func TestLearningItemDraftFilteringAdminContrastHTTP(t *testing.T) {
	courseID, nodeID, publishedID, draftID := learningItemDraftFilteringHTTPIds()
	now := time.Date(2026, time.July, 26, 10, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)

	t.Run("admin_list_includes_draft", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
		expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
		rows := learningItemRowWithPublishState(
			draftID, courseID, nodeID, "Draft Lesson", models.LearningItemTypeArticle,
			nil, metadata, 0, now, models.LearningItemPublishStateDraft,
		)
		rows.AddRow(
			publishedID, courseID, nodeID, "Published Lesson", models.LearningItemTypeVideo,
			nil, metadata, 1, models.LearningItemPublishStatePublished, now, now,
		)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`).
			WithArgs(courseID, nodeID).
			WillReturnRows(rows)

		resp, payload := doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, nodeID), nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		list, ok := payload["data"].([]any)
		if !ok || len(list) != 2 {
			t.Fatalf("list = %v", payload["data"])
		}
		var sawDraft, sawPublished bool
		for _, raw := range list {
			item, _ := raw.(map[string]any)
			switch item["publish_state"] {
			case string(models.LearningItemPublishStateDraft):
				sawDraft = true
				if item["id"] != draftID.String() {
					t.Fatalf("draft id = %v", item["id"])
				}
			case string(models.LearningItemPublishStatePublished):
				sawPublished = true
			}
		}
		if !sawDraft || !sawPublished {
			t.Fatalf("admin list missing draft/published: %v", payload["data"])
		}
	})

	t.Run("admin_get_returns_draft", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
			WithArgs(courseID, nodeID, draftID, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				draftID, courseID, nodeID, "Draft Lesson", models.LearningItemTypeArticle,
				nil, metadata, 0, now, models.LearningItemPublishStateDraft,
			))
		resp, payload := doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, nodeID)+"/"+draftID.String(), nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["id"] != draftID.String() || data["publish_state"] != string(models.LearningItemPublishStateDraft) {
			t.Fatalf("data = %v", data)
		}
	})

	t.Run("non_admin_rejected", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		resp, _ := doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, nodeID), nil)
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("non-admin status = %d, want 403/401", resp.StatusCode)
		}
	})
}
