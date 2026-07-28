package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goqu "github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"go.uber.org/zap"
)

var (
	ErrQuizNotFound              = errors.New("quiz not found")
	ErrInvalidReleasePolicy      = errors.New("invalid result release policy")
	ErrScheduledDateRequired     = errors.New("scheduled release policy requires a valid future timestamp")
	ErrManualReleaseNotPermitted = errors.New("manual release is not applicable for immediate release policy")
)

type QuizResultAdminService struct {
	db              *goqu.Database
	quizModel       *models.QuizModel
	auditModel      *models.QuizResultReleaseAuditModel
	analyticsSvc    *AssessmentAnalyticsService
	learnerCache    *LearnerAnalyticsCache
	instructorCache *InstructorAnalyticsCache
	logger          *zap.Logger
}

func NewQuizResultAdminService(db *goqu.Database, logger *zap.Logger) *QuizResultAdminService {
	return &QuizResultAdminService{
		db:           db,
		quizModel:    models.InitQuizModel(db),
		auditModel:   models.NewQuizResultReleaseAuditModel(db),
		analyticsSvc: NewAssessmentAnalyticsService(db, logger),
		logger:       logger,
	}
}

func (svc *QuizResultAdminService) SetLearnerAnalyticsCache(cache *LearnerAnalyticsCache) {
	svc.learnerCache = cache
	if svc.analyticsSvc != nil {
		svc.analyticsSvc.SetLearnerAnalyticsCache(cache)
	}
}

func (svc *QuizResultAdminService) SetInstructorAnalyticsCache(cache *InstructorAnalyticsCache) {
	svc.instructorCache = cache
	if svc.analyticsSvc != nil {
		svc.analyticsSvc.SetInstructorAnalyticsCache(cache)
	}
}

func (svc *QuizResultAdminService) GetReleaseStatus(quizID uuid.UUID, correlationID string) (structs.QuizResultReleaseStatusResponse, error) {
	if _, err := svc.analyticsSvc.EnsureScheduledReleaseEffective(quizID, correlationID); err != nil && !errors.Is(err, ErrQuizNotFound) {
		svc.logger.Warn("scheduled release evaluator failed during status read", zap.Error(err), zap.String("quiz_id", quizID.String()))
	}

	quiz, err := svc.quizModel.GetSelfPacedMetaByID(quizID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return structs.QuizResultReleaseStatusResponse{}, ErrQuizNotFound
		}
		return structs.QuizResultReleaseStatusResponse{}, err
	}

	// Count total submitted attempts for this quiz
	count, err := svc.db.From("assessment_attempts").
		Where(goqu.Ex{
			"quiz_id": quizID,
			"status":  []string{models.AttemptStatusSubmitted, models.AttemptStatusAutoSubmitted},
		}).
		Count()
	if err != nil {
		svc.logger.Error("GetReleaseStatus count error", zap.Error(err))
		return structs.QuizResultReleaseStatusResponse{}, err
	}

	now := time.Now().UTC()
	isCurrentlyReleased := false
	policy := quiz.ResultReleasePolicy
	if policy == "" {
		policy = string(structs.ResultReleasePolicyImmediate)
	}

	switch policy {
	case string(structs.ResultReleasePolicyImmediate):
		isCurrentlyReleased = true
	case string(structs.ResultReleasePolicyManual):
		isCurrentlyReleased = quiz.ResultsReleased
	case string(structs.ResultReleasePolicyScheduled):
		if quiz.ResultsReleased {
			isCurrentlyReleased = true
		} else if quiz.ResultsScheduledAt.Valid && !quiz.ResultsScheduledAt.Time.IsZero() {
			schedTime := quiz.ResultsScheduledAt.Time.UTC()
			if now.Equal(schedTime) || now.After(schedTime) {
				isCurrentlyReleased = true
			}
		}
	default:
		isCurrentlyReleased = false
	}

	var scheduledAtStr *string
	if quiz.ResultsScheduledAt.Valid && !quiz.ResultsScheduledAt.Time.IsZero() {
		formatted := quiz.ResultsScheduledAt.Time.UTC().Format(time.RFC3339)
		scheduledAtStr = &formatted
	}

	var releasedAtStr *string
	if quiz.ResultsReleasedAt.Valid && !quiz.ResultsReleasedAt.Time.IsZero() {
		formatted := quiz.ResultsReleasedAt.Time.UTC().Format(time.RFC3339)
		releasedAtStr = &formatted
	}

	return structs.QuizResultReleaseStatusResponse{
		QuizID:                 quiz.ID.String(),
		ResultReleasePolicy:    policy,
		ResultsReleased:        quiz.ResultsReleased,
		ResultsScheduledAt:     scheduledAtStr,
		ResultsReleasedAt:      releasedAtStr,
		IsCurrentlyReleased:    isCurrentlyReleased,
		ShowScore:              quiz.ShowScore,
		ShowPassFail:           quiz.ShowPassFail,
		AllowAnswerReview:      quiz.AllowAnswerReview,
		ShowCorrectness:        quiz.ShowCorrectness,
		ShowExplanations:       quiz.ShowExplanations,
		TotalSubmittedAttempts: int(count),
	}, nil
}

