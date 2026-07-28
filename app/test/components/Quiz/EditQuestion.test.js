import { mockNuxtImport } from "@nuxt/test-utils/runtime";

import { mount, flushPromises } from "@vue/test-utils";

import { beforeEach, describe, expect, it, vi } from "vitest";

import EditQuestion from "~/components/Quiz/EditQuestion.vue";

import constants from "~/test/constants";

const updateQuestionMock = vi.hoisted(() => vi.fn());

const listRevisionsMock = vi.hoisted(() => vi.fn());

vi.mock("notivue", () => ({
  usePush: vi.fn(() => ({
    success: vi.fn(),

    error: vi.fn(),
  })),
}));

vi.mock("@/composables/quiz_questions", () => ({
  getQuizQuestionAPIError: (_error, fallback) => fallback,

  useQuizQuestionsApi: () => ({
    updateQuestion: updateQuestionMock,

    listRevisions: listRevisionsMock,
  }),
}));

mockNuxtImport("useRuntimeConfig", () => () => ({
  public: {
    apiUrl: "http://api.test/api/v1",

    maxImageFileSize: 1024 * 1024,
  },
}));

mockNuxtImport("useRequestHeaders", () => () => ({ cookie: "session=test" }));

mockNuxtImport("useNuxtApp", () => () => ({
  $validImageTypes: ["image/png", "image/jpeg"],
}));

const baseQuestion = {
  question: "What is the capital of France?",

  question_media: "text",

  options: { 1: "Paris", 2: "Berlin", 3: "Madrid", 4: "Rome" },

  correct_answer: "[1]",

  question_id: "123",

  question_type_id: 1,

  options_media: "text",

  points: 1,

  duration_in_seconds: 30,

  lineage_id: "lineage-1",

  revision_number: 2,

  official_answer: "[1]",

  authoritative_answer: "[1]",

  answer_review_status: "CONFIRMED",

  answer_revision_reason: "",

  answer_revision_source: "",
};

const mountComponent = (question = baseQuestion) =>
  mount(EditQuestion, {
    props: {
      question,

      quizId: "quiz-123",

      questionId: "q-123",
    },

    global: {
      stubs: {
        VFileInput: constants.slotTemplate,

        CodeBlockComponent: true,

        NavigationLink: {
          template:
            "<button type='button' @click=\"$emit('click')\"><slot /></button>",
        },

        QuestionRevisionHistory: {
          template:
            "<div data-testid='revision-history'>Revision history</div>",
        },
      },
    },
  });

describe("EditQuestion.vue answer authority", () => {
  beforeEach(() => {
    updateQuestionMock.mockReset();

    updateQuestionMock.mockResolvedValue({});

    listRevisionsMock.mockReset();

    listRevisionsMock.mockResolvedValue([
      {
        id: "rev-2",

        revision_number: 2,

        answer_review_status: "CONFIRMED",

        created_at: "2026-07-27T12:00:00Z",
      },
    ]);
  });

  it("renders answer authority controls and revision history slot", async () => {
    const wrapper = mountComponent();

    await flushPromises();

    expect(wrapper.find("#mcq-answer-review-status").exists()).toBe(true);

    expect(wrapper.find("#mcq-answer-review-status").element.value).toBe(
      "CONFIRMED"
    );

    expect(wrapper.find("[data-testid='revision-history']").exists()).toBe(
      true
    );
  });

  it("includes authority fields when saving", async () => {
    const wrapper = mountComponent();

    await flushPromises();

    await wrapper.find("#mcq-answer-review-status").setValue("REVISED");

    await flushPromises();

    await wrapper
      .find("#mcq-answer-revision-reason")
      .setValue("Official key fixed");

    await wrapper

      .find("#mcq-answer-revision-source")

      .setValue("Commission notice");

    const saveButtons = wrapper.findAll("button");

    await saveButtons[saveButtons.length - 1].trigger("click");

    await flushPromises();

    expect(updateQuestionMock).toHaveBeenCalledWith(
      "quiz-123",

      "q-123",

      expect.objectContaining({
        answer_review_status: "REVISED",

        answer_revision_reason: "Official key fixed",

        answer_revision_source: "Commission notice",

        official_answer: [1],

        authoritative_answer: [1],
      })
    );
  });
});
