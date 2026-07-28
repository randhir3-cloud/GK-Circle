import { buildAuthorityPayload } from "@/utils/question_authority";

const encodeID = (value) => encodeURIComponent(value);

export const getQuizQuestionAPIError = (error, fallback) =>
  error?.data?.data || error?.data?.message || error?.message || fallback;

export const useQuizQuestionsApi = () => {
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

  const questionPath = (quizId, questionId = "") => {
    const base = `/quizzes/${encodeID(quizId)}/questions`;
    return questionId ? `${base}/${encodeID(questionId)}` : base;
  };

  const buildQuestionBody = (payload) => {
    const authority = buildAuthorityPayload({
      answers: payload.answers,
      officialAnswerKeys: payload.official_answer,
      authoritativeAnswerKeys: payload.authoritative_answer,
      answerReviewStatus: payload.answer_review_status,
      answerRevisionReason: payload.answer_revision_reason,
      answerRevisionSource: payload.answer_revision_source,
    });

    return {
      question: payload.question,
      type: payload.type,
      options: payload.options,
      answers: authority.answers,
      official_answer: authority.official_answer,
      authoritative_answer: authority.authoritative_answer,
      answer_review_status: authority.answer_review_status,
      answer_revision_reason: authority.answer_revision_reason,
      answer_revision_source: authority.answer_revision_source,
      points: payload.points,
      duration_in_seconds: payload.duration_in_seconds,
      question_media: payload.question_media,
      options_media: payload.options_media,
      resource: payload.resource,
    };
  };

  return {
    getQuestion: (quizId, questionId) =>
      request(questionPath(quizId, questionId)),
    listRevisions: (quizId, questionId) =>
      request(`${questionPath(quizId, questionId)}/revisions`),
    createQuestion: (quizId, payload) =>
      request(questionPath(quizId), {
        method: "POST",
        body: buildQuestionBody(payload),
      }),
    updateQuestion: (quizId, questionId, payload) =>
      request(questionPath(quizId, questionId), {
        method: "PUT",
        body: buildQuestionBody(payload),
      }),
    buildQuestionBody,
  };
};
