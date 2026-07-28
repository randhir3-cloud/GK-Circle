package v1

import (
	"bytes"
	"encoding/json"
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

func TestCreateAssessmentAttempt_Unauthenticated(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/attempts", ctrl.CreateAttempt)

	body := []byte(`{"snapshot_id":"` + uuid.New().String() + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+uuid.New().String()+"/attempts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
}

func TestGetAttemptInstructions_OmitsKeysAndItems(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	snapshotID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	mock.ExpectQuery(`SELECT (.+) FROM "shared_quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"permission"}))
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status", "duration_seconds", "max_attempts",
			"negative_marks_per_question", "allow_answer_review",
		}).AddRow(quizID, "PCS Practice", "Read carefully", "editor-1", true, "SELF_PACED", "PUBLISHED", int64(1800), 2, 0.33, false))
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
			uuid.New(), snapshotID, 0, nil, questionID, uuid.New(), 1,
			"Secret stem?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}))
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/instructions", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		return c.Next()
	}, ctrl.GetInstructions)

	req := httptest.NewRequest(
		http.MethodGet,
		"/quizzes/"+quizID.String()+"/attempts/instructions?snapshot_id="+snapshotID.String(),
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
	if regexp.MustCompile(`"answers"|"official_answer"|"authoritative_answer"|"Secret stem"|"items"`).Match(raw) {
		t.Fatalf("instructions leaked keys/items: %s", raw)
	}
	var envelope struct {
		Data structs.AssessmentAttemptInstructionsResponse `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !envelope.Data.CanStart || envelope.Data.Quiz.Title != "PCS Practice" {
		t.Fatalf("data = %+v", envelope.Data)
	}
	if envelope.Data.Snapshot.QuestionCount != 1 {
		t.Fatalf("question count = %d", envelope.Data.Snapshot.QuestionCount)
	}
}

func TestCreateAssessmentAttempt_LearnerNoKeys(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	snapshotID := uuid.New()
	attemptID := uuid.New()
	questionID := uuid.New()
	lineageID := uuid.New()
	now := time.Now().UTC()
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})

	ctrl.attemptSvc.AttemptModelForTest().SetUUIDGenerator(func() (uuid.UUID, error) { return attemptID, nil })

	// ResolveEditorPreview: not creator
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	// GetPermission — no share row
	mock.ExpectQuery(`SELECT (.+) FROM "shared_quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"permission"}))

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status", "duration_seconds", "max_attempts",
			"negative_marks_per_question", "allow_answer_review",
		}).AddRow(quizID, "PCS Practice", "Instructions body", "editor-1", true, "SELF_PACED", "PUBLISHED", nil, 3, 0.0, false))
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}))
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
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
			uuid.New(), snapshotID, 0, nil, questionID, lineageID, 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil, now,
		))
	mock.ExpectQuery(`SELECT MAX`).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "assessment_attempts"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "assessment_attempt_snapshot_items"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "assessment_analytics_events"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))
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
			uuid.New(), snapshotID, 0, nil, questionID, lineageID, 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "snapshot_item_id", "position", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}).AddRow(
			uuid.New(), attemptID, uuid.New(), 0, questionID, lineageID, 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil, now,
		))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/attempts", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		c.Locals(constants.ContextUid, "learner-1")
		return c.Next()
	}, ctrl.CreateAttempt)

	body := []byte(`{"snapshot_id":"` + snapshotID.String() + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/attempts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
	if regexp.MustCompile(`"answers"|"official_answer"|"authoritative_answer"|"answer_review_status"`).Match(raw) {
		t.Fatalf("learner response leaked keys: %s", raw)
	}
	var envelope struct {
		Data structs.AssessmentAttemptResponse `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data.TestSnapshotID != snapshotID.String() {
		t.Fatalf("snapshot id = %s", envelope.Data.TestSnapshotID)
	}
	if envelope.Data.UserID != "learner-1" {
		t.Fatalf("user id = %s", envelope.Data.UserID)
	}
}

func TestGetEditorAssessmentAttempt_IncludesKeys(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))
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
			uuid.New(), snapshotID, 0, nil, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil, now,
		))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id/editor", ctrl.GetEditorAttempt)

	req := httptest.NewRequest(http.MethodGet, "/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/editor", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
	if !regexp.MustCompile(`"answers"`).Match(raw) {
		t.Fatalf("editor response missing answer keys: %s", raw)
	}
}

func TestGetMyAssessmentAttempt_ForeignOwnerNotFound(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	attemptID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{})

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			attemptID, quizID, "owner-1", uuid.New(), 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "other-user", Email: "other@example.com"})
		return c.Next()
	}, ctrl.GetMyAttempt)

	req := httptest.NewRequest(http.MethodGet, "/quizzes/"+quizID.String()+"/attempts/"+attemptID.String(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
}

func TestAutosaveAttemptAnswer_Unauthenticated(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	app := fiber.New()
	app.Put("/quizzes/:quiz_id/attempts/:attempt_id/answers/:question_id", ctrl.AutosaveAnswer)

	body := []byte(`{"selected_options":[1]}`)
	req := httptest.NewRequest(
		http.MethodPut,
		"/quizzes/"+uuid.New().String()+"/attempts/"+uuid.New().String()+"/answers/"+uuid.New().String(),
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
}

func TestAutosaveAttemptAnswer_RejectsClientScoreFields(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	app := fiber.New()
	app.Put("/quizzes/:quiz_id/attempts/:attempt_id/answers/:question_id", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		return c.Next()
	}, ctrl.AutosaveAnswer)

	body := []byte(`{"selected_options":[1],"score":5,"is_correct":true}`)
	req := httptest.NewRequest(
		http.MethodPut,
		"/quizzes/"+uuid.New().String()+"/attempts/"+uuid.New().String()+"/answers/"+uuid.New().String(),
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
}

func TestResumeAttempt_OmitsAnswerKeys(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})
	selected, _ := json.Marshal([]int{1})

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))
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
			uuid.New(), snapshotID, 0, nil, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "snapshot_item_id", "position", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}).AddRow(
			uuid.New(), attemptID, uuid.New(), 0, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(uuid.New(), attemptID, questionID, selected, false, now, nil, nil, nil, now, now))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id/resume", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		return c.Next()
	}, ctrl.ResumeAttempt)

	req := httptest.NewRequest(http.MethodGet, "/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/resume", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
	if regexp.MustCompile(`"official_answer"|"authoritative_answer"|"answer_review_status"|"is_correct"|"score"`).Match(raw) {
		t.Fatalf("resume leaked keys/scores: %s", raw)
	}
	var envelope struct {
		Data structs.AssessmentAttemptResumeResponse `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data.Progress.AnsweredCount != 1 {
		t.Fatalf("progress = %+v", envelope.Data.Progress)
	}
}

func TestSubmitAssessmentAttempt_Unauthenticated(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/attempts/:attempt_id/submit", ctrl.SubmitAttempt)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+uuid.New().String()+"/attempts/"+uuid.New().String()+"/submit", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
}

func TestSubmitAssessmentAttempt_RejectsClientScoreFields(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/attempts/:attempt_id/submit", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		return c.Next()
	}, ctrl.SubmitAttempt)

	body := []byte(`{"total_score":99,"status":"SUBMITTED"}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/quizzes/"+uuid.New().String()+"/attempts/"+uuid.New().String()+"/submit",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
}

func TestGetAttemptResult_Unauthenticated(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id/result", ctrl.GetAttemptResult)

	req := httptest.NewRequest(
		http.MethodGet,
		"/quizzes/"+uuid.New().String()+"/attempts/"+uuid.New().String()+"/result",
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, raw)
	}
}

func TestGetAttemptResult_RejectsInProgress(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	attemptID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at", "created_at", "updated_at",
			"total_score", "max_score", "time_taken_seconds",
		}).AddRow(
			attemptID, quizID, "learner-1", uuid.New(), 1, "IN_PROGRESS",
			`[]`, 0.0, 10.0,
			now, nil, nil, now, now,
			nil, 10.0, nil,
		))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id/result", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		return c.Next()
	}, ctrl.GetAttemptResult)

	req := httptest.NewRequest(
		http.MethodGet,
		"/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/result",
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d expected 409 body=%s", resp.StatusCode, raw)
	}
}

func TestGetAttemptResult_OwnerMismatch_ReturnsNotFound(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	attemptID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at", "created_at", "updated_at",
			"total_score", "max_score", "time_taken_seconds",
		}).AddRow(
			attemptID, quizID, "owner-user", uuid.New(), 1, "SUBMITTED",
			`[]`, 0.0, 10.0,
			now, now, nil, now, now,
			5.0, 10.0, 120,
		))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id/result", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "other-user", Email: "other@example.com"})
		return c.Next()
	}, ctrl.GetAttemptResult)

	req := httptest.NewRequest(
		http.MethodGet,
		"/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/result",
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d expected 404 body=%s", resp.StatusCode, raw)
	}
}

func TestGetAttemptResult_ReviewDisabled_OmitsReviewBlockAndKeys(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	attemptID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	optionsJSON, _ := json.Marshal(map[string]string{"1": "Delhi", "2": "Patna"})
	answersJSON, _ := json.Marshal([]int{2})
	selectedOptionsJSON, _ := json.Marshal([]int{2})

	// 1. GetByID attempt
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at", "created_at", "updated_at",
			"total_score", "max_score", "time_taken_seconds",
		}).AddRow(
			attemptID, quizID, "learner-1", uuid.New(), 1, "SUBMITTED",
			`[]`, 0.0, 10.0,
			now, now, nil, now, now,
			10.0, 10.0, 60,
		))

	// 2. GetSelfPacedMetaByID quiz (allow_answer_review = false)
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(quizID, "BPSC Practice", "Mock Test", "editor-1", true, "SELF_PACED", "PUBLISHED", int64(1800), 1, 0.0, false, "IMMEDIATE", false, nil, true, true, true, true))

	// 3. ListByAttemptID snapshot items
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "snapshot_item_id", "position", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}).AddRow(
			uuid.New(), attemptID, uuid.New(), 0, questionID, uuid.New(), 1,
			"Capital of Bihar?", 1, optionsJSON, answersJSON, answersJSON, answersJSON,
			"CONFIRMED", int16(10), 60, "", "", "Patna is the capital.", now,
		))

	// 4. ListByAttemptID answers
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review", "is_correct", "score",
			"answered_at", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			uuid.New(), attemptID, questionID, selectedOptionsJSON, false, true, 10.0,
			now, 60, now, now,
		))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id/result", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		return c.Next()
	}, ctrl.GetAttemptResult)

	req := httptest.NewRequest(
		http.MethodGet,
		"/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/result",
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d expected 200 body=%s", resp.StatusCode, raw)
	}

	raw, _ := io.ReadAll(resp.Body)
	bodyStr := string(raw)

	// Assertions for Review Disabled
	if regexp.MustCompile(`"review":\s*\{`).MatchString(bodyStr) {
		t.Fatalf("expected review block to be absent/null when review is disabled, got: %s", bodyStr)
	}
	if regexp.MustCompile(`"test_snapshot_id"`).MatchString(bodyStr) {
		t.Fatalf("test_snapshot_id leaked in response: %s", bodyStr)
	}
	if regexp.MustCompile(`"negative_marks_per_question"`).MatchString(bodyStr) {
		t.Fatalf("negative_marks_per_question leaked in response: %s", bodyStr)
	}
	if regexp.MustCompile(`"official_answer"`).MatchString(bodyStr) {
		t.Fatalf("official_answer leaked in response: %s", bodyStr)
	}

	var jsonResp map[string]interface{}
	if err := json.Unmarshal(raw, &jsonResp); err != nil {
		t.Fatalf("json parse error: %v", err)
	}
	data, _ := jsonResp["data"].(map[string]interface{})
	if data["can_review_questions"] != false {
		t.Fatalf("expected can_review_questions = false, got: %v", data["can_review_questions"])
	}
	summary, _ := data["summary"].(map[string]interface{})
	if summary["total_score"] != 10.0 || summary["percentage"] != 100.0 {
		t.Fatalf("unexpected summary score/percentage: %v", summary)
	}
}

func TestGetAttemptResult_ManualRelease_Withheld(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	attemptID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at", "created_at", "updated_at",
			"total_score", "max_score", "time_taken_seconds",
		}).AddRow(
			attemptID, quizID, "learner-1", uuid.New(), 1, "SUBMITTED",
			`[]`, 0.0, 10.0,
			now, now, nil, now, now,
			10.0, 10.0, 60,
		))

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(quizID, "BPSC Practice", "Mock Test", "editor-1", true, "SELF_PACED", "PUBLISHED", int64(1800), 1, 0.0, true, "MANUAL", false, nil, true, true, true, true))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id/result", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		return c.Next()
	}, ctrl.GetAttemptResult)

	req := httptest.NewRequest(
		http.MethodGet,
		"/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/result",
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d expected 200 body=%s", resp.StatusCode, raw)
	}

	raw, _ := io.ReadAll(resp.Body)
	var jsonResp map[string]interface{}
	_ = json.Unmarshal(raw, &jsonResp)
	data, _ := jsonResp["data"].(map[string]interface{})

	if data["can_view_result"] != false {
		t.Fatalf("expected can_view_result = false, got: %v", data["can_view_result"])
	}
	if data["summary"] != nil {
		t.Fatalf("expected summary = nil when result is withheld, got: %v", data["summary"])
	}
	if data["review"] != nil {
		t.Fatalf("expected review = nil when result is withheld, got: %v", data["review"])
	}
}

func TestGetAttemptResult_ScoreHidden_OmitsScoreFields(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	attemptID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()

	optionsJSON, _ := json.Marshal(map[string]string{"1": "Patna", "2": "Gaya"})
	answersJSON, _ := json.Marshal([]int{1})
	selectedOptionsJSON, _ := json.Marshal([]int{1})

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at", "created_at", "updated_at",
			"total_score", "max_score", "time_taken_seconds",
		}).AddRow(
			attemptID, quizID, "learner-1", uuid.New(), 1, "SUBMITTED",
			`[]`, 0.0, 10.0,
			now, now, nil, now, now,
			10.0, 10.0, 60,
		))

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(quizID, "BPSC Practice", "Mock Test", "editor-1", true, "SELF_PACED", "PUBLISHED", int64(1800), 1, 0.0, true, "IMMEDIATE", false, nil, false, true, true, true))

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "snapshot_item_id", "position", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}).AddRow(
			uuid.New(), attemptID, uuid.New(), 0, questionID, uuid.New(), 1,
			"Capital of Bihar?", 1, optionsJSON, answersJSON, answersJSON, answersJSON,
			"CONFIRMED", int16(10), 60, "", "", "Patna is the capital.", now,
		))

	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review", "is_correct", "score",
			"answered_at", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			uuid.New(), attemptID, questionID, selectedOptionsJSON, false, true, 10.0,
			now, 60, now, now,
		))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id/result", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		return c.Next()
	}, ctrl.GetAttemptResult)

	req := httptest.NewRequest(
		http.MethodGet,
		"/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/result",
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	bodyStr := string(raw)

	if regexp.MustCompile(`"total_score"`).MatchString(bodyStr) {
		t.Fatalf("expected total_score to be omitted when show_score=false, got: %s", bodyStr)
	}
	if regexp.MustCompile(`"percentage"`).MatchString(bodyStr) {
		t.Fatalf("expected percentage to be omitted when show_score=false, got: %s", bodyStr)
	}
}

func TestGetAttemptResult_InvalidPolicy_FailsClosed(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitAssessmentAttemptController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	attemptID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at", "created_at", "updated_at",
			"total_score", "max_score", "time_taken_seconds",
		}).AddRow(
			attemptID, quizID, "learner-1", uuid.New(), 1, "SUBMITTED",
			`[]`, 0.0, 10.0,
			now, now, nil, now, now,
			10.0, 10.0, 60,
		))

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(quizID, "BPSC Practice", "Mock Test", "editor-1", true, "SELF_PACED", "PUBLISHED", int64(1800), 1, 0.0, true, "INVALID_POLICY", false, nil, true, true, true, true))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/attempts/:attempt_id/result", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "learner@example.com"})
		return c.Next()
	}, ctrl.GetAttemptResult)

	req := httptest.NewRequest(
		http.MethodGet,
		"/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/result",
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	raw, _ := io.ReadAll(resp.Body)
	var jsonResp map[string]interface{}
	_ = json.Unmarshal(raw, &jsonResp)
	data, _ := jsonResp["data"].(map[string]interface{})

	if data["can_view_result"] != false {
		t.Fatalf("expected invalid policy to fail closed (can_view_result = false), got: %v", data["can_view_result"])
	}
}


