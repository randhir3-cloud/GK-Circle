package models

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	goqu "github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const AssessmentAnalyticsEventsTable = "assessment_analytics_events"

const (
	uqAnalyticsClientEvent        = "uq_analytics_client_event"
	uqAnalyticsAttemptIdempotency = "uq_analytics_attempt_idempotency"
)

var (
	ErrAnalyticsEventDuplicate = errors.New("assessment analytics event duplicate")
	ErrAnalyticsInvalidCursor  = errors.New("assessment analytics invalid cursor")
)

type AssessmentAnalyticsEvent struct {
	ID             uuid.UUID       `db:"id"`
	ClientEventID  uuid.NullUUID   `db:"client_event_id"`
	AttemptID      uuid.UUID       `db:"attempt_id"`
	QuizID         uuid.UUID       `db:"quiz_id"`
	UserID         string          `db:"user_id"`
	QuizOwnerID    sql.NullString  `db:"quiz_owner_id"`
	AttemptRefID   uuid.NullUUID   `db:"attempt_ref_id"`
	QuizRefID      uuid.NullUUID   `db:"quiz_ref_id"`
	EventType      string          `db:"event_type"`
	EventSource    string          `db:"event_source"`
	CorrelationID  string          `db:"correlation_id"`
	IdempotencyKey sql.NullString  `db:"idempotency_key"`
	SchemaVersion  int16           `db:"schema_version"`
	Metadata       json.RawMessage `db:"metadata"`
	OccurredAt     time.Time       `db:"occurred_at"`
	CreatedAt      time.Time       `db:"created_at"`
}

type AnalyticsBatchResult struct {
	Received   int
	Inserted   int
	Duplicates int
	Rejected   int
}

type AnalyticsPaginationParams struct {
	Limit  int
	Cursor string
}

type AssessmentAnalyticsModel struct {
	db *goqu.Database
}

func NewAssessmentAnalyticsModel(db *goqu.Database) *AssessmentAnalyticsModel {
	return &AssessmentAnalyticsModel{db: db}
}

func analyticsEventRecord(event AssessmentAnalyticsEvent) goqu.Record {
	metadata := event.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}
	record := goqu.Record{
		"id":              event.ID,
		"attempt_id":      event.AttemptID,
		"quiz_id":         event.QuizID,
		"user_id":         event.UserID,
		"event_type":      event.EventType,
		"event_source":    event.EventSource,
		"correlation_id":  event.CorrelationID,
		"schema_version":  event.SchemaVersion,
		"metadata":        goqu.L("?::jsonb", string(metadata)),
		"occurred_at":     event.OccurredAt,
	}
	if event.ClientEventID.Valid {
		record["client_event_id"] = event.ClientEventID.UUID
	}
	if event.QuizOwnerID.Valid {
		record["quiz_owner_id"] = event.QuizOwnerID.String
	}
	if event.AttemptRefID.Valid {
		record["attempt_ref_id"] = event.AttemptRefID.UUID
	}
	if event.QuizRefID.Valid {
		record["quiz_ref_id"] = event.QuizRefID.UUID
	}
	if event.IdempotencyKey.Valid {
		record["idempotency_key"] = event.IdempotencyKey.String
	}
	return record
}

// CreateEventTx inserts one analytics event inside an existing transaction.
// Append-only: no UPDATE/DELETE helpers exist on this model.
func (model *AssessmentAnalyticsModel) CreateEventTx(tx *goqu.TxDatabase, event AssessmentAnalyticsEvent) (AssessmentAnalyticsEvent, bool, error) {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.SchemaVersion == 0 {
		event.SchemaVersion = 1
	}
	if len(event.Metadata) == 0 {
		event.Metadata = json.RawMessage(`{}`)
	}

	result, err := tx.Insert(AssessmentAnalyticsEventsTable).
		Rows(analyticsEventRecord(event)).
		OnConflict(goqu.DoNothing()).
		Executor().Exec()
	if err != nil {
		return AssessmentAnalyticsEvent{}, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return AssessmentAnalyticsEvent{}, false, err
	}
	if affected == 0 {
		return event, false, nil
	}
	return event, true, nil
}

