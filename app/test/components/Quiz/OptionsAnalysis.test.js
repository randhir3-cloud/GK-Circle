import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import OptionsAnalysis from "~/components/Quiz/OptionsAnalysis.vue";

const mountComponent = (props = {}) =>
  mount(OptionsAnalysis, {
    props: {
      options: { 1: "Paris", 2: "Rome", 3: "Athens", 4: "Cairo" },
      selectedAnswers: { 3: ["lol_yJc3"] },
      correctAnswer: [2],
      optionsMedia: "text",
      isAdminAnalysis: true,
      ...props,
    },
  });

describe("OptionsAnalysis test", () => {
  it("renders all options", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("Paris");
    expect(wrapper.text()).toContain("Rome");
    expect(wrapper.text()).toContain("Athens");
    expect(wrapper.text()).toContain("Cairo");
  });

  it("applies correct styles for correct answers", () => {
    const wrapper = mountComponent();
    const correct = wrapper
      .findAll("div")
      .filter((node) => node.classes().includes("bg-jv-mint"));
    expect(correct.length).toBeGreaterThanOrEqual(1);
    expect(correct[0].text()).toContain("Rome");
  });

  it("applies styles for wrong selected answers", async () => {
    const wrapper = mountComponent({
      options: { 1: "10", 2: "120", 3: "240", 4: "20" },
      correctAnswer: "[2]",
      selectedAnswer: "[3]",
      selectedAnswers: {},
      optionsMedia: "text",
      isAdminAnalysis: false,
    });
    const wrong = wrapper
      .findAll("div")
      .filter((node) => node.classes().includes("bg-jv-salmon/50"));
    expect(wrong.length).toBeGreaterThanOrEqual(1);
    expect(wrong[0].text()).toContain("240");
  });

  it("handles empty props correctly", async () => {
    const wrapper = mountComponent({
      options: {},
      correctAnswer: [],
      selectedAnswer: "",
      selectedAnswers: {},
      optionsMedia: "",
      isAdminAnalysis: false,
    });
    expect(wrapper.text()).not.toContain("Paris");
  });
});
