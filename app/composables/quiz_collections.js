const encodeID = (value) => encodeURIComponent(value);

export const getQuizCollectionAPIError = (error, fallback) =>
  error?.data?.data || error?.data?.message || error?.message || fallback;

export const useQuizCollectionsApi = () => {
  const { apiUrl } = useRuntimeConfig().public;
  const cookieHeaders = useRequestHeaders(["cookie"]);

  const collectionPath = (quizId, collectionId = "") => {
    const base = `/quizzes/${encodeID(quizId)}/collections`;
    return collectionId ? `${base}/${encodeID(collectionId)}` : base;
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

  const listCollections = (quizId) => request(collectionPath(quizId));

  const getCollection = (quizId, collectionId) =>
    request(collectionPath(quizId, collectionId));

  const createCollection = (quizId, payload) =>
    request(collectionPath(quizId), {
      method: "POST",
      body: payload,
    });

  const updateCollection = (quizId, collectionId, payload) =>
    request(collectionPath(quizId, collectionId), {
      method: "PATCH",
      body: payload,
    });

  const deleteCollection = (quizId, collectionId) =>
    request(collectionPath(quizId, collectionId), {
      method: "DELETE",
    });

  const replaceMembers = (quizId, collectionId, questionIds) =>
    request(`${collectionPath(quizId, collectionId)}/members`, {
      method: "PUT",
      body: { question_ids: questionIds },
    });

  const resolveCollection = (quizId, collectionId) =>
    request(`${collectionPath(quizId, collectionId)}/resolve`);

  return {
    listCollections,
    getCollection,
    createCollection,
    updateCollection,
    deleteCollection,
    replaceMembers,
    resolveCollection,
    getQuizCollectionAPIError,
  };
};
