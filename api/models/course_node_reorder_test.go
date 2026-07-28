package models

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func reorderTestIDs() (courseID, parentID, aID, bID, cID, foreignID uuid.UUID) {
	return uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481"),
		uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"),
		uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6"),
		uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a"),
		uuid.MustParse("019c01cb-7f39-7f26-900f-6947e75e7284"),
		uuid.MustParse("019c01cc-1111-7222-8333-944455556666")
}

func expectReorderCourseLock(mock sqlmock.Sqlmock, courseID uuid.UUID) {
	mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
}

func expectReorderParentLock(mock sqlmock.Sqlmock, courseID, parentID uuid.UUID) {
	mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
		WithArgs(parentID, uint(1)).
		WillReturnRows(courseNodeRows(parentID, courseID, nil, SECTION, "Parent", 0, parentID.String()))
}

func expectReorderSiblingLockRoot(mock sqlmock.Sqlmock, courseID uuid.UUID, rows *sqlmock.Rows) {
	mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
		WithArgs(courseID).
		WillReturnRows(rows)
}

func expectReorderSiblingLockChild(mock sqlmock.Sqlmock, courseID, parentID uuid.UUID, rows *sqlmock.Rows) {
	mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
		WithArgs(courseID, parentID).
		WillReturnRows(rows)
}

func expectReorderPositionUpdate(mock sqlmock.Sqlmock, courseID, nodeID uuid.UUID, position int, _ bool) {
	mock.ExpectQuery(`UPDATE "course_nodes".*RETURNING`).
		WithArgs(position, courseID, nodeID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(nodeID))
}

func siblingRows(courseID uuid.UUID, parentID interface{}, nodes ...struct {
	id       uuid.UUID
	position int
}) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		"id", "course_id", "parent_id", "node_type", "title", "position",
		"path", "status", "created_at", "updated_at",
	})
	for _, node := range nodes {
		path := node.id.String()
		if parent, ok := parentID.(uuid.UUID); ok {
			path = parent.String() + "/" + node.id.String()
		}
		rows.AddRow(
			node.id, courseID, parentID, SECTION, "Node", node.position, path,
			CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime(),
		)
	}
	return rows
}

func TestVerifyReorderUpdatedIDs(t *testing.T) {
	a := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	b := uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a")
	c := uuid.MustParse("019c01cb-7f39-7f26-900f-6947e75e7284")
	expected := map[uuid.UUID]struct{}{a: {}, b: {}}

	if err := verifyReorderUpdatedIDs(expected, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("exact match: %v", err)
	}
	if err := verifyReorderUpdatedIDs(expected, []uuid.UUID{a}); !errors.Is(err, ErrCourseNodeReorderConflict) {
		t.Fatalf("count mismatch: got %v", err)
	}
	if err := verifyReorderUpdatedIDs(expected, []uuid.UUID{a, c}); !errors.Is(err, ErrCourseNodeReorderConflict) {
		t.Fatalf("foreign id: got %v", err)
	}
	if err := verifyReorderUpdatedIDs(expected, []uuid.UUID{a, a}); !errors.Is(err, ErrCourseNodeReorderConflict) {
		t.Fatalf("duplicate returned id: got %v", err)
	}
	// Duplicate returns with equal map length must still fail.
	expectedThree := map[uuid.UUID]struct{}{a: {}, b: {}, c: {}}
	if err := verifyReorderUpdatedIDs(expectedThree, []uuid.UUID{a, a, b}); !errors.Is(err, ErrCourseNodeReorderConflict) {
		t.Fatalf("duplicate hiding missing expected: got %v", err)
	}
}

func TestReorderChildrenValidation(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	_, parentID, aID, bID, _, _ := reorderTestIDs()
	tests := []struct {
		name     string
		courseID uuid.UUID
		parentID uuid.NullUUID
		ordered  []uuid.UUID
		want     error
	}{
		{"course required", uuid.Nil, uuid.NullUUID{}, nil, ErrCourseNodeCourseRequired},
		{"nil parent uuid", uuid.New(), uuid.NullUUID{Valid: true}, []uuid.UUID{aID}, ErrCourseNodeParentNotFound},
		{"nil ordered id", uuid.New(), uuid.NullUUID{}, []uuid.UUID{uuid.Nil}, ErrCourseNodeNotFound},
		{"duplicate ordered id", uuid.New(), uuid.NullUUID{UUID: parentID, Valid: true}, []uuid.UUID{aID, aID}, ErrCourseNodeReorderDuplicate},
		{"duplicate among three", uuid.New(), uuid.NullUUID{}, []uuid.UUID{aID, bID, aID}, ErrCourseNodeReorderDuplicate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := model.ReorderChildren(test.courseID, test.parentID, test.ordered)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation unexpectedly queried the database: %v", err)
	}
}

func TestReorderChildrenSiblingSetMismatch(t *testing.T) {
	courseID, _, aID, bID, cID, foreignID := reorderTestIDs()
	rows := siblingRows(courseID, nil,
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

	t.Run("missing entry", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectReorderCourseLock(mock, courseID)
		expectReorderSiblingLockRoot(mock, courseID, rows)
		mock.ExpectRollback()
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID, bID})
		if !errors.Is(err, ErrCourseNodeReorderMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeReorderMismatch)
		}
	})
	t.Run("extra foreign id", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectReorderCourseLock(mock, courseID)
		expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil,
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
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID, bID, foreignID})
		if !errors.Is(err, ErrCourseNodeReorderMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeReorderMismatch)
		}
	})
	t.Run("external id replaces sibling", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectReorderCourseLock(mock, courseID)
		expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil,
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
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID, foreignID})
		if !errors.Is(err, ErrCourseNodeReorderMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeReorderMismatch)
		}
	})
}

