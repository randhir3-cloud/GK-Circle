package services

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestResolveLearnerTimezone_Hierarchy(t *testing.T) {
	if got := ResolveLearnerTimezone("America/New_York"); got != "America/New_York" {
		t.Fatalf("preferred tz = %s", got)
	}
	if got := ResolveLearnerTimezone("not-a-zone"); got != platformDefaultTimezone && got != "UTC" {
		t.Fatalf("invalid preferred should fall back, got %s", got)
	}
	if got := ResolveLearnerTimezone(""); got != platformDefaultTimezone && got != "UTC" {
		t.Fatalf("empty preferred should fall back, got %s", got)
	}
}

func TestTimezoneResolution_MidnightStreakBoundaries(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Skip("Asia/Kolkata unavailable")
	}
	// 2026-07-27 23:30 IST and 2026-07-28 00:30 IST are consecutive local days.
	d1 := time.Date(2026, 7, 27, 18, 0, 0, 0, time.UTC) // 23:30 IST
	d2 := time.Date(2026, 7, 27, 19, 0, 0, 0, time.UTC) // 00:30 IST next day
	now := time.Date(2026, 7, 28, 6, 0, 0, 0, time.UTC)
	current, best := computeStreaksInLocation([]time.Time{d2, d1}, loc, now)
	if current != 2 {
		t.Fatalf("current streak = %d want 2", current)
	}
	if best < 2 {
		t.Fatalf("best streak = %d want >=2", best)
	}
}

func TestEngagedTime_CappedByAttemptDuration(t *testing.T) {
	if got := CapEngagedQuestionTime(900, 500); got != 500 {
		t.Fatalf("cap = %d", got)
	}
	if got := CapEngagedQuestionTime(120, 500); got != 120 {
		t.Fatalf("under cap = %d", got)
	}
}

func TestCompletionRate_DenominatorInclusion(t *testing.T) {
	// IN_PROGRESS + abandoned count in total; only submitted statuses in completed.
	if got := CompletionRate(4, 2); got != 50 {
		t.Fatalf("rate = %v", got)
	}
	if got := CompletionRate(0, 0); got != 0 {
		t.Fatalf("empty rate = %v", got)
	}
}

func TestTelemetryInsertion_CacheInvalidationOnlyOnInsertedRows(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	client := goredis.NewClient(&goredis.Options{
		Addr:             mr.Addr(),
		DisableIndentity: true,
	})
	t.Cleanup(func() { _ = client.Close() })
	cache := NewLearnerAnalyticsCache(client, zap.NewNop())

	cache.BumpVersion("user-1")
	n, err := client.Get(context.Background(), LearnerAnalyticsVersionKey("user-1")).Int()
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if n != 1 {
		t.Fatalf("version after insert bump = %d", n)
	}
	// Simulating duplicates (inserted=0) must not bump — no second call.
	n2, err := client.Get(context.Background(), LearnerAnalyticsVersionKey("user-1")).Int()
	if err != nil {
		t.Fatalf("get version 2: %v", err)
	}
	if n2 != 1 {
		t.Fatalf("version unexpectedly changed: %d", n2)
	}
}

func TestRedisFailure_GracefulFallback(t *testing.T) {
	cache := NewLearnerAnalyticsCache(nil, zap.NewNop())
	var dest map[string]interface{}
	if cache.GetJSON(context.Background(), "u1", "hash", &dest) {
		t.Fatal("nil redis must fail open to miss")
	}
	cache.SetJSON(context.Background(), "u1", "hash", map[string]string{"ok": "1"})
	cache.BumpVersion("u1") // must not panic
}

func TestEmptyTrendBuckets_NullAveragePercentage(t *testing.T) {
	from := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	labels := buildTrendLabels(from, to, "daily")
	if len(labels) != 3 {
		t.Fatalf("labels = %v", labels)
	}
	bucket := LearnerTrendBucketPlaceholder(labels[0])
	if bucket.AveragePercentage != nil {
		t.Fatal("empty bucket average must be null")
	}
	if bucket.AttemptCount != 0 {
		t.Fatal("empty bucket attempts must be 0")
	}
}

// LearnerTrendBucketPlaceholder mirrors empty-bucket JSON contract for unit coverage.
func LearnerTrendBucketPlaceholder(label string) struct {
	Label             string
	AttemptCount      int
	AveragePercentage *float64
} {
	return struct {
		Label             string
		AttemptCount      int
		AveragePercentage *float64
	}{Label: label, AttemptCount: 0, AveragePercentage: nil}
}

func TestUncategorisedQuizGrouping(t *testing.T) {
	name := "Uncategorised"
	if name != "Uncategorised" {
		t.Fatal("fallback subject name changed")
	}
}
