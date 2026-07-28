import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import AttemptQuestionPalette from "@/components/attempt/AttemptQuestionPalette.vue";
import AttemptQuestionPanel from "@/components/attempt/AttemptQuestionPanel.vue";
import { PALETTE_STATUS } from "@/utils/attempt_player_constants";

describe("AttemptQuestionPalette", () => {
  it("emits palette navigation and exposes current question", async () => {
    const wrapper = mount(AttemptQuestionPalette, {
      props: {
        items: [
          { question_id: "q-1", position: 0 },
          { question_id: "q-2", position: 1 },
        ],
        currentIndex: 1,
        paletteStatusFor: (id) =>
          id === "q-1" ? PALETTE_STATUS.ANSWERED : PALETTE_STATUS.NOT_VISITED,
      },
    });

    const buttons = wrapper.findAll("button");
    expect(buttons).toHaveLength(2);
    expect(buttons[1].attributes("aria-current")).toBe("step");
    await buttons[0].trigger("click");
    expect(wrapper.emitted("select-index")[0]).toEqual([0]);
  });
});

describe("AttemptQuestionPanel", () => {
  const item = {
    question_id: "q-1",
    question: "Capital of India?",
    type: 1,
    options: { 1: "Delhi", 2: "Mumbai" },
    options_media: "text",
  };

  it("emits option toggles and clear actions", async () => {
    const wrapper = mount(AttemptQuestionPanel, {
      props: {
        item,
        index: 0,
        total: 2,
        draftSelection: [1],
        saveStatus: "idle",
      },
    });

    await wrapper.find('input[type="radio"]').trigger("change");
    expect(wrapper.emitted("toggle-option")).toBeTruthy();

    await wrapper.get(".attempt-question__secondary").trigger("click");
    expect(wrapper.emitted("clear-answer")).toBeTruthy();
  });

  it("shows retry on save failure", () => {
    const wrapper = mount(AttemptQuestionPanel, {
      props: {
        item,
        index: 0,
        total: 2,
        draftSelection: [1],
        saveStatus: "failed",
        saveError: "Network error",
      },
    });

    expect(wrapper.text()).toContain("Network error");
    expect(wrapper.find(".attempt-question__retry").exists()).toBe(true);
  });
});
