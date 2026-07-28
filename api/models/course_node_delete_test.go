package models

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func deleteTestIDs() (courseID, rootID, childID, grandID, siblingID uuid.UUID) {
	return uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481"),
		uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"),
		uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6"),
		uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a"),
		uuid.MustParse("019c01cb-7f39-7f26-900f-6947e75e7284")
}

func expectDeleteCourseLock(mock sqlmock.Sqlmock, courseID uuid.UUID) {
	mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(courseID))
}

func expectDeleteTargetLock(mock sqlmock.Sqlmock, courseID, nodeID uuid.UUID, parentID interface{}, position int, path string) {
	mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
		WithArgs(courseID, nodeID, uint(1)).
		WillReturnRows(courseNodeRows(nodeID, courseID, parentID, SECTION, "Node", position, path))
}

func expectDeleteSubtreeLock(mock sqlmock.Sqlmock, courseID uuid.UUID, rootPath string, rows *sqlmock.Rows) {
	mock.ExpectQuery(regexp.QuoteMeta(courseNodeSubtreeDeleteLockQuery)).
		WithArgs(courseID, rootPath, courseNodeDescendantPathPattern(rootPath)).
		WillReturnRows(rows)
}

func expectDeleteSubtree(mock sqlmock.Sqlmock, courseID uuid.UUID, rootPath string, deleted ...uuid.UUID) {
	rows := sqlmock.NewRows([]string{"id"})
	for _, id := range deleted {
		rows.AddRow(id)
	}
	mock.ExpectQuery(regexp.QuoteMeta(courseNodeSubtreeDeleteQuery)).
		WithArgs(courseID, rootPath, courseNodeDescendantPathPattern(rootPath)).
		WillReturnRows(rows)
}

func TestVerifyDeleteIDs(t *testing.T) {
	a := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	b := uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a")
	c := uuid.MustParse("019c01cb-7f39-7f26-900f-6947e75e7284")
	expected := map[uuid.UUID]struct{}{a: {}, b: {}}

	if err := verifyDeleteIDs(expected, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("exact match: %v", err)
	}
	if err := verifyDeleteIDs(expected, []uuid.UUID{a}); !errors.Is(err, ErrCourseNodeDeleteConflict) {
		t.Fatalf("missing id: got %v", err)
	}
	if err := verifyDeleteIDs(expected, []uuid.UUID{a, b, c}); !errors.Is(err, ErrCourseNodeDeleteConflict) {
		t.Fatalf("unexpected id: got %v", err)
	}
	if err := verifyDeleteIDs(expected, []uuid.UUID{a, a}); !errors.Is(err, ErrCourseNodeDeleteConflict) {
		t.Fatalf("duplicate id: got %v", err)
	}
	expectedThree := map[uuid.UUID]struct{}{a: {}, b: {}, c: {}}
	if err := verifyDeleteIDs(expectedThree, []uuid.UUID{a, a, b}); !errors.Is(err, ErrCourseNodeDeleteConflict) {
		t.Fatalf("duplicate hiding missing: got %v", err)
	}
}

func TestDeleteSubtreeValidation(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	_, nodeID, _, _, _ := deleteTestIDs()
	tests := []struct {
		name     string
		courseID uuid.UUID
		nodeID   uuid.UUID
		want     error
	}{
		{"nil course", uuid.Nil, nodeID, ErrCourseNodeCourseRequired},
		{"nil node", uuid.New(), uuid.Nil, ErrCourseNodeNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := model.DeleteSubtree(test.courseID, test.nodeID)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation queried database: %v", err)
	}
}

func TestDeleteSubtreeExistenceFailures(t *testing.T) {
	courseID, nodeID, _, _, _ := deleteTestIDs()
	t.Run("missing course", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, nodeID)
		if !errors.Is(err, ErrCourseNodeCourseNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeCourseNotFound)
		}
	})
	t.Run("missing node", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		mock.ExpectQuery(`SELECT .* FROM "course_nodes".*FOR UPDATE`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}))
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, nodeID)
		if !errors.Is(err, ErrCourseNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeNotFound)
		}
	})
	t.Run("cross course node", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		mock.ExpectQuery(`SELECT .* FROM "course_nodes".*FOR UPDATE`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}))
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, nodeID)
		if !errors.Is(err, ErrCourseNodeNotFound) {
			t.Fatalf("cross-course appears as not found: got %v", err)
		}
	})
}

