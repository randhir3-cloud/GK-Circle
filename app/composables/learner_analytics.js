/**
 * EXAM-P8-T02 learner analytics API client.
 */
export const getLearnerAnalyticsAPIError = (error, fallback) =>
  error?.data?.data || error?.data?.message || error?.message || fallback;

export const useLearnerAnalyticsApi = () => {
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
      : null;

  return {
    timezone,
    getDashboard: (tz = timezone) =>
      request(withTimezone("/analytics/dashboard", tz)),
    getActivity: ({ limit, cursor, timezone: tz = timezone } = {}) => {
      const params = new URLSearchParams();
      if (limit) params.set("limit", String(limit));
      if (cursor) params.set("cursor", cursor);
      if (tz) params.set("timezone", tz);
      const query = params.toString();
      return request(`/analytics/activity${query ? `?${query}` : ""}`);
    },
    getTrends: ({
      granularity = "daily",
      from,
      to,
      timezone: tz = timezone,
    } = {}) => {
      const params = new URLSearchParams();
      params.set("granularity", granularity);
      if (from) params.set("from", from);
      if (to) params.set("to", to);
      if (tz) params.set("timezone", tz);
      return request(`/analytics/trends?${params.toString()}`);
    },
    getSubjects: (tz = timezone) =>
      request(withTimezone("/analytics/subjects", tz)),
    getAttemptTimeline: (attemptId, tz = timezone) =>
      request(
        withTimezone(
          `/analytics/attempts/${encodeURIComponent(attemptId)}/timeline`,
          tz
        )
      ),
  };
};
