package models

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

func TestProjectPublishedLearningChain(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")

	lowerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	higherID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	t.Run("nil course ID", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		_, err := model.ProjectPublishedLearningChain(uuid.Nil, nodeID)
		if !errors.Is(err, ErrCourseNotFound) {
			t.Fatalf("expected ErrCourseNotFound, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("nil node ID", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		_, err := model.ProjectPublishedLearningChain(courseID, uuid.Nil)
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("expected ErrLearningItemNodeNotFound, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("node not found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		// First query check for course_id and node_id matches nothing
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		// Second check for node_id matches nothing
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))

		_, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("expected ErrLearningItemNodeNotFound, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("node belongs to another course", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		wrongCourseID := uuid.New()
		// First query check for course_id and node_id matches nothing
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		// Second check for node_id matches the wrong course node
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(nodeID, wrongCourseID))

		_, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if !errors.Is(err, ErrLearningItemCrossCourse) {
			t.Fatalf("expected ErrLearningItemCrossCourse, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("validation-query database failure", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		dbErr := errors.New("db error")
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnError(dbErr)

		_, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if err == nil || !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("expected ErrLearningItemPersistence wrapping db error, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("projection-query database failure", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)

		dbErr := errors.New("query error")
		mock.ExpectQuery(`SELECT "id", "title" FROM "learning_items"`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnError(dbErr)

		_, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if err == nil || !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("expected ErrLearningItemPersistence, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("empty node", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)

		mock.ExpectQuery(`SELECT "id", "title" FROM "learning_items"`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title"}))

		result, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil empty slice, got nil")
		}
		if len(result) != 0 {
			t.Fatalf("expected empty slice, got length %d", len(result))
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("one published item", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)

		rows := sqlmock.NewRows([]string{"id", "title"}).
			AddRow(lowerID, "Item 1")

		mock.ExpectQuery(`SELECT "id", "title" FROM "learning_items"`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(rows)

		result, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result))
		}
		if result[0].ID != lowerID || result[0].Title != "Item 1" {
			t.Fatalf("unexpected item value: %v", result[0])
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("many published items and gaps in positions", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)

		// Simulating gaps in positions (e.g. 5 and 12)
		rows := sqlmock.NewRows([]string{"id", "title"}).
			AddRow(lowerID, "Item 1").
			AddRow(higherID, "Item 2")

		mock.ExpectQuery(`SELECT "id", "title" FROM "learning_items"`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(rows)

		result, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 items, got %d", len(result))
		}
		if result[0].ID != lowerID || result[1].ID != higherID {
			t.Fatalf("unexpected items layout: %v", result)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("drafts excluded by SQL filter", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)

		// The mock database should only return the published items
		// The SQL query itself filters by publish_state = 'PUBLISHED'
		rows := sqlmock.NewRows([]string{"id", "title"}).
			AddRow(lowerID, "Published 1").
			AddRow(higherID, "Published 2")

		mock.ExpectQuery(`SELECT "id", "title" FROM "learning_items"`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(rows)

		result, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 published items, got %d", len(result))
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("duplicate positions ordered by ID", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)

		rows := sqlmock.NewRows([]string{"id", "title"}).
			AddRow(lowerID, "Item 1").
			AddRow(higherID, "Item 2")

		mock.ExpectQuery(`SELECT "id", "title" FROM "learning_items"`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(rows)

		result, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 items, got %d", len(result))
		}
		// Confirm lower ID comes first
		if result[0].ID != lowerID || result[1].ID != higherID {
			t.Fatalf("duplicate positions order mismatch: %v", result)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})

	t.Run("SQL boundary assertions", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)

		mock.ExpectQuery(`SELECT "id", "title" FROM "learning_items"`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title"}))

		_, err := model.ProjectPublishedLearningChain(courseID, nodeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Now generate the SQL query string using goqu to check boundaries
		sqlStr, _, err := model.db.From(LearningItemsTable).
			Select("id", "title").
			Where(goqu.Ex{
				"course_id":      courseID,
				"course_node_id": nodeID,
				"publish_state":  LearningItemPublishStatePublished,
			}).
			Order(goqu.I("position").Asc(), goqu.I("id").Asc()).
			Prepared(true).
			ToSQL()
		if err != nil {
			t.Fatalf("failed to generate SQL: %v", err)
		}

		// Perform assertions on sqlStr
		// Check that it selects ONLY id and title
		if !strings.Contains(sqlStr, `SELECT "id", "title" FROM "learning_items"`) {
			t.Fatalf("unexpected SELECT clause: %s", sqlStr)
		}
		// Check that it contains WHERE clause filters
		if !strings.Contains(sqlStr, `"course_id" = ?`) && !strings.Contains(sqlStr, `"course_id" = $1`) {
			t.Fatalf("missing course_id filter: %s", sqlStr)
		}
		if !strings.Contains(sqlStr, `"course_node_id" = ?`) && !strings.Contains(sqlStr, `"course_node_id" = $2`) {
			t.Fatalf("missing course_node_id filter: %s", sqlStr)
		}
		if !strings.Contains(sqlStr, `"publish_state" = ?`) && !strings.Contains(sqlStr, `"publish_state" = $3`) {
			t.Fatalf("missing publish_state filter: %s", sqlStr)
		}
		// Check ordering
		if !strings.Contains(sqlStr, `ORDER BY "position" ASC, "id" ASC`) {
			t.Fatalf("missing ORDER BY position ASC, id ASC clause: %s", sqlStr)
		}

		// Prohibited keywords checks
		prohibited := []string{
			"metadata",
			"description",
			"child_node_ids",
			"parent_id",
			"path",
			"depth",
			"WITH RECURSIVE",
			"MAX_DEPTH",
		}
		for _, term := range prohibited {
			if matched, _ := regexp.MatchString(`(?i)\b`+regexp.QuoteMeta(term)+`\b`, sqlStr); matched {
				t.Fatalf("SQL query violated boundaries by including prohibited term: %q", term)
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL: %v", err)
		}
	})
}
