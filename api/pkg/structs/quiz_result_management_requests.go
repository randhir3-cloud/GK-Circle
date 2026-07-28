package structs

type ResultReleasePolicy string

const (
	ResultReleasePolicyImmediate ResultReleasePolicy = "IMMEDIATE"
	ResultReleasePolicyManual    ResultReleasePolicy = "MANUAL"
	ResultReleasePolicyScheduled ResultReleasePolicy = "SCHEDULED"
)

const (
	AuditPolicyUpdate  = "POLICY_UPDATE"
	AuditManualRelease = "MANUAL_RELEASE"
)

type UpdateQuizResultSettingsRequest struct {
	ResultReleasePolicy ResultReleasePolicy `json:"result_release_policy"`
	ResultsScheduledAt  *string             `json:"results_scheduled_at"`
	ShowScore           *bool               `json:"show_score"`
	ShowPassFail        *bool               `json:"show_pass_fail"`
	AllowAnswerReview   *bool               `json:"allow_answer_review"`
	ShowCorrectness     *bool               `json:"show_correctness"`
	ShowExplanations    *bool               `json:"show_explanations"`
}

type QuizResultReleaseStatusResponse struct {
	QuizID                 string  `json:"quiz_id"`
	ResultReleasePolicy    string  `json:"result_release_policy"`
	ResultsReleased        bool    `json:"results_released"`
	ResultsScheduledAt     *string `json:"results_scheduled_at"`
	ResultsReleasedAt      *string `json:"results_released_at"`
	IsCurrentlyReleased    bool    `json:"is_currently_released"`
	ShowScore              bool    `json:"show_score"`
	ShowPassFail           bool    `json:"show_pass_fail"`
	AllowAnswerReview      bool    `json:"allow_answer_review"`
	ShowCorrectness        bool    `json:"show_correctness"`
	ShowExplanations       bool    `json:"show_explanations"`
	TotalSubmittedAttempts int     `json:"total_submitted_attempts"`
}
