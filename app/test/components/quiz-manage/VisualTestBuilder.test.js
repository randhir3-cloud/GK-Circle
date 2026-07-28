import { mount, flushPromises } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import VisualTestBuilder from "@/components/quiz-manage/VisualTestBuilder.vue";

const listCollections = vi.fn();
const getCollection = vi.fn();
const createCollection = vi.fn();
const updateCollection = vi.fn();
const deleteCollection = vi.fn();
const replaceMembers = vi.fn();
const resolveCollection = vi.fn();
const createQuestion = vi.fn();

vi.mock("@/composables/quiz_collections", () => ({
  useQuizCollectionsApi: () => ({
    listCollections,
    getCollection,
    createCollection,
    updateCollection,
    deleteCollection,
    replaceMembers,
    resolveCollection,
    getQuizCollectionAPIError: (_error, fallback) => fallback,
  }),
}));

vi.mock("@/composables/quiz_questions", () => ({
  useQuizQuestionsApi: () => ({
    createQuestion,
  }),
  getQuizQuestionAPIError: (_error, fallback) => fallback,
}));

const bankQuestions = [
  { question_id: "q-1", question: "Capital of France?" },
  { question_id: "q-2", question: "Who wrote Arthashastra?" },
];

describe("VisualTestBuilder", () => {
  beforeEach(() => {
    listCollections.mockReset();
    getCollection.mockReset();
    createCollection.mockReset();
    updateCollection.mockReset();
    deleteCollection.mockReset();
    replaceMembers.mockReset();
    resolveCollection.mockReset();
    createQuestion.mockReset();

    listCollections.mockResolvedValue([
      {
        id: "c-static",
        title: "Section A",
        kind: "STATIC",
        position: 0,
        members: [{ question_id: "q-1", position: 0 }],
      },
      {
        id: "c-dynamic",
        title: "History pool",
        kind: "DYNAMIC",
        position: 1,
        filter: { subject: "History" },
      },
    ]);
    getCollection.mockImplementation(async (_quizId, collectionId) => {
      if (collectionId === "c-dynamic") {
        return {
          id: "c-dynamic",
          title: "History pool",
          kind: "DYNAMIC",
          position: 1,
          filter: { subject: "History" },
        };
      }
      return {
        id: "c-static",
        title: "Section A",
        kind: "STATIC",
        position: 0,
        members: [{ question_id: "q-1", position: 0 }],
      };
    });
    resolveCollection.mockResolvedValue({
      collection_id: "c-static",
      kind: "STATIC",
      resolution_status: "RESOLVED",
      question_ids: ["q-1"],
    });
  });

  const mountBuilder = async (props = {}) => {
    const wrapper = mount(VisualTestBuilder, {
      props: {
        quizId: "quiz-1",
        questions: bankQuestions,
        canEdit: true,
        defaultPoints: 5,
        defaultDuration: 30,
        ...props,
      },
      global: {
        stubs: {
          DynamicFilterFields: {
            props: ["modelValue"],
            template: `<div data-testid="dynamic-filter-fields" />`,
          },
          QuestionFormCard: {
            props: ["mode", "quizId", "saving"],
            emits: ["save", "cancel"],
            template: `
              <div data-testid="question-form-card">
                <button type="button" data-testid="stub-save" @click="$emit('save', {
                  payload: {
                    question: 'New inline question?',
                    type: 1,
                    options: { 1: 'A', 2: 'B' },
                    answers: [1],
                    official_answer: [1],
                    authoritative_answer: [1],
                    answer_review_status: 'UNREVIEWED',
                    answer_revision_reason: '',
                    answer_revision_source: '',
                    question_media: 'text',
                    options_media: 'text',
                    resource: '',
                  }
                })">Save</button>
              </div>
            `,
          },
        },
      },
    });
    await flushPromises();
    return wrapper;
  };

  it("loads collections and shows STATIC ordered membership", async () => {
    const wrapper = await mountBuilder();

    expect(wrapper.text()).toContain("Visual Test Builder");
    expect(wrapper.text()).toContain("Section A");
    expect(wrapper.text()).toContain("Ordered STATIC membership");
    expect(wrapper.find('[data-testid="static-member-list"]').text()).toContain(
      "Capital of France?"
    );
  });

  it("shows DYNAMIC resolve preview with metadata-pending messaging", async () => {
    resolveCollection.mockResolvedValue({
      collection_id: "c-dynamic",
      kind: "DYNAMIC",
      resolution_status: "METADATA_PENDING",
      message:
        "question bank taxonomy metadata is not yet available; filter criteria are stored but not resolved",
      question_ids: [],
    });

    const wrapper = await mountBuilder();
    await wrapper
      .findAll("button")
      .find((button) => button.text().includes("History pool"))
      ?.trigger("click");
    await flushPromises();

    expect(wrapper.text()).toContain("DYNAMIC collection");
    expect(wrapper.text()).toContain(
      "DYNAMIC filters stored — taxonomy resolution pending"
    );
    expect(wrapper.text()).toContain("Dynamically resolved question IDs");
    expect(wrapper.find('[data-testid="dynamic-filter-fields"]').exists()).toBe(
      true
    );
  });

  it("creates a STATIC collection through the T01 API", async () => {
    createCollection.mockResolvedValue({
      id: "c-new",
      title: "New section",
      kind: "STATIC",
      position: 2,
    });
    getCollection.mockResolvedValue({
      id: "c-new",
      title: "New section",
      kind: "STATIC",
      position: 2,
      members: [],
    });
    resolveCollection.mockResolvedValue({
      collection_id: "c-new",
      kind: "STATIC",
      resolution_status: "RESOLVED",
      question_ids: [],
    });

    const wrapper = await mountBuilder();
    await wrapper
      .findAll("button")
      .find((button) => button.text().includes("Static collection"))
      ?.trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("New STATIC collection");
    await wrapper.find('input[type="text"]').setValue("New section");
    await wrapper.find("form").trigger("submit.prevent");
    await flushPromises();

    expect(createCollection).toHaveBeenCalledWith(
      "quiz-1",
      expect.objectContaining({
        title: "New section",
        kind: "STATIC",
      })
    );
  });

  it("saves STATIC membership via replaceMembers", async () => {
    replaceMembers.mockResolvedValue({
      id: "c-static",
      title: "Section A",
      kind: "STATIC",
      position: 0,
      members: [
        { question_id: "q-1", position: 0 },
        { question_id: "q-2", position: 1 },
      ],
    });
    resolveCollection.mockResolvedValue({
      collection_id: "c-static",
      kind: "STATIC",
      resolution_status: "RESOLVED",
      question_ids: ["q-1", "q-2"],
    });

    const wrapper = await mountBuilder();
    wrapper.vm.addQuestionId = "q-2";
    wrapper.vm.addMember();
    await wrapper.vm.$nextTick();
    await wrapper
      .findAll("button")
      .find((button) => button.text().includes("Save membership"))
      ?.trigger("click");
    await flushPromises();

    expect(replaceMembers).toHaveBeenCalledWith("quiz-1", "c-static", [
      "q-1",
      "q-2",
    ]);
  });

  it("creates inline questions through the shared Question Bank API and links STATIC membership", async () => {
    createQuestion.mockResolvedValue("q-new");
    replaceMembers.mockResolvedValue({
      id: "c-static",
      title: "Section A",
      kind: "STATIC",
      position: 0,
      members: [
        { question_id: "q-1", position: 0 },
        { question_id: "q-new", position: 1 },
      ],
    });
    resolveCollection.mockResolvedValue({
      collection_id: "c-static",
      kind: "STATIC",
      resolution_status: "RESOLVED",
      question_ids: ["q-1", "q-new"],
    });

    const wrapper = await mountBuilder();
    await wrapper.find('[data-testid="inline-add-question"]').trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.find('[data-testid="inline-question-form"]').exists()).toBe(
      true
    );
    expect(
      wrapper.find('[data-testid="link-to-static-collection"]').exists()
    ).toBe(true);

    await wrapper.find('[data-testid="stub-save"]').trigger("click");
    await flushPromises();

    expect(createQuestion).toHaveBeenCalledWith(
      "quiz-1",
      expect.objectContaining({
        question: "New inline question?",
        points: 5,
        duration_in_seconds: 30,
        answer_review_status: "UNREVIEWED",
      })
    );
    expect(replaceMembers).toHaveBeenCalledWith("quiz-1", "c-static", [
      "q-1",
      "q-new",
    ]);
    expect(wrapper.emitted("question-created")?.[0]?.[0]).toEqual({
      questionId: "q-new",
      linkedToCollection: true,
      collectionId: "c-static",
    });
  });

  it("creates inline questions into the bank only for DYNAMIC collections", async () => {
    createQuestion.mockResolvedValue("q-dyn");
    resolveCollection.mockResolvedValue({
      collection_id: "c-dynamic",
      kind: "DYNAMIC",
      resolution_status: "METADATA_PENDING",
      question_ids: [],
    });

    const wrapper = await mountBuilder();
    await wrapper
      .findAll("button")
      .find((button) => button.text().includes("History pool"))
      ?.trigger("click");
    await flushPromises();

    await wrapper.find('[data-testid="inline-add-question"]').trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("linked to the quiz bank only");
    await wrapper.find('[data-testid="stub-save"]').trigger("click");
    await flushPromises();

    expect(createQuestion).toHaveBeenCalled();
    expect(replaceMembers).not.toHaveBeenCalled();
    expect(wrapper.emitted("question-created")?.[0]?.[0]).toEqual({
      questionId: "q-dyn",
      linkedToCollection: false,
      collectionId: null,
    });
  });
});
