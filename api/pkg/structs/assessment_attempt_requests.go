package structs

// ReqCreateAssessmentAttempt binds an attempt to an immutable test snapshot.
// Owner identity must come from the authenticated session — never from this body.
type ReqCreateAssessmentAttempt struct {
	SnapshotID string `json:"snapshot_id"`
}

// ReqAutosaveAttemptAnswer updates a single question answer for an IN_PROGRESS attempt.
// Do not send user_id, score, or is_correct — those are rejected server-side.
type ReqAutosaveAttemptAnswer struct {
	SelectedOptions  *[]int `json:"selected_options"`
	Clear            bool   `json:"clear"`
	IsMarkedReview   bool   `json:"is_marked_review"`
	TimeTakenSeconds *int   `json:"time_taken_seconds,omitempty"`
	// Forbidden client fields — presence is rejected.
	UserID    *string  `json:"user_id,omitempty"`
	Score     *float64 `json:"score,omitempty"`
	IsCorrect *bool    `json:"is_correct,omitempty"`
}

// AttemptAnswerLearnerResponse is score-free and answer-key-safe.
type AttemptAnswerLearnerResponse struct {
	QuestionID       string  `json:"question_id"`
	SelectedOptions  []int   `json:"selected_options"`
	IsMarkedReview   bool    `json:"is_marked_review"`
	AnsweredAt       *string `json:"answered_at,omitempty"`
	TimeTakenSeconds *int    `json:"time_taken_seconds,omitempty"`
	UpdatedAt        string  `json:"updated_at"`
}

type AttemptResumeProgressResponse struct {
	TotalQuestions    int `json:"total_questions"`
	AnsweredCount     int `json:"answered_count"`
	MarkedReviewCount int `json:"marked_review_count"`
	UnansweredCount   int `json:"unanswered_count"`
}

// ReqSubmitAssessmentAttempt rejects client-controlled scoring fields.
type ReqSubmitAssessmentAttempt struct {
	UserID    *string  `json:"user_id,omitempty"`
	Score     *float64 `json:"score,omitempty"`
	IsCorrect *bool    `json:"is_correct,omitempty"`
	Status    *string  `json:"status,omitempty"`
	TotalScore *float64 `json:"total_score,omitempty"`
}

type AttemptResultSummaryResponse struct {
	CorrectCount    int     `json:"correct_count"`
	IncorrectCount  int     `json:"incorrect_count"`
	UnansweredCount int     `json:"unanswered_count"`
	UnscoredCount   int     `json:"unscored_count"`
	TotalScore      float64 `json:"total_score"`
	MaxScore        float64 `json:"max_score"`
}

