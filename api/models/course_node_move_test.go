package models

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func moveTestIDs() (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	return uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481"),
		uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"),
		uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6"),
		uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a"),
		uuid.MustParse("019c01cb-7f39-7f26-900f-6947e75e7284")
}

func moveNodeRow(nodeID, courseID uuid.UUID, parentID interface{}, position int, path string) *sqlmock.Rows {
	return courseNodeRows(nodeID, courseID, parentID, SECTION, "Node", position, path)
}

func expectMoveCourseLock(mock sqlmock.Sqlmock, courseID uuid.UUID) {
	mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
}

func expectMoveNodeLock(mock sqlmock.Sqlmock, courseID, nodeID uuid.UUID, parentID interface{}, position int, path string) {
	mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
		WithArgs(courseID, nodeID, uint(1)).
		WillReturnRows(moveNodeRow(nodeID, courseID, parentID, position, path))
}

func expectMoveParentLock(mock sqlmock.Sqlmock, courseID, parentID uuid.UUID, path string) {
	mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
		WithArgs(parentID, uint(1)).
		WillReturnRows(moveNodeRow(parentID, courseID, nil, 0, path))
}

func expectSubtreeLock(mock sqlmock.Sqlmock, courseID uuid.UUID, rootPath string, rows *sqlmock.Rows) {
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, course_id, parent_id, node_type, title, position, path, status, created_at, updated_at\nFROM course_nodes\nWHERE course_id = $1 AND (path = $2 OR path LIKE $3)\nORDER BY id\nFOR UPDATE")).
		WithArgs(courseID, rootPath, courseNodeDescendantPathPattern(rootPath)).
		WillReturnRows(rows)
}

func TestMoveNodeValidation(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	_, nodeID, _, _, _ := moveTestIDs()
	tests := []struct {
		name     string
		courseID uuid.UUID
		nodeID   uuid.UUID
		parentID uuid.NullUUID
		position int
		want     error
	}{
		{"course required", uuid.Nil, nodeID, uuid.NullUUID{}, 0, ErrCourseNodeCourseRequired},
		{"node required", uuid.New(), uuid.Nil, uuid.NullUUID{}, 0, ErrCourseNodeNotFound},
		{"invalid nullable parent", uuid.New(), nodeID, uuid.NullUUID{Valid: true}, 0, ErrCourseNodeParentNotFound},
		{"negative position", uuid.New(), nodeID, uuid.NullUUID{}, -1, ErrCourseNodePositionInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := model.MoveNode(test.courseID, test.nodeID, test.parentID, test.position)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation unexpectedly queried the database: %v", err)
	}
}

func TestMoveNodeSameParentNoopAndReorderDeferral(t *testing.T) {
	courseID, nodeID, _, _, _ := moveTestIDs()
	path := nodeID.String()
	t.Run("no op", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectMoveCourseLock(mock, courseID)
		expectMoveNodeLock(mock, courseID, nodeID, nil, 2, path)
		mock.ExpectCommit()
		if err := model.MoveNode(courseID, nodeID, uuid.NullUUID{}, 2); err != nil {
			t.Fatalf("no-op move: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("reorder deferred", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectMoveCourseLock(mock, courseID)
		expectMoveNodeLock(mock, courseID, nodeID, nil, 2, path)
		mock.ExpectRollback()
		err := model.MoveNode(courseID, nodeID, uuid.NullUUID{}, 3)
		if !errors.Is(err, ErrCourseNodeInvalidMove) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeInvalidMove)
		}
	})
}

func TestMoveNodeScopeFailures(t *testing.T) {
	courseID, nodeID, _, destinationID, _ := moveTestIDs()
	t.Run("missing course", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectRollback()
		err := model.MoveNode(courseID, nodeID, uuid.NullUUID{}, 0)
		if !errors.Is(err, ErrCourseNodeCourseNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeCourseNotFound)
		}
	})
	t.Run("missing node", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectMoveCourseLock(mock, courseID)
		mock.ExpectQuery(`SELECT .* FROM "course_nodes".*FOR UPDATE`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at"}))
		mock.ExpectRollback()
		err := model.MoveNode(courseID, nodeID, uuid.NullUUID{}, 0)
		if !errors.Is(err, ErrCourseNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeNotFound)
		}
	})
	t.Run("cross course destination", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		otherCourseID := uuid.New()
		mock.ExpectBegin()
		expectMoveCourseLock(mock, courseID)
		expectMoveNodeLock(mock, courseID, nodeID, nil, 0, nodeID.String())
		mock.ExpectQuery(`SELECT .* FROM "course_nodes".*FOR UPDATE`).
			WithArgs(destinationID, uint(1)).
			WillReturnRows(moveNodeRow(destinationID, otherCourseID, nil, 0, destinationID.String()))
		mock.ExpectRollback()
		err := model.MoveNode(courseID, nodeID, uuid.NullUUID{UUID: destinationID, Valid: true}, 2)
		if !errors.Is(err, ErrCourseNodeCrossCourseParent) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeCrossCourseParent)
		}
	})
}

