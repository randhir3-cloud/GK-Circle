package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
)

const (
	QuizAssessmentModeSelfPaced = "SELF_PACED"
	QuizStatusPublished         = "PUBLISHED"
)

// AssessmentAttemptService owns self-paced attempt lifecycle bound to immutable snapshots
// and attempt-linked resolved snapshot items (EXAM-P5-T01–T04).
type AssessmentAttemptService struct {
	db                 *goqu.Database
	attemptModel       *models.AssessmentAttemptModel
	answerModel        *models.AttemptAnswerModel
	attemptItemModel   *models.AssessmentAttemptSnapshotItemModel
	snapshotModel      *models.TestSnapshotModel
	quizModel          *models.QuizModel
	sharedModel        *models.SharedQuizzesModel
	analyticsSvc       *AssessmentAnalyticsService
	learnerCache       *LearnerAnalyticsCache
	logger             *zap.Logger
}

func NewAssessmentAttemptService(db *goqu.Database, logger *zap.Logger) *AssessmentAttemptService {
	return &AssessmentAttemptService{
		db:               db,
		attemptModel:     models.InitAssessmentAttemptModel(db),
		answerModel:      models.InitAttemptAnswerModel(db),
		attemptItemModel: models.InitAssessmentAttemptSnapshotItemModel(db),
		snapshotModel:    models.InitTestSnapshotModel(db),
		quizModel:        models.InitQuizModel(db),
		sharedModel:      models.InitSharedQuizzesModel(db, logger),
		analyticsSvc:     NewAssessmentAnalyticsService(db, logger),
		logger:           logger,
	}
}

func (svc *AssessmentAttemptService) SetLearnerAnalyticsCache(cache *LearnerAnalyticsCache) {
	svc.learnerCache = cache
	if svc.analyticsSvc != nil {
		svc.analyticsSvc.SetLearnerAnalyticsCache(cache)
	}
}

type CreateAssessmentAttemptRequest struct {
	QuizID     uuid.UUID
	UserID     string
	UserEmail  string
	SnapshotID uuid.UUID
	// IsEditorPreview is true when the caller has quiz write/share permission.
	IsEditorPreview bool
	CorrelationID   string
}

type AssessmentAttemptLearnerView struct {
	Attempt  models.AssessmentAttempt
	Snapshot models.TestSnapshotLearnerView
}

type AssessmentAttemptEditorView struct {
	Attempt  models.AssessmentAttempt
	Snapshot models.TestSnapshot
}

// AutosaveAnswerRequest is learner-owned; user_id/score/correctness are never taken from the client.
type AutosaveAnswerRequest struct {
	QuizID           uuid.UUID
	AttemptID        uuid.UUID
	QuestionID       uuid.UUID
	UserID           string
	SelectedOptions  []int
	ClearAnswer      bool
	IsMarkedReview   bool
	TimeTakenSeconds *int
	CorrelationID    string
}

// AttemptAnswerLearnerView is answer-key-safe and score-free for resume/autosave responses.
type AttemptAnswerLearnerView struct {
	QuestionID       uuid.UUID
	SelectedOptions  []int
	IsMarkedReview   bool
	AnsweredAt       *time.Time
	TimeTakenSeconds *int
	UpdatedAt        time.Time
}

type AttemptResumeProgress struct {
	TotalQuestions     int `json:"total_questions"`
	AnsweredCount      int `json:"answered_count"`
	MarkedReviewCount  int `json:"marked_review_count"`
	UnansweredCount    int `json:"unanswered_count"`
}

type AssessmentAttemptStatusView struct {
	Status           string
	ExpiresAt        sql.NullTime
	RemainingSeconds *int64
}

type AssessmentAttemptResumeView struct {
	Attempt          models.AssessmentAttempt
	Snapshot         models.TestSnapshotLearnerView
	Answers          []AttemptAnswerLearnerView
	Progress         AttemptResumeProgress
	RemainingSeconds *int64
}

type QuestionReviewOptionView struct {
	ID       int
	Text     string
	Media    string
	Selected bool
	Correct  *bool
}

type QuestionReviewItemView struct {
	ID               uuid.UUID
	Position         int
	Question         string
	Type             int
	Options          []QuestionReviewOptionView
	OptionsMedia     string
	QuestionMedia    string
	Resource         string
	Points           int16
	IsMarkedReview   bool
	IsCorrect        *bool
	Score            *float64
	TimeTakenSeconds *int
	Explanation      *string
}

type AssessmentResultSummaryView struct {
	TotalScore      *float64
	MaxScore        *float64
	Percentage      *float64
	Passed          *bool
	DurationSeconds int
	Answered        int
	Correct         *int
	Incorrect       *int
	Unanswered      int
	Unscored        int
}

type AssessmentResultReviewView struct {
	Questions []QuestionReviewItemView
}

type AssessmentAttemptResultDetailView struct {
	Attempt             models.AssessmentAttempt
	CanViewResult       bool
	CanShowScore        bool
	CanShowPassFail     bool
	CanReviewQuestions  bool
	CanShowCorrectness  bool
	CanShowExplanations bool
	Message             string
	Summary             *AssessmentResultSummaryView
	Review              *AssessmentResultReviewView
}

// AttemptResultAnswerView includes correctness/marks after submit but never answer keys.
type AttemptResultAnswerView struct {
	QuestionID       uuid.UUID
	SelectedOptions  []int
	IsMarkedReview   bool
	IsCorrect        *bool
	Score            *float64
	AnsweredAt       *time.Time
	TimeTakenSeconds *int
	UpdatedAt        time.Time
}

type AttemptResultSummary struct {
	CorrectCount    int
	IncorrectCount  int
	UnansweredCount int
	UnscoredCount   int
	TotalScore      float64
	MaxScore        float64
}

type AssessmentAttemptResultView struct {
	Attempt  models.AssessmentAttempt
	Snapshot models.TestSnapshotLearnerView
	Answers  []AttemptResultAnswerView
	Summary  AttemptResultSummary
}

