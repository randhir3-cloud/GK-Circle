package v1

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
)

type courseNodeHTTPEnv struct {
	app  *fiber.App
	mock sqlmock.Sqlmock
}

func newCourseNodeHTTPEnv(t *testing.T, auth fiber.Handler) *courseNodeHTTPEnv {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	logger := zap.NewNop()
	cfg := config.AppConfig{
		Quiz: config.QuizConfig{PublicQuizAdminEmails: []string{adminEmail}},
	}
	db := goqu.New("postgres", sqlDB)
	courseCtrl, err := InitCourseController(db, logger, &cfg)
	if err != nil {
		t.Fatalf("init course controller: %v", err)
	}
	nodeCtrl, err := InitCourseNodeController(db, logger, &cfg)
	if err != nil {
		t.Fatalf("init node controller: %v", err)
	}

	app := fiber.New()
	admin := app.Group("/api/v1/admin")
	admin.Use(auth)
	courses := admin.Group("/courses")
	courses.Post("/", courseCtrl.CreateCourse)
	courses.Get("/", courseCtrl.ListCourses)

	nodes := courses.Group("/:" + constants.CourseId + "/nodes")
	nodes.Post("/", nodeCtrl.CreateNode)
	nodes.Get("/", nodeCtrl.ListRoots)
	nodes.Get("/tree", nodeCtrl.GetTree)
	nodes.Post("/reorder", nodeCtrl.ReorderChildren)
	nodes.Get("/:"+constants.NodeId, nodeCtrl.GetByID)
	nodes.Get("/:"+constants.NodeId+"/children", nodeCtrl.ListChildren)
	nodes.Patch("/:"+constants.NodeId+"/move", nodeCtrl.MoveNode)
	nodes.Delete("/:"+constants.NodeId, nodeCtrl.DeleteSubtree)

	courses.Get("/:"+constants.CourseId, courseCtrl.GetCourse)
	courses.Patch("/:"+constants.CourseId, courseCtrl.UpdateCourse)

	return &courseNodeHTTPEnv{app: app, mock: mock}
}

func nodesBase(courseID uuid.UUID) string {
	return "/api/v1/admin/courses/" + courseID.String() + "/nodes"
}

