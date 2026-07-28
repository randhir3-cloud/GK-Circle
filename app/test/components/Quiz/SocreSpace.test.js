import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import ScoreSpace from "~/components/Quiz/ScoreSpace.vue";

vi.mock("~~/composables/avatar", () => ({
  getAvatarUrlByName: vi
    .fn()
    .mockImplementation(
      (name) => `https://api.dicebear.com/9.x/bottts/svg?seed=${name || "Eden"}`
    ),
}));

const buildProps = (overrides = {}) => ({
  data: {
    status: "success",
    data: {
      duration: 20,
      options: {
        1: { value: "Blue", isAnswer: false },
        2: { value: "Red", isAnswer: true },
        3: { value: "Green", isAnswer: false },
        4: { value: "Pink", isAnswer: false },
      },
      options_media: "text",
      question: "What is the color of the Strawberry?",
      question_media: "text",
      quiz_id: "1",
      question_no: 1,
      totalQuestions: 5,
      rankList: [
        {
          rank: 1,
          points: 0,
          score: 0,
          response_time: 3390,
          username: "john",
          firstname: "doe",
          img_key: "Eden",
          streak_count: 0,
        },
      ],
      resource: "",
      userResponses: [
        {
          id: "user1",
          answers: {
            String: "[3]",
            Valid: true,
          },
        },
      ],
    },
    event: "show_score",
    action: "show score page during quiz",
    component: "Score",
  },
  isAdmin: true,
  userName: "",
  selectedAnswer: 0,
  analysisTab: "ranking",
  quizState: "running",
  ...overrides,
});

const mountComponent = (props = buildProps()) =>
  mount(ScoreSpace, {
    props,
    global: {
      plugins: [
        createTestingPinia({
          createSpy: vi.fn,
          initialState: {
            music: { music: false },
          },
        }),
      ],
      stubs: {
        CodeBlockComponent: true,
        AnswerSubmissionChart: true,
      },
    },
  });

describe("ScoreSpace test", () => {
  it("renders correctly with default props", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("h3").text()).toContain(
      "What is the color of the Strawberry?"
    );
    expect(wrapper.find('[aria-hidden="true"]').exists()).toBe(true);
  });

  it("renders the correct answer and selected incorrect answer", async () => {
    const wrapper = mountComponent(
      buildProps({
        selectedAnswer: 3,
        isAdmin: false,
      })
    );
    expect(wrapper.get('[aria-label="Correct answer"]').exists()).toBe(true);
    expect(wrapper.text()).toContain("Red");
    expect(wrapper.get('[aria-label="Your incorrect answer"]').exists()).toBe(
      true
    );
    expect(wrapper.text()).toContain("Green");
  });

  it("emits askSkipTimer when the skip button is clicked", async () => {
    const wrapper = mountComponent();
    const skipButton = wrapper
      .findAll("button")
      .find((button) => button.text().includes("Skip"));
    expect(skipButton).toBeTruthy();
    await skipButton.trigger("click");
    expect(wrapper.emitted("askSkipTimer")).toBeTruthy();
    expect(skipButton.attributes("disabled")).toBeDefined();
  });

  it("changes the analysis tab when a tab is clicked", async () => {
    const wrapper = mountComponent();
    const chartTab = wrapper
      .findAll("button")
      .find((button) => button.text().includes("Chart"));
    await chartTab.trigger("click");
    expect(wrapper.emitted("changeAnalysisTab")[0]).toEqual(["chart"]);
  });

  it("renders the rank list correctly", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("doe");
    expect(wrapper.text()).toContain("1");
  });
});
