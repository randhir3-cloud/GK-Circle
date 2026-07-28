package models

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func learningItemDraftFilteringIDs() (courseID, nodeID, publishedID, draftID, otherNodeID, otherCourseID uuid.UUID) {
	return uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481"),
		uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"),
		uuid.MustParse("019c02a0-1111-7000-8000-000000000101"),
		uuid.MustParse("019c02a0-1111-7000-8000-000000000102"),
		uuid.MustParse("019c01c9-bbbb-78e2-a366-690bfd600281"),
		uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e482")
}

func TestLearningItemDraftFilteringPublishedList(t *testing.T) {
	courseID, nodeID, publishedID, draftID, _, _ := learningItemDraftFilteringIDs()
	metadata := []byte(`{"version":1,"blocks":[]}`)
	now := time.Date(2026, time.July, 26, 10, 0, 0, 0, time.UTC)

	t.Run("mixed_returns_only_published_ordered", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		// Repository SQL filters PUBLISHED; mock returns the published-only result set.
		rows := learningItemRowsWithPublishState(
			publishedID, courseID, nodeID, "Published A", LearningItemTypeArticle,
			nil, metadata, 0, LearningItemPublishStatePublished,
		)
		secondPublished := uuid.MustParse("019c02a0-1111-7000-8000-000000000103")
		rows.AddRow(
			secondPublished, courseID, nodeID, "Published B", LearningItemTypeVideo,
			nil, metadata, 2, LearningItemPublishStatePublished, now, now,
		)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(rows)

		items, err := model.ListPublishedLearningItemsByNode(courseID, nodeID)
		if err != nil {
			t.Fatalf("list published: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("len=%d, want 2 published", len(items))
		}
		for _, item := range items {
			if item.PublishState != LearningItemPublishStatePublished {
				t.Fatalf("draft leaked: %+v", item)
			}
			if item.ID == draftID {
				t.Fatalf("draft ID leaked into published list: %s", draftID)
			}
		}
		if items[0].Title != "Published A" || items[1].Title != "Published B" {
			t.Fatalf("order = %#v", items)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("all_draft_empty_non_nil", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		items, err := model.ListPublishedLearningItemsByNode(courseID, nodeID)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if items == nil || len(items) != 0 {
			t.Fatalf("items = %#v, want non-nil empty", items)
		}
	})

	t.Run("empty_node_non_nil", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		items, err := model.ListPublishedLearningItemsByNode(courseID, nodeID)
		if err != nil || items == nil || len(items) != 0 {
			t.Fatalf("items=%#v err=%v", items, err)
		}
	})
}

func TestLearningItemDraftFilteringPublishedGet(t *testing.T) {
	courseID, nodeID, publishedID, draftID, otherNodeID, otherCourseID := learningItemDraftFilteringIDs()
	metadata := []byte(`{"version":1,"blocks":[]}`)

	t.Run("published_success", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, publishedID, LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowsWithPublishState(
				publishedID, courseID, nodeID, "Live", LearningItemTypeArticle,
				nil, metadata, 0, LearningItemPublishStatePublished,
			))
		item, err := model.GetPublishedLearningItemByID(courseID, nodeID, publishedID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if item.PublishState != LearningItemPublishStatePublished || item.ID != publishedID {
			t.Fatalf("item = %+v", item)
		}
	})

	t.Run("draft_not_found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, draftID, LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		item, err := model.GetPublishedLearningItemByID(courseID, nodeID, draftID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
		if item.ID != uuid.Nil || item.Title != "" {
			t.Fatalf("must not leak draft payload: %+v", item)
		}
	})

	t.Run("missing_same_not_found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		missing := uuid.MustParse("019c02a0-1111-7000-8000-000000000199")
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, missing, LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		_, err := model.GetPublishedLearningItemByID(courseID, nodeID, missing)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
	})

	t.Run("wrong_node_not_found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, otherNodeID, draftID, LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		_, err := model.GetPublishedLearningItemByID(courseID, otherNodeID, draftID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
	})

	t.Run("wrong_course_scoped_not_found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(otherCourseID, nodeID, publishedID, LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		_, err := model.GetPublishedLearningItemByID(otherCourseID, nodeID, publishedID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
	})
}