func TestReorderChildrenEmptySiblingSet(t *testing.T) {
	courseID, parentID, aID, _, _, _ := reorderTestIDs()
	t.Run("empty success", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectReorderCourseLock(mock, courseID)
		expectReorderParentLock(mock, courseID, parentID)
		expectReorderSiblingLockChild(mock, courseID, parentID, siblingRows(courseID, parentID))
		mock.ExpectCommit()
		if err := model.ReorderChildren(courseID, uuid.NullUUID{UUID: parentID, Valid: true}, nil); err != nil {
			t.Fatalf("empty reorder: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("empty with ordered ids", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectReorderCourseLock(mock, courseID)
		expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil))
		mock.ExpectRollback()
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID})
		if !errors.Is(err, ErrCourseNodeReorderMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeReorderMismatch)
		}
	})
	t.Run("empty commit failure", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectReorderCourseLock(mock, courseID)
		expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil))
		mock.ExpectCommit().WillReturnError(errors.New("commit failure"))
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{})
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
	})
}

func TestReorderChildrenSingleSiblingNoop(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID, _, aID, _, _, _ := reorderTestIDs()
	mock.ExpectBegin()
	expectReorderCourseLock(mock, courseID)
	expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
	))
	mock.ExpectCommit()
	if err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID}); err != nil {
		t.Fatalf("single sibling no-op: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no-op must not issue UPDATE statements: %v", err)
	}
}

func TestReorderChildrenTwoNodeSwap(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID, _, aID, bID, _, _ := reorderTestIDs()
	// aID < bID lexicographically; locked in id ASC.
	mock.ExpectBegin()
	expectReorderCourseLock(mock, courseID)
	expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil,
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
	expectReorderPositionUpdate(mock, courseID, aID, 4, false)
	expectReorderPositionUpdate(mock, courseID, bID, 5, false)
	expectReorderPositionUpdate(mock, courseID, aID, 1, true)
	expectReorderPositionUpdate(mock, courseID, bID, 0, true)
	mock.ExpectCommit()

	if err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{bID, aID}); err != nil {
		t.Fatalf("two-node swap: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReorderChildrenThreeNodeRotation(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID, parentID, aID, bID, cID, _ := reorderTestIDs()
	mock.ExpectBegin()
	expectReorderCourseLock(mock, courseID)
	expectReorderParentLock(mock, courseID, parentID)
	expectReorderSiblingLockChild(mock, courseID, parentID, siblingRows(courseID, parentID,
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
	))
	// maxExisting=2, count=3, temporaryBase=6
	expectReorderPositionUpdate(mock, courseID, aID, 6, false)
	expectReorderPositionUpdate(mock, courseID, bID, 7, false)
	expectReorderPositionUpdate(mock, courseID, cID, 8, false)
	// ordered [b,c,a] => b:0, c:1, a:2; phase 2 still id ASC
	expectReorderPositionUpdate(mock, courseID, aID, 2, true)
	expectReorderPositionUpdate(mock, courseID, bID, 0, true)
	expectReorderPositionUpdate(mock, courseID, cID, 1, true)
	mock.ExpectCommit()

	err := model.ReorderChildren(
		courseID,
		uuid.NullUUID{UUID: parentID, Valid: true},
		[]uuid.UUID{bID, cID, aID},
	)
	if err != nil {
		t.Fatalf("three-node rotation: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReorderChildrenAlreadyCanonicalNoop(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID, _, aID, bID, _, _ := reorderTestIDs()
	mock.ExpectBegin()
	expectReorderCourseLock(mock, courseID)
	expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil,
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
	if err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID, bID}); err != nil {
		t.Fatalf("canonical no-op: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("canonical no-op must not UPDATE: %v", err)
	}
}

func TestReorderChildrenPositionOverflow(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID, _, aID, _, _, _ := reorderTestIDs()
	mock.ExpectBegin()
	expectReorderCourseLock(mock, courseID)
	expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil,
		struct {
			id       uuid.UUID
			position int
		}{aID, math.MaxInt32},
	))
	// Positions do not match canonical index 0, so staging is attempted and overflows.
	mock.ExpectRollback()
	err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID})
	if !errors.Is(err, ErrCourseNodePositionInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrCourseNodePositionInvalid)
	}
}

