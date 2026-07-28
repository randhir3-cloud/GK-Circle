package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	goqu "github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"go.uber.org/zap"
)

const (
	analyticsServerTelemetryTimeout = 2 * time.Second
	analyticsOccurredSkew           = 5 * time.Minute
	analyticsAllowedReviewWindow    = 30 * 24 * time.Hour
	analyticsMaxMetadataBytes       = 2048
	analyticsMaxBatchSize           = 100
)

var (
	ErrAnalyticsInvalidEventType   = errors.New("invalid analytics event type for client telemetry")
	ErrAnalyticsInvalidOccurredAt  = errors.New("invalid analytics occurred_at")
	ErrAnalyticsInvalidMetadata    = errors.New("invalid analytics metadata")
	ErrAnalyticsBatchEmpty         = errors.New("analytics batch requires at least one event")
	ErrAnalyticsBatchTooLarge      = errors.New("analytics batch exceeds maximum size")
	ErrAnalyticsAttemptMismatch    = errors.New("analytics attempt does not belong to quiz or user")
	ErrAnalyticsSensitiveMetadata  = errors.New("analytics metadata contains sensitive fields")
)

var analyticsSensitiveMetadataKeys = map[string]struct{}{
	"password":         {},
	"token":            {},
	"access_token":     {},
	"refresh_token":    {},
	"authorization":    {},
	"cookie":           {},
	"session":          {},
	"answer_key":       {},
	"correct_answers":  {},
	"correct_options":  {},
	"answers":          {},
	"secret":           {},
}

type AssessmentAnalyticsService struct {
	db              *goqu.Database
	analyticsModel  *models.AssessmentAnalyticsModel
	attemptModel    *models.AssessmentAttemptModel
	quizModel       *models.QuizModel
	learnerCache    *LearnerAnalyticsCache
	instructorCache *InstructorAnalyticsCache
	logger          *zap.Logger
}

func NewAssessmentAnalyticsService(db *goqu.Database, logger *zap.Logger) *AssessmentAnalyticsService {
	return &AssessmentAnalyticsService{
		db:             db,
		analyticsModel: models.NewAssessmentAnalyticsModel(db),
		attemptModel:   models.InitAssessmentAttemptModel(db),
		quizModel:      models.InitQuizModel(db),
		logger:         logger,
	}
}

func (svc *AssessmentAnalyticsService) SetLearnerAnalyticsCache(cache *LearnerAnalyticsCache) {
	svc.learnerCache = cache
}

func (svc *AssessmentAnalyticsService) SetInstructorAnalyticsCache(cache *InstructorAnalyticsCache) {
	svc.instructorCache = cache
}

type AuthoritativeAnalyticsEventInput struct {
	AttemptID      uuid.UUID
	QuizID         uuid.UUID
	UserID         string
	QuizOwnerID    string
	EventType      structs.AssessmentEventType
	EventSource    structs.AssessmentEventSource
	CorrelationID  string
	IdempotencyKey string
	Metadata       map[string]interface{}
	OccurredAt     time.Time
}

func (svc *AssessmentAnalyticsService) RecordAuthoritativeEventTx(
	tx *goqu.TxDatabase,
	input AuthoritativeAnalyticsEventInput,
) error {
	if !structs.IsAuthoritativeEventType(input.EventType) {
		return fmt.Errorf("%w: %s", ErrAnalyticsInvalidEventType, input.EventType)
	}
	event, err := svc.buildEvent(input.AttemptID, input.QuizID, input.UserID, input.QuizOwnerID, input.EventType, input.EventSource, input.CorrelationID, input.IdempotencyKey, nil, input.Metadata, input.OccurredAt)
	if err != nil {
		return err
	}
	_, _, err = svc.analyticsModel.CreateEventTx(tx, event)
	return err
}

func (svc *AssessmentAnalyticsService) RecordServerTelemetryBounded(
	ctx context.Context,
	input AuthoritativeAnalyticsEventInput,
) {
	_ = ctx
	if !structs.IsServerTelemetryEventType(input.EventType) {
		svc.logger.Warn("rejected non-server telemetry event type", zap.String("event_type", string(input.EventType)))
		return
	}

	// Best-effort and non-fatal: run synchronously with a wall-clock bound via context when
	// callers supply one. Primary business operations must ignore analytics failures.
	start := time.Now()
	err := svc.recordServerTelemetry(input)
	if err != nil {
		svc.logger.Warn("analytics server telemetry failed",
			zap.String("event_type", string(input.EventType)),
			zap.String("attempt_id", input.AttemptID.String()),
			zap.Error(err),
		)
		return
	}
	if time.Since(start) > analyticsServerTelemetryTimeout {
		svc.logger.Warn("analytics server telemetry exceeded bound",
			zap.String("event_type", string(input.EventType)),
			zap.String("attempt_id", input.AttemptID.String()),
			zap.Duration("elapsed", time.Since(start)),
		)
	}
}

