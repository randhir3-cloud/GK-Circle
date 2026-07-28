const encodeID = (value) => encodeURIComponent(value);

export const getCourseAdminAPIError = (error, fallback) =>
  error?.data?.data || error?.data?.message || error?.message || fallback;

export const useCourseLearningItemsApi = () => {
  const { apiUrl } = useRuntimeConfig().public;
  const cookieHeaders = useRequestHeaders(["cookie"]);

  const request = (path, options = {}) =>
    $fetch(`${apiUrl}${path}`, {
      credentials: "include",
      headers: {
        ...cookieHeaders,
        Accept: "application/json",
      },
      ...options,
    });

  const data = async (path, options) => {
    const response = await request(path, options);
    return response?.data;
  };

  return {
    listCourses: () => data("/admin/courses"),
    listQuizzes: () => data("/quizzes"),
    createCourse: (body) => data("/admin/courses", { method: "POST", body }),
    updateCourse: (courseId, body) =>
      data(`/admin/courses/${encodeID(courseId)}`, {
        method: "PATCH",
        body,
      }),
    getCourse: (courseId) => data(`/admin/courses/${encodeID(courseId)}`),
    listRootNodes: (courseId) =>
      data(`/admin/courses/${encodeID(courseId)}/nodes`),
    getTree: (courseId) =>
      data(`/admin/courses/${encodeID(courseId)}/nodes/tree`),
    createNode: (courseId, body) =>
      data(`/admin/courses/${encodeID(courseId)}/nodes`, {
        method: "POST",
        body,
      }),
    listChildren: (courseId, nodeId) =>
      data(
        `/admin/courses/${encodeID(courseId)}/nodes/${encodeID(
          nodeId
        )}/children`
      ),
    listItems: (courseId, nodeId) =>
      data(
        `/admin/courses/${encodeID(courseId)}/nodes/${encodeID(
          nodeId
        )}/learning-items`
      ),
    createItem: (courseId, nodeId, body) =>
      data(
        `/admin/courses/${encodeID(courseId)}/nodes/${encodeID(
          nodeId
        )}/learning-items`,
        { method: "POST", body }
      ),
    updateItem: (courseId, nodeId, itemId, body) =>
      data(
        `/admin/courses/${encodeID(courseId)}/nodes/${encodeID(
          nodeId
        )}/learning-items/${encodeID(itemId)}`,
        { method: "PATCH", body }
      ),
    deleteItem: (courseId, nodeId, itemId) =>
      data(
        `/admin/courses/${encodeID(courseId)}/nodes/${encodeID(
          nodeId
        )}/learning-items/${encodeID(itemId)}`,
        { method: "DELETE" }
      ),
  };
};
