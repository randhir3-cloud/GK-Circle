package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func newLearningItemModelTest(t *testing.T) (*LearningItemModel, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	return InitLearningItemModel(goqu.New("postgres", sqlDB)), mock
}

func learningItemRows(
	itemID, courseID, nodeID uuid.UUID,
	title string,
	itemType LearningItemType,
	description interface{},
	metadata []byte,
	position int,
) *sqlmock.Rows {
	return learningItemRowsWithPublishState(
		itemID, courseID, nodeID, title, itemType, description, metadata, position,
		LearningItemPublishStateDraft,
	)
}

func learningItemRowsWithPublishState(
	itemID, courseID, nodeID uuid.UUID,
	title string,
	itemType LearningItemType,
	description interface{},
	metadata []byte,
	position int,
	publishState LearningItemPublishState,
) *sqlmock.Rows {
	now := time.Date(2026, time.July, 26, 8, 30, 0, 0, time.UTC)
	return sqlmock.NewRows([]string{
		"id", "course_id", "course_node_id", "title", "item_type",
		"description", "metadata", "position", "publish_state", "created_at", "updated_at",
	}).AddRow(
		itemID, courseID, nodeID, title, itemType, description, metadata,
		position, publishState, now, now,
	)
}

func expectLearningItemNodeLock(mock sqlmock.Sqlmock, courseID, nodeID uuid.UUID, exists bool) {
	rows := sqlmock.NewRows([]string{"id", "course_id"})
	if exists {
		rows.AddRow(nodeID, courseID)
	}
	mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes".*FOR UPDATE`).
		WithArgs(courseID, nodeID, uint(1)).
		WillReturnRows(rows)
}

func expectLearningItemNodeLookup(mock sqlmock.Sqlmock, courseID, nodeID uuid.UUID, exists bool) {
	rows := sqlmock.NewRows([]string{"id", "course_id"})
	if exists {
		rows.AddRow(nodeID, courseID)
	}
	mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
		WithArgs(courseID, nodeID, uint(1)).
		WillReturnRows(rows)
}

func expectLearningItemMaxPosition(mock sqlmock.Sqlmock, nodeID uuid.UUID, max interface{}) {
	rows := sqlmock.NewRows([]string{"max"}).AddRow(max)
	mock.ExpectQuery(`SELECT MAX\("position"\) FROM "learning_items"`).
		WithArgs(nodeID, uint(1)).
		WillReturnRows(rows)
}

func TestInitLearningItemModel(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	if model == nil || model.db == nil || model.newUUID == nil {
		t.Fatal("learning item model was not fully initialized")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected SQL: %v", err)
	}
}

func TestCreateLearningItemValidation(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	quizID := uuid.MustParse("019c02a0-1111-7000-8000-000000000101")

	tests := []struct {
		name   string
		params CreateLearningItemParams
		want   error
	}{
		{
			name:   "title required",
			params: CreateLearningItemParams{CourseID: courseID, CourseNodeID: nodeID, Title: "  ", ItemType: LearningItemTypeArticle},
			want:   ErrLearningItemTitleRequired,
		},
		{
			name:   "type invalid",
			params: CreateLearningItemParams{CourseID: courseID, CourseNodeID: nodeID, Title: "Item", ItemType: "UNKNOWN"},
			want:   ErrLearningItemTypeInvalid,
		},
		{
			name: "quiz reference requires quiz",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Test",
				ItemType: LearningItemTypeQuizRef,
			},
			want: ErrLearningItemQuizRequired,
		},
		{
			name: "non quiz item rejects quiz",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Article",
				ItemType: LearningItemTypeArticle, QuizID: &quizID,
			},
			want: ErrLearningItemQuizForbidden,
		},
		{
			name: "metadata null invalid",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Item",
				ItemType: LearningItemTypeArticle, Metadata: json.RawMessage("null"),
			},
			want: ErrLearningItemMetadataInvalid,
		},
		{
			name: "metadata array invalid",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Item",
				ItemType: LearningItemTypeArticle, Metadata: json.RawMessage(`[]`),
			},
			want: ErrLearningItemMetadataInvalid,
		},
		{
			name: "metadata version invalid",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Item",
				ItemType: LearningItemTypeArticle, Metadata: json.RawMessage(`{"version":0,"blocks":[]}`),
			},
			want: ErrLearningItemMetadataVersionInvalid,
		},
		{
			name: "metadata block type invalid",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Item",
				ItemType: LearningItemTypeArticle,
				Metadata: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"WIDGET","data":{}}]}`),
			},
			want: ErrLearningItemBlockTypeInvalid,
		},
		{
			name: "metadata duplicate block id",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Item",
				ItemType: LearningItemTypeArticle,
				Metadata: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{}},{"id":"b1","type":"LINK","data":{}}]}`),
			},
			want: ErrLearningItemBlockDuplicate,
		},
		{
			name: "placeholder invalid",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Item",
				ItemType: LearningItemTypeArticle,
				Metadata: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"{{student-name}}"}}]}`),
			},
			want: ErrLearningItemPlaceholderInvalid,
		},
		{
			name: "placeholder syntax",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Item",
				ItemType: LearningItemTypeArticle,
				Metadata: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"{{name"}}]}`),
			},
			want: ErrLearningItemPlaceholderSyntax,
		},
		{
			name: "visibility mode invalid",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Item",
				ItemType: LearningItemTypeArticle,
				Metadata: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{},"visibility":{"mode":"PUBLIC"}}]}`),
			},
			want: ErrLearningItemVisibilityModeInvalid,
		},
		{
			name: "visibility null invalid",
			params: CreateLearningItemParams{
				CourseID: courseID, CourseNodeID: nodeID, Title: "Item",
				ItemType: LearningItemTypeArticle,
				Metadata: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{},"visibility":null}]}`),
			},
			want: ErrLearningItemVisibilityInvalid,
		},
		{
			name:   "nil node",
			params: CreateLearningItemParams{CourseID: courseID, Title: "Item", ItemType: LearningItemTypeArticle},
			want:   ErrLearningItemNodeNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := model.CreateLearningItem(test.params)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation unexpectedly queried database: %v", err)
	}
}

