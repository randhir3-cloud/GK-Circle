package models

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// learningItemOwnershipChain returns 10 nested CourseNode IDs (depth 1..10).
// Depth is conceptual only — LearningItem SQL scopes by course_id + course_node_id.
func learningItemOwnershipChain() (courseID uuid.UUID, chain []uuid.UUID) {
	courseID = uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	chain = make([]uuid.UUID, 10)
	for i := 0; i < 10; i++ {
		chain[i] = uuid.MustParse(fmt.Sprintf("019c02b0-aaaa-7000-8000-%012d", i+1))
	}
	return courseID, chain
}

func assertLearningItemOwnershipSQLHasNoDepthPredicates(t *testing.T, query string) {
	t.Helper()
	lower := strings.ToLower(query)
	for _, banned := range []string{"max_depth", "max_*_depth", "nesting", `"path"`, " path ", "depth"} {
		if strings.Contains(lower, banned) {
			t.Fatalf("ownership SQL must not reference depth/path predicates; matched %q in: %s", banned, query)
		}
	}
}

func expectLearningItemOwnershipCreate(
	t *testing.T,
	mock sqlmock.Sqlmock,
	courseID, nodeID, itemID uuid.UUID,
	title string,
	publishState LearningItemPublishState,
) {
	t.Helper()
	metadata := []byte(`{"version":1,"blocks":[]}`)
	mock.ExpectBegin()
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemMaxPosition(mock, nodeID, nil)
	mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
		WithArgs(
			courseID,
			nodeID,
			sqlmock.AnyArg(),
			itemID,
			LearningItemTypeArticle,
			metadata,
			0,
			publishState,
			title,
		).
		WillReturnRows(learningItemRowsWithPublishState(
			itemID, courseID, nodeID, title, LearningItemTypeArticle,
			nil, metadata, 0, publishState,
		))
	mock.ExpectCommit()
}

