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
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"go.uber.org/zap"
)

const (
	platformDefaultTimezone     = "Asia/Kolkata"
	engagedQuestionTimeCapSecs  = 600
	learnerActivityDefaultLimit = 20
	learnerActivityMaxLimit     = 100
	learnerTrendMaxRangeDays    = 366
)

var (
	ErrLearnerAnalyticsInvalidRange  = errors.New("invalid analytics date range")
	ErrLearnerAnalyticsInvalidCursor = errors.New("invalid analytics activity cursor")
	ErrLearnerAnalyticsForbidden     = errors.New("analytics attempt access forbidden")
	ErrLearnerAnalyticsNotFound      = errors.New("analytics attempt not found")
)

// Bulk result-release visibility projection (single expression, zero N+1).
const resultVisibleSQL = `(q.result_release_policy = 'IMMEDIATE' OR q.results_released = true OR (q.result_release_policy = 'SCHEDULED' AND q.results_scheduled_at IS NOT NULL AND NOW() >= q.results_scheduled_at))`

type LearnerAnalyticsAggregationService struct {
	db     *goqu.Database
	cache  *LearnerAnalyticsCache
	logger *zap.Logger
}

func NewLearnerAnalyticsAggregationService(
	db *goqu.Database,
	cache *LearnerAnalyticsCache,
	logger *zap.Logger,
) *LearnerAnalyticsAggregationService {
	if cache == nil {
		cache = NewLearnerAnalyticsCache(nil, logger)
	}
	return &LearnerAnalyticsAggregationService{db: db, cache: cache, logger: logger}
}

func (svc *LearnerAnalyticsAggregationService) Cache() *LearnerAnalyticsCache {
	return svc.cache
}

func ResolveLearnerTimezone(preferred string) string {
	preferred = strings.TrimSpace(preferred)
	if preferred != "" {
		if _, err := time.LoadLocation(preferred); err == nil {
			return preferred
		}
	}
	if _, err := time.LoadLocation(platformDefaultTimezone); err == nil {
		return platformDefaultTimezone
	}
	return "UTC"
}

func CapEngagedQuestionTime(boundedTelemetrySum, authoritativeAttemptDuration int64) int64 {
	if boundedTelemetrySum < 0 {
		boundedTelemetrySum = 0
	}
	if authoritativeAttemptDuration < 0 {
		authoritativeAttemptDuration = 0
	}
	if boundedTelemetrySum > authoritativeAttemptDuration {
		return authoritativeAttemptDuration
	}
	return boundedTelemetrySum
}

func CompletionRate(totalAttempts, completedAttempts int) float64 {
	if totalAttempts <= 0 {
		return 0
	}
	return math.Round((float64(completedAttempts)/float64(totalAttempts))*10000) / 100
}

type learnerAttemptRow struct {
	ID                uuid.UUID      `db:"id"`
	QuizID            uuid.UUID      `db:"quiz_id"`
	UserID            string         `db:"user_id"`
	Status            string         `db:"status"`
	TotalScore        sql.NullFloat64 `db:"total_score"`
	MaxScore          sql.NullFloat64 `db:"max_score"`
	TimeTakenSeconds  sql.NullInt64  `db:"time_taken_seconds"`
	SubmittedAt       sql.NullTime   `db:"submitted_at"`
	CreatedAt         time.Time      `db:"created_at"`
	QuizTitle         string         `db:"quiz_title"`
	CategoryID        sql.NullString `db:"category_id"`
	CategoryName      sql.NullString `db:"category_name"`
	ResultsVisible    bool           `db:"results_visible"`
}

