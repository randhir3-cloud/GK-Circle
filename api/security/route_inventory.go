package security

// ReviewRoute describes an HTTP route that may return answer-key material after
// authorisation. Inventory is used by regression tests to prevent drift.
type ReviewRoute struct {
	Name                   string
	Method                 string
	Path                   string
	RequiresAuth           bool
	RequiresReviewAccess   bool
	AllowsAnswerKeysWhenOK bool
}

// ProtectedReviewRoutes lists user-facing review endpoints secured in EXAM-P2-T03.
func ProtectedReviewRoutes() []ReviewRoute {
	return []ReviewRoute{
		{
			Name:                   "analytics_board_user",
			Method:                 "GET",
			Path:                   "/api/v1/analytics_board/user",
			RequiresAuth:           true,
			RequiresReviewAccess:   true,
			AllowsAnswerKeysWhenOK: true,
		},
		{
			Name:                   "final_score_user",
			Method:                 "GET",
			Path:                   "/api/v1/final_score/user",
			RequiresAuth:           true,
			RequiresReviewAccess:   true,
			AllowsAnswerKeysWhenOK: true,
		},
		{
			Name:                   "user_played_quiz_review",
			Method:                 "GET",
			Path:                   "/api/v1/user_played_quizes/:user_played_quiz_id",
			RequiresAuth:           true,
			RequiresReviewAccess:   true,
			AllowsAnswerKeysWhenOK: true,
		},
	}
}

// PreReleasePayloadRoute describes payloads that must never include answer keys
// before a question closes or review is authorised.
type PreReleasePayloadRoute struct {
	Name             string
	Transport        string
	AllowsAnswerKeys bool
}

// PreReleasePayloadRoutes lists delivery surfaces audited in EXAM-P2-T04.
func PreReleasePayloadRoutes() []PreReleasePayloadRoute {
	return []PreReleasePayloadRoute{
		{
			Name:             "live_question_delivery_websocket",
			Transport:        "websocket",
			AllowsAnswerKeys: false,
		},
		{
			Name:             "public_quiz_catalog",
			Transport:        "http",
			AllowsAnswerKeys: false,
		},
		{
			Name:             "quiz_questions_list_unauthenticated",
			Transport:        "http",
			AllowsAnswerKeys: false,
		},
	}
}

// EditorAnswerKeyRoutes lists endpoints that intentionally expose answer keys to
// authorised editors or session hosts after authentication.
func EditorAnswerKeyRoutes() []ReviewRoute {
	return []ReviewRoute{
		{
			Name:                   "analytics_board_admin",
			Method:                 "GET",
			Path:                   "/api/v1/analytics_board/admin",
			RequiresAuth:           true,
			RequiresReviewAccess:   false,
			AllowsAnswerKeysWhenOK: true,
		},
		{
			Name:                   "final_score_admin",
			Method:                 "GET",
			Path:                   "/api/v1/final_score/admin",
			RequiresAuth:           true,
			RequiresReviewAccess:   false,
			AllowsAnswerKeysWhenOK: true,
		},
		{
			Name:                   "quiz_questions_editor_list",
			Method:                 "GET",
			Path:                   "/api/v1/quizzes/:quiz_id/questions",
			RequiresAuth:           true,
			RequiresReviewAccess:   false,
			AllowsAnswerKeysWhenOK: true,
		},
	}
}
