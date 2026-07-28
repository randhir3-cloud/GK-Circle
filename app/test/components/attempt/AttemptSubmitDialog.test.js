import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import AttemptSubmitDialog from "@/components/attempt/AttemptSubmitDialog.vue";

describe("AttemptSubmitDialog.vue", () => {
  it("renders answered/unanswered counts and time remaining", () => {
    const wrapper = mount(AttemptSubmitDialog, {
      props: {
        answeredCount: 8,
        unansweredCount: 2,
        totalQuestions: 10,
        formattedTime: "14:20",
        submitting: false,
      },
    });

    expect(wrapper.text()).toContain("Submit Assessment Attempt?");
    expect(wrapper.text()).toContain("8 / 10");
    expect(wrapper.text()).toContain("2");
    expect(wrapper.text()).toContain("14:20");
    expect(wrapper.text()).toContain("You have 2 unanswered questions");
  });

  it("emits cancel when Continue Attempt button is clicked", async () => {
    const wrapper = mount(AttemptSubmitDialog, {
      props: {
        answeredCount: 5,
        unansweredCount: 0,
        totalQuestions: 5,
        submitting: false,
      },
    });

    const cancelButton = wrapper.find(".submit-dialog__button--cancel");
    await cancelButton.trigger("click");
    expect(wrapper.emitted("cancel")).toHaveLength(1);
  });

  it("emits confirm when Confirm Submission button is clicked", async () => {
    const wrapper = mount(AttemptSubmitDialog, {
      props: {
        answeredCount: 5,
        unansweredCount: 0,
        totalQuestions: 5,
        submitting: false,
      },
    });

    const confirmButton = wrapper.find(".submit-dialog__button--confirm");
    await confirmButton.trigger("click");
    expect(wrapper.emitted("confirm")).toHaveLength(1);
  });

  it("disables buttons when submitting is true", () => {
    const wrapper = mount(AttemptSubmitDialog, {
      props: {
        answeredCount: 5,
        unansweredCount: 0,
        totalQuestions: 5,
        submitting: true,
      },
    });

    const confirmButton = wrapper.find(".submit-dialog__button--confirm");
    const cancelButton = wrapper.find(".submit-dialog__button--cancel");
    expect(confirmButton.attributes("disabled")).toBeDefined();
    expect(cancelButton.attributes("disabled")).toBeDefined();
    expect(confirmButton.text()).toBe("Submitting…");
  });

  it("has proper ARIA dialog attributes", () => {
    const wrapper = mount(AttemptSubmitDialog, {
      props: {
        answeredCount: 5,
        unansweredCount: 0,
        totalQuestions: 5,
      },
    });

    const dialog = wrapper.find('[role="dialog"]');
    expect(dialog.exists()).toBe(true);
    expect(dialog.attributes("aria-modal")).toBe("true");
    expect(dialog.attributes("aria-labelledby")).toBe("submit-dialog-title");
    expect(dialog.attributes("aria-describedby")).toBe("submit-dialog-desc");
  });
});
