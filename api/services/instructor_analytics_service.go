package services

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	goqu "github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"go.uber.org/zap"
)

const (
	instructorDefaultLimit = 20
	instructorMaxLimit     = 100
)

var (
	ErrInstructorInvalidCursor = errors.New("invalid instructor analytics cursor")
	ErrInstructorInvalidRange  = errors.New("invalid instructor analytics date range")
	ErrInstructorQuizNotFound  = errors.New("quiz not found")
)

type InstructorAnalyticsService struct {
	db     *goqu.Database
	cache  *InstructorAnalyticsCache
	logger *zap.Logger
}

func NewInstructorAnalyticsService(db *goqu.Database, cache *InstructorAnalyticsCache, logger *zap.Logger) *InstructorAnalyticsService {
	if cache == nil {
		cache = NewInstructorAnalyticsCache(nil, logger)
	}
	return &InstructorAnalyticsService{db: db, cache: cache, logger: logger}
}

func (svc *InstructorAnalyticsService) Cache() *InstructorAnalyticsCache {
	return svc.cache
}

// Pure read-only release evaluator (zero state mutation, zero event emission).
func EvaluateResultReleaseState(
	policy string,
	resultsReleased bool,
	resultsScheduledAt sql.NullTime,
	resultsReleasedAt sql.NullTime,
	hasOverrideEvent bool,
	nowTime time.Time,
	loc *time.Location,
) string {
	if loc == nil {
		loc = time.UTC
	}
	policy = strings.ToUpper(strings.TrimSpace(policy))
	if policy == "" {
		policy = string(structs.ResultReleasePolicyImmediate)
	}

	switch policy {
	case string(structs.ResultReleasePolicyImmediate):
		return "IMMEDIATE_RELEASE"

	case string(structs.ResultReleasePolicyManual):
		if resultsReleased {
			return "COMPLETED_MANUAL"
		}
		return "PENDING_MANUAL"

	case string(structs.ResultReleasePolicyScheduled):
		if resultsReleased {
			if hasOverrideEvent {
				return "MANUALLY_OVERRIDDEN_SCHEDULED"
			}
			if resultsScheduledAt.Valid && resultsReleasedAt.Valid && resultsReleasedAt.Time.Before(resultsScheduledAt.Time) {
				return "MANUALLY_OVERRIDDEN_SCHEDULED"
			}
			return "COMPLETED_MANUAL"
		}
		if !resultsScheduledAt.Valid || resultsScheduledAt.Time.IsZero() {
			return "PENDING_MANUAL"
		}

		schedLocal := resultsScheduledAt.Time.In(loc)
		nowLocal := nowTime.In(loc)
		schedYear, schedMonth, schedDay := schedLocal.Date()
		nowYear, nowMonth, nowDay := nowLocal.Date()

		if schedYear == nowYear && schedMonth == nowMonth && schedDay == nowDay {
			return "SCHEDULED_FOR_TODAY"
		}
		if nowLocal.Before(schedLocal) {
			return "UPCOMING_SCHEDULED"
		}
		return "OVERDUE_SCHEDULED"

	default:
		if resultsReleased {
			return "COMPLETED_MANUAL"
		}
		return "PENDING_MANUAL"
	}
}

func FormatDisplayName(firstName, lastName, username string) string {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	username = strings.TrimSpace(username)

	if firstName != "" && lastName != "" {
		runes := []rune(firstName)
		return fmt.Sprintf("%s. %s", string(runes[0]), lastName)
	}
	if firstName != "" {
		return firstName
	}
	if username != "" {
		return username
	}
	return "Learner"
}