func TestCreateLearningItemAppendPositions(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")

	cases := []struct {
		name        string
		maxPosition interface{}
		wantPos     int
	}{
		{name: "first item", maxPosition: nil, wantPos: 0},
		{name: "append after zero", maxPosition: int64(0), wantPos: 1},
		{name: "append after one", maxPosition: int64(1), wantPos: 2},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			model, mock := newLearningItemModelTest(t)
			itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000001")
			model.newUUID = func() (uuid.UUID, error) { return itemID, nil }
			metadata := []byte(`{"version":1,"blocks":[]}`)

			mock.ExpectBegin()
			expectLearningItemNodeLock(mock, courseID, nodeID, true)
			expectLearningItemMaxPosition(mock, nodeID, test.maxPosition)
			mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
				WithArgs(
					courseID,
					nodeID,
					sqlmock.AnyArg(),
					itemID,
					LearningItemTypeArticle,
					metadata,
					test.wantPos,
					LearningItemPublishStateDraft,
					"Lesson",
				).
				WillReturnRows(learningItemRows(
					itemID, courseID, nodeID, "Lesson", LearningItemTypeArticle,
					nil, metadata, test.wantPos,
				))
			mock.ExpectCommit()

			item, err := model.CreateLearningItem(CreateLearningItemParams{
				CourseID:     courseID,
				CourseNodeID: nodeID,
				Title:        "Lesson",
				ItemType:     LearningItemTypeArticle,
			})
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if item.Position != test.wantPos {
				t.Fatalf("position = %d, want %d", item.Position, test.wantPos)
			}
			if item.PublishState != LearningItemPublishStateDraft {
				t.Fatalf("publish_state = %q, want DRAFT", item.PublishState)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet SQL expectations: %v", err)
			}
		})
	}
}

func TestCreateLearningItemNodeNotFoundAndCrossCourse(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	otherCourseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e482")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")

	t.Run("missing node", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemNodeLock(mock, courseID, nodeID, false)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		mock.ExpectRollback()

		_, err := model.CreateLearningItem(CreateLearningItemParams{
			CourseID: courseID, CourseNodeID: nodeID, Title: "Item", ItemType: LearningItemTypeVideo,
		})
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNodeNotFound)
		}
	})

	t.Run("cross course", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemNodeLock(mock, courseID, nodeID, false)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(nodeID, otherCourseID))
		mock.ExpectRollback()

		_, err := model.CreateLearningItem(CreateLearningItemParams{
			CourseID: courseID, CourseNodeID: nodeID, Title: "Item", ItemType: LearningItemTypePDF,
		})
		if !errors.Is(err, ErrLearningItemCrossCourse) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemCrossCourse)
		}
	})
}

