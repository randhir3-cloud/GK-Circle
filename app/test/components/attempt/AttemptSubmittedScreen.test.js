import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import AttemptSubmittedScreen from "@/components/attempt/AttemptSubmittedScreen.vue";
import {
  ATTEMPT_STATUS_AUTO_SUBMITTED,
  ATTEMPT_STATUS_SUBMITTED,
} from "@/utils/attempt_player_constants";

describe("AttemptSubmittedScreen.vue", () => {
  it("renders manual submission details", () => {
    const wrapper = mount(AttemptSubmittedScreen, {
      props: {
        attemptId: "att-12345",
        status: ATTEMPT_STATUS_SUBMITTED,
        submittedAt: "2026-07-28T10:00:00Z",
        summary: { answered_count: 9, total_questions: 10 },
        instructionsPath: "/quizzes/q1/instructions",
      },
      global: {
        stubs: {
          NuxtLink: {
            template: "<a><slot /></a>",
          },
        },
      },
    });

    expect(wrapper.text()).toContain("Attempt Submitted");
    expect(wrapper.text()).toContain("Submitted manually");
    expect(wrapper.text()).toContain("att-12345");
    expect(wrapper.text()).toContain("9 of 10");
    const links = wrapper.findAll("a");
    expect(links[0].text()).toContain("View Assessment Results");
    expect(links[1].attributes("to")).toBe("/quizzes/q1/instructions");
  });

  it("renders auto-submission details when time expired", () => {
    const wrapper = mount(AttemptSubmittedScreen, {
      props: {
        attemptId: "att-67890",
        status: ATTEMPT_STATUS_AUTO_SUBMITTED,
        submittedAt: "2026-07-28T10:05:00Z",
        instructionsPath: "/quizzes/q1/instructions",
      },
      global: {
        stubs: {
          NuxtLink: {
            template: "<a><slot /></a>",
          },
        },
      },
    });

    expect(wrapper.text()).toContain("Attempt Auto-Submitted");
    expect(wrapper.text()).toContain("Auto-submitted (time expired)");
    expect(wrapper.text()).toContain("att-67890");
    expect(wrapper.text()).toContain("time expired");
  });
});