func EncodeCursor(parts ...interface{}) string {
	strParts := make([]string, len(parts))
	for i, p := range parts {
		switch v := p.(type) {
		case time.Time:
			strParts[i] = v.UTC().Format(time.RFC3339Nano)
		case *time.Time:
			if v != nil {
				strParts[i] = v.UTC().Format(time.RFC3339Nano)
			} else {
				strParts[i] = "NULL"
			}
		case nil:
			strParts[i] = "NULL"
		default:
			strParts[i] = fmt.Sprintf("%v", v)
		}
	}
	raw := strings.Join(strParts, "|")
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(cursor string) ([]string, error) {
	if strings.TrimSpace(cursor) == "" {
		return nil, ErrInstructorInvalidCursor
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, ErrInstructorInvalidCursor
	}
	return strings.Split(string(raw), "|"), nil
}

// -----------------------------------------------------------------------------
// 1. Portfolio Overview
// -----------------------------------------------------------------------------

func (svc *InstructorAnalyticsService) GetPortfolioOverview(
	ctx context.Context,
	instructorID string,
	preferredTimezone string,
) (structs.InstructorOverviewResponse, error) {
	tz := ResolveLearnerTimezone(preferredTimezone)
	hash := CanonicalEndpointHash("overview", tz)

	var cached structs.InstructorOverviewResponse
	if svc.cache.GetPortfolioJSON(ctx, instructorID, hash, &cached) {
		return cached, nil
	}

	loc, _ := time.LoadLocation(tz)
	now := time.Now().UTC()

	// 1. Count owned quizzes and release states
	type quizRow struct {
		ID                  uuid.UUID    `db:"id"`
		ResultReleasePolicy string       `db:"result_release_policy"`
		ResultsReleased     bool         `db:"results_released"`
		ResultsScheduledAt  sql.NullTime `db:"results_scheduled_at"`
		ResultsReleasedAt   sql.NullTime `db:"results_released_at"`
	}
	var quizzes []quizRow
	err := svc.db.From("quizzes").
		Select("id", "result_release_policy", "results_released", "results_scheduled_at", "results_released_at").
		Where(goqu.Ex{"creator_id": instructorID}).
		ScanStructs(&quizzes)
	if err != nil {
		return structs.InstructorOverviewResponse{}, err
	}

	totalQuizzes := len(quizzes)
	if totalQuizzes == 0 {
		resp := structs.InstructorOverviewResponse{
			ResolvedTimezone: tz,
			TotalQuizzes:     0,
		}
		svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
		return resp, nil
	}

	quizIDs := make([]uuid.UUID, totalQuizzes)
	for i, q := range quizzes {
		quizIDs[i] = q.ID
	}

	// Fetch override audit events for owned quizzes to refine release classification
	var overrideQuizIDs []uuid.UUID
	_ = svc.db.From("assessment_analytics_events").
		Select("quiz_id").
		Distinct().
		Where(goqu.Ex{
			"quiz_id":    quizIDs,
			"event_type": "RESULT_RELEASE_OVERRIDE_APPLIED",
		}).
		ScanVals(&overrideQuizIDs)
	overrideSet := map[uuid.UUID]bool{}
	for _, qid := range overrideQuizIDs {
		overrideSet[qid] = true
	}

	releasedQuizzes := 0
	pendingReleaseQuizzes := 0
	for _, q := range quizzes {
		class := EvaluateResultReleaseState(
			q.ResultReleasePolicy,
			q.ResultsReleased,
			q.ResultsScheduledAt,
			q.ResultsReleasedAt,
			overrideSet[q.ID],
			now,
			loc,
		)
		if class == "IMMEDIATE_RELEASE" || class == "COMPLETED_MANUAL" || class == "MANUALLY_OVERRIDDEN_SCHEDULED" {
			releasedQuizzes++
		} else {
			pendingReleaseQuizzes++
		}
	}

	// 2. Aggregate attempts across owned quizzes
	type attemptAgg struct {
		TotalAttempts            int             `db:"total_attempts"`
		CompletedAttempts        int             `db:"completed_attempts"`
		CompletedScoredCount     int             `db:"completed_scored_count"`
		SumAttemptPct            sql.NullFloat64 `db:"sum_attempt_pct"`
		SumDuration              sql.NullInt64   `db:"sum_duration"`
		CompletedDurationCount   int             `db:"completed_duration_count"`
		UniqueLearners           int             `db:"unique_learners"`
	}

	var agg attemptAgg
	_, err = svc.db.From(goqu.T("assessment_attempts").As("a")).
		Select(
			goqu.COUNT("a.id").As("total_attempts"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') THEN 1 END)").As("completed_attempts"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN 1 END)").As("completed_scored_count"),
			goqu.L("SUM(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN (a.total_score::double precision / NULLIF(a.max_score::double precision, 0)) * 100 END)").As("sum_attempt_pct"),
			goqu.L("SUM(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.time_taken_seconds IS NOT NULL AND a.time_taken_seconds >= 0 THEN a.time_taken_seconds END)").As("sum_duration"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.time_taken_seconds IS NOT NULL AND a.time_taken_seconds >= 0 THEN 1 END)").As("completed_duration_count"),
			goqu.COUNT(goqu.DISTINCT("a.user_id")).As("unique_learners"),
		).
		Where(goqu.Ex{"a.quiz_id": quizIDs}).
		ScanStruct(&agg)
	if err != nil {
		return structs.InstructorOverviewResponse{}, err
	}

	completionRate := CompletionRate(agg.TotalAttempts, agg.CompletedAttempts)

	var avgPct *float64
	if agg.CompletedScoredCount > 0 && agg.SumAttemptPct.Valid {
		val := math.Round((agg.SumAttemptPct.Float64/float64(agg.CompletedScoredCount))*100) / 100
		avgPct = &val
	}

	var avgDuration *int64
	if agg.CompletedDurationCount > 0 && agg.SumDuration.Valid {
		dur := agg.SumDuration.Int64 / int64(agg.CompletedDurationCount)
		avgDuration = &dur
	}

	resp := structs.InstructorOverviewResponse{
		ResolvedTimezone:             tz,
		TotalQuizzes:                 totalQuizzes,
		TotalAttempts:                agg.TotalAttempts,
		CompletedAttempts:            agg.CompletedAttempts,
		CompletedScoredAttemptsCount: agg.CompletedScoredCount,
		CompletionRate:               completionRate,
		AverageScorePercentage:       avgPct,
		UniqueLearners:               agg.UniqueLearners,
		AverageDurationSeconds:       avgDuration,
		ReleasedQuizzes:              releasedQuizzes,
		PendingReleaseQuizzes:        pendingReleaseQuizzes,
	}

	svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
	return resp, nil
}

// -----------------------------------------------------------------------------
// 2. Owned Quiz List
// -----------------------------------------------------------------------------

func (svc *InstructorAnalyticsService) GetOwnedQuizList(
	ctx context.Context,
	instructorID string,
	preferredTimezone string,
	cursor string,
	limit int,
	sortBy string,
	sortDir string,
) (structs.InstructorQuizListResponse, error) {
	tz := ResolveLearnerTimezone(preferredTimezone)
	if limit <= 0 || limit > instructorMaxLimit {
		limit = instructorDefaultLimit
	}
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortDir = strings.ToUpper(strings.TrimSpace(sortDir))
	if sortDir != "ASC" {
		sortDir = "DESC"
	}

	hash := CanonicalEndpointHash("quizzes", tz, cursor, strconv.Itoa(limit), sortBy, sortDir)
	var cached structs.InstructorQuizListResponse
	if svc.cache.GetPortfolioJSON(ctx, instructorID, hash, &cached) {
		return cached, nil
	}

	type quizRow struct {
		ID                     uuid.UUID       `db:"id"`
		Title                  string          `db:"title"`
		CategoryName           sql.NullString  `db:"category_name"`
		ResultReleasePolicy    string          `db:"result_release_policy"`
		ResultsReleased        bool            `db:"results_released"`
		CreatedAt              time.Time       `db:"created_at"`
		TotalAttempts          int             `db:"total_attempts"`
		CompletedAttempts      int             `db:"completed_attempts"`
		CompletedScoredCount   int             `db:"completed_scored_count"`
		SumAttemptPct          sql.NullFloat64 `db:"sum_attempt_pct"`
		UniqueLearners         int             `db:"unique_learners"`
	}

	query := svc.db.From(goqu.T("quizzes").As("q")).
		LeftJoin(goqu.T("quiz_categories").As("c"), goqu.On(goqu.I("c.id").Eq(goqu.I("q.category_id")))).
		LeftJoin(goqu.T("assessment_attempts").As("a"), goqu.On(goqu.I("a.quiz_id").Eq(goqu.I("q.id")))).
		Select(
			goqu.I("q.id"),
			goqu.I("q.title"),
			goqu.I("c.name").As("category_name"),
			goqu.I("q.result_release_policy"),
			goqu.I("q.results_released"),
			goqu.I("q.created_at"),
			goqu.COUNT("a.id").As("total_attempts"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') THEN 1 END)").As("completed_attempts"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN 1 END)").As("completed_scored_count"),
			goqu.L("SUM(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN (a.total_score::double precision / NULLIF(a.max_score::double precision, 0)) * 100 END)").As("sum_attempt_pct"),
			goqu.COUNT(goqu.DISTINCT("a.user_id")).As("unique_learners"),
		).
		Where(goqu.Ex{"q.creator_id": instructorID}).
		GroupBy(goqu.I("q.id"), goqu.I("q.title"), goqu.I("c.name"), goqu.I("q.result_release_policy"), goqu.I("q.results_released"), goqu.I("q.created_at"))

	// Apply cursor pagination
	if cursor != "" {
		parts, err := DecodeCursor(cursor)
		if err == nil && len(parts) >= 2 {
			cursorTime, _ := time.Parse(time.RFC3339Nano, parts[0])
			cursorID, _ := uuid.Parse(parts[len(parts)-1])
			if sortDir == "DESC" {
				query = query.Where(goqu.L("(q.created_at, q.id) < (?, ?)", cursorTime, cursorID))
			} else {
				query = query.Where(goqu.L("(q.created_at, q.id) > (?, ?)", cursorTime, cursorID))
			}
		}
	}

	if sortDir == "DESC" {
		query = query.Order(goqu.I("q.created_at").Desc(), goqu.I("q.id").Desc())
	} else {
		query = query.Order(goqu.I("q.created_at").Asc(), goqu.I("q.id").Asc())
	}

	var rows []quizRow
	err := query.Limit(uint(limit + 1)).ScanStructs(&rows)
	if err != nil {
		return structs.InstructorQuizListResponse{}, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]structs.InstructorQuizListItem, len(rows))
	for i, r := range rows {
		var catName *string
		if r.CategoryName.Valid && strings.TrimSpace(r.CategoryName.String) != "" {
			val := r.CategoryName.String
			catName = &val
		}

		var avgPct *float64
		if r.CompletedScoredCount > 0 && r.SumAttemptPct.Valid {
			val := math.Round((r.SumAttemptPct.Float64/float64(r.CompletedScoredCount))*100) / 100
			avgPct = &val
		}

		title := r.Title
		if decoded, err := unescapeString(title); err == nil {
			title = decoded
		}

		items[i] = structs.InstructorQuizListItem{
			QuizID:                 r.ID.String(),
			Title:                  title,
			CategoryName:           catName,
			TotalAttempts:          r.TotalAttempts,
			CompletedAttempts:      r.CompletedAttempts,
			AverageScorePercentage: avgPct,
			UniqueLearners:         r.UniqueLearners,
			ResultReleasePolicy:    r.ResultReleasePolicy,
			ResultsReleased:        r.ResultsReleased,
			CreatedAt:              r.CreatedAt.UTC().Format(time.RFC3339),
		}
	}

	nextCursor := ""
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor = EncodeCursor(last.CreatedAt, last.ID)
	}

	resp := structs.InstructorQuizListResponse{
		Quizzes:    items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}

	svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
	return resp, nil
}

// Helper to safely unescape titles
func unescapeString(s string) (string, error) {
	return strings.ReplaceAll(s, "%20", " "), nil
}

// -----------------------------------------------------------------------------
// 3. Learner Performance List
// -----------------------------------------------------------------------------

func (svc *InstructorAnalyticsService) GetLearnerList(
	ctx context.Context,
	instructorID string,
	preferredTimezone string,
	cursor string,
	limit int,
	search string,
	quizIDFilter string,
	statusFilter string,
	sortBy string,
	sortDir string,
) (structs.InstructorLearnerListResponse, error) {
	tz := ResolveLearnerTimezone(preferredTimezone)
	if limit <= 0 || limit > instructorMaxLimit {
		limit = instructorDefaultLimit
	}
	search = strings.TrimSpace(search)
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))
	if sortBy == "" {
		sortBy = "last_activity_at"
	}
	sortDir = strings.ToUpper(strings.TrimSpace(sortDir))
	if sortDir != "ASC" {
		sortDir = "DESC"
	}

	hash := CanonicalEndpointHash("learners", tz, cursor, strconv.Itoa(limit), search, quizIDFilter, statusFilter, sortBy, sortDir)
	var cached structs.InstructorLearnerListResponse
	if svc.cache.GetPortfolioJSON(ctx, instructorID, hash, &cached) {
		return cached, nil
	}

	// 1. Get owned quiz IDs
	var ownedQuizIDs []uuid.UUID
	err := svc.db.From("quizzes").Select("id").Where(goqu.Ex{"creator_id": instructorID}).ScanVals(&ownedQuizIDs)
	if err != nil {
		return structs.InstructorLearnerListResponse{}, err
	}
	if len(ownedQuizIDs) == 0 {
		resp := structs.InstructorLearnerListResponse{ResolvedTimezone: tz, Learners: []structs.InstructorLearnerItem{}}
		svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
		return resp, nil
	}

	if quizIDFilter != "" {
		parsedQID, err := uuid.Parse(quizIDFilter)
		if err == nil {
			isOwned := false
			for _, qid := range ownedQuizIDs {
				if qid == parsedQID {
					isOwned = true
					break
				}
			}
			if isOwned {
				ownedQuizIDs = []uuid.UUID{parsedQID}
			} else {
				ownedQuizIDs = nil
			}
		}
	}

	if len(ownedQuizIDs) == 0 {
		resp := structs.InstructorLearnerListResponse{ResolvedTimezone: tz, Learners: []structs.InstructorLearnerItem{}}
		svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
		return resp, nil
	}

	type learnerRow struct {
		UserID                 string          `db:"user_id"`
		FirstName              sql.NullString  `db:"first_name"`
		LastName               sql.NullString  `db:"last_name"`
		Username               sql.NullString  `db:"username"`
		ImgKey                 sql.NullString  `db:"img_key"`
		UniqueQuizzes          int             `db:"unique_quizzes"`
		TotalAttempts          int             `db:"total_attempts"`
		CompletedAttempts      int             `db:"completed_attempts"`
		CompletedScoredCount   int             `db:"completed_scored_count"`
		SumAttemptPct          sql.NullFloat64 `db:"sum_attempt_pct"`
		SumDuration            sql.NullInt64   `db:"sum_duration"`
		LastActivityAt         time.Time       `db:"last_activity_at"`
	}

	whereClause := goqu.Ex{"a.quiz_id": ownedQuizIDs}
	if statusFilter != "" {
		whereClause["a.status"] = statusFilter
	}

	query := svc.db.From(goqu.T("assessment_attempts").As("a")).
		LeftJoin(goqu.T("users").As("u"), goqu.On(goqu.I("u.id").Eq(goqu.I("a.user_id")))).
		Select(
			goqu.I("a.user_id"),
			goqu.I("u.first_name"),
			goqu.I("u.last_name"),
			goqu.I("u.username"),
			goqu.I("u.img_key"),
			goqu.COUNT(goqu.DISTINCT("a.quiz_id")).As("unique_quizzes"),
			goqu.COUNT("a.id").As("total_attempts"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') THEN 1 END)").As("completed_attempts"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN 1 END)").As("completed_scored_count"),
			goqu.L("SUM(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN (a.total_score / a.max_score) * 100 END)").As("sum_attempt_pct"),
			goqu.L("SUM(CASE WHEN a.time_taken_seconds IS NOT NULL AND a.time_taken_seconds >= 0 THEN a.time_taken_seconds END)").As("sum_duration"),
			goqu.MAX(goqu.L("COALESCE(a.submitted_at, a.created_at)")).As("last_activity_at"),
		).
		Where(whereClause).
		GroupBy(goqu.I("a.user_id"), goqu.I("u.first_name"), goqu.I("u.last_name"), goqu.I("u.username"), goqu.I("u.img_key"))

	if search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(goqu.Or(
			goqu.L("LOWER(u.first_name) LIKE ?", pattern),
			goqu.L("LOWER(u.last_name) LIKE ?", pattern),
			goqu.L("LOWER(u.username) LIKE ?", pattern),
			goqu.L("LOWER(a.user_id) LIKE ?", pattern),
		))
	}

	// Apply cursor pagination
	if cursor != "" {
		parts, err := DecodeCursor(cursor)
		if err == nil && len(parts) >= 2 {
			cursorTime, _ := time.Parse(time.RFC3339Nano, parts[0])
			cursorUID := parts[len(parts)-1]
			if sortDir == "DESC" {
				query = query.Having(goqu.L("(MAX(COALESCE(a.submitted_at, a.created_at)), a.user_id) < (?, ?)", cursorTime, cursorUID))
			} else {
				query = query.Having(goqu.L("(MAX(COALESCE(a.submitted_at, a.created_at)), a.user_id) > (?, ?)", cursorTime, cursorUID))
			}
		}
	}

	if sortDir == "DESC" {
		query = query.Order(goqu.L("MAX(COALESCE(a.submitted_at, a.created_at))").Desc(), goqu.I("a.user_id").Desc())
	} else {
		query = query.Order(goqu.L("MAX(COALESCE(a.submitted_at, a.created_at))").Asc(), goqu.I("a.user_id").Asc())
	}

	var rows []learnerRow
	err = query.Limit(uint(limit + 1)).ScanStructs(&rows)
	if err != nil {
		return structs.InstructorLearnerListResponse{}, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	userIDs := make([]string, len(rows))
	for i, r := range rows {
		userIDs[i] = r.UserID
	}

	// Calculate telemetry-based engaged question time per user using exact P8-T02 contract
	engagedTimeByUser := map[string]int64{}
	if len(userIDs) > 0 {
		type telemRow struct {
			UserID    string          `db:"user_id"`
			AttemptID uuid.UUID       `db:"attempt_id"`
			Metadata  json.RawMessage `db:"metadata"`
		}
		var telems []telemRow
		_ = svc.db.From("assessment_analytics_events").
			Select("user_id", "attempt_id", "metadata").
			Where(goqu.Ex{
				"user_id":    userIDs,
				"quiz_id":    ownedQuizIDs,
				"event_type": "QUESTION_TIME_SPENT",
			}).
			ScanStructs(&telems)

		// Group telemetry duration by (user_id, attempt_id)
		teleByAttempt := map[string]map[uuid.UUID]int64{}
		for _, t := range telems {
			dur := extractDurationSeconds(t.Metadata)
			if dur <= 0 {
				continue
			}
			if dur > engagedQuestionTimeCapSecs {
				dur = engagedQuestionTimeCapSecs
			}
			if teleByAttempt[t.UserID] == nil {
				teleByAttempt[t.UserID] = map[uuid.UUID]int64{}
			}
			teleByAttempt[t.UserID][t.AttemptID] += dur
		}

		// Fetch authoritative attempt durations for capping
		type attDurRow struct {
			UserID           string        `db:"user_id"`
			AttemptID        uuid.UUID     `db:"id"`
			TimeTakenSeconds sql.NullInt64 `db:"time_taken_seconds"`
		}
		var attDurs []attDurRow
		_ = svc.db.From("assessment_attempts").
			Select("user_id", "id", "time_taken_seconds").
			Where(goqu.Ex{"user_id": userIDs, "quiz_id": ownedQuizIDs}).
			ScanStructs(&attDurs)

		authDurByAttempt := map[uuid.UUID]int64{}
		for _, ad := range attDurs {
			var dur int64
			if ad.TimeTakenSeconds.Valid && ad.TimeTakenSeconds.Int64 > 0 {
				dur = ad.TimeTakenSeconds.Int64
			}
			authDurByAttempt[ad.AttemptID] = dur
		}

		for uid, attMap := range teleByAttempt {
			var userEngaged int64
			for attID, telemSum := range attMap {
				authDur := authDurByAttempt[attID]
				userEngaged += CapEngagedQuestionTime(telemSum, authDur)
			}
			engagedTimeByUser[uid] = userEngaged
		}
	}

	items := make([]structs.InstructorLearnerItem, len(rows))
	for i, r := range rows {
		fname := ""
		if r.FirstName.Valid {
			fname = r.FirstName.String
		}
		lname := ""
		if r.LastName.Valid {
			lname = r.LastName.String
		}
		uname := ""
		if r.Username.Valid {
			uname = r.Username.String
		}

		displayName := FormatDisplayName(fname, lname, uname)

		var avatarURL *string
		if r.ImgKey.Valid && strings.TrimSpace(r.ImgKey.String) != "" {
			url := r.ImgKey.String
			avatarURL = &url
		}

		compRate := CompletionRate(r.TotalAttempts, r.CompletedAttempts)

		var avgPct *float64
		if r.CompletedScoredCount > 0 && r.SumAttemptPct.Valid {
			val := math.Round((r.SumAttemptPct.Float64/float64(r.CompletedScoredCount))*100) / 100
			avgPct = &val
		}

		var durationSecs int64
		if r.SumDuration.Valid && r.SumDuration.Int64 > 0 {
			durationSecs = r.SumDuration.Int64
		}

		items[i] = structs.InstructorLearnerItem{
			LearnerID:                  r.UserID,
			DisplayName:                displayName,
			AvatarURL:                  avatarURL,
			UniqueQuizzesAttempted:     r.UniqueQuizzes,
			TotalAttempts:              r.TotalAttempts,
			CompletedAttempts:          r.CompletedAttempts,
			CompletionRate:             compRate,
			AveragePercentage:          avgPct,
			AssessmentDurationSeconds:  durationSecs,
			EngagedQuestionTimeSeconds: engagedTimeByUser[r.UserID],
			LastActivityAt:             r.LastActivityAt.UTC().Format(time.RFC3339),
		}
	}

	nextCursor := ""
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor = EncodeCursor(last.LastActivityAt, last.UserID)
	}

	resp := structs.InstructorLearnerListResponse{
		ResolvedTimezone: tz,
		Learners:         items,
		NextCursor:       nextCursor,
		HasMore:          hasMore,
	}

	svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
	return resp, nil
}

