package models

import (
	"database/sql"
	"time"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type QuestionRevision struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	QuestionID           uuid.UUID `json:"question_id" db:"question_id"`
	LineageID            uuid.UUID `json:"lineage_id" db:"lineage_id"`
	RevisionNumber       int       `json:"revision_number" db:"revision_number"`
	Question             string    `json:"question" db:"question"`
	Type                 int       `json:"type" db:"type"`
	Options              string    `json:"options" db:"options"`
	Answers              string    `json:"answers" db:"answers"`
	OfficialAnswer       string    `json:"official_answer" db:"official_answer"`
	AuthoritativeAnswer  string    `json:"authoritative_answer" db:"authoritative_answer"`
	AnswerReviewStatus   string    `json:"answer_review_status" db:"answer_review_status"`
	AnswerRevisionReason sql.NullString `json:"answer_revision_reason" db:"answer_revision_reason"`
	AnswerRevisionSource sql.NullString `json:"answer_revision_source" db:"answer_revision_source"`
	Points               int16     `json:"points" db:"points"`
	DurationInSeconds    int       `json:"duration_in_seconds" db:"duration_in_seconds"`
	QuestionMedia        string    `json:"question_media" db:"question_media"`
	OptionsMedia         string    `json:"options_media" db:"options_media"`
	Resource             sql.NullString `json:"resource" db:"resource"`
	CreatedBy            sql.NullString `json:"created_by" db:"created_by"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
}

type QuestionRevisionModel struct {
	db     *goqu.Database
	logger *zap.Logger
}

func InitQuestionRevisionModel(goquDB *goqu.Database, logger *zap.Logger) *QuestionRevisionModel {
	return &QuestionRevisionModel{db: goquDB, logger: logger}
}

func (model *QuestionRevisionModel) RecordRevision(
	transaction *goqu.TxDatabase,
	question Question,
	authority AnswerAuthorityFields,
	optionsJSON string,
	answersJSON string,
	officialJSON string,
	authoritativeJSON string,
	createdBy string,
) error {
	revisionID, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	record := goqu.Record{
		"id":                    revisionID,
		"question_id":           question.ID,
		"lineage_id":            authority.LineageID,
		"revision_number":       authority.RevisionNumber,
		"question":              question.Question,
		"type":                  question.Type,
		"options":               optionsJSON,
		"answers":               answersJSON,
		"official_answer":       officialJSON,
		"authoritative_answer":  authoritativeJSON,
		"answer_review_status":  authority.AnswerReviewStatus,
		"answer_revision_reason": authority.AnswerRevisionReason,
		"answer_revision_source": authority.AnswerRevisionSource,
		"points":                question.Points,
		"duration_in_seconds":   question.DurationInSeconds,
		"question_media":        question.QuestionMedia,
		"options_media":         question.OptionsMedia,
		"resource":              question.Resource.String,
		"created_by":            createdBy,
		"created_at":            goqu.L("now()"),
	}

	_, err = transaction.Insert(constants.QuestionRevisionsTable).Rows(record).Executor().Exec()
	return err
}

func (model *QuestionRevisionModel) ListByLineageID(lineageID string) ([]QuestionRevision, error) {
	revisions := []QuestionRevision{}
	err := model.db.From(constants.QuestionRevisionsTable).
		Select(
			"id",
			"question_id",
			"lineage_id",
			"revision_number",
			"question",
			"type",
			"options",
			"answers",
			"official_answer",
			"authoritative_answer",
			"answer_review_status",
			"answer_revision_reason",
			"answer_revision_source",
			"points",
			"duration_in_seconds",
			"question_media",
			"options_media",
			"resource",
			"created_by",
			"created_at",
		).
		Where(goqu.Ex{"lineage_id": lineageID}).
		Order(goqu.I("revision_number").Desc()).
		ScanStructs(&revisions)
	if err != nil {
		return nil, err
	}
	return revisions, nil
}
