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

func newTestSnapshotModelTest(t *testing.T) (*TestSnapshotModel, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return InitTestSnapshotModel(goqu.New("postgres", sqlDB)), mock
}

func TestTestSnapshotModelCreateSnapshotRejectsEmpty(t *testing.T) {
	model, _ := newTestSnapshotModelTest(t)
	_, err := model.CreateSnapshot(CreateTestSnapshotParams{
		QuizID:              uuid.New(),
		CreatedBy:           "editor-1",
		SourceCollectionIDs: []uuid.UUID{uuid.New()},
		Items:               nil,
	})
	if err != ErrTestSnapshotEmpty {
		t.Fatalf("err = %v", err)
	}
}

func TestTestSnapshotModelCreateSnapshotRejectsDuplicates(t *testing.T) {
	model, _ := newTestSnapshotModelTest(t)
	qid := uuid.New()
	_, err := model.CreateSnapshot(CreateTestSnapshotParams{
		QuizID:              uuid.New(),
		CreatedBy:           "editor-1",
		SourceCollectionIDs: []uuid.UUID{uuid.New()},
		Items: []CreateTestSnapshotItemParams{
			{Position: 0, Freeze: SnapshotQuestionFreeze{QuestionID: qid}},
			{Position: 1, Freeze: SnapshotQuestionFreeze{QuestionID: qid}},
		},
	})
	if err != ErrTestSnapshotDuplicateQuestion {
		t.Fatalf("err = %v", err)
	}
}

func TestTestSnapshotModelCreateAndReadPersistsFrozenPayload(t *testing.T) {
	model, mock := newTestSnapshotModelTest(t)
	quizID := uuid.New()
	collectionID := uuid.New()
	snapshotID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea6001")
	questionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea6002")
	lineageID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea6003")
	itemID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea6004")
	now := time.Date(2026, 7, 27, 18, 0, 0, 0, time.UTC)

	call := 0
	model.newUUID = func() (uuid.UUID, error) {
		call++
		if call == 1 {
			return snapshotID, nil
		}
		return itemID, nil
	}
	options, _ := json.Marshal(map[string]string{"1": "Paris", "2": "Berlin"})
	answers, _ := json.Marshal([]int{1})

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
			itemID, snapshotID, 0, collectionID, questionID, lineageID, 2,
			"Capital?", 1, options, answers, answers, answers,
			"CONFIRMED", int16(5), 30, "text", "text", nil, now,
		))

	points := int16(5)
	duration := 30
	snapshot, err := model.CreateSnapshot(CreateTestSnapshotParams{
		QuizID:              quizID,
		CreatedBy:           "editor-1",
		SourceCollectionIDs: []uuid.UUID{collectionID},
		Items: []CreateTestSnapshotItemParams{
			{
				Position:     0,
				CollectionID: &collectionID,
				Freeze: SnapshotQuestionFreeze{
					QuestionID:         questionID,
					LineageID:          lineageID,
					RevisionNumber:     2,
					Question:           "Capital?",
					Type:               1,
					OptionsJSON:        options,
					AnswersJSON:        answers,
					OfficialAnswerJSON: answers,
					AuthoritativeJSON:  answers,
					AnswerReviewStatus: "CONFIRMED",
					Points:             &points,
					DurationInSeconds:  &duration,
					QuestionMedia:      "text",
					OptionsMedia:       "text",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if snapshot.QuestionCount != 1 || len(snapshot.Items) != 1 {
		t.Fatalf("items = %d count=%d", len(snapshot.Items), snapshot.QuestionCount)
	}
	if snapshot.Items[0].RevisionNumber != 2 || snapshot.Items[0].Answers[0] != 1 {
		t.Fatalf("frozen payload mismatch: %+v", snapshot.Items[0])
	}

	learner := model.ToLearnerView(snapshot)
	payload, err := json.Marshal(learner)
	if err != nil {
		t.Fatalf("marshal learner: %v", err)
	}
	body := string(payload)
	if regexp.MustCompile(`"answers"|"official_answer"|"authoritative_answer"|"answer_review_status"`).MatchString(body) {
		t.Fatalf("learner view leaked answer keys: %s", body)
	}
}

func TestTestSnapshotLearnerViewOmitsAnswerKeys(t *testing.T) {
	model := InitTestSnapshotModel(nil)
	view := model.ToLearnerView(TestSnapshot{
		ID:            uuid.New(),
		QuizID:        uuid.New(),
		Status:        TestSnapshotStatusCreated,
		QuestionCount: 1,
		Items: []TestSnapshotItem{
			{
				Position:            0,
				QuestionID:          uuid.New(),
				LineageID:           uuid.New(),
				RevisionNumber:      1,
				Question:            "Q?",
				Type:                1,
				Options:             map[string]string{"1": "A"},
				Answers:             []int{1},
				OfficialAnswer:      []int{1},
				AuthoritativeAnswer: []int{1},
				AnswerReviewStatus:  "CONFIRMED",
			},
		},
	})
	if len(view.Items) != 1 {
		t.Fatalf("items = %d", len(view.Items))
	}
	raw, _ := json.Marshal(view.Items[0])
	if regexp.MustCompile(`answers|official_answer|authoritative_answer|answer_review`).MatchString(string(raw)) {
		t.Fatalf("leaked keys: %s", raw)
	}
}