// -----------------------------------------------------------------------------
// 4. Release Monitoring
// -----------------------------------------------------------------------------

func (svc *InstructorAnalyticsService) GetReleaseMonitoring(
	ctx context.Context,
	instructorID string,
	preferredTimezone string,
	cursor string,
	limit int,
	classificationFilter string,
) (structs.InstructorReleaseMonitoringResponse, error) {
	tz := ResolveLearnerTimezone(preferredTimezone)
	if limit <= 0 || limit > instructorMaxLimit {
		limit = instructorDefaultLimit
	}
	classificationFilter = strings.ToUpper(strings.TrimSpace(classificationFilter))

	hash := CanonicalEndpointHash("releases", tz, cursor, strconv.Itoa(limit), classificationFilter)
	var cached structs.InstructorReleaseMonitoringResponse
	if svc.cache.GetPortfolioJSON(ctx, instructorID, hash, &cached) {
		return cached, nil
	}

	loc, _ := time.LoadLocation(tz)
	now := time.Now().UTC()

	// Fetch all owned quizzes with completed attempt counts
	type rawReleaseRow struct {
		ID                     uuid.UUID      `db:"id"`
		Title                  string         `db:"title"`
		ResultReleasePolicy    string         `db:"result_release_policy"`
		ResultsReleased        bool           `db:"results_released"`
		ResultsScheduledAt     sql.NullTime   `db:"results_scheduled_at"`
		ResultsReleasedAt      sql.NullTime   `db:"results_released_at"`
		CategoryName           sql.NullString `db:"category_name"`
		CreatedAt              time.Time      `db:"created_at"`
		CompletedAttemptsCount int            `db:"completed_attempts_count"`
	}

	var rows []rawReleaseRow
	err := svc.db.From(goqu.T("quizzes").As("q")).
		LeftJoin(goqu.T("quiz_categories").As("c"), goqu.On(goqu.I("c.id").Eq(goqu.I("q.category_id")))).
		LeftJoin(goqu.T("assessment_attempts").As("a"), goqu.On(goqu.I("a.quiz_id").Eq(goqu.I("q.id")))).
		Select(
			goqu.I("q.id"),
			goqu.I("q.title"),
			goqu.I("q.result_release_policy"),
			goqu.I("q.results_released"),
			goqu.I("q.results_scheduled_at"),
			goqu.I("q.results_released_at"),
			goqu.I("c.name").As("category_name"),
			goqu.I("q.created_at"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') THEN 1 END)").As("completed_attempts_count"),
		).
		Where(goqu.Ex{"q.creator_id": instructorID}).
		GroupBy(goqu.I("q.id"), goqu.I("q.title"), goqu.I("q.result_release_policy"), goqu.I("q.results_released"), goqu.I("q.results_scheduled_at"), goqu.I("q.results_released_at"), goqu.I("c.name"), goqu.I("q.created_at")).
		Order(goqu.I("q.created_at").Desc(), goqu.I("q.id").Desc()).
		ScanStructs(&rows)
	if err != nil {
		return structs.InstructorReleaseMonitoringResponse{}, err
	}

	if len(rows) == 0 {
		resp := structs.InstructorReleaseMonitoringResponse{
			ResolvedTimezone: tz,
			Summary:          map[string]int{},
			Quizzes:          []structs.InstructorReleaseItem{},
		}
		svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
		return resp, nil
	}

	quizIDs := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		quizIDs[i] = r.ID
	}

	var overrideQuizIDs []uuid.UUID
	_ = svc.db.From("assessment_analytics_events").
		Select("quiz_id").
		Distinct().
		Where(goqu.Ex{
			"quiz_id":    quizIDs,
			"event_type": "RESULT_RELEASE_OVERRIDE_APPLIED",
		}).
		ScanVals(&overrideQuizIDs)
	overrideSet := map[uuid.UUID]bool{}
	for _, qid := range overrideQuizIDs {
		overrideSet[qid] = true
	}

	summaryCounts := map[string]int{
		"IMMEDIATE_RELEASE":             0,
		"PENDING_MANUAL":                0,
		"COMPLETED_MANUAL":              0,
		"UPCOMING_SCHEDULED":            0,
		"OVERDUE_SCHEDULED":             0,
		"SCHEDULED_FOR_TODAY":           0,
		"MANUALLY_OVERRIDDEN_SCHEDULED": 0,
	}

	allEvaluated := make([]structs.InstructorReleaseItem, 0, len(rows))
	for _, r := range rows {
		class := EvaluateResultReleaseState(
			r.ResultReleasePolicy,
			r.ResultsReleased,
			r.ResultsScheduledAt,
			r.ResultsReleasedAt,
			overrideSet[r.ID],
			now,
			loc,
		)

		summaryCounts[class]++

		if classificationFilter != "" && class != classificationFilter {
			continue
		}

		cat := "Uncategorised"
		if r.CategoryName.Valid && strings.TrimSpace(r.CategoryName.String) != "" {
			cat = r.CategoryName.String
		}

		var schedStr *string
		if r.ResultsScheduledAt.Valid && !r.ResultsScheduledAt.Time.IsZero() {
			val := r.ResultsScheduledAt.Time.UTC().Format(time.RFC3339)
			schedStr = &val
		}

		var relStr *string
		if r.ResultsReleasedAt.Valid && !r.ResultsReleasedAt.Time.IsZero() {
			val := r.ResultsReleasedAt.Time.UTC().Format(time.RFC3339)
			relStr = &val
		}

		title := r.Title
		if decoded, err := unescapeString(title); err == nil {
			title = decoded
		}

		allEvaluated = append(allEvaluated, structs.InstructorReleaseItem{
			QuizID:                 r.ID.String(),
			Title:                  title,
			ResultReleasePolicy:    r.ResultReleasePolicy,
			ResultsReleased:        r.ResultsReleased,
			ResultsScheduledAt:     schedStr,
			ResultsReleasedAt:      relStr,
			Category:               cat,
			Classification:         class,
			CompletedAttemptsCount: r.CompletedAttemptsCount,
		})
	}

	// Apply cursor pagination in-memory on evaluated items
	startIndex := 0
	if cursor != "" {
		parts, err := DecodeCursor(cursor)
		if err == nil && len(parts) >= 1 {
			targetID := parts[len(parts)-1]
			for idx, item := range allEvaluated {
				if item.QuizID == targetID {
					startIndex = idx + 1
					break
				}
			}
		}
	}

	endIndex := startIndex + limit
	hasMore := false
	if endIndex < len(allEvaluated) {
		hasMore = true
	} else {
		endIndex = len(allEvaluated)
	}

	var pagedItems []structs.InstructorReleaseItem
	if startIndex < len(allEvaluated) {
		pagedItems = allEvaluated[startIndex:endIndex]
	} else {
		pagedItems = []structs.InstructorReleaseItem{}
	}

	nextCursor := ""
	if hasMore && len(pagedItems) > 0 {
		last := pagedItems[len(pagedItems)-1]
		nextCursor = EncodeCursor(last.QuizID)
	}

	resp := structs.InstructorReleaseMonitoringResponse{
		ResolvedTimezone: tz,
		Summary:          summaryCounts,
		Quizzes:          pagedItems,
		NextCursor:       nextCursor,
		HasMore:          hasMore,
	}

	svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
	return resp, nil
}

