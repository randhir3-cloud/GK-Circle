package models

import (
	"encoding/json"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/google/uuid"
)

// AnswerAuthorityFields holds ADR-024 answer authority on the live question row.
type AnswerAuthorityFields struct {
	LineageID            uuid.UUID `json:"lineage_id" db:"lineage_id"`
	RevisionNumber       int       `json:"revision_number" db:"revision_number"`
	OfficialAnswer       []int     `json:"official_answer"`
	AuthoritativeAnswer  []int     `json:"authoritative_answer"`
	AnswerReviewStatus   string    `json:"answer_review_status" db:"answer_review_status"`
	AnswerRevisionReason string    `json:"answer_revision_reason" db:"answer_revision_reason"`
	AnswerRevisionSource string    `json:"answer_revision_source" db:"answer_revision_source"`
}

type QuestionLineageMeta struct {
	LineageID      uuid.UUID `db:"lineage_id"`
	RevisionNumber int       `db:"revision_number"`
}

func validAnswerReviewStatus(status string) bool {
	switch status {
	case constants.AnswerReviewUnreviewed,
		constants.AnswerReviewConfirmed,
		constants.AnswerReviewDisputed,
		constants.AnswerReviewRevised:
		return true
	default:
		return false
	}
}

func normalizeAnswerReviewStatus(status string) string {
	if status == "" {
		return constants.AnswerReviewUnreviewed
	}
	if validAnswerReviewStatus(status) {
		return status
	}
	return constants.AnswerReviewUnreviewed
}

func resolveAnswerKeys(primary []int, fallback []int) []int {
	if len(primary) > 0 {
		return primary
	}
	return fallback
}

func marshalAnswerKeys(keys []int) (string, error) {
	payload, err := json.Marshal(keys)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func parseAnswerKeys(raw []byte) ([]int, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	keys := []int{}
	if err := json.Unmarshal(raw, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

// ApplyAnswerAuthority merges request authority onto operational answers and validates status.
func ApplyAnswerAuthority(
	answers []int,
	officialAnswer []int,
	authoritativeAnswer []int,
	reviewStatus string,
	revisionReason string,
	revisionSource string,
) (AnswerAuthorityFields, []int, error) {
	status := normalizeAnswerReviewStatus(reviewStatus)
	if reviewStatus != "" && !validAnswerReviewStatus(reviewStatus) {
		return AnswerAuthorityFields{}, nil, errInvalidAnswerReviewStatus
	}

	authoritative := resolveAnswerKeys(authoritativeAnswer, answers)
	official := resolveAnswerKeys(officialAnswer, authoritative)
	operationalAnswers := authoritative

	return AnswerAuthorityFields{
		OfficialAnswer:       official,
		AuthoritativeAnswer:  authoritative,
		AnswerReviewStatus:   status,
		AnswerRevisionReason: revisionReason,
		AnswerRevisionSource: revisionSource,
	}, operationalAnswers, nil
}

var errInvalidAnswerReviewStatus = &answerAuthorityError{msg: constants.ErrAnswerReviewStatusInvalid}

type answerAuthorityError struct {
	msg string
}

func (e *answerAuthorityError) Error() string {
	return e.msg
}
