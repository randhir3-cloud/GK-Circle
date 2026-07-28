package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrImportJobCommitInProgress = errors.New("import job commit already in progress")
	ErrImportJobNotCommitable    = errors.New("import job cannot be committed")
	ErrImportJobNoValidRows      = errors.New("import job has no valid rows to commit")
	ErrImportCommitDuplicates    = errors.New("import commit blocked: duplicate question detected")
)

const QuestionImportJobsTable = "question_import_jobs"

const (
	ImportJobStatusPreviewed  = "PREVIEWED"
	ImportJobStatusFailed     = "FAILED"
	ImportJobStatusCommitting = "COMMITTING"
	ImportJobStatusCommitted  = "COMMITTED"
)

// ImportPreviewRow is one CSV row ready for bank create (authority defaults applied).
type ImportPreviewRow struct {
	RowNumber             int               `json:"row_number"`
	Question              string            `json:"question"`
	Type                  int               `json:"type"`
	Options               map[string]string `json:"options"`
	Answers               []int             `json:"answers"`
	OfficialAnswer        []int             `json:"official_answer"`
	AuthoritativeAnswer   []int             `json:"authoritative_answer"`
	AnswerReviewStatus    string            `json:"answer_review_status"`
	Points                int16             `json:"points"`
	DurationInSeconds     int               `json:"duration_in_seconds"`
	QuestionMedia         string            `json:"question_media"`
	OptionsMedia          string            `json:"options_media"`
	Resource              string            `json:"resource"`
	RevisionNumber        int               `json:"revision_number"`
}

// ImportRowError captures deterministic validation failures for a CSV row.
type ImportRowError struct {
	RowNumber           int      `json:"row_number"`
	Messages            []string `json:"messages"`
	Kind                string   `json:"kind,omitempty"`
	DuplicateOfRow      *int     `json:"duplicate_of_row,omitempty"`
	DuplicateQuestionID *string  `json:"duplicate_question_id,omitempty"`
}

// ImportPreviewPayload is stored on the job and returned by the API.
type ImportPreviewPayload struct {
	ValidRows []ImportPreviewRow `json:"valid_rows"`
	Errors    []ImportRowError   `json:"errors"`
}

// ImportCommitResult is persisted when a preview job is committed.
type ImportCommitResult struct {
	QuestionIDs    []string `json:"question_ids,omitempty"`
	CommittedCount int      `json:"committed_count"`
	Error          string   `json:"error,omitempty"`
}

type QuestionImportJob struct {
	ID             uuid.UUID `json:"id" db:"id"`
	QuizID         uuid.UUID `json:"quiz_id" db:"quiz_id"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	Status         string    `json:"status" db:"status"`
	SourceFilename string    `json:"source_filename" db:"source_filename"`
	TotalRows      int       `json:"total_rows" db:"total_rows"`
	ValidRowCount  int       `json:"valid_row_count" db:"valid_row_count"`
	ErrorRowCount  int       `json:"error_row_count" db:"error_row_count"`
	PreviewJSON    []byte    `json:"-" db:"preview_json"`
	CommitResultJSON []byte               `json:"-" db:"commit_result_json"`
	CommittedAt      sql.NullTime         `json:"committed_at,omitempty" db:"committed_at"`
	CreatedAt        time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at" db:"updated_at"`
	Preview          ImportPreviewPayload `json:"preview" db:"-"`
	CommitResult     ImportCommitResult   `json:"commit_result,omitempty" db:"-"`
}

type QuestionImportJobModel struct {
	db     *goqu.Database
	logger *zap.Logger
}

func InitQuestionImportJobModel(db *goqu.Database, logger *zap.Logger) *QuestionImportJobModel {
	return &QuestionImportJobModel{db: db, logger: logger}
}

