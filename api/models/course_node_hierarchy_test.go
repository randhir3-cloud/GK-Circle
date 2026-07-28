package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func expectCourseRead(mock sqlmock.Sqlmock, courseID uuid.UUID, exists bool) {
	rows := sqlmock.NewRows([]string{"id"})
	if exists {
		rows.AddRow(courseID)
	}
	mock.ExpectQuery(`SELECT "id" FROM "courses".*LIMIT`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(rows)
}

func expectHierarchyParentRead(mock sqlmock.Sqlmock, parentID, courseID uuid.UUID) *sqlmock.ExpectedQuery {
	return mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes".*LIMIT`).
		WithArgs(parentID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(parentID, courseID))
}

func hierarchyRows(total int64, nodes ...CourseNode) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		"id", "course_id", "parent_id", "node_type", "title", "position",
		"path", "status", "created_at", "updated_at", "total_nodes",
	})
	if len(nodes) == 0 {
		return rows.AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, total)
	}
	for _, node := range nodes {
		var parentID interface{}
		if node.ParentID.Valid {
			parentID = node.ParentID.UUID
		}
		rows.AddRow(
			node.ID, node.CourseID, parentID, node.NodeType, node.Title, node.Position,
			node.Path, node.Status, node.CreatedAt, node.UpdatedAt, total,
		)
	}
	return rows
}

func TestListRootNodes(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	firstID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	secondID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	expectCourseRead(mock, courseID, true)
	mock.ExpectQuery(`SELECT .* FROM "course_nodes".*"parent_id" IS NULL.*ORDER BY "position" ASC, "id" ASC`).
		WithArgs(courseID).
		WillReturnRows(courseNodeRows(firstID, courseID, nil, SECTION, "Two", 2, firstID.String()).
			AddRow(secondID, courseID, nil, SECTION, "Ten", 10, secondID.String(), CourseStatusDraft, testCourseNodeTime(), testCourseNodeTime()))

	nodes, err := model.ListRootNodes(courseID)
	if err != nil {
		t.Fatalf("list roots: %v", err)
	}
	if len(nodes) != 2 || nodes[0].Position != 2 || nodes[1].Position != 10 {
		t.Fatalf("root ordering = %+v, want positions 2 then 10", nodes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestListRootNodesDistinguishesMissingAndEmptyCourse(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	t.Run("missing course", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		expectCourseRead(mock, courseID, false)
		_, err := model.ListRootNodes(courseID)
		if !errors.Is(err, ErrCourseNodeCourseNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeCourseNotFound)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("empty course", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		expectCourseRead(mock, courseID, true)
		mock.ExpectQuery(`SELECT .* FROM "course_nodes".*"parent_id" IS NULL`).
			WithArgs(courseID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at"}))
		nodes, err := model.ListRootNodes(courseID)
		if err != nil || nodes == nil || len(nodes) != 0 {
			t.Fatalf("nodes, err = %#v, %v; want non-nil empty result", nodes, err)
		}
	})
}

func TestListChildrenValidatesCourseScopedParent(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	otherCourseID := uuid.MustParse("019c01c7-7057-7b53-bd20-23081e70482f")
	parentID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	t.Run("cross course", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		expectCourseRead(mock, courseID, true)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes".*LIMIT`).
			WithArgs(parentID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(parentID, otherCourseID))
		_, err := model.ListChildren(courseID, parentID)
		if !errors.Is(err, ErrCourseNodeCrossCourseParent) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeCrossCourseParent)
		}
	})
	t.Run("ordered children", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		childID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
		expectCourseRead(mock, courseID, true)
		expectHierarchyParentRead(mock, parentID, courseID)
		mock.ExpectQuery(`SELECT .* FROM "course_nodes".*ORDER BY "position" ASC, "id" ASC`).
			WithArgs(courseID, parentID).
			WillReturnRows(courseNodeRows(childID, courseID, parentID, TOPIC, "Child", 2, parentID.String()+"/"+childID.String()))
		nodes, err := model.ListChildren(courseID, parentID)
		if err != nil || len(nodes) != 1 || nodes[0].ParentID.UUID != parentID {
			t.Fatalf("children, err = %#v, %v", nodes, err)
		}
	})
}

