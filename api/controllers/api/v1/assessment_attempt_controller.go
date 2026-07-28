package v1

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

type AssessmentAttemptController struct {
	attemptSvc *services.AssessmentAttemptService
	appConfig  *config.AppConfig
	logger     *zap.Logger
}

func InitAssessmentAttemptController(db *goqu.Database, logger *zap.Logger, appConfig *config.AppConfig) (*AssessmentAttemptController, error) {
	return &AssessmentAttemptController{
		attemptSvc: services.NewAssessmentAttemptService(db, logger),
		appConfig:  appConfig,
		logger:     logger,
	}, nil
}

func (ctrl *AssessmentAttemptController) SetLearnerAnalyticsCache(cache *services.LearnerAnalyticsCache) {
	if ctrl.attemptSvc != nil {
		ctrl.attemptSvc.SetLearnerAnalyticsCache(cache)
	}
}

func (ctrl *AssessmentAttemptController) GetInstructions(c *fiber.Ctx) error {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuizInvalidID)
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	snapshotID, err := uuid.Parse(strings.TrimSpace(c.Query("snapshot_id")))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrTestSnapshotInvalidID)
	}

	isEditor, err := ctrl.attemptSvc.ResolveEditorPreview(quizID, user)
	if err != nil {
		ctrl.logger.Error("error resolving attempt instructions entitlement", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetAttemptInstructions)
	}

	view, err := ctrl.attemptSvc.GetInstructions(quizID, snapshotID, user.ID, isEditor)
	if err != nil {
		return mapAssessmentAttemptError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAssessmentAttemptInstructionsResponse(view))
}

func (ctrl *AssessmentAttemptController) CreateAttempt(c *fiber.Ctx) error {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuizInvalidID)
	}

	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	var req structs.ReqCreateAssessmentAttempt
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	snapshotID, err := uuid.Parse(strings.TrimSpace(req.SnapshotID))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrTestSnapshotInvalidID)
	}

	isEditor, err := ctrl.attemptSvc.ResolveEditorPreview(quizID, user)
	if err != nil {
		ctrl.logger.Error("error resolving attempt editor entitlement", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrCreateAssessmentAttempt)
	}

	view, created, err := ctrl.attemptSvc.Create(services.CreateAssessmentAttemptRequest{
		QuizID:          quizID,
		UserID:          user.ID,
		UserEmail:       user.Email,
		SnapshotID:      snapshotID,
		IsEditorPreview: isEditor,
		CorrelationID:   utils.ResolveAuditCorrelationID(c),
	})
	if err != nil {
		return mapAssessmentAttemptError(c, ctrl.logger, err)
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	return utils.JSONSuccess(c, status, toAssessmentAttemptLearnerResponse(view))
}

func (ctrl *AssessmentAttemptController) ListMyAttempts(c *fiber.Ctx) error {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuizInvalidID)
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	views, err := ctrl.attemptSvc.ListMine(quizID, user.ID)
	if err != nil {
		ctrl.logger.Error("error listing assessment attempts", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrListAssessmentAttempts)
	}
	response := make([]structs.AssessmentAttemptResponse, 0, len(views))
	for _, view := range views {
		response = append(response, toAssessmentAttemptLearnerResponse(view))
	}
	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *AssessmentAttemptController) GetMyAttempt(c *fiber.Ctx) error {
	quizID, attemptID, err := parseQuizAndAttemptIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	view, err := ctrl.attemptSvc.GetLearner(quizID, attemptID, user.ID)
	if err != nil {
		return mapAssessmentAttemptError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAssessmentAttemptLearnerResponse(view))
}

func (ctrl *AssessmentAttemptController) GetEditorAttempt(c *fiber.Ctx) error {
	quizID, attemptID, err := parseQuizAndAttemptIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	view, err := ctrl.attemptSvc.GetEditor(quizID, attemptID)
	if err != nil {
		return mapAssessmentAttemptError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAssessmentAttemptEditorResponse(view))
}

func (ctrl *AssessmentAttemptController) ResumeAttempt(c *fiber.Ctx) error {
	quizID, attemptID, err := parseQuizAndAttemptIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	view, err := ctrl.attemptSvc.Resume(quizID, attemptID, user.ID)
	if err != nil {
		return mapAssessmentAttemptError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAssessmentAttemptResumeResponse(view))
}