// -----------------------------------------------------------------------------
// 5. Instructor Activity Timeline
// -----------------------------------------------------------------------------

func (svc *InstructorAnalyticsService) GetTimeline(
	ctx context.Context,
	instructorID string,
	preferredTimezone string,
	cursor string,
	limit int,
) (structs.InstructorTimelineResponse, error) {
	tz := ResolveLearnerTimezone(preferredTimezone)
	if limit <= 0 || limit > instructorMaxLimit {
		limit = instructorDefaultLimit
	}

	hash := CanonicalEndpointHash("timeline", tz, cursor, strconv.Itoa(limit))
	var cached structs.InstructorTimelineResponse
	if svc.cache.GetPortfolioJSON(ctx, instructorID, hash, &cached) {
		return cached, nil
	}

	// 1. Get owned quiz IDs
	type quizMeta struct {
		ID    uuid.UUID `db:"id"`
		Title string    `db:"title"`
	}
	var ownedQuizzes []quizMeta
	err := svc.db.From("quizzes").Select("id", "title").Where(goqu.Ex{"creator_id": instructorID}).ScanStructs(&ownedQuizzes)
	if err != nil {
		return structs.InstructorTimelineResponse{}, err
	}
	if len(ownedQuizzes) == 0 {
		resp := structs.InstructorTimelineResponse{ResolvedTimezone: tz, Events: []structs.InstructorTimelineEvent{}}
		svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
		return resp, nil
	}

	quizMap := map[uuid.UUID]string{}
	quizIDs := make([]uuid.UUID, len(ownedQuizzes))
	for i, q := range ownedQuizzes {
		quizIDs[i] = q.ID
		title := q.Title
		if decoded, err := unescapeString(title); err == nil {
			title = decoded
		}
		quizMap[q.ID] = title
	}

	allowedEvents := []string{
		"ATTEMPT_STARTED",
		"ATTEMPT_SUBMITTED",
		"ATTEMPT_AUTO_SUBMITTED",
		"RESULT_VIEWED",
		"RESULT_RELEASE_OVERRIDE_APPLIED",
		"RESULT_RELEASE_SCHEDULED_EFFECTIVE",
	}

	type eventRow struct {
		ID          uuid.UUID `db:"id"`
		QuizID      uuid.UUID `db:"quiz_id"`
		EventType   string    `db:"event_type"`
		EventSource string    `db:"event_source"`
		OccurredAt  time.Time `db:"occurred_at"`
		CreatedAt   time.Time `db:"created_at"`
	}

	query := svc.db.From("assessment_analytics_events").
		Select("id", "quiz_id", "event_type", "event_source", "occurred_at", "created_at").
		Where(goqu.Ex{
			"quiz_id":    quizIDs,
			"event_type": allowedEvents,
		})

	if cursor != "" {
		parts, err := DecodeCursor(cursor)
		if err == nil && len(parts) >= 3 {
			cursorOccurred, _ := time.Parse(time.RFC3339Nano, parts[0])
			cursorCreated, _ := time.Parse(time.RFC3339Nano, parts[1])
			cursorID, _ := uuid.Parse(parts[2])
			query = query.Where(goqu.L("(occurred_at, created_at, id) < (?, ?, ?)", cursorOccurred, cursorCreated, cursorID))
		}
	}

	var rows []eventRow
	err = query.Order(goqu.I("occurred_at").Desc(), goqu.I("created_at").Desc(), goqu.I("id").Desc()).
		Limit(uint(limit + 1)).
		ScanStructs(&rows)
	if err != nil {
		return structs.InstructorTimelineResponse{}, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	events := make([]structs.InstructorTimelineEvent, len(rows))
	for i, r := range rows {
		qTitle := quizMap[r.QuizID]
		summary := fmt.Sprintf("%s event recorded for quiz '%s'", r.EventType, qTitle)
		switch r.EventType {
		case "ATTEMPT_STARTED":
			summary = fmt.Sprintf("A learner started an attempt on '%s'", qTitle)
		case "ATTEMPT_SUBMITTED":
			summary = fmt.Sprintf("An attempt was manually submitted on '%s'", qTitle)
		case "ATTEMPT_AUTO_SUBMITTED":
			summary = fmt.Sprintf("An attempt was auto-submitted on '%s'", qTitle)
		case "RESULT_VIEWED":
			summary = fmt.Sprintf("A learner viewed results for '%s'", qTitle)
		case "RESULT_RELEASE_OVERRIDE_APPLIED":
			summary = fmt.Sprintf("Manual result release override applied for '%s'", qTitle)
		case "RESULT_RELEASE_SCHEDULED_EFFECTIVE":
			summary = fmt.Sprintf("Scheduled result release became effective for '%s'", qTitle)
		}

		events[i] = structs.InstructorTimelineEvent{
			ID:          r.ID.String(),
			QuizID:      r.QuizID.String(),
			QuizTitle:   qTitle,
			EventType:   r.EventType,
			EventSource: r.EventSource,
			OccurredAt:  r.OccurredAt.UTC().Format(time.RFC3339),
			Summary:     summary,
		}
	}

	nextCursor := ""
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor = EncodeCursor(last.OccurredAt, last.CreatedAt, last.ID)
	}

	resp := structs.InstructorTimelineResponse{
		ResolvedTimezone: tz,
		Events:           events,
		NextCursor:       nextCursor,
		HasMore:          hasMore,
	}

	svc.cache.SetPortfolioJSON(ctx, instructorID, hash, resp)
	return resp, nil
}