func TestReorderChildrenUpdateConflict(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID, _, aID, bID, _, _ := reorderTestIDs()
	mock.ExpectBegin()
	expectReorderCourseLock(mock, courseID)
	expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil,
		struct {
			id       uuid.UUID
			position int
		}{aID, 0},
		struct {
			id       uuid.UUID
			position int
		}{bID, 1},
	))
	mock.ExpectQuery(`UPDATE "course_nodes".*RETURNING`).
		WithArgs(4, courseID, aID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{bID, aID})
	if !errors.Is(err, ErrCourseNodeReorderConflict) {
		t.Fatalf("error = %v, want %v", err, ErrCourseNodeReorderConflict)
	}
}

func TestReorderChildrenLargeSiblingSet(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	const n = 100
	ids := make([]uuid.UUID, n)
	for i := 0; i < n; i++ {
		ids[i] = uuid.MustParse(fmt.Sprintf("019c01d0-%04x-7000-8000-000000000000", i))
	}
	// Ensure id ASC equals construction order for these UUIDs.
	rows := sqlmock.NewRows([]string{
		"id", "course_id", "parent_id", "node_type", "title", "position",
		"path", "status", "created_at", "updated_at",
	})
	for i, id := range ids {
		rows.AddRow(
			id, courseID, nil, SECTION, "Node", i, id.String(),
			CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime(),
		)
	}
	// Reverse order request.
	ordered := make([]uuid.UUID, n)
	for i := 0; i < n; i++ {
		ordered[i] = ids[n-1-i]
	}

	mock.ExpectBegin()
	expectReorderCourseLock(mock, courseID)
	expectReorderSiblingLockRoot(mock, courseID, rows)
	// maxExisting=99, count=100, temporaryBase=200
	for i, id := range ids {
		expectReorderPositionUpdate(mock, courseID, id, 200+i, false)
	}
	for i, id := range ids {
		expectReorderPositionUpdate(mock, courseID, id, n-1-i, true)
	}
	mock.ExpectCommit()

	if err := model.ReorderChildren(courseID, uuid.NullUUID{}, ordered); err != nil {
		t.Fatalf("large reorder: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReorderChildrenTransactionFailures(t *testing.T) {
	courseID, _, aID, bID, _, _ := reorderTestIDs()
	t.Run("begin", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin().WillReturnError(errors.New("begin failure"))
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID})
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
	})
	t.Run("course lock", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnError(errors.New("driver failure"))
		mock.ExpectRollback()
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID})
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
	})
	t.Run("missing course", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectRollback()
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{aID})
		if !errors.Is(err, ErrCourseNodeCourseNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeCourseNotFound)
		}
	})
	t.Run("update failure rolls back", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectReorderCourseLock(mock, courseID)
		expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil,
			struct {
				id       uuid.UUID
				position int
			}{aID, 0},
			struct {
				id       uuid.UUID
				position int
			}{bID, 1},
		))
		mock.ExpectQuery(`UPDATE "course_nodes".*RETURNING`).
			WithArgs(4, courseID, aID).
			WillReturnError(errors.New("update failure"))
		mock.ExpectRollback()
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{bID, aID})
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
	})
	t.Run("commit failure after updates", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectReorderCourseLock(mock, courseID)
		expectReorderSiblingLockRoot(mock, courseID, siblingRows(courseID, nil,
			struct {
				id       uuid.UUID
				position int
			}{aID, 0},
			struct {
				id       uuid.UUID
				position int
			}{bID, 1},
		))
		expectReorderPositionUpdate(mock, courseID, aID, 4, false)
		expectReorderPositionUpdate(mock, courseID, bID, 5, false)
		expectReorderPositionUpdate(mock, courseID, aID, 1, true)
		expectReorderPositionUpdate(mock, courseID, bID, 0, true)
		mock.ExpectCommit().WillReturnError(errors.New("commit failure"))
		err := model.ReorderChildren(courseID, uuid.NullUUID{}, []uuid.UUID{bID, aID})
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
	})
}