// AssessmentAttemptInstructionsView is the learner start-screen contract (no answer keys / no items).
type AssessmentAttemptInstructionsView struct {
	Quiz             models.QuizSelfPacedMeta
	SnapshotID       uuid.UUID
	SnapshotStatus   string
	QuestionCount    int
	SnapshotCreated  time.Time
	ActiveAttempt    *models.AssessmentAttempt
	AttemptsConsumed int
	CanStart         bool
	CanResume        bool
	BlockReason      string
}

func (svc *AssessmentAttemptService) Create(req CreateAssessmentAttemptRequest) (AssessmentAttemptLearnerView, bool, error) {
	if req.UserID == "" {
		return AssessmentAttemptLearnerView{}, false, models.ErrAssessmentAttemptOwnerMismatch
	}
	if req.SnapshotID == uuid.Nil {
		return AssessmentAttemptLearnerView{}, false, models.ErrAssessmentAttemptSnapshotRequired
	}

	quiz, err := svc.quizModel.GetSelfPacedMetaByID(req.QuizID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AssessmentAttemptLearnerView{}, false, models.ErrAssessmentAttemptNotEntitled
		}
		return AssessmentAttemptLearnerView{}, false, err
	}
	if err := svc.ensureCreateEntitlement(quiz, req); err != nil {
		return AssessmentAttemptLearnerView{}, false, err
	}

	if existing, found, err := svc.attemptModel.GetInProgress(req.QuizID, req.UserID); err != nil {
		return AssessmentAttemptLearnerView{}, false, err
	} else if found {
		view, viewErr := svc.learnerView(existing)
		return view, false, viewErr
	}

	consuming, err := svc.attemptModel.CountConsumingAttempts(req.QuizID, req.UserID)
	if err != nil {
		return AssessmentAttemptLearnerView{}, false, err
	}
	if consuming >= quiz.MaxAttempts {
		return AssessmentAttemptLearnerView{}, false, models.ErrAssessmentAttemptMaxReached
	}

	snapshot, err := svc.snapshotModel.GetSnapshotByID(req.QuizID, req.SnapshotID)
	if err != nil {
		if errors.Is(err, models.ErrTestSnapshotNotFound) {
			return AssessmentAttemptLearnerView{}, false, models.ErrAssessmentAttemptForeignSnapshot
		}
		return AssessmentAttemptLearnerView{}, false, err
	}
	if snapshot.QuizID != req.QuizID {
		return AssessmentAttemptLearnerView{}, false, models.ErrAssessmentAttemptForeignSnapshot
	}
	if snapshot.QuestionCount == 0 || len(snapshot.Items) == 0 {
		return AssessmentAttemptLearnerView{}, false, models.ErrAssessmentAttemptEmptySnapshot
	}

	order := make([]uuid.UUID, 0, len(snapshot.Items))
	freezeItems := make([]models.CreateAttemptSnapshotItemParams, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		order = append(order, item.QuestionID)
		freezeItems = append(freezeItems, models.SnapshotItemFromTestFreeze(item))
	}
	expectedMax := ExpectedMaxScoreFromSnapshotItems(freezeItems)

	startedAt := time.Now().UTC()
	var expiresAt *time.Time
	if quiz.DurationSeconds.Valid && quiz.DurationSeconds.Int64 > 0 {
		expires := startedAt.Add(time.Duration(quiz.DurationSeconds.Int64) * time.Second)
		expiresAt = &expires
	}

	attemptNumber, err := svc.attemptModel.NextAttemptNumber(req.QuizID, req.UserID)
	if err != nil {
		return AssessmentAttemptLearnerView{}, false, err
	}

	quizOwnerID := ""
	if quiz.CreatorID.Valid {
		quizOwnerID = quiz.CreatorID.String
	}
	correlationID := req.CorrelationID

	attempt, created, err := svc.attemptModel.CreateInProgress(models.CreateAssessmentAttemptParams{
		QuizID:                   req.QuizID,
		UserID:                   req.UserID,
		TestSnapshotID:           snapshot.ID,
		AttemptNumber:            attemptNumber,
		QuestionOrder:            order,
		NegativeMarksPerQuestion: quiz.NegativeMarksPerQuestion,
		ExpectedMaxScore:         expectedMax,
		StartedAt:                startedAt,
		ExpiresAt:                expiresAt,
		SnapshotItems:            freezeItems,
		BeforeCommit: func(tx *goqu.TxDatabase, attemptID uuid.UUID) error {
			return svc.analyticsSvc.RecordAuthoritativeEventTx(tx, AuthoritativeAnalyticsEventInput{
				AttemptID:      attemptID,
				QuizID:         req.QuizID,
				UserID:         req.UserID,
				QuizOwnerID:    quizOwnerID,
				EventType:      structs.EventAttemptStarted,
				EventSource:    structs.EventSourceHTTP,
				CorrelationID:  correlationID,
				IdempotencyKey: AttemptStartedIdempotencyKey(attemptID),
				Metadata: map[string]interface{}{
					"snapshot_id":    snapshot.ID.String(),
					"attempt_number": attemptNumber,
				},
				OccurredAt: startedAt,
			})
		},
	})
	if err != nil {
		return AssessmentAttemptLearnerView{}, false, err
	}
	view, viewErr := svc.learnerView(attempt)
	return view, created, viewErr
}