// -----------------------------------------------------------------------------
// 6. Per-Quiz Cohort Summary
// -----------------------------------------------------------------------------

func (svc *InstructorAnalyticsService) GetQuizCohortSummary(
	ctx context.Context,
	quizID uuid.UUID,
	preferredTimezone string,
) (structs.InstructorQuizSummaryResponse, error) {
	tz := ResolveLearnerTimezone(preferredTimezone)
	hash := CanonicalEndpointHash("summary", tz)

	var cached structs.InstructorQuizSummaryResponse
	if svc.cache.GetQuizJSON(ctx, quizID.String(), hash, &cached) {
		return cached, nil
	}

	// Fetch quiz metadata
	type quizMeta struct {
		ID                  uuid.UUID `db:"id"`
		Title               string    `db:"title"`
		ResultReleasePolicy string    `db:"result_release_policy"`
		ResultsReleased     bool      `db:"results_released"`
	}
	var q quizMeta
	found, err := svc.db.From("quizzes").
		Select("id", "title", "result_release_policy", "results_released").
		Where(goqu.Ex{"id": quizID}).
		ScanStruct(&q)
	if err != nil {
		return structs.InstructorQuizSummaryResponse{}, err
	}
	if !found {
		return structs.InstructorQuizSummaryResponse{}, ErrInstructorQuizNotFound
	}

	title := q.Title
	if decoded, err := unescapeString(title); err == nil {
		title = decoded
	}

	// Fetch question count
	var totalQuestions int
	_, _ = svc.db.From("quiz_questions").Where(goqu.Ex{"quiz_id": quizID}).Select(goqu.COUNT("id")).ScanVal(&totalQuestions)

	// Aggregate attempt metrics
	type attemptAgg struct {
		TotalAttempts          int             `db:"total_attempts"`
		CompletedAttempts      int             `db:"completed_attempts"`
		InProgressAttempts     int             `db:"in_progress_attempts"`
		AbandonedAttempts      int             `db:"abandoned_attempts"`
		CompletedScoredCount   int             `db:"completed_scored_count"`
		SumAttemptPct          sql.NullFloat64 `db:"sum_attempt_pct"`
		MaxAttemptPct          sql.NullFloat64 `db:"max_attempt_pct"`
		MinAttemptPct          sql.NullFloat64 `db:"min_attempt_pct"`
		SumDuration            sql.NullInt64   `db:"sum_duration"`
		CompletedDurationCount int             `db:"completed_duration_count"`
		UniqueLearners         int             `db:"unique_learners"`
	}

	var agg attemptAgg
	_, err = svc.db.From(goqu.T("assessment_attempts").As("a")).
		Select(
			goqu.COUNT("a.id").As("total_attempts"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') THEN 1 END)").As("completed_attempts"),
			goqu.L("COUNT(CASE WHEN a.status = 'IN_PROGRESS' THEN 1 END)").As("in_progress_attempts"),
			goqu.L("COUNT(CASE WHEN a.status = 'ABANDONED' THEN 1 END)").As("abandoned_attempts"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN 1 END)").As("completed_scored_count"),
			goqu.L("SUM(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN (a.total_score / a.max_score) * 100 END)").As("sum_attempt_pct"),
			goqu.L("MAX(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN (a.total_score / a.max_score) * 100 END)").As("max_attempt_pct"),
			goqu.L("MIN(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.total_score IS NOT NULL AND a.max_score IS NOT NULL AND a.max_score > 0 THEN (a.total_score / a.max_score) * 100 END)").As("min_attempt_pct"),
			goqu.L("SUM(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.time_taken_seconds IS NOT NULL AND a.time_taken_seconds >= 0 THEN a.time_taken_seconds END)").As("sum_duration"),
			goqu.L("COUNT(CASE WHEN a.status IN ('SUBMITTED', 'AUTO_SUBMITTED') AND a.time_taken_seconds IS NOT NULL AND a.time_taken_seconds >= 0 THEN 1 END)").As("completed_duration_count"),
			goqu.COUNT(goqu.DISTINCT("a.user_id")).As("unique_learners"),
		).
		Where(goqu.Ex{"a.quiz_id": quizID}).
		ScanStruct(&agg)
	if err != nil {
		return structs.InstructorQuizSummaryResponse{}, err
	}

	compRate := CompletionRate(agg.TotalAttempts, agg.CompletedAttempts)

	var avgPct, maxPct, minPct *float64
	if agg.CompletedScoredCount > 0 && agg.SumAttemptPct.Valid {
		val := math.Round((agg.SumAttemptPct.Float64/float64(agg.CompletedScoredCount))*100) / 100
		avgPct = &val
	}
	if agg.MaxAttemptPct.Valid {
		val := math.Round(agg.MaxAttemptPct.Float64*100) / 100
		maxPct = &val
	}
	if agg.MinAttemptPct.Valid {
		val := math.Round(agg.MinAttemptPct.Float64*100) / 100
		minPct = &val
	}

	var avgDuration *int64
	if agg.CompletedDurationCount > 0 && agg.SumDuration.Valid {
		dur := agg.SumDuration.Int64 / int64(agg.CompletedDurationCount)
		avgDuration = &dur
	}

	resp := structs.InstructorQuizSummaryResponse{
		QuizID:                 quizID.String(),
		Title:                  title,
		TotalAttempts:          agg.TotalAttempts,
		CompletedAttempts:      agg.CompletedAttempts,
		InProgressAttempts:     agg.InProgressAttempts,
		AbandonedAttempts:      agg.AbandonedAttempts,
		CompletionRate:         compRate,
		AverageScorePercentage: avgPct,
		HighestScorePercentage: maxPct,
		LowestScorePercentage:  minPct,
		AverageDurationSeconds: avgDuration,
		UniqueLearners:         agg.UniqueLearners,
		ResultReleasePolicy:    q.ResultReleasePolicy,
		ResultsReleased:        q.ResultsReleased,
		TotalQuestions:         totalQuestions,
	}

	svc.cache.SetQuizJSON(ctx, quizID.String(), hash, resp)
	return resp, nil
}

// -----------------------------------------------------------------------------
// 7. Per-Quiz Attempt List
// -----------------------------------------------------------------------------

func (svc *InstructorAnalyticsService) GetQuizAttemptList(
	ctx context.Context,
	quizID uuid.UUID,
	cursor string,
	limit int,
	statusFilter string,
) (structs.InstructorAttemptListResponse, error) {
	if limit <= 0 || limit > instructorMaxLimit {
		limit = instructorDefaultLimit
	}
	statusFilter = strings.TrimSpace(statusFilter)

	hash := CanonicalEndpointHash("attempts", cursor, strconv.Itoa(limit), statusFilter)
	var cached structs.InstructorAttemptListResponse
	if svc.cache.GetQuizJSON(ctx, quizID.String(), hash, &cached) {
		return cached, nil
	}

	type attemptRow struct {
		ID               uuid.UUID       `db:"id"`
		UserID           string          `db:"user_id"`
		FirstName        sql.NullString  `db:"first_name"`
		LastName         sql.NullString  `db:"last_name"`
		Username         sql.NullString  `db:"username"`
		ImgKey           sql.NullString  `db:"img_key"`
		AttemptNumber    int             `db:"attempt_number"`
		Status           string          `db:"status"`
		TotalScore       sql.NullFloat64 `db:"total_score"`
		MaxScore         sql.NullFloat64 `db:"max_score"`
		TimeTakenSeconds sql.NullInt64   `db:"time_taken_seconds"`
		StartedAt        time.Time       `db:"started_at"`
		SubmittedAt      sql.NullTime    `db:"submitted_at"`
		CreatedAt        time.Time       `db:"created_at"`
	}

	whereClause := goqu.Ex{"a.quiz_id": quizID}
	if statusFilter != "" {
		whereClause["a.status"] = statusFilter
	}

	query := svc.db.From(goqu.T("assessment_attempts").As("a")).
		LeftJoin(goqu.T("users").As("u"), goqu.On(goqu.I("u.id").Eq(goqu.I("a.user_id")))).
		Select(
			goqu.I("a.id"),
			goqu.I("a.user_id"),
			goqu.I("u.first_name"),
			goqu.I("u.last_name"),
			goqu.I("u.username"),
			goqu.I("u.img_key"),
			goqu.I("a.attempt_number"),
			goqu.I("a.status"),
			goqu.I("a.total_score"),
			goqu.I("a.max_score"),
			goqu.I("a.time_taken_seconds"),
			goqu.I("a.started_at"),
			goqu.I("a.submitted_at"),
			goqu.I("a.created_at"),
		).
		Where(whereClause)

	if cursor != "" {
		parts, err := DecodeCursor(cursor)
		if err == nil && len(parts) >= 2 {
			cursorTime, _ := time.Parse(time.RFC3339Nano, parts[0])
			cursorID, _ := uuid.Parse(parts[len(parts)-1])
			query = query.Where(goqu.L("(COALESCE(a.submitted_at, a.created_at), a.id) < (?, ?)", cursorTime, cursorID))
		}
	}

	var rows []attemptRow
	err := query.Order(goqu.L("COALESCE(a.submitted_at, a.created_at)").Desc(), goqu.I("a.id").Desc()).
		Limit(uint(limit + 1)).
		ScanStructs(&rows)
	if err != nil {
		return structs.InstructorAttemptListResponse{}, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]structs.InstructorAttemptListItem, len(rows))
	for i, r := range rows {
		fname := ""
		if r.FirstName.Valid {
			fname = r.FirstName.String
		}
		lname := ""
		if r.LastName.Valid {
			lname = r.LastName.String
		}
		uname := ""
		if r.Username.Valid {
			uname = r.Username.String
		}

		displayName := FormatDisplayName(fname, lname, uname)

		var avatarURL *string
		if r.ImgKey.Valid && strings.TrimSpace(r.ImgKey.String) != "" {
			url := r.ImgKey.String
			avatarURL = &url
		}

		var totScore, mScore, pct *float64
		if r.TotalScore.Valid {
			v := r.TotalScore.Float64
			totScore = &v
		}
		if r.MaxScore.Valid {
			v := r.MaxScore.Float64
			mScore = &v
		}
		if r.TotalScore.Valid && r.MaxScore.Valid && r.MaxScore.Float64 > 0 {
			v := math.Round((r.TotalScore.Float64/r.MaxScore.Float64)*10000) / 100
			pct = &v
		}

		var durationSecs *int64
		if r.TimeTakenSeconds.Valid {
			v := r.TimeTakenSeconds.Int64
			durationSecs = &v
		}

		var subStr *string
		if r.SubmittedAt.Valid && !r.SubmittedAt.Time.IsZero() {
			v := r.SubmittedAt.Time.UTC().Format(time.RFC3339)
			subStr = &v
		}

		items[i] = structs.InstructorAttemptListItem{
			AttemptID:        r.ID.String(),
			LearnerID:        r.UserID,
			DisplayName:      displayName,
			AvatarURL:        avatarURL,
			AttemptNumber:    r.AttemptNumber,
			Status:           r.Status,
			TotalScore:       totScore,
			MaxScore:         mScore,
			Percentage:       pct,
			TimeTakenSeconds: durationSecs,
			StartedAt:        r.StartedAt.UTC().Format(time.RFC3339),
			SubmittedAt:      subStr,
		}
	}

	nextCursor := ""
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		actTime := last.CreatedAt
		if last.SubmittedAt.Valid {
			actTime = last.SubmittedAt.Time
		}
		nextCursor = EncodeCursor(actTime, last.ID)
	}

	resp := structs.InstructorAttemptListResponse{
		QuizID:     quizID.String(),
		Attempts:   items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}

	svc.cache.SetQuizJSON(ctx, quizID.String(), hash, resp)
	return resp, nil
}