func TestGetHierarchyBuildsNumericPreorder(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	rootID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	childTwoID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	grandchildID := uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a")
	childTenID := uuid.MustParse("019c01cb-7f39-7f26-900f-6947e75e7284")
	childHundredID := uuid.MustParse("019c01cc-5e0c-7e57-bf98-6fd62c0de9b7")
	rootPath := rootID.String()
	childTwoPath := rootPath + "/" + childTwoID.String()
	expectCourseRead(mock, courseID, true)
	mock.ExpectQuery(regexp.QuoteMeta("WITH RECURSIVE course_rows AS (")).
		WithArgs(courseID).
		WillReturnRows(hierarchyRows(5,
			CourseNode{ID: rootID, CourseID: courseID, NodeType: SECTION, Title: "Root", Position: 0, Path: rootPath, Status: CourseStatusDraft, CreatedAt: testCourseNodeTime(), UpdatedAt: testCourseNodeTime()},
			CourseNode{ID: childTwoID, CourseID: courseID, ParentID: uuid.NullUUID{UUID: rootID, Valid: true}, NodeType: SUBJECT, Title: "Two", Position: 2, Path: childTwoPath, Status: CourseStatusDraft, CreatedAt: testCourseNodeTime(), UpdatedAt: testCourseNodeTime()},
			CourseNode{ID: grandchildID, CourseID: courseID, ParentID: uuid.NullUUID{UUID: childTwoID, Valid: true}, NodeType: TOPIC, Title: "Grandchild", Position: 0, Path: childTwoPath + "/" + grandchildID.String(), Status: CourseStatusDraft, CreatedAt: testCourseNodeTime(), UpdatedAt: testCourseNodeTime()},
			CourseNode{ID: childTenID, CourseID: courseID, ParentID: uuid.NullUUID{UUID: rootID, Valid: true}, NodeType: SUBJECT, Title: "Ten", Position: 10, Path: rootPath + "/" + childTenID.String(), Status: CourseStatusDraft, CreatedAt: testCourseNodeTime(), UpdatedAt: testCourseNodeTime()},
			CourseNode{ID: childHundredID, CourseID: courseID, ParentID: uuid.NullUUID{UUID: rootID, Valid: true}, NodeType: SUBJECT, Title: "Hundred", Position: 100, Path: rootPath + "/" + childHundredID.String(), Status: CourseStatusDraft, CreatedAt: testCourseNodeTime(), UpdatedAt: testCourseNodeTime()},
		))

	hierarchy, err := model.GetHierarchy(courseID)
	if err != nil {
		t.Fatalf("get hierarchy: %v", err)
	}
	if hierarchy.CourseID != courseID || len(hierarchy.Roots) != 1 || hierarchy.Roots[0].Children == nil {
		t.Fatalf("unexpected hierarchy: %#v", hierarchy)
	}
	children := hierarchy.Roots[0].Children
	if len(children) != 3 || children[0].Node.Position != 2 || children[1].Node.Position != 10 || children[2].Node.Position != 100 || len(children[0].Children) != 1 {
		t.Fatalf("numeric preorder was not preserved: %#v", children)
	}
	encoded, err := json.Marshal(hierarchy)
	if err != nil || strings.Contains(string(encoded), `"path":`) {
		t.Fatalf("hierarchy JSON exposed path or failed: %s, %v", encoded, err)
	}
}

func TestGetHierarchyReturnsNonNilEmptyAndDetectsIncompleteGraph(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	t.Run("empty", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		expectCourseRead(mock, courseID, true)
		mock.ExpectQuery(`WITH RECURSIVE`).WithArgs(courseID).WillReturnRows(hierarchyRows(0))
		hierarchy, err := model.GetHierarchy(courseID)
		if err != nil || hierarchy.Roots == nil || len(hierarchy.Roots) != 0 {
			t.Fatalf("hierarchy, err = %#v, %v; want non-nil empty roots", hierarchy, err)
		}
	})
	t.Run("disconnected", func(t *testing.T) {
		model, mock := newCourseNodeModelTest(t)
		expectCourseRead(mock, courseID, true)
		mock.ExpectQuery(`WITH RECURSIVE`).WithArgs(courseID).WillReturnRows(hierarchyRows(1))
		_, err := model.GetHierarchy(courseID)
		if !errors.Is(err, ErrCourseNodeHierarchyIntegrity) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNodeHierarchyIntegrity)
		}
	})
}

func TestGetHierarchyPersistenceFailure(t *testing.T) {
	model, mock := newCourseNodeModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	expectCourseRead(mock, courseID, true)
	mock.ExpectQuery(`WITH RECURSIVE`).WithArgs(courseID).WillReturnError(errors.New("query failure"))
	_, err := model.GetHierarchy(courseID)
	if !errors.Is(err, ErrCourseNodePersistence) {
		t.Fatalf("error = %v, want persistence wrapper", err)
	}
}

func testCourseNodeTime() time.Time {
	return time.Date(2026, time.July, 25, 13, 47, 7, 0, time.UTC)
}