// GetInstructions returns key-safe quiz/snapshot rules plus start/resume eligibility for the player.
func (svc *AssessmentAttemptService) GetInstructions(
	quizID, snapshotID uuid.UUID,
	userID string,
	isEditorPreview bool,
) (AssessmentAttemptInstructionsView, error) {
	if userID == "" {
		return AssessmentAttemptInstructionsView{}, models.ErrAssessmentAttemptOwnerMismatch
	}
	if snapshotID == uuid.Nil {
		return AssessmentAttemptInstructionsView{}, models.ErrAssessmentAttemptSnapshotRequired
	}

	quiz, err := svc.quizModel.GetSelfPacedMetaByID(quizID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AssessmentAttemptInstructionsView{}, models.ErrAssessmentAttemptNotEntitled
		}
		return AssessmentAttemptInstructionsView{}, err
	}
	if err := svc.ensureCreateEntitlement(quiz, CreateAssessmentAttemptRequest{
		QuizID:          quizID,
		UserID:          userID,
		IsEditorPreview: isEditorPreview,
	}); err != nil {
		return AssessmentAttemptInstructionsView{}, err
	}

	snapshot, err := svc.snapshotModel.GetSnapshotByID(quizID, snapshotID)
	if err != nil {
		if errors.Is(err, models.ErrTestSnapshotNotFound) {
			return AssessmentAttemptInstructionsView{}, models.ErrAssessmentAttemptForeignSnapshot
		}
		return AssessmentAttemptInstructionsView{}, err
	}
	if snapshot.QuizID != quizID {
		return AssessmentAttemptInstructionsView{}, models.ErrAssessmentAttemptForeignSnapshot
	}
	if snapshot.QuestionCount == 0 || len(snapshot.Items) == 0 {
		return AssessmentAttemptInstructionsView{}, models.ErrAssessmentAttemptEmptySnapshot
	}

	view := AssessmentAttemptInstructionsView{
		Quiz:            quiz,
		SnapshotID:      snapshot.ID,
		SnapshotStatus:  snapshot.Status,
		QuestionCount:   snapshot.QuestionCount,
		SnapshotCreated: snapshot.CreatedAt,
		CanStart:        true,
	}

	if existing, found, err := svc.attemptModel.GetInProgress(quizID, userID); err != nil {
		return AssessmentAttemptInstructionsView{}, err
	} else if found {
		view.ActiveAttempt = &existing
		view.CanResume = true
		view.CanStart = false
		view.BlockReason = "an in-progress attempt already exists; resume it to continue"
		return view, nil
	}

	consuming, err := svc.attemptModel.CountConsumingAttempts(quizID, userID)
	if err != nil {
		return AssessmentAttemptInstructionsView{}, err
	}
	view.AttemptsConsumed = consuming
	if consuming >= quiz.MaxAttempts {
		view.CanStart = false
		view.BlockReason = models.ErrAssessmentAttemptMaxReached.Error()
	}
	return view, nil
}

func (svc *AssessmentAttemptService) GetLearner(quizID, attemptID uuid.UUID, userID string) (AssessmentAttemptLearnerView, error) {
	attempt, err := svc.attemptModel.GetByID(quizID, attemptID)
	if err != nil {
		return AssessmentAttemptLearnerView{}, err
	}
	if attempt.UserID != userID {
		return AssessmentAttemptLearnerView{}, models.ErrAssessmentAttemptNotFound
	}
	return svc.learnerView(attempt)
}

