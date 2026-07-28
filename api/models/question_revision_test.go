package models

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

func newQuestionRevisionModelTest(t *testing.T) (*QuestionRevisionModel, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return InitQuestionRevisionModel(goqu.New("postgres", sqlDB), nil), mock
}

func TestQuestionRevisionModelListByLineageID(t *testing.T) {
	model, mock := newQuestionRevisionModelTest(t)
	lineageID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	revisionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eee")
	questionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eef")
	now := time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

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

	revisions, err := model.ListByLineageID(lineageID.String())
	if err != nil {
		t.Fatalf("list revisions: %v", err)
	}
	if len(revisions) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(revisions))
	}
	if revisions[0].RevisionNumber != 2 {
		t.Fatalf("revision number: got %d", revisions[0].RevisionNumber)
	}
}