// CreateBatchEvents validates callers have already filtered events; inserts with ON CONFLICT DO NOTHING.
func (model *AssessmentAnalyticsModel) CreateBatchEvents(events []AssessmentAnalyticsEvent) (AnalyticsBatchResult, error) {
	result := AnalyticsBatchResult{Received: len(events)}
	if len(events) == 0 {
		return result, nil
	}

	tx, err := model.db.Begin()
	if err != nil {
		return result, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = tx.Rollback()
		}
	}()

	for i := range events {
		if events[i].ID == uuid.Nil {
			events[i].ID = uuid.New()
		}
		if events[i].SchemaVersion == 0 {
			events[i].SchemaVersion = 1
		}
		if len(events[i].Metadata) == 0 {
			events[i].Metadata = json.RawMessage(`{}`)
		}

		execResult, insertErr := tx.Insert(AssessmentAnalyticsEventsTable).
			Rows(analyticsEventRecord(events[i])).
			OnConflict(goqu.DoNothing()).
			Executor().Exec()
		if insertErr != nil {
			return result, insertErr
		}
		affected, rowsErr := execResult.RowsAffected()
		if rowsErr != nil {
			return result, rowsErr
		}
		if affected == 0 {
			result.Duplicates++
		} else {
			result.Inserted++
		}
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}
	ok = true
	return result, nil
}

func EncodeAnalyticsCursor(createdAt time.Time, id uuid.UUID) string {
	raw := fmt.Sprintf("%s|%s", createdAt.UTC().Format(time.RFC3339Nano), id.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeAnalyticsCursor(cursor string) (time.Time, uuid.UUID, error) {
	if strings.TrimSpace(cursor) == "" {
		return time.Time{}, uuid.Nil, ErrAnalyticsInvalidCursor
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, ErrAnalyticsInvalidCursor
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, ErrAnalyticsInvalidCursor
	}
	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, ErrAnalyticsInvalidCursor
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, ErrAnalyticsInvalidCursor
	}
	return createdAt.UTC(), id, nil
}

func (model *AssessmentAnalyticsModel) listEvents(where goqu.Expression, params AnalyticsPaginationParams) ([]AssessmentAnalyticsEvent, string, error) {
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := model.db.From(AssessmentAnalyticsEventsTable).Where(where)
	if params.Cursor != "" {
		cursorTime, cursorID, err := DecodeAnalyticsCursor(params.Cursor)
		if err != nil {
			return nil, "", err
		}
		query = query.Where(
			goqu.L("(created_at, id) < (?, ?)", cursorTime, cursorID),
		)
	}

	var events []AssessmentAnalyticsEvent
	err := query.
		Order(goqu.I("created_at").Desc(), goqu.I("id").Desc()).
		Limit(uint(limit + 1)).
		ScanStructs(&events)
	if err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(events) > limit {
		last := events[limit-1]
		nextCursor = EncodeAnalyticsCursor(last.CreatedAt, last.ID)
		events = events[:limit]
	}
	return events, nextCursor, nil
}

func (model *AssessmentAnalyticsModel) ListEventsByAttemptID(attemptID uuid.UUID, params AnalyticsPaginationParams) ([]AssessmentAnalyticsEvent, string, error) {
	return model.listEvents(goqu.Ex{"attempt_id": attemptID}, params)
}

func (model *AssessmentAnalyticsModel) ListEventsByQuizID(quizID uuid.UUID, params AnalyticsPaginationParams) ([]AssessmentAnalyticsEvent, string, error) {
	return model.listEvents(goqu.Ex{"quiz_id": quizID}, params)
}

func IsAnalyticsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
		return false
	}
	return pqErr.Constraint == uqAnalyticsClientEvent ||
		pqErr.Constraint == uqAnalyticsAttemptIdempotency
}