func (svc *LearnerAnalyticsAggregationService) loadLearnerAttempts(userID string) ([]learnerAttemptRow, error) {
	var rows []learnerAttemptRow
	err := svc.db.From(goqu.T("assessment_attempts").As("a")).
		LeftJoin(goqu.T("quizzes").As("q"), goqu.On(goqu.I("q.id").Eq(goqu.I("a.quiz_id")))).
		LeftJoin(goqu.T("quiz_categories").As("c"), goqu.On(goqu.I("c.id").Eq(goqu.I("q.category_id")))).
		Select(
			goqu.I("a.id"),
			goqu.I("a.quiz_id"),
			goqu.I("a.user_id"),
			goqu.I("a.status"),
			goqu.I("a.total_score"),
			goqu.I("a.max_score"),
			goqu.I("a.time_taken_seconds"),
			goqu.I("a.submitted_at"),
			goqu.I("a.created_at"),
			goqu.L("COALESCE(q.title, '')").As("quiz_title"),
			goqu.I("q.category_id"),
			goqu.I("c.name").As("category_name"),
			goqu.L(resultVisibleSQL).As("results_visible"),
		).
		Where(goqu.Ex{"a.user_id": userID}).
		ScanStructs(&rows)
	return rows, err
}

func percentageFromAttempt(row learnerAttemptRow) *float64 {
	if !row.ResultsVisible {
		return nil
	}
	if row.Status != models.AttemptStatusSubmitted && row.Status != models.AttemptStatusAutoSubmitted {
		return nil
	}
	if !row.TotalScore.Valid || !row.MaxScore.Valid || row.MaxScore.Float64 <= 0 {
		return nil
	}
	pct := math.Round((row.TotalScore.Float64/row.MaxScore.Float64)*10000) / 100
	return &pct
}

func resultStatus(row learnerAttemptRow) string {
	switch row.Status {
	case models.AttemptStatusInProgress:
		return "In Progress"
	case models.AttemptStatusAbandoned:
		return "Abandoned"
	case models.AttemptStatusSubmitted, models.AttemptStatusAutoSubmitted:
		if row.ResultsVisible {
			return "Released"
		}
		return "Result Pending"
	default:
		return row.Status
	}
}

func (svc *LearnerAnalyticsAggregationService) engagedTimeForUser(userID string, attempts []learnerAttemptRow) (int64, error) {
	authByAttempt := map[uuid.UUID]int64{}
	var authTotal int64
	for _, row := range attempts {
		var dur int64
		if row.TimeTakenSeconds.Valid {
			dur = row.TimeTakenSeconds.Int64
		}
		if dur < 0 {
			dur = 0
		}
		authByAttempt[row.ID] = dur
		authTotal += dur
	}

	type telemRow struct {
		AttemptID uuid.UUID       `db:"attempt_id"`
		Metadata  json.RawMessage `db:"metadata"`
	}
	var telem []telemRow
	err := svc.db.From("assessment_analytics_events").
		Select("attempt_id", "metadata").
		Where(goqu.Ex{
			"user_id":    userID,
			"event_type": string(structs.EventQuestionTimeSpent),
		}).
		ScanStructs(&telem)
	if err != nil {
		return 0, err
	}

	perAttempt := map[uuid.UUID]int64{}
	for _, event := range telem {
		duration := extractDurationSeconds(event.Metadata)
		if duration <= 0 {
			continue
		}
		if duration > engagedQuestionTimeCapSecs {
			duration = engagedQuestionTimeCapSecs
		}
		perAttempt[event.AttemptID] += duration
	}

	var engaged int64
	for attemptID, bounded := range perAttempt {
		auth := authByAttempt[attemptID]
		engaged += CapEngagedQuestionTime(bounded, auth)
	}
	_ = authTotal
	return engaged, nil
}