func (ctrl *AssessmentAttemptController) GetAttemptStatus(c *fiber.Ctx) error {
	quizID, attemptID, err := parseQuizAndAttemptIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	view, err := ctrl.attemptSvc.GetStatus(quizID, attemptID, user.ID)
	if err != nil {
		return mapAssessmentAttemptError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, structs.AssessmentAttemptStatusResponse{
		Status:           view.Status,
		ExpiresAt:        nullTimePtr(view.ExpiresAt),
		RemainingSeconds: view.RemainingSeconds,
	})
}

func (ctrl *AssessmentAttemptController) AutosaveAnswer(c *fiber.Ctx) error {
	quizID, attemptID, err := parseQuizAndAttemptIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	questionID, err := uuid.Parse(strings.TrimSpace(c.Params(constants.QuestionId)))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuestionInvalidID)
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	var req structs.ReqAutosaveAttemptAnswer
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	if req.UserID != nil || req.Score != nil || req.IsCorrect != nil {
		return utils.JSONFail(c, http.StatusBadRequest, models.ErrAttemptAnswerClientScoreForbidden.Error())
	}

	selected := []int{}
	clear := req.Clear
	if req.SelectedOptions != nil {
		selected = *req.SelectedOptions
		if len(selected) == 0 {
			clear = true
		}
	} else if !req.Clear {
		// Allow mark-for-review-only update without touching options when clear=false
		// and selected_options omitted: load existing via upsert with clear=false and empty
		// would clear — so require explicit selected_options or clear.
		return utils.JSONFail(c, http.StatusBadRequest, "selected_options is required (use [] or clear=true to clear)")
	}

	view, err := ctrl.attemptSvc.AutosaveAnswer(services.AutosaveAnswerRequest{
		QuizID:           quizID,
		AttemptID:        attemptID,
		QuestionID:       questionID,
		UserID:           user.ID,
		SelectedOptions:  selected,
		ClearAnswer:      clear,
		IsMarkedReview:   req.IsMarkedReview,
		TimeTakenSeconds: req.TimeTakenSeconds,
		CorrelationID:    utils.ResolveAuditCorrelationID(c),
	})
	if err != nil {
		return mapAssessmentAttemptError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAttemptAnswerLearnerResponse(view))
}

func (ctrl *AssessmentAttemptController) SubmitAttempt(c *fiber.Ctx) error {
	quizID, attemptID, err := parseQuizAndAttemptIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	if len(c.Body()) > 0 {
		var req structs.ReqSubmitAssessmentAttempt
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		if req.UserID != nil || req.Score != nil || req.IsCorrect != nil || req.Status != nil || req.TotalScore != nil {
			return utils.JSONFail(c, http.StatusBadRequest, models.ErrAttemptAnswerClientScoreForbidden.Error())
		}
	}

	view, created, err := ctrl.attemptSvc.Submit(quizID, attemptID, user.ID, utils.ResolveAuditCorrelationID(c))
	if err != nil {
		return mapAssessmentAttemptError(c, ctrl.logger, err)
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	return utils.JSONSuccess(c, status, toAssessmentAttemptResultResponse(view))
}

func (ctrl *AssessmentAttemptController) GetAttemptResult(c *fiber.Ctx) error {
	quizID, attemptID, err := parseQuizAndAttemptIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	view, err := ctrl.attemptSvc.GetResult(quizID, attemptID, user.ID, utils.ResolveAuditCorrelationID(c))
	if err != nil {
		return mapAssessmentAttemptError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAssessmentAttemptResultDetailResponse(view))
}

func parseQuizAndAttemptIDs(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, errors.New(constants.ErrQuizInvalidID)
	}
	attemptID, err := uuid.Parse(strings.TrimSpace(c.Params(constants.AttemptId)))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, errors.New(constants.ErrAssessmentAttemptInvalidID)
	}
	return quizID, attemptID, nil
}

func toAssessmentAttemptLearnerResponse(view services.AssessmentAttemptLearnerView) structs.AssessmentAttemptResponse {
	return structs.AssessmentAttemptResponse{
		ID:                       view.Attempt.ID.String(),
		QuizID:                   view.Attempt.QuizID.String(),
		UserID:                   view.Attempt.UserID,
		TestSnapshotID:           view.Attempt.TestSnapshotID.String(),
		AttemptNumber:            view.Attempt.AttemptNumber,
		Status:                   view.Attempt.Status,
		QuestionOrder:            uuidSliceToStrings(view.Attempt.QuestionOrder),
		NegativeMarksPerQuestion: view.Attempt.NegativeMarksPerQuestion,
		ExpectedMaxScore:         nullFloatPtr(view.Attempt.ExpectedMaxScore),
		StartedAt:                view.Attempt.StartedAt.UTC().Format(time.RFC3339),
		SubmittedAt:              nullTimePtr(view.Attempt.SubmittedAt),
		ExpiresAt:                nullTimePtr(view.Attempt.ExpiresAt),
		CreatedAt:                view.Attempt.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                view.Attempt.UpdatedAt.UTC().Format(time.RFC3339),
		Snapshot:                 toTestSnapshotLearnerResponse(view.Snapshot),
	}
}