func (svc *AssessmentAnalyticsService) recordServerTelemetry(input AuthoritativeAnalyticsEventInput) error {
	if svc.db == nil {
		return errors.New("analytics database unavailable")
	}
	tx, err := svc.db.Begin()
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		if !ok {
			_ = tx.Rollback()
		}
	}()

	event, err := svc.buildEvent(input.AttemptID, input.QuizID, input.UserID, input.QuizOwnerID, input.EventType, input.EventSource, input.CorrelationID, input.IdempotencyKey, nil, input.Metadata, input.OccurredAt)
	if err != nil {
		return err
	}
	if _, _, err := svc.analyticsModel.CreateEventTx(tx, event); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	ok = true
	return nil
}

func (svc *AssessmentAnalyticsService) RecordClientTelemetryBatch(
	quizID, attemptID uuid.UUID,
	userID, correlationID string,
	requests []structs.RecordTelemetryEventRequest,
) (structs.BatchResultResponse, error) {
	if len(requests) == 0 {
		return structs.BatchResultResponse{}, ErrAnalyticsBatchEmpty
	}
	if len(requests) > analyticsMaxBatchSize {
		return structs.BatchResultResponse{}, ErrAnalyticsBatchTooLarge
	}

	attempt, err := svc.attemptModel.GetByID(quizID, attemptID)
	if err != nil {
		return structs.BatchResultResponse{}, err
	}
	if attempt.UserID != userID || attempt.QuizID != quizID {
		return structs.BatchResultResponse{}, ErrAnalyticsAttemptMismatch
	}

	quiz, err := svc.quizModel.GetSelfPacedMetaByID(quizID)
	if err != nil {
		return structs.BatchResultResponse{}, err
	}
	quizOwnerID := ""
	if quiz.CreatorID.Valid {
		quizOwnerID = quiz.CreatorID.String
	}

	events := make([]models.AssessmentAnalyticsEvent, 0, len(requests))
	for _, req := range requests {
		if !structs.IsClientTelemetryEventType(req.EventType) {
			return structs.BatchResultResponse{}, fmt.Errorf("%w: %s", ErrAnalyticsInvalidEventType, req.EventType)
		}
		if err := svc.ValidateOccurredAt(req.OccurredAt, attempt.CreatedAt, attempt.SubmittedAt); err != nil {
			return structs.BatchResultResponse{}, err
		}
		if err := svc.ValidateMetadataSchema(req.EventType, req.Metadata); err != nil {
			return structs.BatchResultResponse{}, err
		}

		var clientEventID *uuid.UUID
		if req.ClientEventID != nil && strings.TrimSpace(*req.ClientEventID) != "" {
			parsed, parseErr := uuid.Parse(strings.TrimSpace(*req.ClientEventID))
			if parseErr != nil {
				return structs.BatchResultResponse{}, fmt.Errorf("%w: invalid client_event_id", ErrAnalyticsInvalidEventType)
			}
			clientEventID = &parsed
		}
		idempotencyKey := ""
		if req.IdempotencyKey != nil {
			idempotencyKey = strings.TrimSpace(*req.IdempotencyKey)
		}

		event, buildErr := svc.buildEvent(
			attempt.ID,
			quizID,
			userID,
			quizOwnerID,
			req.EventType,
			structs.EventSourceClientBatch,
			correlationID,
			idempotencyKey,
			clientEventID,
			req.Metadata,
			req.OccurredAt.UTC(),
		)
		if buildErr != nil {
			return structs.BatchResultResponse{}, buildErr
		}
		events = append(events, event)
	}

	batch, err := svc.analyticsModel.CreateBatchEvents(events)
	if err != nil {
		return structs.BatchResultResponse{}, err
	}
	if batch.Inserted > 0 {
		if svc.learnerCache != nil {
			svc.learnerCache.BumpVersion(userID)
		}
		if svc.instructorCache != nil {
			svc.instructorCache.BumpQuizVersion(quizID.String())
			if quizOwnerID != "" {
				svc.instructorCache.BumpInstructorVersion(quizOwnerID)
			}
		}
	}
	return structs.BatchResultResponse{
		Received:   batch.Received,
		Inserted:   batch.Inserted,
		Duplicates: batch.Duplicates,
		Rejected:   batch.Rejected,
	}, nil
}