func extractDurationSeconds(metadata json.RawMessage) int64 {
	if len(metadata) == 0 {
		return 0
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(metadata, &payload); err != nil {
		return 0
	}
	for _, key := range []string{"duration_seconds", "duration", "seconds"} {
		if raw, ok := payload[key]; ok {
			switch v := raw.(type) {
			case float64:
				return int64(v)
			case string:
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err == nil {
					return n
				}
			}
		}
	}
	return 0
}

func computeStreaks(dates []string) (current, best int) {
	if len(dates) == 0 {
		return 0, 0
	}
	uniq := make([]string, 0, len(dates))
	seen := map[string]struct{}{}
	for _, d := range dates {
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		uniq = append(uniq, d)
	}
	// dates expected descending
	best = 1
	run := 1
	for i := 1; i < len(uniq); i++ {
		prev, _ := time.Parse("2006-01-02", uniq[i-1])
		cur, _ := time.Parse("2006-01-02", uniq[i])
		if prev.Sub(cur) == 24*time.Hour {
			run++
			if run > best {
				best = run
			}
		} else {
			run = 1
		}
	}
	current = 1
	today := time.Now().UTC().Format("2006-01-02")
	// current streak uses first date relative to "today" in caller TZ — caller passes TZ-local dates already sorted desc
	if uniq[0] != today {
		yesterday := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02")
		// Fallback comparison in UTC labels; proper TZ handled by caller via local "today"
		_ = yesterday
	}
	run = 1
	for i := 1; i < len(uniq); i++ {
		prev, _ := time.Parse("2006-01-02", uniq[i-1])
		cur, _ := time.Parse("2006-01-02", uniq[i])
		if prev.Sub(cur) == 24*time.Hour {
			run++
		} else {
			break
		}
	}
	current = run
	return current, best
}

func computeStreaksInLocation(dates []time.Time, loc *time.Location, now time.Time) (current, best int) {
	if len(dates) == 0 {
		return 0, 0
	}
	daySet := map[string]struct{}{}
	dayList := make([]string, 0)
	for _, ts := range dates {
		key := ts.In(loc).Format("2006-01-02")
		if _, ok := daySet[key]; ok {
			continue
		}
		daySet[key] = struct{}{}
		dayList = append(dayList, key)
	}
	// sort descending
	for i := 0; i < len(dayList); i++ {
		for j := i + 1; j < len(dayList); j++ {
			if dayList[j] > dayList[i] {
				dayList[i], dayList[j] = dayList[j], dayList[i]
			}
		}
	}

	best = 0
	run := 0
	var prev *time.Time
	for _, key := range dayList {
		day, _ := time.ParseInLocation("2006-01-02", key, loc)
		if prev == nil || prev.Sub(day) == 24*time.Hour {
			run++
		} else {
			run = 1
		}
		if run > best {
			best = run
		}
		d := day
		prev = &d
	}

	todayLocal := now.In(loc)
	today := time.Date(todayLocal.Year(), todayLocal.Month(), todayLocal.Day(), 0, 0, 0, 0, loc)
	yesterday := today.AddDate(0, 0, -1)
	if len(dayList) == 0 {
		return 0, best
	}
	first, _ := time.ParseInLocation("2006-01-02", dayList[0], loc)
	if !(first.Equal(today) || first.Equal(yesterday)) {
		return 0, best
	}
	current = 1
	for i := 1; i < len(dayList); i++ {
		prevDay, _ := time.ParseInLocation("2006-01-02", dayList[i-1], loc)
		curDay, _ := time.ParseInLocation("2006-01-02", dayList[i], loc)
		if prevDay.Sub(curDay) == 24*time.Hour {
			current++
		} else {
			break
		}
	}
	return current, best
}

func (svc *LearnerAnalyticsAggregationService) GetDashboardSummary(
	ctx context.Context,
	userID, preferredTZ string,
) (structs.LearnerDashboardSummaryResponse, error) {
	tz := ResolveLearnerTimezone(preferredTZ)
	hash := endpointHash("dashboard", tz)
	var cached structs.LearnerDashboardSummaryResponse
	if svc.cache.GetJSON(ctx, userID, hash, &cached) {
		return cached, nil
	}

	attempts, err := svc.loadLearnerAttempts(userID)
	if err != nil {
		return structs.LearnerDashboardSummaryResponse{}, err
	}

	loc, _ := time.LoadLocation(tz)
	if loc == nil {
		loc = time.UTC
	}

	total := len(attempts)
	completed := 0
	scored := 0
	pending := 0
	var durationSum int64
	var pctSum float64
	streakTimes := make([]time.Time, 0)

	for _, row := range attempts {
		if row.TimeTakenSeconds.Valid {
			durationSum += row.TimeTakenSeconds.Int64
		}
		if row.Status == models.AttemptStatusSubmitted || row.Status == models.AttemptStatusAutoSubmitted {
			completed++
			if row.SubmittedAt.Valid {
				streakTimes = append(streakTimes, row.SubmittedAt.Time)
			} else {
				streakTimes = append(streakTimes, row.CreatedAt)
			}
			if !row.ResultsVisible {
				pending++
				continue
			}
			if pct := percentageFromAttempt(row); pct != nil {
				scored++
				pctSum += *pct
			}
		}
	}

	engaged, err := svc.engagedTimeForUser(userID, attempts)
	if err != nil {
		return structs.LearnerDashboardSummaryResponse{}, err
	}
	engaged = CapEngagedQuestionTime(engaged, durationSum)

	current, best := computeStreaksInLocation(streakTimes, loc, time.Now().UTC())
	var avg *float64
	if scored > 0 {
		v := math.Round((pctSum/float64(scored))*100) / 100
		avg = &v
	}

	resp := structs.LearnerDashboardSummaryResponse{
		ResolvedTimezone:               tz,
		TotalAttempts:                  total,
		CompletedAttempts:              completed,
		CompletionRate:                 CompletionRate(total, completed),
		ScoredAttemptCount:             scored,
		PendingResultCount:             pending,
		AveragePercentage:              avg,
		AssessmentDurationSeconds:      durationSum,
		EngagedQuestionTimeSeconds:     engaged,
		EngagedQuestionTimeApproximate: true,
		CurrentStreakDays:              current,
		BestStreakDays:                 best,
	}
	svc.cache.SetJSON(ctx, userID, hash, resp)
	return resp, nil
}

type activityCursor struct {
	ActivityAt time.Time
	CreatedAt  time.Time
	ID         uuid.UUID
}

func encodeActivityCursor(c activityCursor) string {
	raw := fmt.Sprintf("%s|%s|%s",
		c.ActivityAt.UTC().Format(time.RFC3339Nano),
		c.CreatedAt.UTC().Format(time.RFC3339Nano),
		c.ID.String(),
	)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeActivityCursor(raw string) (activityCursor, error) {
	if strings.TrimSpace(raw) == "" {
		return activityCursor{}, ErrLearnerAnalyticsInvalidCursor
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return activityCursor{}, ErrLearnerAnalyticsInvalidCursor
	}
	parts := strings.SplitN(string(decoded), "|", 3)
	if len(parts) != 3 {
		return activityCursor{}, ErrLearnerAnalyticsInvalidCursor
	}
	activityAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return activityCursor{}, ErrLearnerAnalyticsInvalidCursor
	}
	createdAt, err := time.Parse(time.RFC3339Nano, parts[1])
	if err != nil {
		return activityCursor{}, ErrLearnerAnalyticsInvalidCursor
	}
	id, err := uuid.Parse(parts[2])
	if err != nil {
		return activityCursor{}, ErrLearnerAnalyticsInvalidCursor
	}
	return activityCursor{ActivityAt: activityAt.UTC(), CreatedAt: createdAt.UTC(), ID: id}, nil
}

func (svc *LearnerAnalyticsAggregationService) GetRecentActivity(
	ctx context.Context,
	userID, preferredTZ, cursor string,
	limit int,
) (structs.LearnerRecentActivityResponse, error) {
	tz := ResolveLearnerTimezone(preferredTZ)
	if limit <= 0 {
		limit = learnerActivityDefaultLimit
	}
	if limit > learnerActivityMaxLimit {
		limit = learnerActivityMaxLimit
	}

	hash := endpointHash("activity", tz, cursor, strconv.Itoa(limit))
	var cached structs.LearnerRecentActivityResponse
	if svc.cache.GetJSON(ctx, userID, hash, &cached) {
		return cached, nil
	}

	query := svc.db.From(goqu.T("assessment_attempts").As("a")).
		LeftJoin(goqu.T("quizzes").As("q"), goqu.On(goqu.I("q.id").Eq(goqu.I("a.quiz_id")))).
		Select(
			goqu.I("a.id"),
			goqu.I("a.quiz_id"),
			goqu.I("a.user_id"),
			goqu.I("a.status"),
			goqu.I("a.total_score"),
			goqu.I("a.max_score"),
			goqu.I("a.time_taken_seconds"),
			goqu.I("a.submitted_at"),
			goqu.I("a.created_at"),
			goqu.L("COALESCE(q.title, '')").As("quiz_title"),
			goqu.L("NULL").As("category_id"),
			goqu.L("NULL").As("category_name"),
			goqu.L(resultVisibleSQL).As("results_visible"),
		).
		Where(goqu.Ex{"a.user_id": userID})

	if cursor != "" {
		cur, err := decodeActivityCursor(cursor)
		if err != nil {
			return structs.LearnerRecentActivityResponse{}, err
		}
		query = query.Where(goqu.L(
			`(COALESCE(a.submitted_at, a.created_at), a.created_at, a.id) < (?, ?, ?)`,
			cur.ActivityAt, cur.CreatedAt, cur.ID,
		))
	}

	var rows []learnerAttemptRow
	err := query.
		Order(
			goqu.L("COALESCE(a.submitted_at, a.created_at)").Desc(),
			goqu.I("a.created_at").Desc(),
			goqu.I("a.id").Desc(),
		).
		Limit(uint(limit + 1)).
		ScanStructs(&rows)
	if err != nil {
		return structs.LearnerRecentActivityResponse{}, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]structs.LearnerActivityItem, 0, len(rows))
	var nextCursor string
	for i, row := range rows {
		activityAt := row.CreatedAt
		var submittedAt *string
		if row.SubmittedAt.Valid {
			activityAt = row.SubmittedAt.Time
			formatted := row.SubmittedAt.Time.UTC().Format(time.RFC3339)
			submittedAt = &formatted
		}
		item := structs.LearnerActivityItem{
			AttemptID:    row.ID.String(),
			QuizID:       row.QuizID.String(),
			QuizTitle:    row.QuizTitle,
			Status:       row.Status,
			ResultStatus: resultStatus(row),
			Percentage:   percentageFromAttempt(row),
			ActivityAt:   activityAt.UTC().Format(time.RFC3339),
			CreatedAt:    row.CreatedAt.UTC().Format(time.RFC3339),
			SubmittedAt:  submittedAt,
		}
		if row.ResultsVisible {
			if row.TotalScore.Valid {
				v := row.TotalScore.Float64
				item.TotalScore = &v
			}
			if row.MaxScore.Valid {
				v := row.MaxScore.Float64
				item.MaxScore = &v
			}
		}
		if row.TimeTakenSeconds.Valid {
			v := row.TimeTakenSeconds.Int64
			item.DurationSeconds = &v
		}
		items = append(items, item)
		if i == len(rows)-1 {
			nextCursor = encodeActivityCursor(activityCursor{
				ActivityAt: activityAt,
				CreatedAt:  row.CreatedAt,
				ID:         row.ID,
			})
		}
	}
	if !hasMore {
		nextCursor = ""
	}

	resp := structs.LearnerRecentActivityResponse{
		ResolvedTimezone: tz,
		Items:            items,
		NextCursor:       nextCursor,
		HasMore:          hasMore,
	}
	svc.cache.SetJSON(ctx, userID, hash, resp)
	return resp, nil
}

