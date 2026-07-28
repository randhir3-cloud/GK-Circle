package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func newCourseNodeModelTest(t *testing.T) (*CourseNodeModel, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	return InitCourseNodeModel(goqu.New("postgres", sqlDB)), mock
}

func courseNodeRows(nodeID, courseID uuid.UUID, parentID interface{}, nodeType CourseNodeType, title string, position int, path string) *sqlmock.Rows {
	now := time.Date(2026, time.July, 25, 13, 47, 7, 0, time.UTC)
	return sqlmock.NewRows([]string{
		"id", "course_id", "parent_id", "node_type", "title", "position",
		"path", "status", "created_at", "updated_at",
	}).AddRow(
		nodeID, courseID, parentID, nodeType, title, position, path,
		CourseStatusDraft, now, now,
	)
}

func expectCourseNodeCourseLock(mock sqlmock.Sqlmock, courseID uuid.UUID, exists bool) {
	rows := sqlmock.NewRows([]string{"id"})
	if exists {
		rows.AddRow(courseID)
	}
	mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(rows)
}

func expectCourseNodeInsert(
	mock sqlmock.Sqlmock,
	courseID, nodeID uuid.UUID,
	parentID interface{},
	nodeType CourseNodeType,
	title string,
	position int,
	path string,
) *sqlmock.ExpectedQuery {
	return mock.ExpectQuery(`INSERT INTO "course_nodes".*RETURNING`).
		WithArgs(
			courseID,
			nodeID,
			nodeType,
			parentID,
			path,
			position,
			CourseStatusDraft,
			title,
		)
}

func TestInitCourseNodeModel(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	if model == nil || model.db == nil || model.newUUID == nil {
		t.Fatal("course node model was not fully initialized")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected SQL: %v", err)
	}
}