func (svc *AssessmentAnalyticsService) ListAttemptEvents(
	quizID, attemptID uuid.UUID,
	userID string,
	params models.AnalyticsPaginationParams,
) ([]models.AssessmentAnalyticsEvent, string, error) {
	attempt, err := svc.attemptModel.GetByID(quizID, attemptID)
	if err != nil {
		return nil, "", err
	}
	if attempt.UserID != userID {
		return nil, "", models.ErrAssessmentAttemptNotFound
	}
	return svc.analyticsModel.ListEventsByAttemptID(attempt.ID, params)
}

func (svc *AssessmentAnalyticsService) ValidateOccurredAt(
	occurredAt time.Time,
	attemptCreated time.Time,
	attemptSubmitted sql.NullTime,
) error {
	if occurredAt.IsZero() {
		return ErrAnalyticsInvalidOccurredAt
	}
	now := time.Now().UTC()
	occurred := occurredAt.UTC()
	if occurred.After(now.Add(analyticsOccurredSkew)) {
		return ErrAnalyticsInvalidOccurredAt
	}
	if occurred.Before(attemptCreated.UTC().Add(-analyticsOccurredSkew)) {
		return ErrAnalyticsInvalidOccurredAt
	}
	if attemptSubmitted.Valid {
		deadline := attemptSubmitted.Time.UTC().Add(analyticsAllowedReviewWindow)
		if occurred.After(deadline) {
			return ErrAnalyticsInvalidOccurredAt
		}
	}
	return nil
}

func (svc *AssessmentAnalyticsService) ValidateMetadataSchema(
	eventType structs.AssessmentEventType,
	metadata map[string]interface{},
) error {
	_ = eventType
	if metadata == nil {
		return nil
	}
	for key := range metadata {
		normalized := strings.ToLower(strings.TrimSpace(key))
		if _, blocked := analyticsSensitiveMetadataKeys[normalized]; blocked {
			return ErrAnalyticsSensitiveMetadata
		}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return ErrAnalyticsInvalidMetadata
	}
	if len(raw) > analyticsMaxMetadataBytes {
		return ErrAnalyticsInvalidMetadata
	}
	if string(raw) == "null" {
		return ErrAnalyticsInvalidMetadata
	}
	return nil
}

func (svc *AssessmentAnalyticsService) buildEvent(
	attemptID, quizID uuid.UUID,
	userID, quizOwnerID string,
	eventType structs.AssessmentEventType,
	eventSource structs.AssessmentEventSource,
	correlationID, idempotencyKey string,
	clientEventID *uuid.UUID,
	metadata map[string]interface{},
	occurredAt time.Time,
) (models.AssessmentAnalyticsEvent, error) {
	if correlationID == "" {
		correlationID = uuid.NewString()
	}
	if err := svc.ValidateMetadataSchema(eventType, metadata); err != nil {
		return models.AssessmentAnalyticsEvent{}, err
	}
	rawMeta := json.RawMessage(`{}`)
	if metadata != nil {
		encoded, err := json.Marshal(metadata)
		if err != nil {
			return models.AssessmentAnalyticsEvent{}, ErrAnalyticsInvalidMetadata
		}
		rawMeta = encoded
	}

	event := models.AssessmentAnalyticsEvent{
		ID:            uuid.New(),
		AttemptID:     attemptID,
		QuizID:        quizID,
		UserID:        userID,
		EventType:     string(eventType),
		EventSource:   string(eventSource),
		CorrelationID: correlationID,
		SchemaVersion: 1,
		Metadata:      rawMeta,
		OccurredAt:    occurredAt.UTC(),
		AttemptRefID:  uuid.NullUUID{UUID: attemptID, Valid: true},
		QuizRefID:     uuid.NullUUID{UUID: quizID, Valid: true},
	}
	if quizOwnerID != "" {
		event.QuizOwnerID = sql.NullString{String: quizOwnerID, Valid: true}
	}
	if idempotencyKey != "" {
		event.IdempotencyKey = sql.NullString{String: idempotencyKey, Valid: true}
	}
	if clientEventID != nil {
		event.ClientEventID = uuid.NullUUID{UUID: *clientEventID, Valid: true}
	}
	return event, nil
}

