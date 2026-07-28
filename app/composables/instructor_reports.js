export const useInstructorReports = () => {
  const config = useRuntimeConfig();
  const apiBase = config.public.apiBase || "/api/v1";

  // Standard fetch options with credentials
  const fetchOpts = (opts = {}) => ({
    headers: { "Content-Type": "application/json" },
    ...opts,
  });

  // 1. One-time export request
  const requestExport = async (payload) => {
    return await $fetch(
      `${apiBase}/instructor/reports/exports`,
      fetchOpts({
        method: "POST",
        body: payload,
      })
    );
  };

  // 2. Poll job status
  const getExportStatus = async (reportId) => {
    return await $fetch(
      `${apiBase}/instructor/reports/exports/${reportId}`,
      fetchOpts({
        method: "GET",
      })
    );
  };

  // 3. Download URL helper
  const getDownloadUrl = (reportId) => {
    return `${apiBase}/instructor/reports/exports/${reportId}/download`;
  };

  // 4. Delete / cancel report
  const deleteReport = async (reportId) => {
    return await $fetch(
      `${apiBase}/instructor/reports/exports/${reportId}`,
      fetchOpts({
        method: "DELETE",
      })
    );
  };

  // 5. History list (cursor-paginated)
  const getHistory = async (params = {}) => {
    const query = new URLSearchParams();
    if (params.cursor) query.set("cursor", params.cursor);
    if (params.limit) query.set("limit", String(params.limit));
    const qStr = query.toString() ? `?${query.toString()}` : "";
    return await $fetch(
      `${apiBase}/instructor/reports/history${qStr}`,
      fetchOpts({
        method: "GET",
      })
    );
  };

  // 6. Audit log
  const getAuditLog = async (params = {}) => {
    const query = new URLSearchParams();
    if (params.cursor) query.set("cursor", params.cursor);
    if (params.limit) query.set("limit", String(params.limit));
    const qStr = query.toString() ? `?${query.toString()}` : "";
    return await $fetch(
      `${apiBase}/instructor/reports/audit${qStr}`,
      fetchOpts({
        method: "GET",
      })
    );
  };

  // 7. Scheduled report CRUD
  const createSchedule = async (payload) => {
    return await $fetch(
      `${apiBase}/instructor/reports/schedules`,
      fetchOpts({
        method: "POST",
        body: payload,
      })
    );
  };

  const listSchedules = async () => {
    return await $fetch(
      `${apiBase}/instructor/reports/schedules`,
      fetchOpts({
        method: "GET",
      })
    );
  };

  const getSchedule = async (scheduleId) => {
    return await $fetch(
      `${apiBase}/instructor/reports/schedules/${scheduleId}`,
      fetchOpts({
        method: "GET",
      })
    );
  };

  const updateSchedule = async (scheduleId, payload) => {
    return await $fetch(
      `${apiBase}/instructor/reports/schedules/${scheduleId}`,
      fetchOpts({
        method: "PATCH",
        body: payload,
      })
    );
  };

  const deleteSchedule = async (scheduleId) => {
    return await $fetch(
      `${apiBase}/instructor/reports/schedules/${scheduleId}`,
      fetchOpts({
        method: "DELETE",
      })
    );
  };

  // 8. Per-quiz export operations
  const requestQuizExport = async (quizId, payload) => {
    return await $fetch(
      `${apiBase}/quizzes/${quizId}/reports/exports`,
      fetchOpts({
        method: "POST",
        body: payload,
      })
    );
  };

  const getQuizHistory = async (quizId, params = {}) => {
    const query = new URLSearchParams();
    if (params.cursor) query.set("cursor", params.cursor);
    if (params.limit) query.set("limit", String(params.limit));
    const qStr = query.toString() ? `?${query.toString()}` : "";
    return await $fetch(
      `${apiBase}/quizzes/${quizId}/reports/history${qStr}`,
      fetchOpts({
        method: "GET",
      })
    );
  };

  const deleteQuizReport = async (quizId, reportId) => {
    return await $fetch(
      `${apiBase}/quizzes/${quizId}/reports/exports/${reportId}`,
      fetchOpts({
        method: "DELETE",
      })
    );
  };

  return {
    requestExport,
    getExportStatus,
    getDownloadUrl,
    deleteReport,
    getHistory,
    getAuditLog,
    createSchedule,
    listSchedules,
    getSchedule,
    updateSchedule,
    deleteSchedule,
    requestQuizExport,
    getQuizHistory,
    deleteQuizReport,
  };
};