func toAssessmentAttemptEditorResponse(view services.AssessmentAttemptEditorView) structs.AssessmentAttemptEditorResponse {
	return structs.AssessmentAttemptEditorResponse{
		ID:             view.Attempt.ID.String(),
		QuizID:         view.Attempt.QuizID.String(),
		UserID:         view.Attempt.UserID,
		TestSnapshotID: view.Attempt.TestSnapshotID.String(),
		AttemptNumber:  view.Attempt.AttemptNumber,
		Status:         view.Attempt.Status,
		QuestionOrder:  uuidSliceToStrings(view.Attempt.QuestionOrder),
		StartedAt:      view.Attempt.StartedAt.UTC().Format(time.RFC3339),
		SubmittedAt:    nullTimePtr(view.Attempt.SubmittedAt),
		ExpiresAt:      nullTimePtr(view.Attempt.ExpiresAt),
		CreatedAt:      view.Attempt.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      view.Attempt.UpdatedAt.UTC().Format(time.RFC3339),
		Snapshot:       toTestSnapshotResponse(view.Snapshot),
	}
}

func toAssessmentAttemptResumeResponse(view services.AssessmentAttemptResumeView) structs.AssessmentAttemptResumeResponse {
	answers := make([]structs.AttemptAnswerLearnerResponse, 0, len(view.Answers))
	for _, answer := range view.Answers {
		answers = append(answers, toAttemptAnswerLearnerResponse(answer))
	}
	return structs.AssessmentAttemptResumeResponse{
		ID:                       view.Attempt.ID.String(),
		QuizID:                   view.Attempt.QuizID.String(),
		UserID:                   view.Attempt.UserID,
		TestSnapshotID:           view.Attempt.TestSnapshotID.String(),
		AttemptNumber:            view.Attempt.AttemptNumber,
		Status:                   view.Attempt.Status,
		QuestionOrder:            uuidSliceToStrings(view.Attempt.QuestionOrder),
		NegativeMarksPerQuestion: view.Attempt.NegativeMarksPerQuestion,
		ExpectedMaxScore:         nullFloatPtr(view.Attempt.ExpectedMaxScore),
		StartedAt:                view.Attempt.StartedAt.UTC().Format(time.RFC3339),
		SubmittedAt:              nullTimePtr(view.Attempt.SubmittedAt),
		ExpiresAt:                nullTimePtr(view.Attempt.ExpiresAt),
		RemainingSeconds:         view.RemainingSeconds,
		CreatedAt:                view.Attempt.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                view.Attempt.UpdatedAt.UTC().Format(time.RFC3339),
		Snapshot:                 toTestSnapshotLearnerResponse(view.Snapshot),
		Answers:                  answers,
		Progress: structs.AttemptResumeProgressResponse{
			TotalQuestions:    view.Progress.TotalQuestions,
			AnsweredCount:     view.Progress.AnsweredCount,
			MarkedReviewCount: view.Progress.MarkedReviewCount,
			UnansweredCount:   view.Progress.UnansweredCount,
		},
	}
}

