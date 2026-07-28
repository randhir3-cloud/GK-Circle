import { mockNuxtImport } from "@nuxt/test-utils/runtime";
import { mount, flushPromises } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import McqQuestionEditor from "@/components/quiz-manage/McqQuestionEditor.vue";
import QuestionRevisionHistory from "@/components/quiz-manage/QuestionRevisionHistory.vue";

const fetchMock = vi.hoisted(() => vi.fn());
const listRevisionsMock = vi.hoisted(() => vi.fn());

vi.mock("notivue", () => ({
  usePush: vi.fn(() => ({
    success: vi.fn(),
    error: vi.fn(),
  })),
}));

vi.mock("@/composables/quiz_questions", () => ({
  useQuizQuestionsApi: () => ({
    listRevisions: listRevisionsMock,
  }),
}));

mockNuxtImport("useRuntimeConfig", () => () => ({
  public: { maxImageFileSize: 1024 * 1024 },
}));
mockNuxtImport("useNuxtApp", () => () => ({
  $validImageTypes: ["image/png"],
}));
mockNuxtImport("useRequestHeaders", () => () => ({ cookie: "session=test" }));
vi.stubGlobal("$fetch", fetchMock);

const baseQuestion = {
  question: "Capital of France?",
  question_media: "text",
  options: { 1: "Paris", 2: "Berlin", 3: "Madrid", 4: "Rome" },
  correct_answer: "[1]",
  question_type_id: 1,
  options_media: "text",
  official_answer: "[2]",
  authoritative_answer: "[3]",
  answer_review_status: "DISPUTED",
  revision_number: 2,
  lineage_id: "lineage-1",
};

describe("McqQuestionEditor", () => {
  beforeEach(() => {
    listRevisionsMock.mockReset();
    listRevisionsMock.mockResolvedValue([]);
  });

  it("emits authority payload with independent official and authoritative keys", async () => {
    const wrapper = mount(McqQuestionEditor, {
      props: {
        question: baseQuestion,
        mode: "edit",
        quizId: "quiz-1",
        questionId: "q-1",
      },
      global: {
        stubs: {
          NavigationLink: {
            template:
              "<button type='button' @click=\"$emit('click')\"><slot /></button>",
          },
          QuestionRevisionHistory: true,
        },
      },
    });

    expect(wrapper.find("#mcq-answer-review-status").element.value).toBe(
      "DISPUTED"
    );

    const saveButtons = wrapper.findAll("button");
    await saveButtons[saveButtons.length - 1].trigger("click");

    expect(wrapper.emitted("save")?.[0]?.[0]).toEqual({
      payload: expect.objectContaining({
        answers: [1],
        official_answer: [2],
        authoritative_answer: [3],
        answer_review_status: "DISPUTED",
      }),
    });
  });
});

describe("QuestionRevisionHistory", () => {
  beforeEach(() => {
    listRevisionsMock.mockReset();
  });

  it("loads and renders revision rows", async () => {
    listRevisionsMock.mockResolvedValue([
      {
        id: "rev-2",
        revision_number: 2,
        answer_review_status: "CONFIRMED",
        created_at: "2026-07-27T12:00:00Z",
      },
    ]);

    const wrapper = mount(QuestionRevisionHistory, {
      props: {
        quizId: "quiz-1",
        questionId: "q-1",
      },
    });

    await flushPromises();

    expect(listRevisionsMock).toHaveBeenCalledWith("quiz-1", "q-1");
    expect(wrapper.text()).toContain("Revision history");
    expect(wrapper.text()).toContain("#2");
    expect(wrapper.text()).toContain("CONFIRMED");
  });
});