type AttemptResultAnswerResponse struct {
	QuestionID       string   `json:"question_id"`
	SelectedOptions  []int    `json:"selected_options"`
	IsMarkedReview   bool     `json:"is_marked_review"`
	IsCorrect        *bool    `json:"is_correct,omitempty"`
	Score            *float64 `json:"score,omitempty"`
	AnsweredAt       *string  `json:"answered_at,omitempty"`
	TimeTakenSeconds *int     `json:"time_taken_seconds,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
}

// AssessmentAttemptResultResponse is learner-safe: correctness/marks without answer keys.
type AssessmentAttemptResultResponse struct {
	ID                       string                        `json:"id"`
	QuizID                   string                        `json:"quiz_id"`
	UserID                   string                        `json:"user_id"`
	TestSnapshotID           string                        `json:"test_snapshot_id"`
	AttemptNumber            int                           `json:"attempt_number"`
	Status                   string                        `json:"status"`
	QuestionOrder            []string                      `json:"question_order"`
	NegativeMarksPerQuestion float64                       `json:"negative_marks_per_question"`
	ExpectedMaxScore         *float64                      `json:"expected_max_score,omitempty"`
	StartedAt                string                        `json:"started_at"`
	SubmittedAt              *string                       `json:"submitted_at,omitempty"`
	ExpiresAt                *string                       `json:"expires_at,omitempty"`
	CreatedAt                string                        `json:"created_at"`
	UpdatedAt                string                        `json:"updated_at"`
	Snapshot                 TestSnapshotLearnerResponse   `json:"snapshot"`
	Answers                  []AttemptResultAnswerResponse `json:"answers"`
	Summary                  AttemptResultSummaryResponse  `json:"summary"`
}

type QuestionReviewOptionResponse struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	Media    string `json:"media,omitempty"`
	Selected bool   `json:"selected"`
	Correct  *bool  `json:"correct,omitempty"`
}

type QuestionReviewItemResponse struct {
	ID               string                         `json:"id"`
	Position         int                            `json:"position"`
	Question         string                         `json:"question"`
	Type             int                            `json:"type"`
	Options          []QuestionReviewOptionResponse `json:"options"`
	OptionsMedia     string                         `json:"options_media,omitempty"`
	QuestionMedia    string                         `json:"question_media,omitempty"`
	Resource         string                         `json:"resource,omitempty"`
	Points           int16                          `json:"points"`
	IsMarkedReview   bool                           `json:"is_marked_review"`
	IsCorrect        *bool                          `json:"is_correct,omitempty"`
	Score            *float64                       `json:"score,omitempty"`
	TimeTakenSeconds *int                           `json:"time_taken_seconds,omitempty"`
	Explanation      *string                        `json:"explanation,omitempty"`
}

type AssessmentResultSummaryResponse struct {
	TotalScore      *float64 `json:"total_score,omitempty"`
	MaxScore        *float64 `json:"max_score,omitempty"`
	Percentage      *float64 `json:"percentage,omitempty"`
	Passed          *bool    `json:"passed,omitempty"`
	DurationSeconds int      `json:"duration_seconds"`
	Answered        int      `json:"answered"`
	Correct         *int     `json:"correct,omitempty"`
	Incorrect       *int     `json:"incorrect,omitempty"`
	Unanswered      int      `json:"unanswered"`
	Unscored        int      `json:"unscored"`
}

type AssessmentResultReviewResponse struct {
	Questions []QuestionReviewItemResponse `json:"questions"`
}

type AssessmentAttemptResultDetailResponse struct {
	AttemptID           string                           `json:"attempt_id"`
	QuizID              string                           `json:"quiz_id"`
	Status              string                           `json:"status"`
	StartedAt           string                           `json:"started_at"`
	SubmittedAt         *string                          `json:"submitted_at,omitempty"`
	CanViewResult       bool                             `json:"can_view_result"`
	CanShowScore        bool                             `json:"can_show_score"`
	CanShowPassFail     bool                             `json:"can_show_pass_fail"`
	CanReviewQuestions  bool                             `json:"can_review_questions"`
	CanShowCorrectness  bool                             `json:"can_show_correctness"`
	CanShowExplanations bool                             `json:"can_show_explanations"`
	Message             string                           `json:"message,omitempty"`
	Summary             *AssessmentResultSummaryResponse `json:"summary,omitempty"`
	Review              *AssessmentResultReviewResponse  `json:"review,omitempty"`
}

// AssessmentAttemptResumeResponse is the learner resume payload (no answer keys).
type AssessmentAttemptResumeResponse struct {
	ID                       string                         `json:"id"`
	QuizID                   string                         `json:"quiz_id"`
	UserID                   string                         `json:"user_id"`
	TestSnapshotID           string                         `json:"test_snapshot_id"`
	AttemptNumber            int                            `json:"attempt_number"`
	Status                   string                         `json:"status"`
	QuestionOrder            []string                       `json:"question_order"`
	NegativeMarksPerQuestion float64                        `json:"negative_marks_per_question"`
	ExpectedMaxScore         *float64                       `json:"expected_max_score,omitempty"`
	StartedAt                string                         `json:"started_at"`
	SubmittedAt              *string                        `json:"submitted_at,omitempty"`
	ExpiresAt                *string                        `json:"expires_at,omitempty"`
	RemainingSeconds         *int64                         `json:"remaining_seconds,omitempty"`
	CreatedAt                string                         `json:"created_at"`
	UpdatedAt                string                         `json:"updated_at"`
	Snapshot                 TestSnapshotLearnerResponse    `json:"snapshot"`
	Answers                  []AttemptAnswerLearnerResponse `json:"answers"`
	Progress                 AttemptResumeProgressResponse  `json:"progress"`
}

// AssessmentAttemptStatusResponse is a lightweight status payload for periodic timer resync.
type AssessmentAttemptStatusResponse struct {
	Status           string  `json:"status"`
	ExpiresAt        *string `json:"expires_at,omitempty"`
	RemainingSeconds *int64  `json:"remaining_seconds,omitempty"`
}

// AssessmentAttemptResponse is the learner-safe attempt envelope (no answer keys).
type AssessmentAttemptResponse struct {
	ID                       string                      `json:"id"`
	QuizID                   string                      `json:"quiz_id"`
	UserID                   string                      `json:"user_id"`
	TestSnapshotID           string                      `json:"test_snapshot_id"`
	AttemptNumber            int                         `json:"attempt_number"`
	Status                   string                      `json:"status"`
	QuestionOrder            []string                    `json:"question_order"`
	NegativeMarksPerQuestion float64                     `json:"negative_marks_per_question"`
	ExpectedMaxScore         *float64                    `json:"expected_max_score,omitempty"`
	StartedAt                string                      `json:"started_at"`
	SubmittedAt              *string                     `json:"submitted_at,omitempty"`
	ExpiresAt                *string                     `json:"expires_at,omitempty"`
	CreatedAt                string                      `json:"created_at"`
	UpdatedAt                string                      `json:"updated_at"`
	Snapshot                 TestSnapshotLearnerResponse `json:"snapshot"`
}

// AttemptInstructionsQuizResponse is key-safe quiz metadata for the player start screen.
type AttemptInstructionsQuizResponse struct {
	ID                       string  `json:"id"`
	Title                    string  `json:"title"`
	Description              string  `json:"description,omitempty"`
	AssessmentMode           string  `json:"assessment_mode"`
	Status                   string  `json:"status"`
	DurationSeconds          *int64  `json:"duration_seconds,omitempty"`
	MaxAttempts              int     `json:"max_attempts"`
	NegativeMarksPerQuestion float64 `json:"negative_marks_per_question"`
}

// AttemptInstructionsSnapshotResponse summarises the immutable snapshot without items/keys.
type AttemptInstructionsSnapshotResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	QuestionCount int    `json:"question_count"`
	CreatedAt     string `json:"created_at"`
}

// AttemptInstructionsActiveAttemptResponse identifies an owned IN_PROGRESS attempt for Resume.
type AttemptInstructionsActiveAttemptResponse struct {
	ID            string `json:"id"`
	AttemptNumber int    `json:"attempt_number"`
	Status        string `json:"status"`
	StartedAt     string `json:"started_at"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
}

