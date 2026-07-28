package structs

import "time"

type AssessmentEventType string

const (
	EventAttemptStarted                 AssessmentEventType = "ATTEMPT_STARTED"
	EventAttemptSubmitted               AssessmentEventType = "ATTEMPT_SUBMITTED"
	EventAttemptAutoSubmitted           AssessmentEventType = "ATTEMPT_AUTO_SUBMITTED"
	EventResultReleaseOverrideApplied   AssessmentEventType = "RESULT_RELEASE_OVERRIDE_APPLIED"
	EventResultReleaseScheduledEffective AssessmentEventType = "RESULT_RELEASE_SCHEDULED_EFFECTIVE"
	EventAttemptAutosaved               AssessmentEventType = "ATTEMPT_AUTOSAVED"
	EventResultViewed                   AssessmentEventType = "RESULT_VIEWED"
	EventQuestionViewed                 AssessmentEventType = "QUESTION_VIEWED"
	EventAnswerSelected                 AssessmentEventType = "ANSWER_SELECTED"
	EventAnswerChanged                  AssessmentEventType = "ANSWER_CHANGED"
	EventQuestionTimeSpent              AssessmentEventType = "QUESTION_TIME_SPENT"
	EventHintOpened                     AssessmentEventType = "HINT_OPENED"
	EventReviewOpened                   AssessmentEventType = "REVIEW_OPENED"
)

type AssessmentEventSource string

const (
	EventSourceHTTP        AssessmentEventSource = "HTTP"
	EventSourceWorker      AssessmentEventSource = "WORKER"
	EventSourceScheduler   AssessmentEventSource = "SCHEDULER"
	EventSourceClientBatch AssessmentEventSource = "CLIENT_BATCH"
)

func IsAuthoritativeEventType(eventType AssessmentEventType) bool {
	switch eventType {
	case EventAttemptStarted,
		EventAttemptSubmitted,
		EventAttemptAutoSubmitted,
		EventResultReleaseOverrideApplied,
		EventResultReleaseScheduledEffective:
		return true
	default:
		return false
	}
}

func IsServerTelemetryEventType(eventType AssessmentEventType) bool {
	switch eventType {
	case EventAttemptAutosaved, EventResultViewed:
		return true
	default:
		return false
	}
}

func IsClientTelemetryEventType(eventType AssessmentEventType) bool {
	switch eventType {
	case EventQuestionViewed,
		EventAnswerSelected,
		EventAnswerChanged,
		EventQuestionTimeSpent,
		EventHintOpened,
		EventReviewOpened:
		return true
	default:
		return false
	}
}

type RecordTelemetryEventRequest struct {
	ClientEventID  *string                `json:"client_event_id"`
	EventType      AssessmentEventType    `json:"event_type"`
	IdempotencyKey *string                `json:"idempotency_key"`
	Metadata       map[string]interface{} `json:"metadata"`
	OccurredAt     time.Time              `json:"occurred_at"`
}

type RecordTelemetryBatchRequest struct {
	Events []RecordTelemetryEventRequest `json:"events"`
}

type BatchResultResponse struct {
	Received   int `json:"received"`
	Inserted   int `json:"inserted"`
	Duplicates int `json:"duplicates"`
	Rejected   int `json:"rejected"`
}

type AnalyticsPaginationParams struct {
	Limit  int    `json:"limit"`
	Cursor string `json:"cursor"`
}

type AssessmentAnalyticsEventResponse struct {
	ID             string                 `json:"id"`
	ClientEventID  *string                `json:"client_event_id,omitempty"`
	AttemptID      string                 `json:"attempt_id"`
	QuizID         string                 `json:"quiz_id"`
	UserID         string                 `json:"user_id"`
	QuizOwnerID    *string                `json:"quiz_owner_id,omitempty"`
	EventType      string                 `json:"event_type"`
	EventSource    string                 `json:"event_source"`
	CorrelationID  string                 `json:"correlation_id"`
	IdempotencyKey *string                `json:"idempotency_key,omitempty"`
	SchemaVersion  int16                  `json:"schema_version"`
	Metadata       map[string]interface{} `json:"metadata"`
	OccurredAt     string                 `json:"occurred_at"`
	CreatedAt      string                 `json:"created_at"`
}

type AssessmentAnalyticsEventListResponse struct {
	Events     []AssessmentAnalyticsEventResponse `json:"events"`
	NextCursor string                             `json:"next_cursor,omitempty"`
}