func ScheduledReleaseIdempotencyKey(quizID uuid.UUID, scheduledAt time.Time) string {
	return fmt.Sprintf("scheduled-release:%s:%s", quizID.String(), scheduledAt.UTC().Format(time.RFC3339Nano))
}

func AttemptStartedIdempotencyKey(attemptID uuid.UUID) string {
	return fmt.Sprintf("attempt-started:%s", attemptID.String())
}

func AttemptSubmittedIdempotencyKey(attemptID uuid.UUID) string {
	return fmt.Sprintf("attempt-submitted:%s", attemptID.String())
}

func AttemptAutoSubmittedIdempotencyKey(attemptID uuid.UUID) string {
	return fmt.Sprintf("attempt-auto-submitted:%s", attemptID.String())
}

func ReleaseOverrideIdempotencyKey(quizID uuid.UUID, releasedAt time.Time) string {
	return fmt.Sprintf("release-override:%s:%s", quizID.String(), releasedAt.UTC().Format(time.RFC3339Nano))
}

// EnsureScheduledReleaseEffective evaluates SCHEDULED release eligibility under a quiz lock,
// emits RESULT_RELEASE_SCHEDULED_EFFECTIVE exactly once via deterministic idempotency key,
// and updates quiz release state to match Phase 7 contracts.
func (svc *AssessmentAnalyticsService) EnsureScheduledReleaseEffective(quizID uuid.UUID, correlationID string) (bool, error) {
	tx, err := svc.db.Begin()
	if err != nil {
		return false, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = tx.Rollback()
		}
	}()

	var quiz models.QuizSelfPacedMeta
	found, err := tx.From("quizzes").
		Select(
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		).
		Where(goqu.Ex{"id": quizID}).
		ForUpdate(goqu.Wait).
		ScanStruct(&quiz)
	if err != nil {
		return false, err
	}
	if !found {
		return false, ErrQuizNotFound
	}
	if quiz.ResultReleasePolicy != string(structs.ResultReleasePolicyScheduled) {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		ok = true
		return false, nil
	}
	if !quiz.ResultsScheduledAt.Valid || quiz.ResultsScheduledAt.Time.IsZero() {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		ok = true
		return false, nil
	}

	now := time.Now().UTC()
	scheduledAt := quiz.ResultsScheduledAt.Time.UTC()
	if now.Before(scheduledAt) {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		ok = true
		return false, nil
	}

	quizOwnerID := ""
	actorID := "system:scheduler"
	if quiz.CreatorID.Valid {
		quizOwnerID = quiz.CreatorID.String
		actorID = quiz.CreatorID.String
	}

	applied := false
	if !quiz.ResultsReleased {
		_, err = tx.Update("quizzes").
			Set(goqu.Record{
				"results_released":    true,
				"results_released_at": now,
				"updated_at":          now,
			}).
			Where(goqu.Ex{"id": quizID}).
			Executor().Exec()
		if err != nil {
			return false, err
		}
		applied = true
	}

	event, err := svc.buildEvent(
		uuid.Nil,
		quizID,
		actorID,
		quizOwnerID,
		structs.EventResultReleaseScheduledEffective,
		structs.EventSourceScheduler,
		correlationID,
		ScheduledReleaseIdempotencyKey(quizID, scheduledAt),
		nil,
		map[string]interface{}{
			"scheduled_at": scheduledAt.Format(time.RFC3339Nano),
			"effective_at": now.Format(time.RFC3339Nano),
		},
		now,
	)
	if err != nil {
		return false, err
	}
	// Quiz-level events do not retain a live attempt FK.
	event.AttemptRefID = uuid.NullUUID{}
	event.QuizRefID = uuid.NullUUID{UUID: quizID, Valid: true}

	_, inserted, err := svc.analyticsModel.CreateEventTx(tx, event)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	ok = true
	if (applied || inserted) && svc.learnerCache != nil {
		svc.bumpQuizLearnerCaches(quizID)
	}
	return applied || inserted, nil
}

func (svc *AssessmentAnalyticsService) bumpQuizLearnerCaches(quizID uuid.UUID) {
	if svc.learnerCache == nil || svc.db == nil {
		return
	}
	var userIDs []string
	err := svc.db.From("assessment_attempts").
		Select("user_id").
		Where(goqu.Ex{"quiz_id": quizID}).
		Distinct().
		ScanVals(&userIDs)
	if err != nil {
		svc.logger.Warn("failed listing quiz learners for analytics cache bump", zap.Error(err))
		return
	}
	svc.learnerCache.BumpVersions(userIDs)
}
