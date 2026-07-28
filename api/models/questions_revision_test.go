package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func newQuestionModelTest(t *testing.T) (*QuestionModel, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return InitQuestionModel(goqu.New("postgres", sqlDB), zap.NewNop()), mock
}

func TestQuestionModelGetLineageMeta(t *testing.T) {
	model, mock := newQuestionModelTest(t)
	questionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	lineageID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eee")

	mock.ExpectQuery(`SELECT (.+) FROM "questions"`).
		WillReturnRows(sqlmock.NewRows([]string{"lineage_id", "revision_number"}).
			AddRow(lineageID, 3))

	meta, err := model.GetLineageMeta(questionID.String())
	if err != nil {
		t.Fatalf("get lineage meta: %v", err)
	}
	if meta.LineageID != lineageID {
		t.Fatalf("lineage_id = %s, want %s", meta.LineageID, lineageID)
	}
	if meta.RevisionNumber != 3 {
		t.Fatalf("revision_number = %d, want 3", meta.RevisionNumber)
	}
}

func TestQuestionModelGetLineageMetaNotFound(t *testing.T) {
	model, mock := newQuestionModelTest(t)
	questionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")

	mock.ExpectQuery(`SELECT (.+) FROM "questions"`).
		WillReturnRows(sqlmock.NewRows([]string{"lineage_id", "revision_number"}))

	_, err := model.GetLineageMeta(questionID.String())
	if err != sql.ErrNoRows {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
}

func TestQuestionModelListRevisionsByQuestionId(t *testing.T) {
	model, mock := newQuestionModelTest(t)
	questionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	lineageID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eee")
	revisionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eef")
	now := time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT (.+) FROM "questions"`).
		WillReturnRows(sqlmock.NewRows([]string{"lineage_id", "revision_number"}).
			AddRow(lineageID, 2))

	mock.ExpectQuery(`SELECT (.+) FROM "question_revisions"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"question_id",
			"lineage_id",
			"revision_number",
			"question",
			"type",
			"options",
			"answers",
			"official_answer",
			"authoritative_answer",
			"answer_review_status",
			"answer_revision_reason",
			"answer_revision_source",
			"points",
			"duration_in_seconds",
			"question_media",
			"options_media",
			"resource",
			"created_by",
			"created_at",
		}).AddRow(
			revisionID,
			questionID,
			lineageID,
			2,
			"Stem?",
			1,
			`{"1":"A"}`,
			`[1]`,
			`[1]`,
			`[1]`,
			constants.AnswerReviewConfirmed,
			nil,
			nil,
			1,
			30,
			"text",
			"text",
			nil,
			nil,
			now,
		))

	revisions, err := model.ListRevisionsByQuestionId(questionID.String())
	if err != nil {
		t.Fatalf("list revisions: %v", err)
	}
	if len(revisions) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(revisions))
	}
	if revisions[0].LineageID != lineageID {
		t.Fatalf("lineage_id = %s, want %s", revisions[0].LineageID, lineageID)
	}
}
