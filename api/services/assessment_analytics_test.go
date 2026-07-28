package services

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"go.uber.org/zap"
)

func TestValidateOccurredAt_Bounds(t *testing.T) {
	svc := &AssessmentAnalyticsService{logger: zap.NewNop()}
	created := time.Now().UTC().Add(-time.Hour)
	submitted := sql.NullTime{Time: time.Now().UTC().Add(-10 * time.Minute), Valid: true}

	if err := svc.ValidateOccurredAt(time.Now().UTC(), created, submitted); err != nil {
		t.Fatalf("expected valid occurred_at, got %v", err)
	}
	if err := svc.ValidateOccurredAt(time.Now().UTC().Add(10*time.Minute), created, submitted); err == nil {
		t.Fatal("expected future occurred_at rejection")
	}
	if err := svc.ValidateOccurredAt(created.Add(-10*time.Minute), created, submitted); err == nil {
		t.Fatal("expected too-early occurred_at rejection")
	}
	if err := svc.ValidateOccurredAt(submitted.Time.Add(analyticsAllowedReviewWindow+time.Minute), created, submitted); err == nil {
		t.Fatal("expected post-review-window rejection")
	}
}

func TestValidateMetadataSchema_SensitiveAndSize(t *testing.T) {
	svc := &AssessmentAnalyticsService{logger: zap.NewNop()}
	if err := svc.ValidateMetadataSchema(structs.EventQuestionViewed, map[string]interface{}{"question_id": "q1"}); err != nil {
		t.Fatalf("expected valid metadata: %v", err)
	}
	if err := svc.ValidateMetadataSchema(structs.EventQuestionViewed, map[string]interface{}{"password": "x"}); err == nil {
		t.Fatal("expected sensitive metadata rejection")
	}
	huge := map[string]interface{}{"blob": strings.Repeat("a", analyticsMaxMetadataBytes)}
	if err := svc.ValidateMetadataSchema(structs.EventQuestionViewed, huge); err == nil {
		t.Fatal("expected oversized metadata rejection")
	}
}

func TestRecordServerTelemetryBounded_DoesNotPanicOnFailure(t *testing.T) {
	svc := &AssessmentAnalyticsService{logger: zap.NewNop()}
	done := make(chan struct{})
	go func() {
		svc.RecordServerTelemetryBounded(context.Background(), AuthoritativeAnalyticsEventInput{
			AttemptID:     uuid.New(),
			QuizID:        uuid.New(),
			UserID:        "u1",
			EventType:     structs.EventAttemptAutosaved,
			EventSource:   structs.EventSourceHTTP,
			CorrelationID: "c1",
			OccurredAt:    time.Now().UTC(),
		})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("bounded telemetry did not return in time")
	}
}

func TestScheduledReleaseIdempotencyKey_Deterministic(t *testing.T) {
	quizID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	at := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	a := ScheduledReleaseIdempotencyKey(quizID, at)
	b := ScheduledReleaseIdempotencyKey(quizID, at)
	if a != b {
		t.Fatalf("expected deterministic key, got %q vs %q", a, b)
	}
	if !strings.HasPrefix(a, "scheduled-release:") {
		t.Fatalf("unexpected key format: %s", a)
	}
}