func (model *QuestionImportJobModel) CreatePreviewJob(
	quizID string,
	createdBy string,
	sourceFilename string,
	preview ImportPreviewPayload,
) (QuestionImportJob, error) {
	jobID, err := uuid.NewUUID()
	if err != nil {
		return QuestionImportJob{}, err
	}

	parsedQuizID, err := uuid.Parse(quizID)
	if err != nil {
		return QuestionImportJob{}, err
	}

	previewBytes, err := json.Marshal(preview)
	if err != nil {
		return QuestionImportJob{}, err
	}

	status := ImportJobStatusPreviewed
	if len(preview.ValidRows) == 0 && len(preview.Errors) > 0 {
		status = ImportJobStatusFailed
	}

	job := QuestionImportJob{
		ID:             jobID,
		QuizID:         parsedQuizID,
		CreatedBy:      createdBy,
		Status:         status,
		SourceFilename: sourceFilename,
		TotalRows:      len(preview.ValidRows) + len(preview.Errors),
		ValidRowCount:  len(preview.ValidRows),
		ErrorRowCount:  len(preview.Errors),
		PreviewJSON:    previewBytes,
		Preview:        preview,
	}

	_, err = model.db.Insert(QuestionImportJobsTable).Rows(goqu.Record{
		"id":              job.ID,
		"quiz_id":         job.QuizID,
		"created_by":      job.CreatedBy,
		"status":          job.Status,
		"source_filename": job.SourceFilename,
		"total_rows":      job.TotalRows,
		"valid_row_count": job.ValidRowCount,
		"error_row_count": job.ErrorRowCount,
		"preview_json":    string(previewBytes),
		"created_at":      goqu.L("now()"),
		"updated_at":      goqu.L("now()"),
	}).Executor().Exec()
	if err != nil {
		return QuestionImportJob{}, err
	}

	return job, nil
}

func (model *QuestionImportJobModel) hydrateJob(job *QuestionImportJob) error {
	if len(job.PreviewJSON) > 0 {
		if err := json.Unmarshal(job.PreviewJSON, &job.Preview); err != nil {
			return err
		}
	}
	if len(job.CommitResultJSON) > 0 {
		if err := json.Unmarshal(job.CommitResultJSON, &job.CommitResult); err != nil {
			return err
		}
	}
	return nil
}

func (model *QuestionImportJobModel) importJobSelectColumns() []interface{} {
	return []interface{}{
		"id",
		"quiz_id",
		"created_by",
		"status",
		"source_filename",
		"total_rows",
		"valid_row_count",
		"error_row_count",
		"preview_json",
		"commit_result_json",
		"committed_at",
		"created_at",
		"updated_at",
	}
}

func (model *QuestionImportJobModel) GetByQuizAndID(quizID, jobID string) (QuestionImportJob, error) {
	var job QuestionImportJob
	found, err := model.db.From(QuestionImportJobsTable).
		Select(model.importJobSelectColumns()...).
		Where(goqu.Ex{
			"id":      jobID,
			"quiz_id": quizID,
		}).
		Limit(1).
		ScanStruct(&job)
	if err != nil {
		return job, err
	}
	if !found {
		return job, sql.ErrNoRows
	}

	if err := model.hydrateJob(&job); err != nil {
		return job, err
	}

	return job, nil
}

// TryClaimForCommit transitions a preview job into COMMITTING inside an open transaction.
func (model *QuestionImportJobModel) TryClaimForCommit(
	transaction *goqu.TxDatabase,
	quizID, jobID string,
) (QuestionImportJob, bool, error) {
	var job QuestionImportJob
	found, err := transaction.From(QuestionImportJobsTable).
		Select(model.importJobSelectColumns()...).
		Where(goqu.Ex{
			"id":      jobID,
			"quiz_id": quizID,
			"status":  []string{ImportJobStatusPreviewed, ImportJobStatusFailed},
		}).
		Where(goqu.I("valid_row_count").Gt(0)).
		ForUpdate(goqu.Wait).
		Limit(1).
		ScanStruct(&job)
	if err != nil {
		return job, false, err
	}
	if !found {
		return job, false, nil
	}

	_, err = transaction.Update(QuestionImportJobsTable).
		Set(goqu.Record{
			"status":     ImportJobStatusCommitting,
			"updated_at": goqu.L("now()"),
		}).
		Where(goqu.Ex{"id": jobID, "quiz_id": quizID}).
		Executor().Exec()
	if err != nil {
		return job, false, err
	}

	job.Status = ImportJobStatusCommitting
	if err := model.hydrateJob(&job); err != nil {
		return job, false, err
	}

	return job, true, nil
}

