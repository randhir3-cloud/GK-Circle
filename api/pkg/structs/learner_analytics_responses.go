package structs

type LearnerDashboardSummaryResponse struct {
	ResolvedTimezone             string   `json:"resolved_timezone"`
	TotalAttempts                int      `json:"total_attempts"`
	CompletedAttempts            int      `json:"completed_attempts"`
	CompletionRate               float64  `json:"completion_rate"`
	ScoredAttemptCount           int      `json:"scored_attempt_count"`
	PendingResultCount           int      `json:"pending_result_count"`
	AveragePercentage            *float64 `json:"average_percentage"`
	AssessmentDurationSeconds    int64    `json:"assessment_duration_seconds"`
	EngagedQuestionTimeSeconds   int64    `json:"engaged_question_time_seconds"`
	EngagedQuestionTimeApproximate bool   `json:"engaged_question_time_approximate"`
	CurrentStreakDays            int      `json:"current_streak_days"`
	BestStreakDays               int      `json:"best_streak_days"`
}

type LearnerActivityItem struct {
	AttemptID       string   `json:"attempt_id"`
	QuizID          string   `json:"quiz_id"`
	QuizTitle       string   `json:"quiz_title"`
	Status          string   `json:"status"`
	ResultStatus    string   `json:"result_status"`
	Percentage      *float64 `json:"percentage"`
	TotalScore      *float64 `json:"total_score"`
	MaxScore        *float64 `json:"max_score"`
	ActivityAt      string   `json:"activity_at"`
	CreatedAt       string   `json:"created_at"`
	SubmittedAt     *string  `json:"submitted_at,omitempty"`
	DurationSeconds *int64   `json:"duration_seconds,omitempty"`
}

type LearnerRecentActivityResponse struct {
	ResolvedTimezone string                `json:"resolved_timezone"`
	Items            []LearnerActivityItem `json:"items"`
	NextCursor       string                `json:"next_cursor,omitempty"`
	HasMore          bool                  `json:"has_more"`
}

type LearnerTrendBucket struct {
	Label                     string   `json:"label"`
	AttemptCount              int      `json:"attempt_count"`
	ScoredAttemptCount        int      `json:"scored_attempt_count"`
	AveragePercentage         *float64 `json:"average_percentage"`
	AssessmentDurationSeconds int64    `json:"assessment_duration_seconds"`
}

type LearnerPerformanceTrendsResponse struct {
	ResolvedTimezone string               `json:"resolved_timezone"`
	Granularity      string               `json:"granularity"`
	From             string               `json:"from"`
	To               string               `json:"to"`
	Buckets          []LearnerTrendBucket `json:"buckets"`
}

type LearnerSubjectPerformanceItem struct {
	SubjectID                 *string  `json:"subject_id"`
	SubjectName               string   `json:"subject_name"`
	AttemptCount              int      `json:"attempt_count"`
	ScoredAttemptCount        int      `json:"scored_attempt_count"`
	AveragePercentage         *float64 `json:"average_percentage"`
	AssessmentDurationSeconds int64    `json:"assessment_duration_seconds"`
}

type LearnerSubjectPerformanceResponse struct {
	ResolvedTimezone string                          `json:"resolved_timezone"`
	Subjects         []LearnerSubjectPerformanceItem `json:"subjects"`
}

type LearnerTimelineEvent struct {
	ID            string                 `json:"id"`
	EventType     string                 `json:"event_type"`
	EventSource   string                 `json:"event_source"`
	OccurredAt    string                 `json:"occurred_at"`
	CorrelationID string                 `json:"correlation_id"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type LearnerTimelineResponse struct {
	ResolvedTimezone string                 `json:"resolved_timezone"`
	AttemptID        string                 `json:"attempt_id"`
	QuizID           string                 `json:"quiz_id"`
	QuizTitle        string                 `json:"quiz_title"`
	Status           string                 `json:"status"`
	ResultStatus     string                 `json:"result_status"`
	Percentage       *float64               `json:"percentage"`
	Events           []LearnerTimelineEvent `json:"events"`
}
