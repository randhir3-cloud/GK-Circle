import { describe, it, expect, vi, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import QuestionAnalysis from "~/components/Quiz/QuestionAnalysis.vue";
import CodeBlockComponent from "~/components/CodeBlockComponent.vue";
import DeleteDialog from "~/components/DeleteDialog.vue";

const buildQuestion = (overrides = {}) => ({
  question_id: "1",
  question: "What is the color of the Strawberry?",
  type: 1,
  options: {
    1: "red",
    2: "pink",
    3: "black",
    4: "yellow",
  },
  question_media: "text",
  options_media: "text",
  resource: "",
  correct_answer: [1],
  selected_answers: { 1: ["husen_eSE1"] },
  duration: 30,
  avg_response_time: 6333,
  userParticipants: 1,
  correctPercentage: 100,
  ...overrides,
});

const mountComponent = (props = {}) =>
  mount(QuestionAnalysis, {
    props: {
      question: buildQuestion(),
      order: 1,
      isAdminAnalysis: true,
      isForQuiz: false,
      isEditable: "true",
      ...props,
    },
    global: {
      stubs: {
        DeleteDialog: {
          template: '<div class="delete-dialog-stub" />',
          props: ["modelValue", "title", "message"],
          emits: ["confirm-delete", "update:modelValue"],
        },
        CodeBlockComponent: true,
      },
    },
  });

describe("QuestionAnalysis", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders the question order and text", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("Question 1");
    expect(wrapper.find("h3").text()).toBe(
      "What is the color of the Strawberry?"
    );
  });

  it("renders an image if question_media is image with fetchable URL", () => {
    const wrapper = mountComponent({
      question: buildQuestion({
        resource: "https://example.com/test-image.jpg",
        question_media: "image",
        question: "lol",
      }),
    });
    const image = wrapper.find("img");
    expect(image.exists()).toBe(true);
    expect(image.attributes("src")).toBe("https://example.com/test-image.jpg");
  });

  it("renders a code block if question_media is code", () => {
    const wrapper = mountComponent({
      question: buildQuestion({
        resource: "console.log('Hello!');",
        question_media: "code",
      }),
    });
    expect(wrapper.findComponent(CodeBlockComponent).exists()).toBe(true);
  });

  it("renders admin analysis section", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("AVG.");
    expect(wrapper.text()).toContain("6.33");
    expect(wrapper.text()).toContain("M.C.Q.");
    expect(wrapper.text()).toContain("100% correct");
  });

  it("renders user response details for non-admin analysis", () => {
    const wrapper = mountComponent({
      isAdminAnalysis: false,
      question: buildQuestion({
        response_time: 1924,
        is_attend: true,
        selected_answer: { String: "[1]" },
      }),
    });
    expect(wrapper.text()).toContain("1.92");
    expect(wrapper.text()).toContain("Correct");
  });

  it("renders not attempted badge when user has not attempted the question", () => {
    const wrapper = mountComponent({
      isAdminAnalysis: false,
      question: buildQuestion({
        is_attend: false,
        selected_answer: { String: "" },
      }),
    });
    expect(wrapper.text()).toContain("Not Attempted");
  });

  it("renders edit and delete buttons for editable questions", async () => {
    const wrapper = mountComponent({ isEditable: "true" });
    const editButton = wrapper.find("button[title='Edit question']");
    expect(editButton.exists()).toBe(true);
    const deleteButton = wrapper.find("button[title='Delete question']");
    expect(deleteButton.exists()).toBe(true);
    expect(wrapper.findComponent(DeleteDialog).exists()).toBe(true);

    await editButton.trigger("click");
    expect(wrapper.emitted("editQuestion")[0]).toEqual(["1"]);
  });

  it("does not render edit and delete buttons if not editable", () => {
    const wrapper = mountComponent({ isEditable: "" });
    expect(wrapper.find("button[title='Edit question']").exists()).toBe(false);
    expect(wrapper.find("button[title='Delete question']").exists()).toBe(
      false
    );
  });

  it("handles missing question props gracefully", () => {
    const wrapper = mountComponent({ question: {} });
    expect(wrapper.find("h3").text()).toBe("");
    expect(wrapper.find("img").exists()).toBe(false);
  });
});
