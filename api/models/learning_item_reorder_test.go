package models

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func learningItemReorderTestIDs() (courseID, nodeID, aID, bID, cID, foreignID uuid.UUID) {
	return uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481"),
		uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"),
		uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6"),
		uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a"),
		uuid.MustParse("019c01cb-7f39-7f26-900f-6947e75e7284"),
		uuid.MustParse("019c01cc-1111-7222-8333-944455556666")
}

func expectLearningItemReorderCourseLock(mock sqlmock.Sqlmock, courseID uuid.UUID) {
	mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
}

func expectLearningItemReorderSiblingLock(mock sqlmock.Sqlmock, courseID, nodeID uuid.UUID, rows *sqlmock.Rows) {
	mock.ExpectQuery(`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items".*FOR UPDATE`).
		WithArgs(courseID, nodeID).
		WillReturnRows(rows)
}

func expectLearningItemReorderPositionUpdate(mock sqlmock.Sqlmock, courseID, nodeID, itemID uuid.UUID, position int) {
	mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
		WithArgs(position, courseID, nodeID, itemID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(itemID))
}

func learningItemSiblingRows(courseID, nodeID uuid.UUID, items ...struct {
	id       uuid.UUID
	position int
}) *sqlmock.Rows {
	now := time.Date(2026, time.July, 26, 8, 30, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "course_id", "course_node_id", "title", "item_type",
		"description", "metadata", "position", "publish_state", "created_at", "updated_at",
	})
	for _, item := range items {
		rows.AddRow(
			item.id, courseID, nodeID, "Item", LearningItemTypeArticle, nil, []byte(`null`),
			item.position, LearningItemPublishStateDraft, now, now,
		)
	}
	return rows
}

func TestVerifyLearningItemReorderUpdatedIDs(t *testing.T) {
	a := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	b := uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a")
	c := uuid.MustParse("019c01cb-7f39-7f26-900f-6947e75e7284")
	expected := map[uuid.UUID]struct{}{a: {}, b: {}}

	if err := verifyLearningItemReorderUpdatedIDs(expected, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("exact match: %v", err)
	}
	if err := verifyLearningItemReorderUpdatedIDs(expected, []uuid.UUID{a}); !errors.Is(err, ErrLearningItemReorderConflict) {
		t.Fatalf("count mismatch: got %v", err)
	}
	if err := verifyLearningItemReorderUpdatedIDs(expected, []uuid.UUID{a, c}); !errors.Is(err, ErrLearningItemReorderConflict) {
		t.Fatalf("foreign id: got %v", err)
	}
	if err := verifyLearningItemReorderUpdatedIDs(expected, []uuid.UUID{a, a}); !errors.Is(err, ErrLearningItemReorderConflict) {
		t.Fatalf("duplicate returned id: got %v", err)
	}
}