func (svc *AssessmentAttemptService) ListMine(quizID uuid.UUID, userID string) ([]AssessmentAttemptLearnerView, error) {
	attempts, err := svc.attemptModel.ListByQuizAndUser(quizID, userID)
	if err != nil {
		return nil, err
	}
	views := make([]AssessmentAttemptLearnerView, 0, len(attempts))
	for _, attempt := range attempts {
		view, err := svc.learnerView(attempt)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (svc *AssessmentAttemptService) GetEditor(quizID, attemptID uuid.UUID) (AssessmentAttemptEditorView, error) {
	attempt, err := svc.attemptModel.GetByID(quizID, attemptID)
	if err != nil {
		return AssessmentAttemptEditorView{}, err
	}
	snapshot, err := svc.snapshotModel.GetSnapshotByID(quizID, attempt.TestSnapshotID)
	if err != nil {
		return AssessmentAttemptEditorView{}, err
	}
	return AssessmentAttemptEditorView{Attempt: attempt, Snapshot: snapshot}, nil
}

func (svc *AssessmentAttemptService) GetStatus(quizID, attemptID uuid.UUID, userID string) (AssessmentAttemptStatusView, error) {
	attempt, err := svc.attemptModel.GetStatusByID(quizID, attemptID)
	if err != nil {
		return AssessmentAttemptStatusView{}, err
	}
	if attempt.UserID != userID {
		return AssessmentAttemptStatusView{}, models.ErrAssessmentAttemptNotFound
	}
	var remaining *int64
	if attempt.ExpiresAt.Valid {
		rem := int64(time.Until(attempt.ExpiresAt.Time).Seconds())
		if rem < 0 {
			rem = 0
		}
		remaining = &rem
	}
	return AssessmentAttemptStatusView{
		Status:           attempt.Status,
		ExpiresAt:        attempt.ExpiresAt,
		RemainingSeconds: remaining,
	}, nil
}

func (svc *AssessmentAttemptService) Resume(quizID, attemptID uuid.UUID, userID string) (AssessmentAttemptResumeView, error) {
	attempt, err := svc.attemptModel.GetByID(quizID, attemptID)
	if err != nil {
		return AssessmentAttemptResumeView{}, err
	}
	if attempt.UserID != userID {
		return AssessmentAttemptResumeView{}, models.ErrAssessmentAttemptNotFound
	}

	snapshot, err := svc.snapshotModel.GetSnapshotByID(attempt.QuizID, attempt.TestSnapshotID)
	if err != nil {
		return AssessmentAttemptResumeView{}, err
	}
	learnerSnapshot, err := svc.learnerSnapshotFromAttempt(attempt, snapshot)
	if err != nil {
		return AssessmentAttemptResumeView{}, err
	}

	answers, err := svc.answerModel.ListByAttemptID(attempt.ID)
	if err != nil {
		return AssessmentAttemptResumeView{}, err
	}

	learnerAnswers := make([]AttemptAnswerLearnerView, 0, len(answers))
	answeredCount := 0
	markedCount := 0
	for _, answer := range answers {
		view, err := toAttemptAnswerLearnerView(answer)
		if err != nil {
			return AssessmentAttemptResumeView{}, err
		}
		learnerAnswers = append(learnerAnswers, view)
		if len(view.SelectedOptions) > 0 {
			answeredCount++
		}
		if view.IsMarkedReview {
			markedCount++
		}
	}

	total := learnerSnapshot.QuestionCount
	unanswered := total - answeredCount
	if unanswered < 0 {
		unanswered = 0
	}

	var remaining *int64
	if attempt.ExpiresAt.Valid {
		rem := int64(time.Until(attempt.ExpiresAt.Time).Seconds())
		if rem < 0 {
			rem = 0
		}
		remaining = &rem
	}

	return AssessmentAttemptResumeView{
		Attempt:  attempt,
		Snapshot: learnerSnapshot,
		Answers:  learnerAnswers,
		Progress: AttemptResumeProgress{
			TotalQuestions:    total,
			AnsweredCount:     answeredCount,
			MarkedReviewCount: markedCount,
			UnansweredCount:   unanswered,
		},
		RemainingSeconds: remaining,
	}, nil
}

func (svc *AssessmentAttemptService) AutosaveAnswer(req AutosaveAnswerRequest) (AttemptAnswerLearnerView, error) {
	if req.UserID == "" {
		return AttemptAnswerLearnerView{}, models.ErrAssessmentAttemptOwnerMismatch
	}
	if req.QuestionID == uuid.Nil {
		return AttemptAnswerLearnerView{}, models.ErrAttemptAnswerQuestionNotInSnapshot
	}

	tx, err := svc.db.Begin()
	if err != nil {
		return AttemptAnswerLearnerView{}, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = tx.Rollback()
		}
	}()

	attempt, err := svc.attemptModel.GetByIDForUpdate(tx, req.QuizID, req.AttemptID)
	if err != nil {
		return AttemptAnswerLearnerView{}, err
	}
	if attempt.UserID != req.UserID {
		return AttemptAnswerLearnerView{}, models.ErrAssessmentAttemptNotFound
	}
	if attempt.Status != models.AttemptStatusInProgress {
		return AttemptAnswerLearnerView{}, models.ErrAttemptAnswerNotInProgress
	}

	item, err := svc.attemptItemModel.GetByAttemptAndQuestion(attempt.ID, req.QuestionID)
	if err != nil {
		if errors.Is(err, models.ErrAttemptSnapshotItemNotFound) {
			return AttemptAnswerLearnerView{}, models.ErrAttemptAnswerQuestionNotInSnapshot
		}
		return AttemptAnswerLearnerView{}, err
	}
	if err := models.ValidateSelectedOptionsAgainstSnapshot(item.Type, item.Options, req.SelectedOptions, req.ClearAnswer); err != nil {
		return AttemptAnswerLearnerView{}, err
	}

	answer, err := svc.answerModel.UpsertAnswer(tx, models.UpsertAttemptAnswerParams{
		AttemptID:        attempt.ID,
		QuestionID:       req.QuestionID,
		SelectedOptions:  req.SelectedOptions,
		ClearAnswer:      req.ClearAnswer || len(req.SelectedOptions) == 0,
		IsMarkedReview:   req.IsMarkedReview,
		TimeTakenSeconds: req.TimeTakenSeconds,
	})
	if err != nil {
		return AttemptAnswerLearnerView{}, err
	}
	if err := svc.attemptModel.TouchUpdatedAtTx(tx, req.QuizID, attempt.ID, time.Now().UTC()); err != nil {
		return AttemptAnswerLearnerView{}, err
	}
	if err := tx.Commit(); err != nil {
		return AttemptAnswerLearnerView{}, err
	}
	ok = true

	quizOwnerID := ""
	if quizMeta, quizErr := svc.quizModel.GetSelfPacedMetaByID(req.QuizID); quizErr == nil && quizMeta.CreatorID.Valid {
		quizOwnerID = quizMeta.CreatorID.String
	}
	svc.analyticsSvc.RecordServerTelemetryBounded(context.Background(), AuthoritativeAnalyticsEventInput{
		AttemptID:      attempt.ID,
		QuizID:         req.QuizID,
		UserID:         req.UserID,
		QuizOwnerID:    quizOwnerID,
		EventType:      structs.EventAttemptAutosaved,
		EventSource:    structs.EventSourceHTTP,
		CorrelationID:  req.CorrelationID,
		IdempotencyKey: fmt.Sprintf("attempt-autosaved:%s:%s:%d", attempt.ID.String(), req.QuestionID.String(), answer.UpdatedAt.UnixNano()),
		Metadata: map[string]interface{}{
			"question_id": req.QuestionID.String(),
		},
		OccurredAt: time.Now().UTC(),
	})

	return toAttemptAnswerLearnerView(answer)
}

func (svc *AssessmentAttemptService) Submit(quizID, attemptID uuid.UUID, userID, correlationID string) (AssessmentAttemptResultView, bool, error) {
	if userID == "" {
		return AssessmentAttemptResultView{}, false, models.ErrAssessmentAttemptOwnerMismatch
	}

	tx, err := svc.db.Begin()
	if err != nil {
		return AssessmentAttemptResultView{}, false, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = tx.Rollback()
		}
	}()

	attempt, err := svc.attemptModel.GetByIDForUpdate(tx, quizID, attemptID)
	if err != nil {
		return AssessmentAttemptResultView{}, false, err
	}
	if attempt.UserID != userID {
		return AssessmentAttemptResultView{}, false, models.ErrAssessmentAttemptNotFound
	}

	if attempt.Status == models.AttemptStatusSubmitted ||
		attempt.Status == models.AttemptStatusAutoSubmitted {
		if err := tx.Commit(); err != nil {
			return AssessmentAttemptResultView{}, false, err
		}
		ok = true
		view, viewErr := svc.buildResultView(attempt)
		return view, false, viewErr
	}
	if attempt.Status != models.AttemptStatusInProgress {
		return AssessmentAttemptResultView{}, false, models.ErrAttemptAnswerNotInProgress
	}

	attemptItems, err := svc.attemptItemModel.ListByAttemptID(attempt.ID)
	if err != nil {
		return AssessmentAttemptResultView{}, false, err
	}
	if len(attemptItems) == 0 {
		return AssessmentAttemptResultView{}, false, models.ErrAssessmentAttemptEmptySnapshot
	}

	persisted, err := svc.answerModel.ListByAttemptID(attempt.ID)
	if err != nil {
		return AssessmentAttemptResultView{}, false, err
	}
	answersByQuestion := make(map[uuid.UUID]models.AttemptAnswer, len(persisted))
	for _, answer := range persisted {
		answersByQuestion[answer.QuestionID] = answer
	}

	inputs := make([]PCSQuestionScoreInput, 0, len(attemptItems))
	for _, item := range attemptItems {
		selected := []int{}
		if answer, found := answersByQuestion[item.QuestionID]; found {
			decoded, decodeErr := models.DecodeSelectedOptions(answer.SelectedOptions)
			if decodeErr != nil {
				return AssessmentAttemptResultView{}, false, decodeErr
			}
			selected = decoded
		}
		inputs = append(inputs, AttemptSnapshotItemToScoreInput(item, selected))
	}

	scored := ScorePCSAttempt(inputs, PCSScoreConfig{
		NegativeMarksPerQuestion: attempt.NegativeMarksPerQuestion,
	})

	submittedAt := time.Now().UTC()
	timeTaken := int(submittedAt.Sub(attempt.StartedAt).Seconds())
	if timeTaken < 0 {
		timeTaken = 0
	}

	for _, questionResult := range scored.QuestionResults {
		existing := answersByQuestion[questionResult.QuestionID]
		var answeredAt *time.Time
		if existing.AnsweredAt.Valid {
			t := existing.AnsweredAt.Time.UTC()
			answeredAt = &t
		}
		var timeTakenSeconds *int
		if existing.TimeTakenSeconds.Valid {
			seconds := int(existing.TimeTakenSeconds.Int64)
			timeTakenSeconds = &seconds
		}
		score := questionResult.Score
		err := svc.answerModel.ApplyScoreOutcomeTx(tx, models.ApplyScoreOutcomeParams{
			AttemptID:        attempt.ID,
			QuestionID:       questionResult.QuestionID,
			SelectedOptions:  questionResult.SelectedOptions,
			IsMarkedReview:   existing.IsMarkedReview,
			IsCorrect:        questionResult.IsCorrect,
			Score:            score,
			AnsweredAt:       answeredAt,
			TimeTakenSeconds: timeTakenSeconds,
		})
		if err != nil {
			return AssessmentAttemptResultView{}, false, err
		}
	}

	finalStatus := models.AttemptStatusSubmitted
	if attempt.ExpiresAt.Valid && submittedAt.After(attempt.ExpiresAt.Time) {
		finalStatus = models.AttemptStatusAutoSubmitted
	}

	if err := svc.attemptModel.FinalizeAttemptTx(
		tx,
		quizID,
		attempt.ID,
		scored.TotalScore,
		scored.MaxScore,
		timeTaken,
		submittedAt,
		finalStatus,
	); err != nil {
		if errors.Is(err, models.ErrAssessmentAttemptSubmitConflict) {
			_ = tx.Rollback()
			ok = true // already rolled back; avoid double rollback
			existing, getErr := svc.attemptModel.GetByID(quizID, attemptID)
			if getErr != nil {
				return AssessmentAttemptResultView{}, false, getErr
			}
			view, viewErr := svc.buildResultView(existing)
			return view, false, viewErr
		}
		return AssessmentAttemptResultView{}, false, err
	}

	quizOwnerID := ""
	eventType := structs.EventAttemptSubmitted
	idemKey := AttemptSubmittedIdempotencyKey(attempt.ID)
	if finalStatus == models.AttemptStatusAutoSubmitted {
		eventType = structs.EventAttemptAutoSubmitted
		idemKey = AttemptAutoSubmittedIdempotencyKey(attempt.ID)
	}
	if quizMeta, quizErr := svc.quizModel.GetSelfPacedMetaByID(quizID); quizErr == nil && quizMeta.CreatorID.Valid {
		quizOwnerID = quizMeta.CreatorID.String
	}
	if err := svc.analyticsSvc.RecordAuthoritativeEventTx(tx, AuthoritativeAnalyticsEventInput{
		AttemptID:      attempt.ID,
		QuizID:         quizID,
		UserID:         userID,
		QuizOwnerID:    quizOwnerID,
		EventType:      eventType,
		EventSource:    structs.EventSourceHTTP,
		CorrelationID:  correlationID,
		IdempotencyKey: idemKey,
		Metadata: map[string]interface{}{
			"status":             finalStatus,
			"time_taken_seconds": timeTaken,
		},
		OccurredAt: submittedAt,
	}); err != nil {
		return AssessmentAttemptResultView{}, false, err
	}

	if err := tx.Commit(); err != nil {
		return AssessmentAttemptResultView{}, false, err
	}
	ok = true

	finalAttempt, err := svc.attemptModel.GetByID(quizID, attemptID)
	if err != nil {
		return AssessmentAttemptResultView{}, false, err
	}
	view, viewErr := svc.buildResultView(finalAttempt)
	if viewErr == nil && svc.learnerCache != nil {
		svc.learnerCache.BumpVersion(userID)
	}
	return view, true, viewErr
}

