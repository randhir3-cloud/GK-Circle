/**
 * EXAM-P8-T03 instructor analytics API client.
 */
export const getInstructorAnalyticsAPIError = (error, fallback) =>
  error?.data?.data || error?.data?.message || error?.message || fallback;

export const useInstructorAnalyticsApi = () => {
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

  const withTimezone = (path, timezone) => {
    if (!timezone) return path;
    const join = path.includes("?") ? "&" : "?";
    return `${path}${join}timezone=${encodeURIComponent(timezone)}`;
  };

  const timezone =
    typeof Intl !== "undefined"
      ? Intl.DateTimeFormat().resolvedOptions().timeZone
      : "Asia/Kolkata";

  return {
    timezone,
    getOverview: (tz = timezone) =>
      request(withTimezone("/instructor/analytics/overview", tz)),
    getQuizzes: ({
      cursor,
      limit,
      sortBy,
      sortDir,
      timezone: tz = timezone,
    } = {}) => {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      if (limit) params.set("limit", String(limit));
      if (sortBy) params.set("sort_by", sortBy);
      if (sortDir) params.set("sort_dir", sortDir);
      if (tz) params.set("timezone", tz);
      const q = params.toString();
      return request(`/instructor/analytics/quizzes${q ? `?${q}` : ""}`);
    },
    getLearners: ({
      cursor,
      limit,
      search,
      quizId,
      status,
      sortBy,
      sortDir,
      timezone: tz = timezone,
    } = {}) => {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      if (limit) params.set("limit", String(limit));
      if (search) params.set("q", search);
      if (quizId) params.set("quiz_id", quizId);
      if (status) params.set("status", status);
      if (sortBy) params.set("sort_by", sortBy);
      if (sortDir) params.set("sort_dir", sortDir);
      if (tz) params.set("timezone", tz);
      const q = params.toString();
      return request(`/instructor/analytics/learners${q ? `?${q}` : ""}`);
    },
    getReleases: ({
      cursor,
      limit,
      classification,
      timezone: tz = timezone,
    } = {}) => {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      if (limit) params.set("limit", String(limit));
      if (classification) params.set("classification", classification);
      if (tz) params.set("timezone", tz);
      const q = params.toString();
      return request(`/instructor/analytics/releases${q ? `?${q}` : ""}`);
    },
    getTimeline: ({ cursor, limit, timezone: tz = timezone } = {}) => {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      if (limit) params.set("limit", String(limit));
      if (tz) params.set("timezone", tz);
      const q = params.toString();
      return request(`/instructor/analytics/timeline${q ? `?${q}` : ""}`);
    },
    getQuizSummary: (quizId, tz = timezone) =>
      request(withTimezone(`/quizzes/${quizId}/analytics/summary`, tz)),
    getQuizAttempts: (
      quizId,
      { cursor, limit, status, timezone: tz = timezone } = {}
    ) => {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      if (limit) params.set("limit", String(limit));
      if (status) params.set("status", status);
      if (tz) params.set("timezone", tz);
      const q = params.toString();
      return request(
        `/quizzes/${quizId}/analytics/attempts${q ? `?${q}` : ""}`
      );
    },
    getQuestionMetrics: (
      quizId,
      { cursor, limit, timezone: tz = timezone } = {}
    ) => {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      if (limit) params.set("limit", String(limit));
      if (tz) params.set("timezone", tz);
      const q = params.toString();
      return request(
        `/quizzes/${quizId}/analytics/questions${q ? `?${q}` : ""}`
      );
    },
    getEngagement: (quizId, tz = timezone) =>
      request(withTimezone(`/quizzes/${quizId}/analytics/engagement`, tz)),
  };
};