func TestMoveNodeReparentsAndRewritesCompleteSubtree(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID, rootID, childID, destinationID, _ := moveTestIDs()
	oldPath := rootID.String()
	childPath := oldPath + "/" + childID.String()
	destinationPath := destinationID.String()
	newPath := destinationPath + "/" + rootID.String()

	mock.ExpectBegin()
	expectMoveCourseLock(mock, courseID)
	expectMoveNodeLock(mock, courseID, rootID, nil, 0, oldPath)
	expectMoveParentLock(mock, courseID, destinationID, destinationPath)
	expectSubtreeLock(mock, courseID, oldPath, moveNodeRow(rootID, courseID, nil, 0, oldPath).
		AddRow(childID, courseID, rootID, TOPIC, "Child", 0, childPath, CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))
	mock.ExpectQuery(`SELECT "id" FROM "course_nodes".*FOR UPDATE`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(regexp.QuoteMeta(courseNodeSubtreePathRewriteQuery)).
		WithArgs(courseID, oldPath, newPath, courseNodeDescendantPathPattern(oldPath)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(rootID).AddRow(childID))
	mock.ExpectExec(`UPDATE "course_nodes"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := model.MoveNode(courseID, rootID, uuid.NullUUID{UUID: destinationID, Valid: true}, 2)
	if err != nil {
		t.Fatalf("move subtree: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestMoveNodeRejectsCycleAndPositionConflict(t *testing.T) {
	courseID, rootID, childID, destinationID, _ := moveTestIDs()
	oldPath := rootID.String()
	t.Run("destination is descendant", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectMoveCourseLock(mock, courseID)
		expectMoveNodeLock(mock, courseID, rootID, nil, 0, oldPath)
		expectMoveParentLock(mock, courseID, childID, oldPath+"/"+childID.String())
		expectSubtreeLock(mock, courseID, oldPath, moveNodeRow(rootID, courseID, nil, 0, oldPath).
			AddRow(childID, courseID, rootID, TOPIC, "Child", 0, oldPath+"/"+childID.String(), CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))
		mock.ExpectRollback()
		err := model.MoveNode(courseID, rootID, uuid.NullUUID{UUID: childID, Valid: true}, 2)
		if !errors.Is(err, ErrCourseNodeCycle) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeCycle)
		}
	})
	t.Run("destination is self", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectMoveCourseLock(mock, courseID)
		expectMoveNodeLock(mock, courseID, rootID, nil, 0, oldPath)
		expectMoveParentLock(mock, courseID, rootID, oldPath)
		expectSubtreeLock(mock, courseID, oldPath, moveNodeRow(rootID, courseID, nil, 0, oldPath))
		mock.ExpectRollback()
		err := model.MoveNode(courseID, rootID, uuid.NullUUID{UUID: rootID, Valid: true}, 2)
		if !errors.Is(err, ErrCourseNodeCycle) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeCycle)
		}
	})
	t.Run("destination position conflict", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectMoveCourseLock(mock, courseID)
		expectMoveNodeLock(mock, courseID, rootID, nil, 0, oldPath)
		expectMoveParentLock(mock, courseID, destinationID, destinationID.String())
		expectSubtreeLock(mock, courseID, oldPath, moveNodeRow(rootID, courseID, nil, 0, oldPath))
		mock.ExpectQuery(`SELECT "id" FROM "course_nodes".*FOR UPDATE`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectRollback()
		err := model.MoveNode(courseID, rootID, uuid.NullUUID{UUID: destinationID, Valid: true}, 2)
		if !errors.Is(err, ErrCourseNodePositionConflict) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodePositionConflict)
		}
	})
}

func TestMoveNodeDetectsBoundarySafeRewriteMismatchAndRollsBack(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID, rootID, childID, destinationID, _ := moveTestIDs()
	oldPath := rootID.String()
	childPath := oldPath + "/" + childID.String()
	mock.ExpectBegin()
	expectMoveCourseLock(mock, courseID)
	expectMoveNodeLock(mock, courseID, rootID, nil, 0, oldPath)
	expectMoveParentLock(mock, courseID, destinationID, destinationID.String())
	expectSubtreeLock(mock, courseID, oldPath, moveNodeRow(rootID, courseID, nil, 0, oldPath).
		AddRow(childID, courseID, rootID, TOPIC, "Child", 0, childPath, CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))
	mock.ExpectQuery(`SELECT "id" FROM "course_nodes".*FOR UPDATE`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(regexp.QuoteMeta(courseNodeSubtreePathRewriteQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(rootID))
	mock.ExpectRollback()
	err := model.MoveNode(courseID, rootID, uuid.NullUUID{UUID: destinationID, Valid: true}, 2)
	if !errors.Is(err, ErrCourseNodeSubtreeConflict) {
		t.Fatalf("error = %v, want %v", err, ErrCourseNodeSubtreeConflict)
	}
	if isCourseNodePathInSubtree("abcd", "abc") || !isCourseNodePathInSubtree("abc/child", "abc") {
		t.Fatal("path matching is not segment-boundary safe")
	}
	if courseNodeDescendantPathPattern("abc") != "abc/%" {
		t.Fatalf("unexpected descendant pattern: %q", courseNodeDescendantPathPattern("abc"))
	}
}

func TestMoveNodeRollbackOnPersistenceAndCommitFailure(t *testing.T) {
	courseID, rootID, _, _, _ := moveTestIDs()
	t.Run("course lock", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnError(errors.New("driver failure"))
		mock.ExpectRollback()
		err := model.MoveNode(courseID, rootID, uuid.NullUUID{}, 2)
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
	})
	t.Run("commit", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		path := rootID.String()
		mock.ExpectBegin()
		expectMoveCourseLock(mock, courseID)
		expectMoveNodeLock(mock, courseID, rootID, nil, 2, path)
		mock.ExpectCommit().WillReturnError(errors.New("commit failure"))
		err := model.MoveNode(courseID, rootID, uuid.NullUUID{}, 2)
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
	})
}
