const encodeID = (value) => encodeURIComponent(value);

export const getQuizImportJobAPIError = (error, fallback) =>
  error?.data?.data || error?.data?.message || error?.message || fallback;

export const useQuizImportJobsApi = () => {
  const { apiUrl } = useRuntimeConfig().public;
  const cookieHeaders = useRequestHeaders(["cookie"]);

  const importJobPath = (quizId, jobId = "") => {
    const base = `/quizzes/${encodeID(quizId)}/questions/import-jobs`;
    return jobId ? `${base}/${encodeID(jobId)}` : base;
  };

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

  const createPreviewJob = async (quizId, file) => {
    const formData = new FormData();
    formData.append("attachment", file);

    return request(importJobPath(quizId), {
      method: "POST",
      body: formData,
    });
  };

  const getPreviewJob = (quizId, jobId) =>
    request(importJobPath(quizId, jobId));

  const commitPreviewJob = (quizId, jobId) =>
    request(`${importJobPath(quizId, jobId)}/commit`, {
      method: "POST",
    });

  return {
    createPreviewJob,
    getPreviewJob,
    commitPreviewJob,
    getQuizImportJobAPIError,
  };
};
