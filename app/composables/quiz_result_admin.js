export function useQuizResultAdminApi() {
  const { apiUrl } = useRuntimeConfig().public;
  const cookieHeaders = useRequestHeaders(["cookie"]);

  const request = async (path, options = {}) => {
    const response = await $fetch(`${apiUrl}${path}`, {
      credentials: "include",
      headers: {
        ...cookieHeaders,
        Accept: "application/json",
      },
      ...options,
    });
    return response?.data;
  };

  const getReleaseStatus = async (quizId) => {
    return await request(
      `/quizzes/${encodeURIComponent(quizId)}/results/status`
    );
  };

  const updateResultSettings = async (quizId, settings) => {
    return await request(
      `/quizzes/${encodeURIComponent(quizId)}/results/settings`,
      {
        method: "PATCH",
        body: settings,
      }
    );
  };

  const releaseResults = async (quizId) => {
    return await request(
      `/quizzes/${encodeURIComponent(quizId)}/results/release`,
      {
        method: "POST",
      }
    );
  };

  return {
    getReleaseStatus,
    updateResultSettings,
    releaseResults,
  };
}