func TestCreateLearningItemPositionConflict(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000002")
	model.newUUID = func() (uuid.UUID, error) { return itemID, nil }

	mock.ExpectBegin()
	expectLearningItemNodeLock(mock, courseID, nodeID, true)
	expectLearningItemMaxPosition(mock, nodeID, int64(0))
	mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
		WillReturnError(&pq.Error{Code: "23505", Constraint: learningItemsNodePositionConstraint})
	mock.ExpectRollback()

	_, err := model.CreateLearningItem(CreateLearningItemParams{
		CourseID: courseID, CourseNodeID: nodeID, Title: "Item", ItemType: LearningItemTypeLink,
	})
	if !errors.Is(err, ErrLearningItemConflict) {
		t.Fatalf("error = %v, want %v", err, ErrLearningItemConflict)
	}
}

func TestGetListUpdateDeleteLearningItem(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000003")
	metadata := []byte(`{"k":"v"}`)

	t.Run("get", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
			WithArgs(courseID, nodeID, itemID, uint(1)).
			WillReturnRows(learningItemRows(
				itemID, courseID, nodeID, "Lesson", LearningItemTypeArticle,
				"desc", metadata, 0,
			))
		item, err := model.GetLearningItemByID(courseID, nodeID, itemID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if item.Title != "Lesson" || item.Position != 0 {
			t.Fatalf("unexpected item: %+v", item)
		}
	})

	t.Run("get missing", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
			WithArgs(courseID, nodeID, itemID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		_, err := model.GetLearningItemByID(courseID, nodeID, itemID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
	})

	t.Run("list empty non-nil", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`).
			WithArgs(courseID, nodeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		items, err := model.ListLearningItemsByNode(courseID, nodeID)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if items == nil || len(items) != 0 {
			t.Fatalf("items = %#v, want non-nil empty slice", items)
		}
	})

	t.Run("list missing node", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, false)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		_, err := model.ListLearningItemsByNode(courseID, nodeID)
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNodeNotFound)
		}
	})

	t.Run("update presence-aware", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		title := "Updated"
		itemType := LearningItemTypeQuizRef
		newMetaInput := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"hello"}}]}`)
		newMetaStored := []byte(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"hello"},"visibility":{"mode":"ALL"}}]}`)
		mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(nil, itemType, newMetaStored, title, courseID, nodeID, itemID).
			WillReturnRows(learningItemRows(
				itemID, courseID, nodeID, title, itemType, nil, newMetaStored, 0,
			))
		item, err := model.UpdateLearningItem(courseID, nodeID, itemID, UpdateLearningItemParams{
			Title:       &title,
			ItemType:    &itemType,
			Description: OptionalNullableString{Present: true, Null: true},
			Metadata:    OptionalJSONBytes{Present: true, Value: newMetaInput},
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if item.Title != title || item.ItemType != itemType {
			t.Fatalf("unexpected update result: %+v", item)
		}
	})

	t.Run("update invalid metadata rejected", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		_, err := model.UpdateLearningItem(courseID, nodeID, itemID, UpdateLearningItemParams{
			Metadata: OptionalJSONBytes{
				Present: true,
				Value:   json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":null}]}`),
			},
		})
		if !errors.Is(err, ErrLearningItemMetadataInvalid) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemMetadataInvalid)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("invalid metadata update queried database: %v", err)
		}
	})

	t.Run("update invalid placeholder rejected", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		_, err := model.UpdateLearningItem(courseID, nodeID, itemID, UpdateLearningItemParams{
			Metadata: OptionalJSONBytes{
				Present: true,
				Value:   json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"{{student-name}}"}}]}`),
			},
		})
		if !errors.Is(err, ErrLearningItemPlaceholderInvalid) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemPlaceholderInvalid)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("invalid placeholder update queried database: %v", err)
		}
	})

	t.Run("update invalid visibility rejected", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		_, err := model.UpdateLearningItem(courseID, nodeID, itemID, UpdateLearningItemParams{
			Metadata: OptionalJSONBytes{
				Present: true,
				Value:   json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{},"visibility":{"mode":"PUBLIC"}}]}`),
			},
		})
		if !errors.Is(err, ErrLearningItemVisibilityModeInvalid) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemVisibilityModeInvalid)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("invalid visibility update queried database: %v", err)
		}
	})

	t.Run("create persists valid placeholders unchanged", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		itemIDCreate := uuid.MustParse("019c02a0-1111-7000-8000-000000000099")
		model.newUUID = func() (uuid.UUID, error) { return itemIDCreate, nil }
		metaInput := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"Welcome {{student_name}}"}}]}`)
		metaStored := []byte(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"Welcome {{student_name}}"},"visibility":{"mode":"ALL"}}]}`)
		mock.ExpectBegin()
		expectLearningItemNodeLock(mock, courseID, nodeID, true)
		expectLearningItemMaxPosition(mock, nodeID, nil)
		mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
			WithArgs(
				courseID,
				nodeID,
				sqlmock.AnyArg(),
				itemIDCreate,
				LearningItemTypeArticle,
				metaStored,
				0,
				LearningItemPublishStateDraft,
				"Lesson",
			).
			WillReturnRows(learningItemRows(
				itemIDCreate, courseID, nodeID, "Lesson", LearningItemTypeArticle,
				nil, metaStored, 0,
			))
		mock.ExpectCommit()
		item, err := model.CreateLearningItem(CreateLearningItemParams{
			CourseID: courseID, CourseNodeID: nodeID, Title: "Lesson",
			ItemType: LearningItemTypeArticle, Metadata: metaInput,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if !bytes.Contains(item.Metadata, []byte(`{{student_name}}`)) {
			t.Fatalf("placeholder not preserved: %s", item.Metadata)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL expectations: %v", err)
		}
	})

	t.Run("update empty rejected", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		_, err := model.UpdateLearningItem(courseID, nodeID, itemID, UpdateLearningItemParams{})
		if !errors.Is(err, ErrLearningItemUpdateRequired) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemUpdateRequired)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("empty update queried database: %v", err)
		}
	})

	t.Run("update metadata null rejected", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		_, err := model.UpdateLearningItem(courseID, nodeID, itemID, UpdateLearningItemParams{
			Metadata: OptionalJSONBytes{Present: true, Null: true},
		})
		if !errors.Is(err, ErrLearningItemMetadataInvalid) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemMetadataInvalid)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("metadata null update queried database: %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectExec(`DELETE FROM "learning_items"`).
			WithArgs(courseID, nodeID, itemID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		if err := model.DeleteLearningItem(courseID, nodeID, itemID); err != nil {
			t.Fatalf("delete: %v", err)
		}
	})

	t.Run("delete missing", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectExec(`DELETE FROM "learning_items"`).
			WithArgs(courseID, nodeID, itemID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err := model.DeleteLearningItem(courseID, nodeID, itemID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
	})
}