func expectLearningItemOwnershipList(
	t *testing.T,
	mock sqlmock.Sqlmock,
	courseID, nodeID, itemID uuid.UUID,
	title string,
) {
	t.Helper()
	metadata := []byte(`{"version":1,"blocks":[]}`)
	expectLearningItemNodeLookup(mock, courseID, nodeID, true)
	mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`).
		WithArgs(courseID, nodeID).
		WillReturnRows(learningItemRows(
			itemID, courseID, nodeID, title, LearningItemTypeArticle,
			nil, metadata, 0,
		))
}

func TestLearningItemOwnershipAttachListGetAcrossTenNestingLevels(t *testing.T) {
	courseID, chain := learningItemOwnershipChain()
	// Exercise every depth 1..10 to prove there is no LearningItem depth ceiling.
	for depth, nodeID := range chain {
		depth := depth + 1
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			model, mock := newLearningItemModelTest(t)
			itemID := uuid.MustParse(fmt.Sprintf("019c02c0-bbbb-7000-8000-%012d", depth))
			model.newUUID = func() (uuid.UUID, error) { return itemID, nil }
			title := fmt.Sprintf("Depth %d Lesson", depth)

			expectLearningItemOwnershipCreate(t, mock, courseID, nodeID, itemID, title, LearningItemPublishStateDraft)
			item, err := model.CreateLearningItem(CreateLearningItemParams{
				CourseID:     courseID,
				CourseNodeID: nodeID,
				Title:        title,
				ItemType:     LearningItemTypeArticle,
			})
			if err != nil {
				t.Fatalf("create at depth %d: %v", depth, err)
			}
			if item.CourseNodeID != nodeID || item.CourseID != courseID || item.Position != 0 {
				t.Fatalf("create ownership mismatch at depth %d: %+v", depth, item)
			}

			expectLearningItemOwnershipList(t, mock, courseID, nodeID, itemID, title)
			listed, err := model.ListLearningItemsByNode(courseID, nodeID)
			if err != nil {
				t.Fatalf("list at depth %d: %v", depth, err)
			}
			if len(listed) != 1 || listed[0].ID != itemID || listed[0].CourseNodeID != nodeID {
				t.Fatalf("list at depth %d = %#v", depth, listed)
			}

			metadata := []byte(`{"version":1,"blocks":[]}`)
			mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
				WithArgs(courseID, nodeID, itemID, uint(1)).
				WillReturnRows(learningItemRows(
					itemID, courseID, nodeID, title, LearningItemTypeArticle,
					nil, metadata, 0,
				))
			got, err := model.GetLearningItemByID(courseID, nodeID, itemID)
			if err != nil {
				t.Fatalf("get at depth %d: %v", depth, err)
			}
			if got.ID != itemID || got.CourseNodeID != nodeID {
				t.Fatalf("get at depth %d = %+v", depth, got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLearningItemOwnershipRootVsDeepParity(t *testing.T) {
	courseID, chain := learningItemOwnershipChain()
	rootNodeID := chain[0]  // depth 1
	deepNodeID := chain[9]  // depth 10
	otherCourseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e482")
	foreignTitle := "Secret Foreign Course Title Never Leak"

	type nodeCase struct {
		name   string
		nodeID uuid.UUID
		depth  int
	}
	cases := []nodeCase{
		{name: "root_depth_1", nodeID: rootNodeID, depth: 1},
		{name: "deep_depth_10", nodeID: deepNodeID, depth: 10},
	}

	for _, nc := range cases {
		t.Run(nc.name+"_success_create_list_get", func(t *testing.T) {
			model, mock := newLearningItemModelTest(t)
			itemID := uuid.MustParse(fmt.Sprintf("019c02d0-cccc-7000-8000-%012d", nc.depth))
			model.newUUID = func() (uuid.UUID, error) { return itemID, nil }
			title := fmt.Sprintf("Parity Lesson depth %d", nc.depth)

			expectLearningItemOwnershipCreate(t, mock, courseID, nc.nodeID, itemID, title, LearningItemPublishStateDraft)
			created, err := model.CreateLearningItem(CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nc.nodeID, Title: title, ItemType: LearningItemTypeArticle,
			})
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if created.CourseNodeID != nc.nodeID || created.CourseID != courseID {
				t.Fatalf("create ownership: %+v", created)
			}

			expectLearningItemOwnershipList(t, mock, courseID, nc.nodeID, itemID, title)
			listed, err := model.ListLearningItemsByNode(courseID, nc.nodeID)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(listed) != 1 || listed[0].ID != itemID {
				t.Fatalf("list = %#v", listed)
			}

			metadata := []byte(`{"version":1,"blocks":[]}`)
			mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
				WithArgs(courseID, nc.nodeID, itemID, uint(1)).
				WillReturnRows(learningItemRows(
					itemID, courseID, nc.nodeID, title, LearningItemTypeArticle,
					nil, metadata, 0,
				))
			got, err := model.GetLearningItemByID(courseID, nc.nodeID, itemID)
			if err != nil {
				t.Fatalf("get: %v", err)
			}
			if got.ID != itemID {
				t.Fatalf("get = %+v", got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run(nc.name+"_cross_course_create", func(t *testing.T) {
			model, mock := newLearningItemModelTest(t)
			mock.ExpectBegin()
			expectLearningItemNodeLock(mock, courseID, nc.nodeID, false)
			mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
				WithArgs(nc.nodeID, uint(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(nc.nodeID, otherCourseID))
			mock.ExpectRollback()

			_, err := model.CreateLearningItem(CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nc.nodeID, Title: "Item", ItemType: LearningItemTypeVideo,
			})
			if !errors.Is(err, ErrLearningItemCrossCourse) {
				t.Fatalf("error = %v, want %v", err, ErrLearningItemCrossCourse)
			}
			if strings.Contains(err.Error(), foreignTitle) {
				t.Fatalf("cross-course error leaked foreign title payload: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run(nc.name+"_cross_course_list", func(t *testing.T) {
			model, mock := newLearningItemModelTest(t)
			expectLearningItemNodeLookup(mock, courseID, nc.nodeID, false)
			mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
				WithArgs(nc.nodeID, uint(1)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(nc.nodeID, otherCourseID))

			items, err := model.ListLearningItemsByNode(courseID, nc.nodeID)
			if !errors.Is(err, ErrLearningItemCrossCourse) {
				t.Fatalf("error = %v, want %v", err, ErrLearningItemCrossCourse)
			}
			if items != nil {
				t.Fatalf("list must not return items on cross-course: %#v", items)
			}
			if strings.Contains(err.Error(), foreignTitle) {
				t.Fatalf("cross-course list error leaked foreign title: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run(nc.name+"_wrong_course_get_not_found", func(t *testing.T) {
			model, mock := newLearningItemModelTest(t)
			itemID := uuid.MustParse(fmt.Sprintf("019c02d1-dddd-7000-8000-%012d", nc.depth))
			mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
				WithArgs(otherCourseID, nc.nodeID, itemID, uint(1)).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "course_id", "course_node_id", "title", "item_type",
					"description", "metadata", "position", "publish_state", "created_at", "updated_at",
				}))
			item, err := model.GetLearningItemByID(otherCourseID, nc.nodeID, itemID)
			if !errors.Is(err, ErrLearningItemNotFound) {
				t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
			}
			if item.Title != "" || item.ID != uuid.Nil {
				t.Fatalf("get must not leak item payload: %+v", item)
			}
			if strings.Contains(err.Error(), foreignTitle) {
				t.Fatalf("not-found error leaked foreign title: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run(nc.name+"_wrong_node_get_not_found", func(t *testing.T) {
			model, mock := newLearningItemModelTest(t)
			itemID := uuid.MustParse(fmt.Sprintf("019c02d2-eeee-7000-8000-%012d", nc.depth))
			wrongNode := uuid.MustParse("019c02b0-ffff-7000-8000-000000000099")
			mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
				WithArgs(courseID, wrongNode, itemID, uint(1)).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "course_id", "course_node_id", "title", "item_type",
					"description", "metadata", "position", "publish_state", "created_at", "updated_at",
				}))
			item, err := model.GetLearningItemByID(courseID, wrongNode, itemID)
			if !errors.Is(err, ErrLearningItemNotFound) {
				t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
			}
			if item.ID != uuid.Nil {
				t.Fatalf("get must not leak item: %+v", item)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLearningItemOwnershipDeepPublishedReads(t *testing.T) {
	courseID, chain := learningItemOwnershipChain()
	deepNodeID := chain[9]
	itemPublished := uuid.MustParse("019c02e0-1111-7000-8000-000000000001")
	itemDraft := uuid.MustParse("019c02e0-1111-7000-8000-000000000002")
	metadata := []byte(`{"version":1,"blocks":[]}`)

	t.Run("list_published_on_depth_10", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, deepNodeID, true)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, deepNodeID, LearningItemPublishStatePublished).
			WillReturnRows(learningItemRowsWithPublishState(
				itemPublished, courseID, deepNodeID, "Deep Published", LearningItemTypeArticle,
				nil, metadata, 0, LearningItemPublishStatePublished,
			))
		items, err := model.ListPublishedLearningItemsByNode(courseID, deepNodeID)
		if err != nil {
			t.Fatalf("list published: %v", err)
		}
		if len(items) != 1 || items[0].ID != itemPublished || items[0].CourseNodeID != deepNodeID {
			t.Fatalf("list published = %#v", items)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("get_published_on_depth_10", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, deepNodeID, itemPublished, LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowsWithPublishState(
				itemPublished, courseID, deepNodeID, "Deep Published", LearningItemTypeArticle,
				nil, metadata, 0, LearningItemPublishStatePublished,
			))
		item, err := model.GetPublishedLearningItemByID(courseID, deepNodeID, itemPublished)
		if err != nil {
			t.Fatalf("get published: %v", err)
		}
		if item.PublishState != LearningItemPublishStatePublished {
			t.Fatalf("publish_state = %q", item.PublishState)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("get_draft_on_depth_10_not_found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, deepNodeID, itemDraft, LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		_, err := model.GetPublishedLearningItemByID(courseID, deepNodeID, itemDraft)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestLearningItemOwnershipMissingDeepNode(t *testing.T) {
	courseID, chain := learningItemOwnershipChain()
	deepNodeID := chain[9]

	t.Run("list_missing_deep_node", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, deepNodeID, false)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(deepNodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		_, err := model.ListLearningItemsByNode(courseID, deepNodeID)
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNodeNotFound)
		}
	})

	t.Run("create_missing_deep_node", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemNodeLock(mock, courseID, deepNodeID, false)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(deepNodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		mock.ExpectRollback()
		_, err := model.CreateLearningItem(CreateLearningItemParams{
			CourseID: courseID, CourseNodeID: deepNodeID, Title: "Item", ItemType: LearningItemTypeLink,
		})
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNodeNotFound)
		}
	})
}

func TestLearningItemOwnershipSQLShapeNoDepthPredicates(t *testing.T) {
	// Guard future regressions: ownership helpers must stay depth-agnostic.
	queries := []string{
		`SELECT "id", "course_id" FROM "course_nodes" WHERE ("course_id" = $1 AND "id" = $2) FOR UPDATE`,
		`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items" WHERE ("course_id" = $1 AND "course_node_id" = $2) ORDER BY "position" ASC, "id" ASC`,
		`INSERT INTO "learning_items" ("course_id", "course_node_id", "created_at", "description", "id", "item_type", "metadata", "position", "publish_state", "title", "updated_at") VALUES (...)`,
	}
	for _, q := range queries {
		assertLearningItemOwnershipSQLHasNoDepthPredicates(t, q)
	}
}
