/**
 * EXAM-P8-T01: client telemetry buffer for assessment analytics events.
 * Posts Category-3 events only; never mutates attempt scoring/state.
 */

export const CLIENT_TELEMETRY_EVENT_TYPES = new Set([
  "QUESTION_VIEWED",
  "ANSWER_SELECTED",
  "ANSWER_CHANGED",
  "QUESTION_TIME_SPENT",
  "HINT_OPENED",
  "REVIEW_OPENED",
]);

function newClientEventId() {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return crypto.randomUUID();
  }
  return `client-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

/**
 * Pure buffer factory (easy to unit test). Vue composable wraps this.
 */
export function createAssessmentAnalyticsBuffer({
  quizId,
  attemptId,
  correlationId,
  postBatch,
}) {
  const pending = [];

  function track(eventType, metadata = {}) {
    if (!CLIENT_TELEMETRY_EVENT_TYPES.has(eventType)) {
      return;
    }
    pending.push({
      client_event_id: newClientEventId(),
      event_type: eventType,
      metadata: metadata && typeof metadata === "object" ? metadata : {},
      occurred_at: new Date().toISOString(),
    });
  }

  function getPending() {
    return pending.slice();
  }

  async function flush() {
    if (!pending.length) {
      return { received: 0, inserted: 0, duplicates: 0, rejected: 0 };
    }
    const events = pending.splice(0, pending.length);
    const headers = {};
    if (correlationId) {
      headers["X-Correlation-ID"] = correlationId;
    }
    try {
      return await postBatch({ events }, headers);
    } catch (error) {
      // Restore on failure so callers can retry without losing telemetry.
      pending.unshift(...events);
      throw error;
    }
  }

  return {
    quizId,
    attemptId,
    track,
    flush,
    getPending,
  };
}

export function useAssessmentAnalytics(options = {}) {
  const { quizId, attemptId, correlationId = null } = options;
  const api = useAssessmentAttemptsApi();

  const resolvedCorrelationId =
    correlationId ||
    (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
      ? crypto.randomUUID()
      : `corr-${Date.now()}`);

  async function postBatch(payload, headers = {}) {
    return api.postAnalyticsEvents(quizId, attemptId, payload, headers);
  }

  return createAssessmentAnalyticsBuffer({
    quizId,
    attemptId,
    correlationId: resolvedCorrelationId,
    postBatch,
  });
}