func TestLearningItemPublishState(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000088")
	metadata := []byte(`{"version":1,"blocks":[]}`)

	t.Run("create defaults draft", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		model.newUUID = func() (uuid.UUID, error) { return itemID, nil }
		mock.ExpectBegin()
		expectLearningItemNodeLock(mock, courseID, nodeID, true)
		expectLearningItemMaxPosition(mock, nodeID, nil)
		mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
			WithArgs(
				courseID, nodeID, sqlmock.AnyArg(), itemID, LearningItemTypeArticle,
				metadata, 0, LearningItemPublishStateDraft, "Lesson",
			).
			WillReturnRows(learningItemRows(
				itemID, courseID, nodeID, "Lesson", LearningItemTypeArticle, nil, metadata, 0,
			))
		mock.ExpectCommit()
		item, err := model.CreateLearningItem(CreateLearningItemParams{
			CourseID: courseID, CourseNodeID: nodeID, Title: "Lesson", ItemType: LearningItemTypeArticle,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if item.PublishState != LearningItemPublishStateDraft {
			t.Fatalf("publish_state = %q", item.PublishState)
		}
	})

	t.Run("create explicit published", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		model.newUUID = func() (uuid.UUID, error) { return itemID, nil }
		mock.ExpectBegin()
		expectLearningItemNodeLock(mock, courseID, nodeID, true)
		expectLearningItemMaxPosition(mock, nodeID, nil)
		mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
			WithArgs(
				courseID, nodeID, sqlmock.AnyArg(), itemID, LearningItemTypeArticle,
				metadata, 0, LearningItemPublishStatePublished, "Lesson",
			).
			WillReturnRows(learningItemRowsWithPublishState(
				itemID, courseID, nodeID, "Lesson", LearningItemTypeArticle, nil, metadata, 0,
				LearningItemPublishStatePublished,
			))
		mock.ExpectCommit()
		item, err := model.CreateLearningItem(CreateLearningItemParams{
			CourseID: courseID, CourseNodeID: nodeID, Title: "Lesson",
			ItemType: LearningItemTypeArticle, PublishState: LearningItemPublishStatePublished,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if item.PublishState != LearningItemPublishStatePublished {
			t.Fatalf("publish_state = %q", item.PublishState)
		}
	})

	t.Run("create invalid rejected", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		_, err := model.CreateLearningItem(CreateLearningItemParams{
			CourseID: courseID, CourseNodeID: nodeID, Title: "Lesson",
			ItemType: LearningItemTypeArticle, PublishState: "draft",
		})
		if !errors.Is(err, ErrLearningItemPublishStateInvalid) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemPublishStateInvalid)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("invalid publish state queried database: %v", err)
		}
	})

	t.Run("update valid", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		state := LearningItemPublishStatePublished
		mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
			WithArgs(state, courseID, nodeID, itemID).
			WillReturnRows(learningItemRowsWithPublishState(
				itemID, courseID, nodeID, "Lesson", LearningItemTypeArticle, nil, metadata, 0, state,
			))
		item, err := model.UpdateLearningItem(courseID, nodeID, itemID, UpdateLearningItemParams{
			PublishState: &state,
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if item.PublishState != LearningItemPublishStatePublished {
			t.Fatalf("publish_state = %q", item.PublishState)
		}
	})

	t.Run("update invalid rejected", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		state := LearningItemPublishState("published")
		_, err := model.UpdateLearningItem(courseID, nodeID, itemID, UpdateLearningItemParams{
			PublishState: &state,
		})
		if !errors.Is(err, ErrLearningItemPublishStateInvalid) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemPublishStateInvalid)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("invalid publish update queried database: %v", err)
		}
	})
}

