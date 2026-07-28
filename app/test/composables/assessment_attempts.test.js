import { mockNuxtImport } from "@nuxt/test-utils/runtime";
import { mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import AttemptInstructionsPanel from "@/components/attempt/AttemptInstructionsPanel.vue";
import {
  formatDurationSeconds,
  getAssessmentAttemptAPIError,
  useAssessmentAttemptsApi,
} from "@/composables/assessment_attempts";

const fetchMock = vi.hoisted(() => vi.fn());

mockNuxtImport("useRuntimeConfig", () => () => ({
  public: { apiUrl: "http://api.test/api/v1" },
}));
mockNuxtImport("useRequestHeaders", () => () => ({ cookie: "session=test" }));
vi.stubGlobal("$fetch", fetchMock);

beforeEach(() => {
  fetchMock.mockReset();
});

describe("assessment_attempts composable", () => {
  it("loads instructions and creates an attempt with snapshot_id", async () => {
    fetchMock
      .mockResolvedValueOnce({
        data: {
          quiz: { title: "PCS Practice", max_attempts: 2 },
          snapshot: { question_count: 10 },
          can_start: true,
        },
      })
      .mockResolvedValueOnce({
        data: { id: "attempt-1", status: "IN_PROGRESS" },
      });

    const api = useAssessmentAttemptsApi();
    const instructions = await api.getInstructions("quiz-1", "snap-1");
    const created = await api.createAttempt("quiz-1", "snap-1");

    expect(instructions.can_start).toBe(true);
    expect(created.id).toBe("attempt-1");
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://api.test/api/v1/quizzes/quiz-1/attempts/instructions?snapshot_id=snap-1",
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://api.test/api/v1/quizzes/quiz-1/attempts",
      expect.objectContaining({
        method: "POST",
        body: { snapshot_id: "snap-1" },
      })
    );
  });

  it("resumes an attempt and maps API errors", async () => {
    fetchMock.mockResolvedValueOnce({
      data: { id: "attempt-2", progress: { answered_count: 1 } },
    });
    const api = useAssessmentAttemptsApi();
    await api.resumeAttempt("quiz-1", "attempt-2");
    expect(fetchMock).toHaveBeenCalledWith(
      "http://api.test/api/v1/quizzes/quiz-1/attempts/attempt-2/resume",
      expect.objectContaining({ credentials: "include" })
    );
    expect(
      getAssessmentAttemptAPIError(
        { data: { data: "assessment attempt not found" } },
        "fallback"
      )
    ).toBe("assessment attempt not found");
  });

  it("formats duration labels", () => {
    expect(formatDurationSeconds(null)).toBe("Untimed");
    expect(formatDurationSeconds(90)).toBe("1 min");
    expect(formatDurationSeconds(3660)).toBe("1h 1m");
  });

  it("autosaves through the Phase 5 PUT endpoint", async () => {
    fetchMock.mockResolvedValueOnce({
      data: { question_id: "q-1", selected_options: [1] },
    });
    const api = useAssessmentAttemptsApi();
    await api.autosaveAnswer("quiz-1", "attempt-1", "q-1", {
      selected_options: [1],
      clear: false,
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "http://api.test/api/v1/quizzes/quiz-1/attempts/attempt-1/answers/q-1",
      expect.objectContaining({
        method: "PUT",
        body: { selected_options: [1], clear: false },
      })
    );
  });
});

describe("AttemptInstructionsPanel", () => {
  const baseInstructions = {
    quiz: {
      title: "UPPCS Prelims Drill",
      description: "Answer carefully.",
      max_attempts: 3,
      negative_marks_per_question: 0.33,
      duration_seconds: 1800,
    },
    snapshot: { question_count: 20 },
    attempts_consumed: 1,
    can_start: true,
    can_resume: false,
  };

  it("renders rules without answer-key fields and emits start", async () => {
    const wrapper = mount(AttemptInstructionsPanel, {
      props: { instructions: baseInstructions },
    });

    expect(wrapper.text()).toContain("UPPCS Prelims Drill");
    expect(wrapper.text()).toContain("20");
    expect(wrapper.text()).toContain("−0.33 per incorrect answer");
    expect(wrapper.text()).not.toContain("authoritative_answer");
    expect(wrapper.text()).not.toContain("official_answer");

    await wrapper.get("button").trigger("click");
    expect(wrapper.emitted("start")).toHaveLength(1);
  });

  it("offers resume when an active attempt exists", async () => {
    const wrapper = mount(AttemptInstructionsPanel, {
      props: {
        instructions: {
          ...baseInstructions,
          can_start: false,
          can_resume: true,
          active_attempt: {
            id: "attempt-9",
            attempt_number: 2,
            status: "IN_PROGRESS",
          },
          block_reason:
            "an in-progress attempt already exists; resume it to continue",
        },
      },
    });

    expect(wrapper.text()).toContain("Resume attempt");
    await wrapper.get("button").trigger("click");
    expect(wrapper.emitted("resume")[0]).toEqual(["attempt-9"]);
  });
});
