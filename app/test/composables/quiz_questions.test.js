import { mockNuxtImport } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getQuizQuestionAPIError,
  useQuizQuestionsApi,
} from "@/composables/quiz_questions";

const fetchMock = vi.hoisted(() => vi.fn());

mockNuxtImport("useRuntimeConfig", () => () => ({
  public: { apiUrl: "http://api.test/api/v1" },
}));
mockNuxtImport("useRequestHeaders", () => () => ({ cookie: "session=test" }));
vi.stubGlobal("$fetch", fetchMock);

beforeEach(() => {
  fetchMock.mockReset();
  fetchMock.mockResolvedValue({ data: { id: "q-1" } });
});

describe("quiz_questions composable", () => {
  it("creates and updates questions with authority fields", async () => {
    const api = useQuizQuestionsApi();

    await api.createQuestion("quiz 1", {
      question: "Stem",
      type: 1,
      options: { 1: "A", 2: "B" },
      answers: [1],
      official_answer: [2],
      authoritative_answer: [2],
      answer_review_status: "CONFIRMED",
      answer_revision_reason: "",
      answer_revision_source: "",
      question_media: "text",
      options_media: "text",
      resource: "",
    });

    await api.updateQuestion("quiz 1", "q 1", {
      question: "Stem updated",
      type: 1,
      options: { 1: "A", 2: "B" },
      answers: [2],
      official_answer: [2],
      authoritative_answer: [2],
      answer_review_status: "REVISED",
      answer_revision_reason: "Key fix",
      answer_revision_source: "Notice",
      question_media: "text",
      options_media: "text",
      resource: "",
      points: 10,
      duration_in_seconds: 30,
    });

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://api.test/api/v1/quizzes/quiz%201/questions",
      expect.objectContaining({
        method: "POST",
        body: expect.objectContaining({
          official_answer: [2],
          authoritative_answer: [2],
          answer_review_status: "CONFIRMED",
        }),
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://api.test/api/v1/quizzes/quiz%201/questions/q%201",
      expect.objectContaining({
        method: "PUT",
        body: expect.objectContaining({
          answer_review_status: "REVISED",
          answer_revision_reason: "Key fix",
        }),
      })
    );
  });

  it("lists revisions for a question", async () => {
    fetchMock.mockResolvedValueOnce({
      data: [{ id: "rev-1", revision_number: 1 }],
    });
    const api = useQuizQuestionsApi();

    const revisions = await api.listRevisions("quiz-1", "q-1");

    expect(revisions).toEqual([{ id: "rev-1", revision_number: 1 }]);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://api.test/api/v1/quizzes/quiz-1/questions/q-1/revisions",
      expect.objectContaining({ credentials: "include" })
    );
  });

  it("unwraps API errors", () => {
    expect(
      getQuizQuestionAPIError(
        { data: { data: "validation failed" } },
        "fallback"
      )
    ).toBe("validation failed");
    expect(getQuizQuestionAPIError(new Error("network"), "fallback")).toBe(
      "network"
    );
  });
});