func (svc *QuizResultAdminService) UpdateResultSettings(
	quizID uuid.UUID,
	actorID string,
	req structs.UpdateQuizResultSettingsRequest,
	ipAddress, userAgent, correlationID string,
) (structs.QuizResultReleaseStatusResponse, error) {
	// Execute in transaction with SELECT FOR UPDATE row lock
	tx, err := svc.db.Begin()
	if err != nil {
		return structs.QuizResultReleaseStatusResponse{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Lock quiz row
	var currentQuiz models.QuizSelfPacedMeta
	found, err := tx.From("quizzes").
		Select(
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		).
		Where(goqu.Ex{"id": quizID}).
		ForUpdate(goqu.Wait).
		ScanStruct(&currentQuiz)
	if err != nil {
		return structs.QuizResultReleaseStatusResponse{}, err
	}
	if !found {
		return structs.QuizResultReleaseStatusResponse{}, ErrQuizNotFound
	}

	newPolicy := string(req.ResultReleasePolicy)
	if newPolicy == "" {
		newPolicy = currentQuiz.ResultReleasePolicy
	}
	if newPolicy != string(structs.ResultReleasePolicyImmediate) &&
		newPolicy != string(structs.ResultReleasePolicyManual) &&
		newPolicy != string(structs.ResultReleasePolicyScheduled) {
		return structs.QuizResultReleaseStatusResponse{}, ErrInvalidReleasePolicy
	}

	now := time.Now().UTC()
	var newScheduledAt sql.NullTime
	var newReleasedAt sql.NullTime = currentQuiz.ResultsReleasedAt
	newResultsReleased := currentQuiz.ResultsReleased

	switch newPolicy {
	case string(structs.ResultReleasePolicyImmediate):
		newScheduledAt = sql.NullTime{Valid: false}
		newResultsReleased = true
		newReleasedAt = sql.NullTime{Time: now, Valid: true}
	case string(structs.ResultReleasePolicyManual):
		newScheduledAt = sql.NullTime{Valid: false}
	case string(structs.ResultReleasePolicyScheduled):
		if req.ResultsScheduledAt != nil && *req.ResultsScheduledAt != "" {
			t, parseErr := time.Parse(time.RFC3339, *req.ResultsScheduledAt)
			if parseErr != nil {
				return structs.QuizResultReleaseStatusResponse{}, ErrScheduledDateRequired
			}
			newScheduledAt = sql.NullTime{Time: t.UTC(), Valid: true}
		} else if currentQuiz.ResultsScheduledAt.Valid {
			newScheduledAt = currentQuiz.ResultsScheduledAt
		} else {
			return structs.QuizResultReleaseStatusResponse{}, ErrScheduledDateRequired
		}
	}

	newShowScore := currentQuiz.ShowScore
	if req.ShowScore != nil {
		newShowScore = *req.ShowScore
	}
	newShowPassFail := currentQuiz.ShowPassFail
	if req.ShowPassFail != nil {
		newShowPassFail = *req.ShowPassFail
	}
	newAllowAnswerReview := currentQuiz.AllowAnswerReview
	if req.AllowAnswerReview != nil {
		newAllowAnswerReview = *req.AllowAnswerReview
	}
	newShowCorrectness := currentQuiz.ShowCorrectness
	if req.ShowCorrectness != nil {
		newShowCorrectness = *req.ShowCorrectness
	}
	newShowExplanations := currentQuiz.ShowExplanations
	if req.ShowExplanations != nil {
		newShowExplanations = *req.ShowExplanations
	}

	// Capture previous state for audit log
	prevStateJSON, _ := json.Marshal(map[string]interface{}{
		"policy":               currentQuiz.ResultReleasePolicy,
		"results_released":     currentQuiz.ResultsReleased,
		"results_scheduled_at": currentQuiz.ResultsScheduledAt,
		"show_score":           currentQuiz.ShowScore,
		"show_pass_fail":       currentQuiz.ShowPassFail,
		"allow_answer_review":  currentQuiz.AllowAnswerReview,
		"show_correctness":     currentQuiz.ShowCorrectness,
		"show_explanations":    currentQuiz.ShowExplanations,
	})

	newStateMap := map[string]interface{}{
		"policy":               newPolicy,
		"results_released":     newResultsReleased,
		"results_scheduled_at": newScheduledAt,
		"show_score":           newShowScore,
		"show_pass_fail":       newShowPassFail,
		"allow_answer_review":  newAllowAnswerReview,
		"show_correctness":     newShowCorrectness,
		"show_explanations":    newShowExplanations,
	}
	newStateJSON, _ := json.Marshal(newStateMap)

	// Update DB
	_, err = tx.Update("quizzes").
		Set(goqu.Record{
			"result_release_policy": newPolicy,
			"results_released":      newResultsReleased,
			"results_scheduled_at":  newScheduledAt,
			"results_released_at":   newReleasedAt,
			"show_score":            newShowScore,
			"show_pass_fail":        newShowPassFail,
			"allow_answer_review":   newAllowAnswerReview,
			"show_correctness":      newShowCorrectness,
			"show_explanations":     newShowExplanations,
			"updated_at":            now,
		}).
		Where(goqu.Ex{"id": quizID}).
		Executor().Exec()
	if err != nil {
		return structs.QuizResultReleaseStatusResponse{}, fmt.Errorf("update quiz settings: %w", err)
	}

	// Audit Log
	_, err = tx.Insert(models.QuizResultReleaseAuditLogsTable).Rows(goqu.Record{
		"id":              uuid.New(),
		"quiz_id":         quizID,
		"actor_id":        actorID,
		"event_type":      structs.AuditPolicyUpdate,
		"previous_policy": sql.NullString{String: currentQuiz.ResultReleasePolicy, Valid: true},
		"new_policy":      sql.NullString{String: newPolicy, Valid: true},
		"previous_state":  sql.NullString{String: string(prevStateJSON), Valid: true},
		"new_state":       sql.NullString{String: string(newStateJSON), Valid: true},
		"ip_address":      sql.NullString{String: ipAddress, Valid: ipAddress != ""},
		"user_agent":      sql.NullString{String: userAgent, Valid: userAgent != ""},
		"correlation_id":  sql.NullString{String: correlationID, Valid: correlationID != ""},
		"schema_version":  1,
		"created_at":      now,
	}).Executor().Exec()
	if err != nil {
		return structs.QuizResultReleaseStatusResponse{}, fmt.Errorf("insert audit log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return structs.QuizResultReleaseStatusResponse{}, fmt.Errorf("commit transaction: %w", err)
	}

	if svc.instructorCache != nil {
		svc.instructorCache.BumpQuizVersion(quizID.String())
		if currentQuiz.CreatorID.Valid {
			svc.instructorCache.BumpInstructorVersion(currentQuiz.CreatorID.String)
		}
	}

	return svc.GetReleaseStatus(quizID, correlationID)
}

func (svc *QuizResultAdminService) ReleaseResults(
	quizID uuid.UUID,
	actorID string,
	ipAddress, userAgent, correlationID string,
) (structs.QuizResultReleaseStatusResponse, error) {
	tx, err := svc.db.Begin()
	if err != nil {
		return structs.QuizResultReleaseStatusResponse{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var currentQuiz models.QuizSelfPacedMeta
	found, err := tx.From("quizzes").
		Select(
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		).
		Where(goqu.Ex{"id": quizID}).
		ForUpdate(goqu.Wait).
		ScanStruct(&currentQuiz)
	if err != nil {
		return structs.QuizResultReleaseStatusResponse{}, err
	}
	if !found {
		return structs.QuizResultReleaseStatusResponse{}, ErrQuizNotFound
	}

	if currentQuiz.ResultReleasePolicy == string(structs.ResultReleasePolicyImmediate) {
		return structs.QuizResultReleaseStatusResponse{}, ErrManualReleaseNotPermitted
	}

	now := time.Now().UTC()

	// Idempotent: If already released, return success without duplicate state mutation
	releasedNow := false
	if !currentQuiz.ResultsReleased {
		releasedNow = true
		prevStateJSON, _ := json.Marshal(map[string]interface{}{
			"results_released":    currentQuiz.ResultsReleased,
			"results_released_at": currentQuiz.ResultsReleasedAt,
		})

		newStateJSON, _ := json.Marshal(map[string]interface{}{
			"results_released":    true,
			"results_released_at": now,
		})

		_, err = tx.Update("quizzes").
			Set(goqu.Record{
				"results_released":    true,
				"results_released_at": now,
				"updated_at":          now,
			}).
			Where(goqu.Ex{"id": quizID}).
			Executor().Exec()
		if err != nil {
			return structs.QuizResultReleaseStatusResponse{}, fmt.Errorf("update release results: %w", err)
		}

		_, err = tx.Insert(models.QuizResultReleaseAuditLogsTable).Rows(goqu.Record{
			"id":              uuid.New(),
			"quiz_id":         quizID,
			"actor_id":        actorID,
			"event_type":      structs.AuditManualRelease,
			"previous_policy": sql.NullString{String: currentQuiz.ResultReleasePolicy, Valid: true},
			"new_policy":      sql.NullString{String: currentQuiz.ResultReleasePolicy, Valid: true},
			"previous_state":  sql.NullString{String: string(prevStateJSON), Valid: true},
			"new_state":       sql.NullString{String: string(newStateJSON), Valid: true},
			"ip_address":      sql.NullString{String: ipAddress, Valid: ipAddress != ""},
			"user_agent":      sql.NullString{String: userAgent, Valid: userAgent != ""},
			"correlation_id":  sql.NullString{String: correlationID, Valid: correlationID != ""},
			"schema_version":  1,
			"created_at":      now,
		}).Executor().Exec()
		if err != nil {
			return structs.QuizResultReleaseStatusResponse{}, fmt.Errorf("insert manual release audit log: %w", err)
		}

		quizOwnerID := ""
		if currentQuiz.CreatorID.Valid {
			quizOwnerID = currentQuiz.CreatorID.String
		}
		overrideEvent, buildErr := svc.analyticsSvc.buildEvent(
			uuid.Nil,
			quizID,
			actorID,
			quizOwnerID,
			structs.EventResultReleaseOverrideApplied,
			structs.EventSourceHTTP,
			correlationID,
			ReleaseOverrideIdempotencyKey(quizID, now),
			nil,
			map[string]interface{}{
				"policy": currentQuiz.ResultReleasePolicy,
			},
			now,
		)
		if buildErr != nil {
			return structs.QuizResultReleaseStatusResponse{}, buildErr
		}
		overrideEvent.AttemptRefID = uuid.NullUUID{}
		overrideEvent.QuizRefID = uuid.NullUUID{UUID: quizID, Valid: true}
		if _, _, err := svc.analyticsSvc.analyticsModel.CreateEventTx(tx, overrideEvent); err != nil {
			return structs.QuizResultReleaseStatusResponse{}, fmt.Errorf("insert release override analytics: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return structs.QuizResultReleaseStatusResponse{}, fmt.Errorf("commit transaction: %w", err)
	}

	if releasedNow {
		svc.analyticsSvc.bumpQuizLearnerCaches(quizID)
	}
	if svc.instructorCache != nil {
		svc.instructorCache.BumpQuizVersion(quizID.String())
		if currentQuiz.CreatorID.Valid {
			svc.instructorCache.BumpInstructorVersion(currentQuiz.CreatorID.String)
		}
	}

	return svc.GetReleaseStatus(quizID, correlationID)
}
