package structs

type InstructorOverviewResponse struct {
	ResolvedTimezone             string   `json:"resolved_timezone"`
	TotalQuizzes                 int      `json:"total_quizzes"`
	TotalAttempts                int      `json:"total_attempts"`
	CompletedAttempts            int      `json:"completed_attempts"`
	CompletedScoredAttemptsCount int      `json:"completed_scored_attempts_count"`
	CompletionRate               float64  `json:"completion_rate"`
	AverageScorePercentage       *float64 `json:"average_score_percentage"`
	UniqueLearners               int      `json:"unique_learners"`
	AverageDurationSeconds       *int64   `json:"average_duration_seconds"`
	ReleasedQuizzes              int      `json:"released_quizzes"`
	PendingReleaseQuizzes        int      `json:"pending_release_quizzes"`
}

type InstructorQuizListItem struct {
	QuizID                 string   `json:"quiz_id"`
	Title                  string   `json:"title"`
	CategoryName           *string  `json:"category_name"`
	TotalAttempts          int      `json:"total_attempts"`
	CompletedAttempts      int      `json:"completed_attempts"`
	AverageScorePercentage *float64 `json:"average_score_percentage"`
	UniqueLearners         int      `json:"unique_learners"`
	ResultReleasePolicy    string   `json:"result_release_policy"`
	ResultsReleased        bool     `json:"results_released"`
	CreatedAt              string   `json:"created_at"`
}

type InstructorQuizListResponse struct {
	Quizzes    []InstructorQuizListItem `json:"quizzes"`
	NextCursor string                   `json:"next_cursor,omitempty"`
	HasMore    bool                     `json:"has_more"`
}

type InstructorLearnerItem struct {
	LearnerID                  string   `json:"learner_id"`
	DisplayName                string   `json:"display_name"`
	AvatarURL                  *string  `json:"avatar_url"`
	UniqueQuizzesAttempted     int      `json:"unique_quizzes_attempted"`
	TotalAttempts              int      `json:"total_attempts"`
	CompletedAttempts          int      `json:"completed_attempts"`
	CompletionRate             float64  `json:"completion_rate"`
	AveragePercentage          *float64 `json:"average_percentage"`
	AssessmentDurationSeconds  int64    `json:"assessment_duration_seconds"`
	EngagedQuestionTimeSeconds int64    `json:"engaged_question_time_seconds"`
	LastActivityAt             string   `json:"last_activity_at"`
}

type InstructorLearnerListResponse struct {
	ResolvedTimezone string                  `json:"resolved_timezone"`
	Learners         []InstructorLearnerItem `json:"learners"`
	NextCursor       string                  `json:"next_cursor,omitempty"`
	HasMore          bool                    `json:"has_more"`
}

type InstructorReleaseItem struct {
	QuizID                 string  `json:"quiz_id"`
	Title                  string  `json:"title"`
	ResultReleasePolicy    string  `json:"result_release_policy"`
	ResultsReleased        bool    `json:"results_released"`
	ResultsScheduledAt     *string `json:"results_scheduled_at,omitempty"`
	ResultsReleasedAt      *string `json:"results_released_at,omitempty"`
	Category               string  `json:"category"`
	Classification         string  `json:"classification"`
	CompletedAttemptsCount int     `json:"completed_attempts_count"`
}

type InstructorReleaseMonitoringResponse struct {
	ResolvedTimezone string                  `json:"resolved_timezone"`
	Summary          map[string]int          `json:"summary"`
	Quizzes          []InstructorReleaseItem `json:"quizzes"`
	NextCursor       string                  `json:"next_cursor,omitempty"`
	HasMore          bool                    `json:"has_more"`
}

type InstructorTimelineEvent struct {
	ID          string `json:"id"`
	QuizID      string `json:"quiz_id"`
	QuizTitle   string `json:"quiz_title"`
	EventType   string `json:"event_type"`
	EventSource string `json:"event_source"`
	OccurredAt  string `json:"occurred_at"`
	Summary     string `json:"summary"`
}