func (svc *LearnerAnalyticsAggregationService) GetPerformanceTrends(
	ctx context.Context,
	userID, preferredTZ, granularity string,
	from, to time.Time,
) (structs.LearnerPerformanceTrendsResponse, error) {
	tz := ResolveLearnerTimezone(preferredTZ)
	granularity = strings.ToLower(strings.TrimSpace(granularity))
	if granularity == "" {
		granularity = "daily"
	}
	if granularity != "daily" && granularity != "weekly" && granularity != "monthly" {
		return structs.LearnerPerformanceTrendsResponse{}, ErrLearnerAnalyticsInvalidRange
	}
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return structs.LearnerPerformanceTrendsResponse{}, ErrLearnerAnalyticsInvalidRange
	}
	if to.Sub(from) > time.Duration(learnerTrendMaxRangeDays)*24*time.Hour {
		return structs.LearnerPerformanceTrendsResponse{}, ErrLearnerAnalyticsInvalidRange
	}

	hash := endpointHash("trends", tz, granularity, from.Format(time.RFC3339), to.Format(time.RFC3339))
	var cached structs.LearnerPerformanceTrendsResponse
	if svc.cache.GetJSON(ctx, userID, hash, &cached) {
		return cached, nil
	}

	loc, _ := time.LoadLocation(tz)
	if loc == nil {
		loc = time.UTC
	}

	attempts, err := svc.loadLearnerAttempts(userID)
	if err != nil {
		return structs.LearnerPerformanceTrendsResponse{}, err
	}

	type agg struct {
		attemptCount int
		scoredCount  int
		pctSum       float64
		duration     int64
	}
	bucketMap := map[string]*agg{}

	for _, row := range attempts {
		ts := row.CreatedAt
		if row.SubmittedAt.Valid {
			ts = row.SubmittedAt.Time
		}
		local := ts.In(loc)
		if local.Before(from.In(loc)) || local.After(to.In(loc)) {
			continue
		}
		label := trendBucketLabel(local, granularity)
		bucket, ok := bucketMap[label]
		if !ok {
			bucket = &agg{}
			bucketMap[label] = bucket
		}
		bucket.attemptCount++
		if row.TimeTakenSeconds.Valid {
			bucket.duration += row.TimeTakenSeconds.Int64
		}
		if pct := percentageFromAttempt(row); pct != nil {
			bucket.scoredCount++
			bucket.pctSum += *pct
		}
	}

	labels := buildTrendLabels(from.In(loc), to.In(loc), granularity)
	buckets := make([]structs.LearnerTrendBucket, 0, len(labels))
	for _, label := range labels {
		item := structs.LearnerTrendBucket{Label: label}
		if data, ok := bucketMap[label]; ok {
			item.AttemptCount = data.attemptCount
			item.ScoredAttemptCount = data.scoredCount
			item.AssessmentDurationSeconds = data.duration
			if data.scoredCount > 0 {
				v := math.Round((data.pctSum/float64(data.scoredCount))*100) / 100
				item.AveragePercentage = &v
			}
		}
		buckets = append(buckets, item)
	}

	resp := structs.LearnerPerformanceTrendsResponse{
		ResolvedTimezone: tz,
		Granularity:      granularity,
		From:             from.UTC().Format(time.RFC3339),
		To:               to.UTC().Format(time.RFC3339),
		Buckets:          buckets,
	}
	svc.cache.SetJSON(ctx, userID, hash, resp)
	return resp, nil
}