func TestDeleteSubtreeLeaf(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID, rootID, childID, _, _ := deleteTestIDs()
	childPath := rootID.String() + "/" + childID.String()

	mock.ExpectBegin()
	expectDeleteCourseLock(mock, courseID)
	expectDeleteTargetLock(mock, courseID, childID, rootID, 0, childPath)
	expectDeleteSubtreeLock(mock, courseID, childPath, courseNodeRows(childID, courseID, rootID, TOPIC, "Leaf", 0, childPath))
	expectDeleteSubtree(mock, courseID, childPath, childID)
	mock.ExpectCommit()

	if err := model.DeleteSubtree(courseID, childID); err != nil {
		t.Fatalf("delete leaf: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteSubtreeBranchAndDeep(t *testing.T) {
	courseID, rootID, childID, grandID, siblingID := deleteTestIDs()
	rootPath := rootID.String()
	childPath := rootPath + "/" + childID.String()
	grandPath := childPath + "/" + grandID.String()
	siblingPath := rootPath + "/" + siblingID.String()

	t.Run("branch", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, childID, rootID, 0, childPath)
		expectDeleteSubtreeLock(mock, courseID, childPath,
			courseNodeRows(childID, courseID, rootID, SUBJECT, "Branch", 0, childPath).
				AddRow(grandID, courseID, childID, TOPIC, "Deep", 0, grandPath, CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))
		expectDeleteSubtree(mock, courseID, childPath, grandID, childID)
		mock.ExpectCommit()
		if err := model.DeleteSubtree(courseID, childID); err != nil {
			t.Fatalf("delete branch: %v", err)
		}
	})

	t.Run("deep subtree", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, rootID, nil, 0, rootPath)
		expectDeleteSubtreeLock(mock, courseID, rootPath,
			courseNodeRows(rootID, courseID, nil, SECTION, "Root", 0, rootPath).
				AddRow(childID, courseID, rootID, SUBJECT, "Child", 0, childPath, CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()).
				AddRow(grandID, courseID, childID, TOPIC, "Grand", 0, grandPath, CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))
		expectDeleteSubtree(mock, courseID, rootPath, grandID, childID, rootID)
		mock.ExpectCommit()
		if err := model.DeleteSubtree(courseID, rootID); err != nil {
			t.Fatalf("delete deep: %v", err)
		}
	})

	t.Run("root subtree leaves unrelated sibling", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		// Deleting child branch must not touch sibling; mock only expects child subtree delete.
		_ = siblingPath
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, childID, rootID, 0, childPath)
		expectDeleteSubtreeLock(mock, courseID, childPath,
			courseNodeRows(childID, courseID, rootID, SUBJECT, "Branch", 0, childPath).
				AddRow(grandID, courseID, childID, TOPIC, "Deep", 0, grandPath, CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))
		expectDeleteSubtree(mock, courseID, childPath, grandID, childID)
		mock.ExpectCommit()
		if err := model.DeleteSubtree(courseID, childID); err != nil {
			t.Fatalf("unrelated branch survival scope: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("delete must not touch unrelated sibling: %v", err)
		}
	})
}

