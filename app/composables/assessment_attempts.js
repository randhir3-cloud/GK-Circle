const encodeID = (value) => encodeURIComponent(value);

export const getAssessmentAttemptAPIError = (error, fallback) =>
  error?.data?.data || error?.data?.message || error?.message || fallback;

export const useAssessmentAttemptsApi = () => {
  const { apiUrl } = useRuntimeConfig().public;
  const cookieHeaders = useRequestHeaders(["cookie"]);

  const request = async (path, options = {}) => {
    const { headers: optionHeaders, ...rest } = options;
    const response = await $fetch(`${apiUrl}${path}`, {
      credentials: "include",
      headers: {
        ...cookieHeaders,
        Accept: "application/json",
        ...(optionHeaders || {}),
      },
      ...rest,
    });
    return response?.data;
  };

  const attemptBase = (quizId) => `/quizzes/${encodeID(quizId)}/attempts`;

  return {
    getInstructions: (quizId, snapshotId) =>
      request(
        `${attemptBase(quizId)}/instructions?snapshot_id=${encodeID(
          snapshotId
        )}`
      ),
    listMine: (quizId) => request(attemptBase(quizId)),
    createAttempt: (quizId, snapshotId) =>
      request(attemptBase(quizId), {
        method: "POST",
        body: { snapshot_id: snapshotId },
      }),
    getAttempt: (quizId, attemptId) =>
      request(`${attemptBase(quizId)}/${encodeID(attemptId)}`),
    resumeAttempt: (quizId, attemptId) =>
      request(`${attemptBase(quizId)}/${encodeID(attemptId)}/resume`),
    getAttemptStatus: (quizId, attemptId) =>
      request(`${attemptBase(quizId)}/${encodeID(attemptId)}/status`),
    getAttemptResult: (quizId, attemptId) =>
      request(`${attemptBase(quizId)}/${encodeID(attemptId)}/result`),
    submitAttempt: (quizId, attemptId) =>
      request(`${attemptBase(quizId)}/${encodeID(attemptId)}/submit`, {
        method: "POST",
      }),
    autosaveAnswer: (quizId, attemptId, questionId, body) =>
      request(
        `${attemptBase(quizId)}/${encodeID(attemptId)}/answers/${encodeID(
          questionId
        )}`,
        {
          method: "PUT",
          body,
        }
      ),
    postAnalyticsEvents: (quizId, attemptId, body, headers = {}) =>
      request(
        `${attemptBase(quizId)}/${encodeID(attemptId)}/analytics/events`,
        {
          method: "POST",
          body,
          headers,
        }
      ),
    listAnalyticsEvents: (quizId, attemptId, { limit, cursor } = {}) => {
      const params = new URLSearchParams();
      if (limit) params.set("limit", String(limit));
      if (cursor) params.set("cursor", cursor);
      const query = params.toString();
      return request(
        `${attemptBase(quizId)}/${encodeID(attemptId)}/analytics/events${
          query ? `?${query}` : ""
        }`
      );
    },
  };
};

export const formatDurationSeconds = (seconds) => {
  if (seconds == null || Number.isNaN(Number(seconds))) return "Untimed";
  const total = Math.max(0, Math.floor(Number(seconds)));
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const secs = total % 60;
  if (hours > 0) return `${hours}h ${minutes}m`;
  if (minutes > 0) return `${minutes} min`;
  return `${secs} sec`;
};