// -----------------------------------------------------------------------------
// 8. Per-Question Quality Metrics
// -----------------------------------------------------------------------------

func (svc *InstructorAnalyticsService) GetQuestionMetrics(
	ctx context.Context,
	quizID uuid.UUID,
	cursor string,
	limit int,
) (structs.InstructorQuestionMetricsResponse, error) {
	if limit <= 0 || limit > instructorMaxLimit {
		limit = instructorDefaultLimit
	}

	hash := CanonicalEndpointHash("questions", cursor, strconv.Itoa(limit))
	var cached structs.InstructorQuestionMetricsResponse
	if svc.cache.GetQuizJSON(ctx, quizID.String(), hash, &cached) {
		return cached, nil
	}

	// 1. Fetch paginated questions for quiz
	type qRow struct {
		ID          uuid.UUID `db:"id"`
		Question    string    `db:"question"`
		Type        int       `db:"type"`
		OrderNo     int       `db:"order_no"`
	}

	query := svc.db.From(goqu.T("quiz_questions").As("qq")).
		InnerJoin(goqu.T("questions").As("q"), goqu.On(goqu.I("q.id").Eq(goqu.I("qq.question_id")))).
		Select(
			goqu.I("q.id"),
			goqu.I("q.question"),
			goqu.I("q.type"),
			goqu.I("qq.order_no"),
		).
		Where(goqu.Ex{"qq.quiz_id": quizID})

	if cursor != "" {
		parts, err := DecodeCursor(cursor)
		if err == nil && len(parts) >= 2 {
			cursorOrder, _ := strconv.Atoi(parts[0])
			cursorID, _ := uuid.Parse(parts[1])
			query = query.Where(goqu.L("(qq.order_no, q.id) > (?, ?)", cursorOrder, cursorID))
		}
	}

	var questions []qRow
	err := query.Order(goqu.I("qq.order_no").Asc(), goqu.I("q.id").Asc()).
		Limit(uint(limit + 1)).
		ScanStructs(&questions)
	if err != nil {
		return structs.InstructorQuestionMetricsResponse{}, err
	}

	hasMore := len(questions) > limit
	if hasMore {
		questions = questions[:limit]
	}

	if len(questions) == 0 {
		resp := structs.InstructorQuestionMetricsResponse{
			QuizID:    quizID.String(),
			Questions: []structs.InstructorQuestionMetricsItem{},
		}
		svc.cache.SetQuizJSON(ctx, quizID.String(), hash, resp)
		return resp, nil
	}

	questionIDs := make([]uuid.UUID, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID
	}

	// Fetch count of completed attempts for this quiz
	var eligibleCompletedAttempts int
	_, _ = svc.db.From("assessment_attempts").
		Where(goqu.Ex{
			"quiz_id": quizID,
			"status":  []string{"SUBMITTED", "AUTO_SUBMITTED"},
		}).
		Select(goqu.COUNT("id")).
		ScanVal(&eligibleCompletedAttempts)

	// Fetch attempt percentage ranks for discrimination index calculation
	type attemptRank struct {
		AttemptID uuid.UUID `db:"id"`
		Pct       float64   `db:"pct"`
	}
	var rankedAttempts []attemptRank
	_ = svc.db.From("assessment_attempts").
		Select(
			"id",
			goqu.L("(total_score / NULLIF(max_score, 0)) * 100").As("pct"),
		).
		Where(goqu.Ex{
			"quiz_id": quizID,
			"status":  []string{"SUBMITTED", "AUTO_SUBMITTED"},
		}).
		Where(goqu.L("total_score IS NOT NULL AND max_score IS NOT NULL AND max_score > 0")).
		Order(goqu.L("(total_score / NULLIF(max_score, 0)) * 100").Desc(), goqu.I("submitted_at").Asc(), goqu.I("id").Asc()).
		ScanStructs(&rankedAttempts)

	topGroupSet := map[uuid.UUID]bool{}
	bottomGroupSet := map[uuid.UUID]bool{}

	totalEligible := len(rankedAttempts)
	if totalEligible >= 10 {
		groupSize := int(math.Floor(float64(totalEligible) * 0.27))
		if groupSize < 1 {
			groupSize = 1
		}
		for i := 0; i < groupSize; i++ {
			topGroupSet[rankedAttempts[i].AttemptID] = true
		}
		for i := totalEligible - groupSize; i < totalEligible; i++ {
			bottomGroupSet[rankedAttempts[i].AttemptID] = true
		}
	}

	// Fetch attempt_answers for these questions across completed attempts
	type ansRow struct {
		QuestionID       uuid.UUID       `db:"question_id"`
		AttemptID        uuid.UUID       `db:"attempt_id"`
		IsCorrect        sql.NullBool    `db:"is_correct"`
		TimeTakenSeconds sql.NullInt64   `db:"time_taken_seconds"`
		SelectedOptions  json.RawMessage `db:"selected_options"`
	}

	var answers []ansRow
	err = svc.db.From(goqu.T("attempt_answers").As("aa")).
		InnerJoin(goqu.T("assessment_attempts").As("a"), goqu.On(goqu.I("a.id").Eq(goqu.I("aa.attempt_id")))).
		Select(
			goqu.I("aa.question_id"),
			goqu.I("aa.attempt_id"),
			goqu.I("aa.is_correct"),
			goqu.I("aa.time_taken_seconds"),
			goqu.I("aa.selected_options"),
		).
		Where(goqu.Ex{
			"aa.question_id": questionIDs,
			"a.status":        []string{"SUBMITTED", "AUTO_SUBMITTED"},
		}).
		ScanStructs(&answers)
	if err != nil {
		return structs.InstructorQuestionMetricsResponse{}, err
	}

	// Group answers by question ID
	type qStats struct {
		totalAnswered int
		correctCount  int
		incorrectCount int
		topCorrect    int
		topTotal      int
		bottomCorrect int
		bottomTotal   int
		sumTime       int64
		timeCount     int
		distMap       map[string]int
	}

	statsMap := map[uuid.UUID]*qStats{}
	for _, qid := range questionIDs {
		statsMap[qid] = &qStats{distMap: map[string]int{}}
	}

	for _, a := range answers {
		st := statsMap[a.QuestionID]
		if st == nil {
			continue
		}
		st.totalAnswered++
		if a.IsCorrect.Valid {
			if a.IsCorrect.Bool {
				st.correctCount++
			} else {
				st.incorrectCount++
			}
		}
		if a.TimeTakenSeconds.Valid && a.TimeTakenSeconds.Int64 >= 0 {
			st.sumTime += a.TimeTakenSeconds.Int64
			st.timeCount++
		}

		if topGroupSet[a.AttemptID] {
			st.topTotal++
			if a.IsCorrect.Valid && a.IsCorrect.Bool {
				st.topCorrect++
			}
		}
		if bottomGroupSet[a.AttemptID] {
			st.bottomTotal++
			if a.IsCorrect.Valid && a.IsCorrect.Bool {
				st.bottomCorrect++
			}
		}

		// Option distribution parsing (objective questions only)
		if len(a.SelectedOptions) > 0 {
			var singleOpt int
			if err := json.Unmarshal(a.SelectedOptions, &singleOpt); err == nil {
				key := strconv.Itoa(singleOpt)
				st.distMap[key]++
			} else {
				var multiOpts []int
				if err := json.Unmarshal(a.SelectedOptions, &multiOpts); err == nil {
					for _, opt := range multiOpts {
						key := strconv.Itoa(opt)
						st.distMap[key]++
					}
				}
			}
		}
	}

	items := make([]structs.InstructorQuestionMetricsItem, len(questions))
	for i, q := range questions {
		st := statsMap[q.ID]
		unanswered := eligibleCompletedAttempts - st.totalAnswered
		if unanswered < 0 {
			unanswered = 0
		}

		var diffIdx *float64
		if st.totalAnswered > 0 {
			val := math.Round((float64(st.correctCount)/float64(st.totalAnswered))*10000) / 100
			diffIdx = &val
		}

		var discIdx *float64
		if totalEligible >= 10 && st.topTotal > 0 && st.bottomTotal > 0 {
			pTop := float64(st.topCorrect) / float64(st.topTotal)
			pBottom := float64(st.bottomCorrect) / float64(st.bottomTotal)
			val := math.Round((pTop-pBottom)*10000) / 100
			discIdx = &val
		}

		var avgTime *float64
		if st.timeCount > 0 {
			val := math.Round((float64(st.sumTime)/float64(st.timeCount))*10) / 10
			avgTime = &val
		}

		qText := q.Question
		if decoded, err := unescapeString(qText); err == nil {
			qText = decoded
		}

		items[i] = structs.InstructorQuestionMetricsItem{
			QuestionID:          q.ID.String(),
			QuestionText:        qText,
			QuestionType:        q.Type,
			OrderNumber:         q.OrderNo,
			TotalAnswered:       st.totalAnswered,
			CorrectCount:        st.correctCount,
			IncorrectCount:      st.incorrectCount,
			UnansweredCount:     unanswered,
			DifficultyIndex:     diffIdx,
			DiscriminationIndex: discIdx,
			AnswerDistribution:  st.distMap,
			AverageTimeSeconds:  avgTime,
		}
	}

	nextCursor := ""
	if hasMore && len(questions) > 0 {
		last := questions[len(questions)-1]
		nextCursor = EncodeCursor(last.OrderNo, last.ID)
	}

	resp := structs.InstructorQuestionMetricsResponse{
		QuizID:     quizID.String(),
		Questions:  items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}

	svc.cache.SetQuizJSON(ctx, quizID.String(), hash, resp)
	return resp, nil
}