func toAttemptAnswerLearnerResponse(view services.AttemptAnswerLearnerView) structs.AttemptAnswerLearnerResponse {
	selected := view.SelectedOptions
	if selected == nil {
		selected = []int{}
	}
	resp := structs.AttemptAnswerLearnerResponse{
		QuestionID:       view.QuestionID.String(),
		SelectedOptions:  selected,
		IsMarkedReview:   view.IsMarkedReview,
		TimeTakenSeconds: view.TimeTakenSeconds,
		UpdatedAt:        view.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if view.AnsweredAt != nil {
		formatted := view.AnsweredAt.UTC().Format(time.RFC3339)
		resp.AnsweredAt = &formatted
	}
	return resp
}

func toAssessmentAttemptResultResponse(view services.AssessmentAttemptResultView) structs.AssessmentAttemptResultResponse {
	answers := make([]structs.AttemptResultAnswerResponse, 0, len(view.Answers))
	for _, answer := range view.Answers {
		selected := answer.SelectedOptions
		if selected == nil {
			selected = []int{}
		}
		entry := structs.AttemptResultAnswerResponse{
			QuestionID:       answer.QuestionID.String(),
			SelectedOptions:  selected,
			IsMarkedReview:   answer.IsMarkedReview,
			IsCorrect:        answer.IsCorrect,
			Score:            answer.Score,
			TimeTakenSeconds: answer.TimeTakenSeconds,
		}
		if !answer.UpdatedAt.IsZero() {
			entry.UpdatedAt = answer.UpdatedAt.UTC().Format(time.RFC3339)
		}
		if answer.AnsweredAt != nil {
			formatted := answer.AnsweredAt.UTC().Format(time.RFC3339)
			entry.AnsweredAt = &formatted
		}
		answers = append(answers, entry)
	}
	return structs.AssessmentAttemptResultResponse{
		ID:                       view.Attempt.ID.String(),
		QuizID:                   view.Attempt.QuizID.String(),
		UserID:                   view.Attempt.UserID,
		TestSnapshotID:           view.Attempt.TestSnapshotID.String(),
		AttemptNumber:            view.Attempt.AttemptNumber,
		Status:                   view.Attempt.Status,
		QuestionOrder:            uuidSliceToStrings(view.Attempt.QuestionOrder),
		NegativeMarksPerQuestion: view.Attempt.NegativeMarksPerQuestion,
		ExpectedMaxScore:         nullFloatPtr(view.Attempt.ExpectedMaxScore),
		StartedAt:                view.Attempt.StartedAt.UTC().Format(time.RFC3339),
		SubmittedAt:              nullTimePtr(view.Attempt.SubmittedAt),
		ExpiresAt:                nullTimePtr(view.Attempt.ExpiresAt),
		CreatedAt:                view.Attempt.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                view.Attempt.UpdatedAt.UTC().Format(time.RFC3339),
		Snapshot:                 toTestSnapshotLearnerResponse(view.Snapshot),
		Answers:                  answers,
		Summary: structs.AttemptResultSummaryResponse{
			CorrectCount:    view.Summary.CorrectCount,
			IncorrectCount:  view.Summary.IncorrectCount,
			UnansweredCount: view.Summary.UnansweredCount,
			UnscoredCount:   view.Summary.UnscoredCount,
			TotalScore:      view.Summary.TotalScore,
			MaxScore:        view.Summary.MaxScore,
		},
	}
}

func toAssessmentAttemptInstructionsResponse(view services.AssessmentAttemptInstructionsView) structs.AssessmentAttemptInstructionsResponse {
	quiz := structs.AttemptInstructionsQuizResponse{
		ID:                       view.Quiz.ID.String(),
		Title:                    view.Quiz.Title,
		AssessmentMode:           view.Quiz.AssessmentMode,
		Status:                   view.Quiz.Status,
		MaxAttempts:              view.Quiz.MaxAttempts,
		NegativeMarksPerQuestion: view.Quiz.NegativeMarksPerQuestion,
	}
	if view.Quiz.Description.Valid {
		quiz.Description = view.Quiz.Description.String
	}
	if view.Quiz.DurationSeconds.Valid {
		seconds := view.Quiz.DurationSeconds.Int64
		quiz.DurationSeconds = &seconds
	}
	response := structs.AssessmentAttemptInstructionsResponse{
		Quiz: quiz,
		Snapshot: structs.AttemptInstructionsSnapshotResponse{
			ID:            view.SnapshotID.String(),
			Status:        view.SnapshotStatus,
			QuestionCount: view.QuestionCount,
			CreatedAt:     view.SnapshotCreated.UTC().Format(time.RFC3339),
		},
		AttemptsConsumed: view.AttemptsConsumed,
		CanStart:         view.CanStart,
		CanResume:        view.CanResume,
		BlockReason:      view.BlockReason,
	}
	if view.ActiveAttempt != nil {
		active := &structs.AttemptInstructionsActiveAttemptResponse{
			ID:            view.ActiveAttempt.ID.String(),
			AttemptNumber: view.ActiveAttempt.AttemptNumber,
			Status:        view.ActiveAttempt.Status,
			StartedAt:     view.ActiveAttempt.StartedAt.UTC().Format(time.RFC3339),
			ExpiresAt:     nullTimePtr(view.ActiveAttempt.ExpiresAt),
		}
		response.ActiveAttempt = active
	}
	return response
}

func uuidSliceToStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func nullTimePtr(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.UTC().Format(time.RFC3339)
	return &formatted
}

func nullFloatPtr(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	v := value.Float64
	return &v
}

func toAssessmentAttemptResultDetailResponse(view services.AssessmentAttemptResultDetailView) structs.AssessmentAttemptResultDetailResponse {
	resp := structs.AssessmentAttemptResultDetailResponse{
		AttemptID:           view.Attempt.ID.String(),
		QuizID:              view.Attempt.QuizID.String(),
		Status:              view.Attempt.Status,
		StartedAt:           view.Attempt.StartedAt.UTC().Format(time.RFC3339),
		CanViewResult:       view.CanViewResult,
		CanShowScore:        view.CanShowScore,
		CanShowPassFail:     view.CanShowPassFail,
		CanReviewQuestions:  view.CanReviewQuestions,
		CanShowCorrectness:  view.CanShowCorrectness,
		CanShowExplanations: view.CanShowExplanations,
		Message:             view.Message,
	}

	if view.Attempt.SubmittedAt.Valid {
		formatted := view.Attempt.SubmittedAt.Time.UTC().Format(time.RFC3339)
		resp.SubmittedAt = &formatted
	}

	if view.Summary != nil {
		resp.Summary = &structs.AssessmentResultSummaryResponse{
			TotalScore:      view.Summary.TotalScore,
			MaxScore:        view.Summary.MaxScore,
			Percentage:      view.Summary.Percentage,
			Passed:          view.Summary.Passed,
			DurationSeconds: view.Summary.DurationSeconds,
			Answered:        view.Summary.Answered,
			Correct:         view.Summary.Correct,
			Incorrect:       view.Summary.Incorrect,
			Unanswered:      view.Summary.Unanswered,
			Unscored:        view.Summary.Unscored,
		}
	}

	if view.Review != nil {
		questions := make([]structs.QuestionReviewItemResponse, 0, len(view.Review.Questions))
		for _, q := range view.Review.Questions {
			opts := make([]structs.QuestionReviewOptionResponse, 0, len(q.Options))
			for _, o := range q.Options {
				opts = append(opts, structs.QuestionReviewOptionResponse{
					ID:       o.ID,
					Text:     o.Text,
					Media:    o.Media,
					Selected: o.Selected,
					Correct:  o.Correct,
				})
			}
			questions = append(questions, structs.QuestionReviewItemResponse{
				ID:               q.ID.String(),
				Position:         q.Position,
				Question:         q.Question,
				Type:             q.Type,
				Options:          opts,
				OptionsMedia:     q.OptionsMedia,
				QuestionMedia:    q.QuestionMedia,
				Resource:         q.Resource,
				Points:           q.Points,
				IsMarkedReview:   q.IsMarkedReview,
				IsCorrect:        q.IsCorrect,
				Score:            q.Score,
				TimeTakenSeconds: q.TimeTakenSeconds,
				Explanation:      q.Explanation,
			})
		}
		resp.Review = &structs.AssessmentResultReviewResponse{
			Questions: questions,
		}
	}

	return resp
}

func mapAssessmentAttemptError(c *fiber.Ctx, logger *zap.Logger, err error) error {
	switch {
	case errors.Is(err, models.ErrAssessmentAttemptNotFound),
		errors.Is(err, models.ErrTestSnapshotNotFound):
		return utils.JSONFail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, models.ErrAssessmentAttemptNotEntitled),
		errors.Is(err, models.ErrAssessmentAttemptQuizNotPublished),
		errors.Is(err, models.ErrAssessmentAttemptNotSelfPaced),
		errors.Is(err, models.ErrAssessmentAttemptOwnerMismatch):
		return utils.JSONError(c, http.StatusForbidden, err.Error())
	case errors.Is(err, models.ErrAssessmentAttemptNotSubmitted):
		return utils.JSONFail(c, http.StatusConflict, err.Error())
	case errors.Is(err, models.ErrAssessmentAttemptMaxReached),
		errors.Is(err, models.ErrAssessmentAttemptEmptySnapshot),
		errors.Is(err, models.ErrAssessmentAttemptForeignSnapshot),
		errors.Is(err, models.ErrAssessmentAttemptSnapshotRequired),
		errors.Is(err, models.ErrAttemptAnswerInvalidOptions),
		errors.Is(err, models.ErrAttemptAnswerInvalidOptionRef),
		errors.Is(err, models.ErrAttemptAnswerCardinality),
		errors.Is(err, models.ErrAttemptAnswerQuestionNotInSnapshot),
		errors.Is(err, models.ErrAttemptAnswerNotInProgress),
		errors.Is(err, models.ErrAttemptAnswerClientScoreForbidden),
		errors.Is(err, models.ErrAssessmentAttemptAlreadyTerminal),
		errors.Is(err, models.ErrAssessmentAttemptSubmitConflict):
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	default:
		logger.Error("assessment attempt operation failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetAssessmentAttempt)
	}
}
