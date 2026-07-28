import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";

import ResultReleaseStatusCard from "../../../components/instructor/ResultReleaseStatusCard.vue";
import ReleaseResultsDialog from "../../../components/instructor/ReleaseResultsDialog.vue";
import ResultPreviewPanel from "../../../components/instructor/ResultPreviewPanel.vue";

describe("EXAM-P7-T03 Frontend Instructor Result Management Components", () => {
  it("ResultReleaseStatusCard renders status metrics accurately", () => {
    const wrapper = mount(ResultReleaseStatusCard, {
      props: {
        policy: "SCHEDULED",
        isCurrentlyReleased: false,
        scheduledAt: "2026-12-31T23:59:59Z",
        releasedAt: null,
        totalSubmittedAttempts: 12,
      },
    });

    expect(wrapper.text()).toContain("Results Withheld");
    expect(wrapper.text()).toContain("SCHEDULED");
    expect(wrapper.text()).toContain("12");
  });

  it("ReleaseResultsDialog emits confirm when clicked", async () => {
    const wrapper = mount(ReleaseResultsDialog, {
      props: {
        isOpen: true,
        totalSubmittedAttempts: 5,
        isSubmitting: false,
      },
    });

    const confirmBtn = wrapper
      .findAll("button")
      .find((b) => b.text().includes("Confirm Release"));
    expect(confirmBtn).toBeTruthy();

    await confirmBtn.trigger("click");
    expect(wrapper.emitted("confirm")).toBeTruthy();
  });

  it("ResultPreviewPanel toggles preview state purely client-side", async () => {
    const wrapper = mount(ResultPreviewPanel, {
      props: {
        mockPermissions: {
          showScore: true,
          showPassFail: true,
          allowAnswerReview: true,
          showCorrectness: true,
          showExplanations: true,
        },
      },
    });

    expect(wrapper.text()).toContain("85.0 / 100.0");

    const withheldBtn = wrapper
      .findAll("button")
      .find((b) => b.text().includes("Results Withheld"));
    await withheldBtn.trigger("click");

    expect(wrapper.text()).toContain("Results Pending Release");
  });
});