func trendBucketLabel(ts time.Time, granularity string) string {
	switch granularity {
	case "weekly":
		year, week := ts.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case "monthly":
		return ts.Format("2006-01")
	default:
		return ts.Format("2006-01-02")
	}
}

func buildTrendLabels(from, to time.Time, granularity string) []string {
	labels := []string{}
	switch granularity {
	case "weekly":
		cur := startOfISOWeek(from)
		end := startOfISOWeek(to)
		for !cur.After(end) {
			labels = append(labels, trendBucketLabel(cur, granularity))
			cur = cur.AddDate(0, 0, 7)
		}
	case "monthly":
		cur := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location())
		for !cur.After(end) {
			labels = append(labels, trendBucketLabel(cur, granularity))
			cur = cur.AddDate(0, 1, 0)
		}
	default:
		cur := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
		for !cur.After(end) {
			labels = append(labels, trendBucketLabel(cur, granularity))
			cur = cur.AddDate(0, 0, 1)
		}
	}
	return labels
}

func startOfISOWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, -(weekday - 1))
}

func (svc *LearnerAnalyticsAggregationService) GetSubjectPerformance(
	ctx context.Context,
	userID, preferredTZ string,
) (structs.LearnerSubjectPerformanceResponse, error) {
	tz := ResolveLearnerTimezone(preferredTZ)
	hash := endpointHash("subjects", tz)
	var cached structs.LearnerSubjectPerformanceResponse
	if svc.cache.GetJSON(ctx, userID, hash, &cached) {
		return cached, nil
	}

	attempts, err := svc.loadLearnerAttempts(userID)
	if err != nil {
		return structs.LearnerSubjectPerformanceResponse{}, err
	}

	type agg struct {
		subjectID *string
		name      string
		attempts  int
		scored    int
		pctSum    float64
		duration  int64
	}
	grouped := map[string]*agg{}

	for _, row := range attempts {
		key := "uncategorised"
		name := "Uncategorised"
		var subjectID *string
		if row.CategoryID.Valid && row.CategoryID.String != "" {
			key = row.CategoryID.String
			id := row.CategoryID.String
			subjectID = &id
			if row.CategoryName.Valid && row.CategoryName.String != "" {
				name = row.CategoryName.String
			} else {
				name = "Uncategorised"
			}
		}
		bucket, ok := grouped[key]
		if !ok {
			bucket = &agg{subjectID: subjectID, name: name}
			grouped[key] = bucket
		}
		bucket.attempts++
		if row.TimeTakenSeconds.Valid {
			bucket.duration += row.TimeTakenSeconds.Int64
		}
		if pct := percentageFromAttempt(row); pct != nil {
			bucket.scored++
			bucket.pctSum += *pct
		}
	}

	subjects := make([]structs.LearnerSubjectPerformanceItem, 0, len(grouped))
	for _, bucket := range grouped {
		item := structs.LearnerSubjectPerformanceItem{
			SubjectID:                 bucket.subjectID,
			SubjectName:               bucket.name,
			AttemptCount:              bucket.attempts,
			ScoredAttemptCount:        bucket.scored,
			AssessmentDurationSeconds: bucket.duration,
		}
		if bucket.scored > 0 {
			v := math.Round((bucket.pctSum/float64(bucket.scored))*100) / 100
			item.AveragePercentage = &v
		}
		subjects = append(subjects, item)
	}

	// Deterministic order: name asc, Uncategorised last
	for i := 0; i < len(subjects); i++ {
		for j := i + 1; j < len(subjects); j++ {
			ai, aj := subjects[i], subjects[j]
			aiUncat := ai.SubjectID == nil
			ajUncat := aj.SubjectID == nil
			swap := false
			if aiUncat != ajUncat {
				swap = aiUncat && !ajUncat
			} else if ai.SubjectName > aj.SubjectName {
				swap = true
			}
			if swap {
				subjects[i], subjects[j] = subjects[j], subjects[i]
			}
		}
	}

	resp := structs.LearnerSubjectPerformanceResponse{
		ResolvedTimezone: tz,
		Subjects:         subjects,
	}
	svc.cache.SetJSON(ctx, userID, hash, resp)
	return resp, nil
}

