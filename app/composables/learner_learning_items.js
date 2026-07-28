const encodeID = (value) => encodeURIComponent(value);

export const getLearnerLearningItemAPIError = (error, fallback) =>
  error?.data?.data || error?.data?.message || error?.message || fallback;

export const isCourseEnrollmentRequiredError = (errorOrMessage) => {
  const message =
    typeof errorOrMessage === "string"
      ? errorOrMessage
      : getLearnerLearningItemAPIError(errorOrMessage, "");
  return message === "course enrollment required";
};

export const useLearnerLearningItemsApi = () => {
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

  const itemBase = (courseId, nodeId) =>
    `/learner/courses/${encodeID(courseId)}/nodes/${encodeID(
      nodeId
    )}/learning-items`;

  const enrollmentPath = (courseId) =>
    `/learner/courses/${encodeID(courseId)}/enrollment`;

  return {
    listItems: (courseId, nodeId) => request(itemBase(courseId, nodeId)),
    getItem: (courseId, nodeId, itemId) =>
      request(`${itemBase(courseId, nodeId)}/${encodeID(itemId)}`),
    getEnrollment: (courseId) => request(enrollmentPath(courseId)),
    enroll: (courseId) => request(enrollmentPath(courseId), { method: "POST" }),
    unenroll: (courseId) =>
      request(enrollmentPath(courseId), { method: "DELETE" }),
    listPublishedCourses: () => request("/learner/courses"),
    getPublishedCourse: (courseId) =>
      request(`/learner/courses/${encodeID(courseId)}`),
    getPublishedOutline: (courseId) =>
      request(`/learner/courses/${encodeID(courseId)}/nodes/tree`),
  };
};