func TestCreateCourseNodeValidation(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")

	tests := []struct {
		name   string
		params CreateCourseNodeParams
		want   error
	}{
		{
			name:   "course required",
			params: CreateCourseNodeParams{Title: "Node", NodeType: CourseNodeTypeSection},
			want:   ErrCourseNodeCourseRequired,
		},
		{
			name:   "title required",
			params: CreateCourseNodeParams{CourseID: courseID, Title: "  ", NodeType: CourseNodeTypeSection},
			want:   ErrCourseNodeTitleRequired,
		},
		{
			name:   "type invalid",
			params: CreateCourseNodeParams{CourseID: courseID, Title: "Node", NodeType: "UNKNOWN"},
			want:   ErrCourseNodeTypeInvalid,
		},
		{
			name:   "position invalid",
			params: CreateCourseNodeParams{CourseID: courseID, Title: "Node", NodeType: CourseNodeTypeSection, Position: -1},
			want:   ErrCourseNodePositionInvalid,
		},
		{
			name: "valid parent cannot contain nil UUID",
			params: CreateCourseNodeParams{
				CourseID: courseID,
				ParentID: uuid.NullUUID{Valid: true},
				Title:    "Node",
				NodeType: CourseNodeTypeSection,
			},
			want: ErrCourseNodeParentNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := model.CreateCourseNode(test.params)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation unexpectedly queried database: %v", err)
	}
}

func TestCreateCourseNodeTopLevelTypes(t *testing.T) {
	for _, nodeType := range []CourseNodeType{CourseNodeTypeSection, CourseNodeTypeSubject} {
		t.Run(string(nodeType), func(t *testing.T) {
			model, mock := newCourseNodeModelTest(t)
			courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
			nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
			model.newUUID = func() (uuid.UUID, error) { return nodeID, nil }
			path := encodeCourseNodePath("", nodeID)

			mock.ExpectBegin()
			expectCourseNodeCourseLock(mock, courseID, true)
			expectCourseNodeInsert(mock, courseID, nodeID, nil, nodeType, "Foundation", 0, path).
				WillReturnRows(courseNodeRows(nodeID, courseID, nil, nodeType, "Foundation", 0, path))
			mock.ExpectCommit()

			node, err := model.CreateCourseNode(CreateCourseNodeParams{
				CourseID: courseID,
				NodeType: nodeType,
				Title:    " Foundation ",
				Position: 0,
			})
			if err != nil {
				if wrapped, ok := err.(*courseNodeError); ok {
					t.Fatalf("create top-level node: %v (cause: %v)", err, wrapped.cause)
				}
				t.Fatalf("create top-level node: %v", err)
			}
			if node.ID != nodeID || node.ParentID.Valid || node.Status != CourseStatusDraft {
				t.Fatalf("unexpected node: %+v", node)
			}
			if strings.Count(node.Path, nodeID.String()) != 1 {
				t.Fatalf("path %q must contain generated node ID exactly once", node.Path)
			}
			if strings.Contains(node.Path, courseID.String()) {
				t.Fatalf("path %q contains Course ID", node.Path)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet SQL expectations: %v", err)
			}
		})
	}
}

func TestCreateCourseNodeChild(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	parentID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	nodeID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	parentPath := encodeCourseNodePath("", parentID)
	path := encodeCourseNodePath(parentPath, nodeID)
	model.newUUID = func() (uuid.UUID, error) { return nodeID, nil }

	mock.ExpectBegin()
	expectCourseNodeCourseLock(mock, courseID, true)
	mock.ExpectQuery(`SELECT "id", "course_id", "path" FROM "course_nodes".*FOR UPDATE`).
		WithArgs(parentID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "course_id", "path"}).
			AddRow(parentID, courseID, parentPath))
	expectCourseNodeInsert(mock, courseID, nodeID, parentID, CourseNodeTypeTopic, "Functions", 2, path).
		WillReturnRows(courseNodeRows(nodeID, courseID, parentID, CourseNodeTypeTopic, "Functions", 2, path))
	mock.ExpectCommit()

	node, err := model.CreateCourseNode(CreateCourseNodeParams{
		CourseID: courseID,
		ParentID: uuid.NullUUID{UUID: parentID, Valid: true},
		NodeType: CourseNodeTypeTopic,
		Title:    "Functions",
		Position: 2,
	})
	if err != nil {
		t.Fatalf("create child node: %v", err)
	}
	if !node.ParentID.Valid || node.ParentID.UUID != parentID {
		t.Fatalf("parent = %+v, want %s", node.ParentID, parentID)
	}
	if !strings.HasPrefix(node.Path, parentPath) || !strings.HasSuffix(node.Path, nodeID.String()) {
		t.Fatalf("child path %q does not preserve parent prefix and child suffix", node.Path)
	}
	if strings.Count(node.Path, nodeID.String()) != 1 || strings.Contains(node.Path, courseID.String()) {
		t.Fatalf("child path violates identifier invariants: %q", node.Path)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestCreateCourseNodeMissingCourse(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	mock.ExpectBegin()
	expectCourseNodeCourseLock(mock, courseID, false)
	mock.ExpectRollback()

	_, err := model.CreateCourseNode(CreateCourseNodeParams{
		CourseID: courseID,
		NodeType: CourseNodeTypeSection,
		Title:    "Missing",
	})
	if !errors.Is(err, ErrCourseNodeCourseNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrCourseNodeCourseNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestCreateCourseNodeParentValidation(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	otherCourseID := uuid.MustParse("019c01c7-7057-7b53-bd20-23081e70482f")
	parentID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")

	t.Run("missing", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectCourseNodeCourseLock(mock, courseID, true)
		mock.ExpectQuery(`SELECT .* FROM "course_nodes".*FOR UPDATE`).
			WithArgs(parentID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id", "path"}))
		mock.ExpectRollback()

		_, err := model.CreateCourseNode(CreateCourseNodeParams{
			CourseID: courseID,
			ParentID: uuid.NullUUID{UUID: parentID, Valid: true},
			NodeType: CourseNodeTypeTopic,
			Title:    "Missing parent",
		})
		if !errors.Is(err, ErrCourseNodeParentNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeParentNotFound)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL expectations: %v", err)
		}
	})

	t.Run("cross course", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		expectCourseNodeCourseLock(mock, courseID, true)
		mock.ExpectQuery(`SELECT .* FROM "course_nodes".*FOR UPDATE`).
			WithArgs(parentID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id", "path"}).
				AddRow(parentID, otherCourseID, encodeCourseNodePath("", parentID)))
		mock.ExpectRollback()

		_, err := model.CreateCourseNode(CreateCourseNodeParams{
			CourseID: courseID,
			ParentID: uuid.NullUUID{UUID: parentID, Valid: true},
			NodeType: CourseNodeTypeTopic,
			Title:    "Cross course",
		})
		if !errors.Is(err, ErrCourseNodeCrossCourseParent) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeCrossCourseParent)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL expectations: %v", err)
		}
	})
}