func (svc *LearnerAnalyticsAggregationService) GetAttemptTimeline(
	ctx context.Context,
	userID, preferredTZ string,
	attemptID uuid.UUID,
) (structs.LearnerTimelineResponse, error) {
	tz := ResolveLearnerTimezone(preferredTZ)
	hash := endpointHash("timeline", tz, attemptID.String())
	var cached structs.LearnerTimelineResponse
	if svc.cache.GetJSON(ctx, userID, hash, &cached) {
		return cached, nil
	}

	var row learnerAttemptRow
	found, err := svc.db.From(goqu.T("assessment_attempts").As("a")).
		LeftJoin(goqu.T("quizzes").As("q"), goqu.On(goqu.I("q.id").Eq(goqu.I("a.quiz_id")))).
		Select(
			goqu.I("a.id"),
			goqu.I("a.quiz_id"),
			goqu.I("a.user_id"),
			goqu.I("a.status"),
			goqu.I("a.total_score"),
			goqu.I("a.max_score"),
			goqu.I("a.time_taken_seconds"),
			goqu.I("a.submitted_at"),
			goqu.I("a.created_at"),
			goqu.L("COALESCE(q.title, '')").As("quiz_title"),
			goqu.L("NULL").As("category_id"),
			goqu.L("NULL").As("category_name"),
			goqu.L(resultVisibleSQL).As("results_visible"),
		).
		Where(goqu.Ex{"a.id": attemptID}).
		Limit(1).
		ScanStruct(&row)
	if err != nil {
		return structs.LearnerTimelineResponse{}, err
	}
	if !found {
		return structs.LearnerTimelineResponse{}, ErrLearnerAnalyticsNotFound
	}
	if row.UserID != userID {
		return structs.LearnerTimelineResponse{}, ErrLearnerAnalyticsForbidden
	}

	var events []models.AssessmentAnalyticsEvent
	err = svc.db.From("assessment_analytics_events").
		Where(goqu.Ex{"attempt_id": attemptID, "user_id": userID}).
		Order(goqu.I("occurred_at").Asc(), goqu.I("created_at").Asc(), goqu.I("id").Asc()).
		ScanStructs(&events)
	if err != nil {
		return structs.LearnerTimelineResponse{}, err
	}

	outEvents := make([]structs.LearnerTimelineEvent, 0, len(events))
	for _, event := range events {
		meta := map[string]interface{}{}
		if len(event.Metadata) > 0 {
			_ = json.Unmarshal(event.Metadata, &meta)
		}
		sanitizeTimelineMetadata(meta)
		outEvents = append(outEvents, structs.LearnerTimelineEvent{
			ID:            event.ID.String(),
			EventType:     event.EventType,
			EventSource:   event.EventSource,
			OccurredAt:    event.OccurredAt.UTC().Format(time.RFC3339),
			CorrelationID: event.CorrelationID,
			Metadata:      meta,
		})
	}

	resp := structs.LearnerTimelineResponse{
		ResolvedTimezone: tz,
		AttemptID:        row.ID.String(),
		QuizID:           row.QuizID.String(),
		QuizTitle:        row.QuizTitle,
		Status:           row.Status,
		ResultStatus:     resultStatus(row),
		Percentage:       percentageFromAttempt(row),
		Events:           outEvents,
	}
	svc.cache.SetJSON(ctx, userID, hash, resp)
	return resp, nil
}

func sanitizeTimelineMetadata(meta map[string]interface{}) {
	for key := range meta {
		normalized := strings.ToLower(strings.TrimSpace(key))
		if _, blocked := analyticsSensitiveMetadataKeys[normalized]; blocked {
			delete(meta, key)
		}
	}
}

func (svc *LearnerAnalyticsAggregationService) ListUserIDsForQuiz(quizID uuid.UUID) ([]string, error) {
	var userIDs []string
	err := svc.db.From("assessment_attempts").
		Select("user_id").
		Where(goqu.Ex{"quiz_id": quizID}).
		Distinct().
		ScanVals(&userIDs)
	return userIDs, err
}
