package models

import (
	"database/sql"
	"time"

	goqu "github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

const QuizResultReleaseAuditLogsTable = "quiz_result_release_audit_logs"

type QuizResultReleaseAuditLog struct {
	ID             uuid.UUID      `db:"id"`
	QuizID         uuid.UUID      `db:"quiz_id"`
	ActorID        string         `db:"actor_id"`
	EventType      string         `db:"event_type"`
	PreviousPolicy sql.NullString `db:"previous_policy"`
	NewPolicy      sql.NullString `db:"new_policy"`
	PreviousState  sql.NullString `db:"previous_state"`
	NewState       sql.NullString `db:"new_state"`
	IPAddress      sql.NullString `db:"ip_address"`
	UserAgent      sql.NullString `db:"user_agent"`
	CorrelationID  sql.NullString `db:"correlation_id"`
	SchemaVersion  int16          `db:"schema_version"`
	CreatedAt      time.Time      `db:"created_at"`
}

type QuizResultReleaseAuditModel struct {
	db *goqu.Database
}

func NewQuizResultReleaseAuditModel(db *goqu.Database) *QuizResultReleaseAuditModel {
	return &QuizResultReleaseAuditModel{db: db}
}

func (model *QuizResultReleaseAuditModel) CreateAuditLog(log QuizResultReleaseAuditLog) (QuizResultReleaseAuditLog, error) {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	if log.SchemaVersion == 0 {
		log.SchemaVersion = 1
	}

	_, err := model.db.Insert(QuizResultReleaseAuditLogsTable).Rows(goqu.Record{
		"id":              log.ID,
		"quiz_id":         log.QuizID,
		"actor_id":        log.ActorID,
		"event_type":      log.EventType,
		"previous_policy": log.PreviousPolicy,
		"new_policy":      log.NewPolicy,
		"previous_state":  log.PreviousState,
		"new_state":       log.NewState,
		"ip_address":      log.IPAddress,
		"user_agent":      log.UserAgent,
		"correlation_id":  log.CorrelationID,
		"schema_version":  log.SchemaVersion,
		"created_at":      log.CreatedAt,
	}).Executor().Exec()

	return log, err
}

func (model *QuizResultReleaseAuditModel) ListAuditLogsByQuizID(quizID uuid.UUID) ([]QuizResultReleaseAuditLog, error) {
	var logs []QuizResultReleaseAuditLog
	err := model.db.From(QuizResultReleaseAuditLogsTable).
		Where(goqu.Ex{"quiz_id": quizID}).
		Order(goqu.I("created_at").Desc()).
		ScanStructs(&logs)
	return logs, err
}

func (model *QuizResultReleaseAuditModel) ListAuditLogsByActorID(actorID string) ([]QuizResultReleaseAuditLog, error) {
	var logs []QuizResultReleaseAuditLog
	err := model.db.From(QuizResultReleaseAuditLogsTable).
		Where(goqu.Ex{"actor_id": actorID}).
		Order(goqu.I("created_at").Desc()).
		ScanStructs(&logs)
	return logs, err
}