func (svc *AssessmentAttemptService) GetResult(quizID, attemptID uuid.UUID, userID, correlationID string) (AssessmentAttemptResultDetailView, error) {
	attempt, err := svc.attemptModel.GetByID(quizID, attemptID)
	if err != nil {
		return AssessmentAttemptResultDetailView{}, err
	}
	if attempt.UserID != userID {
		return AssessmentAttemptResultDetailView{}, models.ErrAssessmentAttemptNotFound
	}
	if attempt.Status != models.AttemptStatusSubmitted &&
		attempt.Status != models.AttemptStatusAutoSubmitted {
		return AssessmentAttemptResultDetailView{}, models.ErrAssessmentAttemptNotSubmitted
	}

	quiz, err := svc.quizModel.GetSelfPacedMetaByID(attempt.QuizID)
	if err != nil {
		return AssessmentAttemptResultDetailView{}, err
	}

	now := time.Now().UTC()
	canViewResult := false
	policy := quiz.ResultReleasePolicy
	if policy == "" {
		policy = "IMMEDIATE"
	}

	switch policy {
	case "IMMEDIATE":
		canViewResult = true
	case "MANUAL":
		canViewResult = quiz.ResultsReleased
	case "SCHEDULED":
		// Exact-once scheduled release evaluator (Phase 7 contract + analytics event).
		if _, evalErr := svc.analyticsSvc.EnsureScheduledReleaseEffective(quizID, correlationID); evalErr != nil {
			svc.logger.Warn("scheduled release evaluator failed", zap.Error(evalErr), zap.String("quiz_id", quizID.String()))
		}
		quiz, err = svc.quizModel.GetSelfPacedMetaByID(attempt.QuizID)
		if err != nil {
			return AssessmentAttemptResultDetailView{}, err
		}
		if quiz.ResultsReleased {
			canViewResult = true
		} else if quiz.ResultsScheduledAt.Valid && !quiz.ResultsScheduledAt.Time.IsZero() {
			schedTime := quiz.ResultsScheduledAt.Time.UTC()
			if now.Equal(schedTime) || now.After(schedTime) {
				canViewResult = true
			}
		}
	default:
		// Fail closed for unknown/invalid policy
		canViewResult = false
	}

	if !canViewResult {
		return AssessmentAttemptResultDetailView{
			Attempt:             attempt,
			CanViewResult:       false,
			CanShowScore:        false,
			CanShowPassFail:     false,
			CanReviewQuestions:  false,
			CanShowCorrectness:  false,
			CanShowExplanations: false,
			Message:             "Results for this assessment have not been released yet.",
			Summary:             nil,
			Review:              nil,
		}, nil
	}

	attemptItems, err := svc.attemptItemModel.ListByAttemptID(attempt.ID)
	if err != nil {
		return AssessmentAttemptResultDetailView{}, err
	}

	// Deterministic ordering by snapshot item position
	sort.Slice(attemptItems, func(i, j int) bool {
		return attemptItems[i].Position < attemptItems[j].Position
	})

	answers, err := svc.answerModel.ListByAttemptID(attempt.ID)
	if err != nil {
		return AssessmentAttemptResultDetailView{}, err
	}
	answersByID := make(map[uuid.UUID]models.AttemptAnswer, len(answers))
	for _, answer := range answers {
		answersByID[answer.QuestionID] = answer
	}

	totalScore := 0.0
	if attempt.TotalScore.Valid {
		totalScore = attempt.TotalScore.Float64
	}
	maxScore := 0.0
	if attempt.MaxScore.Valid {
		maxScore = attempt.MaxScore.Float64
	}

	var percentage float64
	if maxScore > 0 {
		percentage = math.Floor((totalScore / maxScore) * 100)
		if percentage < 0 {
			percentage = 0
		}
	}

	durationSeconds := 0
	if attempt.TimeTakenSeconds.Valid {
		durationSeconds = int(attempt.TimeTakenSeconds.Int64)
	}

	answeredCount := 0
	correctCount := 0
	incorrectCount := 0
	unansweredCount := 0
	unscoredCount := 0

	for _, item := range attemptItems {
		answer, found := answersByID[item.QuestionID]
		if !found {
			unansweredCount++
			continue
		}
		if item.Type == constants.Survey {
			unscoredCount++
			continue
		}
		decodedOptions, _ := models.DecodeSelectedOptions(answer.SelectedOptions)
		if len(decodedOptions) == 0 {
			unansweredCount++
			continue
		}
		answeredCount++
		if answer.IsCorrect.Valid {
			if answer.IsCorrect.Bool {
				correctCount++
			} else {
				incorrectCount++
			}
		} else {
			unscoredCount++
		}
	}

	canShowScore := quiz.ShowScore
	canShowPassFail := quiz.ShowPassFail
	canReviewQuestions := quiz.AllowAnswerReview
	canShowCorrectness := canReviewQuestions && quiz.ShowCorrectness
	canShowExplanations := canReviewQuestions && quiz.ShowExplanations

	var summaryScore *float64
	var summaryMax *float64
	var summaryPct *float64
	if canShowScore {
		summaryScore = &totalScore
		summaryMax = &maxScore
		summaryPct = &percentage
	}

	var passFailStatus *bool
	if canShowPassFail && summaryPct != nil {
		passVal := *summaryPct >= 50.0 // Standard pass threshold
		passFailStatus = &passVal
	}

	var ptrCorrect *int
	var ptrIncorrect *int
	if canShowCorrectness {
		ptrCorrect = &correctCount
		ptrIncorrect = &incorrectCount
	}

	summaryView := &AssessmentResultSummaryView{
		TotalScore:      summaryScore,
		MaxScore:        summaryMax,
		Percentage:      summaryPct,
		Passed:          passFailStatus,
		DurationSeconds: durationSeconds,
		Answered:        answeredCount,
		Correct:         ptrCorrect,
		Incorrect:       ptrIncorrect,
		Unanswered:      unansweredCount,
		Unscored:        unscoredCount,
	}

	var reviewView *AssessmentResultReviewView
	if canReviewQuestions {
		reviewItems := make([]QuestionReviewItemView, 0, len(attemptItems))
		for _, item := range attemptItems {
			answer, found := answersByID[item.QuestionID]
			selectedMap := make(map[int]bool)
			isMarked := false
			var isCorrect *bool
			var score *float64
			var timeTakenSeconds *int

			if found {
				isMarked = answer.IsMarkedReview
				if decoded, err := models.DecodeSelectedOptions(answer.SelectedOptions); err == nil {
					for _, opt := range decoded {
						selectedMap[opt] = true
					}
				}
				if canShowCorrectness && answer.IsCorrect.Valid {
					v := answer.IsCorrect.Bool
					isCorrect = &v
				}
				if canShowScore && canShowCorrectness && answer.Score.Valid {
					v := answer.Score.Float64
					score = &v
				}
				if answer.TimeTakenSeconds.Valid {
					s := int(answer.TimeTakenSeconds.Int64)
					timeTakenSeconds = &s
				}
			}

			// Build option presentation array
			rawOptionsMap := item.Options
			correctAnswers := item.Answers
			correctAnswersMap := make(map[int]bool, len(correctAnswers))
			for _, c := range correctAnswers {
				correctAnswersMap[c] = true
			}

			optionKeys := make([]int, 0, len(rawOptionsMap))
			for kStr := range rawOptionsMap {
				var kInt int
				if _, err := fmt.Sscanf(kStr, "%d", &kInt); err == nil {
					optionKeys = append(optionKeys, kInt)
				}
			}
			sort.Ints(optionKeys)

			optionsView := make([]QuestionReviewOptionView, 0, len(optionKeys))
			for _, optID := range optionKeys {
				optText := rawOptionsMap[strconv.Itoa(optID)]
				isSelected := selectedMap[optID]
				var isOptionCorrect *bool
				if canShowCorrectness {
					c := correctAnswersMap[optID]
					isOptionCorrect = &c
				}
				optionsView = append(optionsView, QuestionReviewOptionView{
					ID:       optID,
					Text:     optText,
					Selected: isSelected,
					Correct:  isOptionCorrect,
				})
			}

			var exp *string
			if canShowExplanations && item.Resource != nil && *item.Resource != "" {
				exp = item.Resource
			}

			var pts int16
			if item.Points != nil {
				pts = *item.Points
			}

			resStr := ""
			if item.Resource != nil {
				resStr = *item.Resource
			}

			reviewItems = append(reviewItems, QuestionReviewItemView{
				ID:               item.QuestionID,
				Position:         item.Position,
				Question:         item.Question,
				Type:             item.Type,
				Options:          optionsView,
				OptionsMedia:     item.OptionsMedia,
				QuestionMedia:    item.QuestionMedia,
				Resource:         resStr,
				Points:           pts,
				IsMarkedReview:   isMarked,
				IsCorrect:        isCorrect,
				Score:            score,
				TimeTakenSeconds: timeTakenSeconds,
				Explanation:      exp,
			})
		}
		reviewView = &AssessmentResultReviewView{Questions: reviewItems}
	}

	quizOwnerID := ""
	if quiz.CreatorID.Valid {
		quizOwnerID = quiz.CreatorID.String
	}
	svc.analyticsSvc.RecordServerTelemetryBounded(context.Background(), AuthoritativeAnalyticsEventInput{
		AttemptID:      attempt.ID,
		QuizID:         quizID,
		UserID:         userID,
		QuizOwnerID:    quizOwnerID,
		EventType:      structs.EventResultViewed,
		EventSource:    structs.EventSourceHTTP,
		CorrelationID:  correlationID,
		IdempotencyKey: fmt.Sprintf("result-viewed:%s:%s:%d", attempt.ID.String(), userID, time.Now().UTC().Unix()),
		Metadata:       map[string]interface{}{"policy": policy},
		OccurredAt:     time.Now().UTC(),
	})

	return AssessmentAttemptResultDetailView{
		Attempt:             attempt,
		CanViewResult:       canViewResult,
		CanShowScore:        canShowScore,
		CanShowPassFail:     canShowPassFail,
		CanReviewQuestions:  canReviewQuestions,
		CanShowCorrectness:  canShowCorrectness,
		CanShowExplanations: canShowExplanations,
		Summary:             summaryView,
		Review:              reviewView,
	}, nil
}

