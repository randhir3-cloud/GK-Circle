import { describe, expect, it, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import AnalyticsSummaryCard from "~/components/analytics/AnalyticsSummaryCard.vue";
import StudyTimeCard from "~/components/analytics/StudyTimeCard.vue";
import RecentActivityTable from "~/components/analytics/RecentActivityTable.vue";

describe("learner analytics components", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("renders summary card values", () => {
    const wrapper = mount(AnalyticsSummaryCard, {
      props: {
        label: "Completion rate",
        value: "50%",
        hint: "1 of 2 attempts",
      },
    });
    expect(wrapper.text()).toContain("Completion rate");
    expect(wrapper.text()).toContain("50%");
    expect(wrapper.text()).toContain("1 of 2 attempts");
  });

  it("labels engaged study time as approximate", () => {
    const wrapper = mount(StudyTimeCard, {
      props: {
        assessmentDurationSeconds: 120,
        engagedQuestionTimeSeconds: 90,
      },
    });
    expect(wrapper.text()).toContain("Approximate telemetry");
    expect(wrapper.text()).toContain("2m");
  });

  it("shows Result Pending without inventing a score", async () => {
    const wrapper = mount(RecentActivityTable, {
      props: {
        items: [
          {
            attempt_id: "a1",
            quiz_title: "BPSC Mock",
            status: "SUBMITTED",
            result_status: "Result Pending",
            percentage: null,
            activity_at: "2026-07-28T10:00:00Z",
          },
        ],
        hasMore: false,
      },
    });
    expect(wrapper.text()).toContain("Result Pending");
    expect(wrapper.text()).not.toContain("100%");
    await wrapper.find("tbody tr").trigger("click");
    expect(wrapper.emitted("select-attempt")?.[0]).toEqual(["a1"]);
  });
});

describe("learner_analytics composable helpers", () => {
  it("exports error helper", async () => {
    const mod = await import("~/composables/learner_analytics");
    expect(mod.getLearnerAnalyticsAPIError({ message: "x" }, "fallback")).toBe(
      "x"
    );
  });
});