func TestDeleteSubtreeConflictOnVerification(t *testing.T) {
	courseID, rootID, childID, _, _ := deleteTestIDs()
	rootPath := rootID.String()
	childPath := rootPath + "/" + childID.String()

	t.Run("missing returned id", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, rootID, nil, 0, rootPath)
		expectDeleteSubtreeLock(mock, courseID, rootPath,
			courseNodeRows(rootID, courseID, nil, SECTION, "Root", 0, rootPath).
				AddRow(childID, courseID, rootID, SUBJECT, "Child", 0, childPath, CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))
		expectDeleteSubtree(mock, courseID, rootPath, rootID)
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, rootID)
		if !errors.Is(err, ErrCourseNodeDeleteConflict) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeDeleteConflict)
		}
	})

	t.Run("unexpected returned id", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		foreign := uuid.MustParse("019c01cc-1111-7222-8333-944455556666")
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, rootID, nil, 0, rootPath)
		expectDeleteSubtreeLock(mock, courseID, rootPath, courseNodeRows(rootID, courseID, nil, SECTION, "Root", 0, rootPath))
		expectDeleteSubtree(mock, courseID, rootPath, rootID, foreign)
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, rootID)
		if !errors.Is(err, ErrCourseNodeDeleteConflict) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeDeleteConflict)
		}
	})

	t.Run("duplicate returned id", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, rootID, nil, 0, rootPath)
		expectDeleteSubtreeLock(mock, courseID, rootPath,
			courseNodeRows(rootID, courseID, nil, SECTION, "Root", 0, rootPath).
				AddRow(childID, courseID, rootID, SUBJECT, "Child", 0, childPath, CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))
		expectDeleteSubtree(mock, courseID, rootPath, rootID, rootID)
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, rootID)
		if !errors.Is(err, ErrCourseNodeDeleteConflict) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeDeleteConflict)
		}
	})

	t.Run("boundary unsafe path rejected", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		unsafePath := rootPath + "d"
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, rootID, nil, 0, rootPath)
		expectDeleteSubtreeLock(mock, courseID, rootPath,
			courseNodeRows(rootID, courseID, nil, SECTION, "Root", 0, rootPath).
				AddRow(childID, courseID, nil, SECTION, "Unsafe", 1, unsafePath, CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, rootID)
		if !errors.Is(err, ErrCourseNodeDeleteConflict) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeDeleteConflict)
		}
	})

	t.Run("empty subtree lock", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, rootID, nil, 0, rootPath)
		expectDeleteSubtreeLock(mock, courseID, rootPath, sqlmock.NewRows([]string{
			"id", "course_id", "parent_id", "node_type", "title", "position",
			"path", "status", "created_at", "updated_at",
		}))
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, rootID)
		if !errors.Is(err, ErrCourseNodeDeleteConflict) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeDeleteConflict)
		}
	})
}

func TestDeleteSubtreeTransactionFailures(t *testing.T) {
	courseID, rootID, _, _, _ := deleteTestIDs()
	rootPath := rootID.String()

	t.Run("begin", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin().WillReturnError(errors.New("begin failure"))
		err := model.DeleteSubtree(courseID, rootID)
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
	})
	t.Run("course lock query", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnError(errors.New("driver failure"))
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, rootID)
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
	})
	t.Run("delete failure", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, rootID, nil, 0, rootPath)
		expectDeleteSubtreeLock(mock, courseID, rootPath, courseNodeRows(rootID, courseID, nil, SECTION, "Root", 0, rootPath))
		mock.ExpectQuery(regexp.QuoteMeta(courseNodeSubtreeDeleteQuery)).
			WithArgs(courseID, rootPath, courseNodeDescendantPathPattern(rootPath)).
			WillReturnError(errors.New("delete failure"))
		mock.ExpectRollback()
		err := model.DeleteSubtree(courseID, rootID)
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
	})
	t.Run("commit failure", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectDeleteCourseLock(mock, courseID)
		expectDeleteTargetLock(mock, courseID, rootID, nil, 0, rootPath)
		expectDeleteSubtreeLock(mock, courseID, rootPath, courseNodeRows(rootID, courseID, nil, SECTION, "Root", 0, rootPath))
		expectDeleteSubtree(mock, courseID, rootPath, rootID)
		mock.ExpectCommit().WillReturnError(errors.New("commit failure"))
		err := model.DeleteSubtree(courseID, rootID)
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
	})
}