func TestCreateCourseNodePositionConflicts(t *testing.T) {
	for _, constraint := range []string{
		courseNodesTopLevelPositionConstraint,
		courseNodesChildPositionConstraint,
	} {
		t.Run(constraint, func(t *testing.T) {
			model, mock := newCourseNodeModelTest(t)
			courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
			nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
			model.newUUID = func() (uuid.UUID, error) { return nodeID, nil }
			path := encodeCourseNodePath("", nodeID)

			mock.ExpectBegin()
			expectCourseNodeCourseLock(mock, courseID, true)
			expectCourseNodeInsert(mock, courseID, nodeID, nil, CourseNodeTypeSection, "Node", 0, path).
				WillReturnError(&pq.Error{Code: "23505", Constraint: constraint})
			mock.ExpectRollback()

			_, err := model.CreateCourseNode(CreateCourseNodeParams{
				CourseID: courseID,
				NodeType: CourseNodeTypeSection,
				Title:    "Node",
			})
			if !errors.Is(err, ErrCourseNodePositionConflict) {
				t.Fatalf("error = %v, want %v", err, ErrCourseNodePositionConflict)
			}
			var pqErr *pq.Error
			if !errors.As(err, &pqErr) {
				t.Fatal("mapped error does not preserve PostgreSQL cause")
			}
			if err.Error() != ErrCourseNodePositionConflict.Error() {
				t.Fatalf("public error leaked SQL detail: %q", err.Error())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet SQL expectations: %v", err)
			}
		})
	}
}

func TestCreateCourseNodePersistenceFailures(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	driverErr := errors.New("driver detail")

	t.Run("begin", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin().WillReturnError(driverErr)
		_, err := model.CreateCourseNode(CreateCourseNodeParams{
			CourseID: courseID, NodeType: CourseNodeTypeSection, Title: "Node",
		})
		if !errors.Is(err, ErrCourseNodePersistence) || !errors.Is(err, driverErr) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
		if strings.Contains(err.Error(), driverErr.Error()) {
			t.Fatalf("public error leaked driver detail: %q", err.Error())
		}
	})

	t.Run("course lock rollback", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
			WithArgs(courseID, uint(1)).
			WillReturnError(driverErr)
		mock.ExpectRollback()

		_, err := model.CreateCourseNode(CreateCourseNodeParams{
			CourseID: courseID, NodeType: CourseNodeTypeSection, Title: "Node",
		})
		if !errors.Is(err, ErrCourseNodePersistence) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
		if strings.Contains(err.Error(), driverErr.Error()) {
			t.Fatalf("public error leaked driver detail: %q", err.Error())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL expectations: %v", err)
		}
	})

	t.Run("insert rollback", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		model.newUUID = func() (uuid.UUID, error) { return nodeID, nil }
		path := encodeCourseNodePath("", nodeID)
		mock.ExpectBegin()
		expectCourseNodeCourseLock(mock, courseID, true)
		expectCourseNodeInsert(mock, courseID, nodeID, nil, CourseNodeTypeSection, "Node", 0, path).
			WillReturnError(driverErr)
		mock.ExpectRollback()

		_, err := model.CreateCourseNode(CreateCourseNodeParams{
			CourseID: courseID, NodeType: CourseNodeTypeSection, Title: "Node",
		})
		if !errors.Is(err, ErrCourseNodePersistence) || !errors.Is(err, driverErr) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL expectations: %v", err)
		}
	})

	t.Run("returning scan rollback", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		model.newUUID = func() (uuid.UUID, error) { return nodeID, nil }
		path := encodeCourseNodePath("", nodeID)
		mock.ExpectBegin()
		expectCourseNodeCourseLock(mock, courseID, true)
		expectCourseNodeInsert(mock, courseID, nodeID, nil, CourseNodeTypeSection, "Node", 0, path).
			WillReturnRows(courseNodeRows(nodeID, courseID, nil, CourseNodeTypeSection, "Node", 0, path).
				RowError(0, driverErr))
		mock.ExpectRollback()

		_, err := model.CreateCourseNode(CreateCourseNodeParams{
			CourseID: courseID, NodeType: CourseNodeTypeSection, Title: "Node",
		})
		if !errors.Is(err, ErrCourseNodePersistence) || !errors.Is(err, driverErr) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL expectations: %v", err)
		}
	})

	t.Run("commit", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		model.newUUID = func() (uuid.UUID, error) { return nodeID, nil }
		path := encodeCourseNodePath("", nodeID)
		mock.ExpectBegin()
		expectCourseNodeCourseLock(mock, courseID, true)
		expectCourseNodeInsert(mock, courseID, nodeID, nil, CourseNodeTypeSection, "Node", 0, path).
			WillReturnRows(courseNodeRows(nodeID, courseID, nil, CourseNodeTypeSection, "Node", 0, path))
		mock.ExpectCommit().WillReturnError(driverErr)

		_, err := model.CreateCourseNode(CreateCourseNodeParams{
			CourseID: courseID, NodeType: CourseNodeTypeSection, Title: "Node",
		})
		if !errors.Is(err, ErrCourseNodePersistence) || !errors.Is(err, driverErr) {
			t.Fatalf("error = %v, want persistence wrapper", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL expectations: %v", err)
		}
	})
}