// -----------------------------------------------------------------------------
// 9. Per-Quiz Engagement Metrics
// -----------------------------------------------------------------------------

func (svc *InstructorAnalyticsService) GetEngagementMetrics(
	ctx context.Context,
	quizID uuid.UUID,
) (structs.InstructorEngagementResponse, error) {
	hash := CanonicalEndpointHash("engagement")
	var cached structs.InstructorEngagementResponse
	if svc.cache.GetQuizJSON(ctx, quizID.String(), hash, &cached) {
		return cached, nil
	}

	acceptedEvents := []string{
		"QUESTION_VIEWED",
		"ANSWER_SELECTED",
		"ANSWER_CHANGED",
		"QUESTION_TIME_SPENT",
		"HINT_OPENED",
		"REVIEW_OPENED",
		"RESULT_VIEWED",
	}

	type eventCountRow struct {
		EventType string `db:"event_type"`
		Cnt       int    `db:"cnt"`
	}

	var counts []eventCountRow
	err := svc.db.From("assessment_analytics_events").
		Select("event_type", goqu.COUNT("id").As("cnt")).
		Where(goqu.Ex{
			"quiz_id":    quizID,
			"event_type": acceptedEvents,
		}).
		GroupBy("event_type").
		ScanStructs(&counts)
	if err != nil {
		return structs.InstructorEngagementResponse{}, err
	}

	typeMap := map[string]int{}
	totalAcceptedEvents := 0
	for _, c := range counts {
		typeMap[c.EventType] = c.Cnt
		totalAcceptedEvents += c.Cnt
	}

	var uniqueLearners int
	_, _ = svc.db.From("assessment_analytics_events").
		Where(goqu.Ex{
			"quiz_id":    quizID,
			"event_type": acceptedEvents,
		}).
		Select(goqu.COUNT(goqu.DISTINCT("user_id"))).
		ScanVal(&uniqueLearners)

	var uniqueAttempts int
	_, _ = svc.db.From("assessment_analytics_events").
		Where(goqu.Ex{
			"quiz_id":    quizID,
			"event_type": acceptedEvents,
		}).
		Select(goqu.COUNT(goqu.DISTINCT("attempt_id"))).
		ScanVal(&uniqueAttempts)

	avgEventsPerAttempt := 0.0
	if uniqueAttempts > 0 {
		avgEventsPerAttempt = math.Round((float64(totalAcceptedEvents)/float64(uniqueAttempts))*100) / 100
	}

	resp := structs.InstructorEngagementResponse{
		QuizID:                  quizID.String(),
		TotalQuestionViews:      typeMap["QUESTION_VIEWED"],
		TotalAnswerSelections:   typeMap["ANSWER_SELECTED"],
		TotalAnswerChanges:      typeMap["ANSWER_CHANGED"],
		TotalHintsOpened:        typeMap["HINT_OPENED"],
		TotalReviewsOpened:      typeMap["REVIEW_OPENED"],
		TotalResultViews:        typeMap["RESULT_VIEWED"],
		UniqueEngagedLearners:   uniqueLearners,
		UniqueEngagedAttempts:   uniqueAttempts,
		AverageEventsPerAttempt: avgEventsPerAttempt,
	}

	svc.cache.SetQuizJSON(ctx, quizID.String(), hash, resp)
	return resp, nil
}