func TestLearningItemSQLDoesNotLeak(t *testing.T) {
	err := newLearningItemPersistenceError(errors.New(`pq: detail "secret"`))
	if regexp.MustCompile(`(?i)pq:|secret|SQLSTATE`).MatchString(err.Error()) {
		t.Fatalf("persistence error leaked SQL details: %v", err)
	}
	if !errors.Is(err, ErrLearningItemPersistence) {
		t.Fatalf("unwrap failed: %v", err)
	}
}

func TestPublishedLearningItemReads(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000010")
	itemDraft := uuid.MustParse("019c02a0-1111-7000-8000-000000000011")
	itemPublished := uuid.MustParse("019c02a0-1111-7000-8000-000000000012")
	metadata := []byte(`{"version":1,"blocks":[]}`)

	t.Run("get published success", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, itemID, LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(learningItemRowsWithPublishState(
				itemID, courseID, nodeID, "Published Lesson", LearningItemTypeArticle,
				"desc", metadata, 0, LearningItemPublishStatePublished,
			))
		item, err := model.GetPublishedLearningItemByID(courseID, nodeID, itemID)
		if err != nil {
			t.Fatalf("get published: %v", err)
		}
		if item.PublishState != LearningItemPublishStatePublished || item.Title != "Published Lesson" {
			t.Fatalf("unexpected item: %+v", item)
		}
	})

	t.Run("get draft or missing is not found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
			WithArgs(courseID, nodeID, itemID, LearningItemPublishStatePublished, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "course_node_id", "title", "item_type",
				"description", "metadata", "position", "publish_state", "created_at", "updated_at",
			}))
		_, err := model.GetPublishedLearningItemByID(courseID, nodeID, itemID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
	})

	t.Run("list published only ordered", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		rows := learningItemRowsWithPublishState(
			itemPublished, courseID, nodeID, "A", LearningItemTypeArticle,
			nil, metadata, 0, LearningItemPublishStatePublished,
		)
		rows.AddRow(
			itemDraft, courseID, nodeID, "B", LearningItemTypeVideo,
			nil, metadata, 1, LearningItemPublishStatePublished,
			time.Date(2026, time.July, 26, 8, 30, 0, 0, time.UTC),
			time.Date(2026, time.July, 26, 8, 30, 0, 0, time.UTC),
		)
		// SQL filter means the DB returns only published rows; mock returns published-only set.
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
			WillReturnRows(rows)
		items, err := model.ListPublishedLearningItemsByNode(courseID, nodeID)
		if err != nil {
			t.Fatalf("list published: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("len=%d, want 2", len(items))
		}
		if items[0].Title != "A" || items[1].Title != "B" {
			t.Fatalf("order = %#v", items)
		}
		for _, item := range items {
			if item.PublishState != LearningItemPublishStatePublished {
				t.Fatalf("draft leaked into published list: %+v", item)
			}
		}
	})

	t.Run("list published empty non-nil", func(t *testing.T) {
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
			t.Fatalf("items = %#v, want non-nil empty slice", items)
		}
	})

	t.Run("list published missing node", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, false)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		_, err := model.ListPublishedLearningItemsByNode(courseID, nodeID)
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNodeNotFound)
		}
	})
}
