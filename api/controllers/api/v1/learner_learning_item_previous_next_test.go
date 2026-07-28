package v1

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

func learningItemAdjacentTestIDs() (courseID, nodeID, aID, bID, cID, otherNodeID, otherCourseID uuid.UUID) {
	return uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481"),
		uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"),
		uuid.MustParse("019c02f0-aaaa-7000-8000-000000000001"),
		uuid.MustParse("019c02f0-aaaa-7000-8000-000000000002"),
		uuid.MustParse("019c02f0-aaaa-7000-8000-000000000003"),
		uuid.MustParse("019c01c9-bbbb-78e2-a366-690bfd600281"),
		uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e482")
}

func TestLearnerLearningItemPreviousNext(t *testing.T) {
	courseID, nodeID, aID, bID, cID, _, _ := learningItemAdjacentTestIDs()
	now := time.Date(2026, time.July, 26, 9, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)

	t.Run("navigation returned (middle item)", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		base := learnerLearningItemsBase(courseID, nodeID) + "/" + bID.String()

		// 1. GetPublishedLearningItemByID lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, bID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				bID, courseID, nodeID, "Current B", models.LearningItemTypeArticle,
				nil, metadata, 1, now, models.LearningItemPublishStatePublished,
			))
		// 2. ensureLearningItemNodeInCourse lookup
		expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
		// 3. GetAdjacentPublishedLearningItems current lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, bID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				bID, courseID, nodeID, "Current B", models.LearningItemTypeArticle,
				nil, metadata, 1, now, models.LearningItemPublishStatePublished,
			))
		// 4. Fetch previous
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 1, 1, bID, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				aID, courseID, nodeID, "Prev A", models.LearningItemTypeArticle,
				nil, metadata, 0, now, models.LearningItemPublishStatePublished,
			))
		// 5. Fetch next
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 1, 1, bID, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				cID, courseID, nodeID, "Next C", models.LearningItemTypeArticle,
				nil, metadata, 2, now, models.LearningItemPublishStatePublished,
			))

		resp, payload := doJSON(t, env.app, http.MethodGet, base, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload = %v", resp.StatusCode, payload)
		}

		data, _ := payload["data"].(map[string]any)
		item, _ := data["learning_item"].(map[string]any)
		if item["id"] != bID.String() || item["title"] != "Current B" {
			t.Fatalf("unexpected current learning item: %v", item)
		}

		prev, _ := data["previous"].(map[string]any)
		if prev["id"] != aID.String() || prev["title"] != "Prev A" {
			t.Fatalf("unexpected previous navigation: %v", prev)
		}

		next, _ := data["next"].(map[string]any)
		if next["id"] != cID.String() || next["title"] != "Next C" {
			t.Fatalf("unexpected next navigation: %v", next)
		}

		// Ensure no draft state or position leaked in navigation DTO
		if _, hasPosition := prev["position"]; hasPosition {
			t.Fatalf("navigation DTO must exclude position: %v", prev)
		}
		if _, hasState := prev["publish_state"]; hasState {
			t.Fatalf("navigation DTO must exclude publish_state: %v", prev)
		}

		if err := env.mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("first item (previous is null)", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		base := learnerLearningItemsBase(courseID, nodeID) + "/" + aID.String()

		// 1. GetPublishedLearningItemByID lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, aID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				aID, courseID, nodeID, "First A", models.LearningItemTypeArticle,
				nil, metadata, 0, now, models.LearningItemPublishStatePublished,
			))
		// 2. ensureLearningItemNodeInCourse lookup
		expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
		// 3. GetAdjacentPublishedLearningItems current lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, aID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				aID, courseID, nodeID, "First A", models.LearningItemTypeArticle,
				nil, metadata, 0, now, models.LearningItemPublishStatePublished,
			))
		// 4. Fetch previous (empty)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, aID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		// 5. Fetch next
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, aID, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				bID, courseID, nodeID, "Second B", models.LearningItemTypeArticle,
				nil, metadata, 1, now, models.LearningItemPublishStatePublished,
			))

		resp, payload := doJSON(t, env.app, http.MethodGet, base, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload = %v", resp.StatusCode, payload)
		}

		data, _ := payload["data"].(map[string]any)
		if data["previous"] != nil {
			t.Fatalf("expected previous to be null, got: %v", data["previous"])
		}

		next, _ := data["next"].(map[string]any)
		if next["id"] != bID.String() || next["title"] != "Second B" {
			t.Fatalf("unexpected next navigation: %v", next)
		}
	})

	t.Run("last item (next is null)", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		base := learnerLearningItemsBase(courseID, nodeID) + "/" + cID.String()

		// 1. GetPublishedLearningItemByID lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, cID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				cID, courseID, nodeID, "Last C", models.LearningItemTypeArticle,
				nil, metadata, 2, now, models.LearningItemPublishStatePublished,
			))
		// 2. ensureLearningItemNodeInCourse lookup
		expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
		// 3. GetAdjacentPublishedLearningItems current lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, cID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				cID, courseID, nodeID, "Last C", models.LearningItemTypeArticle,
				nil, metadata, 2, now, models.LearningItemPublishStatePublished,
			))
		// 4. Fetch previous
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 2, 2, cID, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				bID, courseID, nodeID, "Second B", models.LearningItemTypeArticle,
				nil, metadata, 1, now, models.LearningItemPublishStatePublished,
			))
		// 5. Fetch next (empty)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 2, 2, cID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))

		resp, payload := doJSON(t, env.app, http.MethodGet, base, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload = %v", resp.StatusCode, payload)
		}

		data, _ := payload["data"].(map[string]any)
		if data["next"] != nil {
			t.Fatalf("expected next to be null, got: %v", data["next"])
		}

		prev, _ := data["previous"].(map[string]any)
		if prev["id"] != bID.String() || prev["title"] != "Second B" {
			t.Fatalf("unexpected previous navigation: %v", prev)
		}
	})

	t.Run("single item (both null)", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		base := learnerLearningItemsBase(courseID, nodeID) + "/" + aID.String()

		// 1. GetPublishedLearningItemByID lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, aID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				aID, courseID, nodeID, "Only", models.LearningItemTypeArticle,
				nil, metadata, 0, now, models.LearningItemPublishStatePublished,
			))
		// 2. ensureLearningItemNodeInCourse lookup
		expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
		// 3. GetAdjacentPublishedLearningItems current lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, aID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				aID, courseID, nodeID, "Only", models.LearningItemTypeArticle,
				nil, metadata, 0, now, models.LearningItemPublishStatePublished,
			))
		// 4. Fetch previous (empty)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, aID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		// 5. Fetch next (empty)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 0, 0, aID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))

		resp, payload := doJSON(t, env.app, http.MethodGet, base, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload = %v", resp.StatusCode, payload)
		}

		data, _ := payload["data"].(map[string]any)
		if data["previous"] != nil || data["next"] != nil {
			t.Fatalf("expected both previous and next to be null, got: %v", data)
		}
	})

	t.Run("skipped drafts", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		base := learnerLearningItemsBase(courseID, nodeID) + "/" + bID.String()

		// 1. GetPublishedLearningItemByID lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, bID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				bID, courseID, nodeID, "Current B", models.LearningItemTypeArticle,
				nil, metadata, 5, now, models.LearningItemPublishStatePublished,
			))
		// 2. ensureLearningItemNodeInCourse lookup
		expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
		// 3. GetAdjacentPublishedLearningItems current lookup
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, bID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				bID, courseID, nodeID, "Current B", models.LearningItemTypeArticle,
				nil, metadata, 5, now, models.LearningItemPublishStatePublished,
			))
		// 4. Fetch previous (skips intermediate drafts to find A)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 5, 5, bID, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				aID, courseID, nodeID, "Published A", models.LearningItemTypeArticle,
				nil, metadata, 1, now, models.LearningItemPublishStatePublished,
			))
		// 5. Fetch next (skips intermediate drafts to find C)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, models.LearningItemPublishStatePublished, 5, 5, bID, uint(1)).
			WillReturnRows(learningItemRowWithPublishState(
				cID, courseID, nodeID, "Published C", models.LearningItemTypeArticle,
				nil, metadata, 10, now, models.LearningItemPublishStatePublished,
			))

		resp, payload := doJSON(t, env.app, http.MethodGet, base, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload = %v", resp.StatusCode, payload)
		}

		data, _ := payload["data"].(map[string]any)
		prev, _ := data["previous"].(map[string]any)
		next, _ := data["next"].(map[string]any)

		if prev["id"] != aID.String() || prev["title"] != "Published A" {
			t.Fatalf("unexpected previous navigation (draft not skipped?): %v", prev)
		}
		if next["id"] != cID.String() || next["title"] != "Published C" {
			t.Fatalf("unexpected next navigation (draft not skipped?): %v", next)
		}
	})

	t.Run("current draft 404", func(t *testing.T) {
		env := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		expectLearnerCourseEnrollment(env.mock, nonAdminUser().ID, courseID, true)
		base := learnerLearningItemsBase(courseID, nodeID) + "/" + bID.String()

		// 1. GetPublishedLearningItemByID lookup (fails on draft item)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, bID, models.LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))

		resp, payload := doJSON(t, env.app, http.MethodGet, base, nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got status = %d payload = %v", resp.StatusCode, payload)
		}

		// Save response properties for draft error comparison
		draftStatus := resp.StatusCode
		draftData, _ := payload["data"].(string)

		// Verification of stable error contract
		if draftData != constants.ErrLearningItemNotFound {
			t.Fatalf("unexpected public message for draft: %q", draftData)
		}

		t.Run("current missing 404", func(t *testing.T) {
			env2 := newLearnerLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
			expectLearnerCourseEnrollment(env2.mock, nonAdminUser().ID, courseID, true)
			base2 := learnerLearningItemsBase(courseID, nodeID) + "/" + uuid.New().String()

			env2.mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "course_id", "course_node_id", "title", "item_type",
					"description", "metadata", "position", "publish_state", "created_at", "updated_at",
				}))

			resp2, payload2 := doJSON(t, env2.app, http.MethodGet, base2, nil)
			if resp2.StatusCode != draftStatus {
				t.Fatalf("mismatched status codes between draft and missing")
			}
			missingData, _ := payload2["data"].(string)
			if missingData != draftData {
				t.Fatalf("responses for draft and missing must be indistinguishable; got %q vs %q", draftData, missingData)
			}
		})
	})
}