func (svc *AssessmentAttemptService) buildResultView(attempt models.AssessmentAttempt) (AssessmentAttemptResultView, error) {
	snapshot, err := svc.snapshotModel.GetSnapshotByID(attempt.QuizID, attempt.TestSnapshotID)
	if err != nil {
		return AssessmentAttemptResultView{}, err
	}
	learnerSnapshot, err := svc.learnerSnapshotFromAttempt(attempt, snapshot)
	if err != nil {
		return AssessmentAttemptResultView{}, err
	}
	attemptItems, err := svc.attemptItemModel.ListByAttemptID(attempt.ID)
	if err != nil {
		return AssessmentAttemptResultView{}, err
	}
	answers, err := svc.answerModel.ListByAttemptID(attempt.ID)
	if err != nil {
		return AssessmentAttemptResultView{}, err
	}

	answersByID := make(map[uuid.UUID]AttemptResultAnswerView, len(answers))
	for _, answer := range answers {
		view, err := toAttemptResultAnswerView(answer)
		if err != nil {
			return AssessmentAttemptResultView{}, err
		}
		answersByID[answer.QuestionID] = view
	}

	summary := AttemptResultSummary{}
	if attempt.TotalScore.Valid {
		summary.TotalScore = attempt.TotalScore.Float64
	}
	if attempt.MaxScore.Valid {
		summary.MaxScore = attempt.MaxScore.Float64
	}

	orderedAnswers := make([]AttemptResultAnswerView, 0, len(attemptItems))
	for _, item := range attemptItems {
		answer, found := answersByID[item.QuestionID]
		if !found {
			summary.UnansweredCount++
			orderedAnswers = append(orderedAnswers, AttemptResultAnswerView{
				QuestionID:      item.QuestionID,
				SelectedOptions: []int{},
			})
			continue
		}
		orderedAnswers = append(orderedAnswers, answer)
		if item.Type == constants.Survey {
			summary.UnscoredCount++
			continue
		}
		if len(answer.SelectedOptions) == 0 {
			summary.UnansweredCount++
			continue
		}
		if answer.IsCorrect == nil {
			summary.UnscoredCount++
			continue
		}
		if *answer.IsCorrect {
			summary.CorrectCount++
		} else {
			summary.IncorrectCount++
		}
	}

	return AssessmentAttemptResultView{
		Attempt:  attempt,
		Snapshot: learnerSnapshot,
		Answers:  orderedAnswers,
		Summary:  summary,
	}, nil
}