// AssessmentAttemptInstructionsResponse powers EXAM-P6-T01 instructions + start/resume UI.
type AssessmentAttemptInstructionsResponse struct {
	Quiz              AttemptInstructionsQuizResponse           `json:"quiz"`
	Snapshot          AttemptInstructionsSnapshotResponse       `json:"snapshot"`
	ActiveAttempt     *AttemptInstructionsActiveAttemptResponse `json:"active_attempt,omitempty"`
	AttemptsConsumed  int                                       `json:"attempts_consumed"`
	CanStart          bool                                      `json:"can_start"`
	CanResume         bool                                      `json:"can_resume"`
	BlockReason       string                                    `json:"block_reason,omitempty"`
}

// AssessmentAttemptEditorResponse includes frozen snapshot answer keys for authorised editors.
type AssessmentAttemptEditorResponse struct {
	ID             string               `json:"id"`
	QuizID         string               `json:"quiz_id"`
	UserID         string               `json:"user_id"`
	TestSnapshotID string               `json:"test_snapshot_id"`
	AttemptNumber  int                  `json:"attempt_number"`
	Status         string               `json:"status"`
	QuestionOrder  []string             `json:"question_order"`
	StartedAt      string               `json:"started_at"`
	SubmittedAt    *string              `json:"submitted_at,omitempty"`
	ExpiresAt      *string              `json:"expires_at,omitempty"`
	CreatedAt      string               `json:"created_at"`
	UpdatedAt      string               `json:"updated_at"`
	Snapshot       TestSnapshotResponse `json:"snapshot"`
}
