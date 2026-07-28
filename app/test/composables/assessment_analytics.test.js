import { describe, expect, it, vi, beforeEach } from "vitest";
import {
  createAssessmentAnalyticsBuffer,
  CLIENT_TELEMETRY_EVENT_TYPES,
} from "~/composables/assessment_analytics";

describe("assessment_analytics composable helpers", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("buffers client telemetry events with client_event_id and occurred_at", () => {
    const buffer = createAssessmentAnalyticsBuffer({
      quizId: "quiz-1",
      attemptId: "attempt-1",
      correlationId: "corr-1",
      postBatch: vi.fn(),
    });

    buffer.track("QUESTION_VIEWED", { question_id: "q1" });
    buffer.track("ANSWER_SELECTED", { question_id: "q1", option_id: 2 });

    const pending = buffer.getPending();
    expect(pending).toHaveLength(2);
    expect(CLIENT_TELEMETRY_EVENT_TYPES.has(pending[0].event_type)).toBe(true);
    expect(pending[0].client_event_id).toBeTruthy();
    expect(pending[0].occurred_at).toBeTruthy();
    expect(pending[0].metadata.question_id).toBe("q1");
  });

  it("flushes buffered events to batch endpoint with correlation header", async () => {
    const postBatch = vi.fn().mockResolvedValue({
      received: 1,
      inserted: 1,
      duplicates: 0,
      rejected: 0,
    });
    const buffer = createAssessmentAnalyticsBuffer({
      quizId: "quiz-1",
      attemptId: "attempt-1",
      correlationId: "corr-xyz",
      postBatch,
    });

    buffer.track("HINT_OPENED", { question_id: "q2" });
    const result = await buffer.flush();

    expect(postBatch).toHaveBeenCalledTimes(1);
    const [payload, headers] = postBatch.mock.calls[0];
    expect(payload.events).toHaveLength(1);
    expect(headers["X-Correlation-ID"]).toBe("corr-xyz");
    expect(result.inserted).toBe(1);
    expect(buffer.getPending()).toHaveLength(0);
  });

  it("ignores non-client event types", () => {
    const buffer = createAssessmentAnalyticsBuffer({
      quizId: "quiz-1",
      attemptId: "attempt-1",
      correlationId: "corr-1",
      postBatch: vi.fn(),
    });
    buffer.track("ATTEMPT_STARTED", {});
    expect(buffer.getPending()).toHaveLength(0);
  });
});