func TestGetCourseNodeByID(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	path := encodeCourseNodePath("", nodeID)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes" WHERE (("course_id" = $1) AND ("id" = $2)) LIMIT $3`,
	)).
		WithArgs(courseID, nodeID, uint(1)).
		WillReturnRows(courseNodeRows(nodeID, courseID, nil, CourseNodeTypeSubject, "Subject", 0, path))

	node, err := model.GetCourseNodeByID(courseID, nodeID)
	if err != nil {
		t.Fatalf("get CourseNode: %v", err)
	}
	if node.ID != nodeID || node.CourseID != courseID {
		t.Fatalf("lookup returned wrong node: %+v", node)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestGetCourseNodeByIDNotFoundIsCourseScoped(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")

	mock.ExpectQuery(`SELECT .* FROM "course_nodes" WHERE .*"course_id".*"id".*`).
		WithArgs(courseID, nodeID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "parent_id", "node_type", "title", "position",
			"path", "status", "created_at", "updated_at",
		}))

	_, err := model.GetCourseNodeByID(courseID, nodeID)
	if !errors.Is(err, ErrCourseNodeNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrCourseNodeNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestCourseNodePathExcludedFromJSON(t *testing.T) {
	node := CourseNode{
		ID:   uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"),
		Path: "internal-path",
	}
	encoded, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("marshal CourseNode: %v", err)
	}
	if strings.Contains(string(encoded), "path") || strings.Contains(string(encoded), "internal-path") {
		t.Fatalf("JSON exposed persistence path: %s", encoded)
	}
}

func TestCourseNodeDefaultUUIDStrategy(t *testing.T) {
	model, _ := newCourseNodeModelTest(t)
	if model.newUUID == nil {
		t.Fatal("default UUID generator is nil")
	}
	generated, err := model.newUUID()
	if err != nil {
		t.Fatalf("generate UUID: %v", err)
	}
	if generated == uuid.Nil {
		t.Fatal("canonical UUID generator returned nil UUID")
	}
}

func TestCourseNodeErrorWrapperPreservesCauseWithoutLeakingText(t *testing.T) {
	cause := errors.New("sensitive driver detail")
	err := newCourseNodePersistenceError(cause)
	if !errors.Is(err, ErrCourseNodePersistence) || !errors.Is(err, cause) {
		t.Fatalf("wrapper does not preserve causes: %v", err)
	}
	if err.Error() != ErrCourseNodePersistence.Error() {
		t.Fatalf("wrapper leaked cause: %q", err.Error())
	}
}