func (model *QuestionImportJobModel) FinalizeCommit(
	transaction *goqu.TxDatabase,
	quizID, jobID string,
	result ImportCommitResult,
) (QuestionImportJob, error) {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return QuestionImportJob{}, err
	}

	var job QuestionImportJob
	found, err := transaction.Update(QuestionImportJobsTable).
		Set(goqu.Record{
			"status":             ImportJobStatusCommitted,
			"commit_result_json": string(resultBytes),
			"committed_at":       goqu.L("now()"),
			"updated_at":         goqu.L("now()"),
		}).
		Where(goqu.Ex{
			"id":      jobID,
			"quiz_id": quizID,
			"status":  ImportJobStatusCommitting,
		}).
		Returning(model.importJobSelectColumns()...).
		Executor().ScanStruct(&job)
	if err != nil {
		return job, err
	}
	if !found {
		return job, sql.ErrNoRows
	}

	if err := model.hydrateJob(&job); err != nil {
		return job, err
	}

	return job, nil
}

func (model *QuestionImportJobModel) MarkCommitFailed(quizID, jobID, commitError string) error {
	result := ImportCommitResult{Error: commitError}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	_, err = model.db.Update(QuestionImportJobsTable).
		Set(goqu.Record{
			"status":             ImportJobStatusFailed,
			"commit_result_json": string(resultBytes),
			"updated_at":         goqu.L("now()"),
		}).
		Where(goqu.Ex{
			"id":      jobID,
			"quiz_id": quizID,
		}).
		Where(goqu.I("status").In(ImportJobStatusPreviewed, ImportJobStatusCommitting)).
		Executor().Exec()
	return err
}

// QuestionsFromImportPreviewRows converts stored preview rows into bank questions,
// re-applying answer authority so commit does not trust client-supplied values.
func QuestionsFromImportPreviewRows(rows []ImportPreviewRow) ([]Question, error) {
	questions := make([]Question, 0, len(rows))
	for _, row := range rows {
		resource := sql.NullString{}
		if row.Resource != "" {
			resource = sql.NullString{String: row.Resource, Valid: true}
		}

		question := Question{
			Question:            row.Question,
			Type:                row.Type,
			Options:             row.Options,
			Answers:             row.Answers,
			OfficialAnswer:      row.OfficialAnswer,
			AuthoritativeAnswer: row.AuthoritativeAnswer,
			AnswerReviewStatus:  row.AnswerReviewStatus,
			Points:              row.Points,
			DurationInSeconds:   row.DurationInSeconds,
			QuestionMedia:       row.QuestionMedia,
			OptionsMedia:        row.OptionsMedia,
			Resource:            resource,
			RevisionNumber:      1,
		}

		authority, operationalAnswers, err := ApplyAnswerAuthority(
			question.Answers,
			question.OfficialAnswer,
			question.AuthoritativeAnswer,
			question.AnswerReviewStatus,
			question.AnswerRevisionReason,
			question.AnswerRevisionSource,
		)
		if err != nil {
			return nil, err
		}

		question.Answers = operationalAnswers
		question.OfficialAnswer = authority.OfficialAnswer
		question.AuthoritativeAnswer = authority.AuthoritativeAnswer
		question.AnswerReviewStatus = authority.AnswerReviewStatus
		question.AnswerRevisionReason = authority.AnswerRevisionReason
		question.AnswerRevisionSource = authority.AnswerRevisionSource
		question.RevisionNumber = 1

		questions = append(questions, question)
	}

	return questions, nil
}

// BuildImportPreviewRow applies ADR-024 answer authority defaults for preview rows.
func BuildImportPreviewRow(rowNumber int, question Question) (ImportPreviewRow, error) {
	authority, operationalAnswers, err := ApplyAnswerAuthority(
		question.Answers,
		question.OfficialAnswer,
		question.AuthoritativeAnswer,
		question.AnswerReviewStatus,
		question.AnswerRevisionReason,
		question.AnswerRevisionSource,
	)
	if err != nil {
		return ImportPreviewRow{}, err
	}

	resource := ""
	if question.Resource.Valid {
		resource = question.Resource.String
	}

	return ImportPreviewRow{
		RowNumber:           rowNumber,
		Question:            question.Question,
		Type:                question.Type,
		Options:             question.Options,
		Answers:             operationalAnswers,
		OfficialAnswer:      authority.OfficialAnswer,
		AuthoritativeAnswer: authority.AuthoritativeAnswer,
		AnswerReviewStatus:  authority.AnswerReviewStatus,
		Points:              question.Points,
		DurationInSeconds:   question.DurationInSeconds,
		QuestionMedia:       question.QuestionMedia,
		OptionsMedia:        question.OptionsMedia,
		Resource:            resource,
		RevisionNumber:      1,
	}, nil
}