func toAttemptResultAnswerView(answer models.AttemptAnswer) (AttemptResultAnswerView, error) {
	selected, err := models.DecodeSelectedOptions(answer.SelectedOptions)
	if err != nil {
		return AttemptResultAnswerView{}, err
	}
	view := AttemptResultAnswerView{
		QuestionID:      answer.QuestionID,
		SelectedOptions: selected,
		IsMarkedReview:  answer.IsMarkedReview,
		UpdatedAt:       answer.UpdatedAt,
	}
	if answer.AnsweredAt.Valid {
		t := answer.AnsweredAt.Time.UTC()
		view.AnsweredAt = &t
	}
	if answer.TimeTakenSeconds.Valid {
		seconds := int(answer.TimeTakenSeconds.Int64)
		view.TimeTakenSeconds = &seconds
	}
	if answer.IsCorrect.Valid {
		v := answer.IsCorrect.Bool
		view.IsCorrect = &v
	}
	if answer.Score.Valid {
		v := answer.Score.Float64
		view.Score = &v
	}
	return view, nil
}

func findSnapshotItem(snapshot models.TestSnapshot, questionID uuid.UUID) (models.TestSnapshotItem, bool) {
	for _, item := range snapshot.Items {
		if item.QuestionID == questionID {
			return item, true
		}
	}
	return models.TestSnapshotItem{}, false
}

