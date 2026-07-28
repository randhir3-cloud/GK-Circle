package utils

import (
	"database/sql"
	"time"

	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

// BuildLiveQuestionDeliveryPayload returns the WebSocket question-delivery payload
// for an in-progress live question. It intentionally excludes answer keys.
func BuildLiveQuestionDeliveryPayload(
	question models.Question,
	questionStartTime time.Time,
	totalQuestions int64,
	totalJoinUsers int64,
) map[string]any {
	return map[string]any{
		"id":             question.ID,
		"quiz_id":        question.QuizId,
		"no":             question.OrderNumber,
		"duration":       question.DurationInSeconds,
		"start_time":     questionStartTime.Format(time.RFC3339),
		"server_time":    time.Now().UTC().Format(time.RFC3339Nano),
		"question":       question.Question,
		"options":        question.Options,
		"question_media": question.QuestionMedia,
		"options_media":  question.OptionsMedia,
		"resource":       question.Resource.String,
		"totalQuestions": totalQuestions,
		"totalJoinUser":  totalJoinUsers,
	}
}

// BuildLiveScoreboardPayload returns the post-question scoreboard payload. Answer
// keys are included because the live quiz reveals them after the question closes.
func BuildLiveScoreboardPayload(
	question models.Question,
	userRankBoard any,
	userResponses any,
	totalQuestions int64,
	duration int,
	includeUserResponses bool,
) map[string]any {
	payload := map[string]any{
		"question_no":    question.OrderNumber,
		"quiz_id":        question.QuizId,
		"rankList":       userRankBoard,
		"question":       question.Question,
		"answers":        question.Answers,
		"options":        question.Options,
		"question_media": question.QuestionMedia,
		"options_media":  question.OptionsMedia,
		"resource":       question.Resource.String,
		"duration":       duration,
		"totalQuestions": totalQuestions,
	}
	if includeUserResponses {
		payload["userResponses"] = userResponses
	}
	return payload
}

// EmptySQLNullString helps tests build question fixtures.
func EmptySQLNullString() sql.NullString {
	return sql.NullString{}
}