func expectCourseExists(mock sqlmock.Sqlmock, courseID uuid.UUID, exists bool) {
	rows := sqlmock.NewRows([]string{"id"})
	if exists {
		rows.AddRow(courseID)
	}
	mock.ExpectQuery(`SELECT "id" FROM "courses".*LIMIT`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(rows)
}

func expectCourseLock(mock sqlmock.Sqlmock, courseID uuid.UUID, exists bool) {
	rows := sqlmock.NewRows([]string{"id"})
	if exists {
		rows.AddRow(courseID)
	}
	mock.ExpectQuery(`SELECT "id" FROM "courses".*FOR UPDATE`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(rows)
}

func nodeRow(nodeID, courseID uuid.UUID, parentID interface{}, nodeType models.CourseNodeType, title string, position int, now time.Time) *sqlmock.Rows {
	path := nodeID.String()
	if parent, ok := parentID.(uuid.UUID); ok {
		path = parent.String() + "/" + nodeID.String()
	}
	return sqlmock.NewRows([]string{
		"id", "course_id", "parent_id", "node_type", "title", "position",
		"path", "status", "created_at", "updated_at",
	}).AddRow(nodeID, courseID, parentID, nodeType, title, position, path, models.CourseStatusDraft, now, now)
}

func TestAdminCourseNodeRoutesUnauthenticated(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, kratosAuthMiddleware(t))
	courseID := uuid.New()
	resp, payload := doJSON(t, env.app, http.MethodGet, nodesBase(courseID), nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
}

func TestAdminCourseNodeRoutesNonAdminForbidden(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.New()
	resp, payload := doJSON(t, env.app, http.MethodGet, nodesBase(courseID), nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrCourseAdminForbidden {
		t.Fatalf("forbidden message = %v", payload["data"])
	}
}

func TestAdminCourseNodeCreateRootAndReads(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	now := time.Date(2026, time.July, 26, 8, 0, 0, 0, time.UTC)

	env.mock.ExpectBegin()
	expectCourseLock(env.mock, courseID, true)
	env.mock.ExpectQuery(`INSERT INTO "course_nodes".*RETURNING`).
		WithArgs(
			courseID,
			sqlmock.AnyArg(),
			models.CourseNodeTypeSection,
			nil,
			sqlmock.AnyArg(),
			0,
			models.CourseStatusDraft,
			"Foundation",
		).
		WillReturnRows(nodeRow(nodeID, courseID, nil, models.CourseNodeTypeSection, "Foundation", 0, now))
	env.mock.ExpectCommit()

	resp, payload := doJSON(t, env.app, http.MethodPost, nodesBase(courseID), map[string]any{
		"title":     "Foundation",
		"node_type": string(models.CourseNodeTypeSection),
		"position":  0,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	if data["id"] != nodeID.String() || data["course_id"] != courseID.String() {
		t.Fatalf("create data = %v", data)
	}
	if _, hasPath := data["path"]; hasPath {
		t.Fatal("path must not be exposed")
	}
	if data["parent_id"] != nil {
		t.Fatalf("parent_id = %v, want null", data["parent_id"])
	}

	expectCourseExists(env.mock, courseID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "course_nodes".*"parent_id" IS NULL`).
		WithArgs(courseID).
		WillReturnRows(nodeRow(nodeID, courseID, nil, models.CourseNodeTypeSection, "Foundation", 0, now))

	resp, payload = doJSON(t, env.app, http.MethodGet, nodesBase(courseID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list roots status = %d payload=%v", resp.StatusCode, payload)
	}
	list, ok := payload["data"].([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("roots = %v", payload["data"])
	}

	env.mock.ExpectQuery(`SELECT .* FROM "course_nodes".*LIMIT`).
		WithArgs(courseID, nodeID, uint(1)).
		WillReturnRows(nodeRow(nodeID, courseID, nil, models.CourseNodeTypeSection, "Foundation", 0, now))

	resp, payload = doJSON(t, env.app, http.MethodGet, nodesBase(courseID)+"/"+nodeID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get by id status = %d payload=%v", resp.StatusCode, payload)
	}
}

func TestAdminCourseNodeCreateParentAndPositionRules(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	parentID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	base := nodesBase(courseID)

	t.Run("omitted position", func(t *testing.T) {
		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title":     "Topic",
			"node_type": string(models.CourseNodeTypeTopic),
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
	})

	t.Run("negative position", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title":     "Topic",
			"node_type": string(models.CourseNodeTypeTopic),
			"position":  -1,
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("null position", func(t *testing.T) {
		raw := []byte(`{"title":"Topic","node_type":"TOPIC","position":null}`)
		req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		resp, err := env.app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("empty parent_id", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title":     "Topic",
			"node_type": string(models.CourseNodeTypeTopic),
			"position":  0,
			"parent_id": "",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("whitespace parent_id", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title":     "Topic",
			"node_type": string(models.CourseNodeTypeTopic),
			"position":  0,
			"parent_id": "   ",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("malformed parent_id", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title":     "Topic",
			"node_type": string(models.CourseNodeTypeTopic),
			"position":  0,
			"parent_id": "not-a-uuid",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("null parent creates root", func(t *testing.T) {
		nodeID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
		now := time.Date(2026, time.July, 26, 8, 0, 0, 0, time.UTC)
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, true)
		env.mock.ExpectQuery(`INSERT INTO "course_nodes".*RETURNING`).
			WithArgs(
				courseID,
				sqlmock.AnyArg(),
				models.CourseNodeTypeSection,
				nil,
				sqlmock.AnyArg(),
				0,
				models.CourseStatusDraft,
				"Root",
			).
			WillReturnRows(nodeRow(nodeID, courseID, nil, models.CourseNodeTypeSection, "Root", 0, now))
		env.mock.ExpectCommit()

		raw := []byte(`{"title":"Root","node_type":"SECTION","position":0,"parent_id":null}`)
		req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		resp, err := env.app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("status = %d body=%s", resp.StatusCode, body)
		}
	})

	t.Run("child with parent", func(t *testing.T) {
		nodeID := uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a")
		now := time.Date(2026, time.July, 26, 8, 0, 0, 0, time.UTC)
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "path" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(parentID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id", "path"}).
				AddRow(parentID, courseID, parentID.String()))
		env.mock.ExpectQuery(`INSERT INTO "course_nodes".*RETURNING`).
			WithArgs(
				courseID,
				sqlmock.AnyArg(),
				models.CourseNodeTypeTopic,
				parentID,
				sqlmock.AnyArg(),
				1,
				models.CourseStatusDraft,
				"Child",
			).
			WillReturnRows(nodeRow(nodeID, courseID, parentID, models.CourseNodeTypeTopic, "Child", 1, now))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title":     "Child",
			"node_type": string(models.CourseNodeTypeTopic),
			"position":  1,
			"parent_id": parentID.String(),
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["parent_id"] != parentID.String() {
			t.Fatalf("parent_id = %v", data["parent_id"])
		}
	})

	t.Run("cross-course parent create", func(t *testing.T) {
		otherCourse := uuid.MustParse("019c01c7-7057-7b53-bd20-23081e70482f")
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "path" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(parentID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id", "path"}).
				AddRow(parentID, otherCourse, parentID.String()))
		env.mock.ExpectRollback()

		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"title":     "Child",
			"node_type": string(models.CourseNodeTypeTopic),
			"position":  0,
			"parent_id": parentID.String(),
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
}

func TestAdminCourseNodeBodyInjectionIgnored(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	attackerCourse := uuid.MustParse("019c01c7-7057-7b53-bd20-23081e70482f")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	now := time.Date(2026, time.July, 26, 8, 0, 0, 0, time.UTC)

	env.mock.ExpectBegin()
	expectCourseLock(env.mock, courseID, true)
	env.mock.ExpectQuery(`INSERT INTO "course_nodes".*RETURNING`).
		WithArgs(
			courseID,
			sqlmock.AnyArg(),
			models.CourseNodeTypeTopic,
			nil,
			sqlmock.AnyArg(),
			0,
			models.CourseStatusDraft,
			"Topic",
		).
		WillReturnRows(nodeRow(nodeID, courseID, nil, models.CourseNodeTypeTopic, "Topic", 0, now))
	env.mock.ExpectCommit()

	resp, payload := doJSON(t, env.app, http.MethodPost, nodesBase(courseID), map[string]any{
		"title":     "Topic",
		"node_type": string(models.CourseNodeTypeTopic),
		"position":  0,
		"course_id": attackerCourse.String(),
		"owner_id":  "attacker-owner-id",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	if data["course_id"] != courseID.String() {
		t.Fatalf("course_id = %v, want route course", data["course_id"])
	}
	if _, hasOwner := data["owner_id"]; hasOwner {
		t.Fatal("owner_id must not appear on CourseNode responses")
	}
}

func TestAdminCourseNodeScopedReadsAndMissingCourse(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseA := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	courseBNode := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	now := time.Date(2026, time.July, 26, 8, 0, 0, 0, time.UTC)

	t.Run("cross-course get by id", func(t *testing.T) {
		env.mock.ExpectQuery(`SELECT .* FROM "course_nodes".*LIMIT`).
			WithArgs(courseA, courseBNode, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}))
		resp, _ := doJSON(t, env.app, http.MethodGet, nodesBase(courseA)+"/"+courseBNode.String(), nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("cross-course children", func(t *testing.T) {
		otherCourse := uuid.MustParse("019c01c7-7057-7b53-bd20-23081e70482f")
		expectCourseExists(env.mock, courseA, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes".*LIMIT`).
			WithArgs(courseBNode, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(courseBNode, otherCourse))
		resp, _ := doJSON(t, env.app, http.MethodGet, nodesBase(courseA)+"/"+courseBNode.String()+"/children", nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("missing course list roots", func(t *testing.T) {
		expectCourseExists(env.mock, courseA, false)
		resp, _ := doJSON(t, env.app, http.MethodGet, nodesBase(courseA), nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("empty tree for existing course", func(t *testing.T) {
		expectCourseExists(env.mock, courseA, true)
		env.mock.ExpectQuery(regexp.QuoteMeta("WITH RECURSIVE course_rows AS (")).
			WithArgs(courseA).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at", "total_nodes",
			}).AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, int64(0)))

		resp, payload := doJSON(t, env.app, http.MethodGet, nodesBase(courseA)+"/tree", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		if data["course_id"] != courseA.String() {
			t.Fatalf("course_id = %v", data["course_id"])
		}
		roots, ok := data["roots"].([]any)
		if !ok || len(roots) != 0 {
			t.Fatalf("roots = %#v, want empty slice", data["roots"])
		}
	})

	t.Run("list children success", func(t *testing.T) {
		parentID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
		childID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
		expectCourseExists(env.mock, courseA, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes".*LIMIT`).
			WithArgs(parentID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(parentID, courseA))
		env.mock.ExpectQuery(`SELECT .* FROM "course_nodes".*ORDER BY "position" ASC, "id" ASC`).
			WithArgs(courseA, parentID).
			WillReturnRows(nodeRow(childID, courseA, parentID, models.CourseNodeTypeTopic, "Child", 0, now))

		resp, payload := doJSON(t, env.app, http.MethodGet, nodesBase(courseA)+"/"+parentID.String()+"/children", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
		list, ok := payload["data"].([]any)
		if !ok || len(list) != 1 {
			t.Fatalf("children = %v", payload["data"])
		}
	})
}

func TestAdminCourseNodeMutationRoutesAbsent(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.New()
	nodeID := uuid.New()
	base := nodesBase(courseID)

	// Still intentionally absent after T10. Do not assert absence of
	// PATCH .../move, POST /reorder, or DELETE /:node_id.
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, base + "/" + nodeID.String() + "/move"},
		{http.MethodDelete, base + "/" + nodeID.String() + "/subtree"},
		{http.MethodPatch, base + "/" + nodeID.String()},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			resp, err := env.app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want 404 or 405", resp.StatusCode)
			}
		})
	}
}

func TestAdminCourseNodeInvalidIDs(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	resp, _ := doJSON(t, env.app, http.MethodGet, "/api/v1/admin/courses/not-a-uuid/nodes", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	courseID := uuid.New()
	resp, _ = doJSON(t, env.app, http.MethodGet, nodesBase(courseID)+"/not-a-uuid", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestValidateCreateAdminCourseNodeUnit(t *testing.T) {
	courseID := uuid.New()

	_, err := validateCreateAdminCourseNode(courseID, structs.ReqCreateAdminCourseNode{
		Title:    "",
		NodeType: string(models.CourseNodeTypeSection),
		Position: structs.OptionalInteger{Present: true, Value: 0},
	})
	if err == nil {
		t.Fatal("expected title error")
	}

	_, err = validateCreateAdminCourseNode(courseID, structs.ReqCreateAdminCourseNode{
		Title:    "Title",
		NodeType: "UNKNOWN",
		Position: structs.OptionalInteger{Present: true, Value: 0},
	})
	if err == nil {
		t.Fatal("expected type error")
	}

	params, err := validateCreateAdminCourseNode(courseID, structs.ReqCreateAdminCourseNode{
		Title:    " Topic ",
		NodeType: string(models.CourseNodeTypeTopic),
		Position: structs.OptionalInteger{Present: true, Value: 0},
	})
	if err != nil {
		t.Fatalf("valid create: %v", err)
	}
	if params.Title != "Topic" || params.NodeType != models.CourseNodeTypeTopic || params.ParentID.Valid {
		t.Fatalf("params = %+v", params)
	}
}

func TestAdminCourseNodeMutationAuth(t *testing.T) {
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	base := nodesBase(courseID)

	t.Run("unauthenticated move", func(t *testing.T) {
		env := newCourseNodeHTTPEnv(t, kratosAuthMiddleware(t))
		resp, _ := doJSON(t, env.app, http.MethodPatch, base+"/"+nodeID.String()+"/move", map[string]any{"position": 0})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("non-admin reorder", func(t *testing.T) {
		env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		resp, payload := doJSON(t, env.app, http.MethodPost, base+"/reorder", map[string]any{
			"ordered_node_ids": []string{nodeID.String()},
		})
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}
	})
	t.Run("non-admin delete", func(t *testing.T) {
		env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		resp, _ := doJSON(t, env.app, http.MethodDelete, base+"/"+nodeID.String(), nil)
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
}

func TestAdminCourseNodeMoveValidationAndSuccess(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	path := nodeID.String()
	base := nodesBase(courseID) + "/" + nodeID.String() + "/move"

	t.Run("omitted position", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPatch, base, map[string]any{})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("null position", func(t *testing.T) {
		raw := []byte(`{"position":null}`)
		req := httptest.NewRequest(http.MethodPatch, base, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		resp, err := env.app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("negative position", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPatch, base, map[string]any{"position": -1})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("empty new_parent_id", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPatch, base, map[string]any{
			"position":      0,
			"new_parent_id": "",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("malformed new_parent_id", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPatch, base, map[string]any{
			"position":      0,
			"new_parent_id": "not-a-uuid",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("missing parent 404", func(t *testing.T) {
		parentID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnRows(nodeRow(nodeID, courseID, nil, models.CourseNodeTypeSection, "Node", 0, time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC)))
		env.mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(parentID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}))
		env.mock.ExpectRollback()

		resp, _ := doJSON(t, env.app, http.MethodPatch, base, map[string]any{
			"position":      0,
			"new_parent_id": parentID.String(),
		})
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("root noop move then get tree", func(t *testing.T) {
		now := time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC)
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}).AddRow(nodeID, courseID, nil, models.CourseNodeTypeSection, "Node", 0, path, models.CourseStatusDraft, now, now))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPatch, base, map[string]any{"position": 0})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("move status = %d payload=%v", resp.StatusCode, payload)
		}
		if payload["data"] != "success" {
			t.Fatalf("data = %v", payload["data"])
		}

		expectCourseExists(env.mock, courseID, true)
		env.mock.ExpectQuery(regexp.QuoteMeta("WITH RECURSIVE course_rows AS (")).
			WithArgs(courseID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at", "total_nodes",
			}).AddRow(nodeID, courseID, nil, models.CourseNodeTypeSection, "Node", 0, path, models.CourseStatusDraft, now, now, int64(1)))

		resp, payload = doJSON(t, env.app, http.MethodGet, nodesBase(courseID)+"/tree", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("tree status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		roots, _ := data["roots"].([]any)
		if len(roots) != 1 {
			t.Fatalf("roots after move = %#v", data["roots"])
		}
	})
}

func TestAdminCourseNodeReorderValidationAndSuccess(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	aID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	bID := uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a")
	parentID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	base := nodesBase(courseID) + "/reorder"
	now := time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC)

	t.Run("missing array", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("null array", func(t *testing.T) {
		raw := []byte(`{"ordered_node_ids":null}`)
		req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		resp, err := env.app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("empty array", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_node_ids": []string{},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("duplicate ids", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_node_ids": []string{aID.String(), aID.String()},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("empty string entry", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_node_ids": []string{""},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("whitespace entry", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_node_ids": []string{"   "},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("malformed uuid", func(t *testing.T) {
		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_node_ids": []string{"not-a-uuid"},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("sibling mismatch", func(t *testing.T) {
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(courseID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}).AddRow(aID, courseID, nil, models.CourseNodeTypeSection, "A", 0, aID.String(), models.CourseStatusDraft, now, now).
				AddRow(bID, courseID, nil, models.CourseNodeTypeSection, "B", 1, bID.String(), models.CourseStatusDraft, now, now))
		env.mock.ExpectRollback()

		resp, _ := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"ordered_node_ids": []string{aID.String()},
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("valid child reorder then get children", func(t *testing.T) {
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(parentID, uint(1)).
			WillReturnRows(nodeRow(parentID, courseID, nil, models.CourseNodeTypeSection, "Parent", 0, now))
		env.mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(courseID, parentID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}).AddRow(aID, courseID, parentID, models.CourseNodeTypeTopic, "A", 0, parentID.String()+"/"+aID.String(), models.CourseStatusDraft, now, now))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodPost, base, map[string]any{
			"parent_id":        parentID.String(),
			"ordered_node_ids": []string{aID.String()},
		})
		if resp.StatusCode != http.StatusOK || payload["data"] != "success" {
			t.Fatalf("status = %d payload=%v", resp.StatusCode, payload)
		}

		expectCourseExists(env.mock, courseID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes".*LIMIT`).
			WithArgs(parentID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(parentID, courseID))
		env.mock.ExpectQuery(`SELECT .* FROM "course_nodes".*ORDER BY "position" ASC, "id" ASC`).
			WithArgs(courseID, parentID).
			WillReturnRows(nodeRow(aID, courseID, parentID, models.CourseNodeTypeTopic, "A", 0, now))

		resp, payload = doJSON(t, env.app, http.MethodGet, nodesBase(courseID)+"/"+parentID.String()+"/children", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("children status = %d payload=%v", resp.StatusCode, payload)
		}
		list, ok := payload["data"].([]any)
		if !ok || len(list) != 1 {
			t.Fatalf("children = %v", payload["data"])
		}
	})
}

func TestAdminCourseNodeDeleteAndTree(t *testing.T) {
	env := newCourseNodeHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	rootID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	childID := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	childPath := rootID.String() + "/" + childID.String()
	now := time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC)

	t.Run("missing node", func(t *testing.T) {
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(courseID, childID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}))
		env.mock.ExpectRollback()

		resp, _ := doJSON(t, env.app, http.MethodDelete, nodesBase(courseID)+"/"+childID.String(), nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("missing course", func(t *testing.T) {
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, false)
		env.mock.ExpectRollback()

		resp, _ := doJSON(t, env.app, http.MethodDelete, nodesBase(courseID)+"/"+childID.String(), nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})
	t.Run("delete leaf then get tree", func(t *testing.T) {
		env.mock.ExpectBegin()
		expectCourseLock(env.mock, courseID, true)
		env.mock.ExpectQuery(`SELECT "id", "course_id", "parent_id", "node_type", "title", "position", "path", "status", "created_at", "updated_at" FROM "course_nodes".*FOR UPDATE`).
			WithArgs(courseID, childID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}).AddRow(childID, courseID, rootID, models.CourseNodeTypeTopic, "Leaf", 0, childPath, models.CourseStatusDraft, now, now))
		env.mock.ExpectQuery(regexp.QuoteMeta("SELECT id, course_id, parent_id, node_type, title, position, path, status, created_at, updated_at\nFROM course_nodes\nWHERE course_id = $1 AND (path = $2 OR path LIKE $3)\nORDER BY path ASC\nFOR UPDATE")).
			WithArgs(courseID, childPath, childPath+"/%").
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at",
			}).AddRow(childID, courseID, rootID, models.CourseNodeTypeTopic, "Leaf", 0, childPath, models.CourseStatusDraft, now, now))
		env.mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM course_nodes\nWHERE course_id = $1 AND (path = $2 OR path LIKE $3)\nRETURNING id")).
			WithArgs(courseID, childPath, childPath+"/%").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(childID))
		env.mock.ExpectCommit()

		resp, payload := doJSON(t, env.app, http.MethodDelete, nodesBase(courseID)+"/"+childID.String(), nil)
		if resp.StatusCode != http.StatusOK || payload["data"] != "success" {
			t.Fatalf("delete status = %d payload=%v", resp.StatusCode, payload)
		}

		expectCourseExists(env.mock, courseID, true)
		env.mock.ExpectQuery(regexp.QuoteMeta("WITH RECURSIVE course_rows AS (")).
			WithArgs(courseID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "parent_id", "node_type", "title", "position",
				"path", "status", "created_at", "updated_at", "total_nodes",
			}).AddRow(rootID, courseID, nil, models.CourseNodeTypeSection, "Root", 0, rootID.String(), models.CourseStatusDraft, now, now, int64(1)))

		resp, payload = doJSON(t, env.app, http.MethodGet, nodesBase(courseID)+"/tree", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("tree status = %d payload=%v", resp.StatusCode, payload)
		}
		data, _ := payload["data"].(map[string]any)
		roots, _ := data["roots"].([]any)
		if len(roots) != 1 {
			t.Fatalf("roots after delete = %#v", data["roots"])
		}
	})
}

func TestParseOrderedNodeIDsUnit(t *testing.T) {
	if _, err := parseOrderedNodeIDs(nil); err == nil {
		t.Fatal("nil should fail")
	}
	if _, err := parseOrderedNodeIDs([]string{}); err == nil {
		t.Fatal("empty should fail")
	}
	id := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	parsed, err := parseOrderedNodeIDs([]string{id.String()})
	if err != nil || len(parsed) != 1 || parsed[0] != id {
		t.Fatalf("parsed=%v err=%v", parsed, err)
	}
}
