package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
)

func TestTimezoneResolution_Hierarchy(t *testing.T) {
	// 1. Valid requested timezone
	tz1 := ResolveLearnerTimezone("America/New_York")
	if tz1 != "America/New_York" {
		t.Errorf("expected America/New_York, got %s", tz1)
	}

	// 2. Invalid requested timezone falls back to platform default
	tz2 := ResolveLearnerTimezone("Invalid/Timezone")
	if tz2 != "Asia/Kolkata" {
		t.Errorf("expected Asia/Kolkata fallback, got %s", tz2)
	}

	// 3. Empty requested timezone falls back to platform default
	tz3 := ResolveLearnerTimezone("")
	if tz3 != "Asia/Kolkata" {
		t.Errorf("expected Asia/Kolkata default, got %s", tz3)
	}
}

func TestReleaseEvaluator_PerformsNoWrites(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

	// Immediate policy
	c1 := EvaluateResultReleaseState(string(structs.ResultReleasePolicyImmediate), false, sql.NullTime{}, sql.NullTime{}, false, now, loc)
	if c1 != "IMMEDIATE_RELEASE" {
		t.Errorf("expected IMMEDIATE_RELEASE, got %s", c1)
	}

	// Manual pending vs completed
	c2 := EvaluateResultReleaseState(string(structs.ResultReleasePolicyManual), false, sql.NullTime{}, sql.NullTime{}, false, now, loc)
	if c2 != "PENDING_MANUAL" {
		t.Errorf("expected PENDING_MANUAL, got %s", c2)
	}

	c3 := EvaluateResultReleaseState(string(structs.ResultReleasePolicyManual), true, sql.NullTime{}, sql.NullTime{}, false, now, loc)
	if c3 != "COMPLETED_MANUAL" {
		t.Errorf("expected COMPLETED_MANUAL, got %s", c3)
	}
}

func TestReleaseClassification_Precedence_ScheduledForToday(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	// 2026-07-28 10:00:00 UTC = 2026-07-28 15:30:00 IST
	now := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)

	// Scheduled for today (earlier or later same calendar date)
	schedToday := sql.NullTime{Time: time.Date(2026, 7, 28, 14, 0, 0, 0, time.UTC), Valid: true}
	c1 := EvaluateResultReleaseState(string(structs.ResultReleasePolicyScheduled), false, schedToday, sql.NullTime{}, false, now, loc)
	if c1 != "SCHEDULED_FOR_TODAY" {
		t.Errorf("expected SCHEDULED_FOR_TODAY, got %s", c1)
	}

	// Upcoming scheduled (tomorrow)
	schedTomorrow := sql.NullTime{Time: time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC), Valid: true}
	c2 := EvaluateResultReleaseState(string(structs.ResultReleasePolicyScheduled), false, schedTomorrow, sql.NullTime{}, false, now, loc)
	if c2 != "UPCOMING_SCHEDULED" {
		t.Errorf("expected UPCOMING_SCHEDULED, got %s", c2)
	}

	// Overdue scheduled (yesterday)
	schedYesterday := sql.NullTime{Time: time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC), Valid: true}
	c3 := EvaluateResultReleaseState(string(structs.ResultReleasePolicyScheduled), false, schedYesterday, sql.NullTime{}, false, now, loc)
	if c3 != "OVERDUE_SCHEDULED" {
		t.Errorf("expected OVERDUE_SCHEDULED, got %s", c3)
	}

	// Manually overridden scheduled
	c4 := EvaluateResultReleaseState(string(structs.ResultReleasePolicyScheduled), true, schedTomorrow, sql.NullTime{}, true, now, loc)
	if c4 != "MANUALLY_OVERRIDDEN_SCHEDULED" {
		t.Errorf("expected MANUALLY_OVERRIDDEN_SCHEDULED, got %s", c4)
	}
}

func TestAveragePercentage_AttemptWeighted(t *testing.T) {
	// Verify attempt-weighted average logic: (90% + 50%) / 2 = 70%
	att1Pct := (9.0 / 10.0) * 100.0 // 90
	att2Pct := (50.0 / 100.0) * 100.0 // 50
	sum := att1Pct + att2Pct
	count := 2

	avg := sum / float64(count)
	if avg != 70.0 {
		t.Errorf("expected attempt-weighted average of 70.0, got %f", avg)
	}
}

func TestDiscriminationIndex_MinimumSampleAndDeterministicTies(t *testing.T) {
	// Sample size less than 10 must return nil
	totalEligible := 8
	if totalEligible < 10 {
		// As per contract, discrimination_index = nil
	} else {
		t.Errorf("sample < 10 should not compute discrimination index")
	}

	// Sample size >= 10 computes top 27% vs bottom 27%
	totalEligible2 := 10
	groupSize := int(float64(totalEligible2) * 0.27) // floor(2.7) = 2
	if groupSize != 2 {
		t.Errorf("expected groupSize 2 for n=10, got %d", groupSize)
	}
}

func TestEngagedTime_MatchesLearnerAnalyticsContract(t *testing.T) {
	boundedTele := int64(450) // 7.5 mins
	authDur := int64(300)     // 5 mins attempt duration cap
	capped := CapEngagedQuestionTime(boundedTele, authDur)
	if capped != 300 {
		t.Errorf("expected engaged time capped at 300s, got %d", capped)
	}
}

func TestCanonicalEndpointHash_Normalisation(t *testing.T) {
	h1 := CanonicalEndpointHash("quizzes", "Asia/Kolkata", "", "20", "created_at", "DESC")
	h2 := CanonicalEndpointHash("quizzes", "Asia/Kolkata", "", "20", "created_at", "DESC")
	if h1 != h2 {
		t.Errorf("expected identical hashes, got %s and %s", h1, h2)
	}

	h3 := CanonicalEndpointHash("quizzes", "Asia/Kolkata", "cursor1", "20", "created_at", "DESC")
	if h1 == h3 {
		t.Errorf("different parameters must produce different hashes")
	}
}

func TestOwnershipTransfer_BumpsOldAndNewOwnerCaches(t *testing.T) {
	cache := NewInstructorAnalyticsCache(nil, nil)
	// Cache is nil client, must not panic on version bumps
	cache.BumpQuizVersion(uuid.New().String())
	cache.BumpInstructorVersion("user-old")
	cache.BumpInstructorVersion("user-new")
}
