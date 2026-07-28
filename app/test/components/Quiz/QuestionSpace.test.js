import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import QuestionSpace from "~/components/Quiz/QuestionSpace.vue";
import constants from "~/config/constants";

vi.mock("notivue", () => ({
  usePush: vi.fn(() => ({
    error: vi.fn(),
    warning: vi.fn(),
  })),
}));

const buildProps = (overrides = {}) => ({
  data: {
    status: "success",
    data: {
      duration: 30,
      id: "1",
      no: 3,
      options: {
        1: "Scenic landscapes",
        2: "Cultural experiences",
        3: "Local cuisine",
        4: "Historical sites",
        5: "Adventure activities",
      },
      options_media: "text",
      question:
        "What features of a travel destination are most important to you when planning a vacation?",
      question_media: "text",
      question_time: "",
      quiz_id: "1",
      resource: "",
      totalJoinUser: 1,
      totalQuestions: 3,
      start_time: new Date().toISOString(),
    },
    event: constants.GetQuestion,
    action: "send single question to user",
    component: "Question",
  },
  isAdmin: false,
  canPlay: false,
  ...overrides,
});

const mountComponent = (props = buildProps()) =>
  mount(QuestionSpace, {
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
      },
      mocks: {
        $Fail: constants.Fail,
        $GetQuestion: constants.GetQuestion,
        $Counter: constants.Counter,
      },
    },
  });

describe("QuestionSpace test", () => {
  beforeEach(() => {
    vi.stubGlobal("useNuxtApp", () => ({
      $Fail: constants.Fail,
      $GetQuestion: constants.GetQuestion,
      $Counter: constants.Counter,
    }));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders correctly when a question is present", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("Question 3");
    expect(wrapper.get('[aria-label$="seconds remaining"]').exists()).toBe(
      true
    );
  });

  it("renders the question and options", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("h3").text()).toContain(
      "What features of a travel destination"
    );
    const optionButtons = wrapper
      .findAll('button[type="button"]')
      .filter((button) =>
        /Scenic|Cultural|Local|Historical|Adventure/.test(button.text())
      );
    expect(optionButtons.length).toBe(5);
  });

  it("emits sendAnswer when an answer is selected", async () => {
    const wrapper = mountComponent();
    const optionButtons = wrapper
      .findAll('button[type="button"]')
      .filter((button) => button.text().includes("Scenic landscapes"));
    await optionButtons[0].trigger("click");
    expect(wrapper.emitted("sendAnswer")[0]).toStrictEqual([[1]]);
  });

  it("disables options after submission", async () => {
    const wrapper = mountComponent();
    wrapper.vm.isSubmitted = true;
    await wrapper.vm.$nextTick();
    const optionButtons = wrapper
      .findAll('button[type="button"]')
      .filter((button) =>
        /Scenic|Cultural|Local|Historical|Adventure/.test(button.text())
      );
    optionButtons.forEach((button) => {
      expect(button.attributes("disabled")).toBeDefined();
    });
  });

  it("handles skipping the question", async () => {
    const wrapper = mountComponent(buildProps({ isAdmin: true }));
    wrapper.vm.handleSkip({ preventDefault: vi.fn() });
    expect(wrapper.emitted("askSkip")).toBeTruthy();
  });

  it("renders the countdown if no question is present", async () => {
    const wrapper = mountComponent(
      buildProps({
        data: {
          status: "success",
          data: { count: 3 },
          event: constants.Counter,
        },
      })
    );
    expect(wrapper.text()).toContain("3");
    expect(wrapper.text()).toMatch(/Get Ready|Go!|3/);
  });

  it("renders admin skip button when not last question", async () => {
    const wrapper = mountComponent(
      buildProps({
        isAdmin: true,
        data: {
          status: "success",
          event: constants.GetQuestion,
          data: {
            duration: 30,
            id: "1",
            no: 2,
            options: { 1: "A", 2: "B" },
            options_media: "text",
            question: "Q?",
            question_media: "text",
            quiz_id: "1",
            resource: "",
            totalQuestions: 3,
            start_time: new Date().toISOString(),
          },
        },
      })
    );
    const skipButton = wrapper
      .findAll("button")
      .find((button) => button.text().includes("Skip"));
    expect(skipButton).toBeTruthy();
  });

  it("renders admin finish button when last question", async () => {
    const wrapper = mountComponent(
      buildProps({
        isAdmin: true,
        data: {
          status: "success",
          event: constants.GetQuestion,
          data: {
            duration: 30,
            id: "1",
            no: 3,
            options: { 1: "A", 2: "B" },
            options_media: "text",
            question: "Q?",
            question_media: "text",
            quiz_id: "1",
            resource: "",
            totalQuestions: 3,
            start_time: new Date().toISOString(),
          },
        },
      })
    );
    const finishButton = wrapper
      .findAll("button")
      .find((button) => button.text().includes("Finish Quiz"));
    expect(finishButton).toBeTruthy();
  });
});