func TestReorderLearningItemsValidation(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	_, nodeID, aID, bID, _, _ := learningItemReorderTestIDs()
	tests := []struct {
		name     string
		courseID uuid.UUID
		nodeID   uuid.UUID
		ordered  []uuid.UUID
		want     error
	}{
		{"course required", uuid.Nil, nodeID, nil, ErrCourseNotFound},
		{"node required", uuid.New(), uuid.Nil, []uuid.UUID{aID}, ErrLearningItemNodeNotFound},
		{"nil ordered id", uuid.New(), nodeID, []uuid.UUID{uuid.Nil}, ErrLearningItemNotFound},
		{"duplicate ordered id", uuid.New(), nodeID, []uuid.UUID{aID, aID}, ErrLearningItemReorderDuplicate},
		{"duplicate among three", uuid.New(), nodeID, []uuid.UUID{aID, bID, aID}, ErrLearningItemReorderDuplicate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := model.ReorderLearningItems(test.courseID, test.nodeID, test.ordered)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation unexpectedly queried the database: %v", err)
	}
}

func TestReorderLearningItemsSiblingSetMismatch(t *testing.T) {
	courseID, nodeID, aID, bID, cID, foreignID := learningItemReorderTestIDs()
	rows := learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
		struct {
			id       uuid.UUID
			position int
		}{bID, 1},
		struct {
			id       uuid.UUID
			position int
		}{cID, 2},
	)

	t.Run("missing id", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		expectLearningItemNodeLock(mock, courseID, nodeID, true)
		expectLearningItemReorderSiblingLock(mock, courseID, nodeID, rows)
		mock.ExpectRollback()
		_, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{aID, bID})
		if !errors.Is(err, ErrLearningItemReorderMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemReorderMismatch)
		}
	})
	t.Run("extra foreign id", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		expectLearningItemNodeLock(mock, courseID, nodeID, true)
		expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
			struct {
				id       uuid.UUID
				position int
			}{aID, 0},
			struct {
				id       uuid.UUID
				position int
			}{bID, 1},
		))
		mock.ExpectRollback()
		_, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{aID, bID, foreignID})
		if !errors.Is(err, ErrLearningItemReorderMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemReorderMismatch)
		}
	})
	t.Run("foreign node id replaces sibling", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		expectLearningItemNodeLock(mock, courseID, nodeID, true)
		expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
			struct {
				id       uuid.UUID
				position int
			}{aID, 0},
			struct {
				id       uuid.UUID
				position int
			}{bID, 1},
		))
		mock.ExpectRollback()
		_, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{aID, foreignID})
		if !errors.Is(err, ErrLearningItemReorderMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemReorderMismatch)
		}
	})
}

func TestReorderLearningItemsZeroItemNode(t *testing.T) {
	courseID, nodeID, aID, _, _, _ := learningItemReorderTestIDs()
	t.Run("empty success", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		expectLearningItemNodeLock(mock, courseID, nodeID, true)
		expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID))
		mock.ExpectCommit()
		result, err := model.ReorderLearningItems(courseID, nodeID, nil)
		if err != nil {
			t.Fatalf("empty reorder: %v", err)
		}
		if !result.Noop || result.LearningItemCount != 0 || result.PositionsUpdated != 0 {
			t.Fatalf("result = %+v", result)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("empty with ordered ids", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		expectLearningItemNodeLock(mock, courseID, nodeID, true)
		expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID))
		mock.ExpectRollback()
		_, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{aID})
		if !errors.Is(err, ErrLearningItemReorderMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemReorderMismatch)
		}
	})
}

func TestReorderLearningItemsSingleItemNoop(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID, nodeID, aID, _, _, _ := learningItemReorderTestIDs()
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
	))
	mock.ExpectCommit()
	result, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{aID})
	if err != nil {
		t.Fatalf("single item no-op: %v", err)
	}
	if !result.Noop || result.PositionsUpdated != 0 || result.LearningItemCount != 1 {
		t.Fatalf("result = %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no-op must not issue UPDATE statements: %v", err)
	}
}

func TestReorderLearningItemsTwoItemSwap(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID, nodeID, aID, bID, _, _ := learningItemReorderTestIDs()
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
		struct {
			id       uuid.UUID
			position int
		}{bID, 1},
	))
	// maxExisting=1, count=2, temporaryBase=4
	expectLearningItemReorderPositionUpdate(mock, courseID, nodeID, aID, 4)
	expectLearningItemReorderPositionUpdate(mock, courseID, nodeID, bID, 5)
	expectLearningItemReorderPositionUpdate(mock, courseID, nodeID, aID, 1)
	expectLearningItemReorderPositionUpdate(mock, courseID, nodeID, bID, 0)
	mock.ExpectCommit()

	result, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{bID, aID})
	if err != nil {
		t.Fatalf("two-item swap: %v", err)
	}
	if result.Noop || result.LearningItemCount != 2 || result.PositionsUpdated != 2 {
		t.Fatalf("result = %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReorderLearningItemsAlreadyCanonicalNoop(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID, nodeID, aID, bID, _, _ := learningItemReorderTestIDs()
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
		struct {
			id       uuid.UUID
			position int
		}{bID, 1},
	))
	mock.ExpectCommit()
	result, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{aID, bID})
	if err != nil {
		t.Fatalf("canonical no-op: %v", err)
	}
	if !result.Noop || result.PositionsUpdated != 0 {
		t.Fatalf("result = %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("canonical no-op must not UPDATE: %v", err)
	}
}

