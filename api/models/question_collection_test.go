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

func newQuestionCollectionModelTest(t *testing.T) (*QuestionCollectionModel, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return InitQuestionCollectionModel(goqu.New("postgres", sqlDB)), mock
}

func TestQuestionCollectionModelCreateStaticCollection(t *testing.T) {
	model, mock := newQuestionCollectionModelTest(t)
	quizID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	collectionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5f01")
	now := time.Date(2026, time.July, 27, 14, 0, 0, 0, time.UTC)

	mock.ExpectExec(`INSERT INTO "question_collections"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(
			collectionID,
			quizID,
			"Section A",
			"STATIC",
			0,
			nil,
			"editor-1",
			now,
			now,
		))

	collection, err := model.CreateCollection(CreateQuestionCollectionParams{
		QuizID:    quizID,
		Title:     " Section A ",
		Kind:      QuestionCollectionKindStatic,
		CreatedBy: "editor-1",
	})
	if err != nil {
		t.Fatalf("create static collection: %v", err)
	}
	if collection.Kind != QuestionCollectionKindStatic {
		t.Fatalf("kind = %s", collection.Kind)
	}
	if collection.Title != "Section A" {
		t.Fatalf("title = %q", collection.Title)
	}
}

func TestQuestionCollectionModelCreateDynamicCollectionRequiresFilter(t *testing.T) {
	model, _ := newQuestionCollectionModelTest(t)
	quizID := uuid.New()

	_, err := model.CreateCollection(CreateQuestionCollectionParams{
		QuizID: quizID,
		Title:  "PYQ Pool",
		Kind:   QuestionCollectionKindDynamic,
	})
	if err != ErrQuestionCollectionFilterRequired {
		t.Fatalf("err = %v", err)
	}
}

func TestQuestionCollectionModelCreateDynamicCollectionStoresFilter(t *testing.T) {
	model, mock := newQuestionCollectionModelTest(t)
	quizID := uuid.New()
	collectionID := uuid.New()
	now := time.Now().UTC()
	subject := "History"
	filterJSON, err := json.Marshal(CollectionDynamicFilter{Subject: &subject})
	if err != nil {
		t.Fatalf("marshal filter: %v", err)
	}

	mock.ExpectExec(`INSERT INTO "question_collections"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(
			collectionID,
			quizID,
			"History PYQ",
			"DYNAMIC",
			1,
			filterJSON,
			"editor-1",
			now,
			now,
		))

	collection, err := model.CreateCollection(CreateQuestionCollectionParams{
		QuizID:     quizID,
		Title:      "History PYQ",
		Kind:       QuestionCollectionKindDynamic,
		Position:   1,
		FilterJSON: filterJSON,
		CreatedBy:  "editor-1",
	})
	if err != nil {
		t.Fatalf("create dynamic collection: %v", err)
	}
	if collection.Filter == nil || collection.Filter.Subject == nil || *collection.Filter.Subject != "History" {
		t.Fatalf("filter not attached: %+v", collection.Filter)
	}
}

func TestQuestionCollectionModelReplaceStaticMembers(t *testing.T) {
	model, mock := newQuestionCollectionModelTest(t)
	quizID := uuid.New()
	collectionID := uuid.New()
	questionA := uuid.New()
	questionB := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(
			collectionID, quizID, "Section A", "STATIC", 0, nil, "editor-1", now, now,
		))

	mock.ExpectQuery(`SELECT (.+) FROM "quiz_questions"`).
		WillReturnRows(sqlmock.NewRows([]string{"question_id"}).
			AddRow(questionA).
			AddRow(questionB))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "question_collection_members"`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO "question_collection_members"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "question_collection_members"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT (.+) FROM "question_collection_members"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "collection_id", "question_id", "position", "created_at",
		}).
			AddRow(uuid.New(), collectionID, questionA, 0, now).
			AddRow(uuid.New(), collectionID, questionB, 1, now))

	members, err := model.ReplaceStaticMembers(quizID, collectionID, []uuid.UUID{questionA, questionB})
	if err != nil {
		t.Fatalf("replace members: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("members = %d", len(members))
	}
}

func TestQuestionCollectionModelResolveDynamicMetadataPending(t *testing.T) {
	model, mock := newQuestionCollectionModelTest(t)
	quizID := uuid.New()
	collectionID := uuid.New()
	now := time.Now().UTC()
	year := 2024
	filterJSON, _ := json.Marshal(CollectionDynamicFilter{Year: &year})

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(
			collectionID, quizID, "2024 PYQ", "DYNAMIC", 0, filterJSON, "editor-1", now, now,
		))

	resolution, err := model.ResolveCollection(quizID, collectionID)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolution.ResolutionStatus != CollectionResolutionStatusMetadataPending {
		t.Fatalf("status = %s", resolution.ResolutionStatus)
	}
	if len(resolution.QuestionIDs) != 0 {
		t.Fatalf("expected empty question ids")
	}
}

func TestCollectionDynamicFilterHasMetadataCriteria(t *testing.T) {
	empty := CollectionDynamicFilter{}
	if empty.HasMetadataCriteria() {
		t.Fatal("empty filter should not have metadata criteria")
	}
	subject := "Polity"
	withSubject := CollectionDynamicFilter{Subject: &subject}
	if !withSubject.HasMetadataCriteria() {
		t.Fatal("subject filter should count as metadata criteria")
	}
}