func toAttemptAnswerLearnerView(answer models.AttemptAnswer) (AttemptAnswerLearnerView, error) {
	selected, err := models.DecodeSelectedOptions(answer.SelectedOptions)
	if err != nil {
		return AttemptAnswerLearnerView{}, err
	}
	view := AttemptAnswerLearnerView{
		QuestionID:      answer.QuestionID,
		SelectedOptions: selected,
		IsMarkedReview:  answer.IsMarkedReview,
		UpdatedAt:       answer.UpdatedAt,
	}
	if answer.AnsweredAt.Valid {
		t := answer.AnsweredAt.Time.UTC()
		view.AnsweredAt = &t
	}
	if answer.TimeTakenSeconds.Valid {
		seconds := int(answer.TimeTakenSeconds.Int64)
		view.TimeTakenSeconds = &seconds
	}
	return view, nil
}

func (svc *AssessmentAttemptService) learnerView(attempt models.AssessmentAttempt) (AssessmentAttemptLearnerView, error) {
	snapshot, err := svc.snapshotModel.GetSnapshotByID(attempt.QuizID, attempt.TestSnapshotID)
	if err != nil {
		return AssessmentAttemptLearnerView{}, err
	}
	learnerSnapshot, err := svc.learnerSnapshotFromAttempt(attempt, snapshot)
	if err != nil {
		return AssessmentAttemptLearnerView{}, err
	}
	return AssessmentAttemptLearnerView{
		Attempt:  attempt,
		Snapshot: learnerSnapshot,
	}, nil
}

func (svc *AssessmentAttemptService) learnerSnapshotFromAttempt(
	attempt models.AssessmentAttempt,
	snapshot models.TestSnapshot,
) (models.TestSnapshotLearnerView, error) {
	items, err := svc.attemptItemModel.ListByAttemptID(attempt.ID)
	if err != nil {
		return models.TestSnapshotLearnerView{}, err
	}
	if len(items) == 0 {
		// Backward-compatible fallback for pre-T04 attempts.
		return svc.snapshotModel.ToLearnerView(snapshot), nil
	}
	learnerItems := make([]models.TestSnapshotLearnerItem, 0, len(items))
	for _, item := range items {
		learnerItems = append(learnerItems, item.ToLearnerItem())
	}
	return models.TestSnapshotLearnerView{
		ID:                  snapshot.ID,
		QuizID:              snapshot.QuizID,
		Status:              snapshot.Status,
		SourceCollectionIDs: snapshot.SourceCollectionIDs,
		QuestionCount:       len(learnerItems),
		CreatedAt:           snapshot.CreatedAt,
		Items:               learnerItems,
	}, nil
}

func (svc *AssessmentAttemptService) ensureCreateEntitlement(quiz models.QuizSelfPacedMeta, req CreateAssessmentAttemptRequest) error {
	if quiz.AssessmentMode != QuizAssessmentModeSelfPaced {
		return models.ErrAssessmentAttemptNotSelfPaced
	}
	if req.IsEditorPreview {
		return nil
	}
	if quiz.Status != QuizStatusPublished {
		return models.ErrAssessmentAttemptQuizNotPublished
	}
	// Published SELF_PACED quizzes are attemptable by any authenticated learner.
	// Course enrolment gating is deferred until QUIZ_REFERENCE launch wiring (ADR-024 §6).
	_ = req.UserEmail
	return nil
}

// ResolveEditorPreview reports whether the authenticated user has write/share access.
func (svc *AssessmentAttemptService) ResolveEditorPreview(quizID uuid.UUID, user models.User) (bool, error) {
	isCreator, err := svc.sharedModel.CheckQuizCreatorExists(quizID.String(), user.ID)
	if err != nil {
		return false, err
	}
	if isCreator {
		return true, nil
	}
	permission, err := svc.sharedModel.GetPermissionByQuizAndUser(quizID.String(), user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return permission == constants.SharePermission || permission == constants.WritePermission, nil
}

// AttemptAnswersRepo exposes the answer model for later tasks without write APIs in T01.
func (svc *AssessmentAttemptService) AttemptAnswersRepo() *models.AttemptAnswerModel {
	return svc.answerModel
}

// AttemptModelForTest exposes the attempt model for deterministic UUID wiring in tests.
func (svc *AssessmentAttemptService) AttemptModelForTest() *models.AssessmentAttemptModel {
	return svc.attemptModel
}