func TestReorderLearningItemsTwiceSecondNoop(t *testing.T) {
	courseID, nodeID, aID, bID, _, _ := learningItemReorderTestIDs()

	model, mock := newLearningItemModelTest(t)
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
		struct {
			id       uuid.UUID
			position int
		}{bID, 1},
	))
	expectLearningItemReorderPositionUpdate(mock, courseID, nodeID, aID, 4)
	expectLearningItemReorderPositionUpdate(mock, courseID, nodeID, bID, 5)
	expectLearningItemReorderPositionUpdate(mock, courseID, nodeID, aID, 1)
	expectLearningItemReorderPositionUpdate(mock, courseID, nodeID, bID, 0)
	mock.ExpectCommit()
	first, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{bID, aID})
	if err != nil {
		t.Fatalf("first reorder: %v", err)
	}
	if first.Noop {
		t.Fatal("first reorder should mutate")
	}

	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{bID, 0},
		struct {
			id       uuid.UUID
			position int
		}{aID, 1},
	))
	mock.ExpectCommit()
	second, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{bID, aID})
	if err != nil {
		t.Fatalf("second reorder: %v", err)
	}
	if !second.Noop || second.PositionsUpdated != 0 {
		t.Fatalf("second reorder result = %+v", second)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReorderLearningItemsConcurrentConflict(t *testing.T) {
	// Losing transaction: row lock serialized, but UPDATE finds no matching row.
	model, mock := newLearningItemModelTest(t)
	courseID, nodeID, aID, bID, _, _ := learningItemReorderTestIDs()
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
		struct {
			id       uuid.UUID
			position int
		}{bID, 1},
	))
	mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
		WithArgs(4, courseID, nodeID, aID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	_, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{bID, aID})
	if !errors.Is(err, ErrLearningItemReorderConflict) {
		t.Fatalf("error = %v, want %v", err, ErrLearningItemReorderConflict)
	}
}

func TestReorderLearningItemsUniqueViolationConflict(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID, nodeID, aID, bID, _, _ := learningItemReorderTestIDs()
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
		struct {
			id       uuid.UUID
			position int
		}{bID, 1},
	))
	mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
		WithArgs(4, courseID, nodeID, aID).
		WillReturnError(&pq.Error{Code: "23505", Constraint: learningItemsNodePositionConstraint})
	mock.ExpectRollback()
	_, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{bID, aID})
	if !errors.Is(err, ErrLearningItemReorderConflict) {
		t.Fatalf("error = %v, want %v", err, ErrLearningItemReorderConflict)
	}
}

func TestReorderLearningItemsRollbackOnInjectedFailure(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID, nodeID, aID, bID, _, _ := learningItemReorderTestIDs()
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
		struct {
			id       uuid.UUID
			position int
		}{bID, 1},
	))
	mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
		WithArgs(4, courseID, nodeID, aID).
		WillReturnError(errors.New("update failure"))
	mock.ExpectRollback()
	_, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{bID, aID})
	if !errors.Is(err, ErrLearningItemPersistence) {
		t.Fatalf("error = %v, want persistence wrapper", err)
	}
}

func TestReorderLearningItemsPositionOverflow(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID, nodeID, aID, _, _, _ := learningItemReorderTestIDs()
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, nodeID, learningItemSiblingRows(courseID, nodeID,
		struct {
			id       uuid.UUID
			position int
		}{aID, math.MaxInt32},
	))
	mock.ExpectRollback()
	_, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{aID})
	if !errors.Is(err, ErrLearningItemPositionInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrLearningItemPositionInvalid)
	}
}

func TestReorderLearningItemsMissingNode(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID, nodeID, aID, _, _, _ := learningItemReorderTestIDs()
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemNodeLock(mock, courseID, nodeID, false)
	mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
		WithArgs(nodeID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
	mock.ExpectRollback()
	_, err := model.ReorderLearningItems(courseID, nodeID, []uuid.UUID{aID})
	if !errors.Is(err, ErrLearningItemNodeNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrLearningItemNodeNotFound)
	}
}
