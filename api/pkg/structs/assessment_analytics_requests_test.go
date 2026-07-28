package structs

import "testing"

func TestAssessmentEventTypeClassification(t *testing.T) {
	authoritative := []AssessmentEventType{
		EventAttemptStarted,
		EventAttemptSubmitted,
		EventAttemptAutoSubmitted,
		EventResultReleaseOverrideApplied,
		EventResultReleaseScheduledEffective,
	}
	for _, eventType := range authoritative {
		if !IsAuthoritativeEventType(eventType) {
			t.Fatalf("expected authoritative: %s", eventType)
		}
		if IsServerTelemetryEventType(eventType) || IsClientTelemetryEventType(eventType) {
			t.Fatalf("authoritative type must not be telemetry: %s", eventType)
		}
	}

	server := []AssessmentEventType{EventAttemptAutosaved, EventResultViewed}
	for _, eventType := range server {
		if !IsServerTelemetryEventType(eventType) {
			t.Fatalf("expected server telemetry: %s", eventType)
		}
		if IsAuthoritativeEventType(eventType) || IsClientTelemetryEventType(eventType) {
			t.Fatalf("server telemetry must not overlap: %s", eventType)
		}
	}

	client := []AssessmentEventType{
		EventQuestionViewed,
		EventAnswerSelected,
		EventAnswerChanged,
		EventQuestionTimeSpent,
		EventHintOpened,
		EventReviewOpened,
	}
	for _, eventType := range client {
		if !IsClientTelemetryEventType(eventType) {
			t.Fatalf("expected client telemetry: %s", eventType)
		}
		if IsAuthoritativeEventType(eventType) || IsServerTelemetryEventType(eventType) {
			t.Fatalf("client telemetry must not overlap: %s", eventType)
		}
	}
}
