package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

func TestTestSnapshotServiceCreateFromCollectionsRejectsMetadataPending(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db := goqu.New("postgres", sqlDB)
	svc := NewTestSnapshotService(db, zap.NewNop())
	quizID := uuid.New()
	collectionID := uuid.New()
	now := time.Now().UTC()
	filterJSON := []byte(`{"subject":"History"}`)

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(collectionID, quizID, "History", "DYNAMIC", 0, filterJSON, "editor-1", now, now))

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(collectionID, quizID, "History", "DYNAMIC", 0, filterJSON, "editor-1", now, now))

	_, err = svc.CreateFromCollections(CreateTestSnapshotRequest{
		QuizID:    quizID,
		CreatedBy: "editor-1",
	})
	if err != models.ErrTestSnapshotUnresolvedCollection {
		t.Fatalf("err = %v", err)
	}
}

func TestTestSnapshotServiceCreateFromCollectionsRejectsDuplicates(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db := goqu.New("postgres", sqlDB)
	svc := NewTestSnapshotService(db, zap.NewNop())
	quizID := uuid.New()
	c1 := uuid.New()
	c2 := uuid.New()
	q1 := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).
			AddRow(c1, quizID, "A", "STATIC", 0, nil, "editor-1", now, now).
			AddRow(c2, quizID, "B", "STATIC", 1, nil, "editor-1", now, now))

	// resolve c1
	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(c1, quizID, "A", "STATIC", 0, nil, "editor-1", now, now))
	mock.ExpectQuery(`SELECT (.+) FROM "question_collection_members"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "collection_id", "question_id", "position", "created_at",
		}).AddRow(uuid.New(), c1, q1, 0, now))

	// resolve c2 with same question
	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(c2, quizID, "B", "STATIC", 1, nil, "editor-1", now, now))
	mock.ExpectQuery(`SELECT (.+) FROM "question_collection_members"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "collection_id", "question_id", "position", "created_at",
		}).AddRow(uuid.New(), c2, q1, 0, now))

	_, err = svc.CreateFromCollections(CreateTestSnapshotRequest{
		QuizID:    quizID,
		CreatedBy: "editor-1",
	})
	if err != models.ErrTestSnapshotDuplicateQuestion {
		t.Fatalf("err = %v", err)
	}
}

func TestTestSnapshotServiceCreateFromStaticCollectionsFreezesRevision(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db := goqu.New("postgres", sqlDB)
	svc := NewTestSnapshotService(db, zap.NewNop())
	quizID := uuid.New()
	collectionID := uuid.New()
	questionID := uuid.New()
	lineageID := uuid.New()
	snapshotID := uuid.New()
	itemID := uuid.New()
	now := time.Now().UTC()
	options, _ := json.Marshal(map[string]string{"1": "A", "2": "B"})
	answers, _ := json.Marshal([]int{1})

	call := 0
	svc.snapshotModel.SetUUIDGenerator(func() (uuid.UUID, error) {
		call++
		if call == 1 {
			return snapshotID, nil
		}
		return itemID, nil
	})

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(collectionID, quizID, "Section A", "STATIC", 0, nil, "editor-1", now, now))

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(collectionID, quizID, "Section A", "STATIC", 0, nil, "editor-1", now, now))
	mock.ExpectQuery(`SELECT (.+) FROM "question_collection_members"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "collection_id", "question_id", "position", "created_at",
		}).AddRow(uuid.New(), collectionID, questionID, 0, now))

	mock.ExpectQuery(`SELECT (.+) FROM "quiz_questions"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"question_id", "lineage_id", "revision_number", "question", "type", "options", "answers",
			"official_answer", "authoritative_answer", "answer_review_status", "points", "duration_in_seconds",
			"question_media", "options_media", "resource",
		}).AddRow(
			questionID, lineageID, 3, "Capital?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil,
		))

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "test_snapshots"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "test_snapshot_items"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	sourceJSON, _ := json.Marshal([]uuid.UUID{collectionID})
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_collection_ids", "question_count", "created_at",
		}).AddRow(snapshotID, quizID, "editor-1", "CREATED", sourceJSON, 1, now))
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "snapshot_id", "position", "collection_id", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}).AddRow(
			itemID, snapshotID, 0, collectionID, questionID, lineageID, 3,
			"Capital?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil, now,
		))

	snapshot, err := svc.CreateFromCollections(CreateTestSnapshotRequest{
		QuizID:    quizID,
		CreatedBy: "editor-1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if snapshot.Items[0].RevisionNumber != 3 {
		t.Fatalf("revision = %d", snapshot.Items[0].RevisionNumber)
	}
}