type InstructorTimelineResponse struct {
	ResolvedTimezone string                    `json:"resolved_timezone"`
	Events           []InstructorTimelineEvent `json:"events"`
	NextCursor       string                    `json:"next_cursor,omitempty"`
	HasMore          bool                      `json:"has_more"`
}

type InstructorQuizSummaryResponse struct {
	QuizID                 string   `json:"quiz_id"`
	Title                  string   `json:"title"`
	TotalAttempts          int      `json:"total_attempts"`
	CompletedAttempts      int      `json:"completed_attempts"`
	InProgressAttempts     int      `json:"in_progress_attempts"`
	AbandonedAttempts      int      `json:"abandoned_attempts"`
	CompletionRate         float64  `json:"completion_rate"`
	AverageScorePercentage *float64 `json:"average_score_percentage"`
	HighestScorePercentage *float64 `json:"highest_score_percentage"`
	LowestScorePercentage  *float64 `json:"lowest_score_percentage"`
	AverageDurationSeconds *int64   `json:"average_duration_seconds"`
	UniqueLearners         int      `json:"unique_learners"`
	ResultReleasePolicy    string   `json:"result_release_policy"`
	ResultsReleased        bool     `json:"results_released"`
	TotalQuestions         int      `json:"total_questions"`
}

type InstructorAttemptListItem struct {
	AttemptID        string   `json:"attempt_id"`
	LearnerID        string   `json:"learner_id"`
	DisplayName      string   `json:"display_name"`
	AvatarURL        *string  `json:"avatar_url"`
	AttemptNumber    int      `json:"attempt_number"`
	Status           string   `json:"status"`
	TotalScore       *float64 `json:"total_score"`
	MaxScore         *float64 `json:"max_score"`
	Percentage       *float64 `json:"percentage"`
	TimeTakenSeconds *int64   `json:"time_taken_seconds"`
	StartedAt        string   `json:"started_at"`
	SubmittedAt      *string  `json:"submitted_at,omitempty"`
}

type InstructorAttemptListResponse struct {
	QuizID     string                      `json:"quiz_id"`
	Attempts   []InstructorAttemptListItem `json:"attempts"`
	NextCursor string                      `json:"next_cursor,omitempty"`
	HasMore    bool                        `json:"has_more"`
}

type InstructorQuestionMetricsItem struct {
	QuestionID          string         `json:"question_id"`
	QuestionText        string         `json:"question_text"`
	QuestionType        int            `json:"question_type"`
	OrderNumber         int            `json:"order_number"`
	TotalAnswered       int            `json:"total_answered"`
	CorrectCount        int            `json:"correct_count"`
	IncorrectCount      int            `json:"incorrect_count"`
	UnansweredCount     int            `json:"unanswered_count"`
	DifficultyIndex     *float64       `json:"difficulty_index"`
	DiscriminationIndex *float64       `json:"discrimination_index"`
	AnswerDistribution  map[string]int `json:"answer_distribution"`
	AverageTimeSeconds  *float64       `json:"average_time_seconds"`
}

type InstructorQuestionMetricsResponse struct {
	QuizID     string                          `json:"quiz_id"`
	Questions  []InstructorQuestionMetricsItem `json:"questions"`
	NextCursor string                          `json:"next_cursor,omitempty"`
	HasMore    bool                            `json:"has_more"`
}

type InstructorEngagementResponse struct {
	QuizID                  string  `json:"quiz_id"`
	TotalQuestionViews      int     `json:"total_question_views"`
	TotalAnswerSelections   int     `json:"total_answer_selections"`
	TotalAnswerChanges      int     `json:"total_answer_changes"`
	TotalHintsOpened        int     `json:"total_hints_opened"`
	TotalReviewsOpened      int     `json:"total_reviews_opened"`
	TotalResultViews        int     `json:"total_result_views"`
	UniqueEngagedLearners   int     `json:"unique_engaged_learners"`
	UniqueEngagedAttempts   int     `json:"unique_engaged_attempts"`
	AverageEventsPerAttempt float64 `json:"average_events_per_attempt"`
}
