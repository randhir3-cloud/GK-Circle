package services

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

type QuestionImportService struct {
	quizModel      *models.QuizModel
	importJobModel *models.QuestionImportJobModel
	questionModel  *models.QuestionModel
	db             *goqu.Database
	logger         *zap.Logger
}

func NewQuestionImportService(db *goqu.Database, logger *zap.Logger) *QuestionImportService {
	return &QuestionImportService{
		quizModel:      models.InitQuizModel(db),
		importJobModel: models.InitQuestionImportJobModel(db, logger),
		questionModel:  models.InitQuestionModel(db, logger),
		db:             db,
		logger:         logger,
	}
}

// CreatePreviewJob validates a CSV file for a quiz and stores a preview job
// without inserting questions into the bank.
func (svc *QuestionImportService) CreatePreviewJob(
	quizID string,
	createdBy string,
	filePath string,
	questionTimeLimit string,
) (models.QuestionImportJob, error) {
	if _, err := svc.quizModel.GetQuizById(quizID); err != nil {
		return models.QuestionImportJob{}, err
	}

	csvQuestions, err := utils.ValidateCSVFileFormat(filePath)
	if err != nil {
		return models.QuestionImportJob{}, err
	}

	previewResult, err := utils.PreviewQuestionsFromCSV(csvQuestions, questionTimeLimit)
	if err != nil {
		return models.QuestionImportJob{}, err
	}

	existing, err := svc.questionModel.ListImportFingerprintIndexByQuizID(quizID)
	if err != nil {
		return models.QuestionImportJob{}, err
	}

	validRows, rowErrors := models.ApplyImportDuplicateDetection(
		previewResult.ValidRows,
		previewResult.Errors,
		existing,
	)

	payload := models.ImportPreviewPayload{
		ValidRows: validRows,
		Errors:    rowErrors,
	}

	return svc.importJobModel.CreatePreviewJob(
		quizID,
		createdBy,
		filepath.Base(filePath),
		payload,
	)
}

func (svc *QuestionImportService) GetPreviewJob(quizID, jobID string) (models.QuestionImportJob, error) {
	return svc.importJobModel.GetByQuizAndID(quizID, jobID)
}

// FilterLegacyImportDuplicates removes duplicate rows from a legacy CSV import batch.
func (svc *QuestionImportService) FilterLegacyImportDuplicates(
	quizID string,
	questions []models.Question,
) ([]models.Question, []models.ImportRowError, error) {
	existing, err := svc.questionModel.ListImportFingerprintIndexByQuizID(quizID)
	if err != nil {
		return nil, nil, err
	}

	filtered, rowErrors := models.FilterQuestionsByQuizDuplicates(questions, existing)
	return filtered, rowErrors, nil
}

// CommitPreviewJob persists eligible preview rows through the shared question bank path.
// Valid rows commit transactionally (all-or-nothing). Repeated commits are idempotent.
func (svc *QuestionImportService) CommitPreviewJob(quizID, jobID string) (models.QuestionImportJob, error) {
	if _, err := svc.quizModel.GetQuizById(quizID); err != nil {
		return models.QuestionImportJob{}, err
	}

	existing, err := svc.importJobModel.GetByQuizAndID(quizID, jobID)
	if err != nil {
		return models.QuestionImportJob{}, err
	}

	switch existing.Status {
	case models.ImportJobStatusCommitted:
		return existing, nil
	case models.ImportJobStatusCommitting:
		return models.QuestionImportJob{}, models.ErrImportJobCommitInProgress
	case models.ImportJobStatusFailed:
		if existing.ValidRowCount == 0 {
			return models.QuestionImportJob{}, models.ErrImportJobNotCommitable
		}
	case models.ImportJobStatusPreviewed:
		if existing.ValidRowCount == 0 {
			return models.QuestionImportJob{}, models.ErrImportJobNoValidRows
		}
	default:
		return models.QuestionImportJob{}, models.ErrImportJobNotCommitable
	}

	quizFingerprints, err := svc.questionModel.ListImportFingerprintIndexByQuizID(quizID)
	if err != nil {
		return models.QuestionImportJob{}, err
	}
	if duplicateErrors := models.FindImportDuplicateErrors(existing.Preview.ValidRows, quizFingerprints); len(duplicateErrors) > 0 {
		return models.QuestionImportJob{}, fmt.Errorf("%w: %s", models.ErrImportCommitDuplicates, formatImportRowErrors(duplicateErrors))
	}

	questions, err := models.QuestionsFromImportPreviewRows(existing.Preview.ValidRows)
	if err != nil {
		return models.QuestionImportJob{}, err
	}
	if len(questions) == 0 {
		return models.QuestionImportJob{}, models.ErrImportJobNoValidRows
	}

	isOk := false
	claimed := false
	var commitFailure error
	transaction, err := svc.db.Begin()
	if err != nil {
		return models.QuestionImportJob{}, err
	}

	defer func() {
		if isOk {
			if commitErr := transaction.Commit(); commitErr != nil {
				svc.logger.Error("import commit transaction failed", zap.Error(commitErr))
			}
			return
		}

		if rollbackErr := transaction.Rollback(); rollbackErr != nil {
			svc.logger.Error("import commit rollback failed", zap.Error(rollbackErr))
		}

		if claimed && commitFailure != nil {
			if markErr := svc.importJobModel.MarkCommitFailed(quizID, jobID, commitFailure.Error()); markErr != nil {
				svc.logger.Error("failed to mark import job commit failure", zap.Error(markErr))
			}
		}
	}()

	_, claimed, err = svc.importJobModel.TryClaimForCommit(transaction, quizID, jobID)
	if err != nil {
		return models.QuestionImportJob{}, err
	}
	if !claimed {
		transaction.Rollback()
		isOk = true

		current, loadErr := svc.importJobModel.GetByQuizAndID(quizID, jobID)
		if loadErr != nil {
			return models.QuestionImportJob{}, loadErr
		}
		switch current.Status {
		case models.ImportJobStatusCommitted:
			return current, nil
		case models.ImportJobStatusCommitting:
			return models.QuestionImportJob{}, models.ErrImportJobCommitInProgress
		default:
			return models.QuestionImportJob{}, models.ErrImportJobNotCommitable
		}
	}

	ids, err := svc.questionModel.AppendQuestionsToQuiz(transaction, quizID, questions)
	if err != nil {
		commitFailure = err
		return models.QuestionImportJob{}, err
	}

	questionIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		questionIDs = append(questionIDs, id.String())
	}

	result := models.ImportCommitResult{
		QuestionIDs:    questionIDs,
		CommittedCount: len(questionIDs),
	}

	committedJob, err := svc.importJobModel.FinalizeCommit(transaction, quizID, jobID, result)
	if err != nil {
		commitFailure = err
		return models.QuestionImportJob{}, err
	}

	isOk = true
	return committedJob, nil
}

func formatImportRowErrors(rowErrors []models.ImportRowError) string {
	parts := make([]string, 0, len(rowErrors))
	for _, rowError := range rowErrors {
		parts = append(parts, fmt.Sprintf("row %d: %s", rowError.RowNumber, strings.Join(rowError.Messages, "; ")))
	}
	return strings.Join(parts, "; ")
}

// FormatImportDuplicateErrors renders duplicate row errors for legacy import responses.
func FormatImportDuplicateErrors(rowErrors []models.ImportRowError) string {
	return formatImportRowErrors(rowErrors)
}
