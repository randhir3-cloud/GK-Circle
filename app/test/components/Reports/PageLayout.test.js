import { describe, it, expect, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import PageLayout from "~/components/reports/PageLayout.vue";

describe("QuizAnalysisTabs / PageLayout", () => {
  let wrapper;

  beforeEach(() => {
    wrapper = mount(PageLayout, {
      props: {
        currentTab: "report",
      },
    });
  });

  it("renders the navigation tabs", () => {
    const buttons = wrapper.findAll("button");
    expect(buttons).toHaveLength(2);
    expect(buttons[0].text()).toContain("Questions");
    expect(buttons[1].text()).toContain("Participants");
  });

  it("marks the Questions tab as active for report", () => {
    const questions = wrapper.findAll("button")[0];
    expect(questions.classes().join(" ")).toContain("text-jv-ink");
  });

  it("emits changeTab with participants when Participants is clicked", async () => {
    const participantsTab = wrapper.findAll("button")[1];
    await participantsTab.trigger("click");
    expect(wrapper.emitted().changeTab).toBeTruthy();
    expect(wrapper.emitted().changeTab[0]).toEqual(["participants"]);
  });

  it("marks the Participants tab as active when currentTab is participants", async () => {
    await wrapper.setProps({ currentTab: "participants" });
    const participants = wrapper.findAll("button")[1];
    expect(participants.classes().join(" ")).toContain("text-jv-ink");
  });
});