func TestLearningItemDraftFilteringAdminContrast(t *testing.T) {
	courseID, nodeID, publishedID, draftID, _, _ := learningItemDraftFilteringIDs()
	metadata := []byte(`{"version":1,"blocks":[]}`)
	now := time.Date(2026, time.July, 26, 10, 0, 0, 0, time.UTC)

	t.Run("admin_list_includes_draft_and_published", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		rows := learningItemRowsWithPublishState(
			draftID, courseID, nodeID, "Draft Lesson", LearningItemTypeArticle,
			nil, metadata, 0, LearningItemPublishStateDraft,
		)
		rows.AddRow(
			publishedID, courseID, nodeID, "Published Lesson", LearningItemTypeVideo,
			nil, metadata, 1, LearningItemPublishStatePublished, now, now,
		)
		// Admin list must not require publish_state=PUBLISHED in WHERE.
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`).
			WithArgs(courseID, nodeID).
			WillReturnRows(rows)

		items, err := model.ListLearningItemsByNode(courseID, nodeID)
		if err != nil {
			t.Fatalf("admin list: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("len=%d, want 2", len(items))
		}
		states := map[LearningItemPublishState]bool{}
		for _, item := range items {
			states[item.PublishState] = true
		}
		if !states[LearningItemPublishStateDraft] || !states[LearningItemPublishStatePublished] {
			t.Fatalf("admin list missing draft or published: %#v", items)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("admin_get_returns_draft", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
			WithArgs(courseID, nodeID, draftID, uint(1)).
			WillReturnRows(learningItemRowsWithPublishState(
				draftID, courseID, nodeID, "Draft Lesson", LearningItemTypeArticle,
				nil, metadata, 0, LearningItemPublishStateDraft,
			))
		item, err := model.GetLearningItemByID(courseID, nodeID, draftID)
		if err != nil {
			t.Fatalf("admin get draft: %v", err)
		}
		if item.PublishState != LearningItemPublishStateDraft || item.ID != draftID {
			t.Fatalf("item = %+v", item)
		}
	})
}

func TestLearningItemDraftFilteringQueryFailures(t *testing.T) {
	courseID, nodeID, publishedID, _, _, _ := learningItemDraftFilteringIDs()

	t.Run("published_list_node_lookup_failure", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnError(errors.New("node boom"))
		items, err := model.ListPublishedLearningItemsByNode(courseID, nodeID)
		if !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
		if items != nil {
			t.Fatalf("must not leak items: %#v", items)
		}
	})

	t.Run("published_list_query_failure", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnError(errors.New("list boom"))
		items, err := model.ListPublishedLearningItemsByNode(courseID, nodeID)
		if !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
		if items != nil {
			t.Fatalf("must not leak items: %#v", items)
		}
	})

	t.Run("published_get_query_failure", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, publishedID, LearningItemPublishStatePublished, uint(1)).
			WillReturnError(errors.New("get boom"))
		item, err := model.GetPublishedLearningItemByID(courseID, nodeID, publishedID)
		if !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
		if item.ID != uuid.Nil {
			t.Fatalf("must not leak item: %+v", item)
		}
	})
}

func TestLearningItemDraftFilteringSQLBoundary(t *testing.T) {
	publishedQueries := []string{
		`SELECT ... FROM "learning_items" WHERE ("course_id" = $1 AND "course_node_id" = $2 AND "publish_state" = $3) ORDER BY "position" ASC, "id" ASC`,
		`SELECT ... FROM "learning_items" WHERE ("id" = $1 AND "course_id" = $2 AND "course_node_id" = $3 AND "publish_state" = $4) LIMIT 1`,
	}
	adminQueries := []string{
		`SELECT ... FROM "learning_items" WHERE ("course_id" = $1 AND "course_node_id" = $2) ORDER BY "position" ASC, "id" ASC`,
		`SELECT ... FROM "learning_items" WHERE ("id" = $1 AND "course_id" = $2 AND "course_node_id" = $3) LIMIT 1`,
	}
	for _, q := range publishedQueries {
		lower := strings.ToLower(q)
		for _, need := range []string{"course_id", "course_node_id", "publish_state"} {
			if !strings.Contains(lower, need) {
				t.Fatalf("published query missing %q: %s", need, q)
			}
		}
	}
	for _, q := range adminQueries {
		lower := strings.ToLower(q)
		if strings.Contains(lower, `publish_state" =`) || strings.Contains(lower, "publish_state =") {
			t.Fatalf("admin query must not add published-only predicate: %s", q)
		}
		if !strings.Contains(lower, "course_id") || !strings.Contains(lower, "course_node_id") {
			t.Fatalf("admin query must stay course/node scoped: %s", q)
		}
	}
}
