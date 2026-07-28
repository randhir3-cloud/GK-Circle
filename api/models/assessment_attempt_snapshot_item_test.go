package models

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
)

func TestAssessmentAttemptSnapshotItemToLearnerOmitsKeys(t *testing.T) {
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})
	item := AssessmentAttemptSnapshotItem{
		QuestionID:          uuid.New(),
		LineageID:           uuid.New(),
		RevisionNumber:      2,
		Question:            "Capital?",
		Type:                1,
		OptionsJSON:         options,
		AnswersJSON:         answers,
		OfficialAnswerJSON:  answers,
		AuthoritativeJSON:   answers,
		AnswerReviewStatus:  "CONFIRMED",
		Options:             map[string]string{"1": "A"},
		Answers:             []int{1},
		OfficialAnswer:      []int{1},
		AuthoritativeAnswer: []int{1},
	}
	learner := item.ToLearnerItem()
	payload, err := json.Marshal(learner)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if regexp.MustCompile(`"answers"|"official_answer"|"authoritative_answer"|"answer_review_status"`).Match(payload) {
		t.Fatalf("leaked keys: %s", payload)
	}
}

func TestAssessmentAttemptSnapshotItemListDecodesFreeze(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	model := InitAssessmentAttemptSnapshotItemModel(goqu.New("postgres", sqlDB))

	attemptID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "snapshot_item_id", "position", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}).AddRow(
			uuid.New(), attemptID, uuid.New(), 0, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(1), 30, "text", "text", nil, now,
		))

	items, err := model.ListByAttemptID(attemptID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 || items[0].AuthoritativeAnswer[0] != 1 {
		t.Fatalf("items = %+v", items)
	}
}
